package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/screen"
	"github.com/tdx/go-tdx-mcp/scraper"
	"github.com/tdx/go-tdx-mcp/tdx"
)

func strPtr(s string) *string { return &s }

type _stock struct {
	Code   string          `json:"code"`
	Market int             `json:"market"`
	Bars   []indicator.Bar `json:"bars"`
}

func (s *Server) getTCPClient() *tdx.TDXTCPClient {
	if uc, ok := s.client.(*tdx.UnifiedClient); ok {
		return uc.TCPClient()
	}
	return nil
}

type backtestAsyncReq struct {
	Strategy string            `json:"strategy"`
	Params   map[string]float64 `json:"params"`
	Code     string            `json:"code"`
	Market   int               `json:"market"`
	Period   string            `json:"period"`
	Count    int               `json:"count"`
	Cash     float64           `json:"cash"`
	OHLCV    []indicator.Bar `json:"ohlcv"`
}

type backtestOptimizeReq struct {
	Strategy string            `json:"strategy"`
	Params   map[string][]float64 `json:"params"`
	Code     string            `json:"code"`
	Market   int               `json:"market"`
	Period   string            `json:"period"`
	Count    int               `json:"count"`
	Cash     float64           `json:"cash"`
}

type backtestPortfolioReq struct {
	Strategy string            `json:"strategy"`
	Params   map[string]float64 `json:"params"`
	Stocks   []_stock           `json:"stocks"`
	Period   string             `json:"period"`
	Count    int                `json:"count"`
	Cash     float64            `json:"cash"`
}

type backtestMultiStrategyReq struct {
	Items  []struct {
		Strategy   string            `json:"strategy"`
		Params     map[string]float64 `json:"params"`
		Code       string            `json:"code"`
		Market     int               `json:"market"`
		Allocation float64           `json:"allocation"`
		Bars       []indicator.Bar   `json:"bars"`
	} `json:"items"`
	Period string  `json:"period"`
	Count  int     `json:"count"`
	Cash   float64 `json:"cash"`
}

func (s *Server) handleBacktestAsync(w http.ResponseWriter, r *http.Request) {
	var req backtestAsyncReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Cash <= 0 {
		req.Cash = 1000000
	}
	if req.Period == "" {
		req.Period = "day"
	}
	if req.Count <= 0 {
		req.Count = 500
	}

	bars := req.OHLCV
	if req.Code != "" {
		market := req.Market
		if market < 0 {
			market = 0
		}
		fetchedBars, err := s.fetchKlines(req.Code, market, req.Period, req.Count, 0)
		if err != nil {
			task := &backtest.BacktestTask{
				ID: shortTaskID(), Strategy: req.Strategy, Params: req.Params,
				Cash: req.Cash, Market: market, Code: req.Code,
				Period: req.Period, Count: req.Count, Error: err.Error(),
			}
			s.taskRunner.SubmitJob(task)
			writeJSON(w, map[string]string{"id": task.ID, "status": "failed", "error": err.Error()})
			return
		}
		bars = fetchedBars
	}

	task := &backtest.BacktestTask{
		ID: shortTaskID(), Strategy: req.Strategy, Params: req.Params,
		Cash: req.Cash, Market: req.Market, Code: req.Code,
		Period: req.Period, Count: req.Count,
		Job: func() *backtest.Result {
			st := backtest.NewStrategyWithParams(req.Strategy, req.Params)
			if st == nil {
				return nil
			}
			engine := backtest.NewEngine(req.Cash)
			result := engine.Run(st, bars)
			result.Code = req.Code
			result.Market = req.Market
			result.Period = req.Period
			return result
		},
	}
	s.taskRunner.SubmitJob(task)
	writeJSON(w, map[string]string{"id": task.ID, "status": "pending"})
}

func (s *Server) handleBacktestTasks(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	results := s.taskRunner.ListRecent(limit)
	writeJSON(w, results)
}

func (s *Server) handleBacktestTaskStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/backtest/tasks/")
	if id == "" {
		writeError(w, 400, "task_id 必填")
		return
	}
	result := s.taskRunner.Peek(id)
	if result == nil {
		writeError(w, 404, "任务不存在或已过期")
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleBacktestOptimize(w http.ResponseWriter, r *http.Request) {
	var req backtestOptimizeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Cash <= 0 {
		req.Cash = 1000000
	}
	if req.Period == "" {
		req.Period = "day"
	}
	if req.Count <= 0 {
		req.Count = 500
	}
	if req.Code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	market := req.Market
	if market < 0 {
		market = 0
	}
	bars, err := s.fetchKlines(req.Code, market, req.Period, req.Count, 0)
	if err != nil {
		writeError(w, 500, "获取K线失败: "+err.Error())
		return
	}
	optReq := &backtest.OptimizeRequest{
		Strategy: req.Strategy, Params: req.Params, Bars: bars, Cash: req.Cash,
	}
	result := backtest.RunOptimizer(optReq)
	if result == nil {
		writeError(w, 400, "网格点过多或策略不存在")
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleBacktestPortfolio(w http.ResponseWriter, r *http.Request) {
	var req backtestPortfolioReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Cash <= 0 {
		req.Cash = 1000000
	}
	if req.Period == "" {
		req.Period = "day"
	}
	if req.Count <= 0 {
		req.Count = 500
	}
	st := backtest.NewStrategyWithParams(req.Strategy, req.Params)
	if st == nil {
		writeError(w, 400, "不支持的策略: "+req.Strategy)
		return
	}
	items := make([]backtest.PortfolioItem, 0, len(req.Stocks))
	for _, stock := range req.Stocks {
		var bars []indicator.Bar
		var err error
		if len(stock.Bars) > 0 {
			bars = stock.Bars
		} else {
			market := stock.Market
			if market < 0 {
				market = 0
			}
			bars, err = s.fetchKlines(stock.Code, market, req.Period, req.Count, 0)
			if err != nil {
				continue
			}
		}
		items = append(items, backtest.PortfolioItem{Code: stock.Code, Market: stock.Market, Bars: bars})
	}
	if len(items) == 0 {
		writeError(w, 500, "未获取到任何标的K线数据")
		return
	}
	result := backtest.RunPortfolio(st, items, req.Cash)
	writeJSON(w, result)
}

func (s *Server) handleBacktestMultiStrategy(w http.ResponseWriter, r *http.Request) {
	var req backtestMultiStrategyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Cash <= 0 {
		req.Cash = 1000000
	}
	if req.Period == "" {
		req.Period = "day"
	}
	if req.Count <= 0 {
		req.Count = 500
	}
	multiItems := make([]backtest.MultiStrategyItem, 0, len(req.Items))
	for _, item := range req.Items {
		var bars []indicator.Bar
		var err error
		if len(item.Bars) > 0 {
			bars = item.Bars
		} else {
			market := item.Market
			if market < 0 {
				market = 0
			}
			bars, err = s.fetchKlines(item.Code, market, req.Period, req.Count, 0)
			if err != nil {
				continue
			}
		}
		multiItems = append(multiItems, backtest.MultiStrategyItem{
			Strategy: item.Strategy, Params: item.Params,
			Code: item.Code, Market: item.Market, Bars: bars,
			Allocation: item.Allocation,
		})
	}
	if len(multiItems) == 0 {
		writeError(w, 500, "未获取到任何标的K线数据")
		return
	}
	result := backtest.RunMultiStrategy(multiItems, req.Cash)
	writeJSON(w, result)
}

func (s *Server) handleBacktestStrategies(w http.ResponseWriter, r *http.Request) {
	schemas := backtest.AvailableStrategySchemas()
	writeJSON(w, schemas)
}

func (s *Server) handleSignalScan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Strategies []struct {
			Strategy      string            `json:"strategy"`
			StrategyLabel string            `json:"strategy_label"`
			Params        map[string]float64 `json:"params"`
			Code          string            `json:"code"`
			Market        int               `json:"market"`
			Window        int               `json:"window"`
			Bars          []indicator.Bar   `json:"bars"`
			Period        string            `json:"period"`
			Count         int               `json:"count"`
			StrategyID    string            `json:"strategy_id"`
			StrategyName  string            `json:"strategy_name"`
			Kind          string            `json:"kind"`
			Category      string            `json:"category"`
		} `json:"strategies"`
		Strategy string `json:"strategy"`
		Params   map[string]float64 `json:"params"`
		Market   int              `json:"market"`
		Period   string           `json:"period"`
		Count    int              `json:"count"`
		Codes    []string         `json:"codes"`
		Window   int              `json:"window"`
		TopN     int              `json:"top_n"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}

	targets := make([]backtest.ScanTarget, 0)

	if len(req.Strategies) > 0 {
		if req.Window <= 0 {
			req.Window = 30
		}
		for _, item := range req.Strategies {
			bars := item.Bars
			if len(bars) == 0 && item.Code != "" {
				market := item.Market
				if market < 0 {
					market = 0
				}
				period := item.Period
				if period == "" {
					period = req.Period
				}
				if period == "" {
					period = "1"
				}
				count := item.Count
				if count <= 0 {
					count = 500
				}
				fetched, err := s.fetchKlines(item.Code, market, period, count, 0)
				if err != nil {
					targets = append(targets, backtest.ScanTarget{
						Strategy: item.Strategy, StrategyLabel: item.StrategyLabel, Params: item.Params,
						Code: item.Code, Market: market, Error: err.Error(),
						StrategyID: item.StrategyID, StrategyName: item.StrategyName,
						Kind: item.Kind, Category: item.Category,
					})
					continue
				}
				bars = fetched
			}
			w := req.Window
			if item.Window > 0 {
				w = item.Window
			}
			targets = append(targets, backtest.ScanTarget{
				Strategy: item.Strategy, StrategyLabel: item.StrategyLabel, Params: item.Params,
				Code: item.Code, Market: item.Market, Bars: bars, Window: w,
				StrategyID: item.StrategyID, StrategyName: item.StrategyName,
				Kind: item.Kind, Category: item.Category,
			})
		}
	} else if req.Strategy != "" {
		if req.Period == "" {
			req.Period = "1"
		}
		if req.Count <= 0 {
			req.Count = 500
		}
		if req.Window <= 0 {
			req.Window = 30
		}
		stockCodes := req.Codes
		if len(stockCodes) == 0 {
			stockCodes = fetchSecurityCodesFallback(req.Market, req.TopN)
		}
		for _, code := range stockCodes {
			fetched, err := s.fetchKlines(code, req.Market, req.Period, req.Count, 0)
			if err != nil {
				targets = append(targets, backtest.ScanTarget{
					Strategy: req.Strategy, Params: req.Params,
					Code: code, Market: req.Market, Error: err.Error(),
				})
				continue
			}
			targets = append(targets, backtest.ScanTarget{
				Strategy: req.Strategy, Params: req.Params,
				Code: code, Market: req.Market, Bars: fetched, Window: req.Window,
			})
		}
	}

	if len(targets) == 0 {
		writeJSON(w, backtest.SignalScanResult{Rows: []backtest.ScanRow{}, Total: 0})
		return
	}

	result := backtest.RunSignalScan(targets, req.Window)
	writeJSON(w, result)
}

func fetchSecurityCodesFallback(market int, topN int) []string {
	if topN <= 0 {
		topN = 50
	}
	if topN > 200 {
		topN = 200
	}
	return getTopStocks(market, topN)
}

func getTopStocks(market int, topN int) []string {
	all := []string{
		"000001", "000002", "000004", "000006", "000009",
		"000016", "000021", "000025", "000027", "000028",
		"000039", "000056", "000060", "000063", "000066",
		"000089", "000100", "000157", "000333", "000400",
		"000402", "000423", "000425", "000429", "000488",
		"000501", "000519", "000528", "000538", "000543",
		"000563", "000568", "000581", "000596", "000601",
		"000605", "000617", "000623", "000625", "000630",
		"000636", "000651", "000661", "000665", "000671",
		"000672", "000676", "000680", "000682", "000683",
		"000690", "000692", "000698", "000700", "000703",
		"000706", "000708", "000709", "000717", "000718",
		"000720", "000721", "000723", "000725", "000726",
		"000728", "000729", "000732", "000733", "000738",
		"000739", "000746", "000750", "000751", "000753",
		"000756", "000757", "000758", "000761", "000762",
	}
	if market == 1 {
		all = []string{
			"600000", "600004", "600005", "600006", "600007",
			"600008", "600009", "600010", "600011", "600012",
			"600015", "600016", "600018", "600019", "600020",
			"600021", "600022", "600023", "600025", "600026",
			"600027", "600028", "600029", "600030", "600031",
			"600032", "600033", "600034", "600035", "600036",
			"600037", "600038", "600039", "600048", "600050",
			"600052", "600053", "600054", "600055", "600056",
			"600058", "600059", "600060", "600061", "600062",
			"600063", "600064", "600065", "600066", "600067",
			"600068", "600069", "600070", "600071", "600072",
			"600073", "600074", "600075", "600076", "600077",
			"600078", "600079", "600080", "600081", "600082",
			"600083", "600084", "600085", "600086", "600087",
		}
	}
	if topN >= len(all) {
		return all
	}
	return all[:topN]
}

func (s *Server) handleStrategyStore(w http.ResponseWriter, r *http.Request) {
	store := s.strategyStore
	if store == nil {
		writeError(w, 500, "策略存储不可用")
		return
	}
	switch r.Method {
	case http.MethodGet:
		records, err := store.List()
		if err != nil {
			writeError(w, 500, "读取策略失败: "+err.Error())
			return
		}
		writeJSON(w, records)
	case http.MethodPost:
		var rec SavedStrategy
		if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
			writeError(w, 400, "无效的请求体")
			return
		}
		if rec.Name == "" {
			writeError(w, 400, "name 参数必填")
			return
		}
		saved, err := store.Add(&rec)
		if err != nil {
			writeError(w, 500, "保存策略失败: "+err.Error())
			return
		}
		writeJSON(w, saved)
	default:
		writeError(w, 405, "方法不支持")
	}
}

func (s *Server) handleStrategyStoreByID(w http.ResponseWriter, r *http.Request) {
	store := s.strategyStore
	if store == nil {
		writeError(w, 500, "策略存储不可用")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/strategies/")
	if id == "" {
		writeError(w, 400, "strategy_id 必填")
		return
	}
	switch r.Method {
	case http.MethodGet:
		rec, err := store.Get(id)
		if err != nil || rec == nil {
			writeError(w, 404, "策略不存在")
			return
		}
		writeJSON(w, rec)
	case http.MethodDelete:
		deleted, err := store.Delete(id)
		if err != nil {
			writeError(w, 500, "删除失败: "+err.Error())
			return
		}
		writeJSON(w, map[string]interface{}{"id": id, "deleted": deleted})
	default:
		writeError(w, 405, "方法不支持")
	}
}

func (s *Server) handleMinute(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	if s.client == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	periodCode := 104
	body := map[string]interface{}{
		"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
		"Code": code, "Setcode": market, "Period": periodCode,
		"Startxh": 0, "WantNum": 240, "TQFlag": 0,
	}
	rawResp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
	if err != nil {
		writeError(w, 500, "获取分时数据失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "data": rawResp.Data})
}

func (s *Server) handleMinuteHistory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	date := r.URL.Query().Get("date")
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	if s.client == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	periodCode := 104
	body := map[string]interface{}{
		"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
		"Code": code, "Setcode": market, "Period": periodCode,
		"Startxh": 0, "WantNum": 240, "TQFlag": 0,
	}
	if date != "" {
		body["Date"] = date
	}
	rawResp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
	if err != nil {
		writeError(w, 500, "获取历史分时数据失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "date": date, "data": rawResp.Data})
}

func (s *Server) handleTransaction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	count := queryInt(r, "count", 100)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	if count > 500 {
		count = 500
	}
	if s.client == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	body := map[string]interface{}{
		"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
		"Code": code, "Setcode": market, "Count": count,
	}
	rawResp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBTrans", body)
	if err != nil {
		writeError(w, 500, "获取逐笔成交失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "data": rawResp.Data})
}

func (s *Server) handleTransactionHistory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	date := r.URL.Query().Get("date")
	count := queryInt(r, "count", 100)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	if count > 500 {
		count = 500
	}
	if s.client == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	body := map[string]interface{}{
		"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
		"Code": code, "Setcode": market, "Count": count, "Date": date,
	}
	rawResp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBTrans", body)
	if err != nil {
		writeError(w, 500, "获取历史逐笔成交失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "date": date, "data": rawResp.Data})
}

func (s *Server) handleSecurityList(w http.ResponseWriter, r *http.Request) {
	market, ok := parseMarket(r)
	if !ok {
		writeError(w, 400, "market 参数必填")
		return
	}
	start := queryInt(r, "start", 0)
	count := queryInt(r, "count", 200)
	if count > 1000 {
		count = 1000
	}
	startU16 := uint16(start)
	tcp := s.getTCPClient()
	var reply interface{}
	var tcpErr error
	if tcp != nil {
		reply, tcpErr = tcp.GetSecurityList(market, startU16)
	}
	if tcpErr == nil && reply != nil {
		writeJSON(w, map[string]interface{}{"market": market, "start": start, "count": count, "data": reply})
		return
	}
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_F10_FINANCE_MAINFINADATA&columns=SECURITY_CODE,SECURITY_NAME_ABBR,REPORT_DATE&pageSize=%d&pageNumber=1&sortTypes=-1&sortColumns=SECURITY_CODE", count)
	hc := &http.Client{Timeout: 10 * time.Second}
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取证券列表失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"market": market, "start": start, "count": count, "data": data})
}

func (s *Server) handleSecurityListAll(w http.ResponseWriter, r *http.Request) {
	pages := queryInt(r, "pages", 2)
	if pages > 10 {
		pages = 10
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	allData := make([]interface{}, 0)
	for pg := 0; pg < pages; pg++ {
		for market := 0; market <= 1; market++ {
			reply, err := tcp.GetSecurityList(market, uint16(pg*200))
			if err != nil {
				allData = append(allData, map[string]interface{}{"market": market, "page": pg, "error": err.Error()})
				continue
			}
			allData = append(allData, map[string]interface{}{"market": market, "page": pg, "data": reply})
		}
	}
	writeJSON(w, allData)
}

func (s *Server) handleServerHosts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"hosts":    []string{"218.75.129.43", "115.29.133.24", "113.105.142.162"},
		"current":  "auto",
		"note":     "TQLEX HTTP网关模式，无需切换TCP主机",
	})
}

func (s *Server) handleServerTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hosts   []string `json:"hosts"`
		Timeout int      `json:"timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Timeout = 5000
	}
	if req.Timeout <= 0 || req.Timeout > 30000 {
		req.Timeout = 5000
	}
	allowedHosts := []string{"218.75.129.43", "115.29.133.24", "113.105.142.162"}
	if len(req.Hosts) == 0 {
		req.Hosts = allowedHosts
	}
	type pingResult struct {
		Host      string `json:"host"`
		Latency   int    `json:"latency_ms"`
		Reachable bool   `json:"reachable"`
	}
	results := make([]pingResult, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		host = strings.TrimSpace(host)
		if !isAllowedHost(host, allowedHosts) {
			results = append(results, pingResult{Host: host, Reachable: false})
			continue
		}
		start := time.Now()
		hc := &http.Client{Timeout: time.Duration(req.Timeout) * time.Millisecond}
		_, err := hc.Head(fmt.Sprintf("http://%s:7709/", host))
		latency := int(time.Since(start).Milliseconds())
		results = append(results, pingResult{Host: host, Latency: latency, Reachable: err == nil})
	}
	writeJSON(w, results)
}

func isAllowedHost(host string, allowed []string) bool {
	for _, a := range allowed {
		if host == a {
			return true
		}
	}
	return false
}

func (s *Server) handleServerSwitch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host string `json:"host"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	writeJSON(w, map[string]string{"host": req.Host, "status": "TQLEX HTTP网关模式不支持切换主机", "note": "当前使用HTTP网关，无需TCP主机切换"})
}

func (s *Server) handleBarsIndex(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	period := queryParam(r, "period", "day")
	count := queryInt(r, "count", 200)
	fqType := queryInt(r, "fq_type", 0)
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	if !ok {
		upperCode := strings.ToUpper(code)
		if strings.HasPrefix(upperCode, "SH") {
			code = upperCode[2:]
			market = 1
			ok = true
		} else if strings.HasPrefix(upperCode, "SZ") {
			code = upperCode[2:]
			market = 0
			ok = true
		} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			market = 0
			ok = true
		} else {
			market = 1
			ok = true
		}
	}
	indexCode := normalizeCode(code)
	if indexCode == code {
		upperCode := strings.ToUpper(code)
		if strings.HasPrefix(upperCode, "SH") {
			indexCode = upperCode[2:]
			market = 1
		} else if strings.HasPrefix(upperCode, "SZ") {
			indexCode = upperCode[2:]
			market = 0
		}
	}
	bars, err := s.fetchKlines(indexCode, market, period, count, fqType)
	if err != nil {
		writeError(w, 500, "获取指数K线失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"data": bars, "total": len(bars)})
}

func (s *Server) handleBoardChangeRanking(w http.ResponseWriter, r *http.Request) {
	boardType := r.URL.Query().Get("board_type")
	if boardType == "" {
		boardType = "HY"
	}
	days := queryInt(r, "days", 1)
	topN := queryInt(r, "top_n", 30)
	if topN > 100 {
		topN = 100
	}
	targetDate := r.URL.Query().Get("target_date")
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	blockType := tdx.BlockIndustry
	isTCPAvailable := true
	switch boardType {
	case "GN":
		blockType = tdx.BlockConcept
	case "DQ":
		blockType = tdx.BlockRegion
		isTCPAvailable = false
	case "ZS":
		blockType = tdx.BlockIndex
	default:
		blockType = tdx.BlockIndustry
		isTCPAvailable = false
	}
	var boards []tdx.SectorBoard
	var err error
	if isTCPAvailable {
		boards, err = tcp.GetSectorBoards(blockType)
	} else {
		resp, qerr := s.client.TQLEXQuery(context.Background(), "TdxShare.PBBoardList", map[string]interface{}{"BoardType": boardType, "Count": topN})
		if qerr == nil && resp != nil && resp.Data != nil {
			raw, _ := json.Marshal(resp.Data)
			json.Unmarshal(raw, &boards)
		}
	}
	if len(boards) == 0 {
		es := scraper.NewEastMoneyScraper()
		scraperType := "concept"
		if boardType == "HY" {
			scraperType = "industry"
		} else if boardType == "DQ" {
			scraperType = "region"
		} else if boardType == "GN" {
			scraperType = "concept"
		} else if boardType == "ZS" {
			scraperType = "concept"
		}
		scraperBoards, serr := es.SectorBoards(scraperType)
		if serr == nil && len(scraperBoards) > 0 {
			for _, sb := range scraperBoards {
				boards = append(boards, tdx.SectorBoard{
					Code:     fmt.Sprintf("%v", sb["sector_code"]),
					Name:     fmt.Sprintf("%v", sb["sector_name"]),
					Type:     boardType,
					StockCnt: 0,
				})
			}
		} else {
			err = serr
		}
	}
	if err != nil {
		writeJSON(w, map[string]interface{}{"board_type": boardType, "days": days, "top_n": topN, "target_date": targetDate, "data": []interface{}{}, "error": "获取板块列表失败: " + err.Error()})
		return
	}
	if len(boards) > topN {
		boards = boards[:topN]
	}
	writeJSON(w, map[string]interface{}{"board_type": boardType, "days": days, "top_n": topN, "target_date": targetDate, "count": len(boards), "data": boards, "note": "涨幅排名按成分股数量降序排列"})
}

func (s *Server) handleBoardSummary(w http.ResponseWriter, r *http.Request) {
	boardSymbol := r.URL.Query().Get("board_symbol")
	if boardSymbol == "" {
		writeError(w, 400, "board_symbol 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	stocks, err := tcp.GetSectorBoardStocks(boardSymbol)
	if err != nil {
		es := scraper.NewEastMoneyScraper()
		scraperStocks, serr := es.SectorStocks(boardSymbol)
		if serr == nil && len(scraperStocks) > 0 {
			writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "stock_count": len(scraperStocks), "stocks": scraperStocks, "data": scraperStocks})
			return
		}
		ss := scraper.NewSectorScraper()
		oldStocks, serr2 := ss.FetchBoardStocks(boardSymbol)
		if serr2 == nil && len(oldStocks) > 0 {
			writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "stock_count": len(oldStocks), "stocks": oldStocks, "data": oldStocks})
			return
		}
		writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "data": []interface{}{}, "error": "获取板块成员失败: " + err.Error() + ", scraper: " + serr.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "stock_count": len(stocks), "stocks": stocks, "data": stocks})
}

func (s *Server) handleCompanyCategory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	f10, err := tcp.GetF10(code, market)
	if err != nil {
		writeJSON(w, map[string]interface{}{"code": code, "data": []interface{}{}, "error": "获取F10失败: " + err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "data": f10})
}

func (s *Server) handleCompanyContent(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	filename := r.URL.Query().Get("filename")
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	f10, err := tcp.GetF10(code, market)
	if err != nil {
		writeJSON(w, map[string]interface{}{"code": code, "filename": filename, "data": "", "error": "获取F10失败: " + err.Error()})
		return
	}
	if filename == "" {
		writeJSON(w, map[string]interface{}{"code": code, "data": f10})
		return
	}
	writeJSON(w, map[string]interface{}{"code": code, "filename": filename, "data": f10})
}

func (s *Server) handleExMinute(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	if market == "" {
		market = r.URL.Query().Get("ex_market")
	}
	code := r.URL.Query().Get("code")
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(market)
	dateStr := r.URL.Query().Get("date")
	var date uint32
	if dateStr != "" {
		date = uint32(queryInt(r, "date", 0))
	}
	if date > 0 {
		reply, err := tcp.ExGetHistoryMinuteTimeData(date, cat, code)
		if err != nil {
			writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "date": date, "data": []interface{}{}, "error": "获取扩展市场历史分时失败: " + err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "date": date, "data": reply})
		return
	}
	reply, err := tcp.ExGetMinuteTimeData(cat, code)
	if err != nil {
		writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "data": []interface{}{}, "error": "获取扩展市场分时失败: " + err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "data": reply})
}

func (s *Server) handleExTransaction(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	if market == "" {
		market = r.URL.Query().Get("ex_market")
	}
	code := r.URL.Query().Get("code")
	dateStr := r.URL.Query().Get("date")
	count := queryInt(r, "count", 300)
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	if count > 3000 {
		count = 3000
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(market)
	var date uint32
	if dateStr != "" {
		date = uint32(queryInt(r, "date", 0))
	}
	reply, err := tcp.ExGetTransactionData(date, cat, code)
	if err != nil {
		writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "date": date, "data": []interface{}{}, "error": "获取扩展市场逐笔成交失败: " + err.Error()})
		return
	}
	items := make([]interface{}, len(reply.List))
	for i, item := range reply.List {
		items[i] = item
	}
	if len(items) > count {
		items = items[:count]
	}
	writeJSON(w, map[string]interface{}{"market": market, "code": code, "category": cat, "date": date, "count": len(items), "data": items})
}

func (s *Server) handleExList(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	start := queryInt(r, "start", 0)
	count := queryInt(r, "count", 50)
	if count > 500 {
		count = 500
	}
	if exMarket == "" {
		// ex_market not specified: return full count only
		tcp := s.getTCPClient()
		if tcp == nil {
			writeError(w, 503, "TDX服务不可用")
			return
		}
		reply, err := tcp.ExGetCount()
		if err != nil {
			writeJSON(w, map[string]interface{}{"data": []interface{}{}, "error": "获取商品总数失败: " + err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"total": reply.Count, "data": []interface{}{}})
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(exMarket)
	if count > 0 {
		reply, err := tcp.ExGetList(uint32(start), uint16(count))
		if err != nil {
			writeJSON(w, map[string]interface{}{"ex_market": exMarket, "category": cat, "start": start, "count": count, "data": []interface{}{}, "error": "获取扩展市场商品列表失败: " + err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "category": cat, "start": start, "count": count, "data": reply})
		return
	}
	reply, err := tcp.ExGetCount()
	if err != nil {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "category": cat, "data": []interface{}{}, "error": "获取商品总数失败: " + err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"ex_market": exMarket, "category": cat, "total": reply.Count, "data": []interface{}{}})
}

func (s *Server) handleExQuotes(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	codesStr := r.URL.Query().Get("codes")
	if exMarket == "" || codesStr == "" {
		writeError(w, 400, "ex_market 和 codes 参数必填 (codes逗号分隔)")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(exMarket)
	codes := strings.Split(codesStr, ",")
	if len(codes) > 80 {
		codes = codes[:80]
	}
	var stocks []struct {
		Category uint8
		Code     string
	}
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		stocks = append(stocks, struct{ Category uint8; Code string }{Category: cat, Code: c})
	}
	if len(stocks) == 0 {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": codesStr, "category": cat, "data": []interface{}{}})
		return
	}
	reply, err := tcp.ExGetQuotesEx(stocks)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "codes": codesStr, "category": cat, "data": []interface{}{}, "error": "获取扩展市场批量报价失败: " + err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"ex_market": exMarket, "codes": codesStr, "category": cat, "data": reply})
}

func (s *Server) handleExTransactionAll(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	code := r.URL.Query().Get("code")
	dateStr := r.URL.Query().Get("date")
	if exMarket == "" || code == "" {
		writeError(w, 400, "ex_market 和 code 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(exMarket)
	var date uint32
	if dateStr != "" {
		date = uint32(queryInt(r, "date", 0))
	}
	hkStockMarkets := map[uint8]bool{31: true, 48: true, 27: true, 49: true, 71: true, 98: true}
	if !hkStockMarkets[cat] {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": code, "category": cat, "error": "goods_transaction_all 仅支持港股股票类市场"})
		return
	}
	reply, err := tcp.ExGetTransactionData(date, cat, code)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": code, "category": cat, "date": date, "data": []interface{}{}, "error": "获取扩展市场全部逐笔失败: " + err.Error()})
		return
	}
	items := make([]interface{}, len(reply.List))
	for i, item := range reply.List {
		items[i] = item
	}
	writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": code, "category": cat, "date": date, "total": len(items), "data": items, "note": "返回最近一批逐笔数据(gotdx单次上限),如需全天请多次调用handleExTransaction设置不同date"})
}

func (s *Server) handleExChartSampling(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	code := r.URL.Query().Get("code")
	if exMarket == "" || code == "" {
		writeError(w, 400, "ex_market 和 code 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	cat := tdx.ExMarketCategory(exMarket)
	reply, err := tcp.ExGetChartSampling(cat, code)
	if err != nil {
		writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": code, "category": cat, "data": []interface{}{}, "error": "获取扩展市场分时缩略失败: " + err.Error()})
		return
	}
	writeJSON(w, map[string]interface{}{"ex_market": exMarket, "code": code, "category": cat, "data": reply})
}

func (s *Server) handleFinance(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_F10_FINANCE_MAINFINADATA&columns=ALL&filter=(SECURITY_CODE=%s)&pageSize=10&pageNumber=1&sortBy=REPORT_DATE&sortType=desc", code)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取财务快照失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "data": data})
}

func (s *Server) handleFinancialFileList(w http.ResponseWriter, r *http.Request) {
	hc := &http.Client{Timeout: 10 * time.Second}
	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_F10_FINANCE_MAINFINADATA&columns=REPORT_DATE,SECURITY_CODE&pageSize=20&pageNumber=1"
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取财务文件列表失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, data)
}

func (s *Server) handleFinancialRecords(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeError(w, 400, "filename 参数必填")
		return
	}
	writeJSON(w, map[string]string{"filename": filename, "note": "历史财报文件读取需通达信本地目录，请使用 /financial/report 获取在线财报数据"})
}

func (s *Server) handleFundFlowHistory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	count := queryInt(r, "count", 30)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	if count > 365 {
		count = 365
	}
	setcodeInt := market
	if setcodeInt < 0 {
		setcodeInt = 0
	}
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/fflow/daykline/get?secid=%d.%s&fields1=f1,f2,f3,f7&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f62,f63&klt=101&lmt=%d", setcodeInt, code, count)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取历史资金流向失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "data": data})
}

func (s *Server) handleMarketStrength(w http.ResponseWriter, r *http.Request) {
	preset := queryParam(r, "preset", "balanced")
	topN := queryInt(r, "top_n", 20)
	if topN > 200 {
		topN = 200
	}
	_ = queryParam(r, "vipdoc", "")
	_ = queryParam(r, "universe", "")
	minListedDays := queryInt(r, "min_listed_days", 60)
	minAmount := queryFloat(r, "min_amount", 0)
	w5 := queryFloat(r, "w5", 1.0)
	w20 := queryFloat(r, "w20", 1.0)
	w60 := queryFloat(r, "w60", 1.0)
	volAdjusted := queryInt(r, "vol_adjusted", 0) > 0

	ranker := screen.NewStrengthRanker(preset)
	results := ranker.Rank(topN)
	writeJSON(w, map[string]interface{}{"preset": preset, "w5": w5, "w20": w20, "w60": w60, "vol_adjusted": volAdjusted, "min_listed_days": minListedDays, "min_amount": minAmount, "results": results})
}

func (s *Server) handleFactorList(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	var names []string
	if category != "" {
		names = factor.ListByCategory(category)
	} else {
		names = factor.List()
	}
	writeJSON(w, map[string]interface{}{"category": category, "factors": names})
}

func (s *Server) handleFactorCompute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Factor     string          `json:"factor"`
		FactorNames []string        `json:"factor_names"`
		Bars       []indicator.Bar `json:"bars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	factorNames := req.FactorNames
	if len(factorNames) == 0 && req.Factor != "" {
		factorNames = []string{req.Factor}
	}
	engine := factor.NewEngine()
	values, err := engine.ComputeSingle(req.Bars, factorNames)
	if err != nil {
		writeError(w, 500, "因子计算失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"factors": factorNames, "values": values})
}

func (s *Server) handleFactorAnalyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FactorName  string   `json:"factor"`
		FactorValues []float64 `json:"factor_values"`
		Returns      []float64 `json:"returns"`
		NQuantiles   int     `json:"n_quantiles"`
		Period       int     `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.NQuantiles <= 0 {
		req.NQuantiles = 5
	}
	if len(req.FactorValues) == 0 || len(req.Returns) == 0 {
		writeJSON(w, map[string]string{"note": "需提供 factor_values 和 returns 数据", "period": fmt.Sprintf("%d", req.Period)})
		return
	}
	analyzer := factor.NewAnalyzer(req.FactorValues, req.Returns, req.FactorName, req.NQuantiles)
	report := analyzer.FullReport()
	writeJSON(w, report)
}

type tdxTQLEXResp struct {
	Data   interface{} `json:"data"`
	Errmsg string      `json:"errmsg"`
}

func (s *Server) handleSignalRank(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Signals []struct {
			Code   string            `json:"code"`
			Market int               `json:"market"`
			Name   string            `json:"name"`
			Bars   []indicator.Bar   `json:"bars"`
			Period string            `json:"period"`
			Count  int               `json:"count"`
		} `json:"signals"`
		Codes   []string         `json:"codes"`
		Strategy string           `json:"strategy"`
		Params   map[string]float64 `json:"params"`
		Market   int              `json:"market"`
		Period   string           `json:"period"`
		Count    int              `json:"count"`
		SortBy   string           `json:"sort_by"`
		SortReverse bool         `json:"sort_reverse"`
		TopN     int              `json:"top_n"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Strategy == "" {
		req.Strategy = "MA_CROSS"
	}
	if req.Period == "" {
		req.Period = "1"
	}
	if req.SortBy == "" {
		req.SortBy = "sharpe"
	}

	var inputs []backtest.RankInput

	if len(req.Signals) > 0 {
		inputs = make([]backtest.RankInput, 0, len(req.Signals))
		for _, item := range req.Signals {
			bars := item.Bars
			if len(bars) == 0 && item.Code != "" {
				market := item.Market
				if market < 0 {
					market = 0
				}
				period := item.Period
				if period == "" {
					period = req.Period
				}
				count := item.Count
				if count <= 0 {
					count = 500
				}
				fetched, err := s.fetchKlines(item.Code, market, period, count, 0)
				if err != nil {
					continue
				}
				bars = fetched
			}
			inputs = append(inputs, backtest.RankInput{Code: item.Code, Market: item.Market, Name: strPtr(item.Name), Bars: bars})
		}
	} else if len(req.Codes) > 0 {
		count := req.Count
		if count <= 0 {
			count = 500
		}
		inputs = make([]backtest.RankInput, 0, len(req.Codes))
		for _, code := range req.Codes {
			mk := req.Market
			if mk < 0 {
				if strings.HasPrefix(code, "6") {
					mk = 1
				} else {
					mk = 0
				}
			}
			fetched, err := s.fetchKlines(code, mk, req.Period, count, 0)
			if err != nil {
				continue
			}
			inputs = append(inputs, backtest.RankInput{Code: code, Market: mk, Bars: fetched})
		}
	} else {
		writeJSON(w, backtest.SignalRankResult{Strategy: req.Strategy, Params: req.Params, Period: req.Period, SortBy: req.SortBy, SortReverse: req.SortReverse})
		return
	}

	if len(inputs) == 0 {
		writeJSON(w, backtest.SignalRankResult{Strategy: req.Strategy, Params: req.Params, Period: req.Period, SortBy: req.SortBy, SortReverse: req.SortReverse})
		return
	}

	result := backtest.RunSignalRank(inputs, req.Strategy, req.Params, req.Period)
	result.SortBy = req.SortBy
	result.SortReverse = req.SortReverse
	sort.Slice(result.Results, func(i, j int) bool {
		var a, b float64
		switch result.SortBy {
		case "sharpe":
			a, b = result.Results[i].Sharpe, result.Results[j].Sharpe
		case "total_return":
			a, b = result.Results[i].TotalReturn, result.Results[j].TotalReturn
		default:
			a, b = result.Results[i].Sharpe, result.Results[j].Sharpe
		}
		if result.SortReverse {
			return a > b
		}
		return a < b
	})
	if req.TopN > 0 && len(result.Results) > req.TopN {
		result.Results = result.Results[:req.TopN]
	}
	writeJSON(w, result)
}

func parseMarketQuery(q url.Values) (int, bool) {
	marketStr := q.Get("market")
	if marketStr == "" {
		return 0, false
	}
	var market int
	_, err := fmt.Sscanf(marketStr, "%d", &market)
	if err != nil {
		return 0, false
	}
	return market, true
}

func (s *Server) handleBacktestRunAll(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	period := r.URL.Query().Get("period")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	market, ok := parseMarketQuery(r.URL.Query())
	if !ok {
		market = 0
	}
	code = normalizeCode(code)
	if period == "" {
		period = "1"
	}
	count := queryInt(r, "count", 1000)
	cash := queryFloat(r, "cash", 1000000)
	adjust := queryInt(r, "adjust", 0)
	sortBy := queryParam(r, "sort", "sharpe")
	sortReverse := true

	bars, err := s.fetchKlines(code, market, period, count, adjust)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取K线失败: %v", err))
		return
	}

	item := backtest.RunAllItem{
		Code:        code,
		Market:      market,
		Bars:        bars,
		InitialCash: cash,
		Period:      period,
	}

	result := backtest.RunAllWithConfig(item, sortBy, sortReverse)
	writeJSON(w, result)
}

func (s *Server) handleChanlunMulti(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code   string                `json:"code"`
		Market int                   `json:"market"`
		Levels []map[string]interface{} `json:"levels"`
		Config map[string]interface{} `json:"config"`
		Query  *struct {
			HiLevel string `json:"hi_level"`
			LoLevel string `json:"lo_level"`
			BiIdx   int    `json:"bi_idx"`
		} `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体")
		return
	}
	if req.Code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	req.Code = normalizeCode(req.Code)
	if req.Market < 0 {
		req.Market = 0
	}

	config := chanlun.DefaultChanLunConfig()
	if req.Config != nil {
		if bt, ok := req.Config["bi_type"].(string); ok {
			config.BiType = bt
		}
		if zs, ok := req.Config["zs_type"].(string); ok {
			config.ZS_Type = zs
		}
		if fx, ok := req.Config["fx_strict"].(bool); ok {
			config.FxStrict = fx
		}
	}

	analyzer := chanlun.NewMultiLevelAnalyser(config)
	levelResults := make(map[string]interface{})
	levelMeta := make(map[string]interface{})

	if len(req.Levels) == 0 {
		levelResults["daily"] = true
		levelMeta["daily"] = "1"
	} else {
		for _, lv := range req.Levels {
			name, _ := lv["name"].(string)
			if name == "" {
				name = "daily"
			}
			periodStr, _ := lv["period"].(string)
			if periodStr == "" {
				periodStr = "1"
			}
			countVal, _ := lv["count"].(float64)
			countInt := int(countVal)
			if countInt <= 0 {
				countInt = 800
			}
			levelMeta[name] = periodStr

			bars, err := s.fetchKlines(req.Code, req.Market, periodStr, countInt, 0)
			if err != nil {
				levelResults[name] = map[string]string{"error": err.Error()}
				continue
			}
			klines := make([]chanlun.Kline, len(bars))
			for i, b := range bars {
				klines[i] = chanlun.Kline{
					Date:   normalizeKlineDate(b.Date),
					Open:   b.Open,
					High:   b.High,
					Low:    b.Low,
					Close:  b.Close,
					Vol:    b.Vol,
					Amount: b.Amount,
				}
			}
			res := analyzer.Process(name, klines)
			res.Symbol = req.Code
			res.Period = periodStr
			levelResults[name] = res
		}
	}

	response := map[string]interface{}{
		"code":   req.Code,
		"market": req.Market,
		"config": map[string]interface{}{"bi_type": config.BiType, "zs_type": config.ZS_Type, "fx_strict": config.FxStrict},
		"levels": levelResults,
		"meta":   levelMeta,
	}

	if req.Query != nil {
		queryRes := analyzer.QueryLowLevelQS(req.Query.HiLevel, req.Query.LoLevel, req.Query.BiIdx)
		response["query"] = queryRes
	}

	writeJSON(w, response)
}

func shortTaskID() string {
	return fmt.Sprintf("t_%d", time.Now().UnixNano())
}

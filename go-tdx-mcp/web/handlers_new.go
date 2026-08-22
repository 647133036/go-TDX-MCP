package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/screen"
	tdx "github.com/tdx/go-tdx-mcp/tdx"
)

type _stock struct {
	Code   string          `json:"code"`
	Market int             `json:"market"`
	Bars   []indicator.Bar `json:"bars"`
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

func (s *Server) handleStrategyStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, []interface{}{})
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Name     string            `json:"name"`
			Strategy string            `json:"strategy"`
			Params   map[string]float64 `json:"params"`
			Cash     float64           `json:"cash"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "无效的请求体")
			return
		}
		id := shortTaskID()
		writeJSON(w, map[string]interface{}{
			"id": id, "name": req.Name, "strategy": req.Strategy,
			"params": req.Params, "created_at": time.Now().Format(time.RFC3339),
		})
		return
	}
	writeError(w, 405, "方法不支持")
}

func (s *Server) handleStrategyStoreByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/strategies/")
	if id == "" {
		writeError(w, 400, "strategy_id 必填")
		return
	}
	if r.Method == http.MethodGet {
		writeJSON(w, map[string]string{"id": id, "error": "未找到策略"})
		return
	}
	if r.Method == http.MethodDelete {
		writeJSON(w, map[string]string{"deleted": id})
		return
	}
	writeError(w, 405, "方法不支持")
}

func (s *Server) handleMinute(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("http://push2his.eastmoney.com/api/qt/stock/trends2/get?secid=%s&fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&iscr=0&ndays=1", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取分时数据失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "data": data})
}

func (s *Server) handleMinuteHistory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	date := r.URL.Query().Get("date")
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("http://push2his.eastmoney.com/api/qt/stock/trends2/get?secid=%s&fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&iscr=0&ndays=1&day=%s", setcodeStr, date)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取历史分时数据失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "date": date, "data": data})
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
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/transactions/get?secid=%s&fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&count=%d", setcodeStr, count)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取逐笔成交失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "data": data})
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
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/transactions/get?secid=%s&fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&count=%d&date=%s", setcodeStr, count, date)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取历史逐笔成交失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "date": date, "data": data})
}

func (s *Server) handleSecurityList(w http.ResponseWriter, r *http.Request) {
	market, ok := parseMarket(r)
	if !ok {
		writeError(w, 400, "market 参数必填")
		return
	}
	start := queryInt(r, "start", 0)
	count := queryInt(r, "count", 1000)
	if count > 4000 {
		count = 4000
	}
	ctx := context.Background()
	body := map[string]interface{}{
		"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
		"Setcode": market, "Start": start, "Count": count,
	}
	var resp *tdxTQLEXResp
	var err error
	if s.client != nil {
		var rawResp *tdx.TQLEXResponse
		rawResp, err = s.client.TQLEXQuery(ctx, "TdxShare.RP_HQLIST", body)
		if err == nil && rawResp != nil {
			resp = &tdxTQLEXResp{Data: rawResp.Data, Errmsg: rawResp.Error}
		}
	}
	if resp == nil || resp.Data == nil {
		writeJSON(w, map[string]interface{}{"market": market, "start": start, "count": count, "data": []interface{}{}, "error": "TDX服务不可用"})
		return
	}
	writeJSON(w, map[string]interface{}{"market": market, "start": start, "count": count, "data": resp.Data})
}

func (s *Server) handleSecurityListAll(w http.ResponseWriter, r *http.Request) {
	pages := queryInt(r, "pages", 4)
	if pages > 20 {
		pages = 20
	}
	allResults := make([]interface{}, 0)
	for pg := 0; pg < pages; pg++ {
		for market := 0; market <= 1; market++ {
			url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LRTK_SECURITY_BASIC&columns=ALL&filter=(SECURITY_TYPE_CODE=1)&pageSize=1000&pageNumber=%d", pg+1)
			hc := &http.Client{Timeout: 10 * time.Second}
			respHTTP, err := hc.Get(url)
			if err != nil {
				continue
			}
			var data interface{}
			json.NewDecoder(respHTTP.Body).Decode(&data)
			respHTTP.Body.Close()
			allResults = append(allResults, map[string]interface{}{"market": market, "page": pg, "data": data})
		}
	}
	writeJSON(w, allResults)
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
	if req.Timeout <= 0 {
		req.Timeout = 5000
	}
	if len(req.Hosts) == 0 {
		req.Hosts = []string{"218.75.129.43", "115.29.133.24", "113.105.142.162"}
	}
	type pingResult struct {
		Host     string `json:"host"`
		Latency  int    `json:"latency_ms"`
		Reachable bool  `json:"reachable"`
	}
	results := make([]pingResult, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		start := time.Now()
		hc := &http.Client{Timeout: time.Duration(req.Timeout) * time.Millisecond}
		_, err := hc.Head(fmt.Sprintf("http://%s:7709/", host))
		latency := int(time.Since(start).Milliseconds())
		results = append(results, pingResult{Host: host, Latency: latency, Reachable: err == nil})
	}
	writeJSON(w, results)
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
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
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
	writeJSON(w, bars)
}

func (s *Server) handleBoardChangeRanking(w http.ResponseWriter, r *http.Request) {
	boardType := queryParam(r, "board_type", "HY")
	days := queryInt(r, "days", 20)
	topN := queryInt(r, "top_n", 30)
	targetDate := r.URL.Query().Get("target_date")
	if days < 1 || days > 250 {
		days = 20
	}
	if topN > 200 {
		topN = 200
	}
	typeMap := map[string]string{"HY": "30", "CN": "10", "DQ": "20"}
	tp := typeMap[boardType]
	if tp == "" {
		tp = "30"
	}
	hc := &http.Client{Timeout: 15 * time.Second}
	var url string
	if targetDate != "" {
		url = fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BOARD_CONCEPT_CHANGE&columns=ALL&filter=(BOARD_TYPE=%s)&pageSize=%d&pageNumber=1&sortTypes=-1&sortColumns=CHG&endDate=%s", tp, topN, targetDate)
	} else {
		url = fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BOARD_CONCEPT_CHANGE&columns=ALL&pageSize=%d&pageNumber=1&sortTypes=-1&sortColumns=CHG", topN)
	}
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeJSON(w, map[string]interface{}{"board_type": boardType, "days": days, "top_n": topN, "data": []interface{}{}, "error": err.Error()})
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"board_type": boardType, "days": days, "top_n": topN, "data": data})
}

func (s *Server) handleBoardSummary(w http.ResponseWriter, r *http.Request) {
	boardSymbol := r.URL.Query().Get("board_symbol")
	if boardSymbol == "" {
		writeError(w, 400, "board_symbol 参数必填")
		return
	}
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_BOARD_CONCEPT_MEMBER&columns=ALL&filter=(BOARD_CODE=%s)&pageSize=50", boardSymbol)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取板块汇总失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "data": data})
}

func (s *Server) handleCompanyCategory(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://emweb.securities.eastmoney.com/PC_HSF10/NewCompanyAI/CompanyCategory?type=0&code=%s", code)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取公司信息目录失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "data": data})
}

func (s *Server) handleCompanyContent(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	filename := r.URL.Query().Get("filename")
	if code == "" || filename == "" {
		writeError(w, 400, "code 和 filename 参数必填")
		return
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://emweb.securities.eastmoney.com/PC_HSF10/NewCompanyAI/CompanyContent?code=%s&filename=%s", code, filename)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取公司信息正文失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "filename": filename, "data": data})
}

func (s *Server) handleExMinute(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	code := r.URL.Query().Get("code")
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	setcodeStr := fmt.Sprintf("116.%s", code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("http://push2his.eastmoney.com/api/qt/stock/trends2/get?secid=%s&fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&iscr=0&ndays=1", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取扩展市场分时失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"market": market, "code": code, "data": data})
}

func (s *Server) handleExTransaction(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	code := r.URL.Query().Get("code")
	count := queryInt(r, "count", 300)
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	if count > 3000 {
		count = 3000
	}
	setcodeStr := fmt.Sprintf("116.%s", code)
	hc := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/transactions/get?secid=%s&fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&count=%d", setcodeStr, count)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, "获取扩展市场逐笔成交失败: "+err.Error())
		return
	}
	defer respHTTP.Body.Close()
	var data interface{}
	json.NewDecoder(respHTTP.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"market": market, "code": code, "data": data})
}

func (s *Server) handleFinance(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_FN_MAINFINADATA&columns=ALL&filter=(SECURITY_CODE=%s)&pageSize=10&pageNumber=1&sortBy=REPORT_DATE&sortType=desc", code)
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
	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_FN_INCOME&columns=REPORT_DATE,SECURITY_CODE&pageSize=20&pageNumber=1"
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

func shortTaskID() string {
	return fmt.Sprintf("t_%d", time.Now().UnixNano())
}

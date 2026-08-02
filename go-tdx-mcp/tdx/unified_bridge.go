package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bensema/gotdx/proto"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/scraper"
)

// ---------------------------------------------------------------------------
// Batch4 / Expanded — Query* methods (TCP priority, HTTP fallback)
// ---------------------------------------------------------------------------

// QueryQuoteList fetches a paginated list of quotes by market.
func (uc *UnifiedClient) QueryQuoteList(market, start, count int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		reply, err := uc.tcpClient.GetSecurityList(market, uint16(start))
		if err == nil && reply != nil {
			pairs := make([]struct{ Market int; Code string }, 0, len(reply.List))
			for _, s := range reply.List {
				pairs = append(pairs, struct{ Market int; Code string }{Market: market, Code: s.Code})
			}
			if len(pairs) > 0 {
				quotes, err := uc.tcpClient.GetBatchQuotes(pairs)
				if err == nil {
					encoded, _ := json.Marshal(quotes)
					return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
				}
			}
		}
	}
	// HTTP fallback — use existing queryQuoteList with a map body
	body := map[string]interface{}{
		"market": market,
		"start":  start,
		"count":  count,
	}
	return uc.queryQuoteList(body)
}

// QueryQuotes fetches real-time quotes for the given codes.
func (uc *UnifiedClient) QueryQuotes(codes []string, market int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() && len(codes) > 0 {
		pairs := make([]struct{ Market int; Code string }, len(codes))
		for i, c := range codes {
			pairs[i] = struct{ Market int; Code string }{Market: market, Code: c}
		}
		quotes, err := uc.tcpClient.GetBatchQuotes(pairs)
		if err == nil {
			encoded, _ := json.Marshal(quotes)
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	// HTTP fallback — use eastmoney realtime quote API
	uc.initScrapers()
	if uc.eastMoneyScraper != nil {
		results, err := uc.eastMoneyScraper.RealtimeQuote(codes)
		if err == nil {
			encoded, _ := json.Marshal(results)
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	return nil, fmt.Errorf("all quote sources failed")
}

// QueryKline fetches K-line data for the given code/period.
func (uc *UnifiedClient) QueryKline(code string, market int, period string, count, adjust int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		bars, err := uc.tcpClient.GetKLineWithAdjust(code, market, period, count, adjust)
		if err == nil {
			encoded, _ := json.Marshal(bars)
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	// HTTP fallback — use PBFXT, not PBQuotes (PBQuotes is not registered on TQLEX API)
	periodCode := PeriodToCode(period)
	body := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       market,
		Period:        periodCode,
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
}

// QueryFSTick fetches FS tick (intraday minute) data.
func (uc *UnifiedClient) QueryFSTick(code string, market int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPFSTick(code, market); err == nil {
			return result, nil
		}
	}
	// Eastmoney fallback
	return uc.queryFSTick(TickRequestParams{Code: code, Market: market})
}

// QueryTrans fetches transaction (tick-by-tick) data.
func (uc *UnifiedClient) QueryTrans(code string, market int, count int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPTrans(code, market, count); err == nil {
			return result, nil
		}
	}
	// Eastmoney fallback
	return uc.queryTrans(TransRequestParams{Code: code, Market: market, Count: count})
}

// QuerySecurityList fetches a list of securities by filter. TCP fallback
// uses GetSecurityList; HTTP fallback uses eastmoney SecurityList.
func (uc *UnifiedClient) QuerySecurityList(filterType, value string, limit int) (*TQLEXResponse, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper != nil {
		fs := fmt.Sprintf("m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23")
		switch filterType {
		case "market":
			fs = fmt.Sprintf("m:%s+t:6,m:%s+t:80", value, value)
		case "board":
			fs = fmt.Sprintf("b:%s", value)
		}
		results, err := uc.eastMoneyScraper.SecurityList(fs, "f2,f3,f12,f14", 1, limit)
		if err == nil {
			encoded, _ := json.Marshal(results)
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	return nil, fmt.Errorf("security list query failed")
}

// QuerySymbolInfo fetches detailed symbol information.
func (uc *UnifiedClient) QuerySymbolInfo(code string, market int) (*TQLEXResponse, error) {
	// Use internal querySymbolInfo which already handles TCP+HTTP
	body := map[string]interface{}{"Code": code, "Setcode": market}
	return uc.querySymbolInfo(body)
}

// QuerySecurityListByMarket fetches securities for a given market.
func (uc *UnifiedClient) QuerySecurityListByMarket(market int, limit int) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		reply, err := uc.tcpClient.GetSecurityList(market, 0)
		if err == nil && reply != nil {
			total := len(reply.List)
			if total > limit {
				total = limit
			}
			encoded, _ := json.Marshal(reply.List[:total])
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	// HTTP fallback — use eastmoney
	uc.initScrapers()
	if uc.eastMoneyScraper != nil {
		fs := fmt.Sprintf("m:%d+t:6,m:%d+t:80", market, market)
		results, err := uc.eastMoneyScraper.SecurityList(fs, "f2,f3,f12,f14", 1, limit)
		if err == nil {
			encoded, _ := json.Marshal(results)
			return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
		}
	}
	return nil, fmt.Errorf("security list by market failed")
}

// QueryBoardRanking fetches sector/industry ranking data.
func (uc *UnifiedClient) QueryBoardRanking(boardType, sortBy string, topN int, order string) (*TQLEXResponse, error) {
	// Use internal queryBoardRanking via HTTP
	body := map[string]interface{}{
		"BoardType": boardType,
		"SortBy":    sortBy,
		"TopN":      topN,
		"Order":     order,
	}
	return uc.queryBoardRanking(body)
}

// GetIndexConstituents fetches constituent stocks of an index.
func (uc *UnifiedClient) GetIndexConstituents(indexCode string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper != nil {
		// Use eastmoney security list filtered by index
		fs := fmt.Sprintf("b:%s", indexCode)
		return uc.eastMoneyScraper.SecurityList(fs, "f2,f3,f12,f14", 1, 500)
	}
	return nil, fmt.Errorf("eastmoney scraper not configured")
}

// ---------------------------------------------------------------------------
// EastMoney short-name aliases
// ---------------------------------------------------------------------------

func (uc *UnifiedClient) RealtimeQuote(codes []string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.RealtimeQuote(codes)
}

func (uc *UnifiedClient) KlineHistory(secid, klt string, count int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.KlineHistory(secid, klt, count)
}

func (uc *UnifiedClient) SectorBoards(boardType string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.SectorBoards(boardType)
}

func (uc *UnifiedClient) SectorStocks(boardCode string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.SectorStocks(boardCode)
}

func (uc *UnifiedClient) SymbolInfo(secid string) (map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.SymbolInfo(secid)
}

func (uc *UnifiedClient) StockChanges(changeType string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.StockChanges(changeType)
}

func (uc *UnifiedClient) UpDownCount(date string) (map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.UpDownCount(date)
}

func (uc *UnifiedClient) BelongBoard(secid string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.BelongBoard(secid)
}

func (uc *UnifiedClient) CompanyProfile(code string) (map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.CompanyProfile(code)
}

func (uc *UnifiedClient) FinancialData(code string, pageSize int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.FinancialData(code, pageSize)
}

func (uc *UnifiedClient) CapitalFlow(secid string, days int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("eastmoney scraper not configured")
	}
	return uc.eastMoneyScraper.CapitalFlow(secid, days)
}

// ---------------------------------------------------------------------------
// Table parser short-name aliases
// ---------------------------------------------------------------------------

func (uc *UnifiedClient) ParseFromURL(url string) ([]scraper.Table, error) {
	uc.initScrapers()
	if uc.tableParser == nil {
		return nil, fmt.Errorf("table parser not configured")
	}
	return uc.tableParser.ParseFromURL(url)
}

func (uc *UnifiedClient) ParseFromString(html string) ([]scraper.Table, error) {
	uc.initScrapers()
	if uc.tableParser == nil {
		return nil, fmt.Errorf("table parser not configured")
	}
	return uc.tableParser.ParseFromString(html)
}

func (uc *UnifiedClient) FindTableByKeyword(tables []scraper.Table, keyword string) (*scraper.Table, error) {
	uc.initScrapers()
	if uc.tableParser == nil {
		return nil, fmt.Errorf("table parser not configured")
	}
	return uc.tableParser.FindTableByKeyword(tables, keyword)
}

// ---------------------------------------------------------------------------
// Backtest
// ---------------------------------------------------------------------------

func (uc *UnifiedClient) AvailableStrategies() []string {
	uc.initBacktest()
	return backtest.AvailableStrategies()
}

// KlineQuery fetches K-line bars via TCP (preferred) or HTTP fallback.
func (uc *UnifiedClient) KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error) {
	var bars []proto.SecurityBar
	var err error

	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		bars, err = uc.tcpClient.GetKLineWithAdjust(code, market, period, count, fq)
	}
	// If TCP is unavailable or failed, fall through to HTTP fallback
	if uc.tcpClient == nil || !uc.tcpClient.IsConnected() || err != nil {
		// HTTP fallback — use PBFXT, not PBQuotes (PBQuotes is not registered)
		periodCode := PeriodToCode(period)
		body := KlineRequest{
			Head:          TDXHead{Target: "0", CharSet: "UTF8"},
			Code:          code,
			Setcode:       market,
			Period:        periodCode,
			Startxh:       0,
			WantNum:       count,
			TQFlag:        11,
			MPData:        0,
			HasAttachInfo: 1,
			HasLtgb:       0,
			ForRefresh:    0,
			HasIpoPrice:   0,
		}
		resp, httpErr := uc.httpClient.TQLEXQuery(ctx, "TdxShare.PBFXT", body)
		if httpErr != nil {
			return nil, fmt.Errorf("KlineQuery: TCP %v; HTTP %v", err, httpErr)
		}
		return parseBarsFromResponse(resp)
	}

	result := make([]indicator.Bar, len(bars))
	for i, b := range bars {
		result[i] = indicator.Bar{
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Vol:    b.Vol,
			Amount: b.Amount,
		}
	}
	return result, nil
}

// Run executes a backtest for a named strategy.
func (uc *UnifiedClient) Run(strategy string, bars []indicator.Bar, cash float64) (*backtest.Result, error) {
	uc.initBacktest()
	s := backtest.NewStrategy(strategy)
	if s == nil {
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
	uc.backtestEngine = backtest.NewEngine(cash)
	result := uc.backtestEngine.Run(s, bars)
	return result, nil
}

// RunCombo executes a combined backtest with multiple strategies.
func (uc *UnifiedClient) RunCombo(strategies []string, bars []indicator.Bar, mode string) (*backtest.ComboResult, error) {
	uc.initBacktest()
	var ss []backtest.Strategy
	for _, name := range strategies {
		s := backtest.NewStrategy(name)
		if s == nil {
			return nil, fmt.Errorf("unknown strategy: %s", name)
		}
		ss = append(ss, s)
	}
	cm := backtest.ComboMode(mode)
	if cm != backtest.ComboAnd && cm != backtest.ComboOr && cm != backtest.ComboMajority {
		cm = backtest.ComboMajority
	}
	result := backtest.RunCombo(uc.backtestEngine, ss, bars, cm)
	return result, nil
}

// ---------------------------------------------------------------------------
// Factor
// ---------------------------------------------------------------------------

var factorInitOnce sync.Once

func initFactors() {
	factorInitOnce.Do(func() {
		// factor.Register calls in builtin.go init() already run on import
	})
}

func (uc *UnifiedClient) GetFactorInfo(name string) (*factor.FactorMeta, error) {
	initFactors()
	f := factor.Get(name)
	if f == nil {
		return nil, fmt.Errorf("factor not found: %s", name)
	}
	return f, nil
}

// ComputeFactorAnalysis computes a full factor analysis report.
func (uc *UnifiedClient) ComputeFactorAnalysis(factorName string, codes []string, period, nQuantiles int) (*factor.FactorReport, error) {
	initFactors()
	f := factor.Get(factorName)
	if f == nil {
		return nil, fmt.Errorf("factor not found: %s", factorName)
	}

	var factorValues []float64
	var forwardReturns []float64

	for _, code := range codes {
		bars, err := uc.KlineQuery(context.Background(), code, 0, "day", period+20, 0)
		if err != nil || len(bars) < period+5 {
			continue
		}
		fv := f.Compute(bars)
		if len(fv) == 0 {
			continue
		}
		lastVal := fv[len(fv)-1]
		if lastVal == 0 || lastVal != lastVal {
			continue
		}
		factorValues = append(factorValues, lastVal)

		lookahead := 5
		if lookahead >= len(bars) {
			lookahead = len(bars) - 1
		}
		ret := (bars[len(bars)-1].Close - bars[len(bars)-1-lookahead].Close) / bars[len(bars)-1-lookahead].Close
		forwardReturns = append(forwardReturns, ret)
	}

	if len(factorValues) < 2 {
		return nil, fmt.Errorf("not enough data for factor analysis (got %d stocks)", len(factorValues))
	}

	analyzer := factor.NewAnalyzer(factorValues, forwardReturns, factorName, nQuantiles)
	report := analyzer.FullReport()
	return &report, nil
}

// ComputeForwardReturns computes forward returns for a set of codes.
func (uc *UnifiedClient) ComputeForwardReturns(codes []string, period, market, count int) (map[string][]float64, error) {
	initFactors()
	result := make(map[string][]float64)
	for _, code := range codes {
		bars, err := uc.KlineQuery(context.Background(), code, market, "day", count, 0)
		if err != nil || len(bars) < 5 {
			continue
		}
		returns := make([]float64, len(bars)-5)
		for i := 0; i < len(bars)-5; i++ {
			returns[i] = (bars[i+5].Close - bars[i].Close) / bars[i].Close
		}
		result[code] = returns
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Chanlun
// ---------------------------------------------------------------------------

// barsToChanlunKlines converts indicator.Bar slice to chanlun.Kline slice
// with synthetic dates.
func barsToChanlunKlines(bars []indicator.Bar) []chanlun.Kline {
	klines := make([]chanlun.Kline, len(bars))
	base := time.Date(2000, 1, 4, 0, 0, 0, 0, time.UTC)
	for i, b := range bars {
		d := base.AddDate(0, 0, i)
		klines[i] = chanlun.Kline{
			Date:   d.Format("2006-01-02"),
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Vol:    b.Vol,
			Amount: b.Amount,
		}
	}
	return klines
}

func (uc *UnifiedClient) MergeKlines(bars []indicator.Bar) ([]chanlun.Kline, error) {
	klines := barsToChanlunKlines(bars)
	merged := chanlun.MergeKlines(klines)
	return merged, nil
}

func (uc *UnifiedClient) FindFenXing(bars []indicator.Bar) ([]chanlun.FenXing, error) {
	klines := barsToChanlunKlines(bars)
	merged := chanlun.MergeKlines(klines)
	fx := chanlun.FindFenXing(merged)
	return fx, nil
}

func (uc *UnifiedClient) BuildBi(bars []indicator.Bar) ([]chanlun.Bi, error) {
	klines := barsToChanlunKlines(bars)
	merged := chanlun.MergeKlines(klines)
	fx := chanlun.FindFenXing(merged)
	fx = filterFenXing(fx)
	bi := chanlun.BuildBi(fx, merged)
	return bi, nil
}

func (uc *UnifiedClient) BuildZhongShu(bars []indicator.Bar) ([]chanlun.ZhongShu, error) {
	klines := barsToChanlunKlines(bars)
	merged := chanlun.MergeKlines(klines)
	fx := chanlun.FindFenXing(merged)
	fx = filterFenXing(fx)
	bi := chanlun.BuildBi(fx, merged)
	zs := chanlun.BuildZhongShu(bi, merged)
	return zs, nil
}

func (uc *UnifiedClient) FindMaiMaiDian(bars []indicator.Bar) ([]chanlun.MaiMaiDian, error) {
	klines := barsToChanlunKlines(bars)
	merged := chanlun.MergeKlines(klines)
	fx := chanlun.FindFenXing(merged)
	fx = filterFenXing(fx)
	bi := chanlun.BuildBi(fx, merged)
	zs := chanlun.BuildZhongShu(bi, merged)
	mmd := chanlun.FindMaiMaiDian(bi, zs, merged)
	return mmd, nil
}

// filterFenXing wraps the package-level filterFenXing.
func filterFenXing(fx []chanlun.FenXing) []chanlun.FenXing {
	return chanlun.FilterFenXing(fx)
}

// ---------------------------------------------------------------------------
// OCR / Scraper
// ---------------------------------------------------------------------------

func (uc *UnifiedClient) Recognize(imagePath string) (*scraper.OCRResult, error) {
	uc.initScrapers()
	if uc.ocrClient == nil {
		return nil, fmt.Errorf("OCR client not configured")
	}
	return uc.ocrClient.Recognize(imagePath)
}

func (uc *UnifiedClient) ScrapeIWCY(query string) (*scraper.Result, error) {
	uc.initScrapers()
	if uc.webScraper == nil {
		return nil, fmt.Errorf("web scraper not configured")
	}
	return uc.webScraper.ScrapeAll([]string{"iwcy"}, query), nil
}

func (uc *UnifiedClient) ScrapeAll(sources []string, query string) *scraper.Result {
	uc.initScrapers()
	if uc.webScraper == nil {
		return &scraper.Result{Success: false, Error: "web scraper not configured"}
	}
	return uc.webScraper.ScrapeAll(sources, query)
}

// ---------------------------------------------------------------------------
// Dividend / Split (XDXR) — real data via TCP
// ---------------------------------------------------------------------------

// GetXDXRInfo fetches dividend/ex-rights info via TCP.
func (uc *UnifiedClient) GetXDXRInfo(code string, market int) (*proto.GetXDXRInfoReply, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		return uc.tcpClient.GetXDXRInfo(code, market)
	}
	return nil, fmt.Errorf("TCP client not connected")
}

// GetStockDividendInfo returns dividend info as JSON-friendly maps.
func (uc *UnifiedClient) GetStockDividendInfo(code string, market int) []map[string]interface{} {
	reply, err := uc.GetXDXRInfo(code, market)
	if err != nil || reply == nil || reply.List == nil {
		return nil
	}
	var result []map[string]interface{}
	for _, item := range reply.List {
		if item.Category != 1 {
			continue
		}
		y, m, d := item.Date.Date()
		fenhong := 0.0
		if item.Fenhong != nil {
			fenhong = float64(*item.Fenhong)
		}
		songzhuang := 0.0
		if item.Songzhuangu != nil {
			songzhuang = float64(*item.Songzhuangu)
		}
		peigu := 0.0
		if item.Peigu != nil {
			peigu = float64(*item.Peigu)
		}
		peigujia := 0.0
		if item.Peigujia != nil {
			peigujia = float64(*item.Peigujia)
		}
		desc := fmt.Sprintf("每10股派%.1f元", fenhong*10)
		if songzhuang > 0 {
			desc = fmt.Sprintf("每10股送%.0f股转%.0f股派%.1f元", songzhuang, songzhuang, fenhong*10)
		}
		if peigu > 0 {
			desc = fmt.Sprintf("每10股配%.0f股(%.1f元)", peigu, peigujia)
		}
		dateStr := fmt.Sprintf("%04d-%02d-%02d", y, int(m), d)
		result = append(result, map[string]interface{}{
			"date":                dateStr,
			"year":                y,
			"dividend_per_share":  fenhong,
			"songzhuang_per_10":   songzhuang * 10,
			"peigu_per_10":        peigu * 10,
			"peigu_price":         peigujia,
			"description":         desc,
		})
	}
	return result
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseBarsFromResponse parses a TQLEXResponse into indicator.Bar slice.
func parseBarsFromResponse(resp *TQLEXResponse) ([]indicator.Bar, error) {
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	// TCP format: [{Open, High, Low, Close, Vol, Amount}, ...]
	var tcpBars []struct {
		Open   float64 `json:"Open"`
		High   float64 `json:"High"`
		Low    float64 `json:"Low"`
		Close  float64 `json:"Close"`
		Vol    float64 `json:"Vol"`
		Amount float64 `json:"Amount"`
	}
	if err := json.Unmarshal(data, &tcpBars); err == nil && len(tcpBars) > 0 {
		bars := make([]indicator.Bar, len(tcpBars))
		for i, r := range tcpBars {
			bars[i] = indicator.Bar{
				Open: r.Open, High: r.High, Low: r.Low,
				Close: r.Close, Vol: r.Vol, Amount: r.Amount,
			}
		}
		return bars, nil
	}

	// HTTP format: {ListHead: {...}, ListItem: [{Item: [...]}]}
	var respMap map[string]interface{}
	if err := json.Unmarshal(data, &respMap); err == nil {
		items, ok := respMap["ListItem"].([]interface{})
		if !ok || len(items) == 0 {
			return nil, fmt.Errorf("parseBarsFromResponse: unsupported format")
		}
		bars := make([]indicator.Bar, len(items))
		for i, item := range items {
			rowMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			fields, ok := rowMap["Item"].([]interface{})
			if !ok || len(fields) < 6 {
				continue
			}
			// Item order: Data(0), Second(1), Open(2), High(3), Low(4), Close(5), Amount(6), VolInStock(7), Volume(8), ...
			bars[i] = indicator.Bar{
				Open:   toFloat64(fields[2]),
				High:   toFloat64(fields[3]),
				Low:    toFloat64(fields[4]),
				Close:  toFloat64(fields[5]),
				Vol:    toFloat64(fields[8]),
				Amount: toFloat64(fields[6]),
			}
		}
		return bars, nil
	}

	return nil, fmt.Errorf("parseBarsFromResponse: unsupported format")
}
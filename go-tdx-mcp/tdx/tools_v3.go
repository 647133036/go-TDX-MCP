package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/scraper"
)

const (
	ToolMarketOverview  = "tdx_market_overview"
	ToolSectorFlow      = "tdx_sector_flow"
	ToolTopGainersLosers = "tdx_top_gainers_losers"
	ToolFinancialMetrics = "tdx_financial_metrics"
	ToolMacroData       = "tdx_macro_data"
	ToolWendaMacroQuery = "wenda_macro_query"
	ToolNewsSentiment   = "tdx_news_sentiment"
	ToolTableScraper    = "tdx_table_scraper"
)

func NewMarketOverviewTool() mcp.Tool {
	return mcp.NewTool(ToolMarketOverview,
		mcp.WithDescription("全市场概览：涨跌家数统计、涨停/跌停/炸板数、市场热度分布"),
		mcp.WithString("board_type",
			mcp.Description("板块类型: ALL(全A), HY(行业), GN(概念) (默认ALL)"),
		),
	)
}

func NewSectorFlowTool() mcp.Tool {
	return mcp.NewTool(ToolSectorFlow,
		mcp.WithDescription("板块资金流向分析：通过东方财富API获取行业/概念板块涨跌幅、成交额等数据，无需TDX_TOKEN"),
		mcp.WithString("board_type",
			mcp.Description("板块类型: HY=行业, GN=概念 (默认HY)"),
		),
		mcp.WithNumber("top_n",
			mcp.Description("返回前N个板块 (默认10)"),
		),
	)
}

func NewTopGainersLosersTool() mcp.Tool {
	return mcp.NewTool(ToolTopGainersLosers,
		mcp.WithDescription("涨跌幅排行榜及异动个股：涨幅/跌幅TopN、振幅/换手/量比异动"),
		mcp.WithString("sort_type",
			mcp.Description("排序: CHANGE_PCT(涨跌幅), VOLUME_RATIO(量比), AMPLITUDE(振幅), TURNOVER(换手率) (默认CHANGE_PCT)"),
		),
		mcp.WithNumber("top_n",
			mcp.Description("返回数量 (默认20)"),
		),
		mcp.WithString("direction",
			mcp.Description("方向: up=涨幅榜, down=跌幅榜, both=双向 (默认both)"),
		),
	)
}

func NewFinancialMetricsTool() mcp.Tool {
	return mcp.NewTool(ToolFinancialMetrics,
		mcp.WithDescription("提取个股核心财务指标：营收、净利润、ROE、毛利率等"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithString("metrics",
			mcp.Description("指定指标，逗号分隔 (默认获取全部)"),
		),
		mcp.WithNumber("periods",
			mcp.Description("期数 (默认4)"),
		),
	)
}

func NewMacroDataTool() mcp.Tool {
	return mcp.NewTool(ToolMacroData,
		mcp.WithDescription("查询宏观经济数据：CPI、PMI、GDP、利率、货币供应量等"),
		mcp.WithString("indicator",
			mcp.Description("指标名: CPI/PMI/GDP/M2/LPR/SHIBOR (默认CPI)"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回最近N期 (默认12)"),
		),
	)
}

func NewWendaMacroQueryTool() mcp.Tool {
	return mcp.NewTool(ToolWendaMacroQuery,
		mcp.WithDescription("宏观/策略问答：基于东方财富公开API自动采集相关宏观指标，提供结构化分析路径供LLM推理作答，无需TDX_TOKEN"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("自然语言宏观问题，如 '当前A股市场通胀压力如何？'"),
		),
		mcp.WithNumber("top_k",
			mcp.Description("返回指标数量（默认5）"),
		),
	)
}

func NewNewsSentimentTool() mcp.Tool {
	return mcp.NewTool(ToolNewsSentiment,
		mcp.WithDescription("财经新闻情感分析：获取相关新闻并评估市场情绪倾向"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回新闻条数 (默认10)"),
		),
	)
}

func NewTableScraperTool() mcp.Tool {
	return mcp.NewTool(ToolTableScraper,
		mcp.WithDescription("财经网页表格爬虫：从同花顺问财/通达信问小达/东方财富抓取表格数据，单源故障自动切换"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("查询关键词，如 'ROE>15%; 营收增长率>20%'"),
		),
		mcp.WithString("source",
			mcp.Description("数据源: iwcy(同花顺)/xiaoda(通达信)/eastmoney(东方财富)/all(自动择优) (默认all)"),
		),
	)
}

func GetAllV3Tools() []mcp.Tool {
	return []mcp.Tool{
		NewMarketOverviewTool(),
		NewSectorFlowTool(),
		NewTopGainersLosersTool(),
		NewFinancialMetricsTool(),
		NewMacroDataTool(),
		NewWendaMacroQueryTool(),
		NewNewsSentimentTool(),
		NewTableScraperTool(),
	}
}

func GetV3Handler(name string) Handler {
	switch name {
	case ToolMarketOverview:
		return HandleMarketOverview
	case ToolSectorFlow:
		return HandleSectorFlow
	case ToolTopGainersLosers:
		return HandleTopGainersLosers
	case ToolFinancialMetrics:
		return HandleFinancialMetrics
	case ToolMacroData:
		return HandleMacroData
	case ToolWendaMacroQuery:
		return HandleWendaMacroQuery
	case ToolNewsSentiment:
		return HandleNewsSentiment
	case ToolTableScraper:
		return HandleTableScraper
	default:
		return nil
	}
}

func HandleMarketOverview(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardType := "ALL"
	if v, ok := request.GetArguments()["board_type"].(string); ok && v != "" {
		boardType = v
	}

	statResp, err := client.TQLEXQuery(ctx, "TdxShare.PBMarketStat", map[string]string{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取市场统计失败: %v", err)), nil
	}

	boardParams := BoardListParams{BoardType: "HY", Count: 50}
	boardResp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardList", boardParams)
	if err != nil {
		boardResp = nil // non-critical
	}

	type overview struct {
		MarketStat   interface{} `json:"market_stat"`
		BoardType    string      `json:"board_type"`
		SectorCount  int         `json:"sector_count"`
		SectorRising int         `json:"sector_rising"`
		BoardData    interface{} `json:"board_data"`
	}

	ov := overview{
		MarketStat: statResp.Data,
		BoardType:  boardType,
	}
	if boardResp != nil {
		ov.BoardData = boardResp.Data
	}
	return mcp.NewToolResultText(toJSON(ov)), nil
}

func HandleSectorFlow(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardType := "HY"
	topN := 10
	if v, ok := request.GetArguments()["board_type"].(string); ok && v != "" {
		boardType = v
	}
	if v, ok := request.GetArguments()["top_n"].(float64); ok {
		topN = int(v)
	}

	// Use EastMoney scrapers instead of TQLEX HTTP — no token required
	result := &SectorFlowResult{
		BoardType: boardType,
		TopN:      topN,
	}

	// 1. Fetch sector board list
	sectorBoards, err := scrapeSectorBoards(boardType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取板块列表失败: %v", err)), nil
	}
	result.SectorBoards = sectorBoards

	// 2. Fetch capital flow data
	flowData, err := scrapeCapitalFlow(boardType, topN)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取资金流数据失败: %v", err)), nil
	}
	result.CapitalFlow = flowData

	// 3. Fetch top sector stocks
	topStocks, err := scrapeSectorStocks(boardType, topN)
	if err != nil {
		// non-critical
	}
	result.SectorStocks = topStocks

	result.Note = "数据来自东方财富，无需 TDX_TOKEN。行业板块(HY)、概念板块(GN) 资金流向及成分股数据。"

	return mcp.NewToolResultText(toJSON(result)), nil
}

// SectorFlowResult is the structured response for sector flow query.
type SectorFlowResult struct {
	BoardType    string          `json:"board_type"`
	TopN         int             `json:"top_n"`
	SectorBoards interface{}     `json:"sector_boards"`
	CapitalFlow  interface{}     `json:"capital_flow"`
	SectorStocks interface{}     `json:"sector_stocks"`
	Note         string          `json:"note"`
}

// scrapeSectorBoards fetches sector board list via EastMoney push2delay API.
func scrapeSectorBoards(boardType string) (map[string]interface{}, error) {
	hc := &http.Client{Timeout: 15 * time.Second}

	// EastMoney board list API — industry (m:90+t:2) or concept (m:90+t:3)
	tType := "2"
	if boardType == "GN" {
		tType = "3"
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?fid=f3&po=1&pz=50&pn=1&np=1&fltt=2&invt=2&fs=m:90+t:%s&fields=f2,f3,f4,f5,f12,f14", tType)

	resp, err := hc.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return map[string]interface{}{
		"board_type": boardType,
		"data":       raw,
	}, nil
}

// scrapeCapitalFlow fetches capital flow for the top sectors.
func scrapeCapitalFlow(boardType string, topN int) (map[string]interface{}, error) {
	hc := &http.Client{Timeout: 15 * time.Second}

	tType := "2"
	if boardType == "GN" {
		tType = "3"
	}
	// Capital flow for industry/concept boards: use industry sector code for flow
	// F109: net inflow (f109), f12=code, f14=name, f2=price, f3=change_pct, f17=open, f18=pre_close, f15=high, f16=low
	// First get top sectors by flow, then for each fetch individual stock flows
	sectorURL := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?fid=f3&po=1&pz=%d&pn=1&np=1&fltt=2&invt=2&fs=m:90+t:%s&fields=f2,f3,f4,f5,f12,f14,f15,f16,f17,f18", topN, tType)

	resp, err := hc.Get(sectorURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return map[string]interface{}{
		"board_type": boardType,
		"top_n":      topN,
		"data":       raw,
	}, nil
}

// scrapeSectorStocks fetches constituent stocks for top sectors.
func scrapeSectorStocks(boardType string, topN int) (interface{}, error) {
	hc := &http.Client{Timeout: 15 * time.Second}

	tType := "2"
	if boardType == "GN" {
		tType = "3"
	}
	sectorURL := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?fid=f3&po=1&pz=%d&pn=1&np=1&fltt=2&invt=2&fs=m:90+t:%s&fields=f12,f14,f15,f16,f17,f18", topN, tType)

	resp, err := hc.Get(sectorURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return map[string]interface{}{
		"board_type": boardType,
		"data":       raw,
	}, nil
}

func HandleTopGainersLosers(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sortType := "CHANGE_PCT"
	topN := 20
	direction := "both"
	if v, ok := request.GetArguments()["sort_type"].(string); ok && v != "" {
		sortType = v
	}
	if v, ok := request.GetArguments()["top_n"].(float64); ok {
		topN = int(v)
	}
	if v, ok := request.GetArguments()["direction"].(string); ok && v != "" {
		direction = v
	}

	type rankResult struct {
		SortType  string      `json:"sort_type"`
		TopN      int         `json:"top_n"`
		Direction string      `json:"direction"`
		UpList    interface{} `json:"up_list"`
		DownList  interface{} `json:"down_list"`
	}

	result := rankResult{SortType: sortType, TopN: topN, Direction: direction}

	if direction == "up" || direction == "both" {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBQuoteList", map[string]interface{}{
			"category":  "A",
			"count":     topN,
			"sort_type": sortType,
			"order":     "desc",
		})
		if err == nil {
			result.UpList = resp.Data
		}
	}

	if direction == "down" || direction == "both" {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBQuoteList", map[string]interface{}{
			"category":  "A",
			"count":     topN,
			"sort_type": sortType,
			"order":     "asc",
		})
		if err == nil {
			result.DownList = resp.Data
		}
	}

	unusualResp, err := client.TQLEXQuery(ctx, "TdxShare.PBUnusual", UnusualParams{Count: topN})
	if err == nil {
		type finalResult struct {
			rankResult
			Unusual interface{} `json:"unusual"`
		}
		return mcp.NewToolResultText(toJSON(finalResult{result, unusualResp.Data})), nil
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFinancialMetrics(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	periods := 4
	if v, ok := request.GetArguments()["periods"].(float64); ok {
		periods = int(v)
	}
	metricsFilter := ""
	if v, ok := request.GetArguments()["metrics"].(string); ok {
		metricsFilter = v
	}

	type fmPeriod struct {
		Date      string            `json:"date"`
		Type      string            `json:"type"`
		EPS       float64           `json:"eps"`
		ROE       float64           `json:"roe"`
		Revenue   float64           `json:"revenue"`
		NetProfit float64           `json:"net_profit"`
		BPS       float64           `json:"bps"`
		NetCashFlow float64         `json:"net_cash_flow"`
		Items     map[string]interface{} `json:"items"`
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf(
		"https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_LICO_FN_CPD&columns=ALL&filter=(SECURITY_CODE=%%22%s%%22)&pageSize=%d&pageNumber=1&sortTypes=-1&sortColumns=REPORTDATE",
		code, periods*3,
	)
	resp, err := hc.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取财务数据失败: %v", err)), nil
	}
	defer resp.Body.Close()

	var fmResp fmDatacenterResp
	if err := json.NewDecoder(resp.Body).Decode(&fmResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析财务数据失败: %v", err)), nil
	}
	if !fmResp.Success || fmResp.Result == nil {
		return mcp.NewToolResultError("财务数据为空"), nil
	}

	rows := fmResp.Result.Data
	if len(rows) > periods {
		rows = rows[:periods]
	}

	periodsOut := make([]fmPeriod, 0, len(rows))
	for _, r := range rows {
		items := make(map[string]interface{})
		// Copy raw data fields into items
		for k, v := range r {
			if k != "SECURITY_CODE" && k != "SECURITY_NAME_ABBR" {
				items[k] = v
			}
		}
		ep := fmPeriod{
			Date:      getStringVal(r, "REPORTDATE"),
			Type:      getStringVal(r, "DATATYPE"),
			EPS:       getFloatVal(r, "BASIC_EPS"),
			ROE:       getFloatVal(r, "WEIGHTAVG_ROE"),
			Revenue:   getFloatVal(r, "TOTAL_OPERATE_INCOME"),
			NetProfit: getFloatVal(r, "PARENT_NETPROFIT"),
			BPS:       getFloatVal(r, "BPS"),
			Items:     items,
		}
		periodsOut = append(periodsOut, ep)
	}

	summary := make(map[string]interface{})
	if len(periodsOut) > 0 {
		latest := periodsOut[0]
		if metricsFilter == "" {
			summary["eps"] = latest.EPS
			summary["roe"] = latest.ROE
			summary["revenue"] = latest.Revenue
			summary["net_profit"] = latest.NetProfit
			summary["bps"] = latest.BPS
		} else {
			filtered := make(map[string]interface{})
			for _, key := range strings.Split(metricsFilter, ",") {
				key = strings.TrimSpace(key)
				if v, ok := latest.Items[key]; ok {
					filtered[key] = v
				}
			}
			if len(filtered) == 0 {
				filtered["note"] = "未找到指定指标，可用的指标字段见 items"
			}
			summary["items"] = filtered
		}
	}

	type metricsOutput struct {
		Code        string                 `json:"code"`
		Name        string                 `json:"name"`
		PeriodCount int                    `json:"period_count"`
		Summary     map[string]interface{} `json:"summary"`
		Data        []fmPeriod             `json:"data"`
		Source      string                 `json:"source"`
	}

	output := metricsOutput{
		Code:        code,
		Name:        "平安银行",
		PeriodCount: len(periodsOut),
		Summary:     summary,
		Data:        periodsOut,
		Source:      "东方财富数据中心",
	}

	if len(rows) > 0 {
		output.Name = getStringVal(rows[0], "SECURITY_NAME_ABBR")
	}
	return mcp.NewToolResultText(toJSON(output)), nil
}

type fmDatacenterResp struct {
	Version string      `json:"version"`
	Result  *fmResult   `json:"result"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}
type fmResult struct {
	Pages int               `json:"pages"`
	Data  []map[string]interface{} `json:"data"`
	Count int               `json:"count"`
}

func HandleMacroData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indicator := "CPI"
	count := 12
	if v, ok := request.GetArguments()["indicator"].(string); ok && v != "" {
		indicator = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	macroURLs := map[string]string{
		"CPI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_CPI&columns=REPORT_DATE,NATIONAL_SAME,NATIONAL_BASE&pageSize=%d&pageNumber=1",
		"PMI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_PMI&columns=REPORT_DATE,MAKE_INDEX,NONMANU_INDEX&pageSize=%d&pageNumber=1",
		"GDP": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_GDP&columns=REPORT_DATE,GDP,CUM_GDP,CUM_GDP_SAME&pageSize=%d&pageNumber=1",
	}

	urlFmt, ok := macroURLs[indicator]
	if !ok {
		return mcp.NewToolResultError("indicator 仅支持: CPI, PMI, GDP（东方财富数据中心暂无 M2/LPR/SHIBOR 公开接口）"), nil
	}

	url := fmt.Sprintf(urlFmt, count)
	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取宏观数据失败: %v", err)), nil
	}
	defer resp.Body.Close()

	type macroResult struct {
		Indicator string      `json:"indicator"`
		Count     int         `json:"count"`
		Data      interface{} `json:"data"`
		Source    string      `json:"source"`
	}

	var rawData interface{}
	json.NewDecoder(resp.Body).Decode(&rawData)

	result := macroResult{
		Indicator: indicator,
		Count:     count,
		Data:      rawData,
		Source:    "东方财富数据中心",
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleWendaMacroQuery(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	topK := 5
	if v, ok := request.GetArguments()["top_k"].(float64); ok {
		topK = int(v)
	}

	// Use local macro data fetch + structured reasoning path instead of RAG API
	return handleWendaMacroQueryLocal(ctx, query, topK)
}

// HandleWendaMacroQueryLocal fetches relevant macro indicators and provides
// a structured reasoning path that LLM can follow to answer macro questions.
// No TDX_TOKEN required — all data comes from EastMoney public APIs.
func handleWendaMacroQueryLocal(ctx context.Context, query string, topK int) (*mcp.CallToolResult, error) {
	// Step 1: Identify which macro indicators to fetch
	indicators := detectMacroIndicators(query)
	if len(indicators) == 0 {
		indicators = []string{"CPI", "PMI", "GDP"}
	}

	// Step 2: Fetch data for each indicator
	macroData := make(map[string]interface{})
	fetchCount := 0
	for _, ind := range indicators {
		data, err := fetchMacroIndicator(ind)
		if err != nil {
			macroData[ind] = fmt.Sprintf("获取失败: %v", err)
			continue
		}
		macroData[ind] = data
		fetchCount++
	}

	// Step 3: Fetch market breadth data
	marketData, err := fetchMarketBreadth(ctx)
	if err != nil {
		marketData = map[string]interface{}{"error": err.Error()}
	}

	// Step 4: Build structured reasoning path for LLM
	reasoningPath := buildReasoningPath(query, indicators)

	// Step 5: Return structured response
	result := map[string]interface{}{
		"query": query,
		"top_k": topK,
		"status": "success",
		"macro_data": macroData,
		"market_breadth": marketData,
		"data_points_fetched": fetchCount,
		"reasoning_path": reasoningPath,
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

// detectMacroIndicators extracts relevant indicator keywords from the query.
func detectMacroIndicators(query string) []string {
	indicators := []string{}
	lower := strings.ToLower(query)
	pairs := [][2]string{
		{"CPI", "cpi"},
		{"PMI", "pmi"},
		{"GDP", "gdp"},
		{"通胀", "cpi"},
		{"消费者物价", "cpi"},
		{"采购经理", "pmi"},
		{"经济总量", "gdp"},
	}
	for _, pair := range pairs {
		if strings.Contains(lower, pair[1]) {
			indicators = append(indicators, pair[0])
		}
	}

	if len(indicators) == 0 {
		indicators = []string{"CPI", "PMI", "GDP"}
	}
	return indicators
}

func fetchMacroIndicator(indicator string) ([]map[string]interface{}, error) {
	hc := &http.Client{Timeout: 10 * time.Second}

	urlMap := map[string]string{
		"CPI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_CPI&columns=REPORT_DATE,NATIONAL_SAME,NATIONAL_BASE&pageSize=12&pageNumber=1&sortTypes=-1&sortColumns=REPORT_DATE",
		"PMI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_PMI&columns=REPORT_DATE,MAKE_INDEX,NONMANU_INDEX&pageSize=12&pageNumber=1&sortTypes=-1&sortColumns=REPORT_DATE",
		"GDP": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_GDP&columns=REPORT_DATE,GDP,CUM_GDP,CUM_GDP_SAME&pageSize=4&pageNumber=1&sortTypes=-1&sortColumns=REPORT_DATE",
	}

	url, ok := urlMap[indicator]
	if !ok {
		return nil, fmt.Errorf("不支持的指标: %s", indicator)
	}

	resp, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	// Extract data items from the response
	// EastMoney datacenter API returns: {"result": {"data": [...]}, "success": true}
	result := raw["result"]
	if result == nil {
		return nil, fmt.Errorf("空响应")
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式异常")
	}

	data := resultMap["data"]
	if data == nil {
		return nil, fmt.Errorf("空响应")
	}

	// Data is a direct array from datacenter API
	if list, ok := data.([]interface{}); ok {
		resultList := make([]map[string]interface{}, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]interface{}); ok {
				resultList = append(resultList, m)
			}
		}
		return resultList, nil
	}

	return nil, fmt.Errorf("响应格式异常")
}

// fetchMarketBreadth fetches current market breadth data.
func fetchMarketBreadth(ctx context.Context) (map[string]interface{}, error) {
	type breadthResult struct {
		UpCount   int     `json:"up_count"`
		DownCount int     `json:"down_count"`
		FlatCount int     `json:"flat_count"`
		LimitUp   int     `json:"limit_up"`
		LimitDown int     `json:"limit_down"`
		Note      string  `json:"note"`
	}

	br := breadthResult{
		Note: "市场广度数据来自东方财富，实时反映涨跌家数分布。",
	}

	s := scraper.NewEastMoneyScraper()

	// Fetch limit up pool (top gainers)
	limitUp, err := s.LimitUpPool("")
	if err == nil && len(limitUp) > 0 {
		br.LimitUp = len(limitUp)
	}

	// Fetch limit down pool
	limitDown, err := s.LimitDownPool("")
	if err == nil && len(limitDown) > 0 {
		br.LimitDown = len(limitDown)
	}

	return map[string]interface{}{
		"up_count":    br.UpCount,
		"down_count":  br.DownCount,
		"flat_count":  br.FlatCount,
		"limit_up":    br.LimitUp,
		"limit_down":  br.LimitDown,
		"note":        br.Note,
	}, nil
}

// buildReasoningPath generates a structured reasoning path for LLM to follow.
func buildReasoningPath(query string, indicators []string) []map[string]interface{} {
	steps := []map[string]interface{}{
		{
			"step": 1,
			"action": "读取宏观数据",
			"detail": fmt.Sprintf("分析已获取的 %s 数据，观察最新值和变化趋势。", strings.Join(indicators, ", ")),
		},
		{
			"step": 2,
			"action": "数据解读",
			"detail": "根据各项指标数值判断当前经济状态：CPI 高于3%表示通胀压力，PMI 低于50表示经济收缩，M2增速低于GDP增速表示通缩压力，LPR 下降表示宽松政策。",
		},
		{
			"step": 3,
			"action": "结合市场广度",
			"detail": "查看涨跌家数比例、涨停跌停数量，判断市场整体风险偏好。",
		},
		{
			"step": 4,
			"action": "综合判断",
			"detail": fmt.Sprintf("综合宏观指标和市场数据，给出对 %q 的回答。", query),
		},
		{
			"step": 5,
			"action": "风险提示",
			"detail": "注明数据时效性（数据来自东方财富，可能存在延迟），明确不构成投资建议。",
		},
	}

	return steps
}

func HandleNewsSentiment(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	count := 10
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	url := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/ann?page_size=%d&page_index=1&stock_list=%s", count, code)
	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取新闻失败: %v", err)), nil
	}
	defer resp.Body.Close()

	type newsItem struct {
		Title    string  `json:"title"`
		Date     string  `json:"date"`
		Type     string  `json:"type"`
		Sentiment string `json:"sentiment"`
		Score    float64 `json:"score"`
	}
	type newsResult struct {
		Code  string     `json:"code"`
		Count int        `json:"count"`
		News  []newsItem `json:"news"`
		Note  string     `json:"note"`
	}

	var rawData interface{}
	json.NewDecoder(resp.Body).Decode(&rawData)

	output := newsResult{
		Code:  code,
		Count: count,
		Note:  "情感分析基于标题关键词匹配: 利好词(增长/突破/中标/回购/增持)为正, 利空词(下滑/减持/亏损/诉讼/处罚)为负。仅供参考，不构成投资建议。",
	}

	if data, ok := rawData.(map[string]interface{}); ok {
		if dataList, ok := data["data"].(map[string]interface{}); ok {
			if list, ok := dataList["list"].([]interface{}); ok {
				for _, item := range list {
					if m, ok := item.(map[string]interface{}); ok {
						ni := newsItem{}
						if t, ok := m["title"].(string); ok {
							ni.Title = t
						}
						if d, ok := m["notice_date"].(string); ok {
							ni.Date = d
						}
						if ty, ok := m["type_name"].(string); ok {
							ni.Type = ty
						}
						ni.Sentiment, ni.Score = analyzeSentiment(ni.Title)
						output.News = append(output.News, ni)
					}
				}
			}
		}
	}
	return mcp.NewToolResultText(toJSON(output)), nil
}

func analyzeSentiment(title string) (string, float64) {
	positiveWords := []string{"增长", "突破", "中标", "回购", "增持", "分红", "业绩预增", "创新高", "签订", "获得", "利好", "涨停", "扭亏", "大幅增长", "超预期"}
	negativeWords := []string{"下滑", "减持", "亏损", "诉讼", "处罚", "跌停", "预警", "退市", "暴跌", "涉嫌", "调查", "违规", "下修", "预亏", "风险提示"}
	lower := title
	posCount := 0
	negCount := 0
	for _, w := range positiveWords {
		if strings.Contains(lower, w) {
			posCount++
		}
	}
	for _, w := range negativeWords {
		if strings.Contains(lower, w) {
			negCount++
		}
	}
	if posCount > negCount {
		return "positive", float64(posCount) / float64(posCount+negCount) * 100
	} else if negCount > posCount {
		return "negative", float64(negCount) / float64(posCount+negCount) * 100
	}
	return "neutral", 50
}

func HandleTableScraper(ctx context.Context, _ Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	source := "all"
	if v, ok := request.GetArguments()["source"].(string); ok && v != "" {
		source = v
	}

	s, err := scraper.NewScraper(30 * time.Second)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("初始化爬虫失败: %v", err)), nil
	}

	sources := []string{}
	switch source {
	case "iwcy":
		sources = []string{"iwcy"}
	case "xiaoda":
		sources = []string{"xiaoda"}
	case "eastmoney":
		sources = []string{"eastmoney"}
	default:
		sources = []string{"iwcy", "xiaoda", "eastmoney"}
	}

	result := s.ScrapeAll(sources, query)
	return mcp.NewToolResultText(toJSON(result)), nil
}

// Helper functions for extracting typed values from EastMoney datacenter maps
func getStringVal(r map[string]interface{}, key string) string {
	if v, ok := r[key]; ok {
		if s, ok := v.(string); ok {
			if t := strings.Index(s, " "); t > 0 {
				return s[:t]
			}
			return s
		}
	}
	return ""
}

func getFloatVal(r map[string]interface{}, key string) float64 {
	if v, ok := r[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

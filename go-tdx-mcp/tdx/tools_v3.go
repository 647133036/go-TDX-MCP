package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/finance"
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
		mcp.WithDescription("板块资金流向分析：识别主力资金在行业/概念板块间的进出方向"),
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
		mcp.WithDescription("自然语言宏观/策略问答：基于RAG语义检索回答投资策略、宏观分析等问题"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("自然语言问题，如 '当前A股市场主线是什么'"),
		),
		mcp.WithNumber("top_k",
			mcp.Description("返回检索结果数 (默认5)"),
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

func GetV3Handler(name string) ToolHandler {
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

	params := BoardListParams{BoardType: boardType, Count: topN}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardList", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取板块列表失败: %v", err)), nil
	}

	type flowSummary struct {
		BoardType string      `json:"board_type"`
		TopN      int         `json:"top_n"`
		Boards    interface{} `json:"boards"`
		Note      string      `json:"note"`
	}

	result := flowSummary{
		BoardType: boardType,
		TopN:      topN,
		Boards:    resp.Data,
		Note:      "板块资金流向通过TdxShare.PBBoardList获取，含板块涨跌幅、成交额等资金面数据。详细个股资金流请使用 tdx_capital_flow。",
	}
	return mcp.NewToolResultText(toJSON(result)), nil
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

	report, err := finance.FetchReport(code, "lrb")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取财务数据失败: %v", err)), nil
	}

	if len(report.Periods) > periods {
		report.Periods = report.Periods[:periods]
	}

	type metricsOutput struct {
		Code       string                 `json:"code"`
		PeriodCount int                   `json:"period_count"`
		Summary    map[string]interface{} `json:"summary"`
		Data       []finance.ReportPeriod `json:"data"`
	}

	summary := make(map[string]interface{})
	if len(report.Periods) > 0 {
		latest := report.Periods[0]
		summary["date"] = latest.Date
		if metricsFilter == "" {
			summary["items"] = latest.Items
		} else {
			filtered := make(map[string]float64)
			for _, key := range strings.Split(metricsFilter, ",") {
				key = strings.TrimSpace(key)
				if v, ok := latest.Items[key]; ok {
					filtered[key] = v
				}
			}
			summary["items"] = filtered
		}
	}

	output := metricsOutput{
		Code:        code,
		PeriodCount: len(report.Periods),
		Summary:     summary,
		Data:        report.Periods,
	}
	return mcp.NewToolResultText(toJSON(output)), nil
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
		"CPI":    "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_CPI&columns=TRADE_DATE,NATIONAL_SAME,NATIONAL_BASE&pageSize=%d&pageNumber=1",
		"PMI":    "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_PMI&columns=TRADE_DATE,MAKE_INDEX,NONMANU_INDEX&pageSize=%d&pageNumber=1",
		"GDP":    "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_GDP&columns=TRADE_DATE,GDP,CUM_GDP,CUM_GDP_SAME&pageSize=%d&pageNumber=1",
		"M2":     "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_MONEY_SUPPLY&columns=TRADE_DATE,M2,M2_SAME,M1,M1_SAME&pageSize=%d&pageNumber=1",
		"LPR":    "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_LPR&columns=TRADE_DATE,LPR1Y,LPR5Y&pageSize=%d&pageNumber=1",
		"SHIBOR": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_SHIBOR&columns=TRADE_DATE,ON,ON_CHANGE,W1,W1_CHANGE&pageSize=%d&pageNumber=1",
	}

	urlFmt, ok := macroURLs[indicator]
	if !ok {
		return mcp.NewToolResultError("indicator 必须为: CPI, PMI, GDP, M2, LPR, SHIBOR"), nil
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

	resp, err := client.RAGQuery(ctx, query, topK)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("RAG 检索失败: %v", err)), nil
	}

	type wendaResult struct {
		Query   string      `json:"query"`
		TopK    int         `json:"top_k"`
		Results interface{} `json:"results"`
	}

	result := wendaResult{
		Query:   query,
		TopK:    topK,
		Results: resp.Results,
	}
	return mcp.NewToolResultText(toJSON(result)), nil
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

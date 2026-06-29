package tdx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/scraper"
	"github.com/tdx/go-tdx-mcp/screen"
)

const (
	ToolFactorList       = "tdx_factor_list"
	ToolFactorCompute    = "tdx_factor_compute"
	ToolFactorAnalyze    = "tdx_factor_analyze"
	ToolScreenScan       = "tdx_screen_scan"
	ToolScreenStrength   = "tdx_screen_strength"
	ToolEnhancedBacktest = "tdx_enhanced_backtest"
	ToolTECryptoData     = "tdx_tecrypto_data"
	ToolTECryptoKline    = "tdx_tecrypto_kline"
	ToolFundNAV          = "tdx_fund_nav"
	ToolMarginTrade      = "tdx_margin_trade"
	ToolDragonTiger      = "tdx_dragon_tiger"
	ToolConvertibleBond  = "tdx_convertible_bond"
	ToolFuturesQuote     = "tdx_futures_quote"
	ToolStockCodeResolve = "tdx_stock_code_resolve"
	ToolCSIIndexConstituents = "tdx_csi_index_constituents"
	ToolNewsSearch       = "tdx_news_search"
	ToolCurrentTimestamp = "tdx_current_timestamp"
	// New tools
	ToolTEFundData       = "tdx_tefund_data"
	ToolTEFuturesData    = "tdx_tefutures_data"
	ToolTEMacroData      = "tdx_temacro_data"
	ToolSinaQuotes       = "tdx_sina_quotes"
	ToolSinaHKQuotes     = "tdx_sina_hk_quotes"
	ToolSinaUSQuotes     = "tdx_sina_us_quotes"
	ToolFundHolding      = "tdx_fund_holding"
	ToolFundManagers     = "tdx_fund_managers"
	ToolFundSearch       = "tdx_fund_search"
	ToolHKUSFinancial    = "tdx_hkus_financial"
	ToolHKUSQuote        = "tdx_hkus_quote"
	ToolHKUSBasicInfo    = "tdx_hkus_basic_info"
	ToolHKUSSearchStocks = "tdx_hkus_search_stocks"
	ToolBlockTrades        = "tdx_block_trades"
	ToolBlockTradesByStock = "tdx_block_trades_by_stock"
	ToolBlockTradeStats    = "tdx_block_trade_stats"
	ToolBlockActiveStocks  = "tdx_block_active_stocks"
	ToolSectorBoards       = "tdx_sector_boards"
	ToolSectorBoardStocks  = "tdx_sector_board_stocks"
	ToolMacroDataWeb       = "tdx_macro_data_web"
	ToolNorthboundFlow     = "tdx_northbound_flow"
	ToolNorthboundStocks   = "tdx_northbound_stocks"
	ToolNorthboundDaily    = "tdx_northbound_daily"
	ToolNorthboundHolders  = "tdx_northbound_holders"
	ToolFundNavWeb         = "tdx_fund_nav_web"
	ToolFundNavHistory     = "tdx_fund_nav_history"
	ToolMarginTradeWeb     = "tdx_margin_trade_web"
)

func NewFactorListTool() mcp.Tool {
	return mcp.NewTool(ToolFactorList,
		mcp.WithDescription("列出所有可用的量化因子（momentum/technical/volume/volatility/chanlun/value/quality），共19个因子"),
		mcp.WithString("category",
			mcp.Description("按分类筛选：momentum/technical/volume/volatility/chanlun/value/quality（不填返回全部）"),
		),
	)
}

func NewTECryptoDataTool() mcp.Tool {
	return mcp.NewTool(ToolTECryptoData,
		mcp.WithDescription("加密货币实时报价（Binance API，免费无需Token）"),
		mcp.WithString("symbols_crypto",
			mcp.Required(),
			mcp.Description("逗号分隔的交易对，如 'BTC,ETH,SOL' 或 'BTC/USDT,ETH/USDT'"),
		),
	)
}

func NewTECryptoKlineTool() mcp.Tool {
	return mcp.NewTool(ToolTECryptoKline,
		mcp.WithDescription("加密货币K线数据（Binance API，免费无需Token）"),
		mcp.WithString("symbol_crypto",
			mcp.Required(),
			mcp.Description("交易对，如 'BTC' 或 'BTC/USDT'"),
		),
		mcp.WithString("interval",
			mcp.Description("K线周期: 1m/5m/15m/30m/1h/4h/1d/1w/1M (默认1h)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("K线数量 (默认100, 最大1000)"),
		),
	)
}

func NewFundNAVTool() mcp.Tool {
	return mcp.NewTool(ToolFundNAV,
		mcp.WithDescription("基金净值查询（天天基金网，免费无需Token）"),
		mcp.WithString("fund_code",
			mcp.Required(),
			mcp.Description("基金代码，如 '110011'"),
		),
	)
}

func NewMarginTradeTool() mcp.Tool {
	return mcp.NewTool(ToolMarginTrade,
		mcp.WithDescription("融资融券数据（腾讯证券，免费无需Token）"),
		mcp.WithNumber("limit",
			mcp.Description("返回天数 (默认30)"),
		),
	)
}

func NewDragonTigerTool() mcp.Tool {
	return mcp.NewTool(ToolDragonTiger,
		mcp.WithDescription("龙虎榜数据（东方财富，免费无需Token）"),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认20)"),
		),
	)
}

func NewConvertibleBondTool() mcp.Tool {
	return mcp.NewTool(ToolConvertibleBond,
		mcp.WithDescription("可转债数据（东方财富，免费无需Token）"),
	)
}

func NewFuturesQuoteTool() mcp.Tool {
	return mcp.NewTool(ToolFuturesQuote,
		mcp.WithDescription("期货实时报价（腾讯证券，免费无需Token）"),
		mcp.WithString("symbols_crypto",
			mcp.Required(),
			mcp.Description("逗号分隔的期货代码，如 'au2506,cu2506,rb2510'"),
		),
	)
}

func NewStockCodeResolveTool() mcp.Tool {
	return mcp.NewTool(ToolStockCodeResolve,
		mcp.WithDescription("股票代码解析器（东方财富，免费无需Token）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("逗号分隔的股票代码，如 '000001,600000,300750'"),
		),
	)
}

func NewCSIIndexConstituentsTool() mcp.Tool {
	return mcp.NewTool(ToolCSIIndexConstituents,
		mcp.WithDescription("沪深指数成分股（东方财富，免费无需Token）"),
		mcp.WithString("index_code",
			mcp.Required(),
			mcp.Description("指数代码，如 '1.000300'(沪深300), '0.399001'(深证成指)"),
		),
	)
}

func NewNewsSearchTool() mcp.Tool {
	return mcp.NewTool(ToolNewsSearch,
		mcp.WithDescription("新闻搜索（百度搜索，免费无需Token）"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回数量 (默认20, 最大50)"),
		),
	)
}

func NewCurrentTimestampTool() mcp.Tool {
	return mcp.NewTool(ToolCurrentTimestamp,
		mcp.WithDescription("当前时间戳（毫秒级）"),
	)
}

func NewFactorComputeTool() mcp.Tool {
	return mcp.NewTool(ToolFactorCompute,
		mcp.WithDescription("对指定股票计算因子值，需要先获取K线数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("factors",
			mcp.Required(),
			mcp.Description("逗号分隔的因子名列表"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day/week/month (默认day)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewFactorAnalyzeTool() mcp.Tool {
	return mcp.NewTool(ToolFactorAnalyze,
		mcp.WithDescription("分析单个因子的有效性（IC分析、分层收益、Top-Bottom收益）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("factor_name",
			mcp.Required(),
			mcp.Description("因子名称，如 'momentum_20d'"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认500)"),
		),
	)
}

func NewScreenScanTool() mcp.Tool {
	return mcp.NewTool(ToolScreenScan,
		mcp.WithDescription("对指定股票进行技术信号扫描（MACD金叉/KDJ金叉/放量突破/均线收敛）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewScreenStrengthTool() mcp.Tool {
	return mcp.NewTool(ToolScreenStrength,
		mcp.WithDescription("计算股票强势分排行（5/20/60日动量加权，三种模式：steady稳健/breakout突破/balanced均衡）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("mode",
			mcp.Description("排行模式: steady/breakout/balanced (默认balanced)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewEnhancedBacktestTool() mcp.Tool {
	return mcp.NewTool(ToolEnhancedBacktest,
		mcp.WithDescription("增强回测：支持16种策略+滑点模型+执行模拟，返回完整绩效指标"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("strategy",
			mcp.Required(),
			mcp.Description("策略类型: "+strings.Join(backtest.AvailableStrategies(), ", ")),
		),
		mcp.WithNumber("cash",
			mcp.Description("初始资金 (默认1000000)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认2000)"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day/week/month (默认day)"),
		),
		mcp.WithString("combo_mode",
			mcp.Description("多策略组合模式: and/or/majority（多策略时生效）"),
		),
	)
}

func NewSectorBoardsTool() mcp.Tool {
	return mcp.NewTool(ToolSectorBoards,
		mcp.WithDescription("查询A股板块列表（行业/概念/地域/指数/政策/自定义）"),
		mcp.WithString("board_type",
			mcp.Required(),
			mcp.Description("板块类型: industry(行业)/concept(概念)/region(地域)/index(指数)/policy(政策)/custom(自定义)"),
		),
	)
}

func NewSectorBoardStocksTool() mcp.Tool {
	return mcp.NewTool(ToolSectorBoardStocks,
		mcp.WithDescription("查询指定板块的成分股列表"),
		mcp.WithString("board_code",
			mcp.Required(),
			mcp.Description("板块代码（如行业板块代码）"),
		),
	)
}

func GetAllNewTools() []mcp.Tool {
	return []mcp.Tool{
		NewFactorListTool(),
		NewFactorComputeTool(),
		NewFactorAnalyzeTool(),
		NewScreenScanTool(),
		NewScreenStrengthTool(),
		NewEnhancedBacktestTool(),
		NewTECryptoKlineTool(),
		NewFundNAVTool(),
		NewMarginTradeTool(),
		NewDragonTigerTool(),
		NewConvertibleBondTool(),
		NewFuturesQuoteTool(),
		NewStockCodeResolveTool(),
		NewCSIIndexConstituentsTool(),
		NewNewsSearchTool(),
		NewCurrentTimestampTool(),
		// New tools
		NewTECryptoDataTool(),
		NewTEFundDataTool(),
		NewTEFuturesDataTool(),
		NewTEMacroDataTool(),
		NewSinaQuotesTool(),
		NewSinaHKQuotesTool(),
		NewSinaUSQuotesTool(),
		NewFundHoldingTool(),
		NewFundManagersTool(),
		NewFundSearchTool(),
		NewHKUSFinancialTool(),
		NewHKUSQuoteTool(),
		NewHKUSBasicInfoTool(),
		NewHKUSSearchStocksTool(),
		NewBlockTradesTool(),
		NewBlockTradesByStockTool(),
		NewBlockTradeStatsTool(),
		NewBlockActiveStocksTool(),
		NewSectorBoardsTool(),
		NewSectorBoardStocksTool(),
		NewMacroDataWebTool(),
		NewNorthboundFlowTool(),
		NewNorthboundDailyTool(),
		NewNorthboundStocksTool(),
		NewNorthboundHoldersTool(),
		NewFundNavWebTool(),
		NewFundNavHistoryTool(),
		NewMarginTradeWebTool(),
	}
}

func GetNewHandler(name string) ToolHandler {
	switch name {
	case ToolFactorList:
		return HandleFactorList
	case ToolFactorCompute:
		return HandleFactorCompute
	case ToolFactorAnalyze:
		return HandleFactorAnalyze
	case ToolScreenScan:
		return HandleScreenScan
	case ToolScreenStrength:
		return HandleScreenStrength
	case ToolEnhancedBacktest:
		return HandleEnhancedBacktest
	case ToolTECryptoData:
		return HandleTECryptoData
	case ToolTECryptoKline:
		return HandleTECryptoKline
	case ToolFundNAV:
		return HandleFundNAV
	case ToolMarginTrade:
		return HandleMarginTrade
	case ToolDragonTiger:
		return HandleDragonTiger
	case ToolConvertibleBond:
		return HandleConvertibleBond
	case ToolFuturesQuote:
		return HandleFuturesQuote
	case ToolStockCodeResolve:
		return HandleStockCodeResolve
	case ToolCSIIndexConstituents:
		return HandleCSIIndexConstituents
	case ToolNewsSearch:
		return HandleNewsSearch
	case ToolCurrentTimestamp:
		return HandleCurrentTimestamp
	// New handlers
	case ToolTEFundData:
		return HandleTEFundData
	case ToolTEFuturesData:
		return HandleTEFuturesData
	case ToolTEMacroData:
		return HandleTEMacroData
	case ToolSinaQuotes:
		return HandleSinaQuotes
	case ToolSinaHKQuotes:
		return HandleSinaHKQuotes
	case ToolSinaUSQuotes:
		return HandleSinaUSQuotes
	case ToolFundHolding:
		return HandleFundHolding
	case ToolFundManagers:
		return HandleFundManagers
	case ToolFundSearch:
		return HandleFundSearch
	case ToolHKUSFinancial:
		return HandleHKUSFinancial
	case ToolHKUSQuote:
		return HandleHKUSQuote
	case ToolHKUSBasicInfo:
		return HandleHKUSBasicInfo
	case ToolHKUSSearchStocks:
		return HandleHKUSSearchStocks
	case ToolBlockTrades:
		return HandleBlockTrades
	case ToolBlockTradesByStock:
		return HandleBlockTradesByStock
	case ToolBlockTradeStats:
		return HandleBlockTradeStats
	case ToolBlockActiveStocks:
		return HandleBlockActiveStocks
	case ToolSectorBoards:
		return HandleSectorBoards
	case ToolSectorBoardStocks:
		return HandleSectorBoardStocks
	case ToolMacroDataWeb:
		return HandleMacroDataWeb
	case ToolNorthboundFlow:
		return HandleNorthboundFlow
	case ToolNorthboundDaily:
		return HandleNorthboundDaily
	case ToolNorthboundStocks:
		return HandleNorthboundStocks
	case ToolNorthboundHolders:
		return HandleNorthboundHolders
	case ToolFundNavWeb:
		return HandleFundNavWeb
	case ToolFundNavHistory:
		return HandleFundNavHistory
	case ToolMarginTradeWeb:
		return HandleMarginTradeWeb
	default:
		return nil
	}
}

func HandleFactorList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category, _ := request.GetArguments()["category"].(string)
	type FactorInfo struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	var factors []FactorInfo
	allNames := factor.List()
	for _, name := range allNames {
		f := factor.Get(name)
		if f == nil {
			continue
		}
		if category != "" && f.Category() != category {
			continue
		}
		factors = append(factors, FactorInfo{
			Name:        f.Name(),
			Category:    f.Category(),
			Description: f.Description(),
		})
	}
	return mcp.NewToolResultText(toJSON(factors)), nil
}

func HandleFactorCompute(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	factorStr, err := request.RequireString("factors")
	if err != nil {
		return mcp.NewToolResultError("factors 参数必填"), nil
	}
	factorNames := strings.Split(factorStr, ",")
	for i := range factorNames {
		factorNames[i] = strings.TrimSpace(factorNames[i])
	}

	period := "day"
	count := 200
	if v, ok := request.GetArguments()["period"].(string); ok {
		period = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode(period),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	bars, err := parseKlineBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	engine := factor.NewEngine()
	result, err := engine.ComputeSingle(bars, factorNames)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("计算因子失败: %v", err)), nil
	}

	type FactorResult struct {
		Code   string                       `json:"code"`
		Bars   int                          `json:"bars"`
		Values map[string]map[string]float64 `json:"values"`
	}
	fr := FactorResult{Code: code, Bars: len(bars), Values: make(map[string]map[string]float64)}
	for name, vals := range result {
		if len(vals) > 0 {
			fr.Values[name] = map[string]float64{
				"latest": vals[len(vals)-1],
				"mean":   meanSlice(vals),
			}
		}
	}
	return mcp.NewToolResultText(toJSON(fr)), nil
}

func HandleFactorAnalyze(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	factorName, err := request.RequireString("factor_name")
	if err != nil {
		return mcp.NewToolResultError("factor_name 参数必填"), nil
	}

	count := 500
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode("day"),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	bars, err := parseKlineBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	f := factor.Get(factorName)
	if f == nil {
		return mcp.NewToolResultError("未知因子: " + factorName + "，可用因子: " + strings.Join(factor.List(), ", ")), nil
	}

	factorVals := f.Compute(bars)
	forward5d := make([]float64, len(bars))
	period := 5
	for i := 0; i < len(bars)-period; i++ {
		if bars[i].Close > 0 {
			forward5d[i] = bars[i+period].Close/bars[i].Close - 1
		}
	}

	analyzer := factor.NewAnalyzer(factorVals, forward5d, factorName, 5)
	report := analyzer.FullReport()

	return mcp.NewToolResultText(toJSON(report)), nil
}

func HandleScreenScan(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode("day"),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	bars, err := parseKlineBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	scanner := screen.NewScanner()
	predicates := screen.DefaultScanPredicates()
	results := scanner.ScanBars(code, bars, predicates)

	return mcp.NewToolResultText(toJSON(results)), nil
}

func HandleScreenStrength(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	mode := "balanced"
	if v, ok := request.GetArguments()["mode"].(string); ok {
		mode = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode("day"),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	bars, err := parseKlineBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	ranker := screen.NewStrengthRanker(mode)
	ranker.AddBars(code, bars)
	results := ranker.Rank(1)

	return mcp.NewToolResultText(toJSON(results)), nil
}

func HandleEnhancedBacktest(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	strategy, err := request.RequireString("strategy")
	if err != nil {
		return mcp.NewToolResultError("strategy 参数必填"), nil
	}

	period := "day"
	count := 2000
	cash := 1000000.0
	comboMode := ""
	if v, ok := request.GetArguments()["period"].(string); ok {
		period = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := request.GetArguments()["cash"].(float64); ok {
		cash = v
	}
	if v, ok := request.GetArguments()["combo_mode"].(string); ok {
		comboMode = v
	}

	st := backtest.NewStrategy(strategy)
	if st == nil {
		return mcp.NewToolResultError("strategy 必须为: " + strings.Join(backtest.AvailableStrategies(), ", ")), nil
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode(period),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	bars, err := parseKlineBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	engine := backtest.NewEngine(cash)
	engine.SetCommission(0.00025)
	engine.SetSlippage(0.0005)

	if comboMode != "" {
		strategies := strings.Split(strategy, ",")
		if len(strategies) > 1 {
			var strategyList []backtest.Strategy
			for _, name := range strategies {
				s := backtest.NewStrategy(strings.TrimSpace(name))
				if s != nil {
					strategyList = append(strategyList, s)
				}
			}
			if len(strategyList) >= 2 {
				var mode backtest.ComboMode
				switch comboMode {
				case "and":
					mode = backtest.ComboAnd
				case "or":
					mode = backtest.ComboOr
				case "majority":
					mode = backtest.ComboMajority
				default:
					mode = backtest.ComboMajority
				}
				comboResult := backtest.RunCombo(engine, strategyList, bars, mode)
				comboResult.Results[0].Code = code
				comboResult.Results[0].Market = int(market)
				comboResult.Results[0].Period = period
				return mcp.NewToolResultText(toJSON(comboResult)), nil
			}
		}
	}

	btResult := engine.Run(st, bars)
	btResult.Code = code
	btResult.Market = int(market)
	btResult.Period = period

	return mcp.NewToolResultText(toJSON(btResult)), nil
}

func meanSlice(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	count := 0
	for _, v := range vals {
		if !isNaN(v) {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func isNaN(v float64) bool {
	return v != v
}

func HandleTECryptoData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolsStr, err := request.RequireString("symbols_crypto")
	if err != nil {
		return mcp.NewToolResultError("symbols 参数必填"), nil
	}
	symbols := strings.Split(symbolsStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	teClient := scraper.NewTEEconClient()
	data, err := teClient.GetCryptoData(symbols)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取加密货币数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleTECryptoKline(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol, err := request.RequireString("symbol_crypto")
	if err != nil {
		return mcp.NewToolResultError("symbol 参数必填"), nil
	}

	interval := "1h"
	limit := 100
	if v, ok := request.GetArguments()["interval"].(string); ok {
		interval = v
	}
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	teClient := scraper.NewTEEconClient()
	klines, err := teClient.GetCryptoKline(symbol, interval, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(klines)), nil
}

func HandleFundNAV(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}

	fundClient := scraper.NewEastMoneyFundClient()
	data, err := fundClient.GetFundNetValue(fundCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金净值失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleMarginTrade(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 30
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	marginClient := scraper.NewMarginTradeWebClient()
	data, err := marginClient.GetSummary()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取融资融券数据失败: %v", err)), nil
	}

	if limit < len(data) {
		data = data[:limit]
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleDragonTiger(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 20
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	dtClient := scraper.NewDragonTigerClient()
	data, err := dtClient.GetLatest(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取龙虎榜数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleConvertibleBond(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cbClient := scraper.NewConvertibleBondClient()
	data, err := cbClient.GetAll()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取可转债数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleFuturesQuote(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolsStr, err := request.RequireString("symbols_crypto")
	if err != nil {
		return mcp.NewToolResultError("symbols 参数必填"), nil
	}
	symbols := strings.Split(symbolsStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	futuresClient := scraper.NewFuturesClient()
	var data []*scraper.FuturesData
	for _, sym := range symbols {
		fd, err := futuresClient.GetQuote(sym)
		if err != nil {
			continue
		}
		data = append(data, fd)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取期货数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleStockCodeResolve(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}

	resolver := scraper.NewStockCodeResolver()
	results := resolver.BatchResolve(codes)

	return mcp.NewToolResultText(toJSON(results)), nil
}

func HandleCSIIndexConstituents(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indexCode, err := request.RequireString("index_code")
	if err != nil {
		return mcp.NewToolResultError("index_code 参数必填"), nil
	}

	indexClient := scraper.NewCSIIndexClient()
	data, err := indexClient.GetIndexData(indexCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取指数数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleNewsSearch(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError("keyword 参数必填"), nil
	}

	count := 20
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	newsClient := scraper.NewNewsCrawler()
	articles, err := newsClient.Search(keyword, count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索新闻失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(articles)), nil
}

func HandleCurrentTimestamp(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ts := scraper.GetCurrentTimestamp()
	result := map[string]interface{}{
		"timestamp_ms": ts,
		"timestamp_s":  ts / 1000,
		"iso8601":      time.UnixMilli(ts).UTC().Format(time.RFC3339),
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// --- New tool definitions ---

func NewTEFundDataTool() mcp.Tool {
	return mcp.NewTool(ToolTEFundData,
		mcp.WithDescription("TradingEconomics 基金数据（东方财富基金净值 API，免费无需Token）"),
		mcp.WithString("fund_codes",
			mcp.Required(),
			mcp.Description("逗号分隔的基金代码，如 '110011,161725'"),
		),
	)
}

func NewTEFuturesDataTool() mcp.Tool {
	return mcp.NewTool(ToolTEFuturesData,
		mcp.WithDescription("TradingEconomics 期货数据（腾讯证券 API，免费无需Token）"),
		mcp.WithString("symbols_crypto",
			mcp.Required(),
			mcp.Description("逗号分隔的期货代码，如 'au2506,cu2506,rb2510'"),
		),
	)
}

func NewTEMacroDataTool() mcp.Tool {
	return mcp.NewTool(ToolTEMacroData,
		mcp.WithDescription("TradingEconomics 宏观经济数据（东方财富宏观 API，免费无需Token）"),
		mcp.WithString("indicators",
			mcp.Required(),
			mcp.Description("经济指标，如 'GDP,CPI,PPI'"),
		),
	)
}

func NewSinaQuotesTool() mcp.Tool {
	return mcp.NewTool(ToolSinaQuotes,
		mcp.WithDescription("新浪财经 A 股实时报价（免费无需Token）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("逗号分隔的股票代码，如 '000001,600000,300750'"),
		),
	)
}

func NewSinaHKQuotesTool() mcp.Tool {
	return mcp.NewTool(ToolSinaHKQuotes,
		mcp.WithDescription("新浪财经 港股实时报价（免费无需Token）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("逗号分隔的港股代码，如 '00700,09988'"),
		),
	)
}

func NewSinaUSQuotesTool() mcp.Tool {
	return mcp.NewTool(ToolSinaUSQuotes,
		mcp.WithDescription("新浪财经 美股实时报价（免费无需Token）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("逗号分隔的美股代码，如 'AAPL,TSLA,MSFT'"),
		),
	)
}

func NewFundHoldingTool() mcp.Tool {
	return mcp.NewTool(ToolFundHolding,
		mcp.WithDescription("基金持仓详情（东方财富基金季报 API，免费无需Token）"),
		mcp.WithString("fund_code",
			mcp.Required(),
			mcp.Description("基金代码，如 '110011'"),
		),
		mcp.WithString("report_period",
			mcp.Description("报告期，如 '2024Q1'"),
		),
	)
}

func NewFundManagersTool() mcp.Tool {
	return mcp.NewTool(ToolFundManagers,
		mcp.WithDescription("基金经理信息（东方财富基金基金经理 API，免费无需Token）"),
		mcp.WithString("fund_code",
			mcp.Required(),
			mcp.Description("基金代码，如 '110011'"),
		),
	)
}

func NewFundSearchTool() mcp.Tool {
	return mcp.NewTool(ToolFundSearch,
		mcp.WithDescription("基金搜索（东方财富基金搜索 API，免费无需Token）"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词，如 '沪深300'"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认20)"),
		),
	)
}

func NewHKUSFinancialTool() mcp.Tool {
	return mcp.NewTool(ToolHKUSFinancial,
		mcp.WithDescription("港股/美股财报数据（东方财富 API，免费无需Token）"),
		mcp.WithString("stock_code",
			mcp.Required(),
			mcp.Description("股票代码，如 '00700' 或 'AAPL'"),
		),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: HK=港股, US=美股"),
		),
		mcp.WithString("report_type",
			mcp.Required(),
			mcp.Description("报表类型: income=利润表, balance=资产负债表, cashflow=现金流量表, ratios=关键指标"),
		),
	)
}

func NewHKUSQuoteTool() mcp.Tool {
	return mcp.NewTool(ToolHKUSQuote,
		mcp.WithDescription("港股/美股实时报价（东方财富 API，免费无需Token）"),
		mcp.WithString("stock_code",
			mcp.Required(),
			mcp.Description("股票代码，如 '00700' 或 'AAPL'"),
		),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: HK=港股, US=美股"),
		),
	)
}

func NewHKUSBasicInfoTool() mcp.Tool {
	return mcp.NewTool(ToolHKUSBasicInfo,
		mcp.WithDescription("港股/美股基本信息（东方财富 API，免费无需Token）"),
		mcp.WithString("stock_code",
			mcp.Required(),
			mcp.Description("股票代码，如 '00700' 或 'AAPL'"),
		),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: HK=港股, US=美股"),
		),
	)
}

func NewHKUSSearchStocksTool() mcp.Tool {
	return mcp.NewTool(ToolHKUSSearchStocks,
		mcp.WithDescription("搜索港股/美股（东方财富搜索 API，免费无需Token）"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词，如 '腾讯' 或 'Apple'"),
		),
		mcp.WithString("market",
			mcp.Description("市场: HK=港股, US=美股 (默认全部)"),
		),
	)
}

func NewBlockTradesTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTrades,
		mcp.WithDescription("大宗交易数据（东方财富 API，免费无需Token）"),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

func NewBlockTradesByStockTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTradesByStock,
		mcp.WithDescription("按股票代码查询大宗交易（东方财富 API，免费无需Token）"),
		mcp.WithString("stock_code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

func NewBlockTradeStatsTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTradeStats,
		mcp.WithDescription("大宗交易统计（东方财富 API，免费无需Token）"),
		mcp.WithString("start_date",
			mcp.Required(),
			mcp.Description("开始日期，格式 YYYY-MM-DD"),
		),
		mcp.WithString("end_date",
			mcp.Required(),
			mcp.Description("结束日期，格式 YYYY-MM-DD"),
		),
	)
}

func NewBlockActiveStocksTool() mcp.Tool {
	return mcp.NewTool(ToolBlockActiveStocks,
		mcp.WithDescription("大宗交易活跃股票（东方财富 API，免费无需Token）"),
		mcp.WithNumber("days",
			mcp.Description("查询天数 (默认7)"),
		),
	)
}

// --- New handlers ---

func HandleTEFundData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCodesStr, err := request.RequireString("fund_codes")
	if err != nil {
		return mcp.NewToolResultError("fund_codes 参数必填"), nil
	}
	fundCodes := strings.Split(fundCodesStr, ",")
	for i := range fundCodes {
		fundCodes[i] = strings.TrimSpace(fundCodes[i])
	}

	teClient := scraper.NewTEEconClient()
	data, err := teClient.GetFundData(fundCodes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleTEFuturesData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolsStr, err := request.RequireString("symbols_crypto")
	if err != nil {
		return mcp.NewToolResultError("symbols 参数必填"), nil
	}
	symbols := strings.Split(symbolsStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	teClient := scraper.NewTEEconClient()
	data, err := teClient.GetFuturesData(symbols)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取期货数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleTEMacroData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indicatorsStr, err := request.RequireString("indicators")
	if err != nil {
		return mcp.NewToolResultError("indicators 参数必填"), nil
	}
	indicators := strings.Split(indicatorsStr, ",")
	for i := range indicators {
		indicators[i] = strings.TrimSpace(indicators[i])
	}

	teClient := scraper.NewTEEconClient()
	data, err := teClient.GetMacroData(indicators)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取宏观数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleSinaQuotes(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}

	sinaClient := scraper.NewSinaClient()
	data, err := sinaClient.GetStockQuotes(codes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取 A 股报价失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleSinaHKQuotes(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}

	sinaClient := scraper.NewSinaClient()
	data, err := sinaClient.GetHKStockQuotes(codes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取港股报价失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleSinaUSQuotes(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}

	sinaClient := scraper.NewSinaClient()
	data, err := sinaClient.GetUSStockQuotes(codes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取美股报价失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleFundHolding(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}

	reportPeriod := ""
	if v, ok := request.GetArguments()["report_period"].(string); ok {
		reportPeriod = v
	}

	fundClient := scraper.NewFundHoldingClient()
	data, err := fundClient.GetHoldingReport(fundCode, reportPeriod)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金持仓失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleFundManagers(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}

	fundClient := scraper.NewFundHoldingClient()
	data, err := fundClient.GetFundManagers(fundCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金经理信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleFundSearch(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError("keyword 参数必填"), nil
	}

	limit := 20
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	fundClient := scraper.NewFundHoldingClient()
	data, err := fundClient.SearchFunds(keyword, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索基金失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleHKUSFinancial(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stockCode, err := request.RequireString("stock_code")
	if err != nil {
		return mcp.NewToolResultError("stock_code 参数必填"), nil
	}
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	reportType, err := request.RequireString("report_type")
	if err != nil {
		return mcp.NewToolResultError("report_type 参数必填"), nil
	}

	hkUsClient := scraper.NewHKUSFinancialClient()
	var data []*scraper.HKUSStockFinancial

	switch strings.ToLower(reportType) {
	case "income":
		data, err = hkUsClient.GetIncomeStatement(stockCode, market)
	case "balance":
		data, err = hkUsClient.GetBalanceSheet(stockCode, market)
	case "cashflow":
		data, err = hkUsClient.GetCashFlow(stockCode, market)
	case "ratios":
		data, err = hkUsClient.GetKeyRatios(stockCode, market)
	default:
		return mcp.NewToolResultError("report_type 必须为: income/balance/cashflow/ratios"), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取财报数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleHKUSQuote(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stockCode, err := request.RequireString("stock_code")
	if err != nil {
		return mcp.NewToolResultError("stock_code 参数必填"), nil
	}
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	hkUsClient := scraper.NewHKUSFinancialClient()
	data, err := hkUsClient.GetStockQuote(stockCode, market)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取报价失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleHKUSBasicInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stockCode, err := request.RequireString("stock_code")
	if err != nil {
		return mcp.NewToolResultError("stock_code 参数必填"), nil
	}
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	hkUsClient := scraper.NewHKUSFinancialClient()
	data, err := hkUsClient.GetStockBasicInfo(stockCode, market)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基本信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleHKUSSearchStocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError("keyword 参数必填"), nil
	}

	market := ""
	if v, ok := request.GetArguments()["market"].(string); ok {
		market = v
	}

	hkUsClient := scraper.NewHKUSFinancialClient()

	var results []map[string]interface{}
	if market == "HK" || market == "" {
		hkResults, err := hkUsClient.SearchHKStocks(keyword)
		if err == nil && len(hkResults) > 0 {
			results = append(results, hkResults...)
		}
	}
	if market == "US" || market == "" {
		usResults, err := hkUsClient.SearchUSStocks(keyword)
		if err == nil && len(usResults) > 0 {
			results = append(results, usResults...)
		}
	}

	return mcp.NewToolResultText(toJSON(results)), nil
}

func HandleBlockTrades(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	btClient := scraper.NewBlockTradeClient()
	data, err := btClient.GetBlockTrades(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取大宗交易数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleBlockTradesByStock(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stockCode, err := request.RequireString("stock_code")
	if err != nil {
		return mcp.NewToolResultError("stock_code 参数必填"), nil
	}

	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	btClient := scraper.NewBlockTradeClient()
	data, err := btClient.GetBlockTradesByStock(stockCode, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取大宗交易数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleBlockTradeStats(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startDate, err := request.RequireString("start_date")
	if err != nil {
		return mcp.NewToolResultError("start_date 参数必填"), nil
	}
	endDate, err := request.RequireString("end_date")
	if err != nil {
		return mcp.NewToolResultError("end_date 参数必填"), nil
	}

	btClient := scraper.NewBlockTradeClient()
	data, err := btClient.GetBlockTradeStatistics(startDate, endDate)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取大宗交易统计失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

func HandleBlockActiveStocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := 7
	if v, ok := request.GetArguments()["days"].(float64); ok {
		days = int(v)
	}

	btClient := scraper.NewBlockTradeClient()
	data, err := btClient.GetRecentActiveStocks(days)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取大宗交易活跃股票失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

// HandleSectorBoards fetches sector board lists by type.
func HandleSectorBoards(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardTypeStr, err := request.RequireString("board_type")
	if err != nil {
		return mcp.NewToolResultError("board_type 参数必填"), nil
	}

	var bt BlockType
	switch strings.ToLower(boardTypeStr) {
	case "industry", "gy":
		bt = BlockIndustry
	case "concept", "gn":
		bt = BlockConcept
	case "region", "dy":
		bt = BlockRegion
	case "index", "zs":
		bt = BlockIndex
	case "policy", "zc":
		bt = BlockPolicy
	case "custom", "zdy":
		bt = BlockCustom
	default:
		return mcp.NewToolResultError(fmt.Sprintf("未知板块类型: %s (可选: industry/concept/region/index/policy/custom)", boardTypeStr)), nil
	}

	// Type-assert to get TCP client
	type sectorClient interface {
		GetSectorBoards(bt BlockType) ([]SectorBoard, error)
	}
	sc, ok := client.(sectorClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持板块查询，请使用 UnifiedClient"), nil
	}

	boards, err := sc.GetSectorBoards(bt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询板块失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(boards)), nil
}

// HandleSectorBoardStocks fetches constituent stocks of a board.
func HandleSectorBoardStocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardCode, err := request.RequireString("board_code")
	if err != nil {
		return mcp.NewToolResultError("board_code 参数必填"), nil
	}

	type sectorStocksClient interface {
		GetSectorBoardStocks(boardCode string) ([]string, error)
	}
	sc, ok := client.(sectorStocksClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持板块查询，请使用 UnifiedClient"), nil
	}

	stocks, err := sc.GetSectorBoardStocks(boardCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询板块成分股失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(stocks)), nil
}

// NewMacroDataWebTool returns the macro data tool definition.
func NewMacroDataWebTool() mcp.Tool {
	return mcp.NewTool(ToolMacroDataWeb,
		mcp.WithDescription("宏观经济数据查询（支持CPI/GDP/PMI/LPR/SHIBOR/M2）"),
		mcp.WithString("indicator",
			mcp.Description("指标名称: CPI/GDP/PMI/LPR/SHIBOR/M2 (默认CPI)"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回数据条数 (默认12)"),
		),
	)
}

// NewNorthboundFlowTool returns the northbound flow tool definition.
func NewNorthboundFlowTool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundFlow,
		mcp.WithDescription("北向资金实时净流入（东方财富数据源）"),
	)
}

// NewNorthboundDailyTool returns the northbound daily flow tool definition.
func NewNorthboundDailyTool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundDaily,
		mcp.WithDescription("北向资金历史净流入（东方财富数据源）"),
		mcp.WithNumber("days",
			mcp.Description("查询天数 (默认30)"),
		),
	)
}

// NewNorthboundStocksTool returns the northbound stocks tool definition.
func NewNorthboundStocksTool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundStocks,
		mcp.WithDescription("北向资金持仓个股 Top N（东方财富数据源）"),
		mcp.WithNumber("count",
			mcp.Description("返回数量 (默认20, 最大200)"),
		),
		mcp.WithString("sort_field",
			mcp.Description("排序字段: f62=持股市值 (默认)"),
		),
	)
}

// NewNorthboundHoldersTool returns the northbound holders tool definition.
func NewNorthboundHoldersTool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundHolders,
		mcp.WithDescription("北向资金机构持仓排名（东方财富数据源）"),
		mcp.WithString("mutual_type",
			mcp.Description("市场类型: 001=沪股通, 003=深股通, 空=全部"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("返回数量 (默认20)"),
		),
	)
}

// NewFundNavWebTool returns the fund nav web tool definition.
func NewFundNavWebTool() mcp.Tool {
	return mcp.NewTool(ToolFundNavWeb,
		mcp.WithDescription("基金净值查询（goquery网页解析，天天基金数据源）"),
		mcp.WithString("fund_code",
			mcp.Description("基金代码，如 110011"),
		),
	)
}

// NewFundNavHistoryTool returns the fund nav history tool definition.
func NewFundNavHistoryTool() mcp.Tool {
	return mcp.NewTool(ToolFundNavHistory,
		mcp.WithDescription("基金净值历史查询（goquery网页解析）"),
		mcp.WithString("fund_code",
			mcp.Description("基金代码，如 110011"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认10)"),
		),
	)
}

// NewMarginTradeWebTool returns the margin trade web tool definition.
func NewMarginTradeWebTool() mcp.Tool {
	return mcp.NewTool(ToolMarginTradeWeb,
		mcp.WithDescription("融资融券数据查询（东方财富datacenter API）"),
	)
}

// HandleMacroDataWeb fetches macro data using MacroScraper.
func HandleMacroDataWeb(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indicator := "CPI"
	count := 12
	if v, ok := request.GetArguments()["indicator"].(string); ok && v != "" {
		indicator = strings.ToUpper(v)
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	type macroClient interface {
		GetMacroCPI(int) ([]scraper.MacroIndicator, error)
		GetMacroGDP(int) ([]scraper.MacroIndicator, error)
		GetMacroPMI(int) ([]scraper.MacroIndicator, error)
		GetMacroLPR(int) ([]scraper.MacroIndicator, error)
		GetMacroShibor(int) ([]scraper.MacroIndicator, error)
	}

	mc, ok := client.(macroClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持宏观数据查询，请使用 UnifiedClient"), nil
	}

	var data []scraper.MacroIndicator
	var err error
	switch indicator {
	case "CPI":
		data, err = mc.GetMacroCPI(count)
	case "GDP":
		data, err = mc.GetMacroGDP(count)
	case "PMI":
		data, err = mc.GetMacroPMI(count)
	case "LPR":
		data, err = mc.GetMacroLPR(count)
	case "SHIBOR":
		data, err = mc.GetMacroShibor(count)
	case "M2":
		// M2 not yet exposed as tool indicator, but available via UnifiedClient
		return mcp.NewToolResultError("M2 数据请使用 GetMacroM2 接口"), nil
	default:
		return mcp.NewToolResultError(fmt.Sprintf("不支持的指标: %s (可选: CPI/GDP/PMI/LPR/SHIBOR)", indicator)), nil
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询宏观数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

// HandleNorthboundFlow fetches northbound capital flow.
func HandleNorthboundFlow(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type northClient interface {
		GetNorthboundFlow() ([]scraper.NorthboundFlow, error)
	}

	nc, ok := client.(northClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持北向资金查询，请使用 UnifiedClient"), nil
	}

	flows, err := nc.GetNorthboundFlow()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询北向资金失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(flows)), nil
}

// HandleNorthboundDaily fetches daily northbound flow history.
func HandleNorthboundDaily(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := 30
	if v, ok := request.GetArguments()["days"].(float64); ok {
		days = int(v)
	}

	type northClient interface {
		GetNorthboundDaily(int) ([]scraper.NorthboundFlow, error)
	}

	nc, ok := client.(northClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持北向资金查询，请使用 UnifiedClient"), nil
	}

	flows, err := nc.GetNorthboundDaily(days)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询北向资金失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(flows)), nil
}

// HandleNorthboundStocks fetches top northbound holding stocks.
func HandleNorthboundStocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	count := 20
	sortField := "f62"

	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := request.GetArguments()["sort_field"].(string); ok && v != "" {
		sortField = v
	}

	type northClient interface {
		GetTopNorthboundStocks(string, int) ([]scraper.NorthboundStock, error)
	}

	nc, ok := client.(northClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持北向资金查询，请使用 UnifiedClient"), nil
	}

	stocks, err := nc.GetTopNorthboundStocks(sortField, count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询北向资金持仓失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(stocks)), nil
}

// HandleNorthboundHolders fetches institutional holding rankings from northbound trading.
func HandleNorthboundHolders(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mutualType := ""
	pageSize := 20

	if v, ok := request.GetArguments()["mutual_type"].(string); ok && v != "" {
		mutualType = v
	}
	if v, ok := request.GetArguments()["page_size"].(float64); ok {
		pageSize = int(v)
	}

	type northClient interface {
		GetNorthboundHolders(string, int) ([]*scraper.NorthboundHolder, error)
	}

	nc, ok := client.(northClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持北向资金持仓排名查询，请使用 UnifiedClient"), nil
	}

	holders, err := nc.GetNorthboundHolders(mutualType, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询北向资金持仓排名失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(holders)), nil
}

// HandleFundNavWeb fetches fund NAV via goquery web parser.
func HandleFundNavWeb(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}

	type fundNavClient interface {
		GetFundNav(string) (*scraper.FundNav, error)
	}

	fc, ok := client.(fundNavClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持基金净值查询，请使用 UnifiedClient"), nil
	}

	nav, err := fc.GetFundNav(fundCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金净值失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(nav)), nil
}

// HandleFundNavHistory fetches fund NAV history via goquery web parser.
func HandleFundNavHistory(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}

	limit := 10
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	type fundNavHistoryClient interface {
		GetFundNavHistory(string, int) ([]*scraper.FundNav, error)
	}

	fc, ok := client.(fundNavHistoryClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持基金净值历史查询，请使用 UnifiedClient"), nil
	}

	history, err := fc.GetFundNavHistory(fundCode, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金净值历史失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(history)), nil
}

// HandleMarginTradeWeb fetches margin trading data via eastmoney datacenter API.
func HandleMarginTradeWeb(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type marginTradeClient interface {
		GetMarginTrade() ([]*scraper.MarginTradeData, error)
	}

	mc, ok := client.(marginTradeClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持融资融券查询，请使用 UnifiedClient"), nil
	}

	data, err := mc.GetMarginTrade()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取融资融券数据失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(data)), nil
}

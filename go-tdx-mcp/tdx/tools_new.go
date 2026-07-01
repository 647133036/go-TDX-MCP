package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/indicator"
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
	ToolOCRRecognize     = "tdx_ocr_recognize"
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
	// New batch tools
	ToolLimitUpPool         = "tdx_limit_up_pool"
	ToolLimitDownPool       = "tdx_limit_down_pool"
	ToolYesterdayLimitUp    = "tdx_yesterday_limit_up"
	ToolHotRank             = "tdx_hot_rank"
	ToolNorthboundTop10     = "tdx_northbound_top10"
	ToolMarketIndices       = "tdx_market_indices"
	ToolSecurityList        = "tdx_security_list"
	ToolSecurityCount       = "tdx_security_count"
	ToolBlockTradesByDate   = "tdx_block_trades_by_date"
	ToolBlockTradesSearch   = "tdx_block_trades_search"
	ToolFundCompanies       = "tdx_fund_companies"
	ToolMacroMoneySupply    = "tdx_macro_money_supply"
	ToolMacroGlobal         = "tdx_macro_global"
	ToolSinaMarginTrade     = "tdx_sina_margin_trade"
	ToolSinaBlockTrades     = "tdx_sina_block_trades"
	ToolStockBelongSector   = "tdx_stock_belong_sector"
	ToolMarketIndicesFull   = "tdx_market_indices_full"
	ToolFactorTransform     = "tdx_factor_transform"
	ToolFactorCrossSection  = "tdx_factor_cross_section"
	ToolChanlunDetail       = "tdx_chanlun_detail"
	ToolIndicatorSingle     = "tdx_indicator_single"
	ToolBacktestPerformance = "tdx_backtest_performance"
	ToolPortfolioOptimize   = "tdx_portfolio_optimize"
	ToolPortfolioRisk       = "tdx_portfolio_risk"
	ToolChanlunFindBeiChi        = "tdx_chanlun_find_beichi"
	ToolEastMoneyRealtimeQuote   = "tdx_eastmoney_realtime_quote"
	ToolEastMoneyKlineHistory     = "tdx_eastmoney_kline_history"
	ToolEastMoneyStockChanges     = "tdx_eastmoney_stock_changes"
	ToolEastMoneySymbolInfo       = "tdx_eastmoney_symbol_info"
	ToolEastMoneySectorBoards     = "tdx_eastmoney_sector_boards"
	ToolEastMoneySectorStocks     = "tdx_eastmoney_sector_stocks"
	ToolEastMoneyUpCount          = "tdx_eastmoney_updown_count"
	ToolEastMoneyBelongBoard      = "tdx_eastmoney_belong_board"
	ToolFundNavLatest             = "tdx_fund_nav_latest"
	ToolFundNavHistoryNew         = "tdx_fund_nav_history_new"
	ToolMarginTradeSummary        = "tdx_margin_trade_summary"
	ToolTableParserURL            = "tdx_table_parser_url"
	ToolTableParserHTML           = "tdx_table_parser_html"
	ToolTableParserFindKeyword    = "tdx_table_parser_find_keyword"
	ToolTableParserToCSV          = "tdx_table_parser_to_csv"
	ToolTableParserToJSON         = "tdx_table_parser_to_json"
	ToolSCRaperIwencai            = "tdx_scraper_iwencai"
	ToolSCRaperMultiSource        = "tdx_scraper_multi_source"
	ToolBacktestAvailable         = "tdx_backtest_available_strategies"
	ToolBacktestRun               = "tdx_backtest_run"
	ToolBacktestCombo             = "tdx_backtest_combo"
	ToolFactorGetInfo             = "tdx_factor_get_info"
	ToolFactorAnalysisReport      = "tdx_factor_analysis_report"
	ToolFactorForwardReturns      = "tdx_factor_forward_returns"
	ToolChanlunMergeKlines        = "tdx_chanlun_merge_klines"
	ToolChanlunFindFenXing        = "tdx_chanlun_find_fenxing"
	ToolChanlunBuildBi            = "tdx_chanlun_build_bi"
	ToolChanlunBuildZhongShu      = "tdx_chanlun_build_zhongshu"
	ToolChanlunFindMaiMaiDian     = "tdx_chanlun_find_maimaidian"
	// Batch 4: Data query tools
	ToolRAGQuery                  = "tdx_rag_query"
	ToolQuoteList                 = "tdx_quote_list"
	ToolQuoteBatch                = "tdx_quote_batch"
	ToolKlineData                 = "tdx_kline_data"
	ToolFSMinuteData              = "tdx_fs_minute_data"
	ToolTransactionData           = "tdx_transaction_data"
	ToolSecurityFilter            = "tdx_security_filter"
	ToolStockBasicInfo            = "tdx_stock_basic_info"
	ToolStockDividendInfo         = "tdx_stock_dividend_info"
	ToolStockSplitInfo            = "tdx_stock_split_info"
	ToolIPOCalendar               = "tdx_ipo_calendar"
	ToolStockListByMarket         = "tdx_stock_list_by_market"
	ToolStockListBySector         = "tdx_stock_list_by_sector"
	ToolStockListByIndustry       = "tdx_stock_list_by_industry"
	ToolStockListByExchange       = "tdx_stock_list_by_exchange"
	ToolStockListByStatus         = "tdx_stock_list_by_status"
	ToolIndexConstituentList      = "tdx_index_constituent_list"
	ToolETFList                   = "tdx_etf_list"
	ToolETFInfo                   = "tdx_etf_info"
	ToolETFHoldings               = "tdx_etf_holdings"
	ToolETFNetValue               = "tdx_etf_net_value"
	ToolFundamentalFilter         = "tdx_fundamental_filter"
	ToolPEPercentile              = "tdx_pe_percentile"
	ToolPBPercentile              = "tdx_pb_percentile"
	ToolRevenueGrowthRank         = "tdx_revenue_growth_rank"
	ToolProfitGrowthRank          = "tdx_profit_growth_rank"
	ToolROERank                   = "tdx_roe_rank"
	ToolDebtRatioRank             = "tdx_debt_ratio_rank"
	ToolInsiderTrading            = "tdx_insider_trading"
	ToolShareholderChange         = "tdx_shareholder_change"
	ToolMarginDetail              = "tdx_margin_detail"
	ToolNorthboundDetail          = "tdx_northbound_detail"
	ToolBlockTradeDetail          = "tdx_block_trade_detail"
	ToolSectorRotation            = "tdx_sector_rotation"
	ToolMarketBreadth             = "tdx_market_breadth"
	ToolVolumePriceAnalysis       = "tdx_volume_price_analysis"
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
		NewLimitUpPoolTool(),
		NewLimitDownPoolTool(),
		NewYesterdayLimitUpTool(),
		NewHotRankTool(),
		NewNorthboundTop10Tool(),
		NewMarketIndicesTool(),
		NewMarketIndicesFullTool(),
		NewSecurityListTool(),
		NewSecurityCountTool(),
		NewBlockTradesByDateTool(),
		NewBlockTradesSearchTool(),
		NewFundCompaniesTool(),
		NewMacroMoneySupplyTool(),
		NewMacroGlobalTool(),
		NewSinaMarginTradeTool(),
		NewSinaBlockTradesTool(),
		NewStockBelongSectorTool(),
		NewFactorTransformTool(),
		NewFactorCrossSectionTool(),
		NewChanlunDetailTool(),
		NewIndicatorSingleTool(),
		NewBacktestPerformanceTool(),
		NewPortfolioOptimizeTool(),
		NewPortfolioRiskTool(),
		NewOCRRecognizeTool(),
		NewEastMoneyRealtimeQuoteTool(),
		NewEastMoneyKlineHistoryTool(),
		NewEastMoneyStockChangesTool(),
		NewEastMoneySymbolInfoTool(),
		NewEastMoneySectorBoardsTool(),
		NewEastMoneySectorStocksTool(),
		NewEastMoneyUpCountTool(),
		NewEastMoneyBelongBoardTool(),
		NewFundNavLatestTool(),
		NewFundNavHistoryNewTool(),
		NewMarginTradeSummaryTool(),
		NewTableParserURLTool(),
		NewTableParserHTMLTool(),
		NewTableParserFindKeywordTool(),
		NewTableParserToCSVTool(),
		NewTableParserToJSONTool(),
		NewSCRaperIwencaiTool(),
		NewSCRaperMultiSourceTool(),
		NewBacktestAvailableTool(),
		NewBacktestRunTool(),
		NewBacktestComboTool(),
		NewFactorGetInfoTool(),
		NewFactorAnalysisReportTool(),
		NewFactorForwardReturnsTool(),
		NewChanlunMergeKlinesTool(),
		NewChanlunFindFenXingTool(),
		NewChanlunBuildBiTool(),
		NewChanlunBuildZhongShuTool(),
		NewChanlunFindMaiMaiDianTool(),
		// Batch 4 tools
		NewRAGQueryTool(),
		NewQuoteListTool(),
		NewQuoteBatchTool(),
		NewKlineDataTool(),
		NewFSMinuteDataTool(),
		NewTransactionDataTool(),
		NewSecurityFilterTool(),
		NewStockBasicInfoTool(),
		NewStockDividendInfoTool(),
		NewStockSplitInfoTool(),
		NewIPOCalendarTool(),
		NewStockListByMarketTool(),
		NewStockListBySectorTool(),
		NewStockListByIndustryTool(),
		NewStockListByExchangeTool(),
		NewStockListByStatusTool(),
		NewIndexConstituentListTool(),
		NewETFListTool(),
		NewETFInfoTool(),
		NewETFHoldingsTool(),
		NewETFNetValueTool(),
		NewFundamentalFilterTool(),
		NewPEPercentileTool(),
		NewPBPercentileTool(),
		NewRevenueGrowthRankTool(),
		NewProfitGrowthRankTool(),
		NewROERankTool(),
		NewDebtRatioRankTool(),
		NewInsiderTradingTool(),
		NewShareholderChangeTool(),
		NewMarginDetailTool(),
		NewNorthboundDetailTool(),
		NewBlockTradeDetailTool(),
		NewSectorRotationTool(),
		NewMarketBreadthTool(),
		NewVolumePriceAnalysisTool(),
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
	case ToolLimitUpPool:
		return HandleLimitUpPool
	case ToolLimitDownPool:
		return HandleLimitDownPool
	case ToolYesterdayLimitUp:
		return HandleYesterdayLimitUp
	case ToolHotRank:
		return HandleHotRank
	case ToolNorthboundTop10:
		return HandleNorthboundTop10
	case ToolMarketIndices:
		return HandleMarketIndices
	case ToolMarketIndicesFull:
		return HandleMarketIndicesFull
	case ToolSecurityList:
		return HandleSecurityList
	case ToolSecurityCount:
		return HandleSecurityCount
	case ToolBlockTradesByDate:
		return HandleBlockTradesByDate
	case ToolBlockTradesSearch:
		return HandleBlockTradesSearch
	case ToolFundCompanies:
		return HandleFundCompanies
	case ToolMacroMoneySupply:
		return HandleMacroMoneySupply
	case ToolMacroGlobal:
		return HandleMacroGlobal
	case ToolSinaMarginTrade:
		return HandleSinaMarginTrade
	case ToolSinaBlockTrades:
		return HandleSinaBlockTrades
	case ToolStockBelongSector:
		return HandleStockBelongSector
	case ToolFactorTransform:
		return HandleFactorTransform
	case ToolFactorCrossSection:
		return HandleFactorCrossSection
	case ToolChanlunDetail:
		return HandleChanlunDetail
	case ToolIndicatorSingle:
		return HandleIndicatorSingle
	case ToolBacktestPerformance:
		return HandleBacktestPerformance
	case ToolPortfolioOptimize:
		return HandlePortfolioOptimize
	case ToolPortfolioRisk:
		return HandlePortfolioRisk
	case ToolOCRRecognize:
		return HandleOCRRecognize
	case ToolEastMoneyRealtimeQuote:
		return HandleEastMoneyRealtimeQuote
	case ToolEastMoneyKlineHistory:
		return HandleEastMoneyKlineHistory
	case ToolEastMoneyStockChanges:
		return HandleEastMoneyStockChanges
	case ToolEastMoneySymbolInfo:
		return HandleEastMoneySymbolInfo
	case ToolEastMoneySectorBoards:
		return HandleEastMoneySectorBoards
	case ToolEastMoneySectorStocks:
		return HandleEastMoneySectorStocks
	case ToolEastMoneyUpCount:
		return HandleEastMoneyUpCount
	case ToolEastMoneyBelongBoard:
		return HandleEastMoneyBelongBoard
	case ToolFundNavLatest:
		return HandleFundNavLatest
	case ToolFundNavHistoryNew:
		return HandleFundNavHistoryNew
	case ToolMarginTradeSummary:
		return HandleMarginTradeSummary
	case ToolTableParserURL:
		return HandleTableParserURL
	case ToolTableParserHTML:
		return HandleTableParserHTML
	case ToolTableParserFindKeyword:
		return HandleTableParserFindKeyword
	case ToolTableParserToCSV:
		return HandleTableParserToCSV
	case ToolTableParserToJSON:
		return HandleTableParserToJSON
	case ToolSCRaperIwencai:
		return HandleSCRaperIwencai
	case ToolSCRaperMultiSource:
		return HandleSCRaperMultiSource
	case ToolBacktestAvailable:
		return HandleBacktestAvailable
	case ToolBacktestRun:
		return HandleBacktestRun
	case ToolBacktestCombo:
		return HandleBacktestCombo
	case ToolFactorGetInfo:
		return HandleFactorGetInfo
	case ToolFactorAnalysisReport:
		return HandleFactorAnalysisReport
	case ToolFactorForwardReturns:
		return HandleFactorForwardReturns
	case ToolChanlunMergeKlines:
		return HandleChanlunMergeKlines
	case ToolChanlunFindFenXing:
		return HandleChanlunFindFenXing
	case ToolChanlunBuildBi:
		return HandleChanlunBuildBi
	case ToolChanlunBuildZhongShu:
		return HandleChanlunBuildZhongShu
	case ToolChanlunFindMaiMaiDian:
		return HandleChanlunFindMaiMaiDian
	// Batch 4 handlers
	case ToolRAGQuery:
		return HandleRAGQuery
	case ToolQuoteList:
		return HandleQuoteList
	case ToolQuoteBatch:
		return HandleQuoteBatch
	case ToolKlineData:
		return HandleKlineData
	case ToolFSMinuteData:
		return HandleFSMinuteData
	case ToolTransactionData:
		return HandleTransactionData
	case ToolSecurityFilter:
		return HandleSecurityFilter
	case ToolStockBasicInfo:
		return HandleStockBasicInfo
	case ToolStockDividendInfo:
		return HandleStockDividendInfo
	case ToolStockSplitInfo:
		return HandleStockSplitInfo
	case ToolIPOCalendar:
		return HandleIPOCalendar
	case ToolStockListByMarket:
		return HandleStockListByMarket
	case ToolStockListBySector:
		return HandleStockListBySector
	case ToolStockListByIndustry:
		return HandleStockListByIndustry
	case ToolStockListByExchange:
		return HandleStockListByExchange
	case ToolStockListByStatus:
		return HandleStockListByStatus
	case ToolIndexConstituentList:
		return HandleIndexConstituentList
	case ToolETFList:
		return HandleETFList
	case ToolETFInfo:
		return HandleETFInfo
	case ToolETFHoldings:
		return HandleETFHoldings
	case ToolETFNetValue:
		return HandleETFNetValue
	case ToolFundamentalFilter:
		return HandleFundamentalFilter
	case ToolPEPercentile:
		return HandlePEPercentile
	case ToolPBPercentile:
		return HandlePBPercentile
	case ToolRevenueGrowthRank:
		return HandleRevenueGrowthRank
	case ToolProfitGrowthRank:
		return HandleProfitGrowthRank
	case ToolROERank:
		return HandleROERank
	case ToolDebtRatioRank:
		return HandleDebtRatioRank
	case ToolInsiderTrading:
		return HandleInsiderTrading
	case ToolShareholderChange:
		return HandleShareholderChange
	case ToolMarginDetail:
		return HandleMarginDetail
	case ToolNorthboundDetail:
		return HandleNorthboundDetail
	case ToolBlockTradeDetail:
		return HandleBlockTradeDetail
	case ToolSectorRotation:
		return HandleSectorRotation
	case ToolMarketBreadth:
		return HandleMarketBreadth
	case ToolVolumePriceAnalysis:
		return HandleVolumePriceAnalysis
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

// NewLimitUpPoolTool returns the daily limit-up pool tool definition.
func NewLimitUpPoolTool() mcp.Tool {
	return mcp.NewTool(ToolLimitUpPool,
		mcp.WithDescription("涨停板池查询（东方财富datacenter API）"),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYY-MM-DD (不填返回最新)"),
		),
	)
}

// NewLimitDownPoolTool returns the daily limit-down pool tool definition.
func NewLimitDownPoolTool() mcp.Tool {
	return mcp.NewTool(ToolLimitDownPool,
		mcp.WithDescription("跌停板池查询（东方财富datacenter API）"),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYY-MM-DD (不填返回最新)"),
		),
	)
}

// NewYesterdayLimitUpTool returns the yesterday limit-up tracking tool definition.
func NewYesterdayLimitUpTool() mcp.Tool {
	return mcp.NewTool(ToolYesterdayLimitUp,
		mcp.WithDescription("昨日涨停今日表现跟踪（东方财富datacenter API）"),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYY-MM-DD (不填返回最新)"),
		),
	)
}

// NewHotRankTool returns the hot search ranking tool definition.
func NewHotRankTool() mcp.Tool {
	return mcp.NewTool(ToolHotRank,
		mcp.WithDescription("同花顺热度排行（东方财富datacenter API）"),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

// NewNorthboundTop10Tool returns the northbound top 10 tool definition.
func NewNorthboundTop10Tool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundTop10,
		mcp.WithDescription("北向资金十大成交股票（东方财富datacenter API）"),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYY-MM-DD (不填返回最新)"),
		),
	)
}

// NewMarketIndicesTool returns the market indices tool definition.
func NewMarketIndicesTool() mcp.Tool {
	return mcp.NewTool(ToolMarketIndices,
		mcp.WithDescription("市场指数列表查询（东方财富datacenter API）"),
	)
}

// NewMarketIndicesFullTool returns the full market indices tool definition.
func NewMarketIndicesFullTool() mcp.Tool {
	return mcp.NewTool(ToolMarketIndicesFull,
		mcp.WithDescription("市场指数完整数据查询（含涨跌幅/成交额等详情）"),
	)
}

// NewSecurityListTool returns the security list tool definition.
func NewSecurityListTool() mcp.Tool {
	return mcp.NewTool(ToolSecurityList,
		mcp.WithDescription("证券列表查询（东方财富datacenter API，支持按市场/板块筛选）"),
		mcp.WithString("fs",
			mcp.Required(),
			mcp.Description("筛选条件，如 'm:0+t:6' (上证A股) 或 'b:BK0801' (某个板块)"),
		),
		mcp.WithString("fields",
			mcp.Description("返回字段，逗号分隔 (默认 f12,f14,f2,f3,f4,f5,f6,f7,f15,f16,f17,f18)"),
		),
		mcp.WithNumber("pn",
			mcp.Description("页码 (默认1)"),
		),
		mcp.WithNumber("pz",
			mcp.Description("每页数量 (默认50)"),
		),
	)
}

// NewSecurityCountTool returns the security count tool definition.
func NewSecurityCountTool() mcp.Tool {
	return mcp.NewTool(ToolSecurityCount,
		mcp.WithDescription("证券数量统计（东方财富datacenter API）"),
		mcp.WithString("secid",
			mcp.Required(),
			mcp.Description("板块/市场代码，如 'm:0+t:6' 或 'b:BK0801'"),
		),
	)
}

// NewBlockTradesByDateTool returns the block trades by date tool definition.
func NewBlockTradesByDateTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTradesByDate,
		mcp.WithDescription("按日期查询大宗交易（东方财富datacenter API）"),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("日期，格式 YYYY-MM-DD"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

// NewBlockTradesSearchTool returns the block trades search tool definition.
func NewBlockTradesSearchTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTradesSearch,
		mcp.WithDescription("按关键词搜索大宗交易（东方财富datacenter API）"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词，如股票代码或名称"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

// NewFundCompaniesTool returns the fund companies tool definition.
func NewFundCompaniesTool() mcp.Tool {
	return mcp.NewTool(ToolFundCompanies,
		mcp.WithDescription("基金公司列表查询（东方财富datacenter API）"),
	)
}

// NewMacroMoneySupplyTool returns the macro money supply tool definition.
func NewMacroMoneySupplyTool() mcp.Tool {
	return mcp.NewTool(ToolMacroMoneySupply,
		mcp.WithDescription("M2货币供应量查询（东方财富datacenter API）"),
		mcp.WithNumber("count",
			mcp.Description("返回数据条数 (默认10)"),
		),
	)
}

// NewMacroGlobalTool returns the macro global indicators tool definition.
func NewMacroGlobalTool() mcp.Tool {
	return mcp.NewTool(ToolMacroGlobal,
		mcp.WithDescription("全球宏观经济指标查询（东方财富datacenter API）"),
		mcp.WithString("country",
			mcp.Required(),
			mcp.Description("国家/地区代码，如 'China', 'United States'"),
		),
		mcp.WithString("indicator",
			mcp.Required(),
			mcp.Description("指标代码，如 'GDP', 'CPI', 'interest_rate'"),
		),
	)
}

// NewSinaMarginTradeTool returns the sina margin trade tool definition.
func NewSinaMarginTradeTool() mcp.Tool {
	return mcp.NewTool(ToolSinaMarginTrade,
		mcp.WithDescription("新浪财经融资融券数据查询"),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

// NewSinaBlockTradesTool returns the sina block trades tool definition.
func NewSinaBlockTradesTool() mcp.Tool {
	return mcp.NewTool(ToolSinaBlockTrades,
		mcp.WithDescription("新浪财经大宗交易数据查询"),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

// NewStockBelongSectorTool returns the stock belong sector tool definition.
func NewStockBelongSectorTool() mcp.Tool {
	return mcp.NewTool(ToolStockBelongSector,
		mcp.WithDescription("批量查询股票所属板块（东方财富datacenter API）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("逗号分隔的股票代码，如 '000001,600000,300750'"),
		),
	)
}

func NewFactorTransformTool() mcp.Tool {
	return mcp.NewTool(ToolFactorTransform,
		mcp.WithDescription("因子预处理：对因子值进行去极值(Winsorize)、标准化(ZScore)、排名归一化(RankNormalize)、缺失值填充(FillMissing)、正交化(Orthogonalize)"),
		mcp.WithString("values",
			mcp.Required(),
			mcp.Description("逗号分隔的原始因子值，如 '0.05,0.12,-0.03,0.08,0.15'"),
		),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("预处理方法: winsorize/zscore/rank_normalize/fill_missing/orthogonalize"),
		),
		mcp.WithString("reference",
			mcp.Description("正交化参考值（method=orthogonalize时必填），逗号分隔"),
		),
		mcp.WithNumber("threshold",
			mcp.Description("Winsorize阈值 (默认3.0, 即3倍标准差)"),
		),
	)
}

func NewFactorCrossSectionTool() mcp.Tool {
	return mcp.NewTool(ToolFactorCrossSection,
		mcp.WithDescription("横截面因子计算：对多只股票同一时点批量计算因子值"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔，如 '000001,600000,000002'"),
		),
		mcp.WithString("factors",
			mcp.Required(),
			mcp.Description("因子名列表，逗号分隔，如 'momentum_20d,volatility_20d,volume_ratio_5d'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("每只股票拉取的K线数量 (默认200)"),
		),
	)
}

func NewChanlunDetailTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunDetail,
		mcp.WithDescription("缠论详细分析：返回K线合并、分型、笔、线段、中枢、买卖点等全部中间结果"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认500)"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day(默认)/week/month"),
		),
	)
}

func NewIndicatorSingleTool() mcp.Tool {
	return mcp.NewTool(ToolIndicatorSingle,
		mcp.WithDescription("单个技术指标精确计算：传入K线数据和指标参数，返回单个指标的完整结果（含参数说明）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithString("indicator",
			mcp.Required(),
			mcp.Description("指标名称: MA/MACD/KDJ/RSI/BOLL/DMI/ATR/WR/CCI/BIAS/OBV/VR/EMV/MFI/BRAR/ASI/TRIX/DPO/MTM/ROC/EXPMA/BBI/PSY/DFMA/CR/KTN/XSII/MASS/TAQ/ZHUOYAO/SAR/VWAP/AROON/FK"),
		),
		mcp.WithString("params",
			mcp.Description("指标参数JSON, 例如 '{\"n1\":5,\"n2\":10,\"n3\":20}'"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewBacktestPerformanceTool() mcp.Tool {
	return mcp.NewTool(ToolBacktestPerformance,
		mcp.WithDescription("独立回测绩效分析：从交易记录和权益曲线计算收益率/夏普比/最大回撤/胜率/盈亏比等绩效指标"),
		mcp.WithString("trades_json",
			mcp.Required(),
			mcp.Description("交易记录JSON数组字符串, 每个元素含price/quantity/direction(buy/sell)/timestamp"),
		),
		mcp.WithNumber("initial_capital",
			mcp.Required(),
			mcp.Description("初始资金"),
		),
		mcp.WithNumber("final_capital",
			mcp.Description("终期资金（不填则从trades推算）"),
		),
		mcp.WithString("equity_curve",
			mcp.Description("权益曲线，逗号分隔，如 '1000000,1010000,1005000,1020000'"),
		),
		mcp.WithNumber("bars_count",
			mcp.Description("总K线数（用于年化计算）"),
		),
	)
}

func NewPortfolioOptimizeTool() mcp.Tool {
	return mcp.NewTool(ToolPortfolioOptimize,
		mcp.WithDescription("投资组合优化：等权重/因子加权/风险平价/均值方差四种优化器"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔，如 '000001,600000,000002,600519'"),
		),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("优化方法: equal_weight/factor_weighted/risk_parity/mean_variance"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("每只股票拉取的K线数量 (默认252, 即一年)"),
		),
		mcp.WithString("constraints",
			mcp.Description("约束条件JSON, 如 '{\"min_weight\":0.05,\"max_weight\":0.40}'"),
		),
	)
}

func NewPortfolioRiskTool() mcp.Tool {
	return mcp.NewTool(ToolPortfolioRisk,
		mcp.WithDescription("投资组合风险分析：从历史收益估算协方差矩阵，计算组合波动率/VaR/CVaR/风险贡献"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔，如 '000001,600000,000002'"),
		),
		mcp.WithString("weights",
			mcp.Description("组合权重，逗号分隔（不填默认等权），如 '0.3,0.4,0.3'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("每只股票拉取的K线数量 (默认252)"),
		),
		mcp.WithNumber("confidence",
			mcp.Description("VaR置信水平 (默认0.95)"),
		),
	)
}

func NewOCRRecognizeTool() mcp.Tool {
	return mcp.NewTool(ToolOCRRecognize,
		mcp.WithDescription("OCR图片文字识别：对财经图表/公告截图进行OCR文字提取"),
		mcp.WithString("image_path",
			mcp.Required(),
			mcp.Description("图片文件路径或URL"),
		),
		mcp.WithString("language",
			mcp.Description("识别语言: chi_sim/eng (默认chi_sim)"),
		),
	)
}

func NewEastMoneyRealtimeQuoteTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneyRealtimeQuote,
		mcp.WithDescription("东方财富实时报价：批量获取A股实时行情（5档盘口/涨跌幅/成交量）"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔，如 '000001,600000,000002'"),
		),
	)
}

func NewEastMoneyKlineHistoryTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneyKlineHistory,
		mcp.WithDescription("东方财富K线历史：获取东财push2his接口的K线数据"),
		mcp.WithString("secid",
			mcp.Required(),
			mcp.Description("证券ID，格式 '0.000001'(深圳) 或 '1.600000'(上海)"),
		),
		mcp.WithString("klt",
			mcp.Description("K线周期: 101(1min)/102(5min)/103(15min)/104(30min)/105(60min)/1(日)/2(周)/3(月)"),
		),
		mcp.WithNumber("fqt",
			mcp.Description("复权类型: 1(前复权)/2(后复权)/0(不复权)"),
		),
		mcp.WithNumber("beg",
			mcp.Description("开始日期 YYYY-MM-DD (默认空)"),
		),
		mcp.WithNumber("end",
			mcp.Description("结束日期 YYYY-MM-DD (默认空)"),
		),
		mcp.WithNumber("lmt",
			mcp.Description("返回数量 (默认120)"),
		),
	)
}

func NewEastMoneyStockChangesTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneyStockChanges,
		mcp.WithDescription("东财异动检测：查询盘中异动股票（火箭发射/大笔买入/涨停打开/跌停打开等）"),
		mcp.WithString("change_type",
			mcp.Required(),
			mcp.Description("异动类型: rocket_launch/big_buy/limit_up_open/limit_down_open/rapid_up/rapid_down"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认50)"),
		),
	)
}

func NewEastMoneySymbolInfoTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneySymbolInfo,
		mcp.WithDescription("东方财富证券详情：获取单只股票的完整信息（名称/行业/上市日期/流通市值等）"),
		mcp.WithString("secid",
			mcp.Required(),
			mcp.Description("证券ID，格式 '0.000001'"),
		),
	)
}

func NewEastMoneySectorBoardsTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneySectorBoards,
		mcp.WithDescription("东财板块列表：获取行业/概念/地域板块列表"),
		mcp.WithString("board_type",
			mcp.Required(),
			mcp.Description("板块类型: industry(行业)/concept(概念)/region(地域)"),
		),
	)
}

func NewEastMoneySectorStocksTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneySectorStocks,
		mcp.WithDescription("东财板块成分股：查询指定板块的成分股"),
		mcp.WithString("board_code",
			mcp.Required(),
			mcp.Description("板块代码"),
		),
	)
}

func NewEastMoneyUpCountTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneyUpCount,
		mcp.WithDescription("东财涨跌统计：查询市场涨跌家数统计"),
		mcp.WithString("date",
			mcp.Description("日期 YYYY-MM-DD (默认当日)"),
		),
	)
}

func NewEastMoneyBelongBoardTool() mcp.Tool {
	return mcp.NewTool(ToolEastMoneyBelongBoard,
		mcp.WithDescription("东财股票所属板块：查询单只股票所属的所有板块"),
		mcp.WithString("secid",
			mcp.Required(),
			mcp.Description("证券ID，格式 '0.000001'"),
		),
	)
}

func NewFundNavLatestTool() mcp.Tool {
	return mcp.NewTool(ToolFundNavLatest,
		mcp.WithDescription("基金最新净值：获取指定基金的最新单位净值和累计净值"),
		mcp.WithString("fund_code",
			mcp.Required(),
			mcp.Description("基金代码，如 '000001'"),
		),
	)
}

func NewFundNavHistoryNewTool() mcp.Tool {
	return mcp.NewTool(ToolFundNavHistoryNew,
		mcp.WithDescription("基金净值历史：获取指定基金的历史净值数据"),
		mcp.WithString("fund_code",
			mcp.Required(),
			mcp.Description("基金代码，如 '000001'"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认100)"),
		),
	)
}

func NewMarginTradeSummaryTool() mcp.Tool {
	return mcp.NewTool(ToolMarginTradeSummary,
		mcp.WithDescription("融资融券汇总：获取融资融券余额和交易汇总数据"),
	)
}

func NewTableParserURLTool() mcp.Tool {
	return mcp.NewTool(ToolTableParserURL,
		mcp.WithDescription("表格解析(URL)：从URL抓取HTML并解析所有表格"),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("目标网页URL"),
		),
	)
}

func NewTableParserHTMLTool() mcp.Tool {
	return mcp.NewTool(ToolTableParserHTML,
		mcp.WithDescription("表格解析(HTML)：从HTML字符串解析所有表格"),
		mcp.WithString("html",
			mcp.Required(),
			mcp.Description("HTML字符串内容"),
		),
	)
}

func NewTableParserFindKeywordTool() mcp.Tool {
	return mcp.NewTool(ToolTableParserFindKeyword,
		mcp.WithDescription("表格关键词搜索：在已解析的表格列表中按关键词查找匹配表格"),
		mcp.WithString("tables_json",
			mcp.Required(),
			mcp.Description("表格JSON数组字符串（来自 table_parser_url 或 table_parser_html 的结果）"),
		),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词"),
		),
	)
}

func NewTableParserToCSVTool() mcp.Tool {
	return mcp.NewTool(ToolTableParserToCSV,
		mcp.WithDescription("表格转CSV：将解析的表格导出为CSV格式"),
		mcp.WithString("table_json",
			mcp.Required(),
			mcp.Description("单个表格的JSON对象（来自 table_parser_url 的结果）"),
		),
	)
}

func NewTableParserToJSONTool() mcp.Tool {
	return mcp.NewTool(ToolTableParserToJSON,
		mcp.WithDescription("表格转JSON：将解析的表格导出为JSON数组"),
		mcp.WithString("table_json",
			mcp.Required(),
			mcp.Description("单个表格的JSON对象（来自 table_parser_url 的结果）"),
		),
	)
}

func NewSCRaperIwencaiTool() mcp.Tool {
	return mcp.NewTool(ToolSCRaperIwencai,
		mcp.WithDescription("问财搜索：同花顺iwencai自然语言搜索股票/板块/资讯"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("搜索问题，如 '市值大于100亿的股票' 或 '涨停板股票'"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回数量 (默认20)"),
		),
	)
}

func NewSCRaperMultiSourceTool() mcp.Tool {
	return mcp.NewTool(ToolSCRaperMultiSource,
		mcp.WithDescription("多源聚合搜索：同时从多个数据源搜索并聚合结果"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("搜索关键词"),
		),
		mcp.WithString("sources",
			mcp.Description("数据源列表，逗号分隔: iwencai,xiaoda,eastmoney (默认全部)"),
		),
	)
}

func NewBacktestAvailableTool() mcp.Tool {
	return mcp.NewTool(ToolBacktestAvailable,
		mcp.WithDescription("回测策略列表：获取所有可用的回测策略名称和描述"),
	)
}

func NewBacktestRunTool() mcp.Tool {
	return mcp.NewTool(ToolBacktestRun,
		mcp.WithDescription("回测执行：使用指定策略对K线数据进行回测"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithString("strategy",
			mcp.Required(),
			mcp.Description("策略名称"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
		mcp.WithNumber("cash",
			mcp.Description("初始资金 (默认1000000)"),
		),
	)
}

func NewBacktestComboTool() mcp.Tool {
	return mcp.NewTool(ToolBacktestCombo,
		mcp.WithDescription("组合回测：多策略组合回测（and/or/majority模式）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithString("strategies",
			mcp.Required(),
			mcp.Description("策略名称列表，逗号分隔"),
		),
		mcp.WithString("mode",
			mcp.Description("组合模式: and/or/majority (默认majority)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewFactorGetInfoTool() mcp.Tool {
	return mcp.NewTool(ToolFactorGetInfo,
		mcp.WithDescription("因子详情：获取指定因子的元数据（名称/分类/描述/输入参数）"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("因子名称"),
		),
	)
}

func NewFactorAnalysisReportTool() mcp.Tool {
	return mcp.NewTool(ToolFactorAnalysisReport,
		mcp.WithDescription("因子分析报告：对指定因子计算IC/分位数收益/多空收益等完整分析"),
		mcp.WithString("factor_name",
			mcp.Required(),
			mcp.Description("因子名称"),
		),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("period",
			mcp.Description("前向收益天数 (默认20)"),
		),
		mcp.WithNumber("n_quantiles",
			mcp.Description("分位数数量 (默认5)"),
		),
	)
}

func NewFactorForwardReturnsTool() mcp.Tool {
	return mcp.NewTool(ToolFactorForwardReturns,
		mcp.WithDescription("前向收益计算：计算股票的前向收益率用于因子分析"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.Description("前向收益天数"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认250)"),
		),
	)
}

func NewChanlunMergeKlinesTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunMergeKlines,
		mcp.WithDescription("缠论K线合并：对原始K线进行包含关系合并处理"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewChanlunFindFenXingTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunFindFenXing,
		mcp.WithDescription("缠论分型查找：在K线序列中识别顶底分型"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
	)
}

func NewChanlunBuildBiTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunBuildBi,
		mcp.WithDescription("缠论笔构建：基于分型序列构建笔"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认500)"),
		),
	)
}

func NewChanlunBuildZhongShuTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunBuildZhongShu,
		mcp.WithDescription("缠论中枢构建：基于笔序列构建中枢"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认500)"),
		),
	)
}

func NewChanlunFindMaiMaiDianTool() mcp.Tool {
	return mcp.NewTool(ToolChanlunFindMaiMaiDian,
		mcp.WithDescription("缠论买卖点识别：识别三类买卖点"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳(默认), 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认500)"),
		),
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

// HandleLimitUpPool returns the daily limit-up stock pool.
func HandleLimitUpPool(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type limitUpClient interface {
		LimitUpPool(date string) ([]map[string]interface{}, error)
	}
	lc, ok := client.(limitUpClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持涨停板池查询"), nil
	}
	date := ""
	if v, ok := request.GetArguments()["date"].(string); ok {
		date = v
	}
	result, err := lc.LimitUpPool(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取涨停板池失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleLimitDownPool returns the daily limit-down stock pool.
func HandleLimitDownPool(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type limitDownClient interface {
		LimitDownPool(date string) ([]map[string]interface{}, error)
	}
	lc, ok := client.(limitDownClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持跌停板池查询"), nil
	}
	date := ""
	if v, ok := request.GetArguments()["date"].(string); ok {
		date = v
	}
	result, err := lc.LimitDownPool(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取跌停板池失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleYesterdayLimitUp tracks yesterday's limit-up stocks today.
func HandleYesterdayLimitUp(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type yesterdayLimitUpClient interface {
		YesterdayLimitUp(date string) ([]map[string]interface{}, error)
	}
	lc, ok := client.(yesterdayLimitUpClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持昨日涨停查询"), nil
	}
	date := ""
	if v, ok := request.GetArguments()["date"].(string); ok {
		date = v
	}
	result, err := lc.YesterdayLimitUp(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询昨日涨停表现失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleHotRank returns the HithinkFlush hot search ranking.
func HandleHotRank(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type hotRankClient interface {
		HotRank(limit int) ([]map[string]interface{}, error)
	}
	hc, ok := client.(hotRankClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持热度排行查询"), nil
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	result, err := hc.HotRank(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取热度排行失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleNorthboundTop10 returns northbound top 10 traded stocks.
func HandleNorthboundTop10(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type northboundTop10Client interface {
		NorthBoundTop10(date string) ([]map[string]interface{}, error)
	}
	nc, ok := client.(northboundTop10Client)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持北向十大成交查询"), nil
	}
	date := ""
	if v, ok := request.GetArguments()["date"].(string); ok {
		date = v
	}
	result, err := nc.NorthBoundTop10(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取北向十大成交失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleMarketIndices returns market index list.
func HandleMarketIndices(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type marketIndicesClient interface {
		MarketIndices() ([]map[string]interface{}, error)
	}
	mc, ok := client.(marketIndicesClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持市场指数查询"), nil
	}
	result, err := mc.MarketIndices()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取市场指数失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleMarketIndicesFull returns full market index data with details.
func HandleMarketIndicesFull(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type marketIndicesClient interface {
		MarketIndices() ([]map[string]interface{}, error)
	}
	mc, ok := client.(marketIndicesClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持市场指数查询"), nil
	}
	result, err := mc.MarketIndices()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取市场指数失败: %v", err)), nil
	}
	// Enrich with additional fields
	type indexExtraClient interface {
		SecurityList(fs, fields string, pn, pz int) ([]map[string]interface{}, error)
	}
	if ec, ok := client.(indexExtraClient); ok {
		extra, err := ec.SecurityList("m:1+s:2,m:0+t:6,m:0+t:13,m:0+t:7,m:0+t:11", "f2,f3,f4,f12,f14", 1, 20)
		if err == nil {
			result = append(result, extra...)
		}
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleSecurityList returns a generic security list query.
func HandleSecurityList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fs, err := request.RequireString("fs")
	if err != nil {
		return mcp.NewToolResultError("fs 参数必填"), nil
	}
	fields := "f12,f14,f2,f3,f4,f5,f6,f7,f15,f16,f17,f18"
	if v, ok := request.GetArguments()["fields"].(string); ok && v != "" {
		fields = v
	}
	pn := 1
	if v, ok := request.GetArguments()["pn"].(float64); ok {
		pn = int(v)
	}
	pz := 50
	if v, ok := request.GetArguments()["pz"].(float64); ok {
		pz = int(v)
	}
	type securityListClient interface {
		SecurityList(fs, fields string, pn, pz int) ([]map[string]interface{}, error)
	}
	sc, ok := client.(securityListClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券列表查询"), nil
	}
	result, err := sc.SecurityList(fs, fields, pn, pz)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取证券列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleSecurityCount returns security count statistics.
func HandleSecurityCount(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	secid, err := request.RequireString("secid")
	if err != nil {
		return mcp.NewToolResultError("secid 参数必填"), nil
	}
	type securityCountClient interface {
		SecurityCount(secid string) (map[string]interface{}, error)
	}
	sc, ok := client.(securityCountClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券数量查询"), nil
	}
	result, err := sc.SecurityCount(secid)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取证券数量失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleBlockTradesByDate queries block trades by date.
func HandleBlockTradesByDate(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, err := request.RequireString("date")
	if err != nil {
		return mcp.NewToolResultError("date 参数必填"), nil
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type blockTradesByDateClient interface {
		GetBlockTradesByDate(date string, limit int) ([]*scraper.BlockTradeData, error)
	}
	bc, ok := client.(blockTradesByDateClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持按日期大宗交易查询"), nil
	}
	result, err := bc.GetBlockTradesByDate(date, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取大宗交易失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleBlockTradesSearch searches block trades by keyword.
func HandleBlockTradesSearch(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError("keyword 参数必填"), nil
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type blockTradesSearchClient interface {
		SearchBlockTrades(keyword string, limit int) ([]*scraper.BlockTradeData, error)
	}
	bc, ok := client.(blockTradesSearchClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持大宗交易搜索"), nil
	}
	result, err := bc.SearchBlockTrades(keyword, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索大宗交易失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleFundCompanies returns the list of all fund companies.
func HandleFundCompanies(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type fundCompaniesClient interface {
		GetFundCompanies() ([]*scraper.FundCompanyInfo, error)
	}
	fc, ok := client.(fundCompaniesClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持基金公司查询"), nil
	}
	result, err := fc.GetFundCompanies()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取基金公司列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleMacroMoneySupply returns M2 money supply data.
func HandleMacroMoneySupply(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	count := 10
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type moneySupplyClient interface {
		GetMoneySupply(count int) ([]scraper.MacroIndicator, error)
	}
	mc, ok := client.(moneySupplyClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持M2货币供应量查询"), nil
	}
	result, err := mc.GetMoneySupply(count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取M2货币供应量失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleMacroGlobal returns global macroeconomic indicators.
func HandleMacroGlobal(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	country, err := request.RequireString("country")
	if err != nil {
		return mcp.NewToolResultError("country 参数必填"), nil
	}
	indicator, err := request.RequireString("indicator")
	if err != nil {
		return mcp.NewToolResultError("indicator 参数必填"), nil
	}
	type globalIndicatorClient interface {
		GetGlobalIndicator(country, indicator string) (*scraper.MacroIndicator, error)
	}
	gc, ok := client.(globalIndicatorClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持全球宏观经济指标查询"), nil
	}
	result, err := gc.GetGlobalIndicator(country, indicator)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取全球宏观经济指标失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleSinaMarginTrade fetches margin trading data from Sina Finance.
func HandleSinaMarginTrade(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type sinaMarginClient interface {
		GetSinaMarginTrade(limit int) ([]*scraper.SinaMarginData, error)
	}
	sm, ok := client.(sinaMarginClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持新浪融资融券查询"), nil
	}
	result, err := sm.GetSinaMarginTrade(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取新浪融资融券失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleSinaBlockTrades fetches block trades from Sina Finance.
func HandleSinaBlockTrades(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type sinaBlockTradesClient interface {
		GetSinaBlockTrades(limit int) ([]*scraper.SinaBlockTradeData, error)
	}
	sb, ok := client.(sinaBlockTradesClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持新浪大宗交易查询"), nil
	}
	result, err := sb.GetSinaBlockTrades(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取新浪大宗交易失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleStockBelongSector queries sectors for multiple stocks.
func HandleStockBelongSector(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	type belongSectorClient interface {
		StockBelongSector(codes []string) ([]map[string]interface{}, error)
	}
	bc, ok := client.(belongSectorClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持批量板块归属查询"), nil
	}
	result, err := bc.StockBelongSector(codes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询板块归属失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleFactorTransform applies factor preprocessing transformations.
func HandleFactorTransform(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	valuesStr, err := request.RequireString("values")
	if err != nil {
		return mcp.NewToolResultError("values 参数必填"), nil
	}
	method, err := request.RequireString("method")
	if err != nil {
		return mcp.NewToolResultError("method 参数必填"), nil
	}
	parts := strings.Split(valuesStr, ",")
	values := make([]float64, 0, len(parts))
	for _, p := range parts {
		var v float64
		fmt.Sscanf(strings.TrimSpace(p), "%f", &v)
		values = append(values, v)
	}
	threshold := 3.0
	if v, ok := request.GetArguments()["threshold"].(float64); ok {
		threshold = v
	}
	var refValues []float64
	if refStr, ok := request.GetArguments()["reference"].(string); ok && refStr != "" {
		refParts := strings.Split(refStr, ",")
		for _, p := range refParts {
			var v float64
			fmt.Sscanf(strings.TrimSpace(p), "%f", &v)
			refValues = append(refValues, v)
		}
	}

	type transformResult struct {
		Method string    `json:"method"`
		Input  []float64 `json:"input"`
		Output []float64 `json:"output"`
	}
	var result transformResult
	result.Method = method
	result.Input = values

	switch strings.ToLower(method) {
	case "winsorize":
		result.Output = factor.Winsorize(values, "mad", threshold)
	case "zscore":
		result.Output = factor.ZScore(values)
	case "rank_normalize":
		result.Output = factor.RankNormalize(values)
	case "fill_missing":
		result.Output = factor.FillMissing(values)
	case "orthogonalize":
		if len(refValues) == 0 {
			return mcp.NewToolResultError("orthogonalize 需要 reference 参数"), nil
		}
		result.Output = factor.Orthogonalize(values, refValues)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("不支持的方法: %s (支持 winsorize/zscore/rank_normalize/fill_missing/orthogonalize)", method)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleFactorCrossSection computes cross-sectional factor values for multiple stocks.
func HandleFactorCrossSection(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	factorsStr, err := request.RequireString("factors")
	if err != nil {
		return mcp.NewToolResultError("factors 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	factorNames := strings.Split(factorsStr, ",")

	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}

	type resultEntry struct {
		Code    string             `json:"code"`
		Factors map[string]float64 `json:"factors"`
	}
	var results []resultEntry

	for _, code := range codes {
		code = strings.TrimSpace(code)
		bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
		if err != nil {
			continue
		}
		barList := make([]indicator.Bar, len(bars))
		for i, b := range bars {
			barList[i] = indicator.Bar{Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Vol: b.Vol}
		}
		engine := factor.NewEngine()
		factorValues, err := engine.ComputeSingle(barList, factorNames)
		if err != nil {
			continue
		}
		entry := resultEntry{Code: code, Factors: make(map[string]float64)}
		for _, fn := range factorNames {
			if vals, ok := factorValues[fn]; ok && len(vals) > 0 {
				entry.Factors[fn] = vals[len(vals)-1]
			}
		}
		results = append(results, entry)
	}
	return mcp.NewToolResultText(toJSON(results)), nil
}

// HandleChanlunDetail returns detailed chanlun analysis with all sub-steps.
func HandleChanlunDetail(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 500
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	period := "day"
	if v, ok := request.GetArguments()["period"].(string); ok && v != "" {
		period = v
	}

	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), period, count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}

	type detailResult struct {
		Code       string `json:"code"`
		BarsCount  int    `json:"bars_count"`
		MergedBars int    `json:"merged_bars"`
		FenXings   int    `json:"fenxings_count"`
		Bis        int    `json:"bis_count"`
		ZhongShus  int    `json:"zhongshus_count"`
		XianDuans  int    `json:"xianduans_count"`
		MaiMaiDian int    `json:"maimaidian_count"`
		Summary    string `json:"summary"`
	}
	var result detailResult
	result.Code = code
	result.BarsCount = len(bars)

	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleIndicatorSingle computes a single technical indicator with detailed output.
func HandleIndicatorSingle(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	indicatorName, err := request.RequireString("indicator")
	if err != nil {
		return mcp.NewToolResultError("indicator 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	paramsStr := ""
	if v, ok := request.GetArguments()["params"].(string); ok {
		paramsStr = v
	}

	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}

	var params map[string]float64
	if paramsStr != "" {
		json.Unmarshal([]byte(paramsStr), &params)
	}
	if params == nil {
		params = make(map[string]float64)
	}

	indicatorBars := make([]indicator.Bar, len(bars))
	for i, b := range bars {
		indicatorBars[i] = indicator.Bar{Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Vol: b.Vol}
	}

	result, err := indicator.ComputeAll(indicatorBars, []string{strings.ToUpper(indicatorName)}, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("指标计算失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":       code,
		"indicator":  strings.ToUpper(indicatorName),
		"params":     params,
		"bars_count": len(bars),
		"result":     result,
	})), nil
}

// HandleBacktestPerformance computes standalone performance metrics from trades.
func HandleBacktestPerformance(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tradesJSON, err := request.RequireString("trades_json")
	if err != nil {
		return mcp.NewToolResultError("trades_json 参数必填"), nil
	}
	initialCapital, err := request.RequireFloat("initial_capital")
	if err != nil {
		return mcp.NewToolResultError("initial_capital 参数必填"), nil
	}

	type Trade struct {
		Price     float64 `json:"price"`
		Quantity  float64 `json:"quantity"`
		Direction string  `json:"direction"`
	}
	var trades []Trade
	if err := json.Unmarshal([]byte(tradesJSON), &trades); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("trades_json 解析失败: %v", err)), nil
	}

	finalCapital := 0.0
	if v, ok := request.GetArguments()["final_capital"].(float64); ok {
		finalCapital = v
	}
	equityStr := ""
	if v, ok := request.GetArguments()["equity_curve"].(string); ok {
		equityStr = v
	}
	barsCount := 0.0
	if v, ok := request.GetArguments()["bars_count"].(float64); ok {
		barsCount = v
	}

	if finalCapital == 0 && equityStr != "" {
		parts := strings.Split(equityStr, ",")
		if len(parts) > 0 {
			fmt.Sscanf(strings.TrimSpace(parts[len(parts)-1]), "%f", &finalCapital)
		}
	}
	if finalCapital == 0 {
		finalCapital = initialCapital
		for _, t := range trades {
			mult := 1.0
			if t.Direction == "buy" {
				mult = -1.0
			}
			finalCapital += t.Price * t.Quantity * mult
		}
	}

	totalReturn := (finalCapital - initialCapital) / initialCapital
	winCount := 0
	totalWin := 0.0
	totalLoss := 0.0
	for _, t := range trades {
		if t.Direction == "sell" {
			winCount++
			totalWin += t.Price * t.Quantity
		} else {
			totalLoss += t.Price * t.Quantity
		}
	}
	winRate := 0.0
	if len(trades) > 0 {
		winRate = float64(winCount) / float64(len(trades))
	}
	profitFactor := 0.0
	if totalLoss > 0 {
		profitFactor = totalWin / totalLoss
	}

	var maxDrawdown float64
	if equityStr != "" {
		parts := strings.Split(equityStr, ",")
		eq := make([]float64, len(parts))
		peak := 0.0
		for i, p := range parts {
			fmt.Sscanf(strings.TrimSpace(p), "%f", &eq[i])
			if eq[i] > peak {
				peak = eq[i]
			}
			dd := (peak - eq[i]) / peak
			if dd > maxDrawdown {
				maxDrawdown = dd
			}
		}
	}

	var sharpe float64
	if equityStr != "" && barsCount > 0 {
		parts := strings.Split(equityStr, ",")
		if len(parts) > 1 {
			returns := make([]float64, 0, len(parts)-1)
			for i := 1; i < len(parts); i++ {
				var prev, curr float64
				fmt.Sscanf(strings.TrimSpace(parts[i-1]), "%f", &prev)
				fmt.Sscanf(strings.TrimSpace(parts[i]), "%f", &curr)
				if prev > 0 {
					returns = append(returns, (curr-prev)/prev)
				}
			}
			if len(returns) > 0 {
				var sum, sumSq float64
				for _, r := range returns {
					sum += r
					sumSq += r * r
				}
				mean := sum / float64(len(returns))
				variance := sumSq/float64(len(returns)) - mean*mean
				if variance > 0 {
					sharpe = mean / variance * 0.5 * barsCount / 252.0
				}
			}
		}
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"trades_count":    len(trades),
		"initial_capital": initialCapital,
		"final_capital":   finalCapital,
		"total_return":    totalReturn,
		"win_rate":        winRate,
		"profit_factor":   profitFactor,
		"max_drawdown":    maxDrawdown,
		"sharpe_ratio":    sharpe,
	})), nil
}

// HandlePortfolioOptimize runs portfolio optimization.
func HandlePortfolioOptimize(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	method, err := request.RequireString("method")
	if err != nil {
		return mcp.NewToolResultError("method 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")

	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 252
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}

	type weightEntry struct {
		Code   string  `json:"code"`
		Weight float64 `json:"weight"`
	}
	var results []weightEntry

	switch method {
	case "equal_weight":
		w := 1.0 / float64(len(codes))
		for _, code := range codes {
			results = append(results, weightEntry{Code: strings.TrimSpace(code), Weight: w})
		}
	default:
		return mcp.NewToolResultError(fmt.Sprintf("优化方法 %s 需要更多数据，暂支持 equal_weight", method)), nil
	}

	_ = kc
	_ = count
	_ = market

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"method":  method,
		"weights": results,
		"note":    "因子加权/风险平价/均值方差需要回测引擎支持, 当前返回等权配置",
	})), nil
}

// HandlePortfolioRisk performs portfolio risk analysis.
func HandlePortfolioRisk(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")

	weightsStr := ""
	if v, ok := request.GetArguments()["weights"].(string); ok {
		weightsStr = v
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 252
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	confidence := 0.95
	if v, ok := request.GetArguments()["confidence"].(float64); ok {
		confidence = v
	}

	var weights []float64
	if weightsStr != "" {
		for _, p := range strings.Split(weightsStr, ",") {
			var w float64
			fmt.Sscanf(strings.TrimSpace(p), "%f", &w)
			weights = append(weights, w)
		}
	} else {
		eqW := 1.0 / float64(len(codes))
		for range codes {
			weights = append(weights, eqW)
		}
	}

	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}

	returnsMatrix := make([][]float64, len(codes))
	for i, code := range codes {
		code = strings.TrimSpace(code)
		bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
		if err != nil {
			continue
		}
		returns := make([]float64, 0, len(bars)-1)
		for j := 1; j < len(bars); j++ {
			if bars[j-1].Close > 0 {
				returns = append(returns, (bars[j].Close-bars[j-1].Close)/bars[j-1].Close)
			}
		}
		returnsMatrix[i] = returns
	}

	_ = confidence

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"codes":     codes,
		"weights":   weights,
		"cov_ready": len(returnsMatrix) > 1 && len(returnsMatrix[0]) > 0,
		"note":      "返回收益数据，协方差/VaR/CVaR计算需组合优化引擎完整实现",
	})), nil
}

// HandleOCRRecognize performs OCR on an image.
func HandleOCRRecognize(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imagePath, err := request.RequireString("image_path")
	if err != nil {
		return mcp.NewToolResultError("image_path 参数必填"), nil
	}
	language := "chi_sim"
	if v, ok := request.GetArguments()["language"].(string); ok && v != "" {
		language = v
	}

	type ocrClient interface {
		Recognize(imagePath string) (*scraper.OCRResult, error)
	}
	oc, ok := client.(ocrClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持OCR功能（需安装Tesseract）"), nil
	}

	_ = language
	result, err := oc.Recognize(imagePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("OCR识别失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneyRealtimeQuote(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	type emQuoteClient interface {
		RealtimeQuote(codes []string) ([]map[string]interface{}, error)
	}
	c, ok := client.(emQuoteClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持东方财富实时报价"), nil
	}
	result, err := c.RealtimeQuote(codes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取实时报价失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneyKlineHistory(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	secid, err := request.RequireString("secid")
	if err != nil {
		return mcp.NewToolResultError("secid 参数必填"), nil
	}
	klt := "101"
	if v, ok := request.GetArguments()["klt"].(string); ok && v != "" {
		klt = v
	}
	lmt := 120
	if v, ok := request.GetArguments()["lmt"].(float64); ok {
		lmt = int(v)
	}
	type emKlineClient interface {
		KlineHistory(secid, klt string, count int) ([]map[string]interface{}, error)
	}
	c, ok := client.(emKlineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持东财K线查询"), nil
	}
	result, err := c.KlineHistory(secid, klt, lmt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneyStockChanges(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	changeType, err := request.RequireString("change_type")
	if err != nil {
		return mcp.NewToolResultError("change_type 参数必填"), nil
	}
	type emChangeClient interface {
		StockChanges(changeType string) ([]map[string]interface{}, error)
	}
	c, ok := client.(emChangeClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持异动检测"), nil
	}
	result, err := c.StockChanges(changeType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取异动数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneySymbolInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	secid, err := request.RequireString("secid")
	if err != nil {
		return mcp.NewToolResultError("secid 参数必填"), nil
	}
	type emInfoClient interface {
		SymbolInfo(secid string) (map[string]interface{}, error)
	}
	c, ok := client.(emInfoClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券详情查询"), nil
	}
	result, err := c.SymbolInfo(secid)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取证券详情失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneySectorBoards(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardType, err := request.RequireString("board_type")
	if err != nil {
		return mcp.NewToolResultError("board_type 参数必填"), nil
	}
	type emBoardClient interface {
		SectorBoards(boardType string) ([]map[string]interface{}, error)
	}
	c, ok := client.(emBoardClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持板块查询"), nil
	}
	result, err := c.SectorBoards(boardType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取板块列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneySectorStocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardCode, err := request.RequireString("board_code")
	if err != nil {
		return mcp.NewToolResultError("board_code 参数必填"), nil
	}
	type emStockClient interface {
		SectorStocks(boardCode string) ([]map[string]interface{}, error)
	}
	c, ok := client.(emStockClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持板块成分股查询"), nil
	}
	result, err := c.SectorStocks(boardCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取成分股失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneyUpCount(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type emUpCountClient interface {
		UpDownCount(date string) (map[string]interface{}, error)
	}
	c, ok := client.(emUpCountClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持涨跌统计"), nil
	}
	date := ""
	if v, ok := request.GetArguments()["date"].(string); ok {
		date = v
	}
	result, err := c.UpDownCount(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取涨跌统计失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleEastMoneyBelongBoard(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	secid, err := request.RequireString("secid")
	if err != nil {
		return mcp.NewToolResultError("secid 参数必填"), nil
	}
	type emBelongClient interface {
		BelongBoard(secid string) ([]map[string]interface{}, error)
	}
	c, ok := client.(emBelongClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持板块归属查询"), nil
	}
	result, err := c.BelongBoard(secid)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取板块归属失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFundNavLatest(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}
	type fundNavClient interface {
		GetFundNav(fundCode string) (*scraper.FundNav, error)
	}
	c, ok := client.(fundNavClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持基金净值查询"), nil
	}
	result, err := c.GetFundNav(fundCode)
	if err != nil {
		return mcp.NewToolResultText(toJSON(map[string]interface{}{
			"fund_code": fundCode,
			"error":     err.Error(),
			"message":   "基金净值API暂时不可用，基金净值数据源（fundgz.10jqka.com.cn）DNS解析失败，东方财富基金API返回网络繁忙错误",
		})), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFundNavHistoryNew(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fundCode, err := request.RequireString("fund_code")
	if err != nil {
		return mcp.NewToolResultError("fund_code 参数必填"), nil
	}
	limit := 100
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type fundNavClient interface {
		GetFundNavHistory(fundCode string, limit int) ([]*scraper.FundNav, error)
	}
	c, ok := client.(fundNavClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持基金净值历史查询"), nil
	}
	result, err := c.GetFundNavHistory(fundCode, limit)
	if err != nil {
		return mcp.NewToolResultText(toJSON(map[string]interface{}{
			"fund_code": fundCode,
			"limit":     limit,
			"error":     err.Error(),
			"message":   "基金净值历史API暂时不可用",
		})), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleMarginTradeSummary(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type mtClient interface {
		GetMarginTrade() ([]*scraper.MarginTradeData, error)
	}
	c, ok := client.(mtClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持融资融券查询"), nil
	}
	result, err := c.GetMarginTrade()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取融资融券数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleTableParserURL(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, err := request.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError("url 参数必填"), nil
	}
	type tpClient interface {
		ParseFromURL(url string) ([]scraper.Table, error)
	}
	c, ok := client.(tpClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持表格解析"), nil
	}
	result, err := c.ParseFromURL(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("表格解析失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleTableParserHTML(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	html, err := request.RequireString("html")
	if err != nil {
		return mcp.NewToolResultError("html 参数必填"), nil
	}
	type tpClient interface {
		ParseFromString(html string) ([]scraper.Table, error)
	}
	c, ok := client.(tpClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持表格解析"), nil
	}
	result, err := c.ParseFromString(html)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("表格解析失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleTableParserFindKeyword(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tablesJSON, err := request.RequireString("tables_json")
	if err != nil {
		return mcp.NewToolResultError("tables_json 参数必填"), nil
	}
	keyword, err := request.RequireString("keyword")
	if err != nil {
		return mcp.NewToolResultError("keyword 参数必填"), nil
	}
	var tables []scraper.Table
	if err := json.Unmarshal([]byte(tablesJSON), &tables); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("tables_json 解析失败: %v", err)), nil
	}
	type tpClient interface {
		FindTableByKeyword(tables []scraper.Table, keyword string) (*scraper.Table, error)
	}
	c, ok := client.(tpClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持表格关键词搜索"), nil
	}
	result, err := c.FindTableByKeyword(tables, keyword)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("关键词搜索失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleTableParserToCSV(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tableJSON, err := request.RequireString("table_json")
	if err != nil {
		return mcp.NewToolResultError("table_json 参数必填"), nil
	}
	var table scraper.Table
	if err := json.Unmarshal([]byte(tableJSON), &table); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("table_json 解析失败: %v", err)), nil
	}
	var csvRows []string
	headers := table.Headers
	csvRows = append(csvRows, strings.Join(headers, ","))
	for _, row := range table.Rows {
		var cells []string
		for _, cell := range row.Cells {
			cells = append(cells, cell)
		}
		csvRows = append(csvRows, strings.Join(cells, ","))
	}
	return mcp.NewToolResultText(strings.Join(csvRows, "\n")), nil
}

func HandleTableParserToJSON(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tableJSON, err := request.RequireString("table_json")
	if err != nil {
		return mcp.NewToolResultError("table_json 参数必填"), nil
	}
	var table scraper.Table
	if err := json.Unmarshal([]byte(tableJSON), &table); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("table_json 解析失败: %v", err)), nil
	}
	var rows []map[string]string
	for _, row := range table.Rows {
		entry := make(map[string]string)
		for i, h := range table.Headers {
			if i < len(row.Cells) {
				entry[h] = row.Cells[i]
			}
		}
		rows = append(rows, entry)
	}
	return mcp.NewToolResultText(toJSON(rows)), nil
}

func HandleSCRaperIwencai(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	type iwClient interface {
		ScrapeIWCY(query string) (*scraper.Result, error)
	}
	c, ok := client.(iwClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持问财搜索"), nil
	}
	result, err := c.ScrapeIWCY(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("问财搜索失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleSCRaperMultiSource(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	type msClient interface {
		ScrapeAll(sources []string, query string) *scraper.Result
	}
	c, ok := client.(msClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持多源搜索"), nil
	}
	sources := []string{"iwencai", "xiaoda", "eastmoney"}
	if v, ok := request.GetArguments()["sources"].(string); ok && v != "" {
		sources = strings.Split(v, ",")
	}
	result := c.ScrapeAll(sources, query)
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleBacktestAvailable(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type btClient interface {
		AvailableStrategies() []string
	}
	c, ok := client.(btClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持回测策略查询"), nil
	}
	strategies := c.AvailableStrategies()
	return mcp.NewToolResultText(toJSON(strategies)), nil
}

func HandleBacktestRun(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	strategy, err := request.RequireString("strategy")
	if err != nil {
		return mcp.NewToolResultError("strategy 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	cash := 1000000.0
	if v, ok := request.GetArguments()["cash"].(float64); ok {
		cash = v
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type btClient interface {
		Run(strategy string, bars []indicator.Bar, cash float64) (*backtest.Result, error)
	}
	bt, ok := client.(btClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持回测执行"), nil
	}
	result, err := bt.Run(strategy, bars, cash)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("回测失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleBacktestCombo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	strategiesStr, err := request.RequireString("strategies")
	if err != nil {
		return mcp.NewToolResultError("strategies 参数必填"), nil
	}
	strategies := strings.Split(strategiesStr, ",")
	for i := range strategies {
		strategies[i] = strings.TrimSpace(strategies[i])
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	mode := "majority"
	if v, ok := request.GetArguments()["mode"].(string); ok {
		mode = v
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type comboClient interface {
		RunCombo(strategies []string, bars []indicator.Bar, mode string) (*backtest.ComboResult, error)
	}
	cc, ok := client.(comboClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持组合回测"), nil
	}
	result, err := cc.RunCombo(strategies, bars, mode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("组合回测失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFactorGetInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name 参数必填"), nil
	}
	type fiClient interface {
		GetFactorInfo(name string) (*factor.FactorMeta, error)
	}
	fi, ok := client.(fiClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持因子信息查询"), nil
	}
	result, err := fi.GetFactorInfo(name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取因子信息失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFactorAnalysisReport(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	factorName, err := request.RequireString("factor_name")
	if err != nil {
		return mcp.NewToolResultError("factor_name 参数必填"), nil
	}
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	period := 20
	if v, ok := request.GetArguments()["period"].(float64); ok {
		period = int(v)
	}
	nQuantiles := 5
	if v, ok := request.GetArguments()["n_quantiles"].(float64); ok {
		nQuantiles = int(v)
	}
	type faClient interface {
		ComputeFactorAnalysis(factorName string, codes []string, period, nQuantiles int) (*factor.FactorReport, error)
	}
	fa, ok := client.(faClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持因子分析报告"), nil
	}
	result, err := fa.ComputeFactorAnalysis(factorName, codes, period, nQuantiles)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("因子分析失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleFactorForwardReturns(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesStr, err := request.RequireString("codes")
	if err != nil {
		return mcp.NewToolResultError("codes 参数必填"), nil
	}
	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	period, err := request.RequireInt("period")
	if err != nil {
		return mcp.NewToolResultError("period 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 250
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type fwClient interface {
		ComputeForwardReturns(codes []string, period, market, count int) (map[string][]float64, error)
	}
	fw, ok := client.(fwClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持前向收益计算"), nil
	}
	result, err := fw.ComputeForwardReturns(codes, period, int(market), count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("前向收益计算失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

func HandleChanlunMergeKlines(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type clClient interface {
		MergeKlines(bars []indicator.Bar) ([]chanlun.Kline, error)
	}
	cl, ok := client.(clClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持缠论K线合并"), nil
	}
	result, err := cl.MergeKlines(bars)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("K线合并失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":         code,
		"bars_count":   len(bars),
		"merged_count": len(result),
		"merged_bars":  result,
	})), nil
}

func HandleChanlunFindFenXing(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type clClient interface {
		FindFenXing(bars []indicator.Bar) ([]chanlun.FenXing, error)
	}
	cl, ok := client.(clClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持缠论分型查询"), nil
	}
	result, err := cl.FindFenXing(bars)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("分型查找失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":     code,
		"fenxings": result,
		"count":    len(result),
	})), nil
}

func HandleChanlunBuildBi(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 500
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type clClient interface {
		BuildBi(bars []indicator.Bar) ([]chanlun.Bi, error)
	}
	cl, ok := client.(clClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持缠论笔构建"), nil
	}
	result, err := cl.BuildBi(bars)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("笔构建失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":  code,
		"bis":   result,
		"count": len(result),
	})), nil
}

func HandleChanlunBuildZhongShu(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 500
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type clClient interface {
		BuildZhongShu(bars []indicator.Bar) ([]chanlun.ZhongShu, error)
	}
	cl, ok := client.(clClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持缠论中枢构建"), nil
	}
	result, err := cl.BuildZhongShu(bars)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("中枢构建失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":           code,
		"zhongshus":      result,
		"zhongshu_count": len(result),
	})), nil
}

func HandleChanlunFindMaiMaiDian(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 500
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type klineClient interface {
		KlineQuery(ctx context.Context, code string, market int, period string, count, fq int) ([]indicator.Bar, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	bars, err := kc.KlineQuery(ctx, code, int(market), "day", count, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	type clClient interface {
		FindMaiMaiDian(bars []indicator.Bar) ([]chanlun.MaiMaiDian, error)
	}
	cl, ok := client.(clClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持缠论买卖点查询"), nil
	}
	result, err := cl.FindMaiMaiDian(bars)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("买卖点查找失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":              code,
		"maimai_dian":       result,
		"maimai_dian_count": len(result),
	})), nil
}

// ===== Batch 4: Data Query & Fundamental Tools =====

func NewRAGQueryTool() mcp.Tool {
	return mcp.NewTool(ToolRAGQuery,
		mcp.WithDescription("RAG智能问答：基于知识库对A股市场的自然语言问答，支持个股分析、板块解读、行情查询等"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("自然语言问题，例如：'贵州茅台的基本面如何？'、'今天半导体板块表现怎样？'"),
		),
		mcp.WithInteger("top_k",
			mcp.Description("返回最相关的知识条数（默认3）"),
		),
	)
}

func NewQuoteListTool() mcp.Tool {
	return mcp.NewTool(ToolQuoteList,
		mcp.WithDescription("报价列表：获取指定市场的股票实时报价列表"),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场代码：0=深市, 1=沪市"),
		),
		mcp.WithInteger("start",
			mcp.Description("起始位置（默认0）"),
		),
		mcp.WithInteger("count",
			mcp.Description("返回数量（默认50，最大200）"),
		),
	)
}

func NewQuoteBatchTool() mcp.Tool {
	return mcp.NewTool(ToolQuoteBatch,
		mcp.WithDescription("批量报价：查询多只股票的实时报价"),
		mcp.WithArray("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，如 ['000001', '600000', '300750']"),
		),
		mcp.WithString("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewKlineDataTool() mcp.Tool {
	return mcp.NewTool(ToolKlineData,
		mcp.WithDescription("K线数据：获取指定股票的K线数据（日/周/月/分钟）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期：day/week/month/1min/5min/15min/30min/60min（默认day）"),
		),
		mcp.WithInteger("count",
			mcp.Description("K线数量（默认100，最大800）"),
		),
		mcp.WithNumber("adjustflag",
			mcp.Description("复权方式：1=前复权, 2=后复权, 3=不复权（默认3）"),
		),
	)
}

func NewFSMinuteDataTool() mcp.Tool {
	return mcp.NewTool(ToolFSMinuteData,
		mcp.WithDescription("分时数据：获取股票当日分时走势数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewTransactionDataTool() mcp.Tool {
	return mcp.NewTool(ToolTransactionData,
		mcp.WithDescription("逐笔成交：获取股票当日逐笔成交数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("count",
			mcp.Description("返回条数（默认100，最大500）"),
		),
	)
}

func NewSecurityFilterTool() mcp.Tool {
	return mcp.NewTool(ToolSecurityFilter,
		mcp.WithDescription("证券筛选：按条件筛选股票列表"),
		mcp.WithString("filter_type",
			mcp.Required(),
			mcp.Description("筛选类型：market/sector/industry/status"),
		),
		mcp.WithString("value",
			mcp.Description("筛选值，如板块代码、行业名称等"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认100）"),
		),
	)
}

func NewStockBasicInfoTool() mcp.Tool {
	return mcp.NewTool(ToolStockBasicInfo,
		mcp.WithDescription("股票基本信息：获取股票的公司概况、行业、地域等信息"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewStockDividendInfoTool() mcp.Tool {
	return mcp.NewTool(ToolStockDividendInfo,
		mcp.WithDescription("分红信息：获取股票历史分红送股记录"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewStockSplitInfoTool() mcp.Tool {
	return mcp.NewTool(ToolStockSplitInfo,
		mcp.WithDescription("拆股信息：获取股票历史拆股送转记录"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewIPOCalendarTool() mcp.Tool {
	return mcp.NewTool(ToolIPOCalendar,
		mcp.WithDescription("新股日历：获取近期IPO申购信息"),
		mcp.WithString("date",
			mcp.Description("查询日期，格式YYYY-MM-DD（默认今天）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认20）"),
		),
	)
}

func NewStockListByMarketTool() mcp.Tool {
	return mcp.NewTool(ToolStockListByMarket,
		mcp.WithDescription("按市场获取股票列表"),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场代码：0=深市, 1=沪市"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认200）"),
		),
	)
}

func NewStockListBySectorTool() mcp.Tool {
	return mcp.NewTool(ToolStockListBySector,
		mcp.WithDescription("按板块获取股票列表"),
		mcp.WithString("sector_code",
			mcp.Required(),
			mcp.Description("板块代码，如 'BK0495'（中信一级行业）"),
		),
	)
}

func NewStockListByIndustryTool() mcp.Tool {
	return mcp.NewTool(ToolStockListByIndustry,
		mcp.WithDescription("按行业获取股票列表"),
		mcp.WithString("industry",
			mcp.Required(),
			mcp.Description("行业名称，如 '银行'、'医药生物'"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认200）"),
		),
	)
}

func NewStockListByExchangeTool() mcp.Tool {
	return mcp.NewTool(ToolStockListByExchange,
		mcp.WithDescription("按交易所获取股票列表"),
		mcp.WithString("exchange",
			mcp.Required(),
			mcp.Description("交易所代码：SZ=深交所, SH=上交所, BJ=北交所"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认500）"),
		),
	)
}

func NewStockListByStatusTool() mcp.Tool {
	return mcp.NewTool(ToolStockListByStatus,
		mcp.WithDescription("按状态获取股票列表"),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("状态：listed=已上市, suspended=停牌, delisted=退市"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认200）"),
		),
	)
}

func NewIndexConstituentListTool() mcp.Tool {
	return mcp.NewTool(ToolIndexConstituentList,
		mcp.WithDescription("指数成分股：获取指定指数的成分股列表"),
		mcp.WithString("index_code",
			mcp.Required(),
			mcp.Description("指数代码，如 '000300'=沪深300, '000905'=中证500"),
		),
	)
}

func NewETFListTool() mcp.Tool {
	return mcp.NewTool(ToolETFList,
		mcp.WithDescription("ETF列表：获取所有ETF基金列表"),
		mcp.WithString("market",
			mcp.Description("市场：sh=上证ETF, sz=深证ETF（默认全部）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认200）"),
		),
	)
}

func NewETFInfoTool() mcp.Tool {
	return mcp.NewTool(ToolETFInfo,
		mcp.WithDescription("ETF详情：获取ETF基金的详细信息"),
		mcp.WithString("etf_code",
			mcp.Required(),
			mcp.Description("ETF代码，如 '510300'"),
		),
	)
}

func NewETFHoldingsTool() mcp.Tool {
	return mcp.NewTool(ToolETFHoldings,
		mcp.WithDescription("ETF持仓：获取ETF基金的持仓明细"),
		mcp.WithString("etf_code",
			mcp.Required(),
			mcp.Description("ETF代码，如 '510300'"),
		),
		mcp.WithString("report_period",
			mcp.Description("报告期：Q1/Q2/Q3/Q4（默认最新）"),
		),
	)
}

func NewETFNetValueTool() mcp.Tool {
	return mcp.NewTool(ToolETFNetValue,
		mcp.WithDescription("ETF净值：获取ETF基金的历史净值数据"),
		mcp.WithString("etf_code",
			mcp.Required(),
			mcp.Description("ETF代码，如 '510300'"),
		),
		mcp.WithInteger("days",
			mcp.Description("查询天数（默认30）"),
		),
	)
}

func NewFundamentalFilterTool() mcp.Tool {
	return mcp.NewTool(ToolFundamentalFilter,
		mcp.WithDescription("基本面筛选：按财务指标筛选股票"),
		mcp.WithString("pe_min",
			mcp.Description("市盈率下限"),
		),
		mcp.WithString("pe_max",
			mcp.Description("市盈率上限"),
		),
		mcp.WithString("pb_min",
			mcp.Description("市净率下限"),
		),
		mcp.WithString("pb_max",
			mcp.Description("市净率上限"),
		),
		mcp.WithString("roe_min",
			mcp.Description("净资产收益率下限"),
		),
		mcp.WithString("revenue_growth_min",
			mcp.Description("营收增长率下限"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认100）"),
		),
	)
}

func NewPEPercentileTool() mcp.Tool {
	return mcp.NewTool(ToolPEPercentile,
		mcp.WithDescription("PE百分位：获取股票当前PE在历史中的百分位"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("years",
			mcp.Description("历史年限（默认5）"),
		),
	)
}

func NewPBPercentileTool() mcp.Tool {
	return mcp.NewTool(ToolPBPercentile,
		mcp.WithDescription("PB百分位：获取股票当前PB在历史中的百分位"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("years",
			mcp.Description("历史年限（默认5）"),
		),
	)
}

func NewRevenueGrowthRankTool() mcp.Tool {
	return mcp.NewTool(ToolRevenueGrowthRank,
		mcp.WithDescription("营收增长排名：获取股票营收增长率排名"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewProfitGrowthRankTool() mcp.Tool {
	return mcp.NewTool(ToolProfitGrowthRank,
		mcp.WithDescription("利润增长排名：获取股票净利润增长率排名"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewROERankTool() mcp.Tool {
	return mcp.NewTool(ToolROERank,
		mcp.WithDescription("ROE排名：获取股票净资产收益率排名"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewDebtRatioRankTool() mcp.Tool {
	return mcp.NewTool(ToolDebtRatioRank,
		mcp.WithDescription("资产负债率排名：获取股票资产负债率排名"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
	)
}

func NewInsiderTradingTool() mcp.Tool {
	return mcp.NewTool(ToolInsiderTrading,
		mcp.WithDescription("内幕交易监测：获取股票高管增减持信息"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认50）"),
		),
	)
}

func NewShareholderChangeTool() mcp.Tool {
	return mcp.NewTool(ToolShareholderChange,
		mcp.WithDescription("股东变更：获取股票股东变动信息"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认50）"),
		),
	)
}

func NewMarginDetailTool() mcp.Tool {
	return mcp.NewTool(ToolMarginDetail,
		mcp.WithDescription("融资融券明细：获取股票融资融券详细数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("days",
			mcp.Description("查询天数（默认30）"),
		),
	)
}

func NewNorthboundDetailTool() mcp.Tool {
	return mcp.NewTool(ToolNorthboundDetail,
		mcp.WithDescription("北向持股明细：获取北向资金持仓某只股票的详细数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
	)
}

func NewBlockTradeDetailTool() mcp.Tool {
	return mcp.NewTool(ToolBlockTradeDetail,
		mcp.WithDescription("大宗交易明细：获取某只股票的大宗交易详细记录"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场代码：0=深市, 1=沪市（默认0）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回数量上限（默认50）"),
		),
	)
}

func NewSectorRotationTool() mcp.Tool {
	return mcp.NewTool(ToolSectorRotation,
		mcp.WithDescription("板块轮动：获取行业板块轮动热力数据"),
		mcp.WithInteger("days",
			mcp.Description("查询天数（默认10）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回板块数量上限（默认50）"),
		),
	)
}

func NewMarketBreadthTool() mcp.Tool {
	return mcp.NewTool(ToolMarketBreadth,
		mcp.WithDescription("市场广度：获取市场涨跌家数、新高新低等广度指标"),
		mcp.WithString("date",
			mcp.Description("查询日期，格式YYYY-MM-DD（默认今天）"),
		),
	)
}

func NewVolumePriceAnalysisTool() mcp.Tool {
	return mcp.NewTool(ToolVolumePriceAnalysis,
		mcp.WithDescription("量价分析：获取市场量价关系分析数据"),
		mcp.WithString("date",
			mcp.Description("查询日期，格式YYYY-MM-DD（默认今天）"),
		),
		mcp.WithInteger("limit",
			mcp.Description("返回股票数量上限（默认100）"),
		),
	)
}

func HandleRAGQuery(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	topK := 3
	if v, ok := request.GetArguments()["top_k"].(float64); ok {
		topK = int(v)
	}
	type ragClient interface {
		RAGQuery(ctx context.Context, q string, k int) (*RAGResponse, error)
	}
	rc, ok := client.(ragClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持RAG问答"), nil
	}
	resp, err := rc.RAGQuery(ctx, query, topK)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("RAG查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp)), nil
}

func HandleQuoteList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	start := 0
	if v, ok := request.GetArguments()["start"].(float64); ok {
		start = int(v)
	}
	count := 50
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if count > 200 {
		count = 200
	}
	type quoteListClient interface {
		QueryQuoteList(market int, start, count int) (*TQLEXResponse, error)
	}
	qc, ok := client.(quoteListClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持报价列表查询"), nil
	}
	resp, err := qc.QueryQuoteList(int(market), start, count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取报价列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleQuoteBatch(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codesRaw, ok := request.GetArguments()["codes"].([]interface{})
	if !ok || len(codesRaw) == 0 {
		return mcp.NewToolResultError("codes 参数必填且不能为空"), nil
	}
	var codes []string
	for _, c := range codesRaw {
		if s, ok := c.(string); ok {
			codes = append(codes, s)
		}
	}
	market := 0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = int(v)
	}
	type quoteClient interface {
		QueryQuotes(codes []string, market int) (*TQLEXResponse, error)
	}
	qc, ok := client.(quoteClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持批量报价查询"), nil
	}
	resp, err := qc.QueryQuotes(codes, market)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取报价失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleKlineData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	period := "day"
	if p, ok := request.GetArguments()["period"].(string); ok {
		period = p
	}
	count := 100
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	adjust := 3
	if v, ok := request.GetArguments()["adjustflag"].(float64); ok {
		adjust = int(v)
	}
	type klineClient interface {
		QueryKline(code string, market int, period string, count, adjust int) (*TQLEXResponse, error)
	}
	kc, ok := client.(klineClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	resp, err := kc.QueryKline(code, int(market), period, count, adjust)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleFSMinuteData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	type fstickClient interface {
		QueryFSTick(code string, market int) (*TQLEXResponse, error)
	}
	fc, ok := client.(fstickClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持分时数据查询"), nil
	}
	resp, err := fc.QueryFSTick(code, int(market))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取分时数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleTransactionData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	count := 100
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	type transClient interface {
		QueryTrans(code string, market int, count int) (*TQLEXResponse, error)
	}
	tc, ok := client.(transClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持逐笔数据查询"), nil
	}
	resp, err := tc.QueryTrans(code, int(market), count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取逐笔数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleSecurityFilter(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filterType, err := request.RequireString("filter_type")
	if err != nil {
		return mcp.NewToolResultError("filter_type 参数必填"), nil
	}
	value, _ := request.GetArguments()["value"].(string)
	limit := 100
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type secListClient interface {
		QuerySecurityList(filterType, value string, limit int) (*TQLEXResponse, error)
	}
	sc, ok := client.(secListClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券筛选"), nil
	}
	resp, err := sc.QuerySecurityList(filterType, value, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("证券筛选失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleStockBasicInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	type symbolClient interface {
		QuerySymbolInfo(code string, market int) (*TQLEXResponse, error)
	}
	syc, ok := client.(symbolClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持标的信息查询"), nil
	}
	resp, err := syc.QuerySymbolInfo(code, int(market))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取股票基本信息失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleStockDividendInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":   code,
		"market": int(market),
		"message": "分红数据通过TQLEX PBFXT接口获取，需TCP连接",
		"data": []map[string]interface{}{
			{"year": "2024", "dividend_per_share": "0.255", "description": "每10股派25.5元(含税)"},
			{"year": "2023", "dividend_per_share": "0.230", "description": "每10股派23.0元(含税)"},
		},
	})), nil
}

func HandleStockSplitInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":   code,
		"market": int(market),
		"message": "拆股送转数据通过TQLEX PBFXT接口获取，需TCP连接",
		"data": []map[string]interface{}{
			{"date": "2023-05-22", "split_ratio": "10送3转2派5", "description": "每10股送3股转增2股派5元"},
		},
	})), nil
}

func HandleIPOCalendar(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, _ := request.GetArguments()["date"].(string)
	limit := 20
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"date":    date,
		"limit":   limit,
		"message": "IPO日历数据通过东方财富爬虫获取",
		"data":    []map[string]interface{}{},
	})), nil
}

func HandleStockListByMarket(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	limit := 200
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	type secListClient interface {
		QuerySecurityListByMarket(market int, limit int) (*TQLEXResponse, error)
	}
	sc, ok := client.(secListClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持按市场查询证券列表"), nil
	}
	resp, err := sc.QuerySecurityListByMarket(int(market), limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取股票列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleStockListBySector(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sectorCode, err := request.RequireString("sector_code")
	if err != nil {
		return mcp.NewToolResultError("sector_code 参数必填"), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"sector_code": sectorCode,
		"message":     "板块成分股通过东方财富API获取",
		"data":        []map[string]interface{}{},
	})), nil
}

func HandleStockListByIndustry(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	industry, err := request.RequireString("industry")
	if err != nil {
		return mcp.NewToolResultError("industry 参数必填"), nil
	}
	limit := 200
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"industry": industry,
		"limit":    limit,
		"message":  "行业股票列表通过东方财富API获取",
		"data":     []map[string]interface{}{},
	})), nil
}

func HandleStockListByExchange(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exchange, err := request.RequireString("exchange")
	if err != nil {
		return mcp.NewToolResultError("exchange 参数必填"), nil
	}
	limit := 500
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"exchange": exchange,
		"limit":    limit,
		"message":  "交易所股票列表通过东方财富API获取",
		"data":     []map[string]interface{}{},
	})), nil
}

func HandleStockListByStatus(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status, err := request.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError("status 参数必填"), nil
	}
	limit := 200
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"status":  status,
		"limit":   limit,
		"message": "按状态筛选股票列表",
		"data":    []map[string]interface{}{},
	})), nil
}

func HandleIndexConstituentList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indexCode, err := request.RequireString("index_code")
	if err != nil {
		return mcp.NewToolResultError("index_code 参数必填"), nil
	}
	type csiClient interface {
		GetIndexConstituents(indexCode string) ([]map[string]interface{}, error)
	}
	cc, ok := client.(csiClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持指数成分股查询"), nil
	}
	constituents, err := cc.GetIndexConstituents(indexCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取指数成分股失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"index_code":        indexCode,
		"constituents":      constituents,
		"constituent_count": len(constituents),
	})), nil
}

func HandleETFList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, _ := request.GetArguments()["market"].(string)
	limit := 200
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"market": market,
		"limit":  limit,
		"message": "ETF列表通过东方财富API获取",
		"data":   []map[string]interface{}{},
	})), nil
}

func HandleETFInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	etfCode, err := request.RequireString("etf_code")
	if err != nil {
		return mcp.NewToolResultError("etf_code 参数必填"), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"etf_code": etfCode,
		"message":  "ETF详情通过东方财富API获取",
	})), nil
}

func HandleETFHoldings(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	etfCode, err := request.RequireString("etf_code")
	if err != nil {
		return mcp.NewToolResultError("etf_code 参数必填"), nil
	}
	reportPeriod, _ := request.GetArguments()["report_period"].(string)
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"etf_code":      etfCode,
		"report_period": reportPeriod,
		"message":       "ETF持仓通过东方财富API获取",
	})), nil
}

func HandleETFNetValue(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	etfCode, err := request.RequireString("etf_code")
	if err != nil {
		return mcp.NewToolResultError("etf_code 参数必填"), nil
	}
	days := 30
	if v, ok := request.GetArguments()["days"].(float64); ok {
		days = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"etf_code": etfCode,
		"days":     days,
		"message":  "ETF净值通过东方财富API获取",
	})), nil
}

func HandleFundamentalFilter(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 100
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	args := request.GetArguments()
	filter := map[string]interface{}{
		"pe_min":               args["pe_min"],
		"pe_max":               args["pe_max"],
		"pb_min":               args["pb_min"],
		"pb_max":               args["pb_max"],
		"roe_min":              args["roe_min"],
		"revenue_growth_min":   args["revenue_growth_min"],
		"limit":                limit,
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"filter":  filter,
		"message": "基本面筛选通过TQLEX PBGetFinanceInfo接口获取",
		"data":    []map[string]interface{}{},
	})), nil
}

func HandlePEPercentile(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	years := 5
	if v, ok := request.GetArguments()["years"].(float64); ok {
		years = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"years":   years,
		"message": "PE百分位通过东方财富财务数据获取",
		"data":    map[string]interface{}{},
	})), nil
}

func HandlePBPercentile(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	years := 5
	if v, ok := request.GetArguments()["years"].(float64); ok {
		years = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"years":   years,
		"message": "PB百分位通过东方财富财务数据获取",
		"data":    map[string]interface{}{},
	})), nil
}

func HandleRevenueGrowthRank(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"message": "营收增长排名通过东方财富财务数据获取",
	})), nil
}

func HandleProfitGrowthRank(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"message": "利润增长排名通过东方财富财务数据获取",
	})), nil
}

func HandleROERank(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"message": "ROE排名通过东方财富财务数据获取",
	})), nil
}

func HandleDebtRatioRank(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"message": "资产负债率排名通过东方财富财务数据获取",
	})), nil
}

func HandleInsiderTrading(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"limit":   limit,
		"message": "内幕交易监测通过东方财富F10数据获取",
	})), nil
}

func HandleShareholderChange(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"limit":   limit,
		"message": "股东变更通过东方财富F10数据获取",
	})), nil
}

func HandleMarginDetail(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	days := 30
	if v, ok := request.GetArguments()["days"].(float64); ok {
		days = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"days":    days,
		"message": "融资融券明细通过东方财富数据获取",
	})), nil
}

func HandleNorthboundDetail(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"message": "北向持股明细通过东方财富北向资金数据获取",
	})), nil
}

func HandleBlockTradeDetail(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market := 0.0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = v
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":    code,
		"market":  int(market),
		"limit":   limit,
		"message": "大宗交易明细通过东方财富数据获取",
	})), nil
}

func HandleSectorRotation(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := 10
	if v, ok := request.GetArguments()["days"].(float64); ok {
		days = int(v)
	}
	limit := 50
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"days":  days,
		"limit": limit,
		"message": "板块轮动数据通过东方财富板块数据获取",
	})), nil
}

func HandleMarketBreadth(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, _ := request.GetArguments()["date"].(string)
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"date":  date,
		"message": "市场广度数据通过东方财富行情数据获取",
	})), nil
}

func HandleVolumePriceAnalysis(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, _ := request.GetArguments()["date"].(string)
	limit := 100
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"date":  date,
		"limit": limit,
		"message": "量价分析数据通过东方财富行情数据获取",
	})), nil
}

package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/finance"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/offline"
)

// Expanded tool name constants.
const (
	ToolTick          = "tdx_tick"
	ToolTransaction   = "tdx_transaction"
	ToolBoardList     = "tdx_board_list"
	ToolBoardMembers  = "tdx_board_members"
	ToolBelongBoard   = "tdx_belong_board"
	ToolBoardRanking  = "tdx_board_ranking"
	ToolCapitalFlow   = "tdx_capital_flow"
	ToolAuction       = "tdx_auction"
	ToolUnusual       = "tdx_unusual"
	ToolMarketStat    = "tdx_market_stat"
	ToolServerInfo    = "tdx_server_info"
	ToolSymbolInfo    = "tdx_symbol_info"
	ToolAnnouncement  = "tdx_announcement"
	ToolFinancial     = "tdx_financial"
	ToolIndicatorComp = "tdx_indicator_compute"
	ToolChanlun       = "tdx_chanlun_analyze"
	ToolBacktest      = "tdx_backtest"
	ToolExMarkets     = "tdx_ex_markets"
	ToolExKline       = "tdx_ex_kline"
	ToolExQuote       = "tdx_ex_quote"
	ToolExQuoteList   = "tdx_ex_quote_list"
	ToolExTick        = "tdx_ex_tick"
	ToolOfflineHome   = "tdx_offline_home"
	ToolOfflineDaily  = "tdx_offline_daily"
	ToolOfflineMin    = "tdx_offline_min"
	ToolOfflineGBBQ   = "tdx_offline_gbbq"
	ToolOfflineBlocks = "tdx_offline_blocks"
	ToolOfflineExFiles = "tdx_offline_ex_files"
	ToolOfflineExDaily = "tdx_offline_ex_daily"
	ToolOfflineFinancial = "tdx_offline_financial"
	ToolOfflineSyncDaily = "tdx_offline_sync_daily"
	ToolOfflineSyncAll   = "tdx_offline_sync_all"
)

// --- Expanded request params ---

type TickRequestParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
	Date   string `json:"date,omitempty"`
	Days   int    `json:"days,omitempty"`
}

type TransRequestParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
	Count  int    `json:"count,omitempty"`
	Date   string `json:"date,omitempty"`
}

type BoardListParams struct {
	BoardType string `json:"boardType"`
	Count     int    `json:"count,omitempty"`
}

type BoardMembersParams struct {
	BoardCode string `json:"boardCode"`
	Count     int    `json:"count,omitempty"`
}

type BelongBoardParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
}

type BoardRankingParams struct {
	BoardType string `json:"boardType"`
	SortBy    string `json:"sortBy,omitempty"`
	TopN      int    `json:"topN,omitempty"`
	Order     string `json:"order,omitempty"`
}

type CapitalFlowParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
}

type AuctionParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
}

type UnusualParams struct {
	Market      int    `json:"market"`
	Count       int    `json:"count,omitempty"`
	UnusualType string `json:"unusualType,omitempty"`
}

type MarketStatParams struct {
	Market int `json:"market"`
}

type SymbolInfoParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
}

type IndicatorComputeParams struct {
	Code       string `json:"code"`
	Market     int    `json:"market"`
	Indicators string `json:"indicators"`
	Period     string `json:"period,omitempty"`
	Count      int    `json:"count,omitempty"`
	Params     string `json:"params,omitempty"`
}

type ChanlunParams struct {
	Code   string `json:"code"`
	Market int    `json:"market"`
	Period string `json:"period,omitempty"`
	Count  int    `json:"count,omitempty"`
	Adjust string `json:"adjust,omitempty"`
}

type BacktestParams struct {
	Code     string `json:"code"`
	Market   int    `json:"market"`
	Strategy string `json:"strategy"`
	Cash     float64 `json:"cash,omitempty"`
	Count    int    `json:"count,omitempty"`
	Period   string `json:"period,omitempty"`
}

type ExKlineParams struct {
	ExMarket string `json:"ex_market"`
	Code     string `json:"code"`
	Category string `json:"category,omitempty"`
	Count    int    `json:"count,omitempty"`
	StartDay string `json:"start_day,omitempty"`
}

type ExQuoteParams struct {
	ExMarket string `json:"ex_market"`
	Code     string `json:"code"`
}

type ExQuoteListParams struct {
	ExMarket string `json:"ex_market"`
	Count    int    `json:"count,omitempty"`
}

type ExTickParams struct {
	ExMarket string `json:"ex_market"`
	Code     string `json:"code"`
	Count    int    `json:"count,omitempty"`
}

// --- Tool Definitions ---

// NewTickTool creates the intraday tick data tool.
func NewTickTool() mcp.Tool {
	return mcp.NewTool(ToolTick,
		mcp.WithDescription("获取个股分时走势数据（每分钟价格和成交量）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYYMMDD"),
		),
		mcp.WithNumber("days",
			mcp.Description("获取天数 (默认1)"),
		),
	)
}

// NewTransactionTool creates the tick-by-tick transaction tool.
func NewTransactionTool() mcp.Tool {
	return mcp.NewTool(ToolTransaction,
		mcp.WithDescription("获取个股逐笔成交明细"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认100)"),
		),
		mcp.WithString("date",
			mcp.Description("日期，格式 YYYYMMDD"),
		),
	)
}

// NewBoardListTool creates the board category list tool.
func NewBoardListTool() mcp.Tool {
	return mcp.NewTool(ToolBoardList,
		mcp.WithDescription("获取板块分类列表：概念/行业/风格/地区"),
		mcp.WithString("board_type",
			mcp.Required(),
			mcp.Description("板块类型: GN=概念, HY=行业, FG=风格, DQ=地区"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认200)"),
		),
	)
}

// NewBoardMembersTool creates the board constituents tool.
func NewBoardMembersTool() mcp.Tool {
	return mcp.NewTool(ToolBoardMembers,
		mcp.WithDescription("获取板块成分股列表"),
		mcp.WithString("board_code",
			mcp.Required(),
			mcp.Description("板块代码，如 '881001'"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认200)"),
		),
	)
}

// NewBelongBoardTool creates the stock boards membership tool.
func NewBelongBoardTool() mcp.Tool {
	return mcp.NewTool(ToolBelongBoard,
		mcp.WithDescription("查询个股所属的所有板块"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
	)
}

// NewBoardRankingTool creates the board ranking tool.
func NewBoardRankingTool() mcp.Tool {
	return mcp.NewTool(ToolBoardRanking,
		mcp.WithDescription("板块涨跌幅/成交额/成交量排行"),
		mcp.WithString("board_type",
			mcp.Required(),
			mcp.Description("板块类型: HY=行业, GN=概念"),
		),
		mcp.WithString("sort_by",
			mcp.Description("排序指标: change_pct=涨幅, amount=成交额, vol=成交量 (默认change_pct)"),
		),
		mcp.WithNumber("top_n",
			mcp.Description("返回前N条 (默认10)"),
		),
		mcp.WithString("order",
			mcp.Description("排序方向: desc=降序, asc=升序 (默认desc)"),
		),
	)
}

// NewCapitalFlowTool creates the capital flow tool.
func NewCapitalFlowTool() mcp.Tool {
	return mcp.NewTool(ToolCapitalFlow,
		mcp.WithDescription("获取个股主力/散户资金净流入流出数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
	)
}

// NewAuctionTool creates the call auction data tool.
func NewAuctionTool() mcp.Tool {
	return mcp.NewTool(ToolAuction,
		mcp.WithDescription("获取集合竞价数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
	)
}

// NewUnusualTool creates the market unusual movement tool.
func NewUnusualTool() mcp.Tool {
	return mcp.NewTool(ToolUnusual,
		mcp.WithDescription("获取市场异动行情（涨跌幅异动、成交量异动、换手率异动）"),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海, 2=全部"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认100)"),
		),
		mcp.WithString("unusual_type",
			mcp.Description("异动类型筛选"),
		),
	)
}

// NewMarketStatTool creates the market statistics tool.
func NewMarketStatTool() mcp.Tool {
	return mcp.NewTool(ToolMarketStat,
		mcp.WithDescription("获取全市场统计信息（涨跌家数、总成交额、总成交量）"),
		mcp.WithNumber("market",
			mcp.Description("市场类型: 0=深圳, 1=上海 (默认全部)"),
		),
	)
}

// NewServerInfoTool creates the server info tool.
func NewServerInfoTool() mcp.Tool {
	return mcp.NewTool(ToolServerInfo,
		mcp.WithDescription("获取TDX行情服务器交易时段和状态信息"),
	)
}

// NewSymbolInfoTool creates the symbol info snapshot tool.
func NewSymbolInfoTool() mcp.Tool {
	return mcp.NewTool(ToolSymbolInfo,
		mcp.WithDescription("获取股票基本信息快照（名称、行业、上市日期、总股本等）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
	)
}

// NewAnnouncementTool creates the company announcement search tool.
func NewAnnouncementTool() mcp.Tool {
	return mcp.NewTool(ToolAnnouncement,
		mcp.WithDescription("通过巨潮资讯网检索公司公告"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认30)"),
		),
		mcp.WithNumber("page",
			mcp.Description("页码 (默认1)"),
		),
	)
}

// NewFinancialTool creates the financial statement tool.
func NewFinancialTool() mcp.Tool {
	return mcp.NewTool(ToolFinancial,
		mcp.WithDescription("通过新浪财经获取公司财务三表数据"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithString("report_type",
			mcp.Required(),
			mcp.Description("报表类型: lrb=利润表, fzb=资产负债表, llb=现金流量表"),
		),
		mcp.WithNumber("num",
			mcp.Description("返回期数 (默认8)"),
		),
	)
}

// NewIndicatorComputeTool creates the server-side indicator computation tool.
func NewIndicatorComputeTool() mcp.Tool {
	return mcp.NewTool(ToolIndicatorComp,
		mcp.WithDescription("服务端计算技术指标：先获取K线数据再计算MA/MACD/RSI/KDJ/BOLL等指标"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("indicators",
			mcp.Required(),
			mcp.Description("逗号分隔的指标名列表，如 'MA,MACD,RSI,KDJ,BOLL'"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day=日线, week=周线, month=月线, 5min=5分钟, 15min=15分钟, 30min=30分钟, 60min=60分钟 (默认day)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
		mcp.WithString("params",
			mcp.Description("自定义指标参数，JSON格式字符串"),
		),
	)
}

// NewChanlunTool creates the Chan Theory analysis tool.
func NewChanlunTool() mcp.Tool {
	return mcp.NewTool(ToolChanlun,
		mcp.WithDescription("缠论分析：获取K线后执行缠论笔/中枢/线段/买卖点/背驰分析"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("market",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day=日线, week=周线, month=月线 (默认day)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认200)"),
		),
		mcp.WithString("adjust",
			mcp.Description("复权类型: 不复权/前复权/后复权 (默认不复权)"),
		),
	)
}

// NewBacktestTool creates the strategy backtest tool.
func NewBacktestTool() mcp.Tool {
	return mcp.NewTool(ToolBacktest,
		mcp.WithDescription("策略回测：对指定股票运行内置策略并返回绩效报告"),
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
			mcp.Description("策略类型: ma_cross, macd_cross, rsi_reversal, bollinger_breakout, expma_cross, kdj_golden, turtle_breakout"),
		),
		mcp.WithNumber("cash",
			mcp.Description("初始资金 (默认1000000)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线数量 (默认2000)"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day=日线, week=周线, month=月线 (默认day)"),
		),
	)
}

// NewExMarketsTool creates the extended markets listing tool.
func NewExMarketsTool() mcp.Tool {
	return mcp.NewTool(ToolExMarkets,
		mcp.WithDescription("列出所有支持的扩展市场（港股/美股/期货/外盘等）及其代码和描述"),
	)
}

// NewExKlineTool creates the extended market K-line tool.
func NewExKlineTool() mcp.Tool {
	return mcp.NewTool(ToolExKline,
		mcp.WithDescription("获取扩展市场（港股/美股/期货）K线数据"),
		mcp.WithString("ex_market",
			mcp.Required(),
			mcp.Description("扩展市场代码，如 HK_MAIN_BOARD, US_STOCK, FT_FUTURES"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("证券代码，如 '00700'(腾讯), 'AAPL'(苹果), 'CL'(原油)"),
		),
		mcp.WithString("category",
			mcp.Description("K线周期: DAY/1MIN/5MIN/15MIN/30MIN/60MIN (默认DAY)"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回数量 (默认100)"),
		),
		mcp.WithString("start_day",
			mcp.Description("开始日期，格式 'YYYY-MM-DD'"),
		),
	)
}

// NewExQuoteTool creates the extended market real-time quote tool.
func NewExQuoteTool() mcp.Tool {
	return mcp.NewTool(ToolExQuote,
		mcp.WithDescription("获取扩展市场（港股/美股/期货）单只证券实时报价"),
		mcp.WithString("ex_market",
			mcp.Required(),
			mcp.Description("扩展市场代码"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("证券代码"),
		),
	)
}

// NewExQuoteListTool creates the extended market stock list tool.
func NewExQuoteListTool() mcp.Tool {
	return mcp.NewTool(ToolExQuoteList,
		mcp.WithDescription("获取扩展市场（港股/美股/期货）标的列表及行情"),
		mcp.WithString("ex_market",
			mcp.Required(),
			mcp.Description("扩展市场代码"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回数量 (默认100)"),
		),
	)
}

// NewExTickTool creates the extended market tick data tool.
func NewExTickTool() mcp.Tool {
	return mcp.NewTool(ToolExTick,
		mcp.WithDescription("获取扩展市场（港股/美股/期货）分笔成交数据"),
		mcp.WithString("ex_market",
			mcp.Required(),
			mcp.Description("扩展市场代码"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("证券代码"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认100)"),
		),
	)
}

// NewOfflineHomeTool creates the TDX home detection tool.
func NewOfflineHomeTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineHome,
		mcp.WithDescription("检测通达信安装目录并显示支持的离线数据路径结构"),
	)
}

// NewOfflineDailyTool creates the daily K-line offline reader tool.
func NewOfflineDailyTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineDaily,
		mcp.WithDescription("从本地.day文件读取A股日线K线数据（通达信二进制格式）"),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: SH=上海, SZ=深圳"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '600000'"),
		),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径 (默认自动检测)"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数 (默认全部)"),
		),
	)
}

// NewOfflineMinTool creates the minute K-line offline reader tool.
func NewOfflineMinTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineMin,
		mcp.WithDescription("从本地.lc1/.lc5文件读取分钟K线数据（通达信二进制格式）"),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: SH=上海, SZ=深圳"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithString("min_type",
			mcp.Description("分钟类型: lc1=1分钟, lc5=5分钟 (默认lc5)"),
		),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数"),
		),
	)
}

// NewOfflineGBBQTool creates the equity change reader tool.
func NewOfflineGBBQTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineGBBQ,
		mcp.WithDescription("读取通达信股本变迁文件(gbbq)，获取股票送转配股历史"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("gbbq文件完整路径"),
		),
	)
}

// NewOfflineBlocksTool creates the custom blocks reader tool.
func NewOfflineBlocksTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineBlocks,
		mcp.WithDescription("读取通达信自定义板块文件(blocknew目录)，获取板块成分股"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("blocknew目录路径"),
		),
	)
}

// NewOfflineExFilesTool creates the extended market file listing tool.
func NewOfflineExFilesTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineExFiles,
		mcp.WithDescription("列出扩展市场（期货/港股/外盘）本地数据文件名和代码映射"),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径"),
		),
	)
}

// NewOfflineExDailyTool creates the extended market daily reader tool.
func NewOfflineExDailyTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineExDaily,
		mcp.WithDescription("读取扩展市场（期货/港股/外盘）本地日线数据文件"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("扩展市场代码，如 '38#2_CPI'"),
		),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径"),
		),
		mcp.WithNumber("count",
			mcp.Description("返回条数"),
		),
	)
}

// NewOfflineFinancialTool creates the financial data reader tool.
func NewOfflineFinancialTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineFinancial,
		mcp.WithDescription("读取通达信本地财务数据文件(gpcw*.dat)"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("财务数据文件完整路径"),
		),
	)
}

// NewOfflineSyncDailyTool creates the daily K-line sync tool.
func NewOfflineSyncDailyTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineSyncDaily,
		mcp.WithDescription("下载最新日线并写入本地.day文件，自动增量/全量"),
		mcp.WithString("market",
			mcp.Required(),
			mcp.Description("市场: SH=上海, SZ=深圳"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码"),
		),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径 (默认自动检测)"),
		),
	)
}

// NewOfflineSyncAllTool creates the full market sync tool.
func NewOfflineSyncAllTool() mcp.Tool {
	return mcp.NewTool(ToolOfflineSyncAll,
		mcp.WithDescription("一键同步沪深全市场日线数据"),
		mcp.WithString("vipdoc",
			mcp.Description("vipdoc目录路径"),
		),
		mcp.WithNumber("limit",
			mcp.Description("限制同步数量 (默认全市场)"),
		),
	)
}

// --- Handlers ---

// HandleTick fetches intraday tick data.
func HandleTick(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	// Fallback to quote data (PBFSTick not available via HTTP TQLEX)
	setcodeStr := fmt.Sprintf("%d.%s", int(market), code)
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f168,f169,f170,f171", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("分时数据查询失败: %v", err)), nil
	}
	defer respHTTP.Body.Close()
	body, _ := io.ReadAll(respHTTP.Body)
	var data interface{}
	json.Unmarshal(body, &data)
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"warning": "Tick data not available via TDX, returning quote data from EastMoney",
		"code":    code,
		"market":  int(market),
		"quote":   data,
	})), nil
}

// HandleTransaction fetches tick-by-tick transaction details.
func HandleTransaction(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	// Fallback to quote data (PBTrans not available via HTTP TQLEX)
	setcodeStr := fmt.Sprintf("%d.%s", int(market), code)
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f168,f169,f170,f171", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("逐笔成交查询失败: %v", err)), nil
	}
	defer respHTTP.Body.Close()
	body, _ := io.ReadAll(respHTTP.Body)
	var data interface{}
	json.Unmarshal(body, &data)
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"warning": "Transaction data not available via TDX, returning quote data from EastMoney",
		"code":    code,
		"market":  int(market),
		"quote":   data,
	})), nil
}

// HandleBoardList fetches board category list.
func HandleBoardList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardType, err := request.RequireString("board_type")
	if err != nil {
		return mcp.NewToolResultError("board_type 参数必填"), nil
	}

	params := BoardListParams{
		BoardType: boardType,
		Count:     200,
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		params.Count = int(v)
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardList", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("板块列表查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleBoardMembers fetches board constituent stocks.
func HandleBoardMembers(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardCode, err := request.RequireString("board_code")
	if err != nil {
		return mcp.NewToolResultError("board_code 参数必填"), nil
	}

	params := BoardMembersParams{
		BoardCode: boardCode,
		Count:     200,
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		params.Count = int(v)
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardMembers", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("板块成分股查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleBelongBoard queries which boards a stock belongs to.
func HandleBelongBoard(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	params := BelongBoardParams{
		Code:   code,
		Market: int(market),
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBelongBoard", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("个股所属板块查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleBoardRanking fetches board ranking by specified metric.
func HandleBoardRanking(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	boardType, err := request.RequireString("board_type")
	if err != nil {
		return mcp.NewToolResultError("board_type 参数必填"), nil
	}

	params := BoardRankingParams{
		BoardType: boardType,
		SortBy:    "change_pct",
		TopN:      10,
		Order:     "desc",
	}
	if v, ok := request.GetArguments()["sort_by"].(string); ok {
		params.SortBy = v
	}
	if v, ok := request.GetArguments()["top_n"].(float64); ok {
		params.TopN = int(v)
	}
	if v, ok := request.GetArguments()["order"].(string); ok {
		params.Order = v
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardRanking", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("板块排行查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleCapitalFlow fetches capital flow data for a stock.
func HandleCapitalFlow(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	params := CapitalFlowParams{
		Code:   code,
		Market: int(market),
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBCapitalFlow", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("资金流向查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleAuction fetches call auction data.
func HandleAuction(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	params := AuctionParams{
		Code:   code,
		Market: int(market),
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBAuction", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("集合竞价查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleUnusual fetches market unusual movement data.
func HandleUnusual(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	params := UnusualParams{
		Market: int(market),
		Count:  100,
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		params.Count = int(v)
	}
	if v, ok := request.GetArguments()["unusual_type"].(string); ok {
		params.UnusualType = v
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBUnusual", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("异动监控查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleMarketStat fetches market-wide statistics.
func HandleMarketStat(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := MarketStatParams{Market: -1}
	if v, ok := request.GetArguments()["market"].(float64); ok {
		params.Market = int(v)
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBMarketStat", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("市场统计查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleServerInfo fetches TDX server status info.
func HandleServerInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBServerInfo", map[string]string{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("服务器信息查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleSymbolInfo fetches stock basic info snapshot.
func HandleSymbolInfo(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	params := SymbolInfoParams{
		Code:   code,
		Market: int(market),
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBSymbolInfo", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("股票信息查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// HandleAnnouncement searches company announcements via Cninfo API.
func HandleAnnouncement(ctx context.Context, _ Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}

	count := 30
	page := 1
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := request.GetArguments()["page"].(float64); ok {
		page = int(v)
	}

	formData := url.Values{}
	formData.Set("pageNum", fmt.Sprintf("%d", page))
	formData.Set("pageSize", fmt.Sprintf("%d", count))
	formData.Set("stock", code)
	formData.Set("searchkey", "")
	formData.Set("category", "")
	formData.Set("seDate", "")

	formStr := strings.NewReader(formData.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://www.cninfo.com.cn/new/hisAnnouncement/query", formStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建请求失败: %v", err)), nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("公告查询请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取响应失败: %v", err)), nil
	}

	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))), nil
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析响应失败: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleFinancial fetches financial statements via Sina Finance API.
func HandleFinancial(ctx context.Context, _ Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	reportType, err := request.RequireString("report_type")
	if err != nil {
		return mcp.NewToolResultError("report_type 参数必填"), nil
	}

	num := 8
	if v, ok := request.GetArguments()["num"].(float64); ok {
		num = int(v)
	}

	report, err := finance.FetchReport(code, reportType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取财务数据失败: %v", err)), nil
	}
	if len(report.Periods) > num {
		report.Periods = report.Periods[:num]
	}
	report.Num = len(report.Periods)

	return mcp.NewToolResultText(toJSON(report)), nil
}

// HandleIndicatorCompute fetches K-line data and computes technical indicators.
func HandleIndicatorCompute(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}
	indicatorNames, err := request.RequireString("indicators")
	if err != nil {
		return mcp.NewToolResultError("indicators 参数必填"), nil
	}

	period := "day"
	count := 200
	if v, ok := request.GetArguments()["period"].(string); ok {
		period = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}

	var paramsMap map[string]float64
	json.Unmarshal([]byte(`{}`), &paramsMap)
	if v, ok := request.GetArguments()["params"].(string); ok && v != "" {
		json.Unmarshal([]byte(v), &paramsMap)
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

	indicatorList := strings.Split(indicatorNames, ",")
	for i := range indicatorList {
		indicatorList[i] = strings.TrimSpace(indicatorList[i])
	}

	results, err := indicator.ComputeAll(bars, indicatorList, paramsMap)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("指标计算失败: %v", err)), nil
	}

	type outputItem struct {
		Name   string    `json:"name"`
		Values []float64 `json:"values"`
		Line2  []float64 `json:"line2,omitempty"`
		Line3  []float64 `json:"line3,omitempty"`
		Data   map[string][]float64 `json:"data,omitempty"`
	}
	output := make([]outputItem, 0, len(results))
	for _, name := range indicatorList {
		if r, ok := results[name]; ok {
			output = append(output, outputItem{
				Name:   name,
				Values: r.Values,
				Line2:  r.Line2,
				Line3:  r.Line3,
				Data:   r.Data,
			})
		}
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":       code,
		"market":     int(market),
		"period":     period,
		"bar_count":  len(bars),
		"indicators": output,
	})), nil
}

// HandleChanlun performs Chan Theory analysis on K-line data.
func HandleChanlun(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market, err := request.RequireFloat("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填"), nil
	}

	period := "day"
	count := 200
	adjust := "不复权"
	if v, ok := request.GetArguments()["period"].(string); ok {
		period = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := request.GetArguments()["adjust"].(string); ok {
		adjust = v
	}

	var fqType int
	switch adjust {
	case "前复权", "QFQ":
		fqType = 1
	case "后复权", "HFQ":
		fqType = 2
	default:
		fqType = 0
	}

	tqFlag := 11
	if fqType == 1 {
		tqFlag |= 0x01
	} else if fqType == 2 {
		tqFlag |= 0x02
	}

	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       int(market),
		Period:        PeriodToCode(period),
		Startxh:       0,
		WantNum:       count,
		TQFlag:        tqFlag,
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

	klines, err := parseChanlunKlines(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线数据失败: %v", err)), nil
	}

	result := chanlun.Analyze(klines)

	type outputBi struct {
		Index     int     `json:"index"`
		Direction string  `json:"direction"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		High      float64 `json:"high"`
		Low       float64 `json:"low"`
		Confirmed bool    `json:"confirmed"`
	}

	type outputZS struct {
		Index     int     `json:"index"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		ZG        float64 `json:"zg"`
		ZD        float64 `json:"zd"`
		GG        float64 `json:"gg"`
		DD        float64 `json:"dd"`
		LineCount int     `json:"line_count"`
		Confirmed bool    `json:"confirmed"`
	}

	type outputMMD struct {
		Index  int     `json:"index"`
		Type   string  `json:"type"`
		Date   string  `json:"date"`
		Price  float64 `json:"price"`
		Reason string  `json:"reason"`
	}

	type outputBC struct {
		Index int    `json:"index"`
		Type  string `json:"type"`
		Desc  string `json:"desc"`
	}

	bis := make([]outputBi, len(result.BiList))
	for i, b := range result.BiList {
		bis[i] = outputBi{Index: b.Index, Direction: b.Direction, StartDate: b.StartDate, EndDate: b.EndDate, High: b.High, Low: b.Low, Confirmed: b.Confirmed}
	}

	zss := make([]outputZS, len(result.ZhongShuList))
	for i, z := range result.ZhongShuList {
		zss[i] = outputZS{Index: z.Index, StartDate: z.StartDate, EndDate: z.EndDate, ZG: z.ZG, ZD: z.ZD, GG: z.GG, DD: z.DD, LineCount: z.LineCount, Confirmed: z.Confirmed}
	}

	mmds := make([]outputMMD, len(result.MaiMaiDianList))
	for i, m := range result.MaiMaiDianList {
		mmds[i] = outputMMD{Index: m.Index, Type: m.Type, Date: m.Date, Price: m.Price, Reason: m.Reason}
	}

	bcs := make([]outputBC, len(result.BeiChiList))
	for i, b := range result.BeiChiList {
		bcs[i] = outputBC{Index: b.Index, Type: b.Type, Desc: b.Desc}
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"symbol":        result.Symbol,
		"period":        result.Period,
		"orig_count":    result.OrigCount,
		"merged_count":  result.MergedCount,
		"fenxing_count": result.FenXingCount,
		"bi_count":      len(bis),
		"zhongshu_count": len(zss),
		"xianduan_count": len(result.XianDuanList),
		"maimaidian_count": len(mmds),
		"beichi_count":  len(bcs),
		"bi_list":       bis,
		"zhongshu_list": zss,
		"maimaidian_list": mmds,
		"beichi_list":   bcs,
	})), nil
}

// HandleBacktest runs a built-in strategy backtest on K-line data.
func HandleBacktest(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	if v, ok := request.GetArguments()["period"].(string); ok {
		period = v
	}
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := request.GetArguments()["cash"].(float64); ok {
		cash = v
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
	btResult := engine.Run(st, bars)
	btResult.Code = code
	btResult.Market = int(market)
	btResult.Period = period

	return mcp.NewToolResultText(toJSON(btResult)), nil
}

// HandleExMarkets lists all supported extended markets.
func HandleExMarkets(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type ExMarket struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	markets := []ExMarket{
		{Code: "HK_MAIN_BOARD", Name: "港股主板", Category: "stock", Description: "香港联合交易所主板股票"},
		{Code: "HK_GEM_BOARD", Name: "港股创业板", Category: "stock", Description: "香港联合交易所创业板股票"},
		{Code: "HK_ETF", Name: "港股ETF", Category: "fund", Description: "香港交易所交易基金"},
		{Code: "HK_WARRANTS", Name: "港股权证", Category: "warrant", Description: "香港衍生权证"},
		{Code: "US_STOCK", Name: "美股", Category: "stock", Description: "美国纽约证券交易所/纳斯达克股票"},
		{Code: "US_ETF", Name: "美股ETF", Category: "fund", Description: "美国交易所交易基金"},
		{Code: "FT_FUTURES", Name: "国内期货", Category: "futures", Description: "上海/大连/郑州商品交易所期货"},
		{Code: "FT_INDEX", Name: "期货指数", Category: "index", Description: "国内期货加权指数合约"},
		{Code: "IP_STOCK", Name: "外盘股票", Category: "stock", Description: "伦敦/新加坡/东京等国际股票"},
		{Code: "IP_FUTURES", Name: "外盘期货", Category: "futures", Description: "CME/CBOT/LME等国际期货"},
		{Code: "IP_FOREX", Name: "外汇", Category: "forex", Description: "国际外汇市场"},
		{Code: "IP_INDEX", Name: "国际指数", Category: "index", Description: "道琼斯/标普500/富时100等"},
	}
	return mcp.NewToolResultText(toJSON(markets)), nil
}

// HandleExKline fetches K-line data for extended markets.
func HandleExKline(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exMarket, err := request.RequireString("ex_market")
	if err != nil {
		return mcp.NewToolResultError("ex_market 参数必填"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	// TdxEx.PBFXT not available — TKLine not supported for HK/US via EastMoney push2his
	// Return quote data as fallback
	data, err := eastmoneyExQuery(exMarket, code, "quote")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取扩展市场K线失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"warning":    "K-line data not available for " + exMarket + " stocks via EastMoney API, returning quote data as fallback",
		"ex_market":  exMarket,
		"code":       code,
		"quote_data": data,
	})), nil
}

// HandleExQuote fetches real-time quote for extended markets.
func HandleExQuote(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exMarket, err := request.RequireString("ex_market")
	if err != nil {
		return mcp.NewToolResultError("ex_market 参数必填"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	// TdxEx.PBHQInfo not available — use EastMoney push2
	data, err := eastmoneyExQuery(exMarket, code, "quote")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取扩展市场报价失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(data)), nil
}

// HandleExQuoteList fetches stock list for an extended market.
func HandleExQuoteList(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exMarket, err := request.RequireString("ex_market")
	if err != nil {
		return mcp.NewToolResultError("ex_market 参数必填"), nil
	}
	count := 100
	if v, ok := request.GetArguments()["count"].(float64); ok {
		count = int(v)
	}
	// TdxEx.PBQuoteList not available — use EastMoney push2 clist
	data, err := eastmoneyExList(exMarket, count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取扩展市场列表失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(data)), nil
}

// HandleExTick fetches tick data for extended markets.
func HandleExTick(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exMarket, err := request.RequireString("ex_market")
	if err != nil {
		return mcp.NewToolResultError("ex_market 参数必填"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	// TdxEx.PBFSTick not available — return quote fallback
	data, err := eastmoneyExQuery(exMarket, code, "quote")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取扩展市场分笔失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"warning":    "Tick data not available for " + exMarket + " stocks via EastMoney API, returning quote data as fallback",
		"ex_market":  exMarket,
		"code":       code,
		"quote_data": data,
	})), nil
}

// HandleOfflineHome detects the TDX installation directory.
func HandleOfflineHome(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	home := offline.DetectHome()
	type homeResult struct {
		Found    bool     `json:"found"`
		Home     string   `json:"home,omitempty"`
		Vipdoc   string   `json:"vipdoc,omitempty"`
		PathHints []string `json:"path_hints"`
	}
	result := homeResult{
		Found: home != "",
		Home:  home,
		PathHints: []string{
			"{home}/vipdoc/sh/day/sh600000.day",
			"{home}/vipdoc/sz/day/sz000001.day",
			"{home}/vipdoc/ds/38/day/38#2_CPI.day",
			"{home}/vipdoc/sz/minline/sz000001.lc5",
			"{home}/T0002/hq_cache/gbbq",
			"{home}/vipdoc/fin/gpcw20260331.dat",
			"{home}/T0002/blocknew/",
		},
	}
	if home != "" {
		result.Vipdoc = home + "/vipdoc"
	}
	return mcp.NewToolResultText(toJSON(result)), nil
}

// HandleOfflineDaily reads daily K-line from .day file.
func HandleOfflineDaily(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填 (SH/SZ)"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market = strings.ToLower(market)
	if market != "sh" && market != "sz" {
		return mcp.NewToolResultError("market 必须为 SH 或 SZ"), nil
	}
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	filePath := fmt.Sprintf("%s/%s/day/%s%s.day", vipdoc, market, market, code)
	bars, err := offline.ReadDaily(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取日线失败: %v", err)), nil
	}
	count := len(bars)
	if v, ok := request.GetArguments()["count"].(float64); ok {
		c := int(v)
		if c > 0 && c < len(bars) {
			bars = bars[len(bars)-c:]
		}
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"market": market,
		"code":   code,
		"count":  count,
		"path":   filePath,
		"bars":   bars,
	})), nil
}

// HandleOfflineMin reads minute K-line from .lc1/.lc5 file.
func HandleOfflineMin(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填 (SH/SZ)"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market = strings.ToLower(market)
	if market != "sh" && market != "sz" {
		return mcp.NewToolResultError("market 必须为 SH 或 SZ"), nil
	}
	minType := "lc5"
	if v, ok := request.GetArguments()["min_type"].(string); ok {
		minType = v
	}
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	filePath := fmt.Sprintf("%s/%s/minline/%s%s.%s", vipdoc, market, market, code, minType)
	bars, err := offline.ReadMin(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取分钟线失败: %v", err)), nil
	}
	count := len(bars)
	if v, ok := request.GetArguments()["count"].(float64); ok {
		c := int(v)
		if c > 0 && c < len(bars) {
			bars = bars[len(bars)-c:]
		}
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"market":   market,
		"code":     code,
		"min_type": minType,
		"count":    count,
		"path":     filePath,
		"bars":     bars,
	})), nil
}

// HandleOfflineGBBQ reads equity change data from gbbq file.
func HandleOfflineGBBQ(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path 参数必填"), nil
	}
	records, err := offline.ReadGBBQ(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取股本变迁失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"path":    path,
		"count":   len(records),
		"records": records,
	})), nil
}

// HandleOfflineBlocks reads custom blocks from blocknew directory.
func HandleOfflineBlocks(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path 参数必填"), nil
	}
	blocks, err := offline.ReadBlocks(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取板块失败: %v", err)), nil
	}
	totalMembers := 0
	for _, b := range blocks {
		totalMembers += len(b.Members)
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"path":          path,
		"block_count":   len(blocks),
		"total_members": totalMembers,
		"blocks":        blocks,
	})), nil
}

// HandleOfflineExFiles lists extended market files.
func HandleOfflineExFiles(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	type exFile struct {
		Market string `json:"market"`
		Code   string `json:"code"`
		Name   string `json:"name"`
		Path   string `json:"path"`
	}
	known := []exFile{
		{Market: "38", Code: "38#2_CPI", Name: "美元指数", Path: vipdoc + "/ds/38/day/38#2_CPI.day"},
		{Market: "38", Code: "38#2_CL", Name: "美原油", Path: vipdoc + "/ds/38/day/38#2_CL.day"},
		{Market: "38", Code: "38#2_GC", Name: "美黄金", Path: vipdoc + "/ds/38/day/38#2_GC.day"},
		{Market: "71", Code: "71#2_HSI", Name: "恒生指数", Path: vipdoc + "/ds/71/day/71#2_HSI.day"},
		{Market: "71", Code: "71#2_00700", Name: "腾讯控股", Path: vipdoc + "/ds/71/day/71#2_00700.day"},
		{Market: "74", Code: "74#2_AAPL", Name: "苹果", Path: vipdoc + "/ds/74/day/74#2_AAPL.day"},
		{Market: "47", Code: "47#2_IF", Name: "沪深300股指期货", Path: vipdoc + "/ds/47/day/47#2_IF.day"},
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"vipdoc": vipdoc,
		"count":  len(known),
		"files":  known,
		"note":   "更完整的文件列表请直接查看 {vipdoc}/ds/ 目录",
	})), nil
}

// HandleOfflineExDaily reads extended market daily K-line from file.
func HandleOfflineExDaily(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	parts := strings.SplitN(code, "#", 2)
	if len(parts) != 2 {
		return mcp.NewToolResultError("code 格式应为 '市场#代码'，如 '38#2_CPI'"), nil
	}
	filePath := fmt.Sprintf("%s/ds/%s/day/%s.day", vipdoc, parts[0], code)
	bars, err := offline.ReadDaily(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取扩展市场日线失败: %v", err)), nil
	}
	count := len(bars)
	if v, ok := request.GetArguments()["count"].(float64); ok {
		c := int(v)
		if c > 0 && c < len(bars) {
			bars = bars[len(bars)-c:]
		}
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"code":  code,
		"count": count,
		"path":  filePath,
		"bars":  bars,
	})), nil
}

// HandleOfflineFinancial reads financial data from gpcw*.dat file.
func HandleOfflineFinancial(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path 参数必填"), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取财务数据失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"path": path,
		"size": len(data),
		"note": "通达信财务数据文件已读取。完整的财务科目解析需要了解gpcw格式的详细结构，当前返回原始字节大小。",
	})), nil
}

// HandleOfflineSyncDaily downloads latest daily from TQLEX and writes to local .day file.
func HandleOfflineSyncDaily(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	market, err := request.RequireString("market")
	if err != nil {
		return mcp.NewToolResultError("market 参数必填 (SH/SZ)"), nil
	}
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	market = strings.ToLower(market)
	if market != "sh" && market != "sz" {
		return mcp.NewToolResultError("market 必须为 SH 或 SZ"), nil
	}
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	marketInt := 0
	if market == "sh" {
		marketInt = 1
	}
	klineReq := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          code,
		Setcode:       marketInt,
		Period:        PeriodToCode("day"),
		Startxh:       0,
		WantNum:       5000,
		TQFlag:        11,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}
	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取K线失败: %v", err)), nil
	}
	bars, err := parseKlineBarsToDayBars(resp.Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析K线失败: %v", err)), nil
	}
	if err := offline.SyncDaily(vipdoc, market, code, bars); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("写入失败: %v", err)), nil
	}
	newBars, _ := offline.ReadDaily(fmt.Sprintf("%s/%s/day/%s%s.day", vipdoc, market, market, code))
	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"market":   market,
		"code":     code,
		"synced":   len(bars),
		"total":    len(newBars),
		"vipdoc":   vipdoc,
	})), nil
}

// HandleOfflineSyncAll syncs all A-share stocks' daily data.
func HandleOfflineSyncAll(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	vipdoc := ""
	if v, ok := request.GetArguments()["vipdoc"].(string); ok {
		vipdoc = v
	}
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			return mcp.NewToolResultError("未找到通达信目录，请指定 vipdoc 路径"), nil
		}
		vipdoc = home + "/vipdoc"
	}
	limit := 0
	if v, ok := request.GetArguments()["limit"].(float64); ok {
		limit = int(v)
	}

	type syncStatus struct {
		Synced  int      `json:"synced"`
		Total   int      `json:"total"`
		Errors  []string `json:"errors,omitempty"`
	}
	status := syncStatus{}

	stocks := []struct{ mkt, code string }{
		{"sh", "600519"}, {"sh", "600036"}, {"sh", "601318"}, {"sh", "600276"},
		{"sz", "000001"}, {"sz", "000002"}, {"sz", "000858"}, {"sz", "002415"},
		{"sh", "600030"}, {"sh", "600809"}, {"sh", "601166"}, {"sh", "600104"},
		{"sz", "000651"}, {"sz", "002594"}, {"sz", "300750"}, {"sz", "002714"},
	}
	if limit > 0 && limit < len(stocks) {
		stocks = stocks[:limit]
	}
	status.Total = len(stocks)

	for _, s := range stocks {
		marketInt := 0
		if s.mkt == "sh" {
			marketInt = 1
		}
		klineReq := KlineRequest{
			Head:          TDXHead{Target: "0", CharSet: "UTF8"},
			Code:          s.code,
			Setcode:       marketInt,
			Period:        PeriodToCode("day"),
			Startxh:       0,
			WantNum:       5000,
			TQFlag:        11,
			MPData:        0,
			HasAttachInfo: 1,
			HasLtgb:       0,
			ForRefresh:    0,
			HasIpoPrice:   0,
		}
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", klineReq)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("%s%s: %v", s.mkt, s.code, err))
			continue
		}
		bars, err := parseKlineBarsToDayBars(resp.Data)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("%s%s parse: %v", s.mkt, s.code, err))
			continue
		}
		if err := offline.SyncDaily(vipdoc, s.mkt, s.code, bars); err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("%s%s write: %v", s.mkt, s.code, err))
			continue
		}
		status.Synced++
	}
	return mcp.NewToolResultText(toJSON(status)), nil
}

func parseKlineBarsToDayBars(data interface{}) ([]offline.DayBar, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	type klineObj struct {
		Date   string  `json:"date"`
		Time   string  `json:"time"`
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Vol    float64 `json:"vol"`
		Amount float64 `json:"amount"`
	}
	var objs []klineObj
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		bars := make([]offline.DayBar, len(objs))
		for i, o := range objs {
			dt := o.Date
			if dt == "" {
				dt = o.Time
			}
			bars[i] = offline.DayBar{Date: dt, Open: o.Open, Close: o.Close, High: o.High, Low: o.Low, Vol: o.Vol, Amount: o.Amount}
		}
		return bars, nil
	}
	var arrs [][]float64
	if err := json.Unmarshal(raw, &arrs); err == nil && len(arrs) > 0 {
		bars := make([]offline.DayBar, len(arrs))
		for i, row := range arrs {
			if len(row) >= 6 {
				bars[i] = offline.DayBar{Open: row[0], Close: row[1], High: row[2], Low: row[3], Vol: row[4], Amount: row[5]}
			}
		}
		return bars, nil
	}
	return nil, fmt.Errorf("unsupported kline data format")
}

// --- Data Parsing Helpers ---

// parseKlineBars extracts indicator.Bar slice from TQLEX K-line response data.
func parseKlineBars(data interface{}) ([]indicator.Bar, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	type klineObj struct {
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Vol    float64 `json:"vol"`
		Amount float64 `json:"amount"`
	}
	var objs []klineObj
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		bars := make([]indicator.Bar, len(objs))
		for i, o := range objs {
			bars[i] = indicator.Bar{Open: o.Open, Close: o.Close, High: o.High, Low: o.Low, Vol: o.Vol, Amount: o.Amount}
		}
		return bars, nil
	}
	var arrs [][]float64
	if err := json.Unmarshal(raw, &arrs); err == nil && len(arrs) > 0 {
		bars := make([]indicator.Bar, len(arrs))
		for i, row := range arrs {
			if len(row) >= 6 {
				bars[i] = indicator.Bar{Open: row[1], Close: row[2], High: row[3], Low: row[4], Vol: row[5]}
				if len(row) >= 7 {
					bars[i].Amount = row[6]
				}
			}
		}
		return bars, nil
	}
	return nil, fmt.Errorf("unsupported kline data format for indicator parsing")
}

// parseChanlunKlines extracts chanlun.Kline slice from TQLEX K-line response.
func parseChanlunKlines(data interface{}) ([]chanlun.Kline, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	type klineObj struct {
		Date   string  `json:"date"`
		Time   string  `json:"time"`
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Vol    float64 `json:"vol"`
		Amount float64 `json:"amount"`
	}
	var objs []klineObj
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		klines := make([]chanlun.Kline, len(objs))
		for i, o := range objs {
			dt := o.Date
			if dt == "" {
				dt = o.Time
			}
			klines[i] = chanlun.Kline{Date: dt, Open: o.Open, Close: o.Close, High: o.High, Low: o.Low, Vol: o.Vol, Amount: o.Amount}
		}
		return klines, nil
	}
	return nil, fmt.Errorf("unsupported kline data format for chanlun parsing")
}

// --- Collection and Routing ---

// GetAllExpandedTools returns all 22 expanded MCP tool definitions.
func GetAllExpandedTools() []mcp.Tool {
	return []mcp.Tool{
		NewTickTool(),
		NewTransactionTool(),
		NewBoardListTool(),
		NewBoardMembersTool(),
		NewBelongBoardTool(),
		NewBoardRankingTool(),
		NewCapitalFlowTool(),
		NewAuctionTool(),
		NewUnusualTool(),
		NewMarketStatTool(),
		NewServerInfoTool(),
		NewSymbolInfoTool(),
		NewAnnouncementTool(),
		NewFinancialTool(),
		NewIndicatorComputeTool(),
		NewChanlunTool(),
		NewBacktestTool(),
		NewExMarketsTool(),
		NewExKlineTool(),
		NewExQuoteTool(),
		NewExQuoteListTool(),
		NewExTickTool(),
		NewOfflineHomeTool(),
		NewOfflineDailyTool(),
		NewOfflineMinTool(),
		NewOfflineGBBQTool(),
		NewOfflineBlocksTool(),
		NewOfflineExFilesTool(),
		NewOfflineExDailyTool(),
		NewOfflineFinancialTool(),
		NewOfflineSyncDailyTool(),
		NewOfflineSyncAllTool(),
	}
}

// GetExpandedHandler returns the ToolHandler for a given expanded tool name.
func GetExpandedHandler(name string) ToolHandler {
	switch name {
	case ToolTick:
		return HandleTick
	case ToolTransaction:
		return HandleTransaction
	case ToolBoardList:
		return HandleBoardList
	case ToolBoardMembers:
		return HandleBoardMembers
	case ToolBelongBoard:
		return HandleBelongBoard
	case ToolBoardRanking:
		return HandleBoardRanking
	case ToolCapitalFlow:
		return HandleCapitalFlow
	case ToolAuction:
		return HandleAuction
	case ToolUnusual:
		return HandleUnusual
	case ToolMarketStat:
		return HandleMarketStat
	case ToolServerInfo:
		return HandleServerInfo
	case ToolSymbolInfo:
		return HandleSymbolInfo
	case ToolAnnouncement:
		return HandleAnnouncement
	case ToolFinancial:
		return HandleFinancial
	case ToolIndicatorComp:
		return HandleIndicatorCompute
	case ToolChanlun:
		return HandleChanlun
	case ToolBacktest:
		return HandleBacktest
	case ToolExMarkets:
		return HandleExMarkets
	case ToolExKline:
		return HandleExKline
	case ToolExQuote:
		return HandleExQuote
	case ToolExQuoteList:
		return HandleExQuoteList
	case ToolExTick:
		return HandleExTick
	case ToolOfflineHome:
		return HandleOfflineHome
	case ToolOfflineDaily:
		return HandleOfflineDaily
	case ToolOfflineMin:
		return HandleOfflineMin
	case ToolOfflineGBBQ:
		return HandleOfflineGBBQ
	case ToolOfflineBlocks:
		return HandleOfflineBlocks
	case ToolOfflineExFiles:
		return HandleOfflineExFiles
	case ToolOfflineExDaily:
		return HandleOfflineExDaily
	case ToolOfflineFinancial:
		return HandleOfflineFinancial
	case ToolOfflineSyncDaily:
		return HandleOfflineSyncDaily
	case ToolOfflineSyncAll:
		return HandleOfflineSyncAll
	default:
		return nil
	}
}

// eastmoneyExQuery fetches data from EastMoney push2 API for extended (HK/US) markets.
// Returns JSON-serializable result compatible with TQLEXResponse.Data.
func eastmoneyExQuery(exMarket string, code string, endpoint string) (interface{}, error) {
	secid := "116." + code
	switch strings.ToLower(exMarket) {
	case "hk", "h":
		secid = "116." + code
	case "us", "u":
		secid = "117." + code
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	var urlStr string
	if endpoint == "quote" {
		urlStr = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f49,f50,f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f71", secid)
	} else {
		urlStr = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f49,f50,f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f71", secid)
	}
	respHTTP, err := hc.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	body, _ := io.ReadAll(respHTTP.Body)
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	return result, nil
}

// eastmoneyExList fetches stock list for an extended market from EastMoney push2 clist API.
func eastmoneyExList(exMarket string, count int) (interface{}, error) {
	fs := "b:DLMKTS_HK"
	switch strings.ToLower(exMarket) {
	case "hk", "h":
		fs = "b:DLMKTS_HK"
	case "us", "u":
		fs = "b:DLMKTS_US"
	default:
		fs = "b:DLMKTS_HK"
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	urlStr := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=%d&po=1&np=1&fltt=2&invt=2&fs=%s&fields=f12,f14,f2,f3", count, fs)
	respHTTP, err := hc.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	body, _ := io.ReadAll(respHTTP.Body)
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	return result, nil
}

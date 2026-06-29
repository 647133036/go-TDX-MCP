package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ToolQuotes          = "tdx_quotes"
	ToolKline           = "tdx_kline"
	ToolLookupStock     = "tdx_lookup_stock"
	ToolScreener        = "tdx_screener"
	ToolIndicatorSelect = "tdx_indicator_select"
	ToolApiData         = "tdx_api_data"
)

func AllTools() []mcp.Tool {
	return []mcp.Tool{
		NewQuotesTool(),
		NewKlineTool(),
		NewLookupStockTool(),
		NewScreenerTool(),
		NewIndicatorSelectTool(),
		NewApiDataTool(),
	}
}

func NewQuotesTool() mcp.Tool {
	return mcp.NewTool(ToolQuotes,
		mcp.WithDescription("获取A股实时行情：报价、五档盘口、涨跌幅、成交量等"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithString("setcode",
			mcp.Required(),
			mcp.Description("市场标识: 0=深圳, 1=上海, 2=北交所"),
		),
		mcp.WithString("hasProInfo",
			mcp.Description("是否包含扩展信息，默认 '0'，传 '1' 获取板块/行业信息"),
		),
	)
}

func NewKlineTool() mcp.Tool {
	return mcp.NewTool(ToolKline,
		mcp.WithDescription("获取A股K线历史数据，支持多周期"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithNumber("setcode",
			mcp.Required(),
			mcp.Description("市场类型: 0=深圳, 1=上海, 2=北交所"),
		),
		mcp.WithNumber("period",
			mcp.Required(),
			mcp.Description("K线周期: 4=日线, 5=周线, 6=月线, 3=60分钟, 9=1分钟, 10=5分钟, 11=15分钟, 12=30分钟"),
		),
		mcp.WithNumber("wantNum",
			mcp.Description("返回K线数量 (默认100)"),
		),
		mcp.WithNumber("fqType",
			mcp.Description("复权: 0=不复权, 1=前复权, 2=后复权，通过 TQFlag 位运算实现 (默认0)"),
		),
	)
}

func NewLookupStockTool() mcp.Tool {
	return mcp.NewTool(ToolLookupStock,
		mcp.WithDescription("通过自然语言检索股票/指数/基金代码与名称（RAG语义搜索）"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("检索关键词或自然语言描述，如 '平安银行' 或 '新能源龙头'"),
		),
		mcp.WithString("range",
			mcp.Description("市场范围: AG=A股(默认), HK-GP=港股, JJ=基金, MG-GP=美股, ZS=指数"),
		),
		mcp.WithNumber("topK",
			mcp.Description("返回结果数量 (默认10)"),
		),
	)
}

func NewScreenerTool() mcp.Tool {
	return mcp.NewTool(ToolScreener,
		mcp.WithDescription("自然语言智能选股：根据描述条件筛选股票"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("自然语言选股条件，如 '涨停' 或 '主板 小盘 低价 涨停'"),
		),
		mcp.WithString("rang",
			mcp.Description("市场范围，默认 'AG' (A股)"),
		),
		mcp.WithNumber("pageNo",
			mcp.Description("页码 (默认1)"),
		),
		mcp.WithNumber("pageSize",
			mcp.Description("每页数量 (默认10)"),
		),
	)
}

func NewIndicatorSelectTool() mcp.Tool {
	return mcp.NewTool(ToolIndicatorSelect,
		mcp.WithDescription("金融指标选择与查询：查询财务/技术/估值指标"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("指标查询描述，如 '000001 技术指标' 或 '银行业估值对比'"),
		),
		mcp.WithString("rang",
			mcp.Description("市场范围，默认 'AG' (A股)"),
		),
	)
}

func NewApiDataTool() mcp.Tool {
	return mcp.NewTool(ToolApiData,
		mcp.WithDescription("统一F10内部API调用：公司概况、盈利预测、热点题材、龙虎榜、机构持仓等"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如 '000001'"),
		),
		mcp.WithString("entry",
			mcp.Required(),
			mcp.Description("F10 Entry名称，如 'TdxSharePCCW.tdxf10_gg_gsgk'(公司概况), 'TdxSharePCCW.tdxf10_gg_ybpj'(盈利预测), 'TdxSharePCCW.tdxf10_gg_rdtc'(热点题材)"),
		),
		mcp.WithString("fixedTag",
			mcp.Description("固定标签/子模块标识，如 'gsgy'(公司概要), 'yzyq'(盈利预测), 'zttzbkz'(主题投资板块族谱)"),
		),
		mcp.WithString("extra",
			mcp.Description("额外参数（日期、页码等），JSON数组格式字符串"),
		),
	)
}

type ToolHandler func(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

func CreateToolHandler(client Client, h ToolHandler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h(ctx, client, request)
	}
}

func GetHandler(name string) ToolHandler {
	switch name {
	case ToolQuotes:
		return HandleQuotes
	case ToolKline:
		return HandleKline
	case ToolLookupStock:
		return HandleLookupStock
	case ToolScreener:
		return HandleScreener
	case ToolIndicatorSelect:
		return HandleIndicatorSelect
	case ToolApiData:
		return HandleApiData
	default:
		return nil
	}
}

func toJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func HandleQuotes(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	setcode := "0"
	if v := request.GetString("setcode", ""); v != "" {
		setcode = v
	}
	hasProInfo := "0"
	if v := request.GetString("hasProInfo", ""); v != "" {
		hasProInfo = v
	}

	reqBody := QuoteRequest{
		Head:        TDXHead{Target: "0", CharSet: "UTF8"},
		Code:        strings.TrimSpace(code),
		Setcode:     setcode,
		HasHQInfo:   "1",
		HasExtInfo:  "1",
		BspNum:      "5",
		HasProInfo:  hasProInfo,
		HasCalcInfo: "0",
		HasCwInfo:   "0",
		HasStatInfo: "0",
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBHQInfo", reqBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("行情查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleKline(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	setcode, _ := request.RequireFloat("setcode")
	period, _ := request.RequireFloat("period")
	wantNum := 100.0
	if v := request.GetFloat("wantNum", 0); v > 0 {
		wantNum = v
	}
	fqType := 0.0
	if v := request.GetFloat("fqType", 0); v > 0 {
		fqType = v
	}

	tqFlag := 11
	if fqType == 1 {
		tqFlag |= 0x01
	} else if fqType == 2 {
		tqFlag |= 0x02
	}

	reqBody := KlineRequest{
		Head:          TDXHead{Target: "0", CharSet: "UTF8"},
		Code:          strings.TrimSpace(code),
		Setcode:       int(setcode),
		Period:        int(period),
		Startxh:       0,
		WantNum:       int(wantNum),
		TQFlag:        tqFlag,
		MPData:        0,
		HasAttachInfo: 1,
		HasLtgb:       0,
		ForRefresh:    0,
		HasIpoPrice:   0,
	}

	resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", reqBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("K线查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleLookupStock(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query 参数必填"), nil
	}
	topK := 10
	if v := request.GetFloat("topK", 0); v > 0 {
		topK = int(v)
	}

	resp, err := client.RAGQuery(ctx, strings.TrimSpace(query), topK)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("代码检索失败: %v", err)), nil
	}

	var result strings.Builder
	for i, r := range resp.Results {
		result.WriteString(fmt.Sprintf("%d. %s (%s) - %s [%.2f]\n", i+1, r.Name, r.Code, r.Type, r.Score))
	}
	if result.Len() == 0 {
		result.WriteString("未找到匹配结果")
	}
	return mcp.NewToolResultText(result.String()), nil
}

func HandleScreener(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, err := request.RequireString("message")
	if err != nil {
		return mcp.NewToolResultError("message 参数必填"), nil
	}
	rang := "AG"
	if v := request.GetString("rang", ""); v != "" {
		rang = v
	}
	pageNo := "1"
	if v := request.GetFloat("pageNo", 0); v > 0 {
		pageNo = fmt.Sprintf("%d", int(v))
	}
	pageSize := "10"
	if v := request.GetFloat("pageSize", 0); v > 0 {
		pageSize = fmt.Sprintf("%d", int(v))
	}

	reqBody := ScreenerRequest{{
		Message:  strings.TrimSpace(message),
		Rang:     rang,
		PageNo:   pageNo,
		PageSize: pageSize,
	}}

	resp, err := client.TQLEXQuery(ctx, "JNLPSE:wendaQuery", reqBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("智能选股失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleIndicatorSelect(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, err := request.RequireString("message")
	if err != nil {
		return mcp.NewToolResultError("message 参数必填"), nil
	}
	rang := "AG"
	if v := request.GetString("rang", ""); v != "" {
		rang = v
	}

	reqBody := IndicatorSelectRequest{
		Message: strings.TrimSpace(message),
		Rang:    rang,
	}

	resp, err := client.TQLEXQuery(ctx, "NLPSE:InfoSelectV2", reqBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("指标查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

func HandleApiData(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError("code 参数必填"), nil
	}
	entry, err := request.RequireString("entry")
	if err != nil {
		return mcp.NewToolResultError("entry 参数必填"), nil
	}
	fixedTag := request.GetString("fixedTag", "")
	extra := request.GetString("extra", "")

	params := []interface{}{strings.TrimSpace(code)}
	if fixedTag != "" {
		params = append(params, fixedTag)
	}
	if extra != "" {
		params = append(params, extra)
	}

	reqBody := ApiDataRequest{Params: params}

	resp, err := client.TQLEXQuery(ctx, entry, reqBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("F10查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

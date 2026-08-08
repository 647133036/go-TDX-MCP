package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bensema/gotdx/proto"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tdx/go-tdx-mcp/formula"
	"github.com/tdx/go-tdx-mcp/indicator"
)

type securityLister interface {
	GetSecurityList(market int, start uint16) (*proto.GetSecurityListReply, error)
	GetSecurityCount(market int) (uint16, error)
}

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
		mcp.WithDescription("股票信息查找：通过代码精确匹配或名称关键词模糊匹配，从TDX TCP安全列表检索股票代码、名称、实时价格"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("查询内容：6位股票代码（如000001）或股票名称关键词（如平安、茅台）"),
		),
		mcp.WithNumber("topK",
			mcp.Description("返回结果数量（默认10）"),
		),
	)
}

func NewScreenerTool() mcp.Tool {
	return mcp.NewTool(ToolScreener,
		mcp.WithDescription("通达信公式选股：在全市场范围内用通达信公式筛选符合条件的股票"),
		mcp.WithString("formula",
			mcp.Required(),
			mcp.Description("通达信选股公式，如 'CLOSE > MA(CLOSE, 20)' 或 'CLOSE > REF(CLOSE,1)*1.095'"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场: 0=深圳, 1=上海, 2=北交所 (默认0)"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期: day/week/month/1min/5min/15min/30min/60min (默认day)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线根数 (默认200)"),
		),
	)
}

func NewIndicatorSelectTool() mcp.Tool {
	return mcp.NewTool(ToolIndicatorSelect,
		mcp.WithDescription("通达信公式选股（别名）：同 tdx_screener，用通达信公式在全市场筛选股票"),
		mcp.WithString("formula",
			mcp.Required(),
			mcp.Description("通达信选股公式"),
		),
		mcp.WithNumber("market",
			mcp.Description("市场: 0=深圳, 1=上海, 2=北交所 (默认0)"),
		),
		mcp.WithString("period",
			mcp.Description("K线周期 (默认day)"),
		),
		mcp.WithNumber("count",
			mcp.Description("K线根数 (默认200)"),
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

type Handler func(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

func CreateHandler(client Client, h Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h(ctx, client, request)
	}
}

func GetHandler(name string) Handler {
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
		if h := GetV3Handler(name); h != nil {
			return h
		}
		if h := GetExpandedHandler(name); h != nil {
			return h
		}
		if h := GetNewHandler(name); h != nil {
			return h
		}
		return nil
	}
}

func toJSON(v interface{}) string {
	if v == nil {
		return "null"
	}
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
	if resp.Data == nil {
		return mcp.NewToolResultError("行情查询返回空数据"), nil
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
	if resp.Data == nil {
		return mcp.NewToolResultError("K线查询返回空数据"), nil
	}

	// Serialize resp.Data to []byte regardless of type
	var raw []byte
	switch d := resp.Data.(type) {
	case json.RawMessage:
		raw = []byte(d)
	case []byte:
		raw = d
	case string:
		raw = []byte(d)
	default:
		marshaled, err := json.Marshal(d)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("序列化K线数据失败: %v", err)), nil
		}
		raw = marshaled
	}

	// TCP format: []map[string]interface{} {Year, Open, High, Low, Close, Volume, Amount}
	var tcpBars []map[string]interface{}
	if json.Unmarshal(raw, &tcpBars) == nil && len(tcpBars) > 0 {
		// Only process as TCP format if it has TCP-specific fields (capitalized keys)
		if len(tcpBars) > 0 && tcpBars[0]["Year"] != nil {
			results := make([]map[string]interface{}, 0, len(tcpBars))
			for _, bm := range tcpBars {
				kline := make(map[string]interface{})
				if yr, ok := bm["Year"]; ok {
					ymd := toFloat64v(yr)
					if ymd > 0 {
						yi := int(ymd)
						y := yi / 10000
						m := (yi / 100) % 100
						d := yi % 100
						if m > 0 {
							kline["date"] = fmt.Sprintf("%04d%02d%02d", y, m, d)
						}
					}
				}
				kline["open"] = toFloat64v(bm["Open"])
				kline["high"] = toFloat64v(bm["High"])
				kline["low"] = toFloat64v(bm["Low"])
				kline["close"] = toFloat64v(bm["Close"])
				kline["volume"] = toFloat64v(bm["Volume"])
				kline["amount"] = toFloat64v(bm["Amount"])
				results = append(results, kline)
			}
			return mcp.NewToolResultText(toJSON(results)), nil
		}
	}

	// HTTP format (Sina, East Money, etc.): return as-is
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

// toFloat64v safely extracts a float64 from various JSON-unmarshaled types (tools.go local copy)
func toFloat64v(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch f := v.(type) {
	case float64:
		return f
	case float32:
		return float64(f)
	case int:
		return float64(f)
	case int64:
		return float64(f)
	case uint64:
		return float64(f)
	case uint32:
		return float64(f)
	case int32:
		return float64(f)
	case string:
		ff, _ := strconv.ParseFloat(f, 64)
		return ff
	}
	return 0
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

	// Use TDX TCP security list + keyword matching
	type securityLister interface {
		GetSecurityCount(market int) (uint16, error)
		GetSecurityList(market int, start uint16) (*proto.GetSecurityListReply, error)
	}
	sl, ok := client.(securityLister)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持股票信息查询（无法访问安全列表）"), nil
	}

	// Fetch securities from both markets
	var allSecs []proto.Security
	for _, market := range []int{0, 1} {
		count, err := sl.GetSecurityCount(market)
		if err != nil {
			continue
		}
		for start := uint16(0); start < count; start += 1600 {
			reply, err := sl.GetSecurityList(market, start)
			if err != nil {
				continue
			}
			if reply != nil {
				allSecs = append(allSecs, reply.List...)
			}
		}
	}

	// Keyword match
	matches := make([]struct {
		Code  string  `json:"code"`
		Name  string  `json:"name"`
		Score float64 `json:"score"`
	}, 0, topK)

	for _, sec := range allSecs {
		if sec.Name == "" || sec.Code == "" {
			continue
		}
		score := keywordMatchScore(query, sec.Name)
		if score > 0 {
			matches = append(matches, struct {
				Code  string  `json:"code"`
				Name  string  `json:"name"`
				Score float64 `json:"score"`
			}{Code: sec.Code, Name: sec.Name, Score: score})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > topK {
		matches = matches[:topK]
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(toJSON(map[string]interface{}{
			"query": query,
			"total": 0,
			"results": []string{},
			"note": "未找到匹配结果。请提供准确的股票代码（如 000001）或股票名称关键词。",
		})), nil
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"query":      query,
		"total":      len(matches),
		"match_type": "keyword",
		"results":    matches,
	})), nil
}

// keywordMatchScore computes a relevance score for matching query keywords
// against a stock name. Higher score = more relevant.
func keywordMatchScore(query string, name string) float64 {
	score := 0.0

	// Exact match
	if query == name {
		return 100.0
	}

	// Contains match: query contains stock name (e.g., "平安" matches "平安银行")
	if strings.Contains(query, name) {
		return 90.0
	}

	// Contains match: stock name contains query (e.g., "银行" matches "平安银行")
	if strings.Contains(name, query) {
		return 80.0
	}

	// Character overlap
	queryChars := strings.Split(query, "")
	nameChars := strings.Split(name, "")
	matchCount := 0
	for _, c := range queryChars {
		if strings.Contains(strings.Join(nameChars, ""), c) {
			matchCount++
		}
	}

	if len(queryChars) > 0 {
		ratio := float64(matchCount) / float64(len(queryChars))
		if ratio >= 0.5 {
			score = ratio * 50.0
		}
	}

	return score
}

func HandleScreener(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	formulaCode, err := request.RequireString("formula")
	if err != nil {
		return mcp.NewToolResultError("formula 参数必填（通达信选股公式，如 'CLOSE > MA(CLOSE, 20)'）"), nil
	}
	market := 0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = int(v)
	}
	period := "day"
	if p, ok := request.GetArguments()["period"].(string); ok && p != "" {
		period = p
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok && v > 0 {
		count = int(v)
	}

	// 获取证券列表
	sl, ok := client.(securityLister)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券列表查询"), nil
	}
	totalCount, err := sl.GetSecurityCount(market)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取证券数量失败: %v", err)), nil
	}
	allSecs := make([]proto.Security, 0)
	for start := uint16(0); start < totalCount; start += 1600 {
		reply, err := sl.GetSecurityList(market, start)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取证券列表失败: %v", err)), nil
		}
		if reply != nil {
			allSecs = append(allSecs, reply.List...)
		}
	}

	// 逐个执行公式选股
	kc, ok := client.(klineQueryClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	matches := make([]map[string]interface{}, 0)
	eng := formula.NewEngine()

	for _, sec := range allSecs {
		code := sec.Code
		if code == "" {
			continue
		}
		bars, err := kc.KlineQuery(ctx, code, market, period, count, 0)
		if err != nil {
			continue
		}
		if len(bars) == 0 {
			continue
		}

		indicatorBars := make([]indicator.Bar, len(bars))
		for i, b := range bars {
			indicatorBars[i] = indicator.Bar{
				Open: b.Open, High: b.High, Low: b.Low, Close: b.Close,
				Vol: b.Vol, Amount: b.Amount,
			}
		}

		result, err := eng.Execute(formulaCode, indicatorBars)
		if err != nil {
			continue
		}

		lastValue := 0.0
		if len(result.Outputs) > 0 {
			if data := result.Outputs[0].Data; len(data) > 0 {
				lastValue = data[len(data)-1]
			}
		}

		if lastValue > 0 {
			matches = append(matches, map[string]interface{}{
				"code":     code,
				"name":     sec.Name,
				"setcode":  code,
				"last_val": lastValue,
				"close":    bars[len(bars)-1].Close,
			})
		}
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"formula":   formulaCode,
		"total":     len(allSecs),
		"matched":   len(matches),
		"period":    period,
		"count":     count,
		"results":   matches,
	})), nil
}

func HandleIndicatorSelect(ctx context.Context, client Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	formulaCode, err := request.RequireString("formula")
	if err != nil {
		return mcp.NewToolResultError("formula 参数必填（通达信选股公式）"), nil
	}
	market := 0
	if v, ok := request.GetArguments()["market"].(float64); ok {
		market = int(v)
	}
	period := "day"
	if p, ok := request.GetArguments()["period"].(string); ok && p != "" {
		period = p
	}
	count := 200
	if v, ok := request.GetArguments()["count"].(float64); ok && v > 0 {
		count = int(v)
	}

	sl, ok := client.(securityLister)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持证券列表查询"), nil
	}
	totalCount, err := sl.GetSecurityCount(market)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取证券数量失败: %v", err)), nil
	}
	allSecs := make([]proto.Security, 0)
	for start := uint16(0); start < totalCount; start += 1600 {
		reply, err := sl.GetSecurityList(market, start)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取证券列表失败: %v", err)), nil
		}
		if reply != nil {
			allSecs = append(allSecs, reply.List...)
		}
	}

	kc, ok := client.(klineQueryClient)
	if !ok {
		return mcp.NewToolResultError("当前客户端不支持K线查询"), nil
	}
	matches := make([]map[string]interface{}, 0)
	eng := formula.NewEngine()

	for _, sec := range allSecs {
		code := sec.Code
		if code == "" {
			continue
		}
		bars, err := kc.KlineQuery(ctx, code, market, period, count, 0)
		if err != nil {
			continue
		}
		if len(bars) == 0 {
			continue
		}

		indicatorBars := make([]indicator.Bar, len(bars))
		for i, b := range bars {
			indicatorBars[i] = indicator.Bar{
				Open: b.Open, High: b.High, Low: b.Low, Close: b.Close,
				Vol: b.Vol, Amount: b.Amount,
			}
		}

		result, err := eng.Execute(formulaCode, indicatorBars)
		if err != nil {
			continue
		}

		lastValue := 0.0
		if len(result.Outputs) > 0 {
			if data := result.Outputs[0].Data; len(data) > 0 {
				lastValue = data[len(data)-1]
			}
		}

		if lastValue > 0 {
			matches = append(matches, map[string]interface{}{
				"code":     code,
				"name":     sec.Name,
				"setcode":  code,
				"last_val": lastValue,
				"close":    bars[len(bars)-1].Close,
			})
		}
	}

	return mcp.NewToolResultText(toJSON(map[string]interface{}{
		"formula":   formulaCode,
		"total":     len(allSecs),
		"matched":   len(matches),
		"period":    period,
		"count":     count,
		"results":   matches,
	})), nil
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
	if resp.Data == nil {
		return mcp.NewToolResultError("F10查询返回空数据"), nil
	}
	return mcp.NewToolResultText(toJSON(resp.Data)), nil
}

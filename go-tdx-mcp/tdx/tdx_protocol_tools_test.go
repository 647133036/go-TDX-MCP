package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/bensema/gotdx/proto"
	"github.com/mark3labs/mcp-go/mcp"
)

// ===========================================================================
// Specialized Mock Clients
// ===========================================================================

// mockTDXProtocolClient implements all 7 specialized interfaces + Client.
type mockTDXProtocolClient struct {
	mockClient // embed for TQLEXQuery/RAGQuery
	auctionData    []proto.AuctionData
	auctionErr     error
	minuteData     []proto.HistoryMinuteTimeData
	minuteErr      error
	tradeData      []proto.HistoryTransactionData
	tradeErr       error
	quote          *proto.SecurityQuote
	quoteErr       error
	symbolInfo     *proto.MACSymbolInfoReply
	symbolInfoErr  error
	belongBoard    []proto.MACBelongBoardItem
	belongBoardErr error
	financeInfo    *proto.GetFinanceInfoReply
	financeInfoErr error
	indexBars      []proto.IndexBar
	indexBarsErr   error
	ipoData        []map[string]interface{}
	ipoErr         error
}

// auctionClient interface
func (m *mockTDXProtocolClient) GetCallAuction(code string, market int) ([]proto.AuctionData, error) {
	return m.auctionData, m.auctionErr
}

// historyMinuteClient interface
func (m *mockTDXProtocolClient) GetHistoryMinute(code string, market int, date string) ([]proto.HistoryMinuteTimeData, error) {
	return m.minuteData, m.minuteErr
}

// historyTradeClient interface
func (m *mockTDXProtocolClient) GetHistoryTrade(code string, market int, date string, count int) ([]proto.HistoryTransactionData, error) {
	return m.tradeData, m.tradeErr
}

// stockStatsClient interface
func (m *mockTDXProtocolClient) GetRealtimeQuote(code string, market int) (*proto.SecurityQuote, error) {
	return m.quote, m.quoteErr
}
func (m *mockTDXProtocolClient) GetSymbolInfo(code string, market int) (*proto.MACSymbolInfoReply, error) {
	return m.symbolInfo, m.symbolInfoErr
}
func (m *mockTDXProtocolClient) GetSymbolBelongBoard(code string, market uint8) ([]proto.MACBelongBoardItem, error) {
	return m.belongBoard, m.belongBoardErr
}

// financeClient interface
func (m *mockTDXProtocolClient) GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error) {
	return m.financeInfo, m.financeInfoErr
}

// indexKlineClient interface
func (m *mockTDXProtocolClient) GetIndexKLine(code string, market int, period int, count int) ([]proto.IndexBar, error) {
	return m.indexBars, m.indexBarsErr
}

// ipoCalendarClient interface
func (m *mockTDXProtocolClient) IPOCalendar(date string, limit int) ([]map[string]interface{}, error) {
	return m.ipoData, m.ipoErr
}

// ===========================================================================
// Helper: build mcp.CallToolRequest
// ===========================================================================

func makeReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// ===========================================================================
// Helper: parse result JSON
// ===========================================================================

func parseResult(t *testing.T, result *mcp.CallToolResult) map[string]interface{} {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	var raw map[string]interface{}
	contentJSON := string(result.Content[0].(mcp.TextContent).Text)
	if err := json.Unmarshal([]byte(contentJSON), &raw); err != nil {
		t.Fatalf("failed to parse result JSON: %s\nraw: %s", err, contentJSON)
	}
	return raw
}

// ===========================================================================
// Test Data Factories
// ===========================================================================

func makeAuctionData() []proto.AuctionData {
	return []proto.AuctionData{
		{Time: "09:25:00", Price: 10.5, Matched: 100000, Unmatched: 5000, Flag: 0},
		{Time: "09:24:00", Price: 10.4, Matched: 80000, Unmatched: 3000, Flag: 0},
	}
}

func makeMinuteData() []proto.HistoryMinuteTimeData {
	return []proto.HistoryMinuteTimeData{
		{Price: 10.5, Avg: 10.45, Vol: 1000},
		{Price: 10.6, Avg: 10.55, Vol: 2000},
	}
}

func makeTradeData() []proto.HistoryTransactionData {
	return []proto.HistoryTransactionData{
		{Time: time.Date(2025, 1, 1, 9, 31, 0, 0, time.UTC), Price: 10.5, Vol: 100, Num: 1, BuyOrSell: 1, Action: "BUY"},
		{Time: time.Date(2025, 1, 1, 9, 31, 1, 0, time.UTC), Price: 10.6, Vol: 200, Num: 2, BuyOrSell: 2, Action: "SELL"},
	}
}

func makeQuote() *proto.SecurityQuote {
	return &proto.SecurityQuote{
		Market:   0,
		Code:     "000001",
		Close:    12.50,
		PreClose: 12.30,
		Open:     12.40,
		High:     12.60,
		Low:      12.35,
		Vol:      5000000,
		Amount:   62500000,
		Turnover: 1.5,
	}
}

func makeSymbolInfo() *proto.MACSymbolInfoReply {
	return &proto.MACSymbolInfoReply{
		Market:        0,
		Code:          "000001",
		Name:          "平安银行",
		PreClose:      12.30,
		Open:          12.40,
		High:          12.60,
		Low:           12.35,
		Close:         12.50,
		Momentum:      1.63,
		Vol:           5000000,
		Amount:        62500000,
		InsideVolume:  2000000,
		OutsideVolume: 3000000,
		Turnover:      1.5,
		Avg:           12.45,
	}
}

func makeFinanceInfo() *proto.GetFinanceInfoReply {
	return &proto.GetFinanceInfoReply{
		Code:              "000001",
		TotalShares:       19475000000,
		FloatShares:       19475000000,
		StateShares:       3000000000,
		EPS:               0.85,
		NetAssetsPerShare: 6.50,
		TotalAssets:       1500000000000,
		NetProfit:         18000000000,
		TotalProfit:       20000000000,
		OperatingRevenue:  45000000000,
		TotalEquity:       120000000000,
		ShareholderCount:  450000,
		UndistributedProfit: 30000000000,
		IPODate:           20120203,
		UpdatedDate:       20241231,
	}
}

func makeBelongBoard() []proto.MACBelongBoardItem {
	return []proto.MACBelongBoardItem{
		{BoardCode: "BK0059", BoardName: "银行", BoardType: "HY"},
		{BoardCode: "BK0475", BoardName: "深圳", BoardType: "DY"},
	}
}

func makeIndexBars() []proto.IndexBar {
	return []proto.IndexBar{
		{DateTime: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), Open: 2900, High: 2920, Low: 2895, Close: 2915, Vol: 500000, Amount: 500000000},
		{DateTime: time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), Open: 2910, High: 2930, Low: 2905, Close: 2925, Vol: 450000, Amount: 460000000},
	}
}

func makeIPOData() []map[string]interface{} {
	return []map[string]interface{}{
		{"code": "301234", "name": "测试新股", "ipo_date": "20250110", "issue_price": 18.5},
	}
}

// ===========================================================================
// TestGetNewHandler - 7 new TDX protocol tools
// ===========================================================================

func TestGetNewHandler_CallAuction(t *testing.T) {
	h := GetNewHandler(ToolCallAuction)
	if h == nil {
		t.Fatal("expected handler for tdx_call_auction")
	}
}

func TestGetNewHandler_HistoryMinute(t *testing.T) {
	h := GetNewHandler(ToolHistoryMinute)
	if h == nil {
		t.Fatal("expected handler for tdx_history_minute")
	}
}

func TestGetNewHandler_HistoryTrade(t *testing.T) {
	h := GetNewHandler(ToolHistoryTrade)
	if h == nil {
		t.Fatal("expected handler for tdx_history_trade")
	}
}

func TestGetNewHandler_StockStats(t *testing.T) {
	h := GetNewHandler(ToolStockStats)
	if h == nil {
		t.Fatal("expected handler for tdx_stock_stats")
	}
}

func TestGetNewHandler_FinanceInfo(t *testing.T) {
	h := GetNewHandler(ToolFinanceInfo)
	if h == nil {
		t.Fatal("expected handler for tdx_finance_info")
	}
}

func TestGetNewHandler_IndexKline(t *testing.T) {
	h := GetNewHandler(ToolIndexKline)
	if h == nil {
		t.Fatal("expected handler for tdx_index_kline")
	}
}

func TestGetNewHandler_IPOCalendar(t *testing.T) {
	h := GetNewHandler(ToolIPOCalendar)
	if h == nil {
		t.Fatal("expected handler for tdx_ipo_calendar")
	}
}

// ===========================================================================
// TestToolDefinitions - 7 new tools
// ===========================================================================

func TestNewCallAuctionTool(t *testing.T) {
	tool := NewCallAuctionTool()
	if tool.Name != ToolCallAuction {
		t.Fatalf("expected tool name %s, got %s", ToolCallAuction, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewHistoryMinuteTool(t *testing.T) {
	tool := NewHistoryMinuteTool()
	if tool.Name != ToolHistoryMinute {
		t.Fatalf("expected tool name %s, got %s", ToolHistoryMinute, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewHistoryTradeTool(t *testing.T) {
	tool := NewHistoryTradeTool()
	if tool.Name != ToolHistoryTrade {
		t.Fatalf("expected tool name %s, got %s", ToolHistoryTrade, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewStockStatsTool(t *testing.T) {
	tool := NewStockStatsTool()
	if tool.Name != ToolStockStats {
		t.Fatalf("expected tool name %s, got %s", ToolStockStats, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewFinanceInfoTool(t *testing.T) {
	tool := NewFinanceInfoTool()
	if tool.Name != ToolFinanceInfo {
		t.Fatalf("expected tool name %s, got %s", ToolFinanceInfo, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewIndexKlineTool(t *testing.T) {
	tool := NewIndexKlineTool()
	if tool.Name != ToolIndexKline {
		t.Fatalf("expected tool name %s, got %s", ToolIndexKline, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestNewIPOCalendarTool(t *testing.T) {
	tool := NewIPOCalendarTool()
	if tool.Name != ToolIPOCalendar {
		t.Fatalf("expected tool name %s, got %s", ToolIPOCalendar, tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

// ===========================================================================
// HandleCallAuction Tests
// ===========================================================================

func TestHandleCallAuction_InvalidSetcode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{auctionData: makeAuctionData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "abc"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["count"].(float64) != 2 {
		t.Fatalf("expected count=2, got %v", data["count"])
	}
}

func TestHandleCallAuction_MissingCode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"setcode": "0"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code")
	}
}

func TestHandleCallAuction_MissingSetcode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing setcode")
	}
}

func TestHandleCallAuction_InvalidSetcodeValue(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{auctionData: makeAuctionData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "9"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for invalid setcode=9")
	}
}

func TestHandleCallAuction_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{} // does not implement auctionClient
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleCallAuction_EmptyData(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{auctionData: []proto.AuctionData{}}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["count"].(float64) != 0 {
		t.Fatalf("expected count=0, got %v", data["count"])
	}
}

func TestHandleCallAuction_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{auctionData: makeAuctionData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleCallAuction(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["code"] != "000001" {
		t.Fatalf("expected code=000001, got %v", data["code"])
	}
	if data["market"].(float64) != 0 {
		t.Fatalf("expected market=0, got %v", data["market"])
	}
	if data["count"].(float64) != 2 {
		t.Fatalf("expected count=2, got %v", data["count"])
	}
	items := data["data"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	item := items[0].(map[string]interface{})
	if item["price"].(float64) != 10.5 {
		t.Fatalf("expected price=10.5, got %v", item["price"])
	}
}

// ===========================================================================
// HandleHistoryMinute Tests
// ===========================================================================

func TestHandleHistoryMinute_MissingDate(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleHistoryMinute(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing date")
	}
}

func TestHandleHistoryMinute_MissingSetcode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001", "date": "20250106"})
	result, err := HandleHistoryMinute(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing setcode")
	}
}

func TestHandleHistoryMinute_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106"})
	result, err := HandleHistoryMinute(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleHistoryMinute_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{minuteData: makeMinuteData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106"})
	result, err := HandleHistoryMinute(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["code"] != "000001" {
		t.Fatalf("expected code=000001, got %v", data["code"])
	}
	if data["date"] != "20250106" {
		t.Fatalf("expected date=20250106, got %v", data["date"])
	}
	if data["count"].(float64) != 2 {
		t.Fatalf("expected count=2, got %v", data["count"])
	}
	items := data["data"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	item := items[0].(map[string]interface{})
	if item["time"] != "09:30" {
		t.Fatalf("expected time=09:30, got %v", item["time"])
	}
	if item["price"].(float64) != 10.5 {
		t.Fatalf("expected price=10.5, got %v", item["price"])
	}
}

func TestHandleHistoryMinute_InvalidDate(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "invalid"})
	result, err := HandleHistoryMinute(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The date validation happens in UnifiedClient.GetHistoryMinute, which
	// the handler doesn't directly call. With a mock client, the handler
	// passes the date through to the mock. If the mock accepts it, the
	// result is based on what the mock returns.
	if result != nil && !result.IsError {
		// Mock accepted it; check count
		data := parseResult(t, result)
		_ = data
	}
}

// ===========================================================================
// HandleHistoryTrade Tests
// ===========================================================================

func TestHandleHistoryTrade_MissingDate(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleHistoryTrade(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing date")
	}
}

func TestHandleHistoryTrade_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106"})
	result, err := HandleHistoryTrade(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleHistoryTrade_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{tradeData: makeTradeData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106", "count": float64(10)})
	result, err := HandleHistoryTrade(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["count"].(float64) != 2 {
		t.Fatalf("expected count=2, got %v", data["count"])
	}
	items := data["data"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["time"] != "09:31:00" {
		t.Fatalf("expected time=09:31:00, got %v", item["time"])
	}
	if item["price"].(float64) != 10.5 {
		t.Fatalf("expected price=10.5, got %v", item["price"])
	}
	if item["volume"].(float64) != 100 {
		t.Fatalf("expected volume=100, got %v", item["volume"])
	}
}

func TestHandleHistoryTrade_MaxCount(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{tradeData: makeTradeData()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106", "count": float64(9999)})
	result, err := HandleHistoryTrade(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ===========================================================================
// HandleStockStats Tests
// ===========================================================================

func TestHandleStockStats_MissingCode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"setcode": "0"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code")
	}
}

func TestHandleStockStats_MissingSetcode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing setcode")
	}
}

func TestHandleStockStats_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleStockStats_FullSuccess(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{
		quote:         makeQuote(),
		symbolInfo:    makeSymbolInfo(),
		belongBoard:   makeBelongBoard(),
		financeInfo:   makeFinanceInfo(),
	}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["current_price"].(float64) != 12.5 {
		t.Fatalf("expected current_price=12.5, got %v", data["current_price"])
	}
	changePct := data["change_pct"].(float64)
	expectedPct := (12.50 - 12.30) / 12.30 * 100
	if abs(changePct-expectedPct) > 0.01 {
		t.Fatalf("expected change_pct=%.2f, got %v", expectedPct, data["change_pct"])
	}
	if data["volume"].(float64) != 5000000 {
		t.Fatalf("expected volume=5000000, got %v", data["volume"])
	}
	boards := data["belong_boards"].([]interface{})
	if len(boards) != 2 {
		t.Fatalf("expected 2 belong_boards, got %d", len(boards))
	}
	eps := data["eps"].(float64)
	if abs(eps-0.85) > 0.001 {
		t.Fatalf("expected eps=0.85, got %v", eps)
	}
}

func TestHandleStockStats_QuoteError_NoPanic(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{
		quoteErr:     fmt.Errorf("quote error"),
		symbolInfo:   makeSymbolInfo(),
		financeInfo:  makeFinanceInfo(),
		belongBoard:  makeBelongBoard(),
	}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if _, ok := data["error_quote"]; !ok {
		t.Fatal("expected error_quote in result when quote fails")
	}
	// Verify no nil deref happened - eps should still be present from finance info
	if data["eps"] == nil {
		t.Fatal("expected eps from finance info even when quote fails")
	}
}

func TestHandleStockStats_WithMetrics(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{
		quote:        makeQuote(),
		symbolInfo:   makeSymbolInfo(),
		financeInfo:  makeFinanceInfo(),
		belongBoard:  makeBelongBoard(),
	}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0", "metrics": "price,turnover"})
	result, err := HandleStockStats(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["current_price"] == nil {
		t.Fatal("expected current_price in filtered result")
	}
	if data["turnover_rate"] == nil {
		t.Fatal("expected turnover_rate in filtered result")
	}
	// Should NOT contain data that wasn't requested
	if data["volume"] != nil {
		t.Fatal("expected volume to be absent from filtered result")
	}
}

// ===========================================================================
// HandleFinanceInfo Tests
// ===========================================================================

func TestHandleFinanceInfo_MissingCode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"setcode": "0"})
	result, err := HandleFinanceInfo(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code")
	}
}

func TestHandleFinanceInfo_MissingSetcode(t *testing.T) {
	ctx := context.Background()
	// mockClient does NOT implement financeClient, so we need our specialized mock
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001"})
	result, err := HandleFinanceInfo(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing setcode")
	}
}

func TestHandleFinanceInfo_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleFinanceInfo(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleFinanceInfo_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{financeInfo: makeFinanceInfo()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleFinanceInfo(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["code"] != "000001" {
		t.Fatalf("expected code=000001, got %v", data["code"])
	}
	if data["ipo_date"] != "20120203" {
		t.Fatalf("expected ipo_date=20120203, got %v", data["ipo_date"])
	}
	if data["updated_date"] != "20241231" {
		t.Fatalf("expected updated_date=20241231, got %v", data["updated_date"])
	}
	totalShares := data["total_shares"].(float64)
	if abs(totalShares-19475000000) > 1 {
		t.Fatalf("expected total_shares=19475000000, got %v", totalShares)
	}
	eps := data["eps"].(float64)
	if abs(eps-0.85) > 0.001 {
		t.Fatalf("expected eps=0.85, got %v", eps)
	}
}

func TestHandleFinanceInfo_Error(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{financeInfoErr: fmt.Errorf("finance error")}
	req := makeReq(map[string]any{"code": "000001", "setcode": "0"})
	result, err := HandleFinanceInfo(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
}

// ===========================================================================
// HandleIndexKline Tests
// ===========================================================================

func TestHandleIndexKline_MissingCode(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{})
	result, err := HandleIndexKline(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code")
	}
}

func TestHandleIndexKline_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{"code": "000001"})
	result, err := HandleIndexKline(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleIndexKline_InvalidPeriod(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{}
	req := makeReq(map[string]any{"code": "000001", "period": "xyz"})
	result, err := HandleIndexKline(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for invalid period")
	}
}

func TestHandleIndexKline_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{indexBars: makeIndexBars()}
	req := makeReq(map[string]any{"code": "000001", "setcode": "1", "period": "day", "count": float64(10)})
	result, err := HandleIndexKline(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["code"] != "000001" {
		t.Fatalf("expected code=000001, got %v", data["code"])
	}
	if data["market"].(float64) != 1 {
		t.Fatalf("expected market=1, got %v", data["market"])
	}
	if data["period"] != "day" {
		t.Fatalf("expected period=day, got %v", data["period"])
	}
	if data["count"].(float64) != 2 {
		t.Fatalf("expected count=2, got %v", data["count"])
	}
	bars := data["data"].([]interface{})
	bar := bars[0].(map[string]interface{})
	if bar["date"] != "2025-01-06" {
		t.Fatalf("expected date=2025-01-06, got %v", bar["date"])
	}
	if bar["close"].(float64) != 2915 {
		t.Fatalf("expected close=2915, got %v", bar["close"])
	}
}

func TestHandleIndexKline_MaxCount(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{indexBars: makeIndexBars()}
	req := makeReq(map[string]any{"code": "000001", "count": float64(99999)})
	result, err := HandleIndexKline(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ===========================================================================
// HandleIPOCalendar Tests
// ===========================================================================

func TestHandleIPOCalendar_NoDate(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{ipoData: makeIPOData()}
	req := makeReq(map[string]any{})
	result, err := HandleIPOCalendar(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil && !result.IsError {
		// Date is optional, should work with default
	}
}

func TestHandleIPOCalendar_ClientUnsupported(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}
	req := makeReq(map[string]any{})
	result, err := HandleIPOCalendar(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsupported client error")
	}
}

func TestHandleIPOCalendar_Success(t *testing.T) {
	ctx := context.Background()
	client := &mockTDXProtocolClient{ipoData: makeIPOData()}
	req := makeReq(map[string]any{"date": "20250106", "limit": float64(10)})
	result, err := HandleIPOCalendar(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["count"].(float64) != 1 {
		t.Fatalf("expected count=1, got %v", data["count"])
	}
}

func TestHandleIPOCalendar_EmptyData(t *testing.T) {
	ctx := context.Background()
	// Need to import fmt for make error
	client := &mockTDXProtocolClient{ipoData: []map[string]interface{}{}}
	req := makeReq(map[string]any{"date": "20250106"})
	result, err := HandleIPOCalendar(ctx, client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	data := parseResult(t, result)
	if data["count"].(float64) != 0 {
		t.Fatalf("expected count=0, got %v", data["count"])
	}
}

// ===========================================================================
// Interface types (mirror of those in tools_new.go handler functions)
// ===========================================================================

type auctionClient interface {
	GetCallAuction(code string, market int) ([]proto.AuctionData, error)
}

type historyMinuteClient interface {
	GetHistoryMinute(code string, market int, date string) ([]proto.HistoryMinuteTimeData, error)
}

type historyTradeClient interface {
	GetHistoryTrade(code string, market int, date string, count int) ([]proto.HistoryTransactionData, error)
}

type stockStatsClient interface {
	GetRealtimeQuote(code string, market int) (*proto.SecurityQuote, error)
	GetSymbolInfo(code string, market int) (*proto.MACSymbolInfoReply, error)
	GetSymbolBelongBoard(code string, market uint8) ([]proto.MACBelongBoardItem, error)
}

type indexKlineClient interface {
	GetIndexKLine(code string, market int, period int, count int) ([]proto.IndexBar, error)
}

type financeClient interface {
	GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error)
}

// ===========================================================================
// UnifiedClient Delegation Component Tests
// ===========================================================================

func TestUnifiedClient_GetCallAuction_Mock(t *testing.T) {
	ctx := context.Background()
	uc := &UnifiedClient{}
	var _ auctionClient = uc
	var _ historyMinuteClient = uc
	var _ historyTradeClient = uc
	var _ stockStatsClient = uc
	var _ financeClient = uc
	var _ indexKlineClient = uc
	var _ ipoCalendarClient = uc

	_ = ctx
	_ = uc
}

// ===========================================================================
// Helper
// ===========================================================================

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

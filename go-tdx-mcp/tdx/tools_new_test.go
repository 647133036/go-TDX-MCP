package tdx

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// --- MockClient for handler unit tests ---

type mockClient struct {
	tqlexResp *TQLEXResponse
	tqlexErr  error
	ragResp   *RAGResponse
	ragErr    error
}

func (m *mockClient) TQLEXQuery(ctx context.Context, api string, params interface{}) (*TQLEXResponse, error) {
	if m.tqlexResp != nil {
		return m.tqlexResp, nil
	}
	return m.tqlexResp, m.tqlexErr
}
func (m *mockClient) RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error) {
	if m.ragResp != nil {
		return m.ragResp, nil
	}
	return m.ragResp, m.ragErr
}

// --- TestGetNewHandler ---

func TestGetNewHandler_FactorList(t *testing.T) {
	h := GetNewHandler(ToolFactorList)
	if h == nil {
		t.Fatal("expected handler for tdx_factor_list")
	}
}

func TestGetNewHandler_FactorCompute(t *testing.T) {
	h := GetNewHandler(ToolFactorCompute)
	if h == nil {
		t.Fatal("expected handler for tdx_factor_compute")
	}
}

func TestGetNewHandler_BacktestRun(t *testing.T) {
	h := GetNewHandler(ToolBacktestRun)
	if h == nil {
		t.Fatal("expected handler for tdx_backtest_run")
	}
}

func TestGetNewHandler_EastMoneyRealtimeQuote(t *testing.T) {
	h := GetNewHandler(ToolEastMoneyRealtimeQuote)
	if h == nil {
		t.Fatal("expected handler for tdx_eastmoney_realtime_quote")
	}
}

func TestGetNewHandler_TableParserURL(t *testing.T) {
	h := GetNewHandler(ToolTableParserURL)
	if h == nil {
		t.Fatal("expected handler for tdx_table_parser_url")
	}
}

func TestGetNewHandler_TECryptoData(t *testing.T) {
	h := GetNewHandler(ToolTECryptoData)
	if h == nil {
		t.Fatal("expected handler for tdx_tecrypto_data")
	}
}

func TestGetNewHandler_FundNAV(t *testing.T) {
	h := GetNewHandler(ToolFundNAV)
	if h == nil {
		t.Fatal("expected handler for tdx_fund_nav")
	}
}

func TestGetNewHandler_DragonTiger(t *testing.T) {
	h := GetNewHandler(ToolDragonTiger)
	if h == nil {
		t.Fatal("expected handler for tdx_dragon_tiger")
	}
}

func TestGetNewHandler_CurrentTimestamp(t *testing.T) {
	h := GetNewHandler(ToolCurrentTimestamp)
	if h == nil {
		t.Fatal("expected handler for tdx_current_timestamp")
	}
}

func TestGetNewHandler_Unknown(t *testing.T) {
	h := GetNewHandler("unknown_tool_xyz")
	if h != nil {
		t.Fatal("expected nil for unknown tool")
	}
}

// --- TestHandleFactorList ---

func TestHandleFactorList_AllCategories(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleFactorList(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Fatalf("got error result: %s", getTextContent(result))
	}
	text := getTextContent(result)
	if len(text) == 0 {
		t.Fatal("expected non-empty text content")
	}
}

func TestHandleFactorList_FilterCategory(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"category": "momentum"},
		},
	}
	result, err := HandleFactorList(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("got error result: %s", getTextContent(result))
	}
}

func TestHandleFactorList_InvalidCategory(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"category": "nonexistent"},
		},
	}
	result, err := HandleFactorList(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("got error for invalid category: %s", getTextContent(result))
	}
}

// --- TestHandleCurrentTimestamp ---

func TestHandleCurrentTimestamp(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleCurrentTimestamp(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("got error result: %s", getTextContent(result))
	}
	text := getTextContent(result)
	if len(text) == 0 {
		t.Fatal("expected non-empty timestamp")
	}
}

// --- TestHandleTECryptoData ---

func TestHandleTECryptoData_ValidSymbols(t *testing.T) {
	t.Skip("external crypto API may be unreachable in CI")
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbols_crypto": "BTC,ETH"},
			},
		}
		result, err := HandleTECryptoData(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("TECryptoData warning: %v", err)
			return
		}
		_ = result
	}, 30*time.Second)
}

func TestHandleTECryptoData_EmptySymbols(t *testing.T) {
	t.Skip("external crypto API may be unreachable in CI")
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbols_crypto": ""},
			},
		}
		result, err := HandleTECryptoData(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("TECryptoData warning: %v", err)
			return
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 30*time.Second)
}

// --- TestHandleFundNAV ---

func TestHandleFundNAV_ValidCode(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"fund_code": "110011"},
			},
		}
		result, err := HandleFundNavLatest(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("FundNAV warning: %v", err)
			return
		}
		_ = result
	}, 10*time.Second)
}

func TestHandleFundNAV_InvalidCode(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"fund_code": "invalid"},
			},
		}
		result, err := HandleFundNavLatest(ctx, &mockClient{}, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 10*time.Second)
}

// --- TestHandleDragonTiger ---

func TestHandleDragonTiger_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleDragonTiger(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestHandleDragonTiger_CustomLimit(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"limit": float64(5)},
		},
	}
	result, err := HandleDragonTiger(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleConvertibleBond ---

func TestHandleConvertibleBond(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleConvertibleBond(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleFuturesQuote ---

func TestHandleFuturesQuote_ValidSymbols(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbols_crypto": "au2506,cu2506"},
			},
		}
		result, err := HandleFuturesQuote(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("FuturesQuote warning: %v", err)
			return
		}
		_ = result
	}, 10*time.Second)
}

func TestHandleFuturesQuote_EmptySymbols(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbols_crypto": ""},
			},
		}
		result, err := HandleFuturesQuote(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("FuturesQuote warning: %v", err)
			return
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 10*time.Second)
}

// --- TestHandleStockCodeResolve ---

func TestHandleStockCodeResolve_ValidCodes(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"codes": "000001,600000"},
		},
	}
	result, err := HandleStockCodeResolve(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestHandleStockCodeResolve_InvalidCodes(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"codes": "invalid"},
		},
	}
	result, err := HandleStockCodeResolve(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleCSIIndexConstituents ---

func TestHandleCSIIndexConstituents_HSZ300(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"index_code": "000300"},
		},
	}
	result, err := HandleCSIIndexConstituents(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestHandleCSIIndexConstituents_HSZ50(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"index_code": "000016"},
		},
	}
	result, err := HandleCSIIndexConstituents(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- Test Tool Definitions ---

func TestNewFactorListTool_Definition(t *testing.T) {
	tool := NewFactorListTool()
	if tool.Name != ToolFactorList {
		t.Errorf("expected tool name %s, got %s", ToolFactorList, tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestNewTECryptoDataTool_Definition(t *testing.T) {
	tool := NewTECryptoDataTool()
	if tool.Name != ToolTECryptoData {
		t.Errorf("expected tool name %s, got %s", ToolTECryptoData, tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestNewTECryptoKlineTool_Definition(t *testing.T) {
	tool := NewTECryptoKlineTool()
	if tool.Name != ToolTECryptoKline {
		t.Errorf("expected tool name %s, got %s", ToolTECryptoKline, tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestNewFundNAVTool_Definition(t *testing.T) {
	tool := NewFundNAVTool()
	if tool.Name != ToolFundNAV {
		t.Errorf("expected tool name %s, got %s", ToolFundNAV, tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestNewMarginTradeTool_Definition(t *testing.T) {
	tool := NewMarginTradeTool()
	if tool.Name != ToolMarginTrade {
		t.Errorf("expected tool name %s, got %s", ToolMarginTrade, tool.Name)
	}
}

func TestNewDragonTigerTool_Definition(t *testing.T) {
	tool := NewDragonTigerTool()
	if tool.Name != ToolDragonTiger {
		t.Errorf("expected tool name %s, got %s", ToolDragonTiger, tool.Name)
	}
}

func TestNewConvertibleBondTool_Definition(t *testing.T) {
	tool := NewConvertibleBondTool()
	if tool.Name != ToolConvertibleBond {
		t.Errorf("expected tool name %s, got %s", ToolConvertibleBond, tool.Name)
	}
}

func TestNewFuturesQuoteTool_Definition(t *testing.T) {
	tool := NewFuturesQuoteTool()
	if tool.Name != ToolFuturesQuote {
		t.Errorf("expected tool name %s, got %s", ToolFuturesQuote, tool.Name)
	}
}

func TestNewStockCodeResolveTool_Definition(t *testing.T) {
	tool := NewStockCodeResolveTool()
	if tool.Name != ToolStockCodeResolve {
		t.Errorf("expected tool name %s, got %s", ToolStockCodeResolve, tool.Name)
	}
}

func TestNewCSIIndexConstituentsTool_Definition(t *testing.T) {
	tool := NewCSIIndexConstituentsTool()
	if tool.Name != ToolCSIIndexConstituents {
		t.Errorf("expected tool name %s, got %s", ToolCSIIndexConstituents, tool.Name)
	}
}

// --- TestHandleFactorCompute ---

func TestHandleFactorCompute_MissingCode(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": float64(1), "factors": "MA"},
		},
	}
	result, err := HandleFactorCompute(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code param")
	}
}

func TestHandleFactorCompute_MissingMarket(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "factors": "MA"},
		},
	}
	result, err := HandleFactorCompute(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing market param")
	}
}

func TestHandleFactorCompute_MissingFactors(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleFactorCompute(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing factors param")
	}
}

// --- TestHandleFactorAnalyze ---

func TestHandleFactorAnalyze_MissingCode(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": float64(1), "factor_name": "MA"},
		},
	}
	result, err := HandleFactorAnalyze(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code param")
	}
}

func TestHandleFactorAnalyze_MissingFactorName(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleFactorAnalyze(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing factor_name param")
	}
}

// --- TestHandleScreenScan ---

func TestHandleScreenScan_MissingCode(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": float64(1)},
		},
	}
	result, err := HandleScreenScan(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code param")
	}
}

// --- TestHandleScreenStrength ---

func TestHandleScreenStrength_MissingCode(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": float64(1)},
		},
	}
	result, err := HandleScreenStrength(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing code param")
	}
}

// --- TestHandleEnhancedBacktest ---

func TestHandleEnhancedBacktest_EmptyParams(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleEnhancedBacktest(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleTECryptoKline ---

func TestHandleTECryptoKline_ValidSymbol(t *testing.T) {
	t.Skip("external crypto API may be unreachable in CI")
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbol_crypto": "BTC", "interval": "1d", "limit": float64(10)},
			},
		}
		result, err := HandleTECryptoKline(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("TECryptoKline warning: %v", err)
			return
		}
		_ = result
	}, 30*time.Second)
}

func TestHandleTECryptoKline_DefaultInterval(t *testing.T) {
	t.Skip("external crypto API may be unreachable in CI")
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"symbol_crypto": "ETH"},
			},
		}
		result, err := HandleTECryptoKline(ctx, &mockClient{}, req)
		if err != nil {
			t.Logf("TECryptoKline warning: %v", err)
			return
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 30*time.Second)
}

// --- TestHandleBacktestAvailable ---

func TestHandleBacktestAvailable(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleBacktestAvailable(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// May return error if no client is connected - that's acceptable
	_ = result
}

// --- TestHandleBacktestRun ---

func TestHandleBacktestRun_EmptyParams(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleBacktestRun(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleBacktestCombo ---

func TestHandleBacktestCombo_EmptyParams(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleBacktestCombo(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleFactorGetInfo ---

func TestHandleFactorGetInfo_NoFilter(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"name": "MA"},
		},
	}
	result, err := HandleFactorGetInfo(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestHandleFactorGetInfo_WithCategory(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"name": "MA", "category": "momentum"},
		},
	}
	result, err := HandleFactorGetInfo(ctx, &mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestAllNewToolDefinitionsExist ---

func TestAllNewToolDefinitionsExist(t *testing.T) {
	defs := []struct {
		name string
		fn   func() mcp.Tool
	}{
		{ToolFactorList, NewFactorListTool},
		{ToolTECryptoData, NewTECryptoDataTool},
		{ToolTECryptoKline, NewTECryptoKlineTool},
		{ToolFundNAV, NewFundNAVTool},
		{ToolMarginTrade, NewMarginTradeTool},
		{ToolDragonTiger, NewDragonTigerTool},
		{ToolConvertibleBond, NewConvertibleBondTool},
		{ToolFuturesQuote, NewFuturesQuoteTool},
		{ToolStockCodeResolve, NewStockCodeResolveTool},
		{ToolCSIIndexConstituents, NewCSIIndexConstituentsTool},
	}

	for _, d := range defs {
		t.Run(d.name, func(t *testing.T) {
			tool := d.fn()
			if tool.Name != d.name {
				t.Errorf("tool name mismatch: expected %s, got %s", d.name, tool.Name)
			}
			if tool.Description == "" {
				t.Errorf("tool %s has empty description", d.name)
			}
		})
	}
}

// --- Helper ---

func getTextContent(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

package tdx

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ===========================================================================
// E2E Real Server Tests — 7 new protocol handlers via UnifiedClient
// ===========================================================================

func TestIntegrationE2ERealServer(t *testing.T) {
	ctx := context.Background()
	client := NewUnifiedClient("", 0, "218.75.122.92", 7709)
	defer client.Close()

	dialCtx, dialCancel := context.WithTimeout(ctx, 30*time.Second)
	err := client.initTCP(dialCtx)
	dialCancel()
	if err != nil {
		t.Skipf("cannot connect to TDX server (skipping e2e tests): %v", err)
		return
	}
	if !client.tcpClient.IsConnected() {
		t.Skipf("TCP connection not established, skipping e2e tests")
		return
	}
	t.Log("=== Connected to real TDX server ===")

	// 1. Call Auction
	t.Run("CallAuction", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "0"})
		result, err := HandleCallAuction(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleCallAuction error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Logf("auction unavailable (may be off-hours): %s",
				string(result.Content[0].(mcp.TextContent).Text))
			return
		}
		data := parseResultJSON(t, result)
		t.Logf("auction OK, count=%v, code=%v", data["count"], data["code"])
	})

	// 2. History Minute
	t.Run("HistoryMinute", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106"})
		result, err := HandleHistoryMinute(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleHistoryMinute error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Logf("history minute unavailable: %s",
				string(result.Content[0].(mcp.TextContent).Text))
			return
		}
		data := parseResultJSON(t, result)
		if data["count"].(float64) == 0 {
			t.Logf("history minute returned 0 rows (server may not support it)")
			return
		}
		t.Logf("history minute OK, count=%v", data["count"])
	})

	// 3. History Trade
	t.Run("HistoryTrade", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "0", "date": "20250106"})
		result, err := HandleHistoryTrade(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleHistoryTrade error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Logf("history trade unavailable: %s",
				string(result.Content[0].(mcp.TextContent).Text))
			return
		}
		data := parseResultJSON(t, result)
		t.Logf("history trade OK, count=%v", data["count"])
	})

	// 4. Stock Stats
	t.Run("StockStats", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "0"})
		// UnifiedClient already satisfies stockStatsClient via its GetRealtimeQuote + GetSymbolInfo + GetFinanceInfo
		result, err := HandleStockStats(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleStockStats error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Fatalf("HandleStockStats error result: %s",
				string(result.Content[0].(mcp.TextContent).Text))
		}
		data := parseResultJSON(t, result)
		if data["code"] != "000001" {
			t.Fatalf("expected code=000001, got %v", data["code"])
		}
		if data["eps"] == nil {
			t.Fatalf("expected eps field in result")
		}
		t.Logf("stock stats OK, code=%v, eps=%v, current_price=%v",
			data["code"], data["eps"], data["current_price"])
	})

	// 5. Finance Info
	t.Run("FinanceInfo", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "0"})
		result, err := HandleFinanceInfo(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleFinanceInfo error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Fatalf("HandleFinanceInfo error result: %s",
				string(result.Content[0].(mcp.TextContent).Text))
		}
		data := parseResultJSON(t, result)
		if data["code"] != "000001" {
			t.Fatalf("expected code=000001, got %v", data["code"])
		}
		if data["eps"] == nil {
			t.Fatalf("expected eps in result, got %v", data)
		}
		if data["total_shares"] == nil {
			t.Fatalf("expected total_shares in result")
		}
		t.Logf("finance info OK, eps=%v, total_shares=%v, net_profit=%v",
			data["eps"], data["total_shares"], data["net_profit"])
	})

	// 6. Index Kline
	t.Run("IndexKline", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"code": "000001", "setcode": "1", "period": "day", "count": float64(5)})
		result, err := HandleIndexKline(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleIndexKline error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Fatalf("HandleIndexKline error result: %s",
				string(result.Content[0].(mcp.TextContent).Text))
		}
		data := parseResultJSON(t, result)
		if data["count"].(float64) == 0 {
			t.Fatalf("expected kline data, got count=0")
		}
		if data["code"] != "000001" {
			t.Fatalf("expected code=000001, got %v", data["code"])
		}
		bars := data["data"].([]interface{})
		firstBar := bars[0].(map[string]interface{})
		if firstBar["close"] == nil {
			t.Fatalf("expected close price in bar")
		}
		t.Logf("index kline OK, count=%v, first_close=%v, period=%v",
			data["count"], firstBar["close"], data["period"])
	})

	// 7. IPO Calendar
	t.Run("IPOCalendar", func(t *testing.T) {
		req := makeE2EReq(map[string]any{"date": "20250106"})
		result, err := HandleIPOCalendar(ctx, client, req)
		if err != nil {
			t.Fatalf("HandleIPOCalendar error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Logf("IPO calendar not available: %s",
				string(result.Content[0].(mcp.TextContent).Text))
			return
		}
		data := parseResultJSON(t, result)
		t.Logf("IPO calendar OK, count=%v", data["count"])
	})

	t.Log("=== All 7 new protocol handler E2E tests complete ===")
}

// ===========================================================================
// Helpers (must be unique names to avoid conflict with tdx_protocol_tools_test.go)
// ===========================================================================

func makeE2EReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func parseResultJSON(t *testing.T, result *mcp.CallToolResult) map[string]interface{} {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	var raw map[string]interface{}
	content := string(result.Content[0].(mcp.TextContent).Text)
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		t.Fatalf("failed to parse result JSON: %s\nraw: %s", err, content)
	}
	return raw
}

// ===========================================================================
// Registration + definition verification
// ===========================================================================

func TestGetNewHandler_AllSevenRegistered(t *testing.T) {
	tests := []struct {
		name string
		tool string
	}{
		{"CallAuction", ToolCallAuction},
		{"HistoryMinute", ToolHistoryMinute},
		{"HistoryTrade", ToolHistoryTrade},
		{"StockStats", ToolStockStats},
		{"FinanceInfo", ToolFinanceInfo},
		{"IndexKline", ToolIndexKline},
		{"IPOCalendar", ToolIPOCalendar},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := GetNewHandler(tc.tool)
			if h == nil {
				t.Fatalf("GetNewHandler(%s) returned nil", tc.tool)
			}
		})
	}
}

func TestAllSevenToolDefinitionsInGetAllNewTools(t *testing.T) {
	tools := GetAllNewTools()
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}
	required := []string{
		ToolCallAuction, ToolHistoryMinute, ToolHistoryTrade,
		ToolStockStats, ToolFinanceInfo, ToolIndexKline,
		ToolIPOCalendar,
	}
	for _, name := range required {
		if !toolNames[name] {
			t.Errorf("GetAllNewTools missing tool: %s", name)
		}
	}
}

// ===========================================================================
// Interface compile-time checks
// ===========================================================================

func TestUnifiedClient_CompileTimeInterfaceCheck(t *testing.T) {
	uc := &UnifiedClient{}
	var _ auctionClient = uc
	var _ historyMinuteClient = uc
	var _ historyTradeClient = uc
	var _ stockStatsClient = uc
	var _ financeClient = uc
	var _ indexKlineClient = uc
	var _ ipoCalendarClient = uc
	_ = uc
}

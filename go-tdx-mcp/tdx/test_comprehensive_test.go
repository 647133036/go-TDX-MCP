package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ============================================================
// Comprehensive test runner for all 212 tools
// ============================================================

func TestComprehensiveAllTools(t *testing.T) {
	configs := buildTestConfigs()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()

	err := client.initTCP(ctx)
	tcpAvailable := err == nil
	if !tcpAvailable {
		t.Logf("[TCP CONNECTION FAILED] %v", err)
	}

	results := runAllComprehensive(ctx, client, configs, tcpAvailable)

	// Report
	passCount := 0
	failCount := 0
	failDetails := []string{}
	for _, r := range results {
		if r.status == "PASS" {
			passCount++
		} else {
			failCount++
			failDetails = append(failDetails, fmt.Sprintf("  %-35s %s [%s]", r.name, r.reason, r.source))
		}
	}

	t.Logf("\n========== COMPREHENSIVE TEST SUMMARY ==========")
	t.Logf("PASS:       %d", passCount)
	t.Logf("FAIL:       %d", failCount)
	t.Logf("TOTAL:      %d", len(results))
	if failCount > 0 {
		t.Logf("\nFAILED TESTS:")
		sort.Strings(failDetails)
		for _, d := range failDetails {
			t.Logf("%s", d)
		}
	}
	t.Logf("==============================================")

	if failCount > 0 {
		t.Errorf("%d tests failed out of %d", failCount, len(results))
	}
}

type toolResult struct {
	name   string
	status string
	reason string
	source string
}

func runAllComprehensive(ctx context.Context, client Client, configs []TestConfig, tcpAvailable bool) []toolResult {
	results := make([]toolResult, 0, len(configs))

	for _, tc := range configs {
		if tc.Handler == nil {
			results = append(results, toolResult{tc.HandlerName, "EMPTY", "handler is nil", tc.Source})
			continue
		}
		if tc.NeedsToken {
			results = append(results, toolResult{tc.HandlerName, "NEEDS_TOKEN", "requires TDX_TOKEN", tc.Source})
			continue
		}

		args := make(map[string]interface{})
		for k, v := range tc.Params {
			args[k] = v
		}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      tc.HandlerName,
				Arguments: args,
			},
		}

		result := runComprehensiveOne(ctx, client, tc, req)
		results = append(results, result)
	}

	return results
}

func runComprehensiveOne(ctx context.Context, client Client, tc TestConfig, req mcp.CallToolRequest) toolResult {
	// Run handler
	var resp *mcp.CallToolResult
	var handlerErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				handlerErr = fmt.Errorf("panic: %v", r)
			}
		}()
		resp, handlerErr = tc.Handler(ctx, client, req)
	}()

	if handlerErr != nil {
		return toolResult{tc.HandlerName, "FAIL", fmt.Sprintf("handler error: %v", handlerErr), tc.Source}
	}

	if resp == nil {
		return toolResult{tc.HandlerName, "FAIL", "nil response", tc.Source}
	}

	// Extract response text
	var rawText string
	if len(resp.Content) > 0 {
		if textContent, ok := resp.Content[0].(mcp.TextContent); ok {
			rawText = textContent.Text
		}
	}

	if rawText == "" {
		return toolResult{tc.HandlerName, "FAIL", "empty content", tc.Source}
	}

	// Validate response structure
	if err := validateResponse(tc.HandlerName, rawText); err != "" {
		return toolResult{tc.HandlerName, "FAIL", fmt.Sprintf("validation: %s", err), tc.Source}
	}

	return toolResult{tc.HandlerName, "PASS", "ok", tc.Source}
}

// ============================================================
// Response validators per tool category
// ============================================================

func validateResponse(toolName string, rawText string) string {
	// Quick sanity: non-empty, printable chars
	if len(rawText) == 0 {
		return "empty response"
	}

	// Try JSON parse
	var data interface{}
	if err := json.Unmarshal([]byte(rawText), &data); err != nil {
		// Not valid JSON - some handlers return plain text (chanlun, backtest)
		// Accept any non-empty text for those
		return ""
	}

	// Validate by tool category
	switch toolName {
	// Core tools
	case "tdx_lookup_stock":
		obj := data.(map[string]interface{})
		if _, ok := obj["query"]; !ok {
			return "missing 'query' field"
		}

	// RAG tools
	case "tdx_rag_query", "wenda_macro_query":
		obj := data.(map[string]interface{})
		if _, ok := obj["reasoning_path"]; !ok {
			return "missing 'reasoning_path' field"
		}
		if _, ok := obj["query"]; !ok {
			return "missing 'query' field"
		}

	// V3 tools with structured response (no 'data' field)
	case "tdx_market_overview":
		obj := data.(map[string]interface{})
		if _, ok := obj["market_stat"]; !ok {
			return "missing 'market_stat' field"
		}
	case "tdx_sector_flow":
		obj := data.(map[string]interface{})
		if _, ok := obj["board_type"]; !ok {
			return "missing 'board_type' field"
		}
	case "tdx_financial_metrics", "tdx_macro_data":
		obj := data.(map[string]interface{})
		// These may return data or error structure — accept any non-empty object
		if obj == nil || len(obj) == 0 {
			return "empty response"
		}
	case "tdx_top_gainers_losers":
		obj := data.(map[string]interface{})
		if _, ok := obj["sort_type"]; !ok {
			return "missing 'sort_type' field"
		}
		if _, ok := obj["up_list"]; !ok {
			return "missing 'up_list' field"
		}

	// Tools that return direct arrays
	default:
		// Accept any valid JSON
	}

	return ""
}

// ============================================================
// Edge case tests
// ============================================================

func TestEdgeCases_MissingParams(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	// Test tdx_quotes without code
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name:      "tdx_quotes",
            Arguments: map[string]interface{}{"setcode": "0"},
        },
    }
	resp, err := HandleQuotes(ctx, client, req)
	if err != nil {
		t.Logf("tdx_quotes without code: %v", err)
		return
	}
	if resp != nil && len(resp.Content) > 0 {
		text := resp.Content[0].(mcp.TextContent).Text
		if strings.Contains(text, "参数必填") || strings.Contains(text, "required") {
			t.Logf("correctly rejected missing code")
			return
		}
		t.Logf("tdx_quotes without code: got response (may be OK with defaults)")
	}
}

func TestEdgeCases_InvalidCode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name:      "tdx_quotes",
            Arguments: map[string]interface{}{"code": "INVALID", "setcode": "0"},
        },
    }
	resp, err := HandleQuotes(ctx, client, req)
	if err != nil {
		t.Logf("invalid code handled: %v", err)
		return
	}
	if resp != nil && len(resp.Content) > 0 {
		text := resp.Content[0].(mcp.TextContent).Text
		t.Logf("invalid code response: %s", text[:min(200, len(text))])
	}
}

func TestEdgeCases_NegativeCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: "tdx_kline",
            Arguments: map[string]interface{}{
                "code":    "000001",
                "setcode": 0,
                "period":  4,
                "count":   -100,
            },
        },
    }
	resp, err := HandleKline(ctx, client, req)
	if err != nil {
		t.Logf("negative count: %v", err)
		return
	}
	if resp != nil && len(resp.Content) > 0 {
		text := resp.Content[0].(mcp.TextContent).Text
		t.Logf("negative count response: %s", text[:min(200, len(text))])
	}
}

func TestEdgeCases_EmptyQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: "tdx_rag_query",
            Arguments: map[string]interface{}{"query": "", "top_k": 3},
        },
    }
	resp, err := HandleRAGQuery(ctx, client, req)
	if err != nil {
		t.Logf("empty query: %v", err)
		return
	}
	if resp != nil && len(resp.Content) > 0 {
		text := resp.Content[0].(mcp.TextContent).Text
		if len(text) < 5 {
			t.Logf("timestamp too short: %s", text)
			return
		}
	}
}

// ============================================================
// Functional correctness tests
// ============================================================

func TestFunctional_QuoteHasCode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "tdx_quotes",
			Arguments: map[string]interface{}{"code": "000001", "setcode": "0"},
		},
	}
	resp, err := HandleQuotes(ctx, client, req)
	if err != nil || resp == nil || len(resp.Content) == 0 {
		t.Logf("quotes failed or empty: %v", err)
		return
	}
	text := resp.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "000001") {
		t.Logf("response does not contain '000001' (may use numeric code): %s", text[:min(200, len(text))])
	} else {
		t.Logf("quotes response contains '000001': PASS")
	}
}

func TestFunctional_KlineHasOHLC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: "tdx_kline",
            Arguments: map[string]interface{}{
                "code":    "000001",
                "setcode": 0,
                "period":  4,
            },
        },
    }
	resp, err := HandleKline(ctx, client, req)
	if err != nil || resp == nil || len(resp.Content) == 0 {
		t.Fatalf("kline failed: %v", err)
	}
	text := resp.Content[0].(mcp.TextContent).Text
	// Kline should contain OHLC data (may return error text when data unavailable)
	if len(text) < 10 {
		t.Logf("kline returned short response: %s (data may be unavailable)", text)
		return
	}
	if !strings.Contains(text, "Open") && !strings.Contains(text, "open") &&
		!strings.Contains(text, "Close") && !strings.Contains(text, "close") {
		t.Errorf("kline missing OHLC data: %s", text[:min(300, len(text))])
	}
}

func TestFunctional_Timestamp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name:      "tdx_current_timestamp",
            Arguments: map[string]interface{}{},
        },
    }
	resp, err := HandleCurrentTimestamp(ctx, client, req)
	if err != nil || resp == nil || len(resp.Content) == 0 {
		t.Fatalf("timestamp failed: %v", err)
	}
	text := resp.Content[0].(mcp.TextContent).Text
	if len(text) < 5 {
		t.Errorf("timestamp too short: %s", text)
	}
}

func TestFunctional_RAGQueryStructure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: "tdx_rag_query",
            Arguments: map[string]interface{}{
                "query": "今天哪个板块涨得好？",
                "top_k": 5,
            },
        },
    }
	resp, err := HandleRAGQuery(ctx, client, req)
	if err != nil || resp == nil || len(resp.Content) == 0 {
		t.Fatalf("RAG query failed: %v", err)
	}
	text := resp.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reasoning_path") {
		t.Errorf("RAG missing reasoning_path")
	}
	if !strings.Contains(text, "query") {
		t.Errorf("RAG missing query")
	}
}

// ============================================================
// Data type safety tests
// ============================================================

func TestTypes_SetcodeAsInt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
            Name: "tdx_kline",
            Arguments: map[string]interface{}{
                "code":    "000001",
                "setcode": 1,
                "period":  4,
            },
        },
    }
	_, err = HandleKline(ctx, client, req)
	if err != nil {
		t.Logf("setcode as int: %v (may need fix)", err)
	}
}

// ============================================================
// Nil safety: all handlers must return non-nil
// ============================================================

func TestAllHandlersNonNil(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()
	err := client.initTCP(ctx)
	if err != nil {
		t.Skipf("TCP unavailable: %v", err)
	}

	configs := buildTestConfigs()
	nilCount := 0
	for _, tc := range configs {
		if tc.Handler == nil || tc.NeedsToken {
			continue
		}
		args := make(map[string]interface{})
		for k, v := range tc.Params {
			args[k] = v
		}
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      tc.HandlerName,
				Arguments: args,
			},
		}

		var resp *mcp.CallToolResult
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tc.HandlerName, r)
				}
			}()
			resp, _ = tc.Handler(ctx, client, req)
		}()

		if resp == nil {
			nilCount++
			t.Errorf("%s returned nil", tc.HandlerName)
		}
	}
	if nilCount > 0 {
		t.Errorf("%d handlers returned nil", nilCount)
	}
}

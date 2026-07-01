package tdx

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// --- MockClient for handler unit tests ---

type mockClient2 struct {
	tqlexResp *TQLEXResponse
	tqlexErr  error
	ragResp   *RAGResponse
	ragErr    error
}

func (m *mockClient2) TQLEXQuery(ctx context.Context, api string, params interface{}) (*TQLEXResponse, error) {
	if m.tqlexResp != nil {
		return m.tqlexResp, m.tqlexErr
	}
	if m.tqlexErr != nil {
		return nil, m.tqlexErr
	}
	return &TQLEXResponse{Data: map[string]any{"status": "ok"}}, nil
}
func (m *mockClient2) RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error) {
	if m.ragResp != nil {
		return m.ragResp, m.ragErr
	}
	if m.ragErr != nil {
		return nil, m.ragErr
	}
	return &RAGResponse{}, nil
}

// --- TestGetExpandedHandler ---

func TestGetExpandedHandler_Tick(t *testing.T) {
	h := GetExpandedHandler(ToolTick)
	if h == nil {
		t.Fatal("expected handler for tdx_tick")
	}
}

func TestGetExpandedHandler_Transaction(t *testing.T) {
	h := GetExpandedHandler(ToolTransaction)
	if h == nil {
		t.Fatal("expected handler for tdx_transaction")
	}
}

func TestGetExpandedHandler_BoardList(t *testing.T) {
	h := GetExpandedHandler(ToolBoardList)
	if h == nil {
		t.Fatal("expected handler for tdx_board_list")
	}
}

func TestGetExpandedHandler_ServerInfo(t *testing.T) {
	h := GetExpandedHandler(ToolServerInfo)
	if h == nil {
		t.Fatal("expected handler for tdx_server_info")
	}
}

func TestGetExpandedHandler_SymbolInfo(t *testing.T) {
	h := GetExpandedHandler(ToolSymbolInfo)
	if h == nil {
		t.Fatal("expected handler for tdx_symbol_info")
	}
}

func TestGetExpandedHandler_Financial(t *testing.T) {
	h := GetExpandedHandler(ToolFinancial)
	if h == nil {
		t.Fatal("expected handler for tdx_financial")
	}
}

func TestGetExpandedHandler_IndicatorCompute(t *testing.T) {
	h := GetExpandedHandler(ToolIndicatorComp)
	if h == nil {
		t.Fatal("expected handler for tdx_indicator_compute")
	}
}

func TestGetExpandedHandler_Chanlun(t *testing.T) {
	h := GetExpandedHandler(ToolChanlun)
	if h == nil {
		t.Fatal("expected handler for tdx_chanlun_analyze")
	}
}

func TestGetExpandedHandler_Backtest(t *testing.T) {
	h := GetExpandedHandler(ToolBacktest)
	if h == nil {
		t.Fatal("expected handler for tdx_backtest")
	}
}

func TestGetExpandedHandler_OfflineHome(t *testing.T) {
	h := GetExpandedHandler(ToolOfflineHome)
	if h == nil {
		t.Fatal("expected handler for tdx_offline_home")
	}
}

func TestGetExpandedHandler_OfflineDaily(t *testing.T) {
	h := GetExpandedHandler(ToolOfflineDaily)
	if h == nil {
		t.Fatal("expected handler for tdx_offline_daily")
	}
}

func TestGetExpandedHandler_OfflineSyncAll(t *testing.T) {
	h := GetExpandedHandler(ToolOfflineSyncAll)
	if h == nil {
		t.Fatal("expected handler for tdx_offline_sync_all")
	}
}

func TestGetExpandedHandler_QuoteRealtime(t *testing.T) {
	h := GetExpandedHandler(ToolQuoteRealtime)
	if h == nil {
		t.Fatal("expected handler for tdx_quote_realtime")
	}
}

func TestGetExpandedHandler_KlineExtended(t *testing.T) {
	h := GetExpandedHandler(ToolKlineExtended)
	if h == nil {
		t.Fatal("expected handler for tdx_kline_extended")
	}
}

func TestGetExpandedHandler_DailyLineExtended(t *testing.T) {
	h := GetExpandedHandler(ToolDailyLineExtended)
	if h == nil {
		t.Fatal("expected handler for tdx_daily_line_extended")
	}
}

func TestGetExpandedHandler_MACD(t *testing.T) {
	h := GetExpandedHandler(ToolMACD)
	if h == nil {
		t.Fatal("expected handler for tdx_macd_calc")
	}
}

func TestGetExpandedHandler_KDJ(t *testing.T) {
	h := GetExpandedHandler(ToolKDJ)
	if h == nil {
		t.Fatal("expected handler for tdx_kdj_calc")
	}
}

func TestGetExpandedHandler_RSI(t *testing.T) {
	h := GetExpandedHandler(ToolRSI)
	if h == nil {
		t.Fatal("expected handler for tdx_rsi_calc")
	}
}

func TestGetExpandedHandler_BOLL(t *testing.T) {
	h := GetExpandedHandler(ToolBOLL)
	if h == nil {
		t.Fatal("expected handler for tdx_boll_calc")
	}
}

func TestGetExpandedHandler_EMA(t *testing.T) {
	h := GetExpandedHandler(ToolEMA)
	if h == nil {
		t.Fatal("expected handler for tdx_ema_calc")
	}
}

func TestGetExpandedHandler_OBV(t *testing.T) {
	h := GetExpandedHandler(ToolOBV)
	if h == nil {
		t.Fatal("expected handler for tdx_obv_calc")
	}
}

func TestGetExpandedHandler_ADX(t *testing.T) {
	h := GetExpandedHandler(ToolADX)
	if h == nil {
		t.Fatal("expected handler for tdx_adx_calc")
	}
}

func TestGetExpandedHandler_TechnicalIndicator(t *testing.T) {
	h := GetExpandedHandler(ToolTECHNICAL_INDICATOR)
	if h == nil {
		t.Fatal("expected handler for tdx_technical_indicator")
	}
}

func TestGetExpandedHandler_StockProfile(t *testing.T) {
	h := GetExpandedHandler(ToolStockProfile)
	if h == nil {
		t.Fatal("expected handler for tdx_stock_profile")
	}
}

func TestGetExpandedHandler_SectorRanking(t *testing.T) {
	h := GetExpandedHandler(ToolSectorRanking)
	if h == nil {
		t.Fatal("expected handler for tdx_sector_ranking")
	}
}

func TestGetExpandedHandler_TopGainers(t *testing.T) {
	h := GetExpandedHandler(ToolTopGainers)
	if h == nil {
		t.Fatal("expected handler for tdx_top_gainers")
	}
}

func TestGetExpandedHandler_TopLosers(t *testing.T) {
	h := GetExpandedHandler(ToolTopLosers)
	if h == nil {
		t.Fatal("expected handler for tdx_top_losers")
	}
}

func TestGetExpandedHandler_Unknown(t *testing.T) {
	h := GetExpandedHandler("unknown_tool_xyz")
	if h != nil {
		t.Fatal("expected nil for unknown tool")
	}
}

// --- TestHandleServerInfo ---

func TestHandleServerInfo(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleServerInfo(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	text := getTextContent(result)
	if len(text) == 0 {
		t.Fatal("expected non-empty server info")
	}
}

// --- TestHandleSymbolInfo ---

func TestHandleSymbolInfo_ValidCode(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"code": "000001", "market": float64(1)},
			},
		}
		result, err := HandleSymbolInfo(ctx, &mockClient2{}, req)
		if err != nil {
			t.Logf("SymbolInfo warning: %v", err)
			return
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		_ = result
	}, 10*time.Second)
}

func TestHandleSymbolInfo_MissingCode(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"market": float64(1)},
			},
		}
		result, err := HandleSymbolInfo(ctx, &mockClient2{}, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 10*time.Second)
}

func TestHandleSymbolInfo_MissingMarket(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"code": "000001"},
			},
		}
		result, err := HandleSymbolInfo(ctx, &mockClient2{}, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 10*time.Second)
}

func TestHandleQuoteRealtime_EmptyParams(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{},
			},
		}
		result, err := HandleQuoteRealtime(ctx, &mockClient2{}, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	}, 10*time.Second)
}

// --- TestHandleKlineExtended ---

func TestHandleKlineExtended(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1), "period": "day", "count": float64(10)},
		},
	}
	result, err := HandleKlineExtended(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleMACD ---

func TestHandleMACD(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleMACD(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleKDJ ---

func TestHandleKDJ(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleKDJ(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleRSI ---

func TestHandleRSI(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleRSI(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleBOLL ---

func TestHandleBOLL(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleBOLL(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleTechnicalIndicator ---

func TestHandleTechnicalIndicator(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1), "indicator": "MACD"},
		},
	}
	result, err := HandleTechnicalIndicator(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleStockProfile ---

func TestHandleStockProfile(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleStockProfile(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleSectorRanking ---

func TestHandleSectorRanking(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"sector_type": "concept"},
			},
		}
		result, err := HandleSectorRanking(ctx, &mockClient2{}, req)
		if err != nil {
			t.Logf("SectorRanking warning: %v", err)
			return
		}
		_ = result
	}, 15*time.Second)
}

// --- TestHandleIndustryRanking ---

func TestHandleIndustryRanking(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{},
			},
		}
		result, err := HandleIndustryRanking(ctx, &mockClient2{}, req)
		if err != nil {
			t.Logf("IndustryRanking warning: %v", err)
			return
		}
		_ = result
	}, 15*time.Second)
}

// --- TestHandleTopGainers ---

func TestHandleTopGainers(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"count": float64(10)},
			},
		}
		result, err := HandleTopGainers(ctx, &mockClient2{}, req)
		if err != nil {
			t.Logf("TopGainers warning: %v", err)
			return
		}
		_ = result
	}, 10*time.Second)
}

// --- TestHandleTopLosers ---

func TestHandleTopLosers(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		ctx := context.Background()
		req := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: mcp.CallToolParams{
				Arguments: map[string]any{"count": float64(10)},
			},
		}
		result, err := HandleTopLosers(ctx, &mockClient2{}, req)
		if err != nil {
			t.Logf("TopLosers warning: %v", err)
			return
		}
		_ = result
		}, 10*time.Second)
}

// --- TestHandleOfflineHome ---

func TestHandleOfflineHome(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleOfflineHome(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- TestHandleOfflineSyncDaily ---

func TestHandleOfflineSyncDaily(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleOfflineSyncDaily(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleOfflineSyncAll ---

func TestHandleOfflineSyncAll(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleOfflineSyncAll(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleBoardList ---

func TestHandleBoardList(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleBoardList(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleCapitalFlow ---

func TestHandleCapitalFlow(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleCapitalFlow(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleAnnouncement ---

func TestHandleAnnouncement(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleAnnouncement(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleFinancial ---

func TestHandleFinancial(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleFinancial(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleIndicatorCompute ---

func TestHandleIndicatorCompute(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1), "indicator": "MA"},
		},
	}
	result, err := HandleIndicatorCompute(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleExMarkets ---

func TestHandleExMarkets(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleExMarkets(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleExQuote ---

func TestHandleExQuote(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": "hk", "code": "00700"},
		},
	}
	result, err := HandleExQuote(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleExQuoteList ---

func TestHandleExQuoteList(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": "hk", "count": float64(10)},
		},
	}
	result, err := HandleExQuoteList(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleExKline ---

func TestHandleExKline(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": "hk", "code": "00700", "period": "day", "count": float64(10)},
		},
	}
	result, err := HandleExKline(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleExTick ---

func TestHandleExTick(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"market": "hk", "code": "00700"},
		},
	}
	result, err := HandleExTick(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleBacktest ---

func TestHandleBacktest(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"strategy": "ma_cross"},
		},
	}
	result, err := HandleBacktest(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleChanlun ---

func TestHandleChanlun(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleChanlun(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleAuction ---

func TestHandleAuction(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleAuction(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleUnusual ---

func TestHandleUnusual(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"code": "000001", "market": float64(1)},
		},
	}
	result, err := HandleUnusual(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestHandleMarketStat ---

func TestHandleMarketStat(t *testing.T) {
	ctx := context.Background()
	req := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := HandleMarketStat(ctx, &mockClient2{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// --- TestAllExpandedToolDefinitions ---

func TestAllExpandedToolDefinitions(t *testing.T) {
	defs := []struct {
		name string
		fn   func() mcp.Tool
	}{
		{ToolTick, NewTickTool},
		{ToolTransaction, NewTransactionTool},
		{ToolBoardList, NewBoardListTool},
		{ToolBoardMembers, NewBoardMembersTool},
		{ToolBelongBoard, NewBelongBoardTool},
		{ToolBoardRanking, NewBoardRankingTool},
		{ToolCapitalFlow, NewCapitalFlowTool},
		{ToolAuction, NewAuctionTool},
		{ToolUnusual, NewUnusualTool},
		{ToolMarketStat, NewMarketStatTool},
		{ToolServerInfo, NewServerInfoTool},
		{ToolSymbolInfo, NewSymbolInfoTool},
		{ToolAnnouncement, NewAnnouncementTool},
		{ToolFinancial, NewFinancialTool},
		{ToolIndicatorComp, NewIndicatorComputeTool},
		{ToolChanlun, NewChanlunTool},
		{ToolBacktest, NewBacktestTool},
		{ToolExMarkets, NewExMarketsTool},
		{ToolExKline, NewExKlineTool},
		{ToolExQuote, NewExQuoteTool},
		{ToolExQuoteList, NewExQuoteListTool},
		{ToolExTick, NewExTickTool},
		{ToolOfflineHome, NewOfflineHomeTool},
		{ToolOfflineDaily, NewOfflineDailyTool},
		{ToolOfflineMin, NewOfflineMinTool},
		{ToolOfflineGBBQ, NewOfflineGBBQTool},
		{ToolOfflineBlocks, NewOfflineBlocksTool},
		{ToolOfflineExFiles, NewOfflineExFilesTool},
		{ToolOfflineExDaily, NewOfflineExDailyTool},
		{ToolOfflineFinancial, NewOfflineFinancialTool},
		{ToolOfflineSyncDaily, NewOfflineSyncDailyTool},
		{ToolOfflineSyncAll, NewOfflineSyncAllTool},
		{ToolQuoteRealtime, NewQuoteRealtimeTool},
		{ToolQuoteListExtended, NewQuoteListExtendedTool},
		{ToolKlineExtended, NewKlineExtendedTool},
		{ToolDailyLineExtended, NewDailyLineExtendedTool},
		{ToolWeekLineExtended, NewWeekLineExtendedTool},
		{ToolMonthLineExtended, NewMonthLineExtendedTool},
		{Tool5MinLineExtended, New5MinLineExtendedTool},
		{Tool15MinLineExtended, New15MinLineExtendedTool},
		{Tool30MinLineExtended, New30MinLineExtendedTool},
		{Tool60MinLineExtended, New60MinLineExtendedTool},
		{ToolMACD, NewMACDTool},
		{ToolKDJ, NewKDJTool},
		{ToolRSI, NewRSITool},
		{ToolWR, NewWRTool},
		{ToolBOLL, NewBOLLTool},
		{ToolEMA, NewEMATool},
		{ToolDMA, NewDMATool},
		{ToolASI, NewASITool},
		{ToolVR, NewVRTool},
		{ToolROC, NewROCTool},
		{ToolOBV, NewOVBTool},
		{ToolMFI, NewMFITool},
		{ToolADX, NewADXTool},
		{ToolARBR, NewARBRTool},
		{ToolCCI, NewCCITool},
		{ToolDMI, NewDMITool},
		{ToolTECHNICAL_INDICATOR, NewTechnicalIndicatorTool},
		{ToolStockProfile, NewStockProfileTool},
		{ToolSectorRanking, NewSectorRankingTool},
		{ToolIndustryRanking, NewIndustryRankingTool},
		{ToolTopGainers, NewTopGainersTool},
		{ToolTopLosers, NewTopLosersTool},
	}

	if len(defs) != 64 {
		t.Errorf("expected 64 expanded tool definitions, got %d", len(defs))
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

// runWithTimeout wraps a test function with a timeout to prevent hanging on network calls.
func runWithTimeout(t *testing.T, fn func(*testing.T), timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn(t)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("test timed out after %v", timeout)
	}
}

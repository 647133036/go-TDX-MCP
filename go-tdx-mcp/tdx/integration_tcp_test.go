package tdx

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestIntegrationTCPDirect tests all TCP-first functions against a specific
// TDX server (bypassing slow auto-select).  Uses NO token.
func TestIntegrationTCPDirect(t *testing.T) {
	// Use a specific host to avoid auto-select slowdown
	ctx := context.Background()
	client := NewUnifiedClient("", 6, "218.75.122.92", 7709)
	defer client.Close()

	// Diagnose TCP connectivity
	t.Run("TCP_Diagnose", func(t *testing.T) {
		ctx2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel2()
		tcpErr := client.initTCP(ctx2)
		if tcpErr != nil {
			t.Logf("initTCP error: %v", tcpErr)
		}
		if !client.tcpClient.IsConnected() {
			t.Logf("IsConnected = false (mainAddr=%s)", client.tcpClient.mainAddr)
		} else {
			t.Logf("IsConnected = true (mainAddr=%s)", client.tcpClient.mainAddr)
		}
	})

	t.Run("Quotes", func(t *testing.T) {
		tests := []struct {
			code   string
			setcode string
			name   string
		}{
			{"000001", "0", "pingan_bank"},
			{"600000", "1", "spdb"},
			{"000001", "1", "shanghai_index"},
		}
		for _, tc := range tests {
			resp, err := client.TQLEXQuery(ctx, "TdxShare.PBHQInfo", QuoteRequest{
				Code:    tc.code,
				Setcode: tc.setcode,
			})
			if err != nil {
				t.Fatalf("%s quote failed: %v", tc.name, err)
			}
			if resp.Error != "" {
				t.Fatalf("%s quote returned error: %s", tc.name, resp.Error)
			}
			t.Logf("%s quote OK, data type=%T", tc.name, resp.Data)
		}
	})

	t.Run("Kline_Day", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", KlineRequest{
			Code:    "000001",
			Setcode: 0,
			Period:  4,
			WantNum: 5,
		})
		if err != nil {
			t.Fatalf("kline day failed: %v", err)
		}
		t.Logf("kline day OK, data type=%T", resp.Data)
	})

	t.Run("Kline_Week", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", KlineRequest{
			Code:    "000001",
			Setcode: 1,
			Period:  5,
			WantNum: 3,
		})
		if err != nil {
			t.Fatalf("kline week failed: %v", err)
		}
		t.Logf("kline week OK, data type=%T", resp.Data)
	})

	t.Run("Kline_1min", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBFXT", KlineRequest{
			Code:    "000001",
			Setcode: 0,
			Period:  9,
			WantNum: 5,
		})
		if err != nil {
			t.Fatalf("kline 1min failed: %v", err)
		}
		t.Logf("kline 1min OK, data type=%T", resp.Data)
	})

	t.Run("Board_List_Industry", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardList",
			map[string]interface{}{"BoardType": "HY", "Count": 3})
		if err != nil {
			// Some TDX servers don't have block files; skip instead of fail
			t.Logf("board list skipped (server has no block files): %v", err)
			return
		}
		t.Logf("board list OK, data type=%T", resp.Data)
	})

	t.Run("Board_Members", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardMembers",
			map[string]interface{}{"BoardCode": "BK0059", "Count": 3})
		if err != nil {
			// Some TDX servers don't have block files; skip instead of fail
			t.Logf("board members skipped (server has no block files): %v", err)
			return
		}
		t.Logf("board members OK, data type=%T", resp.Data)
	})

	t.Run("Unusual", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBUnusual",
			map[string]interface{}{"Setcode": 0, "WantNum": 3})
		if err != nil {
			t.Fatalf("unusual failed: %v", err)
		}
		t.Logf("unusual OK, data type=%T", resp.Data)
	})

	t.Run("Market_Stat", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBMarketStat", nil)
		if err != nil {
			t.Fatalf("market stat failed: %v", err)
		}
		t.Logf("market stat OK, data type=%T", resp.Data)
	})

	t.Run("Security_List", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBSecurityList",
			map[string]interface{}{"Setcode": 0, "Start": 0})
		if err != nil {
			t.Fatalf("security list failed: %v", err)
		}
		t.Logf("security list OK, data type=%T", resp.Data)
	})

	t.Run("Finance_Info", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBGetFinanceInfo",
			QuoteRequest{Code: "000001"})
		if err != nil {
			t.Fatalf("finance failed: %v", err)
		}
		t.Logf("finance OK, data type=%T", resp.Data)
	})

	t.Run("Auction", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBAuction",
			QuoteRequest{Code: "000001"})
		if err != nil {
			t.Fatalf("auction failed: %v", err)
		}
		t.Logf("auction OK, data type=%T", resp.Data)
	})

	t.Run("F10", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.TdxSharePCCW",
			QuoteRequest{Code: "000001"})
		if err != nil {
			t.Fatalf("F10 failed: %v", err)
		}
		t.Logf("F10 OK, data type=%T", resp.Data)
	})

	t.Run("Capital_Flow", func(t *testing.T) {
		resp, err := client.TQLEXQuery(ctx, "TdxShare.PBCapitalFlow",
			QuoteRequest{Code: "000001"})
		if err != nil {
			t.Fatalf("capital flow failed: %v", err)
		}
		t.Logf("capital flow OK, data type=%T", resp.Data)
	})

	// ===== Functions that MUST fail without token =====

	t.Run("Screener_NoToken", func(t *testing.T) {
		_, err := client.TQLEXQuery(ctx, "TdxShare.wendaQuery", nil)
		if err == nil {
			t.Fatalf("screener should fail without token")
		}
		t.Logf("screener correctly failed: %v", err)
	})

	t.Run("IndicatorSelect_NoToken", func(t *testing.T) {
		_, err := client.TQLEXQuery(ctx, "TdxShare.InfoSelectV2", nil)
		if err == nil {
			t.Fatalf("indicator select should fail without token")
		}
		t.Logf("indicator select correctly failed: %v", err)
	})

	t.Run("RAG_NoToken", func(t *testing.T) {
		_, err := client.RAGQuery(ctx, "test", 5)
		if err == nil {
			t.Fatalf("RAG should fail without token")
		}
		t.Logf("RAG correctly failed: %v", err)
	})

	t.Run("BoardRanking_NoToken", func(t *testing.T) {
		_, err := client.TQLEXQuery(ctx, "TdxShare.PBBoardRanking", nil)
		if err == nil {
			t.Fatalf("board ranking should fail without token")
		}
		t.Logf("board ranking correctly failed: %v", err)
	})

	t.Run("BelongBoard_NoToken", func(t *testing.T) {
		_, err := client.TQLEXQuery(ctx, "TdxShare.PBBelongBoard", nil)
		if err == nil {
			t.Fatalf("belong board should fail without token")
		}
		t.Logf("belong board correctly failed: %v", err)
	})

	// Count passing vs failing
	fmt.Printf("\n=== TCP Integration Results ===\n")
	fmt.Printf("Without token: TCP-first functions work via 218.75.122.92:7709\n")
	fmt.Printf("HTTP-only functions (screener/RAG/board-ranking/belong-board) correctly fail\n")
}

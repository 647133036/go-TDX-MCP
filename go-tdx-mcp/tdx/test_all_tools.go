package tdx

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type TestConfig struct {
	HandlerName string
	Handler     Handler
	Params      map[string]interface{}
	NeedsToken  bool
	IsCalcOnly  bool
	Source      string
}

type TestResult struct {
	Name   string
	Status string
	Reason string
	Source string
}

func RunAllTests() {
	configs := buildTestConfigs()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	defer client.Close()

	err := client.initTCP(ctx)
	tcpAvailable := err == nil
	if !tcpAvailable {
		fmt.Printf("\n[TCP CONNECTION FAILED] %v\n\n", err)
	}

	results := make([]TestResult, 0, len(configs))
	for _, tc := range configs {
		result := testOneTool(ctx, client, tc, tcpAvailable)
		results = append(results, result)
		fmt.Printf("[%s] %s: %s\n", result.Status, result.Name, result.Reason)
	}

	fmt.Printf("\n========== SUMMARY ==========\n")
	pass, fail, needsToken, empty := 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		case "NEEDS_TOKEN":
			needsToken++
		case "EMPTY":
			empty++
		}
	}
	fmt.Printf("PASS:       %d\n", pass)
	fmt.Printf("FAIL:       %d\n", fail)
	fmt.Printf("NEEDS_TOKEN: %d\n", needsToken)
	fmt.Printf("EMPTY:      %d\n", empty)
	fmt.Printf("TOTAL:      %d\n", len(results))
	fmt.Printf("================================\n")
}

func testOneTool(ctx context.Context, client Client, tc TestConfig, tcpAvailable bool) TestResult {
	if tc.Handler == nil {
		return TestResult{Name: tc.HandlerName, Status: "EMPTY", Reason: "handler is nil", Source: tc.Source}
	}

	if tc.NeedsToken {
		return TestResult{Name: tc.HandlerName, Status: "NEEDS_TOKEN", Reason: "requires TDX_TOKEN", Source: tc.Source}
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

	var result TestResult
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = TestResult{
					Name:   tc.HandlerName,
					Status: "FAIL",
					Reason: fmt.Sprintf("panic: %v", r),
					Source: tc.Source,
				}
			}
		}()

		_, err := tc.Handler(ctx, client, req)

		if err != nil {
			result = TestResult{
				Name:   tc.HandlerName,
				Status: "FAIL",
				Reason: fmt.Sprintf("handler error: %v", err),
				Source: tc.Source,
			}
			return
		}

		result = TestResult{
			Name:   tc.HandlerName,
			Status: "PASS",
			Reason: "ok",
			Source: tc.Source,
		}
	}()

	return result
}

func buildTestConfigs() []TestConfig {
	var configs []TestConfig

	// ===================== Core (tools.go) - 6 =====================
  configs = append(configs, TestConfig{
		HandlerName: ToolQuotes,
		Handler:     HandleQuotes,
		Params:      map[string]interface{}{"code": "000001", "setcode": "0"},
		Source:      "Core",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolKline,
		Handler:     HandleKline,
		Params:      map[string]interface{}{"code": "000001", "setcode": 0, "period": 4},
		Source:      "Core",
	})
   configs = append(configs, TestConfig{
		HandlerName: ToolLookupStock,
		Handler:     HandleLookupStock,
		Params:      map[string]interface{}{"query": "平安"},
		NeedsToken:  false,
		Source:      "Core",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolScreener,
		Handler:     HandleScreener,
		Params:      map[string]interface{}{"formula": "CLOSE > MA(CLOSE, 20)", "period": "day", "count": 100},
		Source:      "Core",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIndicatorSelect,
		Handler:     HandleIndicatorSelect,
		Params:      map[string]interface{}{"formula": "CLOSE > MA(CLOSE, 20)", "period": "day", "count": 100},
		Source:      "Core",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolApiData,
		Handler:     HandleApiData,
		Params:      map[string]interface{}{"code": "000001", "entry": "TdxShare.TdxSharePCCW", "fixedTag": "gsgy"},
		Source:      "Core",
	})

	// ===================== Expanded (tools_expanded.go) - 47 =====================
  configs = append(configs, TestConfig{
		HandlerName: ToolQuoteRealtime,
		Handler:     HandleQuoteRealtime,
		Params:      map[string]interface{}{"codes": []interface{}{"000001"}},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolQuoteListExtended,
		Handler:     HandleQuoteListExtended,
		Params:      map[string]interface{}{"market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolKlineExtended,
		Handler:     HandleKlineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolDailyLineExtended,
		Handler:     HandleDailyLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolWeekLineExtended,
		Handler:     HandleWeekLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMonthLineExtended,
		Handler:     HandleMonthLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: Tool5MinLineExtended,
		Handler:     Handle5MinLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: Tool15MinLineExtended,
		Handler:     Handle15MinLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: Tool30MinLineExtended,
		Handler:     Handle30MinLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: Tool60MinLineExtended,
		Handler:     Handle60MinLineExtended,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMACD,
		Handler:     HandleMACD,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolKDJ,
		Handler:     HandleKDJ,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolRSI,
		Handler:     HandleRSI,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolWR,
		Handler:     HandleWR,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBOLL,
		Handler:     HandleBOLL,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEMA,
		Handler:     HandleEMA,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolDMA,
		Handler:     HandleDMA,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolASI,
		Handler:     HandleASI,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolVR,
		Handler:     HandleVR,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolROC,
		Handler:     HandleROC,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolOBV,
		Handler:     HandleOBV,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMFI,
		Handler:     HandleMFI,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolADX,
		Handler:     HandleADX,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolARBR,
		Handler:     HandleARBR,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolCCI,
		Handler:     HandleCCI,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolDMI,
		Handler:     HandleDMI,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTECHNICAL_INDICATOR,
		Handler:     HandleTechnicalIndicator,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "indicators": "MACD,RSI", "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockProfile,
		Handler:     HandleStockProfile,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSectorRanking,
		Handler:     HandleSectorRanking,
		Params:      map[string]interface{}{"board_type": "1", "limit": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIndustryRanking,
		Handler:     HandleIndustryRanking,
		Params:      map[string]interface{}{"standard": "shenwan", "limit": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTopGainers,
		Handler:     HandleTopGainers,
		Params:      map[string]interface{}{"market": 0, "limit": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTopLosers,
		Handler:     HandleTopLosers,
		Params:      map[string]interface{}{"market": 0, "limit": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTick,
		Handler:     HandleTick,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTransaction,
		Handler:     HandleTransaction,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBoardList,
		Handler:     HandleBoardList,
		Params:      map[string]interface{}{"boardType": "industry", "count": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBoardMembers,
		Handler:     HandleBoardMembers,
		Params:      map[string]interface{}{"boardCode": "001781", "count": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBelongBoard,
		Handler:     HandleBelongBoard,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBoardRanking,
		Handler:     HandleBoardRanking,
		Params:      map[string]interface{}{"boardType": "1", "sortBy": "f146", "topN": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolCapitalFlow,
		Handler:     HandleCapitalFlow,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolAuction,
		Handler:     HandleAuction,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolUnusual,
		Handler:     HandleUnusual,
		Params:      map[string]interface{}{"market": 0, "count": 10},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarketStat,
		Handler:     HandleMarketStat,
		Params:      map[string]interface{}{"market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolServerInfo,
		Handler:     HandleServerInfo,
		Params:      map[string]interface{}{},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSymbolInfo,
		Handler:     HandleSymbolInfo,
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolAnnouncement,
		Handler:     HandleAnnouncement,
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFinancial,
		Handler:     HandleFinancial,
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBacktest,
		Handler:     HandleBacktest,
		Params:      map[string]interface{}{"code": "000001", "market": 0, "strategy": "SMA_CROSS", "period": "day", "count": 100},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolExKline,
		Handler:     HandleExKline,
		Params:      map[string]interface{}{"ex_market": "NASDAQ", "code": "BABA", "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolExMarkets,
		Handler:     HandleExMarkets,
		Params:      map[string]interface{}{},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolExQuote,
		Handler:     HandleExQuote,
		Params:      map[string]interface{}{"ex_market": "NASDAQ", "code": "BABA"},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolExQuoteList,
		Handler:     HandleExQuoteList,
		Params:      map[string]interface{}{"ex_market": "NASDAQ", "count": 5},
		Source:      "Expanded",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolExTick,
		Handler:     HandleExTick,
		Params:      map[string]interface{}{"ex_market": "NASDAQ", "code": "BABA"},
		Source:      "Expanded",
	})

	// ===================== V3 (tools_v3.go) - 8 =====================
  configs = append(configs, TestConfig{
		HandlerName: ToolMarketOverview,
		Handler:     HandleMarketOverview,
		Params:      map[string]interface{}{"board_type": "ALL"},
		Source:      "V3",
	})
   configs = append(configs, TestConfig{
		HandlerName: ToolSectorFlow,
		Handler:     HandleSectorFlow,
		Params:      map[string]interface{}{"board_type": "HY", "top_n": 5},
		NeedsToken:  false,
		Source:      "V3",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTopGainersLosers,
		Handler:     HandleTopGainersLosers,
		Params:      map[string]interface{}{"sort_type": "CHANGE_PCT", "top_n": 10, "direction": "both"},
		Source:      "V3",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFinancialMetrics,
		Handler:     HandleFinancialMetrics,
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "V3",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMacroData,
		Handler:     HandleMacroData,
		Params:      map[string]interface{}{"indicator": "CPI", "count": 6},
		Source:      "V3",
	})
   configs = append(configs, TestConfig{
		HandlerName: ToolWendaMacroQuery,
		Handler:     HandleWendaMacroQuery,
		Params:      map[string]interface{}{"query": "当前A股市场通胀压力如何？CPI和M2的最新数据是多少？", "top_k": 5},
		NeedsToken:  false,
		Source:      "V3",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNewsSentiment,
		Handler:     HandleNewsSentiment,
		Params:      map[string]interface{}{"code": "000001", "count": 5},
		Source:      "V3",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableScraper,
		Handler:     HandleTableScraper,
		Params:      map[string]interface{}{"query": "平安银行", "source": "iwencai"},
		Source:      "V3",
	})

	// ===================== New (tools_new.go) - 130+ =====================
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorList,
		Handler:     GetNewHandler(ToolFactorList),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorCompute,
		Handler:     GetNewHandler(ToolFactorCompute),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "factors": "MA7,MA20", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorAnalyze,
		Handler:     GetNewHandler(ToolFactorAnalyze),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "factor_name": "MA7", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolScreenScan,
		Handler:     GetNewHandler(ToolScreenScan),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolScreenStrength,
		Handler:     GetNewHandler(ToolScreenStrength),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "mode": "balanced", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEnhancedBacktest,
		Handler:     GetNewHandler(ToolEnhancedBacktest),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "strategy": "SMA_CROSS", "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTECryptoData,
		Handler:     GetNewHandler(ToolTECryptoData),
		Params:      map[string]interface{}{"symbols_crypto": "BTC,ETH"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTECryptoKline,
		Handler:     GetNewHandler(ToolTECryptoKline),
		Params:      map[string]interface{}{"symbol_crypto": "BTC", "interval": "1h", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundNAV,
		Handler:     GetNewHandler(ToolFundNAV),
		Params:      map[string]interface{}{"fund_code": "110011"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarginTrade,
		Handler:     GetNewHandler(ToolMarginTrade),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolDragonTiger,
		Handler:     GetNewHandler(ToolDragonTiger),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolConvertibleBond,
		Handler:     GetNewHandler(ToolConvertibleBond),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFuturesQuote,
		Handler:     GetNewHandler(ToolFuturesQuote),
		Params:      map[string]interface{}{"symbols": "CU0"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockCodeResolve,
		Handler:     GetNewHandler(ToolStockCodeResolve),
		Params:      map[string]interface{}{"codes": "000001,000002"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolCSIIndexConstituents,
		Handler:     GetNewHandler(ToolCSIIndexConstituents),
		Params:      map[string]interface{}{"index_code": "000300"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNewsSearch,
		Handler:     GetNewHandler(ToolNewsSearch),
		Params:      map[string]interface{}{"keyword": "平安银行"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolCurrentTimestamp,
		Handler:     GetNewHandler(ToolCurrentTimestamp),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTEFundData,
		Handler:     GetNewHandler(ToolTEFundData),
		Params:      map[string]interface{}{"symbol": "QQQ"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTEFuturesData,
		Handler:     GetNewHandler(ToolTEFuturesData),
		Params:      map[string]interface{}{"symbol": "GC"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTEMacroData,
		Handler:     GetNewHandler(ToolTEMacroData),
		Params:      map[string]interface{}{"indicator": "unemployment_rate"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSinaQuotes,
		Handler:     GetNewHandler(ToolSinaQuotes),
		Params:      map[string]interface{}{"codes": "sh000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSinaHKQuotes,
		Handler:     GetNewHandler(ToolSinaHKQuotes),
		Params:      map[string]interface{}{"codes": "hk00700"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSinaUSQuotes,
		Handler:     GetNewHandler(ToolSinaUSQuotes),
		Params:      map[string]interface{}{"codes": "gb_baba"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundHolding,
		Handler:     GetNewHandler(ToolFundHolding),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundManagers,
		Handler:     GetNewHandler(ToolFundManagers),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundSearch,
		Handler:     GetNewHandler(ToolFundSearch),
		Params:      map[string]interface{}{"keyword": "110011"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHKUSFinancial,
		Handler:     GetNewHandler(ToolHKUSFinancial),
		Params:      map[string]interface{}{"market": "HK", "code": "00700"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHKUSQuote,
		Handler:     GetNewHandler(ToolHKUSQuote),
		Params:      map[string]interface{}{"market": "HK", "code": "00700"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHKUSBasicInfo,
		Handler:     GetNewHandler(ToolHKUSBasicInfo),
		Params:      map[string]interface{}{"market": "HK", "code": "00700"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHKUSSearchStocks,
		Handler:     GetNewHandler(ToolHKUSSearchStocks),
		Params:      map[string]interface{}{"market": "HK", "keyword": "腾讯"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTrades,
		Handler:     GetNewHandler(ToolBlockTrades),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTradesByStock,
		Handler:     GetNewHandler(ToolBlockTradesByStock),
		Params:      map[string]interface{}{"code": "000001", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTradeStats,
		Handler:     GetNewHandler(ToolBlockTradeStats),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockActiveStocks,
		Handler:     GetNewHandler(ToolBlockActiveStocks),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSectorBoards,
		Handler:     GetNewHandler(ToolSectorBoards),
		Params:      map[string]interface{}{"board_type": "HY"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSectorBoardStocks,
		Handler:     GetNewHandler(ToolSectorBoardStocks),
		Params:      map[string]interface{}{"board_symbol": "001781"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMacroDataWeb,
		Handler:     GetNewHandler(ToolMacroDataWeb),
		Params:      map[string]interface{}{"indicator": "CPI", "count": 6},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundFlow,
		Handler:     GetNewHandler(ToolNorthboundFlow),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundDaily,
		Handler:     GetNewHandler(ToolNorthboundDaily),
		Params:      map[string]interface{}{"days": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundStocks,
		Handler:     GetNewHandler(ToolNorthboundStocks),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundHolders,
		Handler:     GetNewHandler(ToolNorthboundHolders),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundNavWeb,
		Handler:     GetNewHandler(ToolFundNavWeb),
		Params:      map[string]interface{}{"fund_code": "110011"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundNavHistory,
		Handler:     GetNewHandler(ToolFundNavHistory),
		Params:      map[string]interface{}{"fund_code": "110011", "limit": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarginTradeWeb,
		Handler:     GetNewHandler(ToolMarginTradeWeb),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolLimitUpPool,
		Handler:     GetNewHandler(ToolLimitUpPool),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolLimitDownPool,
		Handler:     GetNewHandler(ToolLimitDownPool),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolYesterdayLimitUp,
		Handler:     GetNewHandler(ToolYesterdayLimitUp),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHotRank,
		Handler:     GetNewHandler(ToolHotRank),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundTop10,
		Handler:     GetNewHandler(ToolNorthboundTop10),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarketIndices,
		Handler:     GetNewHandler(ToolMarketIndices),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarketIndicesFull,
		Handler:     GetNewHandler(ToolMarketIndicesFull),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSecurityList,
		Handler:     GetNewHandler(ToolSecurityList),
		Params:      map[string]interface{}{"market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSecurityCount,
		Handler:     GetNewHandler(ToolSecurityCount),
		Params:      map[string]interface{}{"market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTradesByDate,
		Handler:     GetNewHandler(ToolBlockTradesByDate),
		Params:      map[string]interface{}{"date": "20250106", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTradesSearch,
		Handler:     GetNewHandler(ToolBlockTradesSearch),
		Params:      map[string]interface{}{"keyword": "000001", "limit": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundCompanies,
		Handler:     GetNewHandler(ToolFundCompanies),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMacroMoneySupply,
		Handler:     GetNewHandler(ToolMacroMoneySupply),
		Params:      map[string]interface{}{"count": 6},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMacroGlobal,
		Handler:     GetNewHandler(ToolMacroGlobal),
		Params:      map[string]interface{}{"country": "US", "indicator": "CPI"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSinaMarginTrade,
		Handler:     GetNewHandler(ToolSinaMarginTrade),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSinaBlockTrades,
		Handler:     GetNewHandler(ToolSinaBlockTrades),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockBelongSector,
		Handler:     GetNewHandler(ToolStockBelongSector),
		Params:      map[string]interface{}{"codes": "000001,000002"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorTransform,
		Handler:     GetNewHandler(ToolFactorTransform),
		Params:      map[string]interface{}{"codes": "000001,000002", "factors": "MA7"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorCrossSection,
		Handler:     GetNewHandler(ToolFactorCrossSection),
		Params:      map[string]interface{}{"codes": "000001,000002", "factors": "MA7"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunDetail,
		Handler:     GetNewHandler(ToolChanlunDetail),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIndicatorSingle,
		Handler:     GetNewHandler(ToolIndicatorSingle),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "indicator": "MACD", "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBacktestPerformance,
		Handler:     GetNewHandler(ToolBacktestPerformance),
		Params:      map[string]interface{}{"trades_json": "[{\"date\":\"20250106\",\"code\":\"000001\",\"type\":\"buy\",\"price\":10,\"amount\":100}]"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolPortfolioOptimize,
		Handler:     GetNewHandler(ToolPortfolioOptimize),
		Params:      map[string]interface{}{"codes": "000001,000002", "weights": "50,50"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolPortfolioRisk,
		Handler:     GetNewHandler(ToolPortfolioRisk),
		Params:      map[string]interface{}{"codes": "000001,000002", "weights": "50,50"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolOCRRecognize,
		Handler:     GetNewHandler(ToolOCRRecognize),
		Params:      map[string]interface{}{"url": "https://example.com/image.png"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyRealtimeQuote,
		Handler:     GetNewHandler(ToolEastMoneyRealtimeQuote),
		Params:      map[string]interface{}{"codes": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyKlineHistory,
		Handler:     GetNewHandler(ToolEastMoneyKlineHistory),
		Params:      map[string]interface{}{"code": "000001", "period": "day", "count": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyStockChanges,
		Handler:     GetNewHandler(ToolEastMoneyStockChanges),
		Params:      map[string]interface{}{"change_type": "limit_up"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneySymbolInfo,
		Handler:     GetNewHandler(ToolEastMoneySymbolInfo),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneySectorBoards,
		Handler:     GetNewHandler(ToolEastMoneySectorBoards),
		Params:      map[string]interface{}{"board_type": "HY"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneySectorStocks,
		Handler:     GetNewHandler(ToolEastMoneySectorStocks),
		Params:      map[string]interface{}{"board_code": "BK0475"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyUpCount,
		Handler:     GetNewHandler(ToolEastMoneyUpCount),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyBelongBoard,
		Handler:     GetNewHandler(ToolEastMoneyBelongBoard),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyCompanyProfile,
		Handler:     GetNewHandler(ToolEastMoneyCompanyProfile),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyFinancialData,
		Handler:     GetNewHandler(ToolEastMoneyFinancialData),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolEastMoneyCapitalFlow,
		Handler:     GetNewHandler(ToolEastMoneyCapitalFlow),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundNavLatest,
		Handler:     GetNewHandler(ToolFundNavLatest),
		Params:      map[string]interface{}{"fund_code": "110011"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundNavHistoryNew,
		Handler:     GetNewHandler(ToolFundNavHistoryNew),
		Params:      map[string]interface{}{"fund_code": "110011", "limit": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarginTradeSummary,
		Handler:     GetNewHandler(ToolMarginTradeSummary),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableParserURL,
		Handler:     GetNewHandler(ToolTableParserURL),
		Params:      map[string]interface{}{"url": "https://finance.eastmoney.com/a/zcfgjj.html"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableParserHTML,
		Handler:     GetNewHandler(ToolTableParserHTML),
		Params:      map[string]interface{}{"html": "<table><tr><th>A</th></tr><tr><td>1</td></tr></table>"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableParserFindKeyword,
		Handler:     GetNewHandler(ToolTableParserFindKeyword),
		Params:      map[string]interface{}{"keyword": "平安"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableParserToCSV,
		Handler:     GetNewHandler(ToolTableParserToCSV),
		Params:      map[string]interface{}{"tables_json": "[{\"headers\":[\"A\"],\"rows\":[[\"1\"]]}]"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTableParserToJSON,
		Handler:     GetNewHandler(ToolTableParserToJSON),
		Params:      map[string]interface{}{"html": "<table><tr><th>A</th></tr><tr><td>1</td></tr></table>"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSCRaperIwencai,
		Handler:     GetNewHandler(ToolSCRaperIwencai),
		Params:      map[string]interface{}{"query": "平安银行"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSCRaperMultiSource,
		Handler:     GetNewHandler(ToolSCRaperMultiSource),
		Params:      map[string]interface{}{"query": "平安银行"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBacktestAvailable,
		Handler:     GetNewHandler(ToolBacktestAvailable),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBacktestRun,
		Handler:     GetNewHandler(ToolBacktestRun),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "strategy": "SMA_CROSS", "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBacktestCombo,
		Handler:     GetNewHandler(ToolBacktestCombo),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "strategy": "SMA_CROSS,MA_CROSS", "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorGetInfo,
		Handler:     GetNewHandler(ToolFactorGetInfo),
		Params:      map[string]interface{}{"factor_name": "MA7"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorAnalysisReport,
		Handler:     GetNewHandler(ToolFactorAnalysisReport),
		Params:      map[string]interface{}{"factor_name": "MA7", "codes": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFactorForwardReturns,
		Handler:     GetNewHandler(ToolFactorForwardReturns),
		Params:      map[string]interface{}{"factor_name": "MA7", "codes": "000001", "market": 0, "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunMergeKlines,
		Handler:     GetNewHandler(ToolChanlunMergeKlines),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunFindFenXing,
		Handler:     GetNewHandler(ToolChanlunFindFenXing),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunBuildBi,
		Handler:     GetNewHandler(ToolChanlunBuildBi),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunBuildZhongShu,
		Handler:     GetNewHandler(ToolChanlunBuildZhongShu),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolChanlunFindMaiMaiDian,
		Handler:     GetNewHandler(ToolChanlunFindMaiMaiDian),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
   configs = append(configs, TestConfig{
		HandlerName: ToolRAGQuery,
		Handler:     GetNewHandler(ToolRAGQuery),
		Params:      map[string]interface{}{"query": "000001 的 K 线走势如何？", "top_k": 3},
		NeedsToken:  false,
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolQuoteList,
		Handler:     GetNewHandler(ToolQuoteList),
		Params:      map[string]interface{}{"count": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolQuoteBatch,
		Handler:     GetNewHandler(ToolQuoteBatch),
		Params:      map[string]interface{}{"codes": "000001,000002"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolKlineData,
		Handler:     GetNewHandler(ToolKlineData),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFSMinuteData,
		Handler:     GetNewHandler(ToolFSMinuteData),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolTransactionData,
		Handler:     GetNewHandler(ToolTransactionData),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "count": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSecurityFilter,
		Handler:     GetNewHandler(ToolSecurityFilter),
		Params:      map[string]interface{}{"market": 0, "start": 0, "count": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockBasicInfo,
		Handler:     GetNewHandler(ToolStockBasicInfo),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockDividendInfo,
		Handler:     GetNewHandler(ToolStockDividendInfo),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockSplitInfo,
		Handler:     GetNewHandler(ToolStockSplitInfo),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIPOCalendar,
		Handler:     GetNewHandler(ToolIPOCalendar),
		Params:      map[string]interface{}{"date": "20250106", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockListByMarket,
		Handler:     GetNewHandler(ToolStockListByMarket),
		Params:      map[string]interface{}{"market": "SZ"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockListBySector,
		Handler:     GetNewHandler(ToolStockListBySector),
		Params:      map[string]interface{}{"sector": "index"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockListByIndustry,
		Handler:     GetNewHandler(ToolStockListByIndustry),
		Params:      map[string]interface{}{"industry": "银行"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockListByExchange,
		Handler:     GetNewHandler(ToolStockListByExchange),
		Params:      map[string]interface{}{"exchange": "SZ"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockListByStatus,
		Handler:     GetNewHandler(ToolStockListByStatus),
		Params:      map[string]interface{}{"status": "normal"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIndexConstituentList,
		Handler:     GetNewHandler(ToolIndexConstituentList),
		Params:      map[string]interface{}{"index_code": "000300"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolETFList,
		Handler:     GetNewHandler(ToolETFList),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolETFInfo,
		Handler:     GetNewHandler(ToolETFInfo),
		Params:      map[string]interface{}{"code": "159919"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolETFHoldings,
		Handler:     GetNewHandler(ToolETFHoldings),
		Params:      map[string]interface{}{"code": "159919"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolETFNetValue,
		Handler:     GetNewHandler(ToolETFNetValue),
		Params:      map[string]interface{}{"code": "159919"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFundamentalFilter,
		Handler:     GetNewHandler(ToolFundamentalFilter),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolPEPercentile,
		Handler:     GetNewHandler(ToolPEPercentile),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolPBPercentile,
		Handler:     GetNewHandler(ToolPBPercentile),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolRevenueGrowthRank,
		Handler:     GetNewHandler(ToolRevenueGrowthRank),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolProfitGrowthRank,
		Handler:     GetNewHandler(ToolProfitGrowthRank),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolROERank,
		Handler:     GetNewHandler(ToolROERank),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolDebtRatioRank,
		Handler:     GetNewHandler(ToolDebtRatioRank),
		Params:      map[string]interface{}{"limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolInsiderTrading,
		Handler:     GetNewHandler(ToolInsiderTrading),
		Params:      map[string]interface{}{"code": "000001", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolShareholderChange,
		Handler:     GetNewHandler(ToolShareholderChange),
		Params:      map[string]interface{}{"code": "000001"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarginDetail,
		Handler:     GetNewHandler(ToolMarginDetail),
		Params:      map[string]interface{}{"code": "000001", "limit": 10},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolNorthboundDetail,
		Handler:     GetNewHandler(ToolNorthboundDetail),
		Params:      map[string]interface{}{"code": "600036"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolBlockTradeDetail,
		Handler:     GetNewHandler(ToolBlockTradeDetail),
		Params:      map[string]interface{}{"code": "000001", "date": "20250106"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolSectorRotation,
		Handler:     GetNewHandler(ToolSectorRotation),
		Params:      map[string]interface{}{"days": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolMarketBreadth,
		Handler:     GetNewHandler(ToolMarketBreadth),
		Params:      map[string]interface{}{},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolVolumePriceAnalysis,
		Handler:     GetNewHandler(ToolVolumePriceAnalysis),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 100},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolCallAuction,
		Handler:     GetNewHandler(ToolCallAuction),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHistoryMinute,
		Handler:     GetNewHandler(ToolHistoryMinute),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "date": "20250106"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolHistoryTrade,
		Handler:     GetNewHandler(ToolHistoryTrade),
		Params:      map[string]interface{}{"code": "000001", "market": 0, "date": "20250106"},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolStockStats,
		Handler:     GetNewHandler(ToolStockStats),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFinanceInfo,
		Handler:     GetNewHandler(ToolFinanceInfo),
		Params:      map[string]interface{}{"code": "000001", "market": 0},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolIndexKline,
		Handler:     GetNewHandler(ToolIndexKline),
		Params:      map[string]interface{}{"code": "000001", "market": 1, "period": "day", "count": 5},
		Source:      "New",
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFormulaParse,
		Handler:     GetNewHandler(ToolFormulaParse),
		Params:      map[string]interface{}{"formula": "CLOSE/OPEN"},
		Source:      "New",
		IsCalcOnly:  true,
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFormulaExecute,
		Handler:     GetNewHandler(ToolFormulaExecute),
		Params:      map[string]interface{}{"formula": "CLOSE/OPEN", "count": 100},
		Source:      "New",
		IsCalcOnly:  true,
	})
  configs = append(configs, TestConfig{
		HandlerName: ToolFormulaList,
		Handler:     GetNewHandler(ToolFormulaList),
		Params:      map[string]interface{}{},
		Source:      "New",
		IsCalcOnly:  true,
	})

	return configs
}


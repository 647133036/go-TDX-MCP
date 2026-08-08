package tdx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRealDataDemo(t *testing.T) {
	fmt.Println("\n=== go-tdx-mcp 实际数据演示 ===")
	fmt.Println("启动:", time.Now().Format("15:04:05"))

	client := NewUnifiedClient("", 5, "218.75.122.92", 7709)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.initTCP(ctx)
	if err != nil {
		t.Logf("[WARN] TCP 连接失败: %v", err)
	} else {
		t.Logf("[OK] 通达信 TCP 连接成功")
	}

	cases := []struct {
		name string
		tool string
		args map[string]interface{}
	}{
		{"实时行情 深证成指", "tdx_quotes", map[string]interface{}{"code": "399001", "setcode": "0"}},
		{"日线K线 平安银行", "tdx_kline", map[string]interface{}{"code": "000001", "setcode": 0, "period": 4, "count": 5}},
		{"MACD指标", "tdx_macd_calc", map[string]interface{}{"code": "000001", "market": 0, "period": "day", "count": 10}},
		{"行业板块Top10", "tdx_board_list", map[string]interface{}{"board_type": "HY", "count": 10}},
		{"概念板块Top10", "tdx_board_list", map[string]interface{}{"board_type": "GN", "count": 10}},
		{"板块资金流Top5", "tdx_sector_flow", map[string]interface{}{"board_type": "HY", "top_n": 5}},
		{"涨幅榜/跌幅榜Top5", "tdx_top_gainers_losers", map[string]interface{}{"sort_type": "CHANGE_PCT", "top_n": 5, "direction": "both"}},
		{"大盘概况", "tdx_market_overview", map[string]interface{}{"board_type": "ALL"}},
		{"财务指标000001", "tdx_financial_metrics", map[string]interface{}{"code": "000001"}},
		{"宏观CPI近6期", "tdx_macro_data", map[string]interface{}{"indicator": "CPI", "count": 6}},
		{"宏观问答通胀", "wenda_macro_query", map[string]interface{}{"query": "当前通胀压力如何CPI M2数据", "top_k": 3}},
		{"RAG今天哪个板块涨得好", "tdx_rag_query", map[string]interface{}{"query": "今天哪个板块涨得好", "top_k": 5}},
		{"RAG平安银行技术面", "tdx_rag_query", map[string]interface{}{"query": "平安银行技术面怎么样", "top_k": 3}},
		{"涨停池", "tdx_limit_up_pool", map[string]interface{}{"limit": 5}},
		{"龙虎榜", "tdx_dragon_tiger", map[string]interface{}{"limit": 5}},
		{"融资融券", "tdx_margin_trade", map[string]interface{}{"limit": 5}},
		{"北向资金", "tdx_northbound_flow", map[string]interface{}{"limit": 5}},
		{"基金净值110011", "tdx_fund_nav", map[string]interface{}{"fund_code": "110011"}},
		{"股票代码查询平安银行", "tdx_lookup_stock", map[string]interface{}{"query": "平安银行"}},
		{"东方财富实时报价", "tdx_eastmoney_realtime_quote", map[string]interface{}{"codes": "000001,399001"}},
		{"选股股价突破MA20", "tdx_screener", map[string]interface{}{"formula": "CLOSE > MA(CLOSE, 20)", "period": "day", "count": 5}},
		{"回测ma_cross", "tdx_backtest", map[string]interface{}{"code": "000001", "market": 0, "strategy": "ma_cross", "period": "day", "count": 200}},
		{"全市场报价Top5", "tdx_quote_list_extended", map[string]interface{}{"market": 0, "count": 5}},
		{"时间戳", "tdx_current_timestamp", map[string]interface{}{}},
	}

	for _, tc := range cases {
		req := mcp.CallToolRequest{}
		req.Params.Name = tc.tool
		req.Params.Arguments = tc.args
		ctx2, cancel2 := context.WithTimeout(context.Background(), 8*time.Second)

		handler := GetHandler(tc.tool)
		var text string
		if handler == nil {
			text = "[HANDLER NOT FOUND]"
			cancel2()
			t.Logf("\n========== %s (%s) ==========", tc.name, tc.tool)
			t.Log(text)
			continue
		}
		resp, hErr := handler(ctx2, client, req)
		cancel2()

		if hErr != nil {
			text = fmt.Sprintf("[ERROR] %v", hErr)
		} else if resp == nil {
			text = "[NIL RESPONSE]"
		} else if len(resp.Content) == 0 {
			text = "[EMPTY CONTENT]"
		} else if tContent, ok := resp.Content[0].(mcp.TextContent); ok {
			text = tContent.Text
		} else {
			text = fmt.Sprintf("[UNKNOWN TYPE] %T", resp.Content[0])
		}

		t.Logf("\n========== %s (%s) ==========", tc.name, tc.tool)
		if len(text) > 2000 {
			t.Logf("%s...\n[truncated, total %d chars]", text[:2000], len(text))
		} else {
			t.Logf("%s", text)
		}
	}

	t.Log("\n=== 演示完成 ===")
}

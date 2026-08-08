package formula

import (
	"fmt"
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// 生成 50 根模拟 K 线数据，价格范围 10-20，带震荡
func generateTestBars() []indicator.Bar {
	bars := make([]indicator.Bar, 50)
	for i := range bars {
		bars[i] = indicator.Bar{
			Open:  12.0 + float64(i%7)*0.3,
			High:  12.5 + float64(i%5)*0.4,
			Low:   11.5 + float64(i%6)*0.3,
			Close: 12.2 + float64(i%8)*0.25,
			Vol:   5000 + float64(i%10)*1000,
			Amount: 1000000 + float64(i%10)*50000,
		}
	}
	return bars
}

// 生成 30 根数据，最后 5 根有 NaN
func generateBarsWithNaN() []indicator.Bar {
	return []indicator.Bar{
		{Open: 10, High: 11, Low: 9, Close: 10.5, Vol: 1000},
		{Open: 10.5, High: 12, Low: 10, Close: 11, Vol: 1200},
		{Open: 11, High: 12.5, Low: 10.5, Close: 12, Vol: 1500},
		{Open: 12, High: 13, Low: 11.5, Close: 12.5, Vol: 1800},
		{Open: 12.5, High: 14, Low: 12, Close: 13, Vol: 2000},
	}
}

// 功能测试：技术指标公式 — MACD
func TestFtMACDFormula(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
MA12:=EMA(C,12);
MA26:=EMA(C,26);
DIF:MA12-MA26;
DEA:EMA(DIF,9);
MACD:(DIF-DEA)*2;
`, bars)
	if err != nil {
		t.Fatalf("MACD formula failed: %v", err)
	}

	t.Logf("Type: %s", result.Type)
	t.Logf("Outputs count: %d", len(result.Outputs))
	for _, out := range result.Outputs {
		nonNaN := 0
		for _, v := range out.Data {
			if !math.IsNaN(v) {
				nonNaN++
			}
		}
		t.Logf("  %s: len=%d, nonNaN=%d, first=%.4f, last=%.4f",
			out.Name, len(out.Data), nonNaN,
			firstNonNaN(out.Data), lastNonNaN(out.Data))
	}
	t.Logf("FunctionCount: %d", result.FunctionCount)
}

// 功能测试：交易信号公式
func TestFtTradeSignalFormula(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
RSI14:RSI(C,14);
BUY:RSI14<30 AND CROSS(RSI14,30);
SELL:RSI14>70 AND CROSS(70,RSI14);
`, bars)
	if err != nil {
		t.Fatalf("trade formula failed: %v", err)
	}

	t.Logf("Type: %s", result.Type)
	t.Logf("TradeSignals: %v", result.TradeSignals)
}

// 功能测试：五彩K线公式
func TestFtColorfulKLineFormula(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
BKCOLOR:IF(C>O,0x008000,0xFF0000);
MA5:MA(C,5),LINETHICK2,COLORBLUE;
MA10:MA(C,10),COLORRED;
`, bars)
	if err != nil {
		t.Fatalf("colorful formula failed: %v", err)
	}

	t.Logf("Type: %s", result.Type)
	t.Logf("BKColor count: %d", len(result.BKColor))
	for _, out := range result.Outputs {
		nonNaN := 0
		for _, v := range out.Data {
			if !math.IsNaN(v) {
				nonNaN++
			}
		}
		styleInfo := ""
		if out.Style != nil {
			styleInfo = fmt.Sprintf(" style=%+v", *out.Style)
		}
		t.Logf("  %s: nonNaN=%d%s", out.Name, nonNaN, styleInfo)
	}
}

// 功能测试：条件选股公式（无输出线）
func TestFtSelectionFormula(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
XG:CLOSE>HIGH*1.03;
`, bars)
	if err != nil {
		t.Fatalf("selection formula failed: %v", err)
	}

	t.Logf("Type: %s", result.Type)
	t.Logf("Outputs: %v", result.Outputs)
}

// 功能测试：绘图函数
func TestFtDrawingFunctions(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
MA5:MA(C,5);
MA10:MA(C,10);
	DRAWTEXT(CROSS(MA5,MA10),H,'UP'),COLORGREEN;
	DRAWICON(CROSS(MA10,MA5),L,1),COLORRED;
`, bars)
	if err != nil {
		t.Fatalf("drawing formula failed: %v", err)
	}

	t.Logf("Drawings count: %d", len(result.Drawings))
	for _, d := range result.Drawings {
		t.Logf("  %s @ bar %d: values=%v text=%q",
			d.Function, d.BarIndex, d.Values, d.Text)
	}
}

// 功能测试：复合指标
func TestFtComplexIndicators(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	indicators := []string{"MACD", "MACD_DEA", "MACD_MACD", "KDJ", "KDJ_D", "KDJ_J",
		"BOLL", "BOLL_UPPER", "BOLL_LOWER", "RSI", "WR", "BBI"}

	for _, name := range indicators {
		src := fmt.Sprintf("%s:%s();", name, name)
		result, err := eng.Execute(src, bars)
		if err != nil {
			t.Logf("%s: ERROR %v", name, err)
			continue
		}
		if len(result.Outputs) == 0 {
			t.Logf("%s: no outputs", name)
			continue
		}
		for _, out := range result.Outputs {
			nonNaN := 0
			for _, v := range out.Data {
				if !math.IsNaN(v) {
					nonNaN++
				}
			}
			t.Logf("  %s: nonNaN=%d, first=%.4f, last=%.4f",
				out.Name, nonNaN, firstNonNaN(out.Data), lastNonNaN(out.Data))
		}
	}
}

// 功能测试：复杂公式 — RSI 多周期交叉
func TestFtRSICrossFormula(t *testing.T) {
	bars := generateTestBars()
	eng := NewEngine()

	result, err := eng.Execute(`
RSI6:RSI(C,6);
RSI12:RSI(C,12);
RSI24:RSI(C,24);
B:CROSS(RSI6,RSI12);
S:CROSS(RSI12,RSI6);
`, bars)
	if err != nil {
		t.Fatalf("RSI cross formula failed: %v", err)
	}

	t.Logf("Type: %s", result.Type)
	for _, out := range result.Outputs {
		nonNaN := 0
		for _, v := range out.Data {
			if !math.IsNaN(v) {
				nonNaN++
			}
		}
		t.Logf("  %s: nonNaN=%d", out.Name, nonNaN)
	}
}

// 功能测试：错误处理 — 无效公式
func TestFtErrorHandling(t *testing.T) {
	eng := NewEngine()
	tests := []struct {
		name    string
		formula string
	}{
		{"undefined var", "X:UNDEFINED_VAR+1;"},
		{"undefined func", "X:BADFUNC();" },
		{"parse error", "X:A B;"},
		{"unterminated block", "{ unclosed comment"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := eng.Execute(tc.formula, generateTestBars())
			if err == nil {
				t.Fatalf("expected error for '%s' but got nil", tc.name)
			}
			t.Logf("error: %v", err)
		})
	}
}

// 功能测试：ListFunctions
func TestFtListFunctions(t *testing.T) {
	eng := NewEngine()
	funcs := eng.ListFunctions()
	t.Logf("Total functions: %d", len(funcs))

	byCategory := make(map[string]int)
	for _, f := range funcs {
		byCategory[f.Category]++
	}
	for cat, count := range byCategory {
		t.Logf("  %s: %d", cat, count)
	}

	// 检查特定函数是否存在
	needles := []string{"MA", "EMA", "MACD", "KDJ", "RSI", "BOLL", "DRAWTEXT", "STICKLINE"}
	found := make(map[string]bool)
	for _, f := range funcs {
		found[f.Name] = true
	}
	for _, name := range needles {
		if !found[name] {
			t.Errorf("function %s not found", name)
		}
	}
}

// 功能测试：Parse 结果
func TestFtParseResults(t *testing.T) {
	eng := NewEngine()

	tests := []struct {
		name    string
		formula string
	}{
		{"indicator", "DIF:EMA(C,12)-EMA(C,26);DEA:EMA(DIF,9);MACD:(DIF-DEA)*2;"},
		{"trade", "ENTERLONG:CROSS(MA(C,5),MA(C,10));EXITLONG:CROSS(MA(C,10),MA(C,5));"},
		{"colorful", "BKCOLOR:IF(C>O,0x00FF00,0xFF0000);"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := eng.Parse(tc.formula)
			if err != nil {
				t.Fatalf("%s parse failed: %v", tc.name, err)
			}
			t.Logf("  type=%s, outputs=%d, drawings=%d",
				result.Type, len(result.Outputs), len(result.Drawings))
			for _, o := range result.Outputs {
				t.Logf("    output: %s (line %d)", o.Name, o.Line)
			}
		})
	}
}

func firstNonNaN(data []float64) float64 {
	for _, v := range data {
		if !math.IsNaN(v) {
			return v
		}
	}
	return math.NaN()
}

func lastNonNaN(data []float64) float64 {
	for i := len(data) - 1; i >= 0; i-- {
		if !math.IsNaN(data[i]) {
			return data[i]
		}
	}
	return math.NaN()
}

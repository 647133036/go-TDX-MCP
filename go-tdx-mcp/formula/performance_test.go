package formula

import (
	"math/rand"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func genBars(n int) []indicator.Bar {
	bars := make([]indicator.Bar, n)
	r := rand.New(rand.NewSource(42))
	for i := 0; i < n; i++ {
		base := 100.0 + r.Float64()*100.0
		bars[i] = indicator.Bar{
			Open:  base,
			High:  base + r.Float64()*5.0,
			Low:   base - r.Float64()*5.0,
			Close: base + r.Float64()*10.0 - 5.0,
			Vol:   10000 + r.Float64()*50000,
			Amount: 100000 + r.Float64()*500000,
		}
	}
	return bars
}

// Design target: 1000 bars, 10 statements, < 10ms
func BenchmarkFormula1000Bars10Stmts(b *testing.B) {
	bars := genBars(1000)
	eng := NewEngine()
	src := `
MA5:MA(C,5);
MA10:MA(C,10);
MA20:MA(C,20);
EMA12:EMA(C,12);
EMA26:EMA(C,26);
MACD:EMA12-EMA26;
RSI6:RSI(C,6);
RSI14:RSI(C,14);
ATR14:ATR(14);
DRAWICON(CROSS(MA5,MA10),H,1);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.Execute(src, bars)
		if err != nil {
			b.Fatalf("execute failed: %v", err)
		}
	}
}

func BenchmarkFormula100Bars(b *testing.B) {
	bars := genBars(100)
	eng := NewEngine()
	src := `
MA5:MA(C,5);
MACD:EMA(C,12)-EMA(C,26);
RSI6:RSI(C,6);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.Execute(src, bars)
		if err != nil {
			b.Fatalf("execute failed: %v", err)
		}
	}
}

func BenchmarkFormula10000Bars(b *testing.B) {
	bars := genBars(10000)
	eng := NewEngine()
	src := `
MA5:MA(C,5);
MA20:MA(C,20);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.Execute(src, bars)
		if err != nil {
			b.Fatalf("execute failed: %v", err)
		}
	}
}

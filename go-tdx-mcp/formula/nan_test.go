package formula

import (
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func TestNaNPropagation(t *testing.T) {
	bars := []indicator.Bar{
		{Open: 10, High: 12, Low: 9, Close: 11, Vol: 100, Amount: 1100},
		{Open: 11, High: 13, Low: 10, Close: 12, Vol: 120, Amount: 1440},
		{Open: 12, High: 14, Low: 11, Close: 13, Vol: 130, Amount: 1690},
	}
	eng := NewEngine()

	t.Run("MA with period > bars returns all NaN", func(t *testing.T) {
		result, err := eng.Execute("X:MA(C,5);", bars)
		if err != nil {
			t.Fatalf("MA failed: %v", err)
		}
		if len(result.Outputs) == 0 {
			t.Fatalf("expected output, got none")
		}
		for i := range result.Outputs[0].Data {
			if !math.IsNaN(result.Outputs[0].Data[i]) {
				t.Fatalf("bar %d: expected NaN, got %v", i, result.Outputs[0].Data[i])
			}
		}
	})

	t.Run("EMA with period > bars returns all NaN", func(t *testing.T) {
		result, err := eng.Execute("X:EMA(C,10);", bars)
		if err != nil {
			t.Fatalf("EMA failed: %v", err)
		}
		if len(result.Outputs) == 0 {
			t.Fatalf("expected output, got none")
		}
		for i, v := range result.Outputs[0].Data {
			if !math.IsNaN(v) {
				t.Fatalf("bar %d: expected NaN, got %v", i, v)
			}
		}
	})

	t.Run("NaN arithmetic propagates", func(t *testing.T) {
		result, err := eng.Execute("X:EMA(C,10)+MA(C,5);", bars)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if len(result.Outputs) == 0 {
			t.Fatalf("expected output, got none")
		}
		for i, v := range result.Outputs[0].Data {
			if !math.IsNaN(v) {
				t.Fatalf("bar %d: expected NaN, got %v", i, v)
			}
		}
	})

	t.Run("DRAWNULL produces NaN", func(t *testing.T) {
		// bar[0] C=10 (not >10), bar[1] C=11 (>10), bar[2] C=13 (>10)
		smallBars := []indicator.Bar{
			{Open: 10, High: 12, Low: 9, Close: 10, Vol: 100, Amount: 1100},
			{Open: 11, High: 13, Low: 10, Close: 11, Vol: 120, Amount: 1440},
			{Open: 12, High: 14, Low: 11, Close: 13, Vol: 130, Amount: 1690},
		}
		eng2 := NewEngine()
		result, err := eng2.Execute("X:DRAWNULL();Y:IF(C>10,C,DRAWNULL());", smallBars)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		var x, y *OutputLine
		for _, o := range result.Outputs {
			if o.Name == "X" {
				x = o
			}
			if o.Name == "Y" {
				y = o
			}
		}
		if x == nil || y == nil {
			t.Fatalf("missing output lines")
		}
		for i := range x.Data {
			if !math.IsNaN(x.Data[i]) {
				t.Fatalf("DRAWNULL bar %d: expected NaN, got %v", i, x.Data[i])
			}
		}
		for i, v := range y.Data {
			if i == 0 && !math.IsNaN(v) {
				t.Fatalf("Y bar 0: expected NaN (C=10 not >10), got %v", v)
			}
			if i > 0 && math.IsNaN(v) {
				t.Fatalf("Y bar %d: expected non-NaN, got %v", i, v)
			}
		}
	})
}

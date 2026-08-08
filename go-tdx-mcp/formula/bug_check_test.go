package formula

import (
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func TestBugs(t *testing.T) {
	bars := []indicator.Bar{
		{Open: 10, High: 11, Low: 9, Close: 10.5},
		{Open: 10.5, High: 12, Low: 10, Close: 11},
		{Open: 11, High: 12.5, Low: 10.5, Close: 12},
	}

	t.Run("STICKLINE_6_args", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("STICKLINE(C>O,H,L,2,0,0x00FF00);", bars)
		if err != nil {
			t.Fatalf("STICKLINE 6 args: %v", err)
		}
	})

	t.Run("STICKLINE_5_args", func(t *testing.T) {
		eng := NewEngine()
		result, err := eng.Execute("STICKLINE(C>O,H,L,2,0);", bars)
		if err != nil {
			t.Fatalf("STICKLINE 5 args: %v", err)
		}
		if len(result.Drawings) != 3 {
			t.Fatalf("expected 3 drawings, got %d", len(result.Drawings))
		}
		for _, d := range result.Drawings {
			if d.Function != "STICKLINE" {
				t.Fatalf("expected STICKLINE, got %s", d.Function)
			}
		}
	})

	t.Run("RGB", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("X:RGB(255,0,0);", bars)
		if err != nil {
			t.Fatalf("RGB: %v", err)
		}
	})

	t.Run("EMA_period_gt_bars", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("X:EMA(C,10);", bars)
		if err != nil {
			t.Fatalf("EMA(C,10): %v", err)
		}
	})

	t.Run("CROSS_scalar_broadcast", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("X:CROSS(C,11);", bars)
		if err != nil {
			t.Fatalf("CROSS(C,11): %v", err)
		}
	})

	t.Run("CROSS_both_scalar", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("X:CROSS(1,2);", bars)
		if err != nil {
			t.Fatalf("CROSS(1,2): %v", err)
		}
	})

	t.Run("FILTER", func(t *testing.T) {
		bars6 := make([]indicator.Bar, 6)
		for i := range bars6 {
			bars6[i] = indicator.Bar{Open: 10, Close: 10}
		}
		eng := NewEngine()
		result, err := eng.Execute("B:FILTER(C,3);", bars6)
		if err != nil {
			t.Fatalf("FILTER: %v", err)
		}
		if len(result.Outputs) != 1 {
			t.Fatalf("expected 1 output, got %d", len(result.Outputs))
		}
		expected := []float64{1, 0, 0, 1, 0, 0}
		for i, v := range expected {
			if result.Outputs[0].Data[i] != v {
				t.Fatalf("FILTER[%d]: expected %v, got %v", i, v, result.Outputs[0].Data[i])
			}
		}
	})

	t.Run("DRAWBAND", func(t *testing.T) {
		eng := NewEngine()
		result, err := eng.Execute("DRAWBAND(C,C,C,C);", bars)
		if err != nil {
			t.Fatalf("DRAWBAND: %v", err)
		}
		if len(result.Drawings) != 3 {
			t.Fatalf("expected 3 drawings, got %d", len(result.Drawings))
		}
	})

	t.Run("DRAWKLINE", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("DRAWKLINE(H,O,L,C);", bars)
		if err != nil {
			t.Fatalf("DRAWKLINE: %v", err)
		}
	})

	t.Run("POLYLINE", func(t *testing.T) {
		eng := NewEngine()
		_, err := eng.Execute("POLYLINE(C,C);", bars)
		if err != nil {
			t.Fatalf("POLYLINE: %v", err)
		}
	})
}

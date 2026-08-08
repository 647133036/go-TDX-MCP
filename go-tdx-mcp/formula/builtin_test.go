package formula

import (
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func newRegistry() *FunctionRegistry {
	return NewFunctionRegistry()
}

func testBarsForFn() []indicator.Bar {
	return []indicator.Bar{
		{Open: 10, High: 12, Low: 9, Close: 11, Vol: 100, Amount: 1100},
		{Open: 11, High: 13, Low: 10, Close: 12, Vol: 120, Amount: 1440},
		{Open: 12, High: 14, Low: 11, Close: 13, Vol: 130, Amount: 1690},
		{Open: 13, High: 15, Low: 12, Close: 14, Vol: 140, Amount: 1960},
		{Open: 14, High: 16, Low: 13, Close: 15, Vol: 150, Amount: 2250},
	}
}

func callFn(t *testing.T, name string, args ...*Value) (*Value, error) {
	t.Helper()
	reg := newRegistry()
	return reg.Call(name, args, testBarsForFn())
}

func seq(vs ...float64) *Value { return NewArrayValue(vs) }
func num(v float64) *Value      { return NewSingleValue(v) }

func assertValue(t *testing.T, got *Value, name string, want ...float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s returned nil value", name)
	}
	if !got.IsArray {
		t.Fatalf("%s expected array, got scalar %v", name, got.Single)
	}
	if len(got.Array) != len(want) {
		t.Fatalf("%s expected length %d, got %d: %v", name, len(want), len(got.Array), got.Array)
	}
	for i, w := range want {
		if math.IsNaN(w) {
			if !math.IsNaN(got.Array[i]) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, got.Array[i])
			}
			continue
		}
		if math.Abs(got.Array[i]-w) > 1e-9 {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, w, got.Array[i])
		}
	}
}

func TestFnMA(t *testing.T) {
	v, err := callFn(t, "MA", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("MA error: %v", err)
	}
	assertValue(t, v, "MA", math.NaN(), math.NaN(), 2, 3, 4)
}

func TestFnEMA(t *testing.T) {
	v, err := callFn(t, "EMA", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("EMA error: %v", err)
	}
	alpha := 2.0 / 4.0
	// seed = SMA(1,2,3) at index 2; out[0],out[1]=0
	want3 := alpha*4 + (1-alpha)*2
	want4 := alpha*5 + (1-alpha)*want3
	assertValue(t, v, "EMA", 0, 0, 2, want3, want4)
}

func TestFnSMA(t *testing.T) {
	v, err := callFn(t, "SMA", seq(1, 2, 3, 4, 5), num(3), num(2))
	if err != nil {
		t.Fatalf("SMA error: %v", err)
	}
	// y[0]=1; y[1]=(2*2+1*1)/3=5/3; y[2]=(2*3+1*5/3)/3=23/9 ...
	want2 := 23.0 / 9.0
	want3 := (2*4 + want2) / 3
	want4 := (2*5 + want3) / 3
	assertValue(t, v, "SMA", 1, 5.0/3.0, want2, want3, want4)
}

func TestFnSUM(t *testing.T) {
	v, err := callFn(t, "SUM", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("SUM error: %v", err)
	}
	assertValue(t, v, "SUM", math.NaN(), math.NaN(), 6, 9, 12)
}

func TestFnMAX(t *testing.T) {
	v, err := callFn(t, "MAX", seq(1, 5, 3, 8, 2), seq(4, 4, 4, 4, 4))
	if err != nil {
		t.Fatalf("MAX error: %v", err)
	}
	assertValue(t, v, "MAX", 4, 5, 4, 8, 4)
}

func TestFnMIN(t *testing.T) {
	v, err := callFn(t, "MIN", seq(1, 5, 3, 8, 2), seq(4, 4, 4, 4, 4))
	if err != nil {
		t.Fatalf("MIN error: %v", err)
	}
	assertValue(t, v, "MIN", 1, 4, 3, 4, 2)
}

func TestFnUnaryMath(t *testing.T) {
	cases := []struct {
		name string
		arg  *Value
		want []float64
	}{
		{"ABS", seq(-1, 2, -3.5), []float64{1, 2, 3.5}},
		{"SQRT", seq(4, 9, 16), []float64{2, 3, 4}},
		{"CEILING", seq(1.2, 2.0, 3.7), []float64{2, 2, 4}},
		{"FLOOR", seq(1.2, 2.0, 3.7), []float64{1, 2, 3}},
		{"INTPART", seq(1.2, -2.7, 3.0), []float64{1, -2, 3}},
		{"FRACPART", seq(1.25, -2.5, 3.0), []float64{0.25, -0.5, 0}},
		{"SIGN", seq(-5, 0, 7), []float64{-1, 0, 1}},
	}
	for _, c := range cases {
		v, err := callFn(t, c.name, c.arg)
		if err != nil {
			t.Fatalf("%s error: %v", c.name, err)
		}
		assertValue(t, v, c.name, c.want...)
	}
}

func TestFnBinaryMath(t *testing.T) {
	v, err := callFn(t, "POW", seq(2, 3), seq(3, 2))
	if err != nil {
		t.Fatalf("POW error: %v", err)
	}
	assertValue(t, v, "POW", 8, 9)

	v, err = callFn(t, "MOD", seq(7, 10), seq(3, 4))
	if err != nil {
		t.Fatalf("MOD error: %v", err)
	}
	assertValue(t, v, "MOD", 1, 2)
}

func TestFnROUND2(t *testing.T) {
	v, err := callFn(t, "ROUND2", seq(1.2345, 2.3456), num(2))
	if err != nil {
		t.Fatalf("ROUND2 error: %v", err)
	}
	if math.Abs(v.Array[0]-1.23) > 1e-9 {
		t.Errorf("ROUND2[0]: expected 1.23, got %v", v.Array[0])
	}
}

func TestFnREF(t *testing.T) {
	v, err := callFn(t, "REF", seq(1, 2, 3, 4, 5), num(2))
	if err != nil {
		t.Fatalf("REF error: %v", err)
	}
	assertValue(t, v, "REF", math.NaN(), math.NaN(), 1, 2, 3)
}

func TestFnREFX(t *testing.T) {
	v, err := callFn(t, "REFX", seq(1, 2, 3, 4, 5), num(1))
	if err != nil {
		t.Fatalf("REFX error: %v", err)
	}
	assertValue(t, v, "REFX", 2, 3, 4, 5, math.NaN())
}

func TestFnHHVLLV(t *testing.T) {
	v, err := callFn(t, "HHV", seq(3, 5, 2, 7, 4), num(3))
	if err != nil {
		t.Fatalf("HHV error: %v", err)
	}
	assertValue(t, v, "HHV", math.NaN(), math.NaN(), 5, 7, 7)

	v, err = callFn(t, "LLV", seq(3, 5, 2, 7, 4), num(3))
	if err != nil {
		t.Fatalf("LLV error: %v", err)
	}
	assertValue(t, v, "LLV", math.NaN(), math.NaN(), 2, 2, 2)
}

func TestFnHHVBARSLLVBARS(t *testing.T) {
	v, err := callFn(t, "HHVBARS", seq(3, 5, 5, 7, 4), num(3))
	if err != nil {
		t.Fatalf("HHVBARS error: %v", err)
	}
	// [3,5,5]->0; [5,5,7]->0; [5,7,4]->1
	assertValue(t, v, "HHVBARS", math.NaN(), math.NaN(), 0, 0, 1)

	v, err = callFn(t, "LLVBARS", seq(5, 3, 4, 2, 6), num(3))
	if err != nil {
		t.Fatalf("LLVBARS error: %v", err)
	}
	// [5,3,4]->1; [3,4,2]->0; [4,2,6]->1
	assertValue(t, v, "LLVBARS", math.NaN(), math.NaN(), 1, 0, 1)
}

func TestFnIF(t *testing.T) {
	v, err := callFn(t, "IF", seq(1, 0, 1, 0, 1), seq(10, 20, 30, 40, 50), seq(-1, -2, -3, -4, -5))
	if err != nil {
		t.Fatalf("IF error: %v", err)
	}
	assertValue(t, v, "IF", 10, -2, 30, -4, 50)
}

func TestFnCROSS(t *testing.T) {
	v, err := callFn(t, "CROSS", seq(1, 2, 1.5, 3, 2), seq(1.5, 1.5, 1.5, 1.5, 1.5))
	if err != nil {
		t.Fatalf("CROSS error: %v", err)
	}
	// i1: 1<=1.5 && 2>1.5 -> 1; i3: 1.5<=1.5 && 3>1.5 -> 1
	assertValue(t, v, "CROSS", 0, 1, 0, 1, 0)
}

func TestFnSTD(t *testing.T) {
	v, err := callFn(t, "STD", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("STD error: %v", err)
	}
	// window [1,2,3] mean2 variance=2/3; [2,3,4] -> 2/3; [3,4,5] -> 2/3
	s := math.Sqrt(2.0 / 3.0)
	assertValue(t, v, "STD", math.NaN(), math.NaN(), s, s, s)
}

func TestFnSTDDEV(t *testing.T) {
	v, err := callFn(t, "STDDEV", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("STDDEV error: %v", err)
	}
	// window [1,2,3] sample var = 1
	s := 1.0
	assertValue(t, v, "STDDEV", math.NaN(), math.NaN(), s, s, s)
}

func TestFnVAR(t *testing.T) {
	v, err := callFn(t, "VAR", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("VAR error: %v", err)
	}
	assertValue(t, v, "VAR", math.NaN(), math.NaN(), 2.0/3.0, 2.0/3.0, 2.0/3.0)
}

func TestFnWMA(t *testing.T) {
	v, err := callFn(t, "WMA", seq(1, 2, 3, 4, 5), num(3))
	if err != nil {
		t.Fatalf("WMA error: %v", err)
	}
	// weights 3,2,1 sum=6: [1,2,3]->(9+4+1)/6; [2,3,4]->(12+6+2)/6; [3,4,5]->(15+8+3)/6
	assertValue(t, v, "WMA", math.NaN(), math.NaN(), 14.0/6.0, 20.0/6.0, 26.0/6.0)
}

func TestFnCOUNT(t *testing.T) {
	v, err := callFn(t, "COUNT", seq(1, 0, 1, 1, 0), num(3))
	if err != nil {
		t.Fatalf("COUNT error: %v", err)
	}
	assertValue(t, v, "COUNT", math.NaN(), math.NaN(), 2, 2, 2)
}

func TestFnEVERYEXIST(t *testing.T) {
	v, err := callFn(t, "EVERY", seq(1, 1, 1, 0, 1), num(3))
	if err != nil {
		t.Fatalf("EVERY error: %v", err)
	}
	assertValue(t, v, "EVERY", 0, 0, 1, 0, 0)

	v, err = callFn(t, "EXIST", seq(0, 0, 1, 0, 0), num(3))
	if err != nil {
		t.Fatalf("EXIST error: %v", err)
	}
	assertValue(t, v, "EXIST", 0, 0, 1, 1, 1)
}

func TestFnBARSLAST(t *testing.T) {
	v, err := callFn(t, "BARSLAST", seq(0, 1, 0, 0, 1))
	if err != nil {
		t.Fatalf("BARSLAST error: %v", err)
	}
	assertValue(t, v, "BARSLAST", math.NaN(), 0, 1, 2, 0)
}

func TestFnBARSLASTCOUNT(t *testing.T) {
	v, err := callFn(t, "BARSLASTCOUNT", seq(0, 1, 1, 1, 0))
	if err != nil {
		t.Fatalf("BARSLASTCOUNT error: %v", err)
	}
	assertValue(t, v, "BARSLASTCOUNT", 0, 1, 2, 3, 0)
}

func TestFnBARSSINCE(t *testing.T) {
	v, err := callFn(t, "BARSSINCE", seq(0, 0, 1, 0, 0))
	if err != nil {
		t.Fatalf("BARSSINCE error: %v", err)
	}
	assertValue(t, v, "BARSSINCE", math.NaN(), math.NaN(), 0, 1, 2)
}

func TestFnBARSCOUNT(t *testing.T) {
	v, err := callFn(t, "BARSCOUNT", seq(1, 2, math.NaN(), 4, 5))
	if err != nil {
		t.Fatalf("BARSCOUNT error: %v", err)
	}
	assertValue(t, v, "BARSCOUNT", 1, 2, 2, 3, 4)
}

func TestFnFILTER(t *testing.T) {
	v, err := callFn(t, "FILTER", seq(1, 0, 0, 1, 0), num(2))
	if err != nil {
		t.Fatalf("FILTER error: %v", err)
	}
	assertValue(t, v, "FILTER", 1, 0, 0, 1, 0)
}

func TestFnBETWEEN(t *testing.T) {
	v, err := callFn(t, "BETWEEN", seq(5, 15, 20, 25), num(10), num(20))
	if err != nil {
		t.Fatalf("BETWEEN error: %v", err)
	}
	assertValue(t, v, "BETWEEN", 0, 1, 1, 0)
}

func TestFnNOT(t *testing.T) {
	v, err := callFn(t, "NOT", seq(1, 0, 0.5))
	if err != nil {
		t.Fatalf("NOT error: %v", err)
	}
	assertValue(t, v, "NOT", 0, 1, 0)
}

func TestFnCURRBARSCOUNT(t *testing.T) {
	reg := newRegistry()
	v, err := reg.Call("CURRBARSCOUNT", nil, testBarsForFn())
	if err != nil {
		t.Fatalf("CURRBARSCOUNT error: %v", err)
	}
	assertValue(t, v, "CURRBARSCOUNT", 5, 4, 3, 2, 1)
}

func TestFnTOTALBARSCOUNT(t *testing.T) {
	reg := newRegistry()
	v, err := reg.Call("TOTALBARSCOUNT", nil, testBarsForFn())
	if err != nil {
		t.Fatalf("TOTALBARSCOUNT error: %v", err)
	}
	assertValue(t, v, "TOTALBARSCOUNT", 5, 5, 5, 5, 5)
}

func TestFnISLASTBAR(t *testing.T) {
	reg := newRegistry()
	v, err := reg.Call("ISLASTBAR", nil, testBarsForFn())
	if err != nil {
		t.Fatalf("ISLASTBAR error: %v", err)
	}
	assertValue(t, v, "ISLASTBAR", 0, 0, 0, 0, 1)
}

func TestFnSUMBARS(t *testing.T) {
	v, err := callFn(t, "SUMBARS", seq(2, 3, 4, 5, 6), num(9))
	if err != nil {
		t.Fatalf("SUMBARS error: %v", err)
	}
	// i0: 2<9 NaN; i1: 2+3=5<9 NaN; i2: 2+3+4=9 -> 3; i3: 5+4=9 -> 2; i4: 6+5=11 -> 2
	assertValue(t, v, "SUMBARS", math.NaN(), math.NaN(), 3, 2, 2)
}

func TestFnVALUEWHEN(t *testing.T) {
	v, err := callFn(t, "VALUEWHEN", seq(1, 0, 0, 1, 0), seq(10, 99, 99, 20, 99))
	if err != nil {
		t.Fatalf("VALUEWHEN error: %v", err)
	}
	assertValue(t, v, "VALUEWHEN", 10, 10, 10, 20, 20)
}

func TestFnDMA(t *testing.T) {
	v, err := callFn(t, "DMA", seq(1, 2, 3, 4, 5), num(0.5))
	if err != nil {
		t.Fatalf("DMA error: %v", err)
	}
	// y0=1; y1=0.5*2+0.5*1=1.5; y2=0.5*3+0.5*1.5=2.25; y3=0.5*4+0.5*2.25=3.125; y4=0.5*5+0.5*3.125=4.0625
	assertValue(t, v, "DMA", 1, 1.5, 2.25, 3.125, 4.0625)
}

func TestFnCONST(t *testing.T) {
	v, err := callFn(t, "CONST", seq(1, 2, 3, 4, 5))
	if err != nil {
		t.Fatalf("CONST error: %v", err)
	}
	assertValue(t, v, "CONST", 5, 5, 5, 5, 5)
}

func TestFnArgCountErrors(t *testing.T) {
	_, err := callFn(t, "MA", seq(1, 2, 3))
	if err == nil {
		t.Fatal("MA with 1 arg should error")
	}
	v, err := callFn(t, "MA", seq(1, 2, 3), num(3), num(2))
	if err == nil {
		t.Fatal("MA with 3 args should error")
	}
	v, err = callFn(t, "CROSS", seq(1, 2), num(2))
	if err != nil {
		t.Fatalf("CROSS with scalar should broadcast: %v", err)
	}
	assertValue(t, v, "CROSS", 0, 0)
	_, err = callFn(t, "MA", seq(1, 2, 3), num(0))
	if err == nil {
		t.Fatal("MA with period 0 should error")
	}
	_, err = callFn(t, "NONEXIST", seq(1))
	if err == nil {
		t.Fatal("undefined function should error")
	}
}

func TestFnDrawingEvents(t *testing.T) {
	v, err := callFn(t, "DRAWTEXT", seq(0, 1, 0, 1, 0), num(10), NewStringValue("buy"))
	if err != nil {
		t.Fatalf("DRAWTEXT error: %v", err)
	}
	if !v.IsDraw {
		t.Fatal("DRAWTEXT should return draw value")
	}
	if len(v.Draws) != 2 {
		t.Fatalf("expected 2 drawings, got %d", len(v.Draws))
	}
	if v.Draws[0].BarIndex != 1 || v.Draws[1].BarIndex != 3 {
		t.Errorf("unexpected bar indices: %v, %v", v.Draws[0].BarIndex, v.Draws[1].BarIndex)
	}
	if v.Draws[0].Text != "buy" {
		t.Errorf("expected text buy, got %q", v.Draws[0].Text)
	}

	v, err = callFn(t, "STICKLINE", seq(1, 0, 1), seq(10, 11, 12), seq(9, 8, 7), num(1), num(0))
	if err != nil {
		t.Fatalf("STICKLINE error: %v", err)
	}
	if len(v.Draws) != 2 {
		t.Fatalf("expected 2 stickline drawings, got %d", len(v.Draws))
	}
}

func TestFnArityMeta(t *testing.T) {
	reg := newRegistry()
	if _, ok := reg.Lookup("MA"); !ok {
		t.Fatal("MA should be registered")
	}
	if _, ok := reg.Lookup("ma"); !ok {
		t.Fatal("lowercase ma should resolve")
	}
	names := reg.Names()
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	required := []string{"MA", "EMA", "CROSS", "REF", "DRAWNULL", "DRAWTEXT", "MACD_placeholder"}
	for _, r := range required {
		if r == "MACD_placeholder" {
			continue
		}
		if !found[r] {
			t.Errorf("function %s missing from registry", r)
		}
	}
}

package formula

import (
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func testBars() []indicator.Bar {
	return []indicator.Bar{
		{Open: 100, High: 105, Low: 99, Close: 104, Vol: 1000, Amount: 104000},
		{Open: 104, High: 108, Low: 102, Close: 106, Vol: 1100, Amount: 116600},
		{Open: 106, High: 109, Low: 101, Close: 103, Vol: 1200, Amount: 123600},
		{Open: 103, High: 110, Low: 102, Close: 109, Vol: 1300, Amount: 141700},
		{Open: 109, High: 113, Low: 107, Close: 112, Vol: 1400, Amount: 156800},
	}
}

func runEval(t *testing.T, src string, bars []indicator.Bar) (*FormulaResult, error) {
	t.Helper()
	toks, err := NewLexer(src).Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	prog, err := NewParser(toks).Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	return NewInterpreter(bars).Execute(prog)
}

func TestEvalSeriesRef(t *testing.T) {
	bars := testBars()
	res, err := runEval(t, "C:CLOSE;", bars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(res.Outputs))
	}
	got := res.Outputs[0].Data
	if len(got) != len(bars) {
		t.Fatalf("expected length %d, got %d", len(bars), len(got))
	}
	for i, b := range bars {
		if got[i] != b.Close {
			t.Errorf("index %d: expected %v, got %v", i, b.Close, got[i])
		}
	}
}

func TestEvalArithmetic(t *testing.T) {
	res, err := runEval(t, "X:CLOSE*2+1;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := res.Outputs[0].Data
	if got[0] != 209 {
		t.Errorf("expected 209, got %v", got[0])
	}
}

func TestEvalScalarBroadcast(t *testing.T) {
	res, err := runEval(t, "X:CLOSE>100;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, v := range res.Outputs[0].Data {
		if v != 1 {
			t.Errorf("index %d: expected 1, got %v", i, v)
		}
	}
}

func TestEvalVariableOrder(t *testing.T) {
	res, err := runEval(t, "A:=CLOSE;B:A+1;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := res.Outputs[0].Data
	for i, b := range testBars() {
		if got[i] != b.Close+1 {
			t.Errorf("index %d: expected %v, got %v", i, b.Close+1, got[i])
		}
	}
}

func TestEvalDivisionByZero(t *testing.T) {
	_, err := runEval(t, "X:1/0;", testBars())
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	if !isEvalError(err) {
		t.Errorf("expected eval error, got %v", err)
	}
}

func TestEvalUndefinedVariable(t *testing.T) {
	_, err := runEval(t, "X:FOO+1;", testBars())
	if err == nil {
		t.Fatal("expected undefined variable error")
	}
}

func TestEvalUndefinedFunction(t *testing.T) {
	_, err := runEval(t, "X:NOSUCHFUNC(C,5);", testBars())
	if err == nil {
		t.Fatal("expected undefined function error")
	}
	if !isEvalError(err) {
		t.Errorf("expected eval error, got %v", err)
	}
}

func TestEvalComparisonOps(t *testing.T) {
	res, err := runEval(t, "X:C>=O;Y:C=104;Z:C<>106;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := res.Outputs[0].Data
	for i, b := range testBars() {
		want := 0.0
		if b.Close >= b.Open {
			want = 1
		}
		if x[i] != want {
			t.Errorf("X index %d: expected %v, got %v", i, want, x[i])
		}
	}
	y := res.Outputs[1].Data
	if y[0] != 1 || y[1] != 0 {
		t.Errorf("expected Y [1,0,...], got %v", y)
	}
	z := res.Outputs[2].Data
	if z[0] != 1 || z[1] != 0 {
		t.Errorf("expected Z [1,0,...], got %v", z)
	}
}

func TestEvalLogicalOps(t *testing.T) {
	res, err := runEval(t, "X:(C>O) AND (H>=L);Y:NOT(C<O);", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, b := range testBars() {
		want := 0.0
		if b.Close > b.Open && b.High >= b.Low {
			want = 1
		}
		if res.Outputs[0].Data[i] != want {
			t.Errorf("index %d: expected %v, got %v", i, want, res.Outputs[0].Data[i])
		}
	}
}

func TestEvalPowerOperator(t *testing.T) {
	res, err := runEval(t, "X:2^3;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Outputs[0].Data[0] != 8 {
		t.Errorf("expected 8, got %v", res.Outputs[0].Data[0])
	}
}

func TestEvalUnaryMinus(t *testing.T) {
	res, err := runEval(t, "X:-CLOSE;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, b := range testBars() {
		if res.Outputs[0].Data[i] != -b.Close {
			t.Errorf("index %d: expected %v, got %v", i, -b.Close, res.Outputs[0].Data[i])
		}
	}
}

func TestEvalNaNPropagation(t *testing.T) {
	res, err := runEval(t, "X:DRAWNULL()*2;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, v := range res.Outputs[0].Data {
		if !math.IsNaN(v) {
			t.Errorf("expected NaN, got %v", v)
		}
	}
}

func TestEvalSeriesVolumeAlias(t *testing.T) {
	res, err := runEval(t, "V1:VOL;V2:V;A1:AMOUNT;A2:AMO;", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Outputs) != 4 {
		t.Fatalf("expected 4 outputs, got %d", len(res.Outputs))
	}
	for i, b := range testBars() {
		if res.Outputs[0].Data[i] != b.Vol || res.Outputs[1].Data[i] != b.Vol {
			t.Errorf("VOL/V index %d: expected %v", i, b.Vol)
		}
		if res.Outputs[2].Data[i] != b.Amount || res.Outputs[3].Data[i] != b.Amount {
			t.Errorf("AMOUNT/AMO index %d: expected %v", i, b.Amount)
		}
	}
}

func isEvalError(err error) bool {
	fe, ok := err.(*FormulaError)
	return ok && fe.Kind == KindEval
}

func TestEvalEmptyBars(t *testing.T) {
	res, err := runEval(t, "X:CLOSE;", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Outputs) != 1 {
		t.Fatalf("expected 1 output line, got %d", len(res.Outputs))
	}
	if len(res.Outputs[0].Data) != 0 {
		t.Errorf("expected empty data for empty bars, got %d", len(res.Outputs[0].Data))
	}
}

func TestEvalNaNValue(t *testing.T) {
	res, err := runEval(t, "X:DRAWNULL();", testBars())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, v := range res.Outputs[0].Data {
		if !math.IsNaN(v) {
			t.Errorf("expected NaN, got %v", v)
		}
	}
}

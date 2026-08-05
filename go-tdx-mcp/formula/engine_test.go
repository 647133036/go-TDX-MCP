package formula

import (
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func newEngine() *Engine {
	return NewEngine()
}

func TestEngineParseIndicator(t *testing.T) {
	e := newEngine()
	res, err := e.Parse("A:=CLOSE;MA5:MA(A,5);")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Type != "indicator" {
		t.Errorf("expected indicator type, got %q", res.Type)
	}
	if len(res.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(res.Outputs))
	}
	if res.Outputs[0].Name != "MA5" {
		t.Errorf("expected output MA5, got %s", res.Outputs[0].Name)
	}
	if len(res.Body) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(res.Body))
	}
}

func TestEngineParseTrade(t *testing.T) {
	e := newEngine()
	res, err := e.Parse("ENTERLONG:CROSS(MA(CLOSE,5),MA(CLOSE,10));")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Type != "trade" {
		t.Errorf("expected trade type, got %q", res.Type)
	}
	if len(res.Outputs) != 1 || res.Outputs[0].Name != "ENTERLONG" {
		t.Errorf("expected ENTERLONG output, got %+v", res.Outputs)
	}
}

func TestEngineParseColorful(t *testing.T) {
	e := newEngine()
	res, err := e.Parse("BKCOLOR:CLOSE>OPEN;")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Type != "colorful" {
		t.Errorf("expected colorful type, got %q", res.Type)
	}
}

func TestEngineParseSelection(t *testing.T) {
	e := newEngine()
	res, err := e.Parse("CLOSE>MA(CLOSE,5);")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Type != "selection" {
		t.Errorf("expected selection type, got %q", res.Type)
	}
}

func TestEngineParseDrawing(t *testing.T) {
	e := newEngine()
	res, err := e.Parse("DRAWTEXT(CROSS(MA(CLOSE,5),MA(CLOSE,10)),LOW,'买点');")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(res.Drawings) != 1 {
		t.Fatalf("expected 1 drawing, got %d", len(res.Drawings))
	}
	if res.Drawings[0].Func != "DRAWTEXT" {
		t.Errorf("expected DRAWTEXT, got %s", res.Drawings[0].Func)
	}
}

func TestEngineExecuteMACD(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("DIF:EMA(CLOSE,12)-EMA(CLOSE,26);DEA:EMA(DIF,9);", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if len(res.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(res.Outputs))
	}
	ind := indicator.MACD(bars, 12, 26, 9)
	for i, v := range res.Outputs[0].Data {
		if math.IsNaN(ind.Values[i]) {
			if !math.IsNaN(v) {
				t.Errorf("DIF[%d]: expected NaN, got %v", i, v)
			}
			continue
		}
		if math.Abs(v-ind.Values[i]) > 1e-9 {
			t.Errorf("DIF[%d]: expected %v, got %v", i, ind.Values[i], v)
		}
	}
	if len(res.NanCounts) != 2 {
		t.Errorf("expected nan_counts for 2 outputs, got %v", res.NanCounts)
	}
}

func TestEngineExecuteKDJGoldenCrossSelection(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("RSV:=(CLOSE-LLV(LOW,9))/(HHV(HIGH,9)-LLV(LOW,9))*100;K:SMA(RSV,3,1);D:SMA(K,3,1);J:3*K-2*D;", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if len(res.Outputs) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(res.Outputs))
	}
	ind := indicator.KDJ(bars, 9, 3, 3)
	for i, v := range res.Outputs[0].Data {
		if math.Abs(v-ind.Values[i]) > 1e-9 {
			t.Errorf("K[%d]: expected %v, got %v", i, ind.Values[i], v)
		}
	}
}

func TestEngineExecuteTradeSignals(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("A1:=MA(CLOSE,5);A2:=MA(CLOSE,10);ENTERLONG:CROSS(A1,A2);EXITLONG:A2>A1;", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.Type != "trade" {
		t.Errorf("expected trade type, got %q", res.Type)
	}
	if len(res.TradeSignals) != 2 {
		t.Fatalf("expected 2 trade signals, got %d", len(res.TradeSignals))
	}
	if _, ok := res.TradeSignals["ENTERLONG"]; !ok {
		t.Error("ENTERLONG signal missing")
	}
	if _, ok := res.TradeSignals["EXITLONG"]; !ok {
		t.Error("EXITLONG signal missing")
	}
	if len(res.TradeSignals["ENTERLONG"]) != len(bars) {
		t.Errorf("ENTERLONG length mismatch: %d != %d", len(res.TradeSignals["ENTERLONG"]), len(bars))
	}
}

func TestEngineExecuteBKColor(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("BKCOLOR:CLOSE>OPEN;", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.Type != "colorful" {
		t.Errorf("expected colorful type, got %q", res.Type)
	}
	if len(res.BKColor) != len(bars) {
		t.Fatalf("expected bkcolor length %d, got %d", len(bars), len(res.BKColor))
	}
}

func TestEngineExecuteDrawingEvents(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("DRAWTEXT(CLOSE>OPEN,LOW,'BUY');", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if len(res.Drawings) == 0 {
		t.Fatal("expected drawing events")
	}
	for _, d := range res.Drawings {
		if d.Function != "DRAWTEXT" {
			t.Errorf("expected DRAWTEXT event, got %s", d.Function)
		}
		if d.Text != "BUY" {
			t.Errorf("expected BUY text, got %q", d.Text)
		}
	}
}

func TestEngineExecuteErrorSerialization(t *testing.T) {
	e := newEngine()
	_, err := e.Execute("X:1/0;", complexBars())
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	fe, ok := err.(*FormulaError)
	if !ok {
		t.Fatalf("expected *FormulaError, got %T", err)
	}
	if fe.Kind != KindEval {
		t.Errorf("expected eval kind, got %q", fe.Kind)
	}
	if fe.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestEngineListFunctions(t *testing.T) {
	e := newEngine()
	infos := e.ListFunctions()
	if len(infos) < 100 {
		t.Fatalf("expected at least 100 functions, got %d", len(infos))
	}
	names := map[string]bool{}
	for _, info := range infos {
		if info.Name == "" {
			t.Error("function with empty name")
		}
		if names[info.Name] {
			t.Errorf("duplicate function %s", info.Name)
		}
		names[info.Name] = true
	}
	if !names["MA"] || !names["MACD"] || !names["DRAWTEXT"] {
		t.Error("expected MA/MACD/DRAWTEXT in function list")
	}
}

func TestEngineExecuteMixedTypes(t *testing.T) {
	e := newEngine()
	bars := complexBars()
	res, err := e.Execute("A1:=MA(CLOSE,5);MA5:A1;ENTERLONG:CROSS(A1,MA(CLOSE,10));", bars)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	// trade type wins even with an output line present
	if res.Type != "trade" {
		t.Errorf("expected trade type, got %q", res.Type)
	}
	if len(res.Outputs) != 1 {
		t.Errorf("expected 1 output line, got %d", len(res.Outputs))
	}
}

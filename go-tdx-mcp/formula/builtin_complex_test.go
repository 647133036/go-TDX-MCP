package formula

import (
	"math"
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// Test data: 40 bars with deterministic values for stable indicator output.
func complexBars() []indicator.Bar {
	bars := make([]indicator.Bar, 40)
	for i := range bars {
		base := 100.0 + float64(i)*0.5
		bars[i] = indicator.Bar{
			Open:   base,
			High:   base + 2 + float64(i%3),
			Low:    base - 2 - float64(i%2),
			Close:  base + 1 + float64(i%5)*0.3,
			Vol:    1000 + float64(i)*10,
			Amount: (1000 + float64(i)*10) * base,
		}
	}
	return bars
}

func callComplex(t *testing.T, name string, args ...*Value) (*Value, error) {
	t.Helper()
	reg := newRegistry()
	return reg.Call(name, args, complexBars())
}

func assertSeriesEqual(t *testing.T, got *Value, name string, want []float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s returned nil", name)
	}
	if !got.IsArray {
		t.Fatalf("%s expected array, got scalar %v", name, got.Single)
	}
	if len(got.Array) != len(want) {
		t.Fatalf("%s expected length %d, got %d", name, len(want), len(got.Array))
	}
	for i := range want {
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got.Array[i]) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, got.Array[i])
			}
			continue
		}
		if math.IsNaN(got.Array[i]) {
			t.Errorf("%s[%d]: expected %v, got NaN", name, i, want[i])
			continue
		}
		if math.Abs(got.Array[i]-want[i]) > 1e-9 {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, want[i], got.Array[i])
		}
	}
}

func TestComplexMACD(t *testing.T) {
	bars := complexBars()
	v, err := callComplex(t, "MACD", num(12), num(26), num(9))
	if err != nil {
		t.Fatalf("MACD error: %v", err)
	}
	assertSeriesEqual(t, v, "MACD", indicator.MACD(bars, 12, 26, 9).Values)

	v, err = callComplex(t, "MACD_DEA", num(12), num(26), num(9))
	if err != nil {
		t.Fatalf("MACD_DEA error: %v", err)
	}
	assertSeriesEqual(t, v, "MACD_DEA", indicator.MACD(bars, 12, 26, 9).Line2)

	v, err = callComplex(t, "MACD_MACD", num(12), num(26), num(9))
	if err != nil {
		t.Fatalf("MACD_MACD error: %v", err)
	}
	assertSeriesEqual(t, v, "MACD_MACD", indicator.MACD(bars, 12, 26, 9).Line3)
}

func TestComplexMACDDefaults(t *testing.T) {
	bars := complexBars()
	v, err := callComplex(t, "MACD")
	if err != nil {
		t.Fatalf("MACD() error: %v", err)
	}
	assertSeriesEqual(t, v, "MACD()", indicator.MACD(bars, 12, 26, 9).Values)
}

func TestComplexKDJ(t *testing.T) {
	bars := complexBars()
	k, err := callComplex(t, "KDJ", num(9), num(3), num(3))
	if err != nil {
		t.Fatalf("KDJ error: %v", err)
	}
	res := indicator.KDJ(bars, 9, 3, 3)
	assertSeriesEqual(t, k, "KDJ", res.Values)

	d, err := callComplex(t, "KDJ_D", num(9), num(3), num(3))
	if err != nil {
		t.Fatalf("KDJ_D error: %v", err)
	}
	assertSeriesEqual(t, d, "KDJ_D", res.Line2)

	j, err := callComplex(t, "KDJ_J", num(9), num(3), num(3))
	if err != nil {
		t.Fatalf("KDJ_J error: %v", err)
	}
	assertSeriesEqual(t, j, "KDJ_J", res.Line3)
}

func TestComplexBOLL(t *testing.T) {
	bars := complexBars()
	mid, err := callComplex(t, "BOLL", num(20), num(2.0))
	if err != nil {
		t.Fatalf("BOLL error: %v", err)
	}
	res := indicator.BOLL(bars, 20, 2.0)
	assertSeriesEqual(t, mid, "BOLL", res.Values)

	upper, err := callComplex(t, "BOLL_UPPER", num(20), num(2.0))
	if err != nil {
		t.Fatalf("BOLL_UPPER error: %v", err)
	}
	assertSeriesEqual(t, upper, "BOLL_UPPER", res.Line2)

	lower, err := callComplex(t, "BOLL_LOWER", num(20), num(2.0))
	if err != nil {
		t.Fatalf("BOLL_LOWER error: %v", err)
	}
	assertSeriesEqual(t, lower, "BOLL_LOWER", res.Line3)
}

func TestComplexDMI(t *testing.T) {
	bars := complexBars()
	res := indicator.DMI(bars, 14, 6)

	pdi, err := callComplex(t, "DMI", num(14), num(6))
	if err != nil {
		t.Fatalf("DMI error: %v", err)
	}
	assertSeriesEqual(t, pdi, "DMI", res.Values)

	mdi, err := callComplex(t, "DMI_MDI", num(14), num(6))
	if err != nil {
		t.Fatalf("DMI_MDI error: %v", err)
	}
	assertSeriesEqual(t, mdi, "DMI_MDI", res.Line2)

	adx, err := callComplex(t, "DMI_ADX", num(14), num(6))
	if err != nil {
		t.Fatalf("DMI_ADX error: %v", err)
	}
	assertSeriesEqual(t, adx, "DMI_ADX", res.Line3)

	adxr, err := callComplex(t, "DMI_ADXR", num(14), num(6))
	if err != nil {
		t.Fatalf("DMI_ADXR error: %v", err)
	}
	assertSeriesEqual(t, adxr, "DMI_ADXR", res.Data["ADXR"])
}

func TestComplexSingleOutput(t *testing.T) {
	bars := complexBars()
	cases := []struct {
		name string
		args []*Value
		want []float64
	}{
		{"RSI", []*Value{num(14)}, indicator.RSI(bars, 14).Values},
		{"CCI", []*Value{num(14)}, indicator.CCI(bars, 14).Values},
		{"WR", []*Value{num(14)}, indicator.WR(bars, 14).Values},
		{"BIAS", []*Value{num(6)}, indicator.BIAS(bars, 6).Values},
		{"VR", []*Value{num(26)}, indicator.VR(bars, 26).Values},
		{"MFI", []*Value{num(14)}, indicator.MFI(bars, 14).Values},
		{"EMV", []*Value{num(14)}, indicator.EMV(bars, 14).Values},
		{"ATR", []*Value{num(14)}, indicator.ATR(bars, 14).Values},
		{"ROC", []*Value{num(12)}, indicator.ROC(bars, 12).Values},
		{"PSY", []*Value{num(12)}, indicator.PSY(bars, 12).Values},
		{"EXPMA", []*Value{num(12)}, indicator.EXPMA(bars, 12).Values},
		{"MTM", []*Value{num(12)}, indicator.MTM(bars, 12).Values},
		{"DPO", []*Value{num(20)}, indicator.DPO(bars, 20).Values},
		{"OBV", nil, indicator.OBV(bars).Values},
		{"VWAP", nil, indicator.VWAP(bars).Values},
		{"ASI", []*Value{num(26)}, indicator.ASI(bars, 26).Values},
		{"AROON", []*Value{num(25)}, indicator.AROON(bars, 25).Values},
		{"TRIX", []*Value{num(12), num(9)}, indicator.TRIX(bars, 12, 9).Values},
		{"BRAR", []*Value{num(26)}, indicator.BRAR(bars, 26).Values},
	}
	for _, c := range cases {
		v, err := callComplex(t, c.name, c.args...)
		if err != nil {
			t.Fatalf("%s error: %v", c.name, err)
		}
		assertSeriesEqual(t, v, c.name, c.want)
	}
}

func TestComplexBBIAndSAR(t *testing.T) {
	bars := complexBars()
	v, err := callComplex(t, "BBI", num(3), num(6), num(12), num(24))
	if err != nil {
		t.Fatalf("BBI error: %v", err)
	}
	assertSeriesEqual(t, v, "BBI", indicator.BBI(bars, 3, 6, 12, 24).Values)

	v, err = callComplex(t, "SAR", num(0.02), num(0.2))
	if err != nil {
		t.Fatalf("SAR error: %v", err)
	}
	assertSeriesEqual(t, v, "SAR", indicator.SAR(bars, 0.02, 0.2).Values)
}

func TestComplexBadArgs(t *testing.T) {
	_, err := callComplex(t, "MACD", seq(1, 2), num(3), num(4))
	if err == nil {
		t.Fatal("MACD with array arg should error")
	}
	_, err = callComplex(t, "KDJ", num(0), num(3), num(3))
	if err == nil {
		t.Fatal("KDJ with period 0 should error")
	}
	_, err = callComplex(t, "ATR", num(-5))
	if err == nil {
		t.Fatal("ATR with negative period should error")
	}
}

func TestComplexRegistryCount(t *testing.T) {
	reg := newRegistry()
	names := reg.Names()
	if len(names) < 100 {
		t.Fatalf("expected at least 100 built-in functions, got %d", len(names))
	}
	required := []string{"MACD", "KDJ", "BOLL", "DMI", "RSI", "CCI", "SAR", "VWAP"}
	for _, r := range required {
		if _, ok := reg.Lookup(r); !ok {
			t.Errorf("function %s missing", r)
		}
	}
}

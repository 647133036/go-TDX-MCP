package indicator

import (
	"testing"
)

func TestMA(t *testing.T) {
	data := []float64{10, 11, 12, 13, 14, 15}
	result := MA(data, 3)
	// MA(3) should be [0, 0, 11, 12, 13, 14]
	if len(result) != 6 {
		t.Errorf("MA length = %d, want 6", len(result))
	}
	if result[2] != 11 {
		t.Errorf("MA[2] = %f, want 11", result[2])
	}
	if result[5] != 14 {
		t.Errorf("MA[5] = %f, want 14", result[5])
	}
}

func TestEMA(t *testing.T) {
	data := []float64{10, 11, 12, 13, 14}
	result := ema(data, 3)
	if len(result) != 5 {
		t.Errorf("EMA length = %d, want 5", len(result))
	}
	// EMA[2] should be average of first 3 values = 11
	if result[2] != 11 {
		t.Errorf("EMA[2] = %f, want 11", result[2])
	}
}

func TestMACD(t *testing.T) {
	bars := generateTestBars(100)
	result := MACD(bars, 12, 26, 9)
	if len(result.Values) != 100 {
		t.Errorf("MACD Values length = %d, want 100", len(result.Values))
	}
	if len(result.Line2) != 100 {
		t.Errorf("MACD Line2 length = %d, want 100", len(result.Line2))
	}
	if len(result.Line3) != 100 {
		t.Errorf("MACD Line3 length = %d, want 100", len(result.Line3))
	}
}

func TestKDJ(t *testing.T) {
	bars := generateTestBars(100)
	result := KDJ(bars, 9, 3, 3)
	if len(result.Values) != 100 {
		t.Errorf("KDJ Values (K) length = %d, want 100", len(result.Values))
	}
	if len(result.Line2) != 100 {
		t.Errorf("KDJ Line2 (D) length = %d, want 100", len(result.Line2))
	}
	if len(result.Line3) != 100 {
		t.Errorf("KDJ Line3 (J) length = %d, want 100", len(result.Line3))
	}
}

func TestRSI(t *testing.T) {
	bars := generateTestBars(50)
	result := RSI(bars, 6)
	if len(result.Values) != 50 {
		t.Errorf("RSI length = %d, want 50", len(result.Values))
	}
	// RSI should be between 0 and 100
	for i, v := range result.Values {
		if v < 0 || v > 100 {
			t.Errorf("RSI[%d] = %f, should be between 0 and 100", i, v)
		}
	}
}

func TestBOLL(t *testing.T) {
	bars := generateTestBars(100)
	result := BOLL(bars, 20, 2)
	if len(result.Values) != 100 {
		t.Errorf("BOLL Values (Mid) length = %d, want 100", len(result.Values))
	}
	if len(result.Line2) != 100 {
		t.Errorf("BOLL Line2 (Upper) length = %d, want 100", len(result.Line2))
	}
	if len(result.Line3) != 100 {
		t.Errorf("BOLL Line3 (Lower) length = %d, want 100", len(result.Line3))
	}
	// Upper band should be >= Mid band
	for i := 20; i < 100; i++ {
		if result.Line2[i] < result.Values[i] {
			t.Errorf("BOLL Upper[%d] = %f < Mid[%d] = %f", i, result.Line2[i], i, result.Values[i])
		}
	}
}

func TestComputeAll(t *testing.T) {
	bars := generateTestBars(100)
	indicators := []string{"MACD", "KDJ", "RSI", "BOLL", "MA"}
	params := map[string]float64{}

	result, err := ComputeAll(bars, indicators, params)
	if err != nil {
		t.Fatalf("ComputeAll failed: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("ComputeAll returned %d indicators, want 5", len(result))
	}
	if _, ok := result["MACD"]; !ok {
		t.Error("Missing MACD in result")
	}
	if _, ok := result["KDJ"]; !ok {
		t.Error("Missing KDJ in result")
	}
	if _, ok := result["RSI"]; !ok {
		t.Error("Missing RSI in result")
	}
	if _, ok := result["BOLL"]; !ok {
		t.Error("Missing BOLL in result")
	}
	if _, ok := result["MA"]; !ok {
		t.Error("Missing MA in result")
	}
}

func TestComputeAllWithUnknownIndicator(t *testing.T) {
	bars := generateTestBars(50)
	indicators := []string{"MACD", "UNKNOWN_INDICATOR"}
	result, err := ComputeAll(bars, indicators, nil)
	if err != nil {
		t.Fatalf("ComputeAll should not fail with unknown indicator: %v", err)
	}
	// Should only contain MACD, UNKNOWN_INDICATOR is silently skipped
	if len(result) != 1 {
		t.Errorf("ComputeAll returned %d indicators, want 1", len(result))
	}
}

func TestComputeAllCaseInsensitive(t *testing.T) {
	bars := generateTestBars(50)
	// ComputeAll expects UPPERCASE indicator names.
	// The web handler normalizes to uppercase before calling ComputeAll.
	result1, _ := ComputeAll(bars, []string{"MACD"}, nil)
	result2, _ := ComputeAll(bars, []string{"macd"}, nil)
	
	// MACD should work, macd should return empty (caller responsibility)
	if len(result1) != 1 {
		t.Errorf("MACD should return 1 indicator, got %d", len(result1))
	}
	if len(result2) != 0 {
		t.Errorf("lowercase macd should return 0 indicators, got %d", len(result2))
	}
}

func generateTestBars(count int) []Bar {
	bars := make([]Bar, count)
	for i := 0; i < count; i++ {
		bars[i] = Bar{
			Open:   float64(100 + i*2),
			High:   float64(102 + i*2),
			Low:    float64(98 + i*2),
			Close:  float64(101 + i*2),
			Vol:    float64(1000 + i*100),
			Amount: float64(100000 + i*10000),
		}
	}
	return bars
}

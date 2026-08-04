package web

import (
	"testing"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// ListItem.Item field order: Data(0), Second(1), Open(2), High(3), Low(4),
// Close(5), Amount(6), VolInStock(7), Volume(8).

func TestExtractBarsFromListItem(t *testing.T) {
	items := []struct{ Item []interface{} }{
		{Item: []interface{}{"2024-01-02", 0, 10.0, 11.0, 9.5, 10.5, 100000.0, 0, 50000.0}},
		{Item: []interface{}{"2024-01-03", 0, 10.5, 12.0, 10.0, 11.5, 120000.0, 0, 60000.0}},
	}
	bars, err := extractBarsFromListItem(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	checks := []struct {
		bar    indicator.Bar
		open   float64
		high   float64
		low    float64
		close  float64
		amount float64
		vol    float64
	}{
		{bars[0], 10.0, 11.0, 9.5, 10.5, 100000.0, 50000.0},
		{bars[1], 10.5, 12.0, 10.0, 11.5, 120000.0, 60000.0},
	}
	for i, c := range checks {
		if c.bar.Open != c.open || c.bar.High != c.high || c.bar.Low != c.low ||
			c.bar.Close != c.close || c.bar.Amount != c.amount || c.bar.Vol != c.vol {
			t.Errorf("bar %d = %+v, want open=%v high=%v low=%v close=%v amount=%v vol=%v",
				i, c.bar, c.open, c.high, c.low, c.close, c.amount, c.vol)
		}
	}
}

func TestExtractBarsFromListItemShortRow(t *testing.T) {
	items := []struct{ Item []interface{} }{
		{Item: []interface{}{"2024-01-02", 0, 10.0, 11.0, 9.5}}, // only 5 fields
	}
	bars, err := extractBarsFromListItem(items)
	if err == nil {
		t.Fatalf("expected error for short rows, got %v bars", bars)
	}
}

func TestExtractKlinesFromListItem(t *testing.T) {
	items := []struct{ Item []interface{} }{
		{Item: []interface{}{"20240102", 0, 10.0, 11.0, 9.5, 10.5, 100000.0, 0, 50000.0}},
	}
	klines, err := extractKlinesFromListItem(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(klines) != 1 {
		t.Fatalf("expected 1 kline, got %d", len(klines))
	}
	k := klines[0]
	if k.Date != "20240102" || k.Open != 10.0 || k.High != 11.0 || k.Low != 9.5 ||
		k.Close != 10.5 || k.Amount != 100000.0 || k.Vol != 50000.0 {
		t.Errorf("kline = %+v", k)
	}
}

func TestParseChanlunKlinesArrItems(t *testing.T) {
	input := [][]interface{}{
		{"20240102", 0, 10.0, 11.0, 9.5, 10.5, 100000.0, 0, 50000.0},
	}
	klines, err := parseChanlunKlines(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(klines) != 1 {
		t.Fatalf("expected 1 kline, got %d", len(klines))
	}
	k := klines[0]
	if k.Open != 10.0 || k.High != 11.0 || k.Low != 9.5 || k.Close != 10.5 ||
		k.Amount != 100000.0 || k.Vol != 50000.0 {
		t.Errorf("kline = %+v", k)
	}
}

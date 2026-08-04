package tdx

import (
	"encoding/json"
	"testing"

	"github.com/tdx/go-tdx-mcp/offline"
)

// ListItem.Item field order: Data(0), Second(1), Open(2), High(3), Low(4),
// Close(5), Amount(6), VolInStock(7), Volume(8).

func listItemJSON() map[string]interface{} {
	return map[string]interface{}{
		"ListHead": map[string]interface{}{"ItemHead": []string{"Data", "Second", "Open", "High", "Low", "Close", "Amount", "VolInStock", "Volume"}},
		"ListItem": []interface{}{
			map[string]interface{}{"Item": []interface{}{"2024-01-02", 0, 10.0, 11.0, 9.5, 10.5, 100000.0, 0, 50000.0}},
			map[string]interface{}{"Item": []interface{}{"2024-01-03", 0, 10.5, 12.0, 10.0, 11.5, 120000.0, 0, 60000.0}},
		},
	}
}

func TestParseKlineBarsListItemMapping(t *testing.T) {
	bars, err := parseKlineBars(listItemJSON())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	if bars[0].Open != 10.0 || bars[0].High != 11.0 || bars[0].Low != 9.5 ||
		bars[0].Close != 10.5 || bars[0].Amount != 100000.0 || bars[0].Vol != 50000.0 {
		t.Errorf("bar0 = %+v", bars[0])
	}
	if bars[1].Open != 10.5 || bars[1].High != 12.0 || bars[1].Low != 10.0 ||
		bars[1].Close != 11.5 || bars[1].Amount != 120000.0 || bars[1].Vol != 60000.0 {
		t.Errorf("bar1 = %+v", bars[1])
	}
}

func TestParseKlineBarsListItemShortRow(t *testing.T) {
	// 6..8 fields used to trigger index-out-of-range on fields[8].
	rows := [][]interface{}{
		{"2024-01-02", 0, 10.0, 11.0, 9.5, 10.5, 100000.0},
		{"2024-01-03", 0, 10.5, 12.0, 10.0, 11.5, 120000.0, 0},
	}
	for _, row := range rows {
		data := map[string]interface{}{
			"ListHead": map[string]interface{}{"ItemHead": []string{"Data", "Second", "Open", "High", "Low", "Close", "Amount", "VolInStock", "Volume"}},
			"ListItem": []interface{}{map[string]interface{}{"Item": row}},
		}
		bars, err := parseKlineBars(data)
		if err == nil && len(bars) > 0 {
			t.Fatalf("expected error or empty result for short row %v, got %v", row, bars)
		}
	}
}

func TestParseKlineBarsToDayBarsMapping(t *testing.T) {
	bars, err := parseKlineBarsToDayBars(listItemJSON())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	if bars[0] != (offline.DayBar{Date: "2024-01-02", Open: 10.0, High: 11.0, Low: 9.5, Close: 10.5, Vol: 50000.0, Amount: 100000.0}) {
		t.Errorf("bar0 = %+v", bars[0])
	}
}

func TestParseBarsFromResponseListItemMapping(t *testing.T) {
	raw, _ := json.Marshal(listItemJSON())
	resp := &TQLEXResponse{Data: json.RawMessage(raw)}
	bars, err := parseBarsFromResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	if bars[0].Open != 10.0 || bars[0].High != 11.0 || bars[0].Low != 9.5 ||
		bars[0].Close != 10.5 || bars[0].Amount != 100000.0 || bars[0].Vol != 50000.0 {
		t.Errorf("bar0 = %+v", bars[0])
	}
}

func TestParseBarsFromResponseShortRow(t *testing.T) {
	data := map[string]interface{}{
		"ListHead": map[string]interface{}{"ItemHead": []string{"Data", "Second", "Open", "High", "Low", "Close", "Amount", "VolInStock", "Volume"}},
		"ListItem": []interface{}{
			map[string]interface{}{"Item": []interface{}{"2024-01-02", 0, 10.0, 11.0, 9.5, 10.5, 100000.0}},
		},
	}
	raw, _ := json.Marshal(data)
	resp := &TQLEXResponse{Data: json.RawMessage(raw)}
	bars, err := parseBarsFromResponse(resp)
	if err == nil && len(bars) > 0 {
		t.Fatalf("expected error or empty result for short row, got %v", bars)
	}
}

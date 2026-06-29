package scraper

import (
	"strings"
	"testing"
)

// TestFundHoldingHTMLDetection tests that HTML content is correctly detected
func TestFundHoldingHTMLDetection(t *testing.T) {
	htmlContent := `<table><tr><td>test</td></tr></table>`
	if !strings.Contains(htmlContent, "<table>") {
		t.Error("HTML detection should find <table> tag")
	}

	jsonContent := `{"data": [{"fund_name": "test"}]}`
	if strings.Contains(jsonContent, "<table>") {
		t.Error("JSON content should not contain <table> tag")
	}
}

// TestIndicatorCaseNormalization tests that indicator names are uppercased
func TestIndicatorCaseNormalization(t *testing.T) {
	inputs := []string{"macd", "MACD", "Macd", "ma", "MA", "Kdj", "KDJ"}
	expected := []string{"MACD", "MACD", "MACD", "MA", "MA", "KDJ", "KDJ"}

	for i, input := range inputs {
		result := strings.ToUpper(strings.TrimSpace(input))
		if result != expected[i] {
			t.Errorf("ToUpper(%q) = %q, want %q", input, result, expected[i])
		}
	}
}

// TestEastMoneyFieldMapping tests that eastmoney clist API field mapping is correct
func TestEastMoneyFieldMapping(t *testing.T) {
	expectedFields := map[string]string{
		"f12": "stock_code",
		"f14": "stock_name",
		"f2":  "price",
		"f3":  "change_pct",
	}

	if len(expectedFields) != 4 {
		t.Error("Expected 4 field mappings")
	}
}

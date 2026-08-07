package scraper

import (
	"strings"
	"testing"
)

const testCurrenciesHTML = `
<html>
<body>
<table>
<tr id="pair_1">
<td>EUR/USD</td>
<td class="pid-1-last">1.1523</td>
<td class="pid-1-high">1.1527</td>
<td class="pid-1-low">1.1519</td>
<td class="pid-1-pc">-0.0001</td>
<td class="pid-1-pcp">-0.01</td>
<td>07:39:45</td>
</tr>
<tr id="pair_2">
<td>GBP/USD</td>
<td class="pid-2-last">1.3453</td>
<td class="pid-2-high">1.3460</td>
<td class="pid-2-low">1.3446</td>
<td class="pid-2-pc">-0.0005</td>
<td class="pid-2-pcp">-0.04</td>
<td>07:39:29</td>
</tr>
<tr id="pair_3">
<td>USD/JPY</td>
<td class="pid-3-last">158.42</td>
<td class="pid-3-high">158.51</td>
<td class="pid-3-low">158.38</td>
<td class="pid-3-pc">-0.01</td>
<td class="pid-3-pcp">-0.01</td>
<td>07:39:11</td>
</tr>
<tr id="pair_2111">
<td>USD/CNY</td>
<td class="pid-2111-last">6.7491</td>
<td class="pid-2111-high">6.7491</td>
<td class="pid-2111-low">6.7491</td>
<td class="pid-2111-pc">0.0000</td>
<td class="pid-2111-pcp">0.00</td>
<td>07:38:53</td>
</tr>
<tr id="pair_1">
<td>-0.01%</td>
<td>1.1523</td>
</tr>
<tr id="pair_2">
<td>-0.04%</td>
<td>1.3453</td>
</tr>
</table>
</body>
</html>
`

func TestParseCurrenciesHTML(t *testing.T) {
	rates := parseCurrenciesHTML(testCurrenciesHTML)

	if len(rates) != 4 {
		t.Fatalf("expected 4 rates, got %d", len(rates))
	}

	tests := []struct {
		name         string
		expectedLast float64
		expectedHigh float64
		expectedLow  float64
	}{
		{"EUR/USD", 1.1523, 1.1527, 1.1519},
		{"GBP/USD", 1.3453, 1.3460, 1.3446},
		{"USD/JPY", 158.42, 158.51, 158.38},
		{"USD/CNY", 6.7491, 6.7491, 6.7491},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var found *struct{}
			for i, r := range rates {
				if strings.Contains(r.Name, tt.name) {
					found = &struct{}{}
					_ = found
					if r.Last != tt.expectedLast {
						t.Errorf("%s Last = %v, want %v", tt.name, r.Last, tt.expectedLast)
					}
					if r.High != tt.expectedHigh {
						t.Errorf("%s High = %v, want %v", tt.name, r.High, tt.expectedHigh)
					}
					if r.Low != tt.expectedLow {
						t.Errorf("%s Low = %v, want %v", tt.name, r.Low, tt.expectedLow)
					}
					_ = i
				}
			}
		})
	}
}

func TestNormalizePrice(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2345", "1.2345"},
		{"+1.2345", "1.2345"},
		{"  -0.0001  ", "-0.0001"},
		{"0.0000", "0.0000"},
	}

	for _, tt := range tests {
		result := normalizePrice(tt.input)
		if result != tt.expected {
			t.Errorf("normalizePrice(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseTDCells(t *testing.T) {
	tds := [][]string{
		{"EUR/USD", "EUR/USD"},
		{"1.1523", "1.1523"},
		{"1.1527", "1.1527"},
		{"1.1519", "1.1519"},
		{"-0.0001", "-0.0001"},
		{"-0.01%", "-0.01%"},
		{"07:39:45", "07:39:45"},
	}

	rate := parseTDCells(tds)

	if rate.Name != "EUR/USD" {
		t.Errorf("Name = %q, want %q", rate.Name, "EUR/USD")
	}
	if rate.Last != 1.1523 {
		t.Errorf("Last = %v, want %v", rate.Last, 1.1523)
	}
	if rate.High != 1.1527 {
		t.Errorf("High = %v, want %v", rate.High, 1.1527)
	}
	if rate.Low != 1.1519 {
		t.Errorf("Low = %v, want %v", rate.Low, 1.1519)
	}
	if rate.Change != -0.0001 {
		t.Errorf("Change = %v, want %v", rate.Change, -0.0001)
	}
	if rate.ChangePct != -0.01 {
		t.Errorf("ChangePct = %v, want %v", rate.ChangePct, -0.01)
	}
	if rate.LastUpdate != "07:39:45" {
		t.Errorf("LastUpdate = %q, want %q", rate.LastUpdate, "07:39:45")
	}
}

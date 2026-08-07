package scraper

import (
	"strings"
	"testing"
)

const testFundsHTML = `
<html>
<body>
<tr><td class="flag"><span class="ceFlags China">&nbsp;</span></td><td class="bold left noWrap elp plusIconTd"><a href="/funds/cmf-csi-white-spirit-index" title="招商中证白酒指数证券投资基金A"><span data-name="招商中证白酒指数证券投资基金A" data-id="1092039">招商中证白酒指数分级</span></a></td><td>161725</td><td>0.561</td><td>-0.50%</td><td>31.3B</td><td>06/08</td></tr>
<tr><td class="flag"><span class="ceFlags USA">&nbsp;</span></td><td class="bold left noWrap elp plusIconTd"><a href="/funds/fidelity-series-commodity-strategy" title="Fidelity Series Commodity Strategy Fund"><span data-name="Fidelity Series Commodity Strategy Fund" data-id="1002228">Fidelity Series Commodity Strategy</span></a></td><td>FCSSX</td><td>112.38</td><td>+0.66%</td><td>1.25B</td><td>05/08</td></tr>
</body>
</html>
`

func TestParseFundsHTML(t *testing.T) {
	quotes := parseFundsHTML(testFundsHTML)

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	t.Run("China fund", func(t *testing.T) {
		for _, q := range quotes {
			if strings.Contains(q.Name, "白酒") {
				if q.ID != 1092039 {
					t.Errorf("ID = %d, want %d", q.ID, 1092039)
				}
				if q.Symbol != "161725" {
					t.Errorf("Symbol = %q, want %q", q.Symbol, "161725")
				}
				if q.Last != 0.561 {
					t.Errorf("Last = %v, want %v", q.Last, 0.561)
				}
				if q.ChangePct != -0.50 {
					t.Errorf("ChangePct = %v, want %v", q.ChangePct, -0.50)
				}
				if q.Volume != "31.3B" {
					t.Errorf("Volume = %q, want %q", q.Volume, "31.3B")
				}
				if q.Date != "06/08" {
					t.Errorf("Date = %q, want %q", q.Date, "06/08")
				}
			}
		}
	})

	t.Run("US fund", func(t *testing.T) {
		for _, q := range quotes {
			if strings.Contains(q.Name, "Fidelity") {
				if q.ID != 1002228 {
					t.Errorf("ID = %d, want %d", q.ID, 1002228)
				}
				if q.Symbol != "FCSSX" {
					t.Errorf("Symbol = %q, want %q", q.Symbol, "FCSSX")
				}
				if q.Last != 112.38 {
					t.Errorf("Last = %v, want %v", q.Last, 112.38)
				}
				if q.ChangePct != 0.66 {
					t.Errorf("ChangePct = %v, want %v", q.ChangePct, 0.66)
				}
				if q.Volume != "1.25B" {
					t.Errorf("Volume = %q, want %q", q.Volume, "1.25B")
				}
			}
		}
	})
}

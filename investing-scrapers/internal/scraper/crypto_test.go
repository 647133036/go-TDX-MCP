package scraper

import (
	"testing"
)

const testCryptoHTML = `
<html>
<head>
<script id="__NEXT_DATA__" type="application/json">
{"props":{"pageProps":{"state":{"multiAssetsCollectionStore":{"multiAssetsCollections":{
  "most-active-crypto-pairs-table":{
    "_collection":[
      {"id":"1","name":"Bitcoin","symbol":"BTC","flagCode":"BTC","title":"Bitcoin","url":"/crypto/bitcoin","last":64303.70,"high":64500,"low":63800,"changeOneDay":300.50,"changeOneDayPercent":0.47,"isCFD":false,"lastUpdateTime":"2026-08-06T23:40:25.000Z"},
      {"id":"2","name":"Ethereum","symbol":"ETH","flagCode":"ETH","title":"Ethereum","url":"/crypto/ethereum","last":2456.80,"high":2480,"low":2420,"changeOneDay":-20.30,"changeOneDayPercent":-0.82,"isCFD":false,"lastUpdateTime":"2026-08-06T23:40:25.000Z"}
    ]
  }
}}}}}}
</script>
</head>
<body></body>
</html>
`

func TestParseCryptoHTML(t *testing.T) {
	quotes := parseCryptoHTML(testCryptoHTML)

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	tests := []struct {
		name      string
		symbol    string
		last      float64
		changePct float64
	}{
		{"Bitcoin", "BTC", 64303.70, 0.47},
		{"Ethereum", "ETH", 2456.80, -0.82},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, q := range quotes {
				if q.Name == tt.name {
					if q.Symbol != tt.symbol {
						t.Errorf("%s Symbol = %q, want %q", tt.name, q.Symbol, tt.symbol)
					}
					if q.Last != tt.last {
						t.Errorf("%s Last = %v, want %v", tt.name, q.Last, tt.last)
					}
					if q.ChangePct != tt.changePct {
						t.Errorf("%s ChangePct = %v, want %v", tt.name, q.ChangePct, tt.changePct)
					}
				}
			}
		})
	}
}

func TestParseCryptoHTML_NoNextData(t *testing.T) {
	quotes := parseCryptoHTML("<html><body>No data</body></html>")
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

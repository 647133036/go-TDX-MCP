package scraper

import (
	"testing"
)

const testIndicesHTML = `
<html>
<head>
<script id="__NEXT_DATA__" type="application/json">
{"props":{"pageProps":{"state":{"multiAssetsCollectionStore":{"multiAssetsCollections":{
  "indices_performance_table":{
    "_collection":[
      {"id":"40820","name":"上证指数","symbol":"SHCOMP","flagCode":"CN","flagName":"中国","title":"上证指数","url":"/indices/shanghai-composite","last":3245.67,"high":3256.78,"low":3234.56,"changeOneDay":12.34,"changeOneDayPercent":0.38,"lastUpdateTime":"2026-08-06T06:00:00.000Z"},
      {"id":"23","name":"标普500","symbol":"SPX","flagCode":"US","flagName":"美国","title":"标普500指数","url":"/indices/us-spx-500","last":5678.90,"high":5690.12,"low":5667.89,"changeOneDay":23.45,"changeOneDayPercent":0.41,"lastUpdateTime":"2026-08-06T21:00:00.000Z"},
      {"id":"166","name":"纳斯达克综合指数","symbol":"IXIC","flagCode":"US","flagName":"美国","title":"纳斯达克综合指数","url":"/indices/nasdaq-composite","last":18234.56,"high":18345.67,"low":18123.45,"changeOneDay":123.45,"changeOneDayPercent":0.68,"lastUpdateTime":"2026-08-06T21:00:00.000Z"}
    ]
  }
}}}}}}
</script>
</head>
<body></body>
</html>
`

func TestParseIndicesHTML(t *testing.T) {
	quotes := parseIndicesHTML(testIndicesHTML)

	if len(quotes) != 3 {
		t.Fatalf("expected 3 quotes, got %d", len(quotes))
	}

	tests := []struct {
		name     string
		symbol   string
		expectedLast float64
		expectedHigh float64
		expectedLow  float64
	}{
		{"上证指数", "SHCOMP", 3245.67, 3256.78, 3234.56},
		{"标普500", "SPX", 5678.90, 5690.12, 5667.89},
		{"纳斯达克综合指数", "IXIC", 18234.56, 18345.67, 18123.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, q := range quotes {
				if q.Name == tt.name {
					if q.Last != tt.expectedLast {
						t.Errorf("%s Last = %v, want %v", tt.name, q.Last, tt.expectedLast)
					}
					if q.High != tt.expectedHigh {
						t.Errorf("%s High = %v, want %v", tt.name, q.High, tt.expectedHigh)
					}
					if q.Low != tt.expectedLow {
						t.Errorf("%s Low = %v, want %v", tt.name, q.Low, tt.expectedLow)
					}
					if q.Symbol != tt.symbol {
						t.Errorf("%s Symbol = %q, want %q", tt.name, q.Symbol, tt.symbol)
					}
				}
			}
		})
	}
}

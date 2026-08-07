package scraper

import (
	"testing"
)

const testCommoditiesHTML = `
<html>
<head>
<script id="__NEXT_DATA__" type="application/json">
{"props":{"pageProps":{"state":{"multiAssetsCollectionStore":{"multiAssetsCollections":{
  "commodities_energy":{
    "_collection":[
      {"id":"8833","name":"伦敦布伦特原油","symbol":"LCO","flagCode":"GB","flagName":"英国","title":"伦敦布伦特原油","url":"/commodities/brent-oil","month":"Oct 26","last":83.25,"high":83.57,"low":83.0,"changeOneDay":-0.29,"changeOneDayPercent":-0.35,"isCFD":true,"lastUpdateTime":"2026-08-06T23:40:25.000Z"}
    ]
  },
  "commodities_metals":{
    "_collection":[
      {"id":"8830","name":"黄金","symbol":"GC","flagCode":"gold","flagName":"","title":"黄金","url":"/commodities/gold","month":"Dec 26","last":4305.0,"high":4307.15,"low":4293.9,"changeOneDay":5.4,"changeOneDayPercent":0.13,"isCFD":true,"lastUpdateTime":"2026-08-06T23:40:25.000Z"},
      {"id":"8836","name":"白银","symbol":"SI","flagCode":"silver","flagName":"","title":"白银","url":"/commodities/silver","month":"Dec 26","last":61.88,"high":61.903,"low":61.606,"changeOneDay":0.27,"changeOneDayPercent":0.44,"isCFD":true,"lastUpdateTime":"2026-08-06T23:40:25.000Z"}
    ]
  },
  "commodities_agriculture":{
    "_collection":[
      {"id":"8832","name":"美国C型咖啡","symbol":"KC","flagCode":"US","flagName":"美国","title":"美国C型咖啡","url":"/commodities/us-coffee-c","month":"Dec 26","last":308.25,"high":311.45,"low":303.6,"changeOneDay":-3.2,"changeOneDayPercent":-1.03,"isCFD":true,"lastUpdateTime":"2026-08-06T23:40:25.000Z"}
    ]
  },
  "commodities_indices":{
    "_collection":[
      {"id":"39972","name":"TR/CC CRB","symbol":"TRCCRB","flagCode":"WW","flagName":"","title":"TR/CC CRB","url":"/indices/thomson-reuters-cr-b","last":374.49,"high":374.49,"low":374.49,"changeOneDay":0,"changeOneDayPercent":0}
    ]
  }
}}}}}}
</script>
</head>
<body></body>
</html>
`

func TestParseCommoditiesHTML(t *testing.T) {
	quotes := parseCommoditiesHTML(testCommoditiesHTML)

	if len(quotes) != 5 {
		t.Fatalf("expected 5 quotes, got %d", len(quotes))
	}

	// Verify each quote
	for _, q := range quotes {
		switch q.Name {
		case "伦敦布伦特原油":
			if q.Last != 83.25 {
				t.Errorf("布伦特原油 Last = %v, want %v", q.Last, 83.25)
			}
			if q.Symbol != "LCO" {
				t.Errorf("布伦特原油 Symbol = %q, want %q", q.Symbol, "LCO")
			}
			if q.Collection != "commodities_energy" {
				t.Errorf("Collection = %q, want %q", q.Collection, "commodities_energy")
			}
		case "黄金":
			if q.Last != 4305.0 {
				t.Errorf("黄金 Last = %v, want %v", q.Last, 4305.0)
			}
			if q.Symbol != "GC" {
				t.Errorf("黄金 Symbol = %q, want %q", q.Symbol, "GC")
			}
			if q.High != 4307.15 {
				t.Errorf("黄金 High = %v, want %v", q.High, 4307.15)
			}
			if q.Low != 4293.9 {
				t.Errorf("黄金 Low = %v, want %v", q.Low, 4293.9)
			}
			if q.ChangePct != 0.13 {
				t.Errorf("黄金 ChangePct = %v, want %v", q.ChangePct, 0.13)
			}
		case "白银":
			if q.Last != 61.88 {
				t.Errorf("白银 Last = %v, want %v", q.Last, 61.88)
			}
			if q.Collection != "commodities_metals" {
				t.Errorf("Collection = %q, want %q", q.Collection, "commodities_metals")
			}
		case "美国C型咖啡":
			if q.Last != 308.25 {
				t.Errorf("咖啡 Last = %v, want %v", q.Last, 308.25)
			}
		case "TR/CC CRB":
			if q.Last != 374.49 {
				t.Errorf("CRB Last = %v, want %v", q.Last, 374.49)
			}
			if q.Collection != "commodities_indices" {
				t.Errorf("Collection = %q, want %q", q.Collection, "commodities_indices")
			}
		default:
			t.Errorf("Unexpected commodity: %s", q.Name)
		}
	}
}

func TestParseCommoditiesHTML_NoNextData(t *testing.T) {
	quotes := parseCommoditiesHTML("<html><body>No data</body></html>")
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

func TestParseCommodityItem_Fields(t *testing.T) {
	item := map[string]interface{}{
		"id":                  "8830",
		"name":                "黄金",
		"symbol":              "GC",
		"last":                4305.0,
		"high":                4307.15,
		"low":                 4293.9,
		"changeOneDay":        5.4,
		"changeOneDayPercent": 0.13,
		"month":               "Dec 26",
		"url":                 "/commodities/gold",
		"flagCode":            "gold",
		"isCFD":               true,
	}

	q := parseCommodityItem(item, "commodities_metals")

	if q.ID != 8830 {
		t.Errorf("ID = %d, want %d", q.ID, 8830)
	}
	if q.Name != "黄金" {
		t.Errorf("Name = %q, want %q", q.Name, "黄金")
	}
	if q.Symbol != "GC" {
		t.Errorf("Symbol = %q, want %q", q.Symbol, "GC")
	}
	if q.Last != 4305.0 {
		t.Errorf("Last = %v, want %v", q.Last, 4305.0)
	}
	if q.High != 4307.15 {
		t.Errorf("High = %v, want %v", q.High, 4307.15)
	}
	if q.Low != 4293.9 {
		t.Errorf("Low = %v, want %v", q.Low, 4293.9)
	}
	if q.Change != 5.4 {
		t.Errorf("Change = %v, want %v", q.Change, 5.4)
	}
	if q.ChangePct != 0.13 {
		t.Errorf("ChangePct = %v, want %v", q.ChangePct, 0.13)
	}
	if q.Month != "Dec 26" {
		t.Errorf("Month = %q, want %q", q.Month, "Dec 26")
	}
	if q.URL != "/commodities/gold" {
		t.Errorf("URL = %q, want %q", q.URL, "/commodities/gold")
	}
	if q.FlagCode != "gold" {
		t.Errorf("FlagCode = %q, want %q", q.FlagCode, "gold")
	}
	if q.IsCFD != true {
		t.Error("IsCFD should be true")
	}
	if q.Collection != "commodities_metals" {
		t.Errorf("Collection = %q, want %q", q.Collection, "commodities_metals")
	}
}

func TestParseCommodityItem_StringLast(t *testing.T) {
	item := map[string]interface{}{
		"id":                  "8830",
		"name":                "黄金",
		"symbol":              "GC",
		"last":                "4299.60",
		"high":                "4307.15",
		"low":                 "4293.9",
		"changeOneDay":        "5.02",
		"changeOneDayPercent": "0.12",
	}

	q := parseCommodityItem(item, "commodities_metals")

	if q.Last != 4299.60 {
		t.Errorf("Last = %v, want %v", q.Last, 4299.60)
	}
	if q.High != 4307.15 {
		t.Errorf("High = %v, want %v", q.High, 4307.15)
	}
}

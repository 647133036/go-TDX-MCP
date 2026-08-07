package scraper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

var singleInstrumentRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type CommodityQuoteScraper struct {
	fetcher *fetcher.Fetcher
}

func NewCommodityQuoteScraper() *CommodityQuoteScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &CommodityQuoteScraper{fetcher: f}
}

func (s *CommodityQuoteScraper) FetchByName(slug string) (*models.CommodityQuoteDetail, error) {
	return fetchInstrumentPage(s.fetcher, "commodities", slug)
}

type IndexQuoteScraper struct {
	fetcher *fetcher.Fetcher
}

func NewIndexQuoteScraper() *IndexQuoteScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &IndexQuoteScraper{fetcher: f}
}

func (s *IndexQuoteScraper) FetchByName(slug string) (*models.CommodityQuoteDetail, error) {
	return fetchInstrumentPage(s.fetcher, "indices", slug)
}

func fetchInstrumentPage(f *fetcher.Fetcher, category, slug string) (*models.CommodityQuoteDetail, error) {
	url := fetcher.BaseURL + "/" + category + "/" + slug
	body, err := f.FetchPage(url)
	if err != nil {
		return nil, err
	}
	m := singleInstrumentRe.FindStringSubmatch(body)
	if m == nil {
		return nil, fmt.Errorf("no __NEXT_DATA__ on %s", url)
	}
	return parseSingleInstrumentPage(m[1])
}

func parseSingleInstrumentPage(jsonStr string) (*models.CommodityQuoteDetail, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	props, ok := data["props"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no props")
	}
	pageProps, ok := props["pageProps"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no pageProps")
	}
	state, ok := pageProps["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no state")
	}

	storeName := "commodityStore"
	if _, ok := state["indexStore"]; ok {
		storeName = "indexStore"
	}
	if _, ok := state["equityStore"]; ok {
		storeName = "equityStore"
	}

	store, ok := state[storeName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no %s", storeName)
	}
	instrument, ok := store["instrument"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no instrument")
	}

	return parseCommodityDetail(instrument, store, pageProps), nil
}

func parseCommodityDetail(instrument map[string]interface{}, store map[string]interface{}, pageProps map[string]interface{}) *models.CommodityQuoteDetail {
	detail := &models.CommodityQuoteDetail{}

	if base, ok := instrument["base"].(map[string]interface{}); ok {
		detail.ID = toInt(base["id"])
		detail.Type = getStr(base, "type")
		detail.URL = getStr(base, "path")
		detail.IsCFD = toBool(base["isCfd"])
		detail.IsOpen = toBool(base["isOpen"])
	}

	if n, ok := instrument["name"].(map[string]interface{}); ok {
		detail.Name = getStr(n, "fullName")
		detail.Symbol = getStr(n, "symbol")
	}
	if en, ok := instrument["englishName"].(map[string]interface{}); ok {
		if detail.Symbol == "" {
			detail.Symbol = getStr(en, "shortName")
		}
	}

	if iid, ok := store["instrumentId"]; ok {
		detail.InstrumentID = toInt(iid)
	}
	if curID, ok := store["currencyId"]; ok {
		detail.CurrencyID = toInt(curID)
	}

	if curr, ok := instrument["currency"].(map[string]interface{}); ok {
		detail.Currency = getStr(curr, "name")
	}

	if price, ok := instrument["price"].(map[string]interface{}); ok {
		detail.Last = getFloat(price, "last")
		detail.Open = getFloat(price, "open")
		detail.High = getFloat(price, "high")
		detail.Low = getFloat(price, "low")
		detail.LastClose = getFloat(price, "lastClose")
		detail.Change = getFloat(price, "change")
		detail.ChangePct = getFloat(price, "changePcr")
		detail.Volume = getFloat(price, "volume")
		detail.FiftyTwoWeekHigh = getFloat(price, "fiftyTwoWeekHigh")
		detail.FiftyTwoWeekLow = getFloat(price, "fiftyTwoWeekLow")
		detail.OneYearChange = getFloat(price, "oneYearChange")
		detail.IsDelayed = getFloat(price, "isDelayed") > 0
		if lu, ok := price["lastUpdateTime"].(string); ok {
			detail.LastUpdate = lu
		}
	}

	if bidding, ok := instrument["bidding"].(map[string]interface{}); ok {
		detail.Bid = getFloat(bidding, "bid")
		detail.Ask = getFloat(bidding, "ask")
	}

	if vol, ok := instrument["volume"].(map[string]interface{}); ok {
		detail.Turnover = getFloat(vol, "_turnover")
	}

	if fund, ok := instrument["fundamental"].(map[string]interface{}); ok {
		detail.OneYearReturn = getFloat(fund, "oneYearReturn")
	}

	if cdata, ok := instrument["commodityData"].(map[string]interface{}); ok {
		detail.Unit = getStr(cdata, "unit")
		detail.PairType = getStr(cdata, "pairTypeDefine")
	}

	if priceChanges, ok := store["priceChanges"].(map[string]interface{}); ok {
		detail.Pct1D = getFloat(priceChanges, "pct_1d")
		detail.Pct1W = getFloat(priceChanges, "pct_1w")
		detail.Pct1M = getFloat(priceChanges, "pct_1m")
		detail.Pct3M = getFloat(priceChanges, "pct_3m")
		detail.Pct6M = getFloat(priceChanges, "pct_6m")
		detail.Pct1Y = getFloat(priceChanges, "pct_1y")
		detail.PctYTD = getFloat(priceChanges, "pct_ytd")
		detail.PctAllTime = getFloat(priceChanges, "pct_all_time")
	}

	_ = pageProps
	return detail
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	}
	return 0
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1"
	case float64:
		return val == 1
	}
	return false
}

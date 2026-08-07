package scraper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

// CommodityScraper scrapes commodity prices from cn.investing.com.
type CommodityScraper struct {
	fetcher *fetcher.Fetcher
}

// NewCommodityScraper creates a new CommodityScraper.
func NewCommodityScraper() *CommodityScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &CommodityScraper{fetcher: f}
}

// FetchAll fetches all commodity prices from the commodities page.
func (s *CommodityScraper) FetchAll() ([]models.CommodityQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/commodities/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch commodities page: %w", err)
	}

	return parseCommoditiesHTML(body), nil
}

// FetchByName fetches a single commodity by searching its page.
func (s *CommodityScraper) FetchByName(name string) (*models.CommodityQuote, error) {
	// For individual commodities, we would need to navigate to their specific page
	// For now, fetch all and filter
	all, err := s.FetchAll()
	if err != nil {
		return nil, err
	}
	for i, q := range all {
		if q.Name == name || q.Symbol == name {
			return &all[i], nil
		}
	}
	return nil, fmt.Errorf("commodity not found: %s", name)
}

func parseCommoditiesHTML(html string) []models.CommodityQuote {
	m := nextDataRe.FindStringSubmatch(html)
	if m == nil {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(m[1]), &data); err != nil {
		return nil
	}

	props, ok := data["props"].(map[string]interface{})
	if !ok {
		return nil
	}

	pageProps, ok := props["pageProps"].(map[string]interface{})
	if !ok {
		return nil
	}

	state, ok := pageProps["state"].(map[string]interface{})
	if !ok {
		return nil
	}

	masc, ok := state["multiAssetsCollectionStore"].(map[string]interface{})
	if !ok {
		return nil
	}

	collections, ok := masc["multiAssetsCollections"].(map[string]interface{})
	if !ok {
		return nil
	}

	quotes := make([]models.CommodityQuote, 0, 30)

	// Only collect from sub-collections that have price data
	priceCollections := []string{
		"commodities_energy",
		"commodities_metals",
		"commodities_agriculture",
		"commodities_indices",
	}

	for _, colName := range priceCollections {
		col, ok := collections[colName].(map[string]interface{})
		if !ok {
			continue
		}

		items, ok := col["_collection"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range items {
			quote := parseCommodityItem(item.(map[string]interface{}), colName)
			quotes = append(quotes, quote)
		}
	}

	return quotes
}

func parseCommodityItem(item map[string]interface{}, collection string) models.CommodityQuote {
	q := models.CommodityQuote{
		Collection: collection,
	}

	if v, ok := item["id"]; ok {
		switch val := v.(type) {
		case float64:
			q.ID = int(val)
		case string:
			q.ID, _ = strconv.Atoi(val)
		}
	}

	if v, ok := item["name"]; ok {
		q.Name, _ = v.(string)
	}
	if v, ok := item["symbol"]; ok {
		q.Symbol, _ = v.(string)
	}
	if v, ok := item["flagCode"]; ok {
		q.FlagCode, _ = v.(string)
	}
	if v, ok := item["flagName"]; ok {
		q.FlagName, _ = v.(string)
	}
	if v, ok := item["month"]; ok {
		q.Month, _ = v.(string)
	}
	if v, ok := item["url"]; ok {
		q.URL, _ = v.(string)
	}

	if v, ok := item["last"]; ok {
		switch val := v.(type) {
		case float64:
			q.Last = val
		case string:
			q.Last, _ = strconv.ParseFloat(val, 64)
		}
	}
	if v, ok := item["high"]; ok {
		switch val := v.(type) {
		case float64:
			q.High = val
		case string:
			q.High, _ = strconv.ParseFloat(val, 64)
		}
	}
	if v, ok := item["low"]; ok {
		switch val := v.(type) {
		case float64:
			q.Low = val
		case string:
			q.Low, _ = strconv.ParseFloat(val, 64)
		}
	}
	if v, ok := item["changeOneDay"]; ok {
		switch val := v.(type) {
		case float64:
			q.Change = val
		case string:
			q.Change, _ = strconv.ParseFloat(val, 64)
		}
	}
	if v, ok := item["changeOneDayPercent"]; ok {
		switch val := v.(type) {
		case float64:
			q.ChangePct = val
		case string:
			q.ChangePct, _ = strconv.ParseFloat(val, 64)
		}
	}
	if v, ok := item["isCFD"]; ok {
		q.IsCFD, _ = v.(bool)
	}
	if v, ok := item["lastUpdateTime"]; ok {
		if t, err := time.Parse(time.RFC3339, v.(string)); err == nil {
			q.LastUpdateTime = t
		}
	}

	return q
}

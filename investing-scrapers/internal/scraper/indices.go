package scraper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

// IndexScraper scrapes stock index prices from cn.investing.com.
type IndexScraper struct {
	fetcher *fetcher.Fetcher
}

// NewIndexScraper creates a new IndexScraper.
func NewIndexScraper() *IndexScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &IndexScraper{fetcher: f}
}

// FetchAll fetches all index prices from the indices page.
func (s *IndexScraper) FetchAll() ([]models.IndexQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/indices/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch indices page: %w", err)
	}

	return parseIndicesHTML(body), nil
}

func parseIndicesHTML(html string) []models.IndexQuote {
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

	quotes := make([]models.IndexQuote, 0, 30)

	// Index collections have different names than commodities
	// Common patterns: "indices_xxx"
	for colName, col := range collections {
		colMap, ok := col.(map[string]interface{})
		if !ok {
			continue
		}

		items, ok := colMap["_collection"].([]interface{})
		if !ok {
			continue
		}

		// Skip collections that don't have price data
		sampleItem, ok := items[0].(map[string]interface{})
		if !ok {
			continue
		}
		_, hasLast := sampleItem["last"]

		for _, item := range items {
			quote := parseIndexItem(item.(map[string]interface{}), colName, hasLast)
			quotes = append(quotes, quote)
		}
	}

	return quotes
}

func parseIndexItem(item map[string]interface{}, collection string, hasLast bool) models.IndexQuote {
	q := models.IndexQuote{
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
	if v, ok := item["title"]; ok {
		if q.Name == "" {
			q.Name, _ = v.(string)
		}
	}
	if v, ok := item["url"]; ok {
		q.URL, _ = v.(string)
	}
	if v, ok := item["month"]; ok {
		q.Month, _ = v.(string)
	}
	if v, ok := item["flagCode"]; ok {
		q.FlagCode, _ = v.(string)
	}
	if v, ok := item["flagName"]; ok {
		q.FlagName, _ = v.(string)
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
	if v, ok := item["lastUpdateTime"]; ok {
		if t, err := time.Parse(time.RFC3339, v.(string)); err == nil {
			q.LastUpdateTime = t
		}
	}

	return q
}

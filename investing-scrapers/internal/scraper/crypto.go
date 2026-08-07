package scraper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

// CryptoScraper scrapes cryptocurrency quotes from the /crypto/ page.
type CryptoScraper struct {
	fetcher *fetcher.Fetcher
}

// NewCryptoScraper creates a new CryptoScraper.
func NewCryptoScraper() *CryptoScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &CryptoScraper{fetcher: f}
}

// FetchAll fetches all crypto quotes from the crypto page.
func (s *CryptoScraper) FetchAll() ([]models.CryptoQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/crypto/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch crypto page: %w", err)
	}
	return parseCryptoHTML(body), nil
}

func parseCryptoHTML(html string) []models.CryptoQuote {
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

	quotes := make([]models.CryptoQuote, 0, 30)

	for colName, col := range collections {
		colMap, ok := col.(map[string]interface{})
		if !ok {
			continue
		}

		items, ok := colMap["_collection"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range items {
			quote := parseCryptoItem(item.(map[string]interface{}), colName)
			quotes = append(quotes, quote)
		}
	}

	return quotes
}

func parseCryptoItem(item map[string]interface{}, collection string) models.CryptoQuote {
	q := models.CryptoQuote{
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
	if v, ok := item["flagCode"]; ok {
		q.FlagCode, _ = v.(string)
	}
	if v, ok := item["flagName"]; ok {
		q.FlagName, _ = v.(string)
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
	if v, ok := item["lastUpdateTime"]; ok {
		if t, err := time.Parse(time.RFC3339, v.(string)); err == nil {
			q.LastUpdateTime = t
		}
	}

	return q
}

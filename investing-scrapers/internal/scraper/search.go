package scraper

import (
	"encoding/json"
	"fmt"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

// SearchScraper queries the investing.com Search API.
type SearchScraper struct {
	fetcher *fetcher.Fetcher
}

// NewSearchScraper creates a new SearchScraper.
func NewSearchScraper() *SearchScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &SearchScraper{fetcher: f}
}

// Search calls the Search API and returns structured results.
func (s *SearchScraper) Search(query string) (*models.SearchResponse, error) {
	body, err := s.fetcher.FetchSearch(query)
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", query, err)
	}

	return parseSearchResponse(body), nil
}

func parseSearchResponse(body string) *models.SearchResponse {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return nil
	}

	resp := &models.SearchResponse{}

	if v, ok := raw["articles"]; ok {
		if arr, ok := v.([]interface{}); ok {
			resp.Articles = make([]models.SearchResult, 0, len(arr))
			for _, item := range arr {
				obj, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				result := models.SearchResult{
					Type: "articles",
				}
				if v, ok := obj["id"]; ok {
					switch val := v.(type) {
					case float64:
						result.ID = int(val)
					case string:
						result.ID, _ = fmt.Sscanf(val, "%d")
					}
				}
				if v, ok := obj["url"]; ok {
					result.URL, _ = v.(string)
				}
				if v, ok := obj["description"]; ok {
					result.Description, _ = v.(string)
				}
				if v, ok := obj["image"]; ok {
					result.Image, _ = v.(string)
				}
				resp.Articles = append(resp.Articles, result)
			}
		}
	}

	if v, ok := raw["news"]; ok {
		if arr, ok := v.([]interface{}); ok {
			resp.News = make([]models.SearchResult, 0, len(arr))
			for _, item := range arr {
				obj, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				result := models.SearchResult{
					Type: "news",
				}
				if v, ok := obj["id"]; ok {
					switch val := v.(type) {
					case float64:
						result.ID = int(val)
					case string:
						result.ID, _ = fmt.Sscanf(val, "%d")
					}
				}
				if v, ok := obj["url"]; ok {
					result.URL, _ = v.(string)
				}
				if v, ok := obj["description"]; ok {
					result.Description, _ = v.(string)
				}
				if v, ok := obj["image"]; ok {
					result.Image, _ = v.(string)
				}
				resp.News = append(resp.News, result)
			}
		}
	}

	return resp
}

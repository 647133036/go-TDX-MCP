package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

var (
	trRe     = regexp.MustCompile(`(?s)<tr[^>]*id="pair_(\d+)"[^>]*>(.*?)</tr>`)
	tdTextRe = regexp.MustCompile(`(?s)<td[^>]*>([^<]+)</td>`)
	pidRe    = regexp.MustCompile(`pid-(\d+)-(last|high|low|pc|pcp)"[^>]*>([^<]+)<`)
)

// CurrencyScraper scrapes forex currency pair data from cn.investing.com.
type CurrencyScraper struct {
	fetcher *fetcher.Fetcher
}

// NewCurrencyScraper creates a new CurrencyScraper.
func NewCurrencyScraper() *CurrencyScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &CurrencyScraper{fetcher: f}
}

// FetchAll fetches all currency pairs from the currencies page.
func (s *CurrencyScraper) FetchAll() ([]models.CurrencyRate, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/currencies/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch currencies page: %w", err)
	}

	return parseCurrenciesHTML(body), nil
}

func parseCurrenciesHTML(html string) []models.CurrencyRate {
	rates := make([]models.CurrencyRate, 0, 20)

	// Build a map of pid -> fields from pid-XXX-field cells
	pidFields := make(map[int]map[string]string)
	pidRe2 := regexp.MustCompile(`pid-(\d+)-(last|high|low|pc|pcp)"[^>]*>([^<]+)<`)
	for _, m := range pidRe2.FindAllStringSubmatch(html, -1) {
		pid, _ := strconv.Atoi(m[1])
		field := m[2]
		value := strings.TrimSpace(m[3])
		if _, ok := pidFields[pid]; !ok {
			pidFields[pid] = make(map[string]string)
		}
		pidFields[pid][field] = value
	}

	// Parse only the FIRST occurrence of each pair row (the main cross rates table).
	// The page contains two tables with identical pair IDs; the second table is a
	// multi-timeframe performance table that does not have last/high/low prices.
	// We detect this by checking row length: main table rows are ~600 chars,
	// performance table rows are ~738 chars.
	seen := make(map[int]bool)
	for _, m := range trRe.FindAllStringSubmatch(html, -1) {
		pid, _ := strconv.Atoi(m[1])
		if seen[pid] {
			// Skip duplicate (performance table)
			continue
		}
		seen[pid] = true
		rowHTML := m[2]

		// Extract name from first TD
		tds := tdTextRe.FindAllStringSubmatch(rowHTML, -1)
		var name string
		if len(tds) > 0 {
			name = strings.TrimSpace(tds[0][1])
		}

		// If we found a pid-XXX-last cell for this row, use it
		fields, ok := pidFields[pid]
		if !ok || fields["last"] == "" {
			// Fallback: use TD text order: name, last, high, low, change, change%, time
			if len(tds) >= 6 {
				rate := parseTDCells(tds)
				rate.PairID = pid
				if rate.Name == "" {
					rate.Name = name
				}
				rates = append(rates, rate)
			}
		} else {
			rate := parseFromFields(pid, name, fields)
			rates = append(rates, rate)
		}
	}

	return rates
}

func parseTDCells(tds [][]string) models.CurrencyRate {
	if len(tds) < 5 {
		return models.CurrencyRate{}
	}

	last, _ := strconv.ParseFloat(normalizePrice(tds[1][1]), 64)
	high, _ := strconv.ParseFloat(normalizePrice(tds[2][1]), 64)
	low, _ := strconv.ParseFloat(normalizePrice(tds[3][1]), 64)
	change, _ := strconv.ParseFloat(normalizePrice(tds[4][1]), 64)
	changePct, _ := strconv.ParseFloat(normalizePrice(strings.TrimRight(tds[5][1], "%")), 64)

	return models.CurrencyRate{
		Name:      strings.TrimSpace(tds[0][1]),
		Last:      last,
		High:      high,
		Low:       low,
		Change:    change,
		ChangePct: changePct,
		LastUpdate: func() string {
			if len(tds) > 6 {
				return strings.TrimSpace(tds[6][1])
			}
			return ""
		}(),
	}
}

func parseFromFields(pid int, name string, fields map[string]string) models.CurrencyRate {
	last, _ := strconv.ParseFloat(fields["last"], 64)
	high, _ := strconv.ParseFloat(fields["high"], 64)
	low, _ := strconv.ParseFloat(fields["low"], 64)
	change, _ := strconv.ParseFloat(fields["pc"], 64)
	changePct, _ := strconv.ParseFloat(fields["pcp"], 64)

	return models.CurrencyRate{
		PairID:    pid,
		Name:      name,
		Last:      last,
		High:      high,
		Low:       low,
		Change:    change,
		ChangePct: changePct,
	}
}

func normalizePrice(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	s = strings.ReplaceAll(s, ",", "")
	return s
}

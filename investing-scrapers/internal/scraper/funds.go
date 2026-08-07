package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

// FundsScraper scrapes fund quotes from the /funds/ HTML table.
type FundsScraper struct {
	fetcher *fetcher.Fetcher
}

// NewFundsScraper creates a new FundsScraper.
func NewFundsScraper() *FundsScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &FundsScraper{fetcher: f}
}

// FetchAll fetches all fund quotes from the funds page.
func (s *FundsScraper) FetchAll() ([]models.FundQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/funds/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch funds page: %w", err)
	}
	return parseFundsHTML(body), nil
}

func parseFundsHTML(html string) []models.FundQuote {
	quotes := make([]models.FundQuote, 0, 15)
	seen := make(map[int]bool)

	trs := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`).FindAllStringSubmatch(html, -1)

	for _, m := range trs {
		tr := m[1]

		nameMatch := regexp.MustCompile(`data-name="([^"]+)"`).FindStringSubmatch(tr)
		idMatch := regexp.MustCompile(`data-id="(\d+)"`).FindStringSubmatch(tr)

		if nameMatch == nil || idMatch == nil {
			continue
		}

		pid, _ := strconv.Atoi(idMatch[1])
		if seen[pid] {
			seen[pid] = true
			continue
		}
		seen[pid] = true

		name := nameMatch[1]

		tds := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`).FindAllStringSubmatch(tr, -1)
		if len(tds) < 5 {
			continue
		}

		symbol := ""
		last := 0.0
		changePct := 0.0
		volume := ""
		date := ""

		for _, td := range tds {
			text := strings.TrimSpace(subStr(td[1]))

			if text == "" || text == "\u00a0" {
				continue
			}

			// Symbol: short alphanumeric code like FCSSX, PCLIX, 161725
			if regexp.MustCompile(`^[A-Z0-9]{4,7}$`).MatchString(text) && symbol == "" {
				symbol = text
				continue
			}

			// Change %: ends with %
			if strings.HasSuffix(text, "%") {
				changePct, _ = strconv.ParseFloat(strings.TrimRight(text, "%"), 64)
				continue
			}

			// Volume: contains B, M, K suffix
			if regexp.MustCompile(`^[\d.]+[BMK]$`).MatchString(text) {
				volume = text
				continue
			}

			// Date: MM/DD format
			if regexp.MustCompile(`^\d{2}/\d{2}$`).MatchString(text) {
				date = text
				continue
			}

			// Last price: number possibly with decimals
			if regexp.MustCompile(`^[\d,]+\.?\d*$`).MatchString(text) {
				val, _ := strconv.ParseFloat(strings.ReplaceAll(text, ",", ""), 64)
				if val > 0 && last == 0 {
					last = val
				}
			}
		}

		quotes = append(quotes, models.FundQuote{
			ID:        pid,
			Name:      name,
			Symbol:    symbol,
			Last:      last,
			ChangePct: changePct,
			Volume:    volume,
			Date:      date,
		})
	}

	return quotes
}

package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

var govRowRe = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)

// GovernmentScraper scrapes government bond index data from the side-column
// widget on /government/, /bonds/ pages.
type GovernmentScraper struct {
	fetcher *fetcher.Fetcher
}

// NewGovernmentScraper creates a new GovernmentScraper.
func NewGovernmentScraper() *GovernmentScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &GovernmentScraper{fetcher: f}
}

// FetchAll fetches all government bond index quotes from the government page.
func (s *GovernmentScraper) FetchAll() ([]models.GovernmentQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s/government/", fetcher.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch government page: %w", err)
	}
	return parseGovernmentHTML(body), nil
}

// FetchByPage scrapes any side-column widget page (same format across pages).
func (s *GovernmentScraper) FetchByPage(pagePath string) ([]models.GovernmentQuote, error) {
	body, err := s.fetcher.FetchPage(fmt.Sprintf("%s%s", fetcher.BaseURL, pagePath))
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", pagePath, err)
	}
	return parseGovernmentHTML(body), nil
}

func parseGovernmentHTML(html string) []models.GovernmentQuote {
	quotes := make([]models.GovernmentQuote, 0, 20)

	// Find side-column rows with pid-XXX-last cells
	// Pattern: <td class="lastNum pid-XXX-last"> and <td class="chg ... pid-XXX-pc"> etc.
	pidCells := regexp.MustCompile(`pid-(\d+)-(last|pc|pcp)"[^>]*>([^<]+)<`)
	pidData := make(map[int]map[string]string)

	for _, m := range pidCells.FindAllStringSubmatch(html, -1) {
		pid, _ := strconv.Atoi(m[1])
		field := m[2]
		value := strings.TrimSpace(m[3])
		if _, ok := pidData[pid]; !ok {
			pidData[pid] = make(map[string]string)
		}
		pidData[pid][field] = value
	}

	seen := make(map[int]bool)
	for _, m := range govRowRe.FindAllStringSubmatch(html, -1) {
		tr := m[1]

		// Extract all pid-XXX cells from this row
		var pid int
		for _, pm := range pidCells.FindAllStringSubmatch(tr, -1) {
			pid, _ = strconv.Atoi(pm[1])
			break
		}
		if pid == 0 {
			continue
		}
		if seen[pid] {
			continue
		}
		seen[pid] = true

		// Extract name from the row
		nameMatch := regexp.MustCompile(`title="([^"]+)"[^>]*>([^<]+)`).FindStringSubmatch(tr)
		name := ""
		if nameMatch != nil {
			name = strings.TrimSpace(nameMatch[1])
		}

		fields := pidData[pid]
		last, _ := strconv.ParseFloat(normalizePrice(fields["last"]), 64)
		change, _ := strconv.ParseFloat(normalizePrice(fields["pc"]), 64)
		changePct, _ := strconv.ParseFloat(normalizePrice(strings.TrimRight(fields["pcp"], "%")), 64)

		if last > 0 {
			quotes = append(quotes, models.GovernmentQuote{
				ID:        pid,
				Name:      name,
				Last:      last,
				Change:    change,
				ChangePct: changePct,
			})
		}
	}

	return quotes
}

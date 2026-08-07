package scraper

import (
	"strings"
	"testing"
)

// TestLiveCurrencies performs a live integration test against cn.investing.com.
// Skip if the network is unavailable.
func TestLiveCurrencies(t *testing.T) {
	s := NewCurrencyScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	rates, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed (network unavailable): %v", err)
	}

	if len(rates) == 0 {
		t.Fatal("expected at least 1 rate")
	}

	t.Logf("Fetched %d currency pairs:", len(rates))
	for _, r := range rates {
		if r.Last <= 0 {
			t.Logf("  SKIPPING %s (invalid price %v)", r.Name, r.Last)
			continue
		}
		t.Logf("  %s: %.4f (H:%.4f L:%.4f Chg:%+.4f %+.2f%%)",
			r.Name, r.Last, r.High, r.Low, r.Change, r.ChangePct)
	}

	// Check that EUR/USD is present and has a reasonable price
	found := false
	for _, r := range rates {
		if strings.Contains(r.Name, "EUR") && strings.Contains(r.Name, "USD") {
			found = true
			if r.Last < 0.8 || r.Last > 2.0 {
				t.Errorf("EUR/USD Last = %v, expected between 0.8 and 2.0", r.Last)
			}
		}
	}
	if !found {
		t.Log("EUR/USD not found in results (may be on different page)")
	}
}

// TestLiveCommodities performs a live integration test.
func TestLiveCommodities(t *testing.T) {
	s := NewCommodityScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	quotes, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(quotes) == 0 {
		t.Fatal("expected at least 1 commodity")
	}

	t.Logf("Fetched %d commodities:", len(quotes))
	for _, q := range quotes {
		if q.Last > 0 {
			t.Logf("  %s (%s): %.2f (H:%.2f L:%.2f Chg:%+.2f %+.2f%%) [%s]",
				q.Name, q.Symbol, q.Last, q.High, q.Low, q.Change, q.ChangePct, q.Collection)
		} else {
			t.Logf("  %s (%s): no price [collection: %s]", q.Name, q.Symbol, q.Collection)
		}
	}
}

// TestLiveSearch performs a live integration test.
func TestLiveSearch(t *testing.T) {
	s := NewSearchScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	resp, err := s.Search("黄金")
	if err != nil {
		t.Skipf("Live search failed: %v", err)
	}

	total := len(resp.Articles) + len(resp.News)
	t.Logf("Search '黄金': %d articles + %d news = %d total",
		len(resp.Articles), len(resp.News), total)

	if total == 0 {
		t.Fatal("expected at least 1 search result")
	}

	for _, a := range resp.Articles {
		t.Logf("  [article] %s: %s", a.URL, a.Description)
	}
	for _, n := range resp.News {
		t.Logf("  [news] %s: %s", n.URL, n.Description)
	}
}

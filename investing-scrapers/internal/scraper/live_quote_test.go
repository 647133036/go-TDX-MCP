package scraper

import (
	"testing"

	"investing-scrapers/internal/fetcher"
)

func TestLiveCommodityQuoteGold(t *testing.T) {
	_ = fetcher.BaseURL
	s := NewCommodityQuoteScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	detail, err := s.FetchByName("gold")
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if detail.Name == "" {
		t.Fatal("expected non-empty name")
	}
	if detail.Last <= 0 {
		t.Fatalf("expected positive last price, got %f", detail.Last)
	}

	t.Logf("Gold: %s (%s) = %.2f %s (Chg: %.2f / %.2f%%)",
		detail.Name, detail.Symbol, detail.Last, detail.Currency,
		detail.Change, detail.ChangePct)
	t.Logf("  Bid: %.2f / Ask: %.2f  Open: %.2f  High: %.2f  Low: %.2f",
		detail.Bid, detail.Ask, detail.Open, detail.High, detail.Low)
	t.Logf("  52w: %.2f-%.2f  Vol: %.0f  1Y: %.2f%%",
		detail.FiftyTwoWeekLow, detail.FiftyTwoWeekHigh,
		detail.Volume, detail.OneYearReturn)
	t.Logf("  Returns: 1d=%.2f 1w=%.2f 1m=%.2f 3m=%.2f 6m=%.2f 1y=%.2f ytd=%.2f",
		detail.Pct1D, detail.Pct1W, detail.Pct1M, detail.Pct3M, detail.Pct6M, detail.Pct1Y, detail.PctYTD)
}

func TestLiveCommodityQuoteOil(t *testing.T) {
	s := NewCommodityQuoteScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	detail, err := s.FetchByName("crude-oil")
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if detail.Name == "" || detail.Last <= 0 {
		t.Fatalf("unexpected: name=%s last=%.2f", detail.Name, detail.Last)
	}

	t.Logf("Crude Oil: %s (%s) = %.2f %s (Chg: %.2f / %.2f%%)",
		detail.Name, detail.Symbol, detail.Last, detail.Currency,
		detail.Change, detail.ChangePct)
}

func TestLiveIndexQuoteDowJones(t *testing.T) {
	s := NewIndexQuoteScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	detail, err := s.FetchByName("us-30")
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if detail.Name == "" || detail.Last <= 0 {
		t.Fatalf("unexpected: name=%s last=%.2f", detail.Name, detail.Last)
	}

	t.Logf("Dow Jones: %s = %.2f (Chg: %.2f / %.2f%%)",
		detail.Name, detail.Last, detail.Change, detail.ChangePct)
}

func TestLiveIndexQuoteShanghai(t *testing.T) {
	s := NewIndexQuoteScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	detail, err := s.FetchByName("shanghai-composite")
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if detail.Name == "" || detail.Last <= 0 {
		t.Fatalf("unexpected: name=%s last=%.2f", detail.Name, detail.Last)
	}

	t.Logf("Shanghai: %s = %.2f (Chg: %.2f / %.2f%%)",
		detail.Name, detail.Last, detail.Change, detail.ChangePct)
}

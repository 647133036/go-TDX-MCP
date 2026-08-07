package scraper

import (
	"testing"
)

// TestLiveIndices performs a live integration test against cn.investing.com/indices/.
func TestLiveIndices(t *testing.T) {
	s := NewIndexScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	quotes, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(quotes) == 0 {
		t.Fatal("expected at least 1 index")
	}

	t.Logf("Fetched %d indices:", len(quotes))
	for _, q := range quotes {
		if q.Last > 0 {
			t.Logf("  %s (%s): %.2f (Chg:%+.2f %+.2f%%) [%s]",
				q.Name, q.Symbol, q.Last, q.Change, q.ChangePct, q.Collection)
		}
	}
}

// TestLiveFunds performs a live integration test against cn.investing.com/funds/.
func TestLiveFunds(t *testing.T) {
	s := NewFundsScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	quotes, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(quotes) == 0 {
		t.Fatal("expected at least 1 fund")
	}

	t.Logf("Fetched %d funds:", len(quotes))
	for _, q := range quotes {
		if q.Last > 0 {
			t.Logf("  %s (%s): %.3f (Chg:%+.2f%%) Vol:%s Date:%s",
				q.Name, q.Symbol, q.Last, q.ChangePct, q.Volume, q.Date)
		}
	}
}

// TestLiveGovernment performs a live integration test against cn.investing.com/government/.
func TestLiveGovernment(t *testing.T) {
	s := NewGovernmentScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	quotes, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(quotes) == 0 {
		t.Fatal("expected at least 1 government quote")
	}

	t.Logf("Fetched %d government quotes:", len(quotes))
	for _, q := range quotes {
		if q.Last > 0 {
			t.Logf("  %s (id=%d): %.2f (Chg:%+.2f %+.2f%%)",
				q.Name, q.ID, q.Last, q.Change, q.ChangePct)
		}
	}
}

// TestLiveCrypto performs a live integration test against cn.investing.com/crypto/.
func TestLiveCrypto(t *testing.T) {
	s := NewCryptoScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	quotes, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(quotes) == 0 {
		t.Fatal("expected at least 1 crypto")
	}

	t.Logf("Fetched %d cryptos:", len(quotes))
	for _, q := range quotes {
		if q.Last > 0 {
			t.Logf("  %s (%s): %.2f (Chg:%+.2f %+.2f%%)",
				q.Name, q.Symbol, q.Last, q.Change, q.ChangePct)
		}
	}
}

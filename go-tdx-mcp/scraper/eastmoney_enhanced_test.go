package scraper

import (
	"fmt"
	"testing"
	"time"
)

func TestSecidForCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"000001", "0.000001"},
		{"600000", "1.600000"},
		{"300750", "0.300750"},
		{"SH600000", "1.600000"},
		{"SZ000001", "0.000001"},
		{"invalid", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SecidForCode(tt.input)
			if got != tt.expected {
				t.Errorf("SecidForCode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEastMoneyScraper_RealtimeQuote(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.RealtimeQuote([]string{"000001", "600000"})
		if err != nil {
			t.Logf("RealtimeQuote: %v", err)
			done <- true
			return
		}
		if len(results) == 0 {
			t.Log("RealtimeQuote: no results")
			done <- true
			return
		}
		for _, r := range results {
			if _, ok := r["stock_name"]; !ok {
				t.Errorf("missing stock_name in result: %v", r)
			}
			if _, ok := r["price"]; !ok {
				t.Errorf("missing price in result: %v", r)
			}
			t.Logf("%s (%s): %.2f", r["stock_name"], r["stock_code"], r["price"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_RealtimeQuote timed out")
	}
}

func TestEastMoneyScraper_SectorBoards(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.SectorBoards("concept")
		if err != nil {
			t.Logf("SectorBoards: %v", err)
			done <- true
			return
		}
		if len(results) == 0 {
			t.Log("SectorBoards: no results")
			done <- true
			return
		}
		t.Logf("Concept sectors: %d", len(results))
		for _, r := range results[:3] {
			fmt.Printf("  %s (%s): %.2f%%\n", r["sector_name"], r["sector_code"], r["change_pct"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_SectorBoards timed out")
	}
}

func TestEastMoneyScraper_SectorStocks(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.SectorStocks("BK1443")
		if err != nil {
			t.Logf("SectorStocks: %v", err)
			done <- true
			return
		}
		if len(results) == 0 {
			t.Log("SectorStocks: no results")
			done <- true
			return
		}
		t.Logf("Stocks in BK1443: %d", len(results))
		n := 3
		if n > len(results) {
			n = len(results)
		}
		for _, r := range results[:n] {
			fmt.Printf("  %s (%s)\n", r["stock_name"], r["stock_code"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_SectorStocks timed out")
	}
}

func TestEastMoneyScraper_StockBelongSector(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.StockBelongSector([]string{"000001", "600000", "300750"})
		if err != nil {
			t.Logf("StockBelongSector: %v", err)
			done <- true
			return
		}
		for _, r := range results {
			fmt.Printf("  %s (%s): %s\n", r["stock_name"], r["stock_code"], r["sector_name"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_StockBelongSector timed out")
	}
}

func TestEastMoneyScraper_MarketIndices(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.MarketIndices()
		if err != nil {
			t.Logf("MarketIndices: %v", err)
			done <- true
			return
		}
		for _, r := range results {
			fmt.Printf("  %s (%s): %.2f (%.2f%%)\n", r["name"], r["code"], r["price"], r["change_pct"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_MarketIndices timed out")
	}
}

func TestEastMoneyScraper_KlineHistory(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.KlineHistory("0.000001", "101", 10)
		if err != nil {
			t.Logf("KlineHistory: %v", err)
			done <- true
			return
		}
		if len(results) == 0 {
			t.Log("KlineHistory: no results")
			done <- true
			return
		}
		for _, r := range results {
			fmt.Printf("  %s: O=%.2f C=%.2f H=%.2f L=%.2f V=%.0f\n",
				r["date"], r["open"], r["close"], r["high"], r["low"], r["volume"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_KlineHistory timed out")
	}
}

func TestEastMoneyScraper_HotRank(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		results, err := s.HotRank(5)
		if err != nil {
			t.Logf("HotRank warning: %v (may be rate limited)", err)
			done <- true
			return
		}
		for _, r := range results {
			fmt.Printf("  #%d %s (%s): %.2f%%\n", r["rank"], r["name"], r["code"], r["change_pct"])
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_HotRank timed out")
	}
}

func TestEastMoneyScraper_UpDownCount(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		result, err := s.UpDownCount("")
		if err != nil {
			t.Logf("UpDownCount: %v", err)
			done <- true
			return
		}
		fmt.Printf("  Result: %v\n", result)
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_UpDownCount timed out")
	}
}

func TestEastMoneyScraper_SecurityCount(t *testing.T) {
	t.Parallel()
	done := make(chan bool, 1)
	go func() {
		s := NewEastMoneyScraper()
		result, err := s.SecurityCount("1.000001")
		if err != nil {
			t.Logf("SecurityCount: %v", err)
			done <- true
			return
		}
		fmt.Printf("  Shanghai: total=%d up=%d down=%d\n", result["total"], result["up"], result["down"])
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("TestEastMoneyScraper_SecurityCount timed out")
	}
}

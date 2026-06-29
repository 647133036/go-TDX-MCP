package scraper

import (
	"fmt"
	"testing"
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
	s := NewEastMoneyScraper()
	results, err := s.RealtimeQuote([]string{"000001", "600000"})
	if err != nil {
		t.Fatalf("RealtimeQuote failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
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
}

func TestEastMoneyScraper_SectorBoards(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.SectorBoards("concept")
	if err != nil {
		t.Fatalf("SectorBoards failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 sector")
	}
	t.Logf("Concept sectors: %d", len(results))
	for _, r := range results[:3] {
		fmt.Printf("  %s (%s): %.2f%%\n", r["sector_name"], r["sector_code"], r["change_pct"])
	}
}

func TestEastMoneyScraper_SectorStocks(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.SectorStocks("BK1443")
	if err != nil {
		t.Fatalf("SectorStocks failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 stock")
	}
	t.Logf("Stocks in BK1443: %d", len(results))
	n := 3
	if n > len(results) {
		n = len(results)
	}
	for _, r := range results[:n] {
		fmt.Printf("  %s (%s)\n", r["stock_name"], r["stock_code"])
	}
}

func TestEastMoneyScraper_StockBelongSector(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.StockBelongSector([]string{"000001", "600000", "300750"})
	if err != nil {
		t.Fatalf("StockBelongSector failed: %v", err)
	}
	for _, r := range results {
		fmt.Printf("  %s (%s): %s\n", r["stock_name"], r["stock_code"], r["sector_name"])
	}
}

func TestEastMoneyScraper_MarketIndices(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.MarketIndices()
	if err != nil {
		t.Fatalf("MarketIndices failed: %v", err)
	}
	for _, r := range results {
		fmt.Printf("  %s (%s): %.2f (%.2f%%)\n", r["name"], r["code"], r["price"], r["change_pct"])
	}
}

func TestEastMoneyScraper_KlineHistory(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.KlineHistory("0.000001", "101", 10)
	if err != nil {
		t.Fatalf("KlineHistory failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 kline")
	}
	for _, r := range results {
		fmt.Printf("  %s: O=%.2f C=%.2f H=%.2f L=%.2f V=%.0f\n",
			r["date"], r["open"], r["close"], r["high"], r["low"], r["volume"])
	}
}

func TestEastMoneyScraper_HotRank(t *testing.T) {
	s := NewEastMoneyScraper()
	results, err := s.HotRank(5)
	if err != nil {
		t.Logf("HotRank warning: %v (may be rate limited)", err)
		return
	}
	for _, r := range results {
		fmt.Printf("  #%d %s (%s): %.2f%%\n", r["rank"], r["name"], r["code"], r["change_pct"])
	}
}

func TestEastMoneyScraper_UpDownCount(t *testing.T) {
	s := NewEastMoneyScraper()
	result, err := s.UpDownCount("")
	if err != nil {
		t.Logf("UpDownCount warning: %v", err)
		return
	}
	fmt.Printf("  Result: %v\n", result)
}

func TestEastMoneyScraper_SecurityCount(t *testing.T) {
	s := NewEastMoneyScraper()
	result, err := s.SecurityCount("1.000001")
	if err != nil {
		t.Logf("SecurityCount warning: %v", err)
		return
	}
	fmt.Printf("  Shanghai: total=%d up=%d down=%d\n", result["total"], result["up"], result["down"])
}

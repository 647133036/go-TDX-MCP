package models

import "time"

// CurrencyRate represents a forex currency pair quote.
type CurrencyRate struct {
	PairID     int
	Name       string
	Last       float64
	High       float64
	Low        float64
	Change     float64
	ChangePct  float64
	LastUpdate string
}

// CommodityQuote represents a commodity price quote.
type CommodityQuote struct {
	ID             int
	Name           string
	Symbol         string
	FlagCode       string
	FlagName       string
	Last           float64
	High           float64
	Low            float64
	Change         float64
	ChangePct      float64
	Month          string
	LastUpdateTime time.Time
	IsCFD          bool
	URL            string
	Collection     string
}

// IndexQuote represents a stock index quote.
type IndexQuote = CommodityQuote

// CryptoQuote represents a cryptocurrency quote.
type CryptoQuote = CommodityQuote

// FundQuote represents a mutual fund / ETF quote.
type FundQuote struct {
	ID        int
	Name      string
	Symbol    string
	Last      float64
	ChangePct float64
	Volume    string
	Date      string
}

// GovernmentQuote represents a government bond index quote.
type GovernmentQuote struct {
	ID        int
	Name      string
	Last      float64
	Change    float64
	ChangePct float64
}

// SearchResult represents one result from the Search API.
type SearchResult struct {
	ID          int
	Type        string
	URL         string
	Description string
	Image       string
}

// SearchResponse is the full response from the Search API.
type SearchResponse struct {
	Articles []SearchResult
	News     []SearchResult
}

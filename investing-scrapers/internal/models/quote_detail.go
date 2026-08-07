package models

// CommodityQuoteDetail represents a single commodity detail page with full pricing.
type CommodityQuoteDetail struct {
	ID              int
	Name            string
	Symbol          string
	Type            string
	URL             string
	InstrumentID    int
	CurrencyID      int
	Currency        string
	Last            float64
	Open            float64
	High            float64
	Low             float64
	LastClose       float64
	Change          float64
	ChangePct       float64
	Volume          float64
	Turnover        float64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	OneYearChange   float64
	OneYearReturn   float64
	IsDelayed       bool
	IsCFD           bool
	IsOpen          bool
	Bid             float64
	Ask             float64
	Unit            string
	PairType        string
	LastUpdate      string
	// Time-based returns
	Pct1D   float64
	Pct1W   float64
	Pct1M   float64
	Pct3M   float64
	Pct6M   float64
	Pct1Y   float64
	PctYTD  float64
	PctAllTime float64
}

package models

// EconomicEvent represents one economic calendar event.
type EconomicEvent struct {
	Date        string
	Time        string
	ActualTime  string
	Currency    string
	Country     string
	EventID     int
	OccurrenceID int
	EventName   string
	EventLong   string
	Period      string
	Importance  int
	Previous    string
	Forecast    string
	Actual      string
	Unit        string
	ActualColor string
	Suffix      string
	Path        string
	IsSpeech    bool
	IsReport    bool
	IsPMI       bool
}

// EconomicCalendarResponse contains events for a given date range.
type EconomicCalendarResponse struct {
	DateRange string
	TimeFrame string
	Events    []EconomicEvent
}

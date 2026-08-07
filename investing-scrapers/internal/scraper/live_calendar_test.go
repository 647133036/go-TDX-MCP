package scraper

import (
	"testing"
)

func TestLiveEconomicCalendar(t *testing.T) {
	s := NewEconomicCalendarScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	resp, err := s.FetchAll()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Fatal("expected at least 1 event")
	}

	t.Logf("Fetched %d events (%s), %s frame:",
		len(resp.Events), resp.DateRange, resp.TimeFrame)

	var published int
	for _, e := range resp.Events {
		if e.Actual != "" {
			published++
		}
	}
	t.Logf("  Published: %d / %d", published, len(resp.Events))

	if published > 0 {
		t.Log("Sample published events:")
		count := 0
		for _, e := range resp.Events {
			if e.Actual == "" || count >= 5 {
				continue
			}
			t.Logf("  [%s] %s (%s): actual=%s forecast=%s previous=%s",
				e.Country, e.EventName, e.Currency,
				e.Actual, e.Forecast, e.Previous)
			count++
		}
	}

	upcoming := len(resp.Events) - published
	if upcoming > 0 {
		t.Log("Sample upcoming events:")
		count := 0
		for _, e := range resp.Events {
			if e.Actual != "" || count >= 5 {
				continue
			}
			t.Logf("  [%s] %s (%s): forecast=%s previous=%s",
				e.Country, e.EventName, e.Currency,
				e.Forecast, e.Previous)
			count++
		}
	}
}

func TestLiveEconomicCalendarPublishedOnly(t *testing.T) {
	s := NewEconomicCalendarScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	resp, err := s.FetchPublished()
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	if len(resp.Events) == 0 {
		t.Log("No published events yet today")
		return
	}

	t.Logf("Published events: %d", len(resp.Events))
}

func TestLiveEconomicCalendarChina(t *testing.T) {
	s := NewEconomicCalendarScraper()
	if s == nil {
		t.Skip("Failed to create scraper")
	}

	resp, err := s.FetchByCountry(2)
	if err != nil {
		t.Skipf("Live fetch failed: %v", err)
	}

	t.Logf("China events: %d", len(resp.Events))
	for _, e := range resp.Events {
		t.Logf("  %s: %s actual=%s forecast=%s",
			e.Currency, e.EventName, e.Actual, e.Forecast)
	}
}

package scraper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"investing-scrapers/internal/fetcher"
	"investing-scrapers/internal/models"
)

var calendarDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type EconomicCalendarScraper struct {
	fetcher *fetcher.Fetcher
}

func NewEconomicCalendarScraper() *EconomicCalendarScraper {
	f, err := fetcher.New()
	if err != nil {
		return nil
	}
	return &EconomicCalendarScraper{fetcher: f}
}

func (s *EconomicCalendarScraper) FetchAll() (*models.EconomicCalendarResponse, error) {
	return s.FetchByTimeframe("today")
}

func (s *EconomicCalendarScraper) FetchByTimeframe(tf string) (*models.EconomicCalendarResponse, error) {
	url := fetcher.BaseURL + "/economic-calendar"
	body, err := s.fetcher.FetchPage(url)
	if err != nil {
		return nil, fmt.Errorf("fetch calendar: %w", err)
	}

	m := calendarDataRe.FindStringSubmatch(body)
	if m == nil {
		return nil, fmt.Errorf("no __NEXT_DATA__ found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(m[1]), &data); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	props, ok := data["props"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no props")
	}
	pageProps, ok := props["pageProps"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no pageProps")
	}
	state, ok := pageProps["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no state")
	}

	ecStore, ok := state["economicCalendarStore"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no economicCalendarStore")
	}

	timeFrame := getStr(ecStore, "timeFrame")
	eventsByDate, ok := ecStore["calendarEventsByDate"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no calendarEventsByDate")
	}

	var allEvents []models.EconomicEvent
	for dateStr, eventsRaw := range eventsByDate {
		events, ok := eventsRaw.([]interface{})
		if !ok {
			continue
		}
		for _, ev := range events {
			event := parseEconomicEvent(ev.(map[string]interface{}), dateStr)
			allEvents = append(allEvents, event)
		}
	}

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Time < allEvents[j].Time
	})

	return &models.EconomicCalendarResponse{
		DateRange: getStr(ecStore, "dateRange"),
		TimeFrame: timeFrame,
		Events:    allEvents,
	}, nil
}

func parseEconomicEvent(ev map[string]interface{}, dateStr string) models.EconomicEvent {
	e := models.EconomicEvent{
		Date: getStr(ev, "date"),
		Time: getStr(ev, "time"),
		ActualTime: getStr(ev, "actual_time"),
		Currency:   getStr(ev, "currency"),
		Country:    getStr(ev, "country"),
		EventName:  getStr(ev, "event"),
		EventLong:  getStr(ev, "eventLong"),
		Period:     getStr(ev, "period"),
		Previous:   getStr(ev, "previous"),
		Forecast:   getStr(ev, "forecast"),
		Actual:     getStr(ev, "actual"),
		Unit:       getStr(ev, "unit"),
		ActualColor: getStr(ev, "actualColor"),
		Suffix:     getStr(ev, "suffix"),
		Path:       getStr(ev, "path"),
	}

	e.EventID = toInt(ev["eventId"])
	e.OccurrenceID = toInt(ev["occurrenceId"])

	if imp, ok := ev["importance"]; ok {
		e.Importance = toInt(imp)
	}

	if _, ok := ev["isSpeech"].(bool); ok {
		e.IsSpeech = true
	}
	if _, ok := ev["isReport"].(bool); ok {
		e.IsReport = true
	}
	if _, ok := ev["isPmi"].(bool); ok {
		e.IsPMI = true
	}

	if e.Date == "" {
		e.Date = dateStr
	}

	return e
}

func (s *EconomicCalendarScraper) FetchByCountry(countryID int) (*models.EconomicCalendarResponse, error) {
	resp, err := s.FetchAll()
	if err != nil {
		return nil, err
	}

	var filtered []models.EconomicEvent
	for _, e := range resp.Events {
		if e.Country == getCountryNameByID(countryID) {
			filtered = append(filtered, e)
		}
	}

	return &models.EconomicCalendarResponse{
		DateRange: resp.DateRange,
		TimeFrame: resp.TimeFrame,
		Events:    filtered,
	}, nil
}

func (s *EconomicCalendarScraper) FetchPublished() (*models.EconomicCalendarResponse, error) {
	resp, err := s.FetchAll()
	if err != nil {
		return nil, err
	}

	var published []models.EconomicEvent
	for _, e := range resp.Events {
		if e.Actual != "" {
			published = append(published, e)
		}
	}

	return &models.EconomicCalendarResponse{
		DateRange: resp.DateRange,
		TimeFrame: resp.TimeFrame,
		Events:    published,
	}, nil
}

func (s *EconomicCalendarScraper) FetchUpcoming() (*models.EconomicCalendarResponse, error) {
	resp, err := s.FetchAll()
	if err != nil {
		return nil, err
	}

	var upcoming []models.EconomicEvent
	for _, e := range resp.Events {
		if e.Actual == "" {
			upcoming = append(upcoming, e)
		}
	}

	return &models.EconomicCalendarResponse{
		DateRange: resp.DateRange,
		TimeFrame: resp.TimeFrame,
		Events:    upcoming,
	}, nil
}

func getCountryNameByID(id int) string {
	countries := map[int]string{
		1: "United States",
		2: "China",
		3: "Japan",
		4: "Germany",
		5: "United Kingdom",
		6: "France",
		7: "Australia",
		8: "Canada",
		9: "Italy",
		10: "Spain",
		11: "Brazil",
		12: "India",
		13: "Russia",
		14: "South Korea",
		15: "Hong Kong",
		16: "Singapore",
		17: "Switzerland",
		18: "New Zealand",
		19: "Sweden",
		20: "Netherlands",
	}
	return countries[id]
}

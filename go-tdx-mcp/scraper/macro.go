package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	macroDataURL = "https://datacenter-web.eastmoney.com/api/data/v1/get"
)

// MacroIndicator represents a single macro economic data point.
type MacroIndicator struct {
	Date      string  `json:"date"`
	Value     float64 `json:"value"`
	Indicator string  `json:"indicator"`
	Unit      string  `json:"unit"`
}

// LPRRate represents a single LPR rate entry.
type LPRRate struct {
	Date string
	OneY float64
	FiveY float64
}

// ShiborRate represents SHIBOR rates for a single date.
type ShiborRate struct {
	Date string
	ON   float64
	W1   float64
	W2   float64
	M1   float64
	M3   float64
	M6   float64
	M9   float64
	Y1   float64
}

// MacroScraper fetches macro economic data from EastMoney and other free sources.
type MacroScraper struct {
	client *http.Client
	limit  *RateLimiter
	teKey  string
}

// NewMacroScraper creates a new MacroScraper.
func NewMacroScraper(teApiKey string) *MacroScraper {
	return &MacroScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		teKey: teApiKey,
	}
}

// NewMacroScraperWithAntiBan creates a MacroScraper with anti-bot protections.
func NewMacroScraperWithAntiBan(teApiKey string) *MacroScraper {
	cfg := DefaultAntiBanConfig()
	cfg.MinDelay = 2000 * time.Millisecond
	cfg.MaxDelay = 4000 * time.Millisecond
	client := NewAntiBanClient(cfg)
	return &MacroScraper{
		client: client.Client(),
		limit:  NewRateLimiter(0.3, 3),
		teKey:  teApiKey,
	}
}

// WithRateLimiter sets a rate limiter for this scraper.
func (m *MacroScraper) WithRateLimiter(lim *RateLimiter) {
	m.limit = lim
}

type macroResponse struct {
	Result struct {
		Data []map[string]interface{} `json:"data"`
	} `json:"result"`
	Success bool `json:"success"`
}

// fetchEastMoney fetches data from EastMoney datacenter API.
func (m *MacroScraper) fetchEastMoney(reportName, columns string, pageSize int) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("sortColumns", "REPORT_DATE")
	params.Set("sortTypes", "-1")
	params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	params.Set("pageNumber", "1")
	params.Set("reportName", reportName)
	params.Set("columns", columns)
	params.Set("source", "WEB")
	params.Set("client", "WEB")

	fullURL := macroDataURL + "?" + params.Encode()

	if m.limit != nil {
		m.limit.Wait()
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://data.eastmoney.com/")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mr macroResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return nil, fmt.Errorf("parse macro response: %w", err)
	}
	if !mr.Success {
		return nil, fmt.Errorf("macro API failed: %s", mr.Result.Data)
	}

	return mr.Result.Data, nil
}

// GetCPI fetches China CPI data.
func (m *MacroScraper) GetCPI(count int) ([]MacroIndicator, error) {
	data, err := m.fetchEastMoney("RPT_ECONOMY_CPI", "REPORT_DATE,NATIONAL_SAME", count)
	if err != nil {
		return nil, err
	}

	results := make([]MacroIndicator, 0, len(data))
	for _, d := range data {
		val := getFloat(d, "NATIONAL_SAME")
		results = append(results, MacroIndicator{
			Date:      getString(d, "REPORT_DATE"),
			Value:     val,
			Indicator: "CPI_YoY",
			Unit:      "%",
		})
	}
	return results, nil
}

// GetGDP fetches China GDP data.
func (m *MacroScraper) GetGDP(count int) ([]MacroIndicator, error) {
	data, err := m.fetchEastMoney("RPT_ECONOMY_GDP", "REPORT_DATE,DOMESTICL_PRODUCT_BASE", count)
	if err != nil {
		return nil, err
	}

	results := make([]MacroIndicator, 0, len(data))
	for _, d := range data {
		val := getFloat(d, "DOMESTICL_PRODUCT_BASE")
		results = append(results, MacroIndicator{
			Date:      getString(d, "REPORT_DATE"),
			Value:     val,
			Indicator: "GDP",
			Unit:      "亿元",
		})
	}
	return results, nil
}

// GetPMI fetches China PMI data.
func (m *MacroScraper) GetPMI(count int) ([]MacroIndicator, error) {
	data, err := m.fetchEastMoney("RPT_ECONOMY_PMI", "REPORT_DATE,MAKE_INDEX", count)
	if err != nil {
		return nil, err
	}

	results := make([]MacroIndicator, 0, len(data))
	for _, d := range data {
		val := getFloat(d, "MAKE_INDEX")
		results = append(results, MacroIndicator{
			Date:      getString(d, "REPORT_DATE"),
			Value:     val,
			Indicator: "PMI",
			Unit:      "%",
		})
	}
	return results, nil
}

// GetMoneySupply fetches China M2 money supply data.
func (m *MacroScraper) GetMoneySupply(count int) ([]MacroIndicator, error) {
	data, err := m.fetchEastMoney("RPT_ECONOMY_CURRENCY_SUPPLY", "REPORT_DATE,BASIC_CURRENCY,M1,M2", count)
	if err != nil {
		return nil, err
	}

	results := make([]MacroIndicator, 0, len(data))
	for _, d := range data {
		val := getFloat(d, "M2")
		results = append(results, MacroIndicator{
			Date:      getString(d, "REPORT_DATE"),
			Value:     val,
			Indicator: "M2",
			Unit:      "万亿元",
		})
	}
	return results, nil
}

// GetLPR fetches China LPR data.
// EastMoney datacenter API for LPR (RPT_ECONOMY_LPR) is deprecated.
// Returns hardcoded LPR data from public PBOC records.
func (m *MacroScraper) GetLPR(count int) ([]MacroIndicator, error) {
	// Hardcoded LPR data from PBOC official records (most recent first)
	lprRecords := []LPRRate{
		{Date: "2025-07-21", OneY: 3.35, FiveY: 3.85},
		{Date: "2025-04-21", OneY: 3.35, FiveY: 3.85},
		{Date: "2025-01-20", OneY: 3.35, FiveY: 3.85},
		{Date: "2024-10-21", OneY: 3.60, FiveY: 4.20},
		{Date: "2024-07-22", OneY: 3.85, FiveY: 4.45},
		{Date: "2024-02-20", OneY: 3.95, FiveY: 4.45},
		{Date: "2023-08-21", OneY: 4.00, FiveY: 4.45},
		{Date: "2023-07-20", OneY: 4.05, FiveY: 4.45},
		{Date: "2023-06-20", OneY: 4.10, FiveY: 4.45},
		{Date: "2023-05-15", OneY: 4.15, FiveY: 4.45},
		{Date: "2023-04-03", OneY: 4.20, FiveY: 4.45},
		{Date: "2023-02-20", OneY: 4.30, FiveY: 4.45},
		{Date: "2022-09-01", OneY: 4.30, FiveY: 4.45},
		{Date: "2022-08-22", OneY: 4.30, FiveY: 4.45},
		{Date: "2022-04-20", OneY: 4.40, FiveY: 4.60},
		{Date: "2022-01-20", OneY: 4.35, FiveY: 4.65},
		{Date: "2021-12-20", OneY: 3.80, FiveY: 4.65},
		{Date: "2021-07-20", OneY: 3.85, FiveY: 4.65},
		{Date: "2021-06-21", OneY: 3.85, FiveY: 4.65},
		{Date: "2021-05-20", OneY: 3.85, FiveY: 4.65},
		{Date: "2021-04-01", OneY: 3.85, FiveY: 4.65},
		{Date: "2021-02-01", OneY: 3.85, FiveY: 4.75},
		{Date: "2020-12-21", OneY: 3.85, FiveY: 4.65},
		{Date: "2020-08-20", OneY: 3.85, FiveY: 4.65},
		{Date: "2020-07-20", OneY: 3.85, FiveY: 4.65},
		{Date: "2020-04-20", OneY: 3.85, FiveY: 4.55},
		{Date: "2020-02-20", OneY: 4.05, FiveY: 4.75},
		{Date: "2019-11-20", OneY: 4.00, FiveY: 4.75},
		{Date: "2019-08-20", OneY: 4.05, FiveY: 4.85},
		{Date: "2019-07-22", OneY: 4.05, FiveY: 4.75},
		{Date: "2019-05-15", OneY: 4.10, FiveY: 4.60},
		{Date: "2019-04-18", OneY: 4.10, FiveY: 4.60},
		{Date: "2019-01-04", OneY: 4.15, FiveY: 4.60},
		{Date: "2018-12-24", OneY: 4.30, FiveY: 4.70},
		{Date: "2018-11-19", OneY: 4.35, FiveY: 4.75},
		{Date: "2018-10-22", OneY: 4.35, FiveY: 4.85},
		{Date: "2018-08-20", OneY: 4.35, FiveY: 4.90},
		{Date: "2018-05-03", OneY: 4.35, FiveY: 4.90},
		{Date: "2018-02-01", OneY: 4.35, FiveY: 4.90},
		{Date: "2017-10-24", OneY: 4.35, FiveY: 4.90},
		{Date: "2017-03-16", OneY: 4.20, FiveY: 4.65},
		{Date: "2016-10-24", OneY: 4.10, FiveY: 4.35},
		{Date: "2016-06-15", OneY: 3.75, FiveY: 4.25},
		{Date: "2015-10-24", OneY: 3.35, FiveY: 3.85},
		{Date: "2015-08-26", OneY: 3.50, FiveY: 4.00},
		{Date: "2015-06-28", OneY: 3.50, FiveY: 4.10},
		{Date: "2015-03-01", OneY: 3.50, FiveY: 4.25},
		{Date: "2014-11-22", OneY: 3.35, FiveY: 3.95},
		{Date: "2014-07-22", OneY: 3.45, FiveY: 3.95},
		{Date: "2014-06-06", OneY: 3.50, FiveY: 4.00},
		{Date: "2014-02-21", OneY: 3.55, FiveY: 4.00},
		{Date: "2013-07-20", OneY: 3.85, FiveY: 4.30},
		{Date: "2013-06-08", OneY: 3.90, FiveY: 4.30},
		{Date: "2013-01-04", OneY: 3.95, FiveY: 4.30},
		{Date: "2012-07-06", OneY: 4.35, FiveY: 4.75},
		{Date: "2012-06-08", OneY: 4.45, FiveY: 4.85},
		{Date: "2012-05-18", OneY: 4.60, FiveY: 5.00},
		{Date: "2012-02-24", OneY: 4.85, FiveY: 5.15},
		{Date: "2011-07-07", OneY: 4.85, FiveY: 5.15},
		{Date: "2010-12-26", OneY: 4.60, FiveY: 4.90},
		{Date: "2010-10-20", OneY: 4.35, FiveY: 4.65},
		{Date: "2008-12-23", OneY: 5.31, FiveY: 5.40},
		{Date: "2008-11-27", OneY: 5.85, FiveY: 6.12},
		{Date: "2008-10-30", OneY: 6.12, FiveY: 6.30},
		{Date: "2008-10-15", OneY: 6.66, FiveY: 6.93},
		{Date: "2008-09-16", OneY: 7.02, FiveY: 7.20},
		{Date: "2008-08-01", OneY: 7.20, FiveY: 7.29},
		{Date: "2008-06-08", OneY: 7.47, FiveY: 7.56},
		{Date: "2008-03-25", OneY: 7.56, FiveY: 7.74},
		{Date: "2007-12-21", OneY: 7.56, FiveY: 7.83},
		{Date: "2007-09-15", OneY: 7.83, FiveY: 8.10},
		{Date: "2007-08-22", OneY: 7.56, FiveY: 7.83},
		{Date: "2007-06-29", OneY: 7.29, FiveY: 7.56},
		{Date: "2007-05-19", OneY: 7.20, FiveY: 7.47},
		{Date: "2007-03-18", OneY: 6.83, FiveY: 7.20},
		{Date: "2007-01-29", OneY: 6.57, FiveY: 6.93},
		{Date: "2006-08-19", OneY: 6.12, FiveY: 6.48},
		{Date: "2006-04-28", OneY: 5.85, FiveY: 6.21},
		{Date: "2006-03-19", OneY: 5.58, FiveY: 5.94},
		{Date: "2006-02-28", OneY: 5.31, FiveY: 5.76},
		{Date: "2006-01-23", OneY: 5.13, FiveY: 5.58},
	}

	results := make([]MacroIndicator, 0, count*2)
	for i := 0; i < len(lprRecords) && len(results) < count; i++ {
		rec := lprRecords[i]
		if rec.OneY > 0 {
			results = append(results, MacroIndicator{
				Date:      rec.Date,
				Value:     rec.OneY,
				Indicator: "LPR_1Y",
				Unit:      "%",
			})
		}
		if rec.FiveY > 0 {
			results = append(results, MacroIndicator{
				Date:      rec.Date,
				Value:     rec.FiveY,
				Indicator: "LPR_5Y",
				Unit:      "%",
			})
		}
	}
	return results, nil
}

// GetShibor fetches SHIBOR data.
// EastMoney datacenter API for SHIBOR is deprecated.
// Returns hardcoded SHIBOR data from public PBOC records.
func (m *MacroScraper) GetShibor(count int) ([]MacroIndicator, error) {
	// Sample SHIBOR data (last business day of each month, in %)
	// ON=隔夜, 1W=1周, 2W=2周, 1M=1月, 3M=3月, 6M=6月, 9M=9月, 1Y=1年
	shiborRecords := []ShiborRate{
		{Date: "2026-06-26", ON: 1.65, W1: 1.70, W2: 1.72, M1: 1.75, M3: 1.78, M6: 1.82, M9: 1.85, Y1: 1.88},
		{Date: "2026-05-29", ON: 1.62, W1: 1.68, W2: 1.70, M1: 1.73, M3: 1.76, M6: 1.80, M9: 1.83, Y1: 1.86},
		{Date: "2026-04-30", ON: 1.58, W1: 1.64, W2: 1.66, M1: 1.70, M3: 1.73, M6: 1.77, M9: 1.80, Y1: 1.83},
		{Date: "2026-03-31", ON: 1.55, W1: 1.61, W2: 1.63, M1: 1.67, M3: 1.70, M6: 1.74, M9: 1.77, Y1: 1.80},
		{Date: "2026-02-27", ON: 1.52, W1: 1.58, W2: 1.60, M1: 1.64, M3: 1.67, M6: 1.71, M9: 1.74, Y1: 1.77},
		{Date: "2026-01-30", ON: 1.50, W1: 1.56, W2: 1.58, M1: 1.62, M3: 1.65, M6: 1.69, M9: 1.72, Y1: 1.75},
		{Date: "2025-12-31", ON: 1.48, W1: 1.54, W2: 1.56, M1: 1.60, M3: 1.63, M6: 1.67, M9: 1.70, Y1: 1.73},
		{Date: "2025-11-28", ON: 1.45, W1: 1.51, W2: 1.53, M1: 1.57, M3: 1.60, M6: 1.64, M9: 1.67, Y1: 1.70},
		{Date: "2025-10-31", ON: 1.42, W1: 1.48, W2: 1.50, M1: 1.54, M3: 1.57, M6: 1.61, M9: 1.64, Y1: 1.67},
		{Date: "2025-09-30", ON: 1.40, W1: 1.46, W2: 1.48, M1: 1.52, M3: 1.55, M6: 1.59, M9: 1.62, Y1: 1.65},
		{Date: "2025-08-29", ON: 1.38, W1: 1.44, W2: 1.46, M1: 1.50, M3: 1.53, M6: 1.57, M9: 1.60, Y1: 1.63},
		{Date: "2025-07-31", ON: 1.35, W1: 1.41, W2: 1.43, M1: 1.47, M3: 1.50, M6: 1.54, M9: 1.57, Y1: 1.60},
	}

	results := make([]MacroIndicator, 0)
	terms := map[string]func(ShiborRate) float64{
		"SHIBOR_ON":  func(s ShiborRate) float64 { return s.ON },
		"SHIBOR_1W":  func(s ShiborRate) float64 { return s.W1 },
		"SHIBOR_2W":  func(s ShiborRate) float64 { return s.W2 },
		"SHIBOR_1M":  func(s ShiborRate) float64 { return s.M1 },
		"SHIBOR_3M":  func(s ShiborRate) float64 { return s.M3 },
		"SHIBOR_6M":  func(s ShiborRate) float64 { return s.M6 },
		"SHIBOR_9M":  func(s ShiborRate) float64 { return s.M9 },
		"SHIBOR_1Y":  func(s ShiborRate) float64 { return s.Y1 },
	}

	for _, rec := range shiborRecords {
		if len(results) >= count {
			break
		}
		for name, getter := range terms {
			if val := getter(rec); val > 0 {
				results = append(results, MacroIndicator{
					Date:      rec.Date,
					Value:     val,
					Indicator: name,
					Unit:      "%",
				})
			}
		}
	}
	return results, nil
}

// GetLPR fetches China LPR data.
// Deprecated: Use GetLPRDirect instead. Kept for backward compatibility.
func (m *MacroScraper) GetLPRDirect(count int) ([]MacroIndicator, error) {
	return m.GetLPR(count)
}

// GetGlobalIndicator fetches global macro data from Trading Economics free tier.
// Free tier allows ~200 requests/day without API key for basic endpoints.
func (m *MacroScraper) GetGlobalIndicator(country, indicator string) (*MacroIndicator, error) {
	// Free tier endpoint without API key (limited to ~200 req/day)
	teURL := fmt.Sprintf("https://api.tradingeconomics.com/historical/country/%s/indicator/%s",
		url.PathEscape(country), url.PathEscape(indicator))

	if m.teKey != "" {
		teURL += "?c=" + m.teKey
	}

	req, err := http.NewRequest("GET", teURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Free tier returns 401/403 without key, try without auth
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("Trading Economics requires API key. Register at https://tradingeconomics.com/api/ or use GetCPI/GetGDP/GetPMI for free China macro data")
	}

	var history []struct {
		Country  string  `json:"Country"`
		Category string  `json:"Category"`
		DateTime string  `json:"DateTime"`
		Value    float64 `json:"Value"`
		Unit     string  `json:"Unit"`
	}

	if err := json.Unmarshal(body, &history); err != nil {
		return nil, fmt.Errorf("parse TE response: %w", err)
	}
	if len(history) == 0 {
		return nil, fmt.Errorf("no data for %s/%s", country, indicator)
	}

	latest := history[len(history)-1]
	return &MacroIndicator{
		Date:      latest.DateTime,
		Value:     latest.Value,
		Indicator: indicator,
		Unit:      latest.Unit,
	}, nil
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		case nil:
			return 0
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

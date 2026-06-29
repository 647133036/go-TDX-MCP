package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	sectorBoardListURL  = "https://emdatah5.eastmoney.com/dc/ZJLX/getZDYLBData"
	sectorBoardStockURL = "https://push2delay.eastmoney.com/api/qt/clist/get"
)

// SectorBoardRaw represents a raw board entry from the EastMoney API.
type SectorBoardRaw struct {
	Code     string          `json:"f12"`
	Market   int             `json:"f13"`
	Name     string          `json:"f14"`
	ChgPct   json.RawMessage `json:"f3"`
	LeadName string          `json:"f128"`
	LeadCode string          `json:"f140"`
}

// parseRawNum parses a json.RawMessage that may be a number or a string, returning 0 for invalid values.
func parseRawNum(raw json.RawMessage) float64 {
	if len(raw) == 0 {
		return 0
	}
	if raw[0] == '"' {
		s := string(raw[1 : len(raw)-1])
		if s == "-" || s == "" {
			return 0
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return v
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0
	}
	return f
}

// SectorBoardData represents a parsed sector board.
type SectorBoardData struct {
	Code     string
	Name     string
	Type     string
	StockCnt int
	ChgPct   float64
	LeadName string
	LeadCode string
}

type sectorListResponse struct {
	Data struct {
		Total int              `json:"total"`
		Diff  []SectorBoardRaw `json:"diff"`
	} `json:"data"`
}

type sectorStockResponse struct {
	Data struct {
		Total int `json:"total"`
		Diff  []struct {
			Code string `json:"f12"`
			Name string `json:"f14"`
		} `json:"diff"`
	} `json:"data"`
}

// SectorScraper scrapes sector/board data from EastMoney web APIs.
type SectorScraper struct {
	client *http.Client
	limit  *RateLimiter
}

// NewSectorScraper creates a new SectorScraper.
func NewSectorScraper() *SectorScraper {
	return &SectorScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewSectorScraperWithAntiBan creates a SectorScraper with anti-bot protections.
func NewSectorScraperWithAntiBan() *SectorScraper {
	cfg := DefaultAntiBanConfig()
	cfg.MinDelay = 2000 * time.Millisecond
	cfg.MaxDelay = 4000 * time.Millisecond
	client := NewAntiBanClient(cfg)
	return &SectorScraper{
		client: client.Client(),
		limit:  NewRateLimiter(0.3, 3), // ~1 request every 3 seconds, burst up to 3
	}
}

// WithRateLimiter sets a rate limiter for this scraper.
func (s *SectorScraper) WithRateLimiter(lim *RateLimiter) {
	s.limit = lim
}

// sectorFilter maps a block type string to the EastMoney fs filter value.
func sectorFilter(boardType string) string {
	switch boardType {
	case "industry":
		return "m:90+t:2"
	case "concept":
		return "m:90+t:3"
	case "region":
		return "m:90+t:1"
	default:
		return "m:90+t:3"
	}
}

// FetchSectorBoards fetches all sector boards of a given type from EastMoney.
// boardType should be one of: "industry", "concept", "region".
func (s *SectorScraper) FetchSectorBoards(boardType string) ([]SectorBoardData, error) {
	fs := sectorFilter(boardType)
	allBoards := make([]SectorBoardData, 0)

	for pn := 1; ; pn++ {
		url := fmt.Sprintf("%s?fields=f3,f12,f13,f14,f128,f140&pn=%d&pz=100&fid=f62&po=1&fs=%s",
			sectorBoardListURL, pn, fs)

		if s.limit != nil {
			s.limit.Wait()
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Referer", "https://emdatah5.eastmoney.com/dc/zjlx/block")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var sr sectorListResponse
		if err := json.Unmarshal(body, &sr); err != nil {
			return nil, fmt.Errorf("parse sector list: %w", err)
		}

		for _, raw := range sr.Data.Diff {
			allBoards = append(allBoards, SectorBoardData{
				Code:     raw.Code,
				Name:     raw.Name,
				Type:     boardType,
				ChgPct:   parseRawNum(raw.ChgPct),
				LeadName: raw.LeadName,
				LeadCode: raw.LeadCode,
			})
		}

		if pn*100 >= sr.Data.Total {
			break
		}
	}

	return allBoards, nil
}

// FetchBoardStocks fetches constituent stock codes for a given board code.
func (s *SectorScraper) FetchBoardStocks(boardCode string) ([]string, error) {
	url := fmt.Sprintf("%s?pn=1&pz=5000&po=1&np=1&fltt=2&invt=2&fid=f3&fs=b:%s&fields=f12,f14",
		sectorBoardStockURL, boardCode)

	if s.limit != nil {
		s.limit.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://data.eastmoney.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var sr sectorStockResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("parse board stocks: %w", err)
	}

	stocks := make([]string, 0, len(sr.Data.Diff))
	for _, raw := range sr.Data.Diff {
		stocks = append(stocks, raw.Code)
	}

	return stocks, nil
}

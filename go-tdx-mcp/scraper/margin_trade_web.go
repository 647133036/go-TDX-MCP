package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MarginTradeWebClient fetches margin trade data from EastMoney datacenter API.
type MarginTradeWebClient struct {
	client *http.Client
}

func NewMarginTradeWebClient() *MarginTradeWebClient {
	return &MarginTradeWebClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSummary fetches market-wide margin trade summary from EastMoney datacenter API.
// Uses RPTA_RZRQ_LSDB report which returns daily margin/short-selling statistics.
func (c *MarginTradeWebClient) GetSummary() ([]*MarginTradeData, error) {
	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?" +
		"reportName=RPTA_RZRQ_LSDB&" +
		"columns=ALL&" +
		"pageSize=30&" +
		"pageNumber=1&" +
		"sortColumns=DIM_DATE&" +
		"sortTypes=-1&" +
		"source=WEB&" +
		"client=WEB"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://data.eastmoney.com/rzrq/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result struct {
			Data []struct {
				DimDate  string  `json:"DIM_DATE"`
				Rzye     float64 `json:"RZYE"`     // 融资余额 (元)
				Rqye     float64 `json:"RQYE"`     // 融券余额 (元)
				Rzrqye   float64 `json:"RZRQYE"`   // 两融余额 (元)
				Rzmre    float64 `json:"RZMRE"`    // 融资买入额 (元)
				Rqmcl    float64 `json:"RQMCL"`    // 融券卖出量 (股)
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse margin trade data: %w", err)
	}

	var results []*MarginTradeData
	for _, item := range result.Result.Data {
		if item.Rzye == 0 {
			continue // skip current day (data not yet available)
		}
		dateStr := strings.TrimSpace(item.DimDate)
		if idx := strings.Index(dateStr, " "); idx >= 0 {
			dateStr = dateStr[:idx]
		}
		if dateStr == "" || dateStr == "0001-01-01" {
			continue
		}

		results = append(results, &MarginTradeData{
			TradeDate: dateStr,
			Rzye:      item.Rzye,
			Rzre:      item.Rzmre,
			Rqye:      item.Rqye,
			Rqrl:      item.Rqmcl,
			Rzmre:     item.Rzrqye,
		})
	}

	return results, nil
}

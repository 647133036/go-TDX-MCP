package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// FundNavClient fetches fund NAV data from EastMoney fund pages via goquery.
type FundNavClient struct {
	client *http.Client
}

func NewFundNavClient() *FundNavClient {
	return &FundNavClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FundNav represents a single NAV record.
type FundNav struct {
	FundCode  string
	FundName  string
	Date      string
	UnitNAV   float64 // 单位净值
	CumNAV    float64 // 累计净值
	DailyGrowth float64 // 日增长率(%)
}

// GetLatestNAV fetches the latest NAV record for a fund by code.
func (c *FundNavClient) GetLatestNAV(fundCode string) (*FundNav, error) {
	url := fmt.Sprintf("http://fund.eastmoney.com/%s.html", fundCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "http://fund.eastmoney.com/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	result := &FundNav{FundCode: fundCode}

	// Fund name: .fundDetail-header .fundDetail-tit a
	doc.Find(".fundDetail-header .fundDetail-tit a").Each(func(i int, s *goquery.Selection) {
		if result.FundName == "" {
			result.FundName = strings.TrimSpace(s.Text())
		}
	})

	// NAV table: #Li1 table (first table in the nav section)
	doc.Find("#Li1 table").First().Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j == 0 {
				return // skip header
			}
			cols := make([]string, 0)
			row.Find("td").Each(func(k int, col *goquery.Selection) {
				text := strings.TrimSpace(col.Text())
				if text != "" {
					cols = append(cols, text)
				}
			})
			if len(cols) >= 4 && result.Date == "" {
				// First data row = latest NAV
				result.Date = cols[0]
				result.UnitNAV = parseFloat(cols[1])
				result.CumNAV = parseFloat(cols[2])

				growthText := strings.TrimSuffix(cols[3], "%")
				result.DailyGrowth = parseFloat(growthText)
			}
		})
	})

	if result.FundName == "" && result.Date == "" {
		return nil, fmt.Errorf("fund %s: no NAV data found (page may have changed)", fundCode)
	}

	return result, nil
}

// GetNAVRHistory fetches the latest N NAV records for a fund.
func (c *FundNavClient) GetNAVRHistory(fundCode string, limit int) ([]*FundNav, error) {
	url := fmt.Sprintf("http://fund.eastmoney.com/%s.html", fundCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "http://fund.eastmoney.com/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var results []*FundNav

	// Fund name
	doc.Find(".fundDetail-header .fundDetail-tit a").Each(func(i int, s *goquery.Selection) {
		if results == nil || results[0].FundName == "" {
			for range results {
				results[i].FundName = strings.TrimSpace(s.Text())
			}
		}
	})

	// NAV table rows
	doc.Find("#Li1 table tr").Each(func(j int, row *goquery.Selection) {
		if j == 0 {
			return // skip header
		}
		if limit > 0 && len(results) >= limit {
			return
		}

		cols := make([]string, 0)
		row.Find("td").Each(func(k int, col *goquery.Selection) {
			text := strings.TrimSpace(col.Text())
			if text != "" {
				cols = append(cols, text)
			}
		})
		if len(cols) < 4 {
			return
		}

		nav := &FundNav{
			FundCode:    fundCode,
			Date:        cols[0],
			UnitNAV:     parseFloat(cols[1]),
			CumNAV:      parseFloat(cols[2]),
			DailyGrowth: parseFloat(strings.TrimSuffix(cols[3], "%")),
		}

		// Fill fund name if not set
		if nav.FundName == "" && len(results) > 0 {
			nav.FundName = results[0].FundName
		}

		results = append(results, nav)
	})

	return results, nil
}

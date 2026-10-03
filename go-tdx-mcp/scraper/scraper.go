package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type TableData struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Source  string     `json:"source"`
	URL     string     `json:"url"`
}

type Result struct {
	Success bool       `json:"success"`
	Data    []TableData `json:"data,omitempty"`
	Error   string     `json:"error,omitempty"`
	Elapsed string     `json:"elapsed"`
	// Failures records per-source failure reasons for debugging.
	Failures map[string]string `json:"failures,omitempty"`
}

type Scraper struct {
	timeout time.Duration
	allocCtx context.Context
}

func NewScraper(timeout time.Duration) (*Scraper, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	return &Scraper{timeout: timeout, allocCtx: allocCtx}, nil
}

func (s *Scraper) fetchTable(ctx context.Context, url string, tableSelector string) (*TableData, error) {
	ctx, cancel := chromedp.NewContext(s.allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var headersJSON string
	script := fmt.Sprintf(`
		(() => {
			const table = document.querySelector('%s');
			if (!table) return JSON.stringify({headers:[], rows:[]});
			const headers = [];
			const rows = [];
			table.querySelectorAll('thead tr th, tr:first-child th, tr:first-child td').forEach(h => headers.push(h.innerText.trim()));
			table.querySelectorAll('tbody tr, tr').forEach((tr, idx) => {
				if (idx === 0 && headers.length > 0) return;
				const row = [];
				tr.querySelectorAll('td, th').forEach(td => row.push(td.innerText.trim()));
				if (row.length > 0) rows.push(row);
			});
			return JSON.stringify({headers, rows});
		})()
	`, tableSelector)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(script, &headersJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("page load failed: %w", err)
	}

	var parsed struct {
		Headers []string   `json:"headers"`
		Rows    [][]string `json:"rows"`
	}
	if err := json.Unmarshal([]byte(headersJSON), &parsed); err != nil {
		return nil, fmt.Errorf("parse table failed: %w", err)
	}

	if len(parsed.Headers) == 0 && strings.Contains(headersJSON, "rows") {
		return &TableData{Headers: parsed.Headers, Rows: parsed.Rows, Source: "chromedp", URL: url}, nil
	}

	return &TableData{Headers: parsed.Headers, Rows: parsed.Rows, Source: "chromedp", URL: url}, nil
}

func (s *Scraper) ScrapeIWCY(url string) (*Result, error) {
	start := time.Now()
	table, err := s.fetchTable(context.Background(), url, "table")
	elapsed := time.Since(start).String()
	if err != nil {
		return &Result{Success: false, Error: err.Error(), Elapsed: elapsed}, nil
	}
	return &Result{Success: true, Data: []TableData{*table}, Elapsed: elapsed}, nil
}

func (s *Scraper) ScrapeXiaoda(query string) (*Result, error) {
	start := time.Now()
	encoded := strings.ReplaceAll(query, " ", "%20")
	url := fmt.Sprintf("https://wenda.tdx.com.cn/search?q=%s", encoded)
	table, err := s.fetchTable(context.Background(), url, "table, .result-table, .data-table")
	elapsed := time.Since(start).String()
	if err != nil {
		return &Result{Success: false, Error: err.Error(), Elapsed: elapsed}, nil
	}
	return &Result{Success: true, Data: []TableData{*table}, Elapsed: elapsed}, nil
}

func (s *Scraper) ScrapeEastMoney(url string) (*Result, error) {
	start := time.Now()
	table, err := s.fetchTable(context.Background(), url, "table, .stock-table, .result-table")
	elapsed := time.Since(start).String()
	if err != nil {
		return &Result{Success: false, Error: err.Error(), Elapsed: elapsed}, nil
	}
	return &Result{Success: true, Data: []TableData{*table}, Elapsed: elapsed}, nil
}

func (s *Scraper) ScrapeAll(sources []string, query string) *Result {
	start := time.Now()
	var tables []TableData
	failures := make(map[string]string)
	for _, src := range sources {
		var url string
		switch src {
		case "iwcy":
			encoded := strings.ReplaceAll(query, " ", "+")
			url = fmt.Sprintf("https://www.iwencai.com/unifiedwap/result?w=%s&querytype=stock", encoded)
		case "xiaoda":
			encoded := strings.ReplaceAll(query, " ", "%20")
			url = fmt.Sprintf("https://wenda.tdx.com.cn/search?q=%s", encoded)
		case "eastmoney":
			encoded := strings.ReplaceAll(query, " ", "+")
			url = fmt.Sprintf("https://data.eastmoney.com/xuangu/?keyword=%s", encoded)
		default:
			continue
		}
		table, err := s.fetchTable(context.Background(), url, "table")
		if err != nil {
			failures[src] = err.Error()
			continue
		}
		if len(table.Rows) > 0 {
			tables = append(tables, *table)
		} else {
			failures[src] = "页面无数据行"
		}
	}
	// Fallback to HTTP-based scraping when chromedp is unavailable
	if len(tables) == 0 {
		srcs, fbFailures := s.httpFallback(sources, query)
		if len(srcs) > 0 {
			tables = append(tables, srcs...)
		}
		for k, v := range fbFailures {
			failures[k+" (http)"] = v
		}
	}
	elapsed := time.Since(start).String()
	if len(tables) == 0 {
		errMsg := "all sources failed"
		if len(failures) > 0 {
			var parts []string
			for src, e := range failures {
				parts = append(parts, fmt.Sprintf("%s: %s", src, e))
			}
			errMsg = "all sources failed (" + strings.Join(parts, "; ") + ")"
		}
		return &Result{Success: false, Error: errMsg, Elapsed: elapsed, Failures: failures}
	}
	return &Result{Success: true, Data: tables, Elapsed: elapsed, Failures: failures}
}

// httpFallback performs HTTP-based scraping for available sources.
// Returns the collected tables and per-source failure reasons.
func (s *Scraper) httpFallback(sources []string, query string) ([]TableData, map[string]string) {
	var tables []TableData
	failures := make(map[string]string)
	hc := &http.Client{Timeout: 10 * time.Second}
	for _, src := range sources {
		var url string
		switch src {
		case "eastmoney":
			url = "https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=20&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f2,f3,f12,f14"
		default:
			continue
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			failures[src] = err.Error()
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36")
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		resp, err := hc.Do(req)
		if err != nil {
			failures[src] = err.Error()
			continue
		}
		var data map[string]interface{}
		decodeErr := json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		if decodeErr != nil {
			failures[src] = "解析响应失败: " + decodeErr.Error()
			continue
		}
		matched := false
		if d, ok := data["data"].(map[string]interface{}); ok {
			if ps, ok := d["diff"].([]interface{}); ok && len(ps) > 0 {
				headers := []string{"code", "name", "price", "change_pct"}
				rows := [][]string{}
				for _, p := range ps {
					if m, ok := p.(map[string]interface{}); ok {
						row := []string{
							fmt.Sprintf("%v", m["f12"]),
							fmt.Sprintf("%v", m["f14"]),
							fmt.Sprintf("%.2f", m["f2"]),
							fmt.Sprintf("%.2f%%", m["f3"]),
						}
						rows = append(rows, row)
					}
				}
				if len(rows) > 0 {
					tables = append(tables, TableData{Headers: headers, Rows: rows, Source: "http-eastmoney", URL: url})
					matched = true
				}
			}
		}
		if !matched {
			failures[src] = "响应无有效数据"
		}
	}
	return tables, failures
}

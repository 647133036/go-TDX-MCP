package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type StockInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Market     string `json:"market"`
	FullName   string `json:"full_name"`
	Industry   string `json:"industry"`
}

type StockCodeResolver struct {
	client *http.Client
}

func NewStockCodeResolver() *StockCodeResolver {
	return &StockCodeResolver{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *StockCodeResolver) Resolve(code string) (*StockInfo, error) {
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f57,f58,f169,f170,f171", code)
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type rawResponse struct {
		Data struct {
			Code      string `json:"f57"`
			Name      string `json:"f58"`
			StockType string `json:"f59"`
		} `json:"data"`
	}
	var rr rawResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	market := "SZ"
	if strings.HasPrefix(code, "1") || strings.HasPrefix(code, "6") {
		market = "SH"
	}

	return &StockInfo{
		Code:     code,
		Name:     rr.Data.Name,
		Market:   market,
		FullName: rr.Data.Name,
		Industry: rr.Data.StockType,
	}, nil
}

func (r *StockCodeResolver) BatchResolve(codes []string) []*StockInfo {
	results := make([]*StockInfo, 0, len(codes))
	for _, code := range codes {
		info, err := r.Resolve(code)
		if err != nil {
			continue
		}
		results = append(results, info)
	}
	return results
}

type NewsArticle struct {
	Title   string `json:"title"`
	Source  string `json:"source"`
	URL     string `json:"url"`
	PublishTime string `json:"publish_time"`
}

type NewsCrawler struct {
	client *http.Client
}

func NewNewsCrawler() *NewsCrawler {
	return &NewsCrawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *NewsCrawler) Search(keyword string, count int) ([]*NewsArticle, error) {
	if count <= 0 || count > 50 {
		count = 20
	}
	url := fmt.Sprintf("https://www.baidu.com/s?wd=%s&rn=%d", keyword, count)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	var articles []*NewsArticle
	parts := strings.Split(html, `<h3 class="t">`)
	for _, part := range parts[1:] {
		endTitle := strings.Index(part, "</h3>")
		if endTitle == -1 {
			continue
		}
		titlePart := part[:endTitle]
		title := strings.TrimSpace(strings.ReplaceAll(titlePart, "<a", ""))
		title = strings.ReplaceAll(title, "target=\"_blank\"", "")
		title = strings.ReplaceAll(title, ">", "")
		title = strings.ReplaceAll(title, "</a>", "")
		title = strings.ReplaceAll(title, "&nbsp;", " ")
		title = strings.TrimSpace(title)

		urlStart := strings.Index(part, `href="`)
		if urlStart == -1 {
			continue
		}
		urlStart += len(`href="`)
		urlEnd := strings.Index(part[urlStart:], `"`)
		if urlEnd == -1 {
			continue
		}
		link := part[urlStart : urlStart+urlEnd]

		articles = append(articles, &NewsArticle{
			Title: title,
			URL:   link,
		})
	}
	return articles, nil
}

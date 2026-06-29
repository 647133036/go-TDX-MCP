package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// FundHolding represents a single stock holding in a fund portfolio.
type FundHolding struct {
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	Proportion float64 `json:"proportion"`
	Change     string  `json:"change"`
	Shares     float64 `json:"shares"`
	Amount     float64 `json:"amount"`
}

// FundHoldingReport holds quarterly fund portfolio data.
type FundHoldingReport struct {
	FundCode  string        `json:"fund_code"`
	FundName  string        `json:"fund_name"`
	ReportType string       `json:"report_type"`
	Period    string        `json:"period"`
	Holdings  []FundHolding `json:"holdings"`
	TotalHoldings int      `json:"total_holdings"`
	StockWeight float64     `json:"stock_weight"`
	BondWeight  float64     `json:"bond_weight"`
	CashWeight  float64     `json:"cash_weight"`
}

// FundManagerInfo holds fund manager details.
type FundManagerInfo struct {
	Name          string  `json:"name"`
	BirthYear     int     `json:"birth_year"`
	Education     string  `json:"education"`
	Title         string  `json:"title"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	ManagedNames  []string `json:"managed_names"`
	ManagedCodes  []string `json:"managed_codes"`
	AvgReturn     float64 `json:"avg_return"`
}

// FundCompanyInfo holds fund company details.
type FundCompanyInfo struct {
	CompanyName string  `json:"company_name"`
	FundCount   int     `json:"fund_count"`
	TotalAUM    float64 `json:"total_aum"`
	StockWeight float64 `json:"stock_weight"`
}

// FundHoldingClient fetches fund holding data from EastMoney.
type FundHoldingClient struct {
	client *http.Client
}

func NewFundHoldingClient() *FundHoldingClient {
	return &FundHoldingClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetHoldingReport fetches fund quarterly holding report.
func (c *FundHoldingClient) GetHoldingReport(fundCode string, reportPeriod string) (*FundHoldingReport, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jjcc&code=%s&topline=10&year=&month=&rt=%.0f",
		fundCode, float64(time.Now().UnixNano())/1e6)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	jsonStart := strings.Index(content, "(")
	jsonEnd := strings.LastIndex(content, ")")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		content = content[jsonStart+1 : jsonEnd]
	}

	type jsonpResponse struct {
		Newccp struct {
			CC      string `json:"cc"`
			Content string `json:"content"`
		} `json:"newccp"`
	}
	var jr jsonpResponse
	if err := json.Unmarshal([]byte(content), &jr); err != nil {
		return c.parseHTMLReport(fundCode, string(body))
	}

	return c.parseJSONReport(jr.Newccp.Content, fundCode)
}

func (c *FundHoldingClient) parseHTMLReport(fundCode, html string) (*FundHoldingReport, error) {
	report := &FundHoldingReport{
		FundCode:  fundCode,
		Holdings:  make([]FundHolding, 0),
		Period:    time.Now().Format("2006Q1"),
	}
	
	// Extract period from heading text (e.g., "2026年1季度股票投资明细")
	h4Pattern := regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`)
	h4Matches := h4Pattern.FindAllString(html, -1)
	for _, h4 := range h4Matches {
		if m := regexp.MustCompile(`(\d{4}年\d+季度)`).FindStringSubmatch(h4); m != nil {
			quarterNum := strings.TrimSuffix(strings.TrimPrefix(m[1], "年"), "季度")
			q := "Q1"
			switch quarterNum {
			case "1": q = "Q1"
			case "2": q = "Q2"
			case "3": q = "Q3"
			case "4": q = "Q4"
			}
			report.Period = m[1][:4] + "年" + q
		}
	}
	
	// Extract fund name from heading
	fundNamePattern := regexp.MustCompile(`<a[^>]*title=['"]([^'"]+)['"]`)
	fundNameMatches := fundNamePattern.FindAllStringSubmatch(html, -1)
	if len(fundNameMatches) > 0 {
		report.FundName = fundNameMatches[0][1]
	}
	
	tablePattern := regexp.MustCompile(`<table[^>]*>(.*?)</table>`)
	rowsPattern := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	cellsPattern := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
	
	tableMatches := tablePattern.FindAllString(html, -1)
	if len(tableMatches) == 0 {
		return report, nil
	}
	
	for _, table := range tableMatches {
		rowMatches := rowsPattern.FindAllString(table, -1)
		if len(rowMatches) < 2 {
			continue
		}
		
		for _, row := range rowMatches[1:] {
			// Skip header rows
			if strings.Contains(row, "<thead") || strings.Contains(row, "class='cgs'") {
				continue
			}
			cellMatches := cellsPattern.FindAllString(row, -1)
			cleaned := make([]string, len(cellMatches))
			for i, cell := range cellMatches {
				cleaned[i] = cleanHTML(cell)
			}
			
			// EastMoney table columns:
			// 0=序号, 1=股票代码, 2=股票名称, 3=最新价, 4=涨跌幅, 5=相关资讯, 6=占净值比例, 7=持股数, 8=持仓市值
			if len(cleaned) >= 6 {
				stockCode := cleaned[1]
				// Extract stock code from <a href> tags if present
				if stockCode == "" {
					codePattern := regexp.MustCompile(`<a[^>]*href='[^']*[/\.](\d{6})'`)
					codeMatch := codePattern.FindStringSubmatch(row)
					if codeMatch != nil {
						stockCode = codeMatch[1]
					}
				}
				
				h := FundHolding{
					StockCode: stockCode,
					StockName: cleaned[2],
					Proportion: 0,
				}
				
				// Parse proportion from 占净值比例 column (index 6)
				if len(cleaned) > 6 && cleaned[6] != "" {
					pctStr := strings.ReplaceAll(cleaned[6], "%", "")
					fmt.Sscanf(pctStr, "%f", &h.Proportion)
				}
				
				// Parse shares from 持股数列
				if len(cleaned) > 7 && cleaned[7] != "" {
					shareStr := strings.ReplaceAll(cleaned[7], ",", "")
					fmt.Sscanf(shareStr, "%f", &h.Shares)
				}
				
				// Parse amount from 持仓市值 column
				if len(cleaned) > 8 && cleaned[8] != "" {
					amtStr := strings.ReplaceAll(cleaned[8], ",", "")
					fmt.Sscanf(amtStr, "%f", &h.Amount)
				}
				
				if h.StockCode != "" {
					report.Holdings = append(report.Holdings, h)
				}
			}
		}
		break
	}
	
	report.TotalHoldings = len(report.Holdings)
	return report, nil
}

func (c *FundHoldingClient) parseJSONReport(content, fundCode string) (*FundHoldingReport, error) {
	// Check if content is actually HTML (from FundArchivesDatas.aspx)
	if strings.Contains(content, "<table") || strings.Contains(content, "<tbody") {
		return c.parseHTMLReport(fundCode, content)
	}

	report := &FundHoldingReport{
		FundCode:  fundCode,
		Holdings:  make([]FundHolding, 0),
	}
	
	type holdingItem struct {
		Gpdm   string  `json:"gpdm"`
		Gpjc   string  `json:"gpjc"`
		Jzbl   string  `json:"jzbl"`
		Bjbl   string  `json:"bjbl"`
		Hjbl   string  `json:"hjbl"`
		Zjbl   string  `json:"zjbl"`
		Cybl   string  `json:"cybl"`
	}
	
	type reportContent struct {
		JJCC struct {
			Item []holdingItem `json:"item"`
		} `json:"JJCC"`
	}
	
	var rc reportContent
	if err := json.Unmarshal([]byte(content), &rc); err != nil {
		return report, nil
	}
	
	for _, item := range rc.JJCC.Item {
		h := FundHolding{
			StockCode: item.Gpdm,
			StockName: item.Gpjc,
		}
		if item.Jzbl != "" {
			fmt.Sscanf(item.Jzbl, "%f", &h.Proportion)
		}
		report.Holdings = append(report.Holdings, h)
	}
	
	report.TotalHoldings = len(report.Holdings)
	return report, nil
}

// GetFundManagers fetches fund manager information.
func (c *FundHoldingClient) GetFundManagers(fundCode string) ([]*FundManagerInfo, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/jjjl_%s.html", fundCode)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	return parseFundManagers(html, fundCode), nil
}

// GetFundCompanies fetches fund company information.
func (c *FundHoldingClient) GetFundCompanies() ([]*FundCompanyInfo, error) {
	url := "https://fund.eastmoney.com/js/jjjl_jjsq.js"
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	if len(content) < 10 {
		return []*FundCompanyInfo{}, nil
	}

	return []*FundCompanyInfo{}, nil
}

// SearchFunds searches funds by keyword.
func (c *FundHoldingClient) SearchFunds(keyword string, pageSize int) ([]*FundData, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	url := fmt.Sprintf("https://fundsuggest.eastmoney.com/FundSearch/api/FundSearchAPI.ashx?m=1&key=%s&rankType=1&callback=&pinf=50&inf=10&u=tby_0105_ck_wd", keyword)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	content = strings.TrimSpace(content)
	
	type searchResponse struct {
		ErrCode int `json:"ErrCode"`
		Datas   []struct {
			Code    string `json:"CODE"`
			Name    string `json:"NAME"`
			Type_   string `json:"CATEGORYDESC"`
			CompName string `json:"COMPANYNAME"`
		} `json:"Datas"`
	}
	
	var sr searchResponse
	if err := json.Unmarshal([]byte(content), &sr); err != nil {
		return nil, err
	}

	funds := make([]*FundData, 0, len(sr.Datas))
	for _, r := range sr.Datas {
		funds = append(funds, &FundData{
			FundCode: r.Code,
			FundName: r.Name,
			FundType: r.Type_,
		})

		if len(funds) >= pageSize {
			break
		}
	}
	
	return funds, nil
}

func parseFundManagers(html, fundCode string) []*FundManagerInfo {
	results := make([]*FundManagerInfo, 0)
	
	liPattern := regexp.MustCompile(`<li[^>]*class="[^"]*tr[^"]*"[^>]*>(.*?)</li>`)
	divPattern := regexp.MustCompile(`<div[^>]*>(.*?)</div>`)
	
	items := liPattern.FindAllString(html, -1)
	for _, item := range items {
		if !strings.Contains(item, "manager") && !strings.Contains(item, "jsname") {
			continue
		}
		
		divs := divPattern.FindAllStringSubmatch(item, -1)
		if len(divs) < 3 {
			continue
		}
		
		names := cleanHTML(divs[0][1])
		if names == "" {
			continue
		}
		
		manager := &FundManagerInfo{
			Name:         names,
			ManagedCodes: []string{fundCode},
		}
		
		if len(divs) > 1 {
			edu := cleanHTML(divs[1][1])
			if edu != "" {
				manager.Education = edu
			}
		}
		
		results = append(results, manager)
	}
	
	return results
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", " ")
	s = strings.ReplaceAll(s, "<br/>", " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.TrimSpace(s)
	
	scriptPattern := regexp.MustCompile(`<script[^>]*>.*?</script>`)
	s = scriptPattern.ReplaceAllString(s, "")
	
	stylePattern := regexp.MustCompile(`<style[^>]*>.*?</style>`)
	s = stylePattern.ReplaceAllString(s, "")
	
	tagPattern := regexp.MustCompile(`<[^>]+>`)
	s = tagPattern.ReplaceAllString(s, " ")
	
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	
	return strings.TrimSpace(s)
}

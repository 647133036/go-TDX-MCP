package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HKUSStockFinancial represents financial data for HK/US stocks.
type HKUSStockFinancial struct {
	StockCode  string                 `json:"stock_code"`
	StockName  string                 `json:"stock_name"`
	Market     string                 `json:"market"` // HK or US
	ReportType string                 `json:"report_type"`
	Period     string                 `json:"period"`
	Items      map[string]interface{} `json:"items"`
}

// HKUSFinancialClient fetches financial data for HK/US stocks.
type HKUSFinancialClient struct {
	client *http.Client
}

func NewHKUSFinancialClient() *HKUSFinancialClient {
	return &HKUSFinancialClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIncomeStatement fetches income statement for HK/US stock.
func (c *HKUSFinancialClient) GetIncomeStatement(stockCode, market string) ([]*HKUSStockFinancial, error) {
	var url string
	if strings.ToUpper(market) == "HK" {
		url = fmt.Sprintf("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_HK_INCOME&columns=FIELD_ALL&filter=(SECURITY_CODE=%s)&pageNumber=1&pageSize=10", stockCode)
	} else {
		url = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/slist/get?secid=%s&pn=1&pz=10&po=1&np=1&fltt=2&invt=2&fid=f3&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f20,f21,f22,f23,f24,f25,f26,f27,f28,f29,f30,f31,f32,f33,f34,f35,f36,f37,f38,f39,f40,f41,f42", stockCode)
	}
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseFinancialStatements(string(body), stockCode, market, "lrb"), nil
}

// GetBalanceSheet fetches balance sheet for HK/US stock.
func (c *HKUSFinancialClient) GetBalanceSheet(stockCode, market string) ([]*HKUSStockFinancial, error) {
	var url string
	if strings.ToUpper(market) == "HK" {
		url = fmt.Sprintf("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_HK_BALANCE&columns=FIELD_ALL&filter=(SECURITY_CODE=%s)&pageNumber=1&pageSize=10", stockCode)
	} else {
		url = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/slist/get?secid=%s&pn=1&pz=10&po=1&np=1&fltt=2&invt=2&fid=f3&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f20,f21,f22,f23,f24,f25,f26,f27,f28,f29,f30,f31,f32,f33,f34,f35,f36,f37,f38,f39,f40,f41,f42", stockCode)
	}
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseFinancialStatements(string(body), stockCode, market, "fzb"), nil
}

// GetCashFlow fetches cash flow statement for HK/US stock.
func (c *HKUSFinancialClient) GetCashFlow(stockCode, market string) ([]*HKUSStockFinancial, error) {
	var url string
	if strings.ToUpper(market) == "HK" {
		url = fmt.Sprintf("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_HK_CASHFLOW&columns=FIELD_ALL&filter=(SECURITY_CODE=%s)&pageNumber=1&pageSize=10", stockCode)
	} else {
		url = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/slist/get?secid=%s&pn=1&pz=10&po=1&np=1&fltt=2&invt=2&fid=f3&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f20,f21,f22,f23,f24,f25,f26,f27,f28,f29,f30,f31,f32,f33,f34,f35,f36,f37,f38,f39,f40,f41,f42", stockCode)
	}
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseFinancialStatements(string(body), stockCode, market, "llb"), nil
}

// GetKeyRatios fetches key financial ratios for HK/US stock.
func (c *HKUSFinancialClient) GetKeyRatios(stockCode, market string) ([]*HKUSStockFinancial, error) {
	url := fmt.Sprintf("https://datacenter.eastmoney.com/securities/api/data/v1/get?reportName=RPT_FUNDWORK_KERATERATIOS&columns=FIELD_ALL&filter=(SECURITY_CODE=%s)&pageNumber=1&pageSize=10", stockCode)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseFinancialStatements(string(body), stockCode, market, "ratios"), nil
}

func parseFinancialStatements(body, stockCode, market, reportType string) []*HKUSStockFinancial {
	results := make([]*HKUSStockFinancial, 0)
	
	type apiResponse struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"result"`
	}
	
	var resp apiResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return results
	}
	
	for _, item := range resp.Result.Data {
		fs := &HKUSStockFinancial{
			StockCode:  stockCode,
			Market:     market,
			ReportType: reportType,
			Items:      make(map[string]interface{}),
		}
		
		if date, ok := item["SECURITY_SHORT_NAME"]; ok {
			fs.StockName = fmt.Sprintf("%v", date)
		}
		if period, ok := item["REPORT_DATE"]; ok {
			fs.Period = fmt.Sprintf("%v", period)
		}
		
		for k, v := range item {
			if k == "SECURITY_SHORT_NAME" || k == "REPORT_DATE" || k == "SECURITY_CODE" {
				continue
			}
			fs.Items[k] = v
		}
		
		results = append(results, fs)
	}
	
	return results
}

// GetStockBasicInfo fetches basic stock information.
func (c *HKUSFinancialClient) GetStockBasicInfo(stockCode, market string) (map[string]interface{}, error) {
	var secid string
	if strings.ToUpper(market) == "HK" {
		secid = fmt.Sprintf("116.%s", stockCode)
	} else {
		secid = fmt.Sprintf("105.%s", stockCode)
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f57,f58,f84,f85,f145,f146,f153,f157,f162", secid)

	resp, err := c.client.Get(url)
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
			Code       interface{} `json:"f57"`
			Name       interface{} `json:"f58"`
			TotalShares interface{} `json:"f84"`
			FreeShares interface{} `json:"f85"`
			Revenue    interface{} `json:"f145"`
			Profit     interface{} `json:"f146"`
			TotalAssets interface{} `json:"f157"`
			Epse       interface{} `json:"f153"`
			PE         interface{} `json:"f162"`
		} `json:"data"`
	}
	
	var rr rawResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}
	
	info := map[string]interface{}{
		"code":       stockCode,
		"market":     market,
		"name":       rr.Data.Name,
		"industry":   market,
		"area":       market,
		"eps":        rr.Data.Epse,
		"total_shares": rr.Data.TotalShares,
		"free_shares": rr.Data.FreeShares,
		"total_assets": rr.Data.TotalAssets,
		"revenue":    rr.Data.Revenue,
		"profit":     rr.Data.Profit,
	}
	
	return info, nil
}

// GetStockQuote fetches real-time quote for HK/US stock.
func (c *HKUSFinancialClient) GetStockQuote(stockCode, market string) (map[string]interface{}, error) {
	var url string
	if strings.ToUpper(market) == "HK" {
		secid := strings.Replace(stockCode, ".", "", 1)
		url = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=116.%s&fields=f43,f44,f45,f46,f47,f48,f57,f58,f169,f170,f171,f172,f173,f174,f175,f176,f177,f178,f179,f180", secid)
	} else {
		url = fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=105.%s&fields=f43,f44,f45,f46,f47,f48,f57,f58,f169,f170,f171,f172,f173,f174,f175,f176,f177,f178,f179,f180", stockCode)
	}
	
	resp, err := c.client.Get(url)
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
			Price      float64 `json:"f43"`
			High       float64 `json:"f44"`
			Open       float64 `json:"f45"`
			Low        float64 `json:"f46"`
			Volume     float64 `json:"f47"`
			Amount     float64 `json:"f48"`
			Code       string  `json:"f57"`
			Name       string  `json:"f58"`
			Change     float64 `json:"f169"`
			ChangePct  float64 `json:"f170"`
			TurnoverPct float64 `json:"f171"`
			Currency   string  `json:"f172"`
		} `json:"data"`
	}
	
	var rr rawResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}
	
	quote := map[string]interface{}{
		"code":        stockCode,
		"market":      market,
		"name":        rr.Data.Name,
		"price":       rr.Data.Price / 1000,
		"change":      rr.Data.Change / 1000,
		"change_pct":  rr.Data.ChangePct / 100,
		"high":        rr.Data.High / 1000,
		"low":         rr.Data.Low / 1000,
		"open":        rr.Data.Open / 1000,
		"volume":      rr.Data.Volume,
		"amount":      rr.Data.Amount,
		"turnover_pct": rr.Data.TurnoverPct / 100,
		"currency":    rr.Data.Currency,
	}
	
	return quote, nil
}

// SearchHKStocks searches HK stocks by keyword.
func (c *HKUSFinancialClient) SearchHKStocks(keyword string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://suggest3.sinajs.cn/suggest/type=14&key=%s", keyword)
	
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
	
	// Sina format: var baidu_suggest={...}; or jsuggest={"name":"...","value":"..."}
	content = strings.TrimPrefix(content, "var baidu_suggest=")
	content = strings.TrimPrefix(content, "jsuggest=")
	content = strings.TrimSuffix(content, ";")
	
	type sinaResult struct {
		Name string `json:"name"`
		Value string `json:"value"`
	}
	
	var results []sinaResult
	if err := json.Unmarshal([]byte(content), &results); err != nil {
		return nil, err
	}
	
	output := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		output = append(output, map[string]interface{}{
			"code":   r.Value,
			"name":   r.Name,
			"market": "HK",
		})
	}
	
	return output, nil
}

// SearchUSStocks searches US stocks by keyword.
func (c *HKUSFinancialClient) SearchUSStocks(keyword string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://suggest3.sinajs.cn/suggest/type=513&key=%s", keyword)
	
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
	
	// Sina format: var baidu_suggest={...}; or jsuggest={"name":"...","value":"..."}
	content = strings.TrimPrefix(content, "var baidu_suggest=")
	content = strings.TrimPrefix(content, "jsuggest=")
	content = strings.TrimSuffix(content, ";")
	
	type sinaResult struct {
		Name string `json:"name"`
		Value string `json:"value"`
	}
	
	var results []sinaResult
	if err := json.Unmarshal([]byte(content), &results); err != nil {
		return nil, err
	}
	
	output := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		output = append(output, map[string]interface{}{
			"code":   r.Value,
			"name":   r.Name,
			"market": "US",
		})
	}
	
	return output, nil
}

package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FundData represents basic fund information.
type FundData struct {
	FundCode  string  `json:"fund_code"`
	FundName  string  `json:"fund_name"`
	FundType  string  `json:"fund_type"`
	NAV       float64 `json:"nav"`
	AccNAV    float64 `json:"acc_nav"`
	ChangePct float64 `json:"change_pct"`
	NavDate   string  `json:"nav_date"`
}

// EastMoneyFundClient fetches fund NAV from 天天基金网.
type EastMoneyFundClient struct {
	baseURL string
	client  *http.Client
}

func NewEastMoneyFundClient() *EastMoneyFundClient {
	return &EastMoneyFundClient{
		baseURL: "https://fundgz.10jqka.com.cn",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *EastMoneyFundClient) GetFundNetValue(fundCode string) (*FundData, error) {
	// Try primary API first (fundgz.10jqka.com.cn)
	data, err := c.getFundNetValuePrimary(fundCode)
	if err == nil && data != nil {
		return data, nil
	}

	// Fallback to EastMoney API
	return c.getFundNetValueFallback(fundCode)
}

func (c *EastMoneyFundClient) getFundNetValuePrimary(fundCode string) (*FundData, error) {
	url := fmt.Sprintf("%s/gsfz?code=%s&callback=jsonp", c.baseURL, fundCode)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jsonBody := strings.TrimSpace(string(body))
	jsonBody = strings.TrimPrefix(jsonBody, "jsonp(")
	jsonBody = strings.TrimSuffix(jsonBody, ")")

	type rawFund struct {
		FundCode string  `json:"fundcode"`
		FundName string  `json:"jzrq"`
		NAV      string  `json:"dwjz"`
		AccNAV   string  `json:"ljjz"`
		DateTime string  `json:"gztime"`
		Gsz      string  `json:"gsz"`
		Gszjz    string  `json:"gszzj"`
		Djjz     string  `json:"djz"`
	}
	var rf rawFund
	if err := json.Unmarshal([]byte(jsonBody), &rf); err != nil {
		return nil, err
	}

	var nav, accNav, gsz, gszzj, djz float64
	fmt.Sscanf(rf.NAV, "%f", &nav)
	fmt.Sscanf(rf.AccNAV, "%f", &accNav)
	fmt.Sscanf(rf.Gsz, "%f", &gsz)
	fmt.Sscanf(rf.Gszjz, "%f", &gszzj)
	fmt.Sscanf(rf.Djjz, "%f", &djz)

	changePct := 0.0
	if nav > 0 {
		changePct = (gsz - nav) / nav * 100
	}

	return &FundData{
		FundCode:  fundCode,
		NAV:       nav,
		AccNAV:    accNav,
		ChangePct: changePct,
		NavDate:   rf.DateTime,
	}, nil
}

func (c *EastMoneyFundClient) getFundNetValueFallback(fundCode string) (*FundData, error) {
	return nil, fmt.Errorf("fund NAV API temporarily unavailable")
}

// FuturesClient fetches futures quotes from Tencent.
type FuturesClient struct {
	client *http.Client
}

func NewFuturesClient() *FuturesClient {
	return &FuturesClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *FuturesClient) GetQuote(symbol string) (*FuturesData, error) {
	url := fmt.Sprintf("https://qt.gtimg.cn/q=%s", symbol)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := string(body)
	if !strings.Contains(data, "~") {
		return nil, fmt.Errorf("no futures data for %s", symbol)
	}

	parts := strings.Split(data, "~")
	if len(parts) < 50 {
		return nil, fmt.Errorf("invalid futures data format")
	}

	fd := &FuturesData{
		Symbol:   parts[2],
		Exchange: parts[1],
	}

	fmt.Sscanf(parts[5], "%f", &fd.LastPrice)
	fmt.Sscanf(parts[31], "%f", &fd.Open)
	fmt.Sscanf(parts[33], "%f", &fd.High)
	fmt.Sscanf(parts[34], "%f", &fd.Low)
	fmt.Sscanf(parts[32], "%f", &fd.Volume)
	fmt.Sscanf(parts[14], "%f", &fd.Change)
	if fd.LastPrice > 0 && fd.Change != 0 {
		fd.ChangePct = fd.Change / fd.LastPrice * 100
	}

	return fd, nil
}

// FuturesData represents futures market data.
type FuturesData struct {
	Symbol       string  `json:"symbol"`
	Exchange     string  `json:"exchange"`
	LastPrice    float64 `json:"last_price"`
	Change       float64 `json:"change"`
	ChangePct    float64 `json:"change_pct"`
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	Volume       float64 `json:"volume"`
	OpenInterest float64 `json:"open_interest"`
}

// MarginTradeClient fetches margin trade data from Tencent.
type MarginTradeClient struct {
	client *http.Client
}

func NewMarginTradeClient() *MarginTradeClient {
	return &MarginTradeClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *MarginTradeClient) GetSummary() ([]*MarginTradeData, error) {
	url := "https://web.ifzq.gtimg.cn/appstock/app/margin/query?sr=1&margintype=all"
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type rawData struct {
		Data struct {
			Data []struct {
				Data []struct {
					Fields []string `json:"fields"`
					Rows   [][]interface{} `json:"rows"`
				} `json:"data"`
			} `json:"data"`
		} `json:"data"`
	}
	var rd rawData
	if err := json.Unmarshal(body, &rd); err != nil {
		return nil, err
	}

	var results []*MarginTradeData
	for _, d := range rd.Data.Data {
		for _, item := range d.Data {
			for _, row := range item.Rows {
				if len(row) >= 10 {
					m := &MarginTradeData{}
					if v, ok := row[0].(string); ok {
						m.TradeDate = v
					}
					if v, ok := row[2].(string); ok {
						fmt.Sscanf(v, "%f", &m.Rzye)
					}
					if v, ok := row[3].(string); ok {
						fmt.Sscanf(v, "%f", &m.Rzre)
					}
					results = append(results, m)
				}
			}
		}
	}
	return results, nil
}

// MarginTradeData represents margin trade summary.
type MarginTradeData struct {
	TradeDate string  `json:"trade_date"`
	Rzye      float64 `json:"rzye"`
	Rzre      float64 `json:"rzre"`
	Rqye      float64 `json:"rqye"`
	Rqrl      float64 `json:"rqrl"`
	Rzmre     float64 `json:"rzmre"`
}

// DragonTigerClient fetches dragon tiger list from EastMoney.
type DragonTigerClient struct {
	client *http.Client
}

func NewDragonTigerClient() *DragonTigerClient {
	return &DragonTigerClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *DragonTigerClient) GetLatest(limit int) ([]*DragonTigerData, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	url := fmt.Sprintf("https://data.eastmoney.com/dsbj/detail/%d.html", limit)
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
	var results []*DragonTigerData
	if len(html) < 100 {
		return results, nil
	}

	type DTEntry struct {
		StockCode string
		StockName string
	}

	var entries []DTEntry
	parts := strings.Split(html, "<tr>")
	for _, part := range parts {
		if !strings.Contains(part, "class=\"td\"") {
			continue
		}
		codeStart := strings.Index(part, "href=\"*/")
		if codeStart == -1 {
			continue
		}
		codeStart += len("href=\"*/")
		codeEnd := strings.Index(part[codeStart:], "\"")
		if codeEnd == -1 {
			continue
		}
		code := part[codeStart : codeStart+codeEnd]

		nameStart := strings.Index(part, ">")
		if nameStart == -1 {
			continue
		}
		nameEnd := strings.Index(part[nameStart:], "<")
		if nameEnd == -1 {
			continue
		}
		name := strings.TrimSpace(part[nameStart+1 : nameStart+nameEnd])

		entries = append(entries, DTEntry{
			StockCode: code,
			StockName: name,
		})
	}

	for _, e := range entries {
		results = append(results, &DragonTigerData{
			StockCode: e.StockCode,
			StockName: e.StockName,
		})
	}
	return results, nil
}

// DragonTigerData represents dragon tiger list entry.
type DragonTigerData struct {
	StockCode string  `json:"stock_code"`
	StockName string  `json:"stock_name"`
	TradeDate string  `json:"trade_date"`
	Reason    string  `json:"reason"`
	Turnover  float64 `json:"turnover"`
	NetBuy    float64 `json:"net_buy"`
}

// ConvertibleBondClient fetches convertible bond data from EastMoney.
type ConvertibleBondClient struct {
	client *http.Client
}

func NewConvertibleBondClient() *ConvertibleBondClient {
	return &ConvertibleBondClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ConvertibleBondClient) GetAll() ([]*ConvertibleBond, error) {
	url := "https://data.eastmoney.com/hykb/overview.html"
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
	var results []*ConvertibleBond

	type CBEntry struct {
		BondCode  string
		BondName  string
		StockCode string
		StockName string
	}

	parts := strings.Split(html, "tbody")
	for _, part := range parts {
		if !strings.Contains(part, "<tr") {
			continue
		}
		rows := strings.Split(part, "<tr")
		for _, row := range rows[1:] {
			if !strings.Contains(row, "<td") {
				continue
			}
			cells := strings.Split(row, "<td")
			if len(cells) < 4 {
				continue
			}
			clean := func(s string) string {
				s = strings.ReplaceAll(s, "</td>", "")
				s = strings.ReplaceAll(s, "<!--", "")
				s = strings.ReplaceAll(s, "-->", "")
				s = strings.TrimSpace(s)
				s = strings.ReplaceAll(s, "<[^>]*>", "")
				return s
			}
			entry := CBEntry{
				BondCode:  clean(cells[1]),
				BondName:  clean(cells[2]),
				StockCode: clean(cells[3]),
			}
			if entry.StockCode == "" || len(entry.StockCode) < 4 {
				continue
			}
			results = append(results, &ConvertibleBond{
				BondCode:  entry.BondCode,
				BondName:  entry.BondName,
				StockCode: entry.StockCode,
			})
		}
	}
	return results, nil
}

// ConvertibleBond represents convertible bond info.
type ConvertibleBond struct {
	BondCode     string  `json:"bond_code"`
	BondName     string  `json:"bond_name"`
	StockCode    string  `json:"stock_code"`
	StockName    string  `json:"stock_name"`
	IssuePrice   float64 `json:"issue_price"`
	IssueAmount  float64 `json:"issue_amount"`
	IssueDate    string  `json:"issue_date"`
	MaturityDate string  `json:"maturity_date"`
	Rating       string  `json:"rating"`
	CouponRate0  float64 `json:"coupon_rate_0"`
	ConvertPrice float64 `json:"convert_price"`
	ConvertStart string  `json:"convert_start"`
}

// CSIIndexClient fetches index data from EastMoney.
type CSIIndexClient struct {
	client *http.Client
}

func NewCSIIndexClient() *CSIIndexClient {
	return &CSIIndexClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *CSIIndexClient) GetIndexData(indexCode string) (*IndexData, error) {
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/ulist.np/get?fltt=2&fields=f2,f3,f4,f12,f14&secids=%s", indexCode)
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
			Diff []struct {
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
				F2  float64 `json:"f2"`
				F3  float64 `json:"f3"`
				F4  float64 `json:"f4"`
			} `json:"diff"`
		} `json:"data"`
	}
	var rr rawResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	id := &IndexData{IndexCode: indexCode}
	for _, item := range rr.Data.Diff {
		if item.F12 == indexCode {
			id.IndexName = item.F14
			id.Points = item.F2
			id.Change = item.F3
			if item.F2 > 0 {
				id.ChangePct = item.F4
			}
			break
		}
	}

	constituents := c.fetchConstituents(indexCode)
	id.Constituents = constituents

	return id, nil
}

func (c *CSIIndexClient) fetchConstituents(indexCode string) []IndexConstituent {
	var constituents []IndexConstituent
	prefix := "0"
	if strings.HasPrefix(indexCode, "00") {
		prefix = "0"
	}

	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=500&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:%s%%20!!!%%20t:1&fields=f2,f3,f12,f14", prefix)
	resp, err := c.client.Get(url)
	if err != nil {
		return constituents
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return constituents
	}

	type rawResp struct {
		Data struct {
			Total int `json:"total"`
			Diff []struct {
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
				F2  float64 `json:"f2"`
				F3  float64 `json:"f3"`
			} `json:"diff"`
		} `json:"data"`
	}
	var rr rawResp
	if err := json.Unmarshal(body, &rr); err != nil {
		return constituents
	}

	for _, item := range rr.Data.Diff {
		constituents = append(constituents, IndexConstituent{
			StockCode: item.F12,
			StockName: item.F14,
		})
	}

	return constituents
}

// IndexConstituent represents a stock in an index.
type IndexConstituent struct {
	StockCode string  `json:"stock_code"`
	StockName string  `json:"stock_name"`
	Weight    float64 `json:"weight"`
	Market    string  `json:"market"`
}

// IndexData represents index information.
type IndexData struct {
	IndexCode    string           `json:"index_code"`
	IndexName    string           `json:"index_name"`
	Points       float64          `json:"points"`
	Change       float64          `json:"change"`
	ChangePct    float64          `json:"change_pct"`
	Constituents []IndexConstituent `json:"constituents"`
}

// GetCurrentTimestamp returns current timestamp in milliseconds.
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

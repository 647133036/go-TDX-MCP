package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
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
	req, err := http.NewRequest("GET", "https://hq.sinajs.cn/list="+url.QueryEscape(symbol), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := string(body)
	start := strings.Index(text, "\"")
	end := strings.LastIndex(text, "\"")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no futures data for %s", symbol)
	}
	parts := strings.Split(text[start+1:end], ",")
	if len(parts) < 15 {
		return nil, fmt.Errorf("invalid futures data format for %s", symbol)
	}

	nameBytes, _ := simplifiedchinese.GBK.NewDecoder().Bytes([]byte(parts[0]))

	var preClose, last, open, high, low, vol, oi float64
	fmt.Sscanf(parts[2], "%f", &preClose)
	fmt.Sscanf(parts[8], "%f", &last)
	fmt.Sscanf(parts[3], "%f", &open)
	fmt.Sscanf(parts[4], "%f", &high)
	fmt.Sscanf(parts[5], "%f", &low)
	fmt.Sscanf(parts[14], "%f", &vol)
	fmt.Sscanf(parts[13], "%f", &oi)

	fd := &FuturesData{
		Symbol:       symbol,
		Exchange:     string(nameBytes),
		LastPrice:    last,
		Open:         open,
		High:         high,
		Low:          low,
		Volume:       vol,
		OpenInterest: oi,
	}
	if last > 0 && preClose > 0 {
		fd.Change = last - preClose
		fd.ChangePct = fd.Change / preClose * 100
	}
	if fd.High > 0 && fd.High < fd.LastPrice {
		fd.High = fd.LastPrice
	}
	if fd.Low > 0 && fd.Low > fd.LastPrice {
		fd.Low = fd.LastPrice
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

// fetchDatacenter fetches rows from the EastMoney datacenter JSON API.
func fetchDatacenter(c *http.Client, reportName, columns, sortColumns string, pageSize int) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("sortColumns", sortColumns)
	params.Set("sortTypes", "-1")
	params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	params.Set("pageNumber", "1")
	params.Set("reportName", reportName)
	params.Set("columns", columns)
	params.Set("source", "WEB")
	params.Set("client", "WEB")

	fullURL := macroDataURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://data.eastmoney.com/")

	resp, err := c.Do(req)
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
		return nil, fmt.Errorf("parse datacenter response: %w", err)
	}
	if !mr.Success {
		return nil, fmt.Errorf("datacenter API failed for %s", reportName)
	}
	return mr.Result.Data, nil
}

func mapStr(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func mapFloat(m map[string]interface{}, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			var f float64
			if n, _ := fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f); n == 1 {
				return f
			}
		}
	}
	return 0
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
	data, err := fetchDatacenter(c.client, "RPT_BILLBOARD_LIST", "ALL", "TRADE_DATE", limit)
	if err != nil {
		return nil, fmt.Errorf("获取龙虎榜失败: %w", err)
	}
	results := make([]*DragonTigerData, 0, len(data))
	for _, d := range data {
		buy := mapFloat(d, "TOTAL_BUY_AMT")
		sell := mapFloat(d, "TOTAL_SELL_AMT")
		results = append(results, &DragonTigerData{
			StockCode: mapStr(d, "SECURITY_CODE", "SECUCODE"),
			StockName: mapStr(d, "SECURITY_NAME_ABBR", "SECURITY_NAME"),
			TradeDate: mapStr(d, "TRADE_DATE", "REPORT_DATE"),
			Reason:    mapStr(d, "BOARD_TYPE", "EXPLAIN", "REASON"),
			Turnover:  buy,
			NetBuy:    buy - sell,
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
	data, err := fetchDatacenter(c.client, "RPT_BOND_CB_LIST", "ALL", "PUBLIC_START_DATE", 200)
	if err != nil {
		return nil, fmt.Errorf("获取可转债失败: %w", err)
	}
	results := make([]*ConvertibleBond, 0, len(data))
	for _, d := range data {
		bondCode := mapStr(d, "SECURITY_CODE", "BOND_CODE", "SECUCODE")
		if bondCode == "" {
			continue
		}
		results = append(results, &ConvertibleBond{
			BondCode:     bondCode,
			BondName:     mapStr(d, "SECURITY_NAME_ABBR", "SECURITY_SHORT_NAME", "BOND_NAME"),
			StockCode:    mapStr(d, "CONVERT_STOCK_CODE", "STOCK_CODE"),
			StockName:    mapStr(d, "CORRECODE_NAME_ABBR", "STOCK_NAME"),
			IssuePrice:   mapFloat(d, "ISSUE_PRICE"),
			IssueAmount:  mapFloat(d, "ACTUAL_ISSUE_SCALE", "ISSUE_SCALE", "ISSUE_AMOUNT"),
			IssueDate:    mapStr(d, "VALUE_DATE", "PUBLIC_START_DATE", "ISSUE_DATE"),
			MaturityDate: mapStr(d, "EXPIRE_DATE", "CEASE_DATE", "MATURITY_DATE"),
			Rating:       mapStr(d, "RATING"),
			CouponRate0:  mapFloat(d, "COUPON_IR", "COUPON_RATE_0"),
			ConvertPrice: mapFloat(d, "INITIAL_TRANSFER_PRICE", "TRANSFER_PRICE", "CONVERT_PRICE"),
			ConvertStart: mapStr(d, "TRANSFER_START_DATE", "CONVERT_START_DATE"),
		})
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

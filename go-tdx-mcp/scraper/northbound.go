package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	northboundURL = "https://push2delay.eastmoney.com/api/qt/kamt.rtmin/get"
)

// NorthboundFlow represents northbound capital flow data.
type NorthboundFlow struct {
	Time           string  `json:"time"`
	NorthNetIn     float64 `json:"north_net_in"`     // 北向净流入(万元)
	SouthNetIn     float64 `json:"south_net_in"`     // 南向净流入(万元)
	NorthCumIn     float64 `json:"north_cum_in"`     // 北向累计流入(万元)
	SouthCumIn     float64 `json:"south_cum_in"`     // 南向累计流入(万元)
}

// NorthboundStock represents individual stock northbound holding data.
type NorthboundStock struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Market    string  `json:"market"`    // "SH"=沪股通, "SZ"=深股通
	ChangePct float64 `json:"change_pct"` // 涨跌幅(%)
	HoldRatio float64 `json:"hold_ratio"` // 持股占比(%)
	HoldShares float64 `json:"hold_shares"` // 持股数量(万股)
	HoldValue float64 `json:"hold_value"`  // 持股市值(万元)
}

// NorthboundHolder represents institutional holding data from northbound capital.
type NorthboundHolder struct {
	SecCode      string  `json:"sec_code"`       // 股票代码
	SecName      string  `json:"sec_name"`       // 股票名称
	Market       string  `json:"market"`         // "SH"=沪股通, "SZ"=深股通
	ReportDate   string  `json:"report_date"`    // 报告期
	TradeDate    string  `json:"trade_date"`     // 持仓日期
	ClosePrice   float64 `json:"close_price"`    // 收盘价
	HoldShares   float64 `json:"hold_shares"`    // 持股数量(股)
	HoldRatio    float64 `json:"hold_ratio"`     // 占自由流通股本比例(%)
	AddShares    float64 `json:"add_shares"`     // 增持/减持数量(股)
	AddRatio     float64 `json:"add_ratio"`      // 增持/减持比例(%)
	OrgQuantity  int     `json:"org_quantity"`   // 机构数量
	HoldCap      float64 `json:"hold_cap"`       // 持仓市值(元)
}

// NorthboundScraper scrapes northbound capital flow data from EastMoney.
type NorthboundScraper struct {
	client *http.Client
	limit  *RateLimiter
}

// NewNorthboundScraper creates a new NorthboundScraper.
func NewNorthboundScraper() *NorthboundScraper {
	return &NorthboundScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewNorthboundScraperWithAntiBan creates a NorthboundScraper with anti-bot protections.
func NewNorthboundScraperWithAntiBan() *NorthboundScraper {
	cfg := DefaultAntiBanConfig()
	cfg.MinDelay = 2000 * time.Millisecond
	cfg.MaxDelay = 4000 * time.Millisecond
	client := NewAntiBanClient(cfg)
	return &NorthboundScraper{
		client: client.Client(),
		limit:  NewRateLimiter(0.3, 3),
	}
}

// WithRateLimiter sets a rate limiter for this scraper.
func (n *NorthboundScraper) WithRateLimiter(lim *RateLimiter) {
	n.limit = lim
}

type northboundRawResponse struct {
	Data struct {
		S2N []string `json:"s2n"`
	} `json:"data"`
}

// GetFlowMinute fetches intraday northbound flow data (1-min intervals during trading).
func (n *NorthboundScraper) GetFlowMinute() ([]NorthboundFlow, error) {
	url := fmt.Sprintf("%s?fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54&ut=b2884a393a59ad64002292a3e90d46a5", northboundURL)

	if n.limit != nil {
		n.limit.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://data.eastmoney.com/hsgt/hsgtV2.html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nr northboundRawResponse
	if err := json.Unmarshal(body, &nr); err != nil {
		return nil, fmt.Errorf("parse northbound flow: %w", err)
	}

	flows := make([]NorthboundFlow, 0, len(nr.Data.S2N))
	for _, line := range nr.Data.S2N {
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}
		flow := NorthboundFlow{
			Time:   parts[0],
			NorthNetIn: parseStrFloat(parts[1]),
			SouthNetIn: parseStrFloat(parts[2]),
			NorthCumIn: parseStrFloat(parts[3]),
		}
		if len(parts) >= 5 {
			flow.SouthCumIn = parseStrFloat(parts[4])
		}
		flows = append(flows, flow)
	}

	return flows, nil
}

// GetDailyFlow fetches daily northbound flow history.
func (n *NorthboundScraper) GetDailyFlow(days int) ([]NorthboundFlow, error) {
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/kamt.kline/get?fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54&klt=101&lmt=%d&ut=b2884a393a59ad64002292a3e90d46a5", days)

	if n.limit != nil {
		n.limit.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://data.eastmoney.com/hsgt/hsgtV2.html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nr struct {
		Data struct {
			HK2SH []string `json:"hk2sh"`
			HK2SZ []string `json:"hk2sz"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &nr); err != nil {
		return nil, fmt.Errorf("parse daily flow: %w", err)
	}

	flows := make([]NorthboundFlow, 0, len(nr.Data.HK2SH))
	for _, line := range nr.Data.HK2SH {
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}
		flow := NorthboundFlow{
			Time:       parts[0],
			NorthNetIn: parseStrFloat(parts[1]),
			SouthNetIn: parseStrFloat(parts[2]),
			NorthCumIn: parseStrFloat(parts[3]),
		}
		if len(parts) >= 5 {
			flow.SouthCumIn = parseStrFloat(parts[4])
		}
		flows = append(flows, flow)
	}

	return flows, nil
}

// GetTopNorthboundStocks fetches top stocks by northbound holding (combined SH+SZ).
func (n *NorthboundScraper) GetTopNorthboundStocks(sortField string, count int) ([]NorthboundStock, error) {
	shStocks, _ := n.GetTopShanghaiNorthbound(sortField, count)
	szStocks, _ := n.GetTopShenzhenNorthbound(sortField, count)

	allStocks := append(shStocks, szStocks...)
	return allStocks, nil
}

// GetTopShanghaiNorthbound fetches top stocks held by Shanghai northbound (沪股通).
func (n *NorthboundScraper) GetTopShanghaiNorthbound(sortField string, count int) ([]NorthboundStock, error) {
	return n.getNorthboundStocks("001", sortField, count)
}

// GetTopShenzhenNorthbound fetches top stocks held by Shenzhen northbound (深股通).
func (n *NorthboundScraper) GetTopShenzhenNorthbound(sortField string, count int) ([]NorthboundStock, error) {
	return n.getNorthboundStocks("003", sortField, count)
}

// getNorthboundStocks fetches northbound holding stocks by market.
// Uses datacenter API RPT_MUTUAL_TOP10DEAL with MUTUAL_TYPE filter:
//   - "001" = 沪股通 (Shanghai-HK Connect)
//   - "003" = 深股通 (Shenzhen-HK Connect)
func (n *NorthboundScraper) getNorthboundStocks(mutualType string, sortField string, count int) ([]NorthboundStock, error) {
	if count <= 0 || count > 200 {
		count = 20
	}

	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_MUTUAL_TOP10DEAL&columns=ALL&filter=(MUTUAL_TYPE%%3D%%22%s%%22)&pageSize=%d&sortColumns=TRADE_DATE,RANK&sortTypes=-1,-1&source=WEB&client=WEB", mutualType, count)

	if n.limit != nil {
		n.limit.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://data.eastmoney.com/hsgt/hsgtV2.html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := n.client.Do(req)
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
				SecurityCode    string  `json:"SECURITY_CODE"`
				SecurityName    string  `json:"SECURITY_NAME"`
				TradeDate       string  `json:"TRADE_DATE"`
				ClosePrice      float64 `json:"CLOSE_PRICE"`
				ChangeRate      float64 `json:"CHANGE_RATE"`
				DealAmt         float64 `json:"DEAL_AMT"`
				MutualRatio     float64 `json:"MUTUAL_RATIO"`
				TurnoverRate    float64 `json:"TURNOVERRATE"`
				Change          float64 `json:"CHANGE"`
				Market          string  `json:"DERIVE_SECURITY_CODE"`
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse northbound stocks: %w", err)
	}

	stocks := make([]NorthboundStock, 0, len(result.Result.Data))
	for _, item := range result.Result.Data {
		market := "SH"
		if mutualType == "003" {
			market = "SZ"
		}
		stocks = append(stocks, NorthboundStock{
			Code:      item.SecurityCode,
			Name:      item.SecurityName,
			Market:    market,
			ChangePct: item.ChangeRate,
			HoldValue: item.DealAmt / 10000, // convert to 万元
		})
	}

	return stocks, nil
}

func parseStrFloat(s string) float64 {
	var v float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &v)
	return v
}

// GetNorthboundHolders fetches institutional holding data from northbound capital.
// Uses datacenter API RPT_MUTUAL_HOLDRANK_NEW with optional MUTUAL_TYPE filter:
//   - "001" = 沪股通 (Shanghai-HK Connect)
//   - "003" = 深股通 (Shenzhen-HK Connect)
//   - "" = all markets
func (n *NorthboundScraper) GetNorthboundHolders(mutualType string, pageSize int) ([]*NorthboundHolder, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	filter := ""
	if mutualType != "" {
		filter = fmt.Sprintf("(MUTUAL_TYPE%%3D%%22%s%%22)", mutualType)
	}

	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_MUTUAL_HOLDRANK_NEW&columns=ALL&pageSize=%d&pageNumber=1&source=WEB&client=WEB", pageSize)
	if filter != "" {
		url += "&filter=" + filter
	}

	if n.limit != nil {
		n.limit.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://data.eastmoney.com/hsgt/hsgtV2.html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := n.client.Do(req)
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
				SecurityCode     string  `json:"SECURITY_CODE"`
				SecurityName     string  `json:"SECURITY_NAME_ABBR"`
				Market           string  `json:"MUTUAL_TYPE"`
				TradeDate        string  `json:"HOLD_DATE"`
				ReportDate       string  `json:"REPORTDATE"`
				ClosePrice       float64 `json:"CLOSE_PRICE"`
				HoldShares       float64 `json:"HOLD_SHARES"`
				HoldSharesRatio  float64 `json:"HOLD_SHARES_RATIO"`
				AddShares        float64 `json:"ADD_SHARES_REPAIR"`
				AddSharesAmp     float64 `json:"ADD_SHARES_AMP"`
				OrgQuantity      int     `json:"ORG_QUANTITY"`
				HoldMarketCap    float64 `json:"HOLD_MARKET_CAP"`
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse northbound holders: %w", err)
	}

	holders := make([]*NorthboundHolder, 0, len(result.Result.Data))
	for _, item := range result.Result.Data {
		market := "SH"
		if item.Market == "003" {
			market = "SZ"
		}

		// Parse trade date (remove time part)
		tradeDate := strings.TrimSpace(item.TradeDate)
		if idx := strings.Index(tradeDate, " "); idx >= 0 {
			tradeDate = tradeDate[:idx]
		}

		reportDate := strings.TrimSpace(item.ReportDate)
		if idx := strings.Index(reportDate, " "); idx >= 0 {
			reportDate = reportDate[:idx]
		}

		holders = append(holders, &NorthboundHolder{
			SecCode:     item.SecurityCode,
			SecName:     item.SecurityName,
			Market:      market,
			TradeDate:   tradeDate,
			ReportDate:  reportDate,
			ClosePrice:  item.ClosePrice,
			HoldShares:  item.HoldShares,
			HoldRatio:   item.HoldSharesRatio,
			AddShares:   item.AddShares,
			AddRatio:    item.AddSharesAmp,
			OrgQuantity: item.OrgQuantity,
			HoldCap:     item.HoldMarketCap,
		})
	}

	return holders, nil
}

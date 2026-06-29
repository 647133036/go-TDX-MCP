package scraper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Sina margin trade data.
type SinaMarginData struct {
	Tradedate  string  `json:"trade_date"`
	Rzye       float64 `json:"rzye"`
	Rzre       float64 `json:"rzre"`
	Rqye       float64 `json:"rqye"`
	Rqrl       float64 `json:"rqrl"`
	Rzmre      float64 `json:"rzmre"`
}

// Sina dragon tiger data.
type SinaDragonTigerData struct {
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	TradeDate  string  `json:"trade_date"`
	Reason     string  `json:"reason"`
	BuyAmount  float64 `json:"buy_amount"`
	SellAmount float64 `json:"sell_amount"`
	NetAmount  float64 `json:"net_amount"`
	Turnover   float64 `json:"turnover"`
}

// Sina convertible bond data.
type SinaConvertibleBond struct {
	BondCode     string  `json:"bond_code"`
	BondName     string  `json:"bond_name"`
	Price        float64 `json:"price"`
	ChangePct    float64 `json:"change_pct"`
	StockCode    string  `json:"stock_code"`
	StockName    string  `json:"stock_name"`
	ConvertPrice float64 `json:"convert_price"`
	ConvertStart string  `json:"convert_start"`
	Rating       string  `json:"rating"`
	RemainSize   float64 `json:"remain_size"`
	Turnover     float64 `json:"turnover"`
}

// Sina block trade data.
type SinaBlockTradeData struct {
	StockCode   string  `json:"stock_code"`
	StockName   string  `json:"stock_name"`
	TradeDate   string  `json:"trade_date"`
	Price       float64 `json:"price"`
	ChangePct   float64 `json:"change_pct"`
	Amount      float64 `json:"amount"`
	Buyer       string  `json:"buyer"`
	Seller      string  `json:"seller"`
	TurnoverPct float64 `json:"turnover_pct"`
}

// SinaClient provides data from Sina Finance APIs.
type SinaClient struct {
	client *http.Client
}

func NewSinaClient() *SinaClient {
	return &SinaClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetMarginTrade fetches margin trade summary from Sina.
func (c *SinaClient) GetMarginTrade(limit int) ([]*SinaMarginData, error) {
	url := "https://stock.finance.sina.com.cn/hkstock/api/jsonv2.py/CN_HKMarginTradeAll/get?page=1&num=30"
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
	if len(data) < 10 {
		return []*SinaMarginData{}, nil
	}

	var results []*SinaMarginData
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "{") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		m := &SinaMarginData{
			Tradedate: parts[0],
		}
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%f", &m.Rzye)
		}
		if len(parts) > 2 {
			fmt.Sscanf(parts[2], "%f", &m.Rzre)
		}
		if len(parts) > 3 {
			fmt.Sscanf(parts[3], "%f", &m.Rqye)
		}
		if len(parts) > 4 {
			fmt.Sscanf(parts[4], "%f", &m.Rqrl)
		}
		if len(parts) > 5 {
			fmt.Sscanf(parts[5], "%f", &m.Rzmre)
		}
		results = append(results, m)
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	return results, nil
}

// GetDragonTiger fetches dragon tiger list from Sina.
func (c *SinaClient) GetDragonTiger(limit int) ([]*SinaDragonTigerData, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	url := fmt.Sprintf("https://stock.finance.sina.com.cn/hkstock/api/jsonv2.py/CN_HKTopInst/get?page=1&num=%d", limit)
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
	if len(data) < 10 {
		return []*SinaDragonTigerData{}, nil
	}

	var results []*SinaDragonTigerData
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 8 {
			continue
		}

		dt := &SinaDragonTigerData{
			StockCode: parts[0],
			StockName: parts[1],
			TradeDate: parts[2],
			Reason:    parts[3],
		}
		if len(parts) > 4 {
			fmt.Sscanf(parts[4], "%f", &dt.BuyAmount)
		}
		if len(parts) > 5 {
			fmt.Sscanf(parts[5], "%f", &dt.SellAmount)
		}
		if len(parts) > 6 {
			fmt.Sscanf(parts[6], "%f", &dt.NetAmount)
		}
		if len(parts) > 7 {
			fmt.Sscanf(parts[7], "%f", &dt.Turnover)
		}
		dt.NetAmount = dt.BuyAmount - dt.SellAmount
		results = append(results, dt)
	}
	return results, nil
}

// GetConvertibleBonds fetches convertible bond list from Sina.
func (c *SinaClient) GetConvertibleBonds() ([]*SinaConvertibleBond, error) {
	url := "https://stock.finance.sina.com.cn/hkstock/api/jsonv2.py/CN_CBAll/get?page=1&num=200"
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
	if len(data) < 10 {
		return []*SinaConvertibleBond{}, nil
	}

	var results []*SinaConvertibleBond
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 10 {
			continue
		}

		cb := &SinaConvertibleBond{
			BondCode:  parts[0],
			BondName:  parts[1],
			StockCode: parts[2],
			StockName: parts[3],
			Rating:    parts[4],
		}
		if len(parts) > 5 {
			fmt.Sscanf(parts[5], "%f", &cb.Price)
		}
		if len(parts) > 6 {
			fmt.Sscanf(parts[6], "%f", &cb.ChangePct)
		}
		if len(parts) > 7 {
			fmt.Sscanf(parts[7], "%f", &cb.ConvertPrice)
		}
		if len(parts) > 8 {
			cb.ConvertStart = parts[8]
		}
		if len(parts) > 9 {
			fmt.Sscanf(parts[9], "%f", &cb.RemainSize)
		}
		results = append(results, cb)
	}
	return results, nil
}

// GetBlockTrades fetches block trade data from Sina.
func (c *SinaClient) GetBlockTrades(limit int) ([]*SinaBlockTradeData, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Sina block trade API
	url := fmt.Sprintf("https://stock.finance.sina.com.cn/hkstock/api/jsonv2.py/CN_BlockTrade/get?page=1&num=%d", limit)
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
	if len(data) < 10 {
		return []*SinaBlockTradeData{}, nil
	}

	var results []*SinaBlockTradeData
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 8 {
			continue
		}

		bt := &SinaBlockTradeData{
			StockCode: parts[0],
			StockName: parts[1],
			TradeDate: parts[2],
			Price:     parseFloat(parts[3]),
			ChangePct: parseFloat(parts[4]),
			Amount:    parseFloat(parts[5]),
			Buyer:     parts[6],
			Seller:    parts[7],
		}
		if len(parts) > 8 {
			bt.TurnoverPct = parseFloat(parts[8])
		}
		results = append(results, bt)
	}
	return results, nil
}

// GetStockQuotes fetches real-time quotes from Sina Finance API.
func (c *SinaClient) GetStockQuotes(securities []string) (map[string]map[string]interface{}, error) {
	if len(securities) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	// Build secid list for Sina API
	secids := make([]string, 0, len(securities))
	for _, sec := range securities {
		sec = strings.TrimSpace(sec)
		lowerSec := strings.ToLower(sec)
		if strings.HasPrefix(lowerSec, "sh") || strings.HasPrefix(lowerSec, "sz") {
			secids = append(secids, lowerSec)
		} else if strings.HasPrefix(sec, "6") || strings.HasPrefix(sec, "9") {
			secids = append(secids, "sh"+sec)
		} else {
			secids = append(secids, "sz"+sec)
		}
	}

	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", strings.Join(secids, ","))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "var hq_str_", "")
	content = strings.ReplaceAll(content, "\";\"", "")
	content = strings.TrimSpace(content)

	results := make(map[string]map[string]interface{})
	lines := strings.Split(content, "\n")

	quotePattern := regexp.MustCompile(`hq_str\["([^"]+)"\]`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := quotePattern.FindStringSubmatch(line)
		var symbol string
		if len(matches) > 1 {
			symbol = matches[1]
		} else {
			parts := strings.Split(line, "=")
			if len(parts) > 0 {
				symbol = parts[0]
			}
		}

		dataParts := strings.Split(line, ",")
		if len(dataParts) < 5 {
			continue
		}

		quote := map[string]interface{}{
			"name":              dataParts[0],
			"open":              parseFloat(dataParts[1]),
			"yesterday_close":   parseFloat(dataParts[2]),
			"price":             parseFloat(dataParts[3]),
			"high":              parseFloat(dataParts[4]),
			"low":               parseFloat(dataParts[5]),
			"volume":            parseFloat(dataParts[8]),
			"amount":            parseFloat(dataParts[9]),
		}

		if len(dataParts) > 10 {
			quote["buy_price"] = parseFloat(dataParts[6])
			quote["sell_price"] = parseFloat(dataParts[7])
		}

		if symbol != "" {
			results[symbol] = quote
		}
	}

	return results, nil
}

// GetHKStockQuotes fetches HK stock quotes from Sina.
func (c *SinaClient) GetHKStockQuotes(securities []string) (map[string]map[string]interface{}, error) {
	if len(securities) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	secids := make([]string, 0, len(securities))
	for _, sec := range securities {
		sec = strings.TrimSpace(sec)
		if len(sec) == 5 {
			secids = append(secids, fmt.Sprintf("rt_hk%s", sec))
		}
	}

	if len(secids) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", strings.Join(secids, ","))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "var hq_str_", "")
	content = strings.ReplaceAll(content, "\";\"", "")
	content = strings.TrimSpace(content)

	results := make(map[string]map[string]interface{})
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		dataParts := strings.Split(line, ",")
		if len(dataParts) < 32 {
			continue
		}

		quote := map[string]interface{}{
			"name":          dataParts[0],
			"open":          parseFloat(dataParts[2]),
			"yesterday_close": parseFloat(dataParts[3]),
			"price":         parseFloat(dataParts[33]),
			"high":          parseFloat(dataParts[37]),
			"low":           parseFloat(dataParts[38]),
			"volume":        parseFloat(dataParts[34]),
			"amount":        parseFloat(dataParts[35]),
			"market_cap":    parseFloat(dataParts[46]),
		}

		results[dataParts[0]] = quote
	}

	return results, nil
}

// GetUSStockQuotes fetches US stock quotes from Sina.
func (c *SinaClient) GetUSStockQuotes(securities []string) (map[string]map[string]interface{}, error) {
	if len(securities) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	secids := make([]string, 0, len(securities))
	for _, sec := range securities {
		sec = strings.ToUpper(strings.TrimSpace(sec))
		if len(sec) > 0 && len(sec) <= 6 {
			secids = append(secids, fmt.Sprintf("gb_%s", strings.ToLower(sec)))
		}
	}

	if len(secids) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", strings.Join(secids, ","))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn/")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "var hq_str_", "")
	content = strings.ReplaceAll(content, "\";\"", "")
	content = strings.TrimSpace(content)

	results := make(map[string]map[string]interface{})
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		dataParts := strings.Split(line, ",")
		if len(dataParts) < 20 {
			continue
		}

		quote := map[string]interface{}{
			"name":            dataParts[0],
			"price":           parseFloat(dataParts[1]),
			"change":          parseFloat(dataParts[2]),
			"change_pct":      parseFloat(dataParts[3]),
			"open":            parseFloat(dataParts[4]),
			"yesterday_close": parseFloat(dataParts[5]),
			"high":            parseFloat(dataParts[6]),
			"low":             parseFloat(dataParts[7]),
			"volume":          parseFloat(dataParts[8]),
			"amount":          parseFloat(dataParts[9]),
		}

		results[dataParts[0]] = quote
	}

	return results, nil
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "-" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

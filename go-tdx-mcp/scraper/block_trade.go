package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BlockTradeData represents a block trade record.
type BlockTradeData struct {
	StockCode   string  `json:"stock_code"`
	StockName   string  `json:"stock_name"`
	TradeDate   string  `json:"trade_date"`
	Price       float64 `json:"price"`
	ChangePct   float64 `json:"change_pct"`
	Amount      float64 `json:"amount"`
	TurnoverPct float64 `json:"turnover_pct"`
	BuyerName   string  `json:"buyer_name"`
	SellerName  string  `json:"seller_name"`
}

// BlockTradeClient fetches block trade data from EastMoney.
type BlockTradeClient struct {
	client *http.Client
}

func NewBlockTradeClient() *BlockTradeClient {
	return &BlockTradeClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetBlockTrades fetches recent block trade records.
func (c *BlockTradeClient) GetBlockTrades(limit int) ([]*BlockTradeData, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=TRADE_DATE&sortTypes=-1&pageSize=%d&pageNumber=1&reportName=RPT_BLOCKTRADE_HISTROY&columns=ALL&source=WEB", limit)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type apiResponse struct {
		Result struct {
			Records []map[string]interface{} `json:"records"`
			Total   int                      `json:"total"`
		} `json:"result"`
	}
	
	var respData apiResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}
	
	results := make([]*BlockTradeData, 0, len(respData.Result.Records))
	for _, record := range respData.Result.Records {
		bt := &BlockTradeData{}
		
		if code, ok := record["SECURITY_CODE"]; ok {
			bt.StockCode = fmt.Sprintf("%v", code)
		}
		if name, ok := record["SECURITY_NAME_ABBR"]; ok {
			bt.StockName = fmt.Sprintf("%v", name)
		}
		if date, ok := record["TRADE_DATE"]; ok {
			bt.TradeDate = fmt.Sprintf("%v", date)
		}
		if price, ok := record["TRADE_PRICE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", price), "%f", &bt.Price)
		}
		if changePct, ok := record["CHANGE_RATE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", changePct), "%f", &bt.ChangePct)
		}
		if amount, ok := record["TRADE_VOLUME"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", amount), "%f", &bt.Amount)
		}
		if turnover, ok := record["TURNOVER_RATE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", turnover), "%f", &bt.TurnoverPct)
		}
		if buyer, ok := record["BUYER_ABBR"]; ok {
			bt.BuyerName = fmt.Sprintf("%v", buyer)
		}
		if seller, ok := record["SELLER_ABBR"]; ok {
			bt.SellerName = fmt.Sprintf("%v", seller)
		}
		
		results = append(results, bt)
	}
	
	return results, nil
}

// GetBlockTradesByDate fetches block trades for a specific date.
func (c *BlockTradeClient) GetBlockTradesByDate(date string, limit int) ([]*BlockTradeData, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	
	filter := fmt.Sprintf("[{\"name\":\"TRADE_DATE\",\"value\":\"%s\"}]", date)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=TRADE_DATE&sortTypes=-1&pageSize=%d&pageNumber=1&reportName=RPT_BLOCKTRADE_HISTROY&columns=ALL&filter=%s&source=WEB", limit, filter)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type apiResponse struct {
		Result struct {
			Records []map[string]interface{} `json:"records"`
		} `json:"result"`
	}
	
	var respData apiResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}
	
	results := make([]*BlockTradeData, 0, len(respData.Result.Records))
	for _, record := range respData.Result.Records {
		bt := &BlockTradeData{
			TradeDate: date,
		}
		
		if code, ok := record["SECURITY_CODE"]; ok {
			bt.StockCode = fmt.Sprintf("%v", code)
		}
		if name, ok := record["SECURITY_NAME_ABBR"]; ok {
			bt.StockName = fmt.Sprintf("%v", name)
		}
		if price, ok := record["TRADE_PRICE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", price), "%f", &bt.Price)
		}
		if changePct, ok := record["CHANGE_RATE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", changePct), "%f", &bt.ChangePct)
		}
		if amount, ok := record["TRADE_VOLUME"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", amount), "%f", &bt.Amount)
		}
		
		results = append(results, bt)
	}
	
	return results, nil
}

// GetBlockTradesByStock fetches block trades for a specific stock.
func (c *BlockTradeClient) GetBlockTradesByStock(stockCode string, limit int) ([]*BlockTradeData, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	
	filter := fmt.Sprintf("[{\"name\":\"SECURITY_CODE\",\"value\":\"%s\"}]", stockCode)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=TRADE_DATE&sortTypes=-1&pageSize=%d&pageNumber=1&reportName=RPT_BLOCKTRADE_HISTROY&columns=ALL&filter=%s&source=WEB", limit, filter)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type apiResponse struct {
		Result struct {
			Records []map[string]interface{} `json:"records"`
		} `json:"result"`
	}
	
	var respData apiResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}
	
	results := make([]*BlockTradeData, 0, len(respData.Result.Records))
	for _, record := range respData.Result.Records {
		bt := &BlockTradeData{
			StockCode: stockCode,
		}
		
		if name, ok := record["SECURITY_NAME_ABBR"]; ok {
			bt.StockName = fmt.Sprintf("%v", name)
		}
		if date, ok := record["TRADE_DATE"]; ok {
			bt.TradeDate = fmt.Sprintf("%v", date)
		}
		if price, ok := record["TRADE_PRICE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", price), "%f", &bt.Price)
		}
		if changePct, ok := record["CHANGE_RATE"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", changePct), "%f", &bt.ChangePct)
		}
		if amount, ok := record["TRADE_VOLUME"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", amount), "%f", &bt.Amount)
		}
		
		results = append(results, bt)
	}
	
	return results, nil
}

// GetBlockTradeStatistics fetches daily block trade statistics.
func (c *BlockTradeClient) GetBlockTradeStatistics(startDate, endDate string) ([]map[string]interface{}, error) {
	filter := fmt.Sprintf("[{\"name\":\"TRADE_DATE\",\"value\":\"[%s,%s]\"}]", startDate, endDate)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=TRADE_DATE&sortTypes=-1&pageSize=100&pageNumber=1&reportName=RPT_BLOCKTRADE_STATS&columns=ALL&filter=%s&source=WEB", filter)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type apiResponse struct {
		Result struct {
			Records []map[string]interface{} `json:"records"`
		} `json:"result"`
	}
	
	var respData apiResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}
	
	return respData.Result.Records, nil
}

// SearchBlockTrades searches block trades by keyword (stock name or code).
func (c *BlockTradeClient) SearchBlockTrades(keyword string, limit int) ([]*BlockTradeData, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	
	// First search the stock
	resolver := NewStockCodeResolver()
	info, err := resolver.Resolve(keyword)
	if err != nil || info == nil {
		return c.GetBlockTrades(limit)
	}
	
	return c.GetBlockTradesByStock(info.Code, limit)
}

// GetRecentActiveStocks gets stocks with most recent block trades.
func (c *BlockTradeClient) GetRecentActiveStocks(days int) ([]map[string]interface{}, error) {
	if days <= 0 || days > 30 {
		days = 7
	}
	
	now := time.Now()
	startDate := now.AddDate(0, 0, -days).Format("2006-01-02")
	endDate := now.Format("2006-01-02")
	
	filter := fmt.Sprintf("[{\"name\":\"TRADE_DATE\",\"value\":\"[%s,%s]\"}]", startDate, endDate)
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=TRADE_VOLUME&sortTypes=-1&pageSize=100&pageNumber=1&reportName=RPT_BLOCKTRADE_HISTROY&columns=ALL&filter=%s&source=WEB", filter)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type apiResponse struct {
		Result struct {
			Records []map[string]interface{} `json:"records"`
		} `json:"result"`
	}
	
	var respData apiResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}
	
	stockMap := make(map[string]map[string]interface{})
	for _, record := range respData.Result.Records {
		code := fmt.Sprintf("%v", record["SECURITY_CODE"])
		name := fmt.Sprintf("%v", record["SECURITY_NAME_ABBR"])
		
		if existing, ok := stockMap[code]; ok {
			existing["trade_count"] = existing["trade_count"].(int) + 1
			if amount, ok := record["TRADE_VOLUME"]; ok {
				existing["total_amount"] = existing["total_amount"].(float64) + parseFloat(fmt.Sprintf("%v", amount))
			}
		} else {
			amount := float64(0)
			if v, ok := record["TRADE_VOLUME"]; ok {
				amount = parseFloat(fmt.Sprintf("%v", v))
			}
			stockMap[code] = map[string]interface{}{
				"code":        code,
				"name":        name,
				"trade_count": 1,
				"total_amount": amount,
			}
		}
	}
	
	results := make([]map[string]interface{}, 0, len(stockMap))
	for _, v := range stockMap {
		results = append(results, v)
	}
	
	return results, nil
}

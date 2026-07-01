package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// EastMoneyScraper uses push2delay/push2ex/push2his APIs discovered from levistock.
type EastMoneyScraper struct {
	client *http.Client
	mu     sync.Mutex
}

func NewEastMoneyScraper() *EastMoneyScraper {
	return &EastMoneyScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *EastMoneyScraper) doJSON(url string, extraHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	rc, ok := result["rc"]
	if ok {
		if rcF, ok := rc.(float64); ok && rcF != 0 {
			return nil, fmt.Errorf("api rc=%v, msg=%v", rc, result["dsc"])
		}
	}
	return body, nil
}

// SecidForCode generates EastMoney secid from stock code.
// 6xx -> 1.xxx (Shanghai), 0xx/3xx -> 0.xxx (Shenzhen)
func SecidForCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.TrimPrefix(code, "SH")
	code = strings.TrimPrefix(code, "SZ")
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return ""
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return ""
		}
	}
	first := code[0]
	if first == '6' {
		return fmt.Sprintf("1.%s", code)
	}
	return fmt.Sprintf("0.%s", code)
}

// RealtimeQuote fetches real-time quotes for multiple stocks via push2delay ulist.
func (e *EastMoneyScraper) RealtimeQuote(codes []string) ([]map[string]interface{}, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	secids := make([]string, 0, len(codes))
	for _, code := range codes {
		s := SecidForCode(code)
		if s != "" {
			secids = append(secids, s)
		}
	}
	if len(secids) == 0 {
		return nil, fmt.Errorf("no valid codes")
	}

	url := fmt.Sprintf(
		"http://push2delay.eastmoney.com/api/qt/ulist.np/get"+
			"?fields=f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f14,f15,f16,f17,f18,f20,f21,f23"+
			"&fltt=2&invt=2&ut=fa5fd1943c7b386f172d6893dbfba10b"+
			"&wbp2u=|0|0|0|web&secids=%s",
		strings.Join(secids, ","),
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type rawItem struct {
		F12 string  `json:"f12"`
		F14 string  `json:"f14"`
		F2  float64 `json:"f2"`
		F3  float64 `json:"f3"`
		F4  float64 `json:"f4"`
		F5  float64 `json:"f5"`
		F6  float64 `json:"f6"`
		F7  float64 `json:"f7"`
		F8  float64 `json:"f8"`
		F9  float64 `json:"f9"`
		F10 float64 `json:"f10"`
		F15 float64 `json:"f15"`
		F16 float64 `json:"f16"`
		F17 float64 `json:"f17"`
		F18 float64 `json:"f18"`
		F20 float64 `json:"f20"`
		F21 float64 `json:"f21"`
		F23 float64 `json:"f23"`
	}

	var parsed struct {
		Data struct {
			Diff []rawItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Diff))
	for _, item := range parsed.Data.Diff {
		results = append(results, map[string]interface{}{
			"stock_code":    item.F12,
			"stock_name":    item.F14,
			"price":         item.F2,
			"change_pct":    item.F3,
			"change_amt":    item.F4,
			"volume":        item.F5,
			"amount":        item.F6,
			"amplitude":     item.F7,
			"turnover_rate": item.F8,
			"pe_ttm":        item.F9,
			"volume_ratio":  item.F10,
			"high":          item.F15,
			"low":           item.F16,
			"open":          item.F17,
			"pre_close":     item.F18,
			"total_market":  item.F20,
			"circ_market":   item.F21,
			"pb":            item.F23,
		})
	}
	return results, nil
}

// SectorBoards fetches all sector boards (industry/concept/region).
// Uses push2delay with fs filter: m:90+t:2 (industry), m:90+t:3 (concept), m:90+t:1 (region).
func (e *EastMoneyScraper) SectorBoards(boardType string) ([]map[string]interface{}, error) {
	fsMap := map[string]string{
		"industry": "m:90+t:2+f:!50",
		"concept":  "m:90+t:3+f:!50",
		"region":   "m:90+t:1+f:!50",
	}
	fs, ok := fsMap[boardType]
	if !ok {
		fs = "m:90+t:3+f:!50"
	}

	var allResults []map[string]interface{}
	for pn := 1; ; pn++ {
		url := fmt.Sprintf(
			"http://push2delay.eastmoney.com/api/qt/clist/get"+
				"?pn=%d&pz=200&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281"+
				"&fltt=2&invt=2&fid=f3&fs=%s"+
				"&fields=f12,f14,f2,f3,f4,f5,f6,f7,f8,f20,f62,f128,f136,f140,f104,f105,f207,f208",
			pn, fs,
		)

		body, err := e.doJSON(url, nil)
		if err != nil {
			return nil, err
		}

		type rawItem struct {
			F12  string          `json:"f12"`
			F14  string          `json:"f14"`
			F2   json.RawMessage `json:"f2"`
			F3   json.RawMessage `json:"f3"`
			F4   json.RawMessage `json:"f4"`
			F5   json.RawMessage `json:"f5"`
			F6   json.RawMessage `json:"f6"`
			F7   json.RawMessage `json:"f7"`
			F8   json.RawMessage `json:"f8"`
			F20  json.RawMessage `json:"f20"`
			F62  json.RawMessage `json:"f62"`
			F128 string          `json:"f128"`
			F136 json.RawMessage `json:"f136"`
			F140 string          `json:"f140"`
			F104 json.RawMessage `json:"f104"`
			F105 json.RawMessage `json:"f105"`
			F207 string          `json:"f207"`
			F208 string          `json:"f208"`
		}

		var parsed struct {
			Data struct {
				Total int       `json:"total"`
				Diff  []rawItem `json:"diff"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, item := range parsed.Data.Diff {
			result := map[string]interface{}{
				"sector_code":   item.F12,
				"sector_name":   item.F14,
				"lead_stock_name": item.F128,
				"lead_stock_code": item.F140,
				"top_drop_name": item.F207,
				"top_drop_code": item.F208,
			}
			result["price"] = parseRawNum(item.F2)
			result["change_pct"] = parseRawNum(item.F3)
			result["change_amt"] = parseRawNum(item.F4)
			result["volume"] = parseRawNum(item.F5)
			result["amount"] = parseRawNum(item.F6)
			result["amplitude"] = parseRawNum(item.F7)
			result["turnover_rate"] = parseRawNum(item.F8)
			result["total_market"] = parseRawNum(item.F20)
			result["main_inflow"] = parseRawNum(item.F62)
			result["lead_stock_chg"] = parseRawNum(item.F136)
			result["up_count"] = parseRawNum(item.F104)
			result["down_count"] = parseRawNum(item.F105)
			allResults = append(allResults, result)
		}

		if len(allResults) >= parsed.Data.Total || parsed.Data.Total == 0 {
			break
		}
	}
	return allResults, nil
}

// SectorStocks fetches constituent stocks for a board code like BK1033.
// Uses fs=b:BK1033+f:!50 (NO SPACE after colon).
func (e *EastMoneyScraper) SectorStocks(boardCode string) ([]map[string]interface{}, error) {
	var allResults []map[string]interface{}
	for pn := 1; ; pn++ {
		url := fmt.Sprintf(
			"http://push2delay.eastmoney.com/api/qt/clist/get"+
				"?pn=%d&pz=500&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281"+
				"&fltt=2&invt=2&fid=f3&fs=b:%s+f:!50"+
				"&fields=f12,f14",
			pn, boardCode,
		)

		body, err := e.doJSON(url, nil)
		if err != nil {
			return nil, err
		}

		type rawItem struct {
			F12 string `json:"f12"`
			F14 string `json:"f14"`
		}

		var parsed struct {
			Data struct {
				Total int       `json:"total"`
				Diff  []rawItem `json:"diff"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}

		for _, item := range parsed.Data.Diff {
			allResults = append(allResults, map[string]interface{}{
				"stock_code": item.F12,
				"stock_name": item.F14,
			})
		}

		if len(allResults) >= parsed.Data.Total || parsed.Data.Total == 0 {
			break
		}
	}
	return allResults, nil
}

// StockBelongSector queries which sectors a stock belongs to.
// Uses push2delay ulist with f100 field.
func (e *EastMoneyScraper) StockBelongSector(codes []string) ([]map[string]interface{}, error) {
	if len(codes) > 100 {
		codes = codes[:100]
	}
	secids := make([]string, 0, len(codes))
	for _, code := range codes {
		s := SecidForCode(code)
		if s != "" {
			secids = append(secids, s)
		}
	}
	if len(secids) == 0 {
		return nil, fmt.Errorf("no valid codes")
	}

	url := fmt.Sprintf(
		"http://push2delay.eastmoney.com/api/qt/ulist.np/get"+
			"?fields=f12,f14,f100&invt=2&ut=fa5fd1943c7b386f172d6893dbfba10b"+
			"&wbp2u=|0|0|0|web&secids=%s",
		strings.Join(secids, ","),
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type rawItem struct {
		F12   string `json:"f12"`
		F14   string `json:"f14"`
		F100  string `json:"f100"`
	}

	var parsed struct {
		Data struct {
			Diff []rawItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Diff))
	for _, item := range parsed.Data.Diff {
		results = append(results, map[string]interface{}{
			"stock_code":  item.F12,
			"stock_name":  item.F14,
			"sector_name": item.F100,
		})
	}
	return results, nil
}

// MarketIndices fetches major market indices (上证/深证/创业板/科创50/沪深300/中证500).
// Uses fs=b:MK0010 filter.
func (e *EastMoneyScraper) MarketIndices() ([]map[string]interface{}, error) {
	url := "https://push2delay.eastmoney.com/api/qt/clist/get" +
		"?np=1&fltt=1&invt=2&fs=b:MK0010" +
		"&fields=f12,f14,f2,f3,f4,f5,f6,f15,f16,f17,f18" +
		"&fid=&pn=1&pz=100&po=1&ut=fa5fd1943c7b386f172d6893dbfba10b&dect=1"

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type rawItem struct {
		F12     string  `json:"f12"`
		F14     string  `json:"f14"`
		F2      float64 `json:"f2"`
		F3      float64 `json:"f3"`
		F4      float64 `json:"f4"`
		F5      float64 `json:"f5"`
		F6      float64 `json:"f6"`
		F15     float64 `json:"f15"`
		F16     float64 `json:"f16"`
		F17     float64 `json:"f17"`
		F18     float64 `json:"f18"`
	}

	var parsed struct {
		Data struct {
			Diff []rawItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Diff))
	for _, item := range parsed.Data.Diff {
		results = append(results, map[string]interface{}{
			"name":       item.F14,
			"code":       item.F12,
			"price":      item.F2 / 100,
			"change_pct": item.F3 / 100,
			"change_amt": item.F4 / 100,
			"volume":     item.F5,
			"amount":     item.F6,
			"high":       item.F15 / 100,
			"low":        item.F16 / 100,
			"open":       item.F17 / 100,
			"pre_close":  item.F18 / 100,
		})
	}
	return results, nil
}

// LimitUpPool fetches today's limit-up stock pool.
// push2ex.getTopicZTPool, price p field needs /1000.
func (e *EastMoneyScraper) LimitUpPool(date string) ([]map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://push2ex.eastmoney.com/getTopicZTPool"+
			"?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wz.ztzt"+
			"&Pageindex=0&pagesize=3000&sort=fbt:asc&date=%s",
		date,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type ztItem struct {
		C          json.RawMessage `json:"c"`
		M          int             `json:"m"`
		N          string          `json:"n"`
		P          json.RawMessage `json:"p"`
		Zdp        float64         `json:"zdp"`
		Amount     float64         `json:"amount"`
		Ltsz       float64         `json:"ltsz"`
		Tshare     float64         `json:"tshare"`
		Hs         float64         `json:"hs"`
		Lbc        int             `json:"lbc"`
		Fbt        json.RawMessage `json:"fbt"`
		Lbt        json.RawMessage `json:"lbt"`
		Fund       float64         `json:"fund"`
		Zbc        int             `json:"zbc"`
		Hybk       string          `json:"hybk"`
		Zttj       map[string]interface{} `json:"zttj"`
	}

	var parsed struct {
		Data struct {
			TC   int             `json:"tc"`
			Qdate json.RawMessage `json:"qdate"`
			Pool []ztItem        `json:"pool"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	qdateStr := ""
	if len(parsed.Data.Qdate) > 0 {
		// Try as string first
		if parsed.Data.Qdate[0] == '"' {
			var s string
			if err := json.Unmarshal(parsed.Data.Qdate, &s); err == nil {
				qdateStr = s
			}
		} else {
			// Try as number
			var n float64
			if err := json.Unmarshal(parsed.Data.Qdate, &n); err == nil {
				qdateStr = fmt.Sprintf("%.0f", n)
			}
		}
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Pool))
	for _, item := range parsed.Data.Pool {
		// Parse C field - could be string or int
		stockCode := "0"
		if len(item.C) > 0 {
			if item.C[0] == '"' {
				var s string
				if err := json.Unmarshal(item.C, &s); err == nil {
					stockCode = s
				}
			} else {
				var n float64
				if err := json.Unmarshal(item.C, &n); err == nil {
					stockCode = fmt.Sprintf("%.0f", n)
				}
			}
		}
		
		price := 0.0
		if item.P != nil {
			var p float64
			if err := json.Unmarshal(item.P, &p); err == nil {
				price = p / 1000
			}
		}

		firstZT := ""
		if len(item.Fbt) > 0 {
			var fbtVal float64
			if json.Unmarshal(item.Fbt, &fbtVal) == nil {
				firstZT = fmt.Sprintf("%.0f", fbtVal)
			}
		}
		lastZT := ""
		if len(item.Lbt) > 0 {
			var lbtVal float64
			if json.Unmarshal(item.Lbt, &lbtVal) == nil {
				lastZT = fmt.Sprintf("%.0f", lbtVal)
			}
		}

		results = append(results, map[string]interface{}{
			"date":           qdateStr,
			"stock_code":     stockCode,
			"stock_name":     item.N,
			"market":         item.M,
			"price":          price,
			"change_pct":     item.Zdp,
			"amount":         item.Amount,
			"circ_market":    item.Ltsz,
			"circ_share":     item.Tshare,
			"turnover_rate":  item.Hs,
			"continuous":     item.Lbc,
			"first_zt_time":  firstZT,
			"last_zt_time":   lastZT,
			"main_inflow":    item.Fund,
			"open_times":     item.Zbc,
			"sector":         item.Hybk,
			"zt_days":        item.Zttj["days"],
			"zt_count":       item.Zttj["ct"],
		})
	}
	return results, nil
}

// LimitDownPool fetches today's limit-down stock pool.
// push2ex.getTopicDTPool.
func (e *EastMoneyScraper) LimitDownPool(date string) ([]map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://push2ex.eastmoney.com/getTopicDTPool"+
			"?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wz.ztzt"+
			"&Pageindex=0&pagesize=3000&sort=fund:asc&date=%s",
		date,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type dtItem struct {
		C    int             `json:"c"`
		M    int             `json:"m"`
		N    string          `json:"n"`
		P    json.RawMessage `json:"p"`
		Zdp  float64         `json:"zdp"`
		Amount float64       `json:"amount"`
		Ltsz float64       `json:"ltsz"`
		Tshare float64      `json:"tshare"`
		Hs   float64       `json:"hs"`
		Days int           `json:"days"`
		Lbt  int           `json:"lbt"`
		Fba  float64       `json:"fba"`
		Fund float64       `json:"fund"`
		Hybk string        `json:"hybk"`
	}

	var parsed struct {
		Data struct {
			Pool []dtItem `json:"pool"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Pool))
	for _, item := range parsed.Data.Pool {
		price := 0.0
		if item.P != nil {
			var p float64
			if err := json.Unmarshal(item.P, &p); err == nil {
				price = p / 1000
			}
		}

		lastDT := ""
		if item.Lbt > 0 {
			lastDT = fmt.Sprintf("%06d", item.Lbt)
		}

		results = append(results, map[string]interface{}{
			"date":        date,
			"stock_code":  fmt.Sprintf("%d", item.C),
			"stock_name":  item.N,
			"market":      item.M,
			"price":       price,
			"change_pct":  item.Zdp,
			"amount":      item.Amount,
			"circ_market": item.Ltsz,
			"circ_share":  item.Tshare,
			"turnover_rate": item.Hs,
			"days":        item.Days,
			"last_dt_time": lastDT,
			"seal_amount": item.Fba,
			"main_inflow": item.Fund,
			"sector":      item.Hybk,
		})
	}
	return results, nil
}

// YesterdayLimitUp fetches yesterday's limit-up stocks and their today performance.
func (e *EastMoneyScraper) YesterdayLimitUp(date string) ([]map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://push2ex.eastmoney.com/getYesterdayZTPool"+
			"?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wz.ztzt"+
			"&Pageindex=0&pagesize=3000&sort=zs:desc&date=%s",
		date,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type yzItem struct {
		C    int             `json:"c"`
		M    int             `json:"m"`
		N    string          `json:"n"`
		P    json.RawMessage `json:"p"`
		Ztp  json.RawMessage `json:"ztp"`
		Zdp  float64         `json:"zdp"`
		Amount float64       `json:"amount"`
		Ltsz float64       `json:"ltsz"`
		Hs   float64       `json:"hs"`
		Zf   float64       `json:"zf"`
		Zs   float64       `json:"zs"`
		Yfbt int           `json:"yfbt"`
		Ylbc int           `json:"ylbc"`
		Hybk string        `json:"hybk"`
		Zttj map[string]interface{} `json:"zttj"`
	}

	var parsed struct {
		Data struct {
			Pool []yzItem `json:"pool"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Pool))
	for _, item := range parsed.Data.Pool {
		price := 0.0
		if item.P != nil {
			var p float64
			if err := json.Unmarshal(item.P, &p); err == nil {
				price = p / 1000
			}
		}
		ztPrice := 0.0
		if item.Ztp != nil {
			var p float64
			if err := json.Unmarshal(item.Ztp, &p); err == nil {
				ztPrice = p / 1000
			}
		}

		yesterdayTime := ""
		if item.Yfbt > 0 {
			yesterdayTime = fmt.Sprintf("%06d", item.Yfbt)
		}

		results = append(results, map[string]interface{}{
			"date":           date,
			"stock_code":     fmt.Sprintf("%d", item.C),
			"stock_name":     item.N,
			"market":         item.M,
			"price":          price,
			"zt_price":       ztPrice,
			"change_pct":     item.Zdp,
			"amount":         item.Amount,
			"circ_market":    item.Ltsz,
			"turnover_rate":  item.Hs,
			"amplitude":      item.Zf,
			"open_ratio":     item.Zs,
			"yesterday_time": yesterdayTime,
			"yesterday_cont": item.Ylbc,
			"sector":         item.Hybk,
			"zt_days":        item.Zttj["days"],
			"zt_count":       item.Zttj["ct"],
		})
	}
	return results, nil
}

// StockChanges fetches real-time market changes (异动股).
// push2ex.getAllStockChanges with type codes.
func (e *EastMoneyScraper) StockChanges(changeType string) ([]map[string]interface{}, error) {
	typeMap := map[string]string{
		"rocket":      "8201", // 火箭发射
		"bounce":      "8202", // 快速反弹
		"accel_drop":  "8203", // 加速下跌
		"dive":        "8204", // 高台跳水
		"big_buy":     "8193", // 大笔买入
		"big_sell":    "8194", // 大笔卖出
		"limit_up":    "8205", // 封涨停板
		"limit_down":  "8206", // 封跌停板
		"open_down":   "8207", // 打开跌停板
		"open_up":     "8208", // 打开涨停板
		"big_buy_pk":  "64",   // 有大买盘
		"big_sell_pk": "128",  // 有大卖盘
		"auction_up":  "8209", // 竞价上涨
		"auction_down":"8210", // 竞价下跌
		"high_5day":   "8211", // 高开5日线
		"low_5day":    "8212", // 低开5日线
		"gap_up":      "8213", // 向上缺口
		"gap_down":    "8214", // 向下缺口
		"high_60day":  "8215", // 60日新高
		"low_60day":   "8216", // 60日新低
	}
	ct, ok := typeMap[changeType]
	if !ok {
		ct = "8201"
	}

	url := fmt.Sprintf(
		"https://push2ex.eastmoney.com/getAllStockChanges"+
			"?type=%s&ut=7eea3edcaed734bea9cbfc24409ed989"+
			"&pageindex=0&pagesize=10000&dpt=wzchanges",
		ct,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type changeItem struct {
		Tm int    `json:"tm"`
		C  string `json:"c"`
		M  int    `json:"m"`
		N  string `json:"n"`
		I  string `json:"i"`
		T  int    `json:"t"`
	}

	var parsed struct {
		Data struct {
			TC      int         `json:"tc"`
			Allstock []changeItem `json:"allstock"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	typeNames := map[int]string{
		8201: "火箭发射", 8202: "快速反弹", 8203: "加速下跌", 8204: "高台跳水",
		8193: "大笔买入", 8194: "大笔卖出", 8205: "封涨停板", 8206: "封跌停板",
		8207: "打开跌停板", 8208: "打开涨停板", 64: "有大买盘", 128: "有大卖盘",
		8209: "竞价上涨", 8210: "竞价下跌", 8211: "高开5日线", 8212: "低开5日线",
		8213: "向上缺口", 8214: "向下缺口", 8215: "60日新高", 8216: "60日新低",
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Allstock))
	for _, item := range parsed.Data.Allstock {
		tm := item.Tm
		if tm > 0 && tm < 100000 {
			tm = tm * 10
		}
		timeStr := fmt.Sprintf("%06d", tm)

		results = append(results, map[string]interface{}{
			"stock_code":  item.C,
			"stock_name":  item.N,
			"market":      item.M,
			"time":        timeStr,
			"change_pct":  item.I,
			"change_type": typeNames[item.T],
		})
	}
	return results, nil
}

// HotRank fetches popular stock ranking from 同花顺 + enriches with EastMoney quotes.
func (e *EastMoneyScraper) HotRank(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	url := "https://dq.10jqka.com.cn/fuyao/hot_list_data/out/hot_list/v1/stock" +
		"?stock_type=a&type=hour&list_type=normal"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.10jqka.com.cn/")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type hotItem struct {
		Code    string `json:"code"`
		Name    string `json:"name"`
		Tag     map[string]interface{} `json:"tag"`
	}

	var parsed struct {
		Status int `json:"status_code"`
		Data struct {
			StockList []hotItem `json:"stock_list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	if parsed.Status != 0 {
		return nil, fmt.Errorf("ths status_code=%d", parsed.Status)
	}

	list := parsed.Data.StockList
	if len(list) > limit {
		list = list[:limit]
	}

	// Enrich with EastMoney quotes
	codes := make([]string, 0, len(list))
	for _, item := range list {
		codes = append(codes, item.Code)
	}

	quotes, err := e.RealtimeQuote(codes)
	if err != nil {
		// fallback without quotes
		results := make([]map[string]interface{}, 0, len(list))
		for i, item := range list {
			results = append(results, map[string]interface{}{
				"rank":   i + 1,
				"code":   item.Code,
				"name":   item.Name,
				"tag":    item.Tag,
			})
		}
		return results, nil
	}

	qMap := make(map[string]map[string]interface{})
	for _, q := range quotes {
		if code, ok := q["stock_code"].(string); ok {
			qMap[code] = q
		}
	}

	results := make([]map[string]interface{}, 0, len(list))
	for i, item := range list {
		r := map[string]interface{}{
			"rank":   i + 1,
			"code":   item.Code,
			"name":   item.Name,
			"tag":    item.Tag,
		}
		if q, ok := qMap[item.Code]; ok {
			r["price"] = q["price"]
			r["change_pct"] = q["change_pct"]
			r["change_amt"] = q["change_amt"]
		}
		results = append(results, r)
	}
	return results, nil
}

// KlineHistory fetches historical K-lines via push2his.
// klt: 101=day, 102=week, 103=month, 104=5min, 105=15min, 106=30min, 107=60min
func (e *EastMoneyScraper) KlineHistory(secid, klt string, count int) ([]map[string]interface{}, error) {
	if klt == "" {
		klt = "101"
	}
	if count <= 0 || count > 10000 {
		count = 1000
	}

	url := fmt.Sprintf(
		"https://push2his.eastmoney.com/api/qt/stock/kline/get"+
			"?secid=%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61"+
			"&klt=%s&fqt=1&beg=19900101&end=20500101&lmt=%d&_%d",
		secid, klt, count, time.Now().UnixMilli(),
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type rawResp struct {
		Data struct {
			Code    string   `json:"code"`
			Market  int      `json:"market"`
			Name    string   `json:"name"`
			Klines  []string `json:"klines"`
		} `json:"data"`
	}
	var rr rawResp
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(rr.Data.Klines))
	for _, line := range rr.Data.Klines {
		parts := strings.Split(line, ",")
		if len(parts) < 7 {
			continue
		}
		k := make(map[string]interface{})
		k["date"] = parts[0]
		var v float64
		fmt.Sscanf(parts[1], "%f", &v); k["open"] = v
		fmt.Sscanf(parts[2], "%f", &v); k["close"] = v
		fmt.Sscanf(parts[3], "%f", &v); k["high"] = v
		fmt.Sscanf(parts[4], "%f", &v); k["low"] = v
		fmt.Sscanf(parts[5], "%f", &v); k["volume"] = v
		fmt.Sscanf(parts[6], "%f", &v); k["amount"] = v
		if len(parts) >= 11 {
			fmt.Sscanf(parts[7], "%f", &v); k["amplitude"] = v
			fmt.Sscanf(parts[8], "%f", &v); k["change_pct"] = v
			fmt.Sscanf(parts[9], "%f", &v); k["change_amt"] = v
			fmt.Sscanf(parts[10], "%f", &v); k["turnover_rate"] = v
		}
		results = append(results, k)
	}
	return results, nil
}

// NorthBoundDaily fetches daily northbound capital flow via datacenter.
func (e *EastMoneyScraper) NorthBoundDaily(date string) ([]map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://datacenter-web.eastmoney.com/api/data/v1/get"+
			"?reportName=RPT_MUTUAL_HISTORY_DATA"+
			"&columns=ALL&filter=(TRADE_DATE='%s')"+
			"&pageNumber=1&pageSize=1&sortColumns=SCHEDULE_DATE&sortTypes=-1",
		date,
	)

	body, err := e.doJSON(url, map[string]string{
		"Referer": "https://data.eastmoney.com/cjsj/hsgt.html",
	})
	if err != nil {
		return nil, err
	}

	type rawResp struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
			Total int                     `json:"total"`
		} `json:"result"`
	}
	var rr rawResp
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	return rr.Result.Data, nil
}

// NorthBoundTop10 fetches top 10 holdings changed by northbound.
func (e *EastMoneyScraper) NorthBoundTop10(date string) ([]map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://datacenter-web.eastmoney.com/api/data/v1/get"+
			"?reportName=RPT_MUTUAL_TOP10DEAL"+
			"&columns=ALL&filter=(TRADE_DATE='%s')"+
			"&pageNumber=1&pageSize=10&sortColumns=NET_BUY_AMOUNT&sortTypes=-1",
		date,
	)

	body, err := e.doJSON(url, map[string]string{
		"Referer": "https://data.eastmoney.com/cjsj/hsgt.html",
	})
	if err != nil {
		return nil, err
	}

	type rawResp struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"result"`
	}
	var rr rawResp
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	return rr.Result.Data, nil
}

// UpDownCount fetches up/down stock counts and other market stats.
func (e *EastMoneyScraper) UpDownCount(date string) (map[string]interface{}, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}

	url := fmt.Sprintf(
		"https://datacenter-web.eastmoney.com/api/data/v1/get"+
			"?reportName=RPT_TABLE_UPDOWN_STAT"+
			"&columns=ALL&filter=(TRADE_DATE='%s')"+
			"&pageNumber=1&pageSize=1",
		date,
	)

	body, err := e.doJSON(url, map[string]string{
		"Referer": "https://data.eastmoney.com/",
	})
	if err != nil {
		return nil, err
	}

	type rawResp struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"result"`
	}
	var rr rawResp
	if err := json.Unmarshal(body, &rr); err != nil {
		return nil, err
	}

	if len(rr.Result.Data) > 0 {
		return rr.Result.Data[0], nil
	}
	return nil, fmt.Errorf("no up/down data for %s", date)
}

// SecurityList fetches stock list with filters.
func (e *EastMoneyScraper) SecurityList(fs, fields string, pn, pz int) ([]map[string]interface{}, error) {
	if fs == "" {
		fs = "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048"
	}
	if fields == "" {
		fields = "f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f14,f15,f16,f17,f18,f20,f21,f23"
	}
	if pn <= 0 {
		pn = 1
	}
	if pz <= 0 || pz > 2000 {
		pz = 200
	}

	url := fmt.Sprintf(
		"http://push2delay.eastmoney.com/api/qt/clist/get"+
			"?pn=%d&pz=%d&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281"+
			"&fltt=2&invt=2&fid=f3&fs=%s&fields=%s",
		pn, pz, fs, fields,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Data struct {
			Total int                    `json:"total"`
			Diff  []map[string]interface{} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	return parsed.Data.Diff, nil
}

// SymbolInfo fetches detailed symbol info via push2.
func (e *EastMoneyScraper) SymbolInfo(secid string) (map[string]interface{}, error) {
	url := fmt.Sprintf(
		"https://push2delay.eastmoney.com/api/qt/stock/get"+
			"?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f163,f168,f169,f170,f171,f173,f292",
		secid,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		return data, nil
	}
	return result, nil
}

// BelongBoard fetches boards a stock belongs to.
func (e *EastMoneyScraper) BelongBoard(secid string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf(
		"https://push2delay.eastmoney.com/api/qt/clist/get"+
			"?pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f3&fs=%s+f:!50&fields=f12,f14",
		secid,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	type rawItem struct {
		F12 string `json:"f12"`
		F14 string `json:"f14"`
	}

	var parsed struct {
		Data struct {
			Diff []rawItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(parsed.Data.Diff))
	for _, item := range parsed.Data.Diff {
		results = append(results, map[string]interface{}{
			"board_code": item.F12,
			"board_name": item.F14,
		})
	}
	return results, nil
}

// SecurityCount fetches count of securities by market.
func (e *EastMoneyScraper) SecurityCount(secid string) (map[string]interface{}, error) {
	url := fmt.Sprintf(
		"https://push2delay.eastmoney.com/api/qt/ulist.np/get"+
			"?fields=f128,f129,f130&secids=%s",
		secid,
	)

	body, err := e.doJSON(url, nil)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Data struct {
			Diff []struct {
				F128 float64 `json:"f128"`
				F129 float64 `json:"f129"`
				F130 float64 `json:"f130"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	if len(parsed.Data.Diff) == 0 {
		return nil, fmt.Errorf("no data for secid %s", secid)
	}

	d := parsed.Data.Diff[0]
	return map[string]interface{}{
		"total": int(d.F128),
		"up":    int(d.F129),
		"down":  int(d.F130),
	}, nil
}

// RandomDelay sleeps for a random duration to avoid rate limiting.
func (e *EastMoneyScraper) RandomDelay() {
	e.mu.Lock()
	defer e.mu.Unlock()
	delay := time.Duration(rand.Intn(1500)+1500) * time.Millisecond
	time.Sleep(delay)
}

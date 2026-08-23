package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleVolumeProfile(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	reply, err := tcp.GetVolumeProfile(code, market)
	if err != nil {
		writeError(w, 500, "获取量价分布失败: "+err.Error())
		return
	}
	profiles := make([]map[string]interface{}, len(reply.VolProfiles))
	for i, p := range reply.VolProfiles {
		profiles[i] = map[string]interface{}{"price": p.Price, "vol": p.Vol, "buy": p.Buy, "sell": p.Sell}
	}
	bidLevels := make([]map[string]interface{}, len(reply.BidLevels))
	for i, l := range reply.BidLevels {
		bidLevels[i] = map[string]interface{}{"price": l.Price, "vol": l.Vol}
	}
	askLevels := make([]map[string]interface{}, len(reply.AskLevels))
	for i, l := range reply.AskLevels {
		askLevels[i] = map[string]interface{}{"price": l.Price, "vol": l.Vol}
	}
	writeJSON(w, map[string]interface{}{
		"code": reply.Code, "market": int(reply.Market),
		"close": reply.Close, "open": reply.Open, "high": reply.High, "low": reply.Low,
		"pre_close": reply.PreClose, "vol": reply.Vol, "amount": reply.Amount,
		"in_vol": reply.InVol, "out_vol": reply.OutVol,
		"bid": bidLevels, "ask": askLevels,
		"profile_count": int(reply.Count), "profile": profiles,
		"server_time": reply.ServerTime,
	})
}

func (s *Server) handleIndexInfo(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	reply, err := tcp.GetIndexInfo(code, market)
	if err != nil {
		writeError(w, 500, "获取指数详情失败: "+err.Error())
		return
	}
	orders := make([]map[string]interface{}, len(reply.Orders))
	for i, o := range reply.Orders {
		orders[i] = map[string]interface{}{"price": o.Price, "vol": o.Vol}
	}
	writeJSON(w, map[string]interface{}{
		"code": reply.Code, "market": int(reply.Market),
		"close": reply.Close, "open": reply.Open, "high": reply.High, "low": reply.Low,
		"pre_close": reply.PreClose, "diff": reply.Diff,
		"vol": reply.Vol, "amount": reply.Amount,
		"up_count": reply.UpCount, "down_count": reply.DownCount,
		"order_count": int(reply.OrderCount), "orders": orders,
		"server_time": reply.ServerTime,
	})
}

func (s *Server) handleIndexMomentum(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	reply, err := tcp.GetIndexMomentum(code, market)
	if err != nil {
		writeError(w, 500, "获取指数动量失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{
		"code": code, "market": market,
		"count": int(reply.Count), "values": reply.Values,
	})
}

func (s *Server) handleHistoryOrders(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	date := r.URL.Query().Get("date")
	if code == "" || !ok || date == "" {
		writeError(w, 400, "code, market, date 参数必填 (date: YYYYMMDD)")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	dateStr := strings.NewReplacer("-", "", "/", "", " ", "").Replace(date)
	if len(dateStr) != 8 {
		writeError(w, 400, "date 格式应为 YYYYMMDD")
		return
	}
	dateNum, _ := strconv.ParseUint(dateStr, 10, 32)
	reply, err := tcp.GetHistoryOrders(uint32(dateNum), code, market)
	if err != nil {
		writeError(w, 500, "获取历史委托单失败: "+err.Error())
		return
	}
	orders := make([]map[string]interface{}, len(reply.List))
	for i, o := range reply.List {
		orders[i] = map[string]interface{}{"price": o.Price, "vol": o.Vol}
	}
	writeJSON(w, map[string]interface{}{
		"code": code, "market": market, "date": dateStr,
		"pre_close": reply.PreClose, "count": int(reply.Count), "orders": orders,
	})
}

func (s *Server) handleExQuotesList(w http.ResponseWriter, r *http.Request) {
	catStr := r.URL.Query().Get("category")
	startStr := r.URL.Query().Get("start")
	countStr := r.URL.Query().Get("count")
	sortStr := r.URL.Query().Get("sort")
	reverseStr := r.URL.Query().Get("reverse")
	cat, _ := strconv.ParseUint(catStr, 10, 8)
	if cat == 0 {
		writeError(w, 400, "category 参数必填 (扩展市场分类代码)")
		return
	}
	start, _ := strconv.ParseUint(startStr, 10, 16)
	count, _ := strconv.ParseUint(countStr, 10, 16)
	if count == 0 {
		count = 100
	}
	sortType, _ := strconv.ParseUint(sortStr, 10, 16)
	reverse := reverseStr == "true" || reverseStr == "1"
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	reply, err := tcp.ExGetQuotesList(uint8(cat), uint16(sortType), reverse, uint16(start), uint16(count))
	if err != nil {
		writeError(w, 500, "获取扩展市场报价失败: "+err.Error())
		return
	}
	items := make([]map[string]interface{}, len(reply.List))
	for i, sq := range reply.List {
		bidLevels := make([]map[string]interface{}, len(sq.BidLevels))
		for j, l := range sq.BidLevels {
			bidLevels[j] = map[string]interface{}{"price": l.Price, "vol": l.Vol}
		}
		askLevels := make([]map[string]interface{}, len(sq.AskLevels))
		for j, l := range sq.AskLevels {
			askLevels[j] = map[string]interface{}{"price": l.Price, "vol": l.Vol}
		}
		items[i] = map[string]interface{}{
			"category": int(sq.Category), "code": sq.Code,
			"open": sq.Open, "high": sq.High, "low": sq.Low, "close": sq.Close,
			"pre_close": sq.PreClose,
			"vol": sq.Vol, "amount": sq.Amount,
			"in_vol": sq.InVol, "out_vol": sq.OutVol,
			"bid": bidLevels, "ask": askLevels,
		}
	}
	writeJSON(w, map[string]interface{}{
		"category": int(cat), "start": int(start), "count": int(count),
		"sort_type": int(sortType), "reverse": reverse,
		"total": len(items), "data": items,
	})
}

func (s *Server) handleXDXR(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	reply, err := tcp.GetXDXRInfo(code, market)
	if err != nil {
		writeError(w, 500, "获取除权除息信息失败: "+err.Error())
		return
	}
	items := make([]map[string]interface{}, len(reply.List))
	for i, item := range reply.List {
		itemData := map[string]interface{}{
			"code":     item.Code,
			"market":   int(item.Market),
			"date":     item.Date.Format("2006-01-02"),
			"category": int(item.Category),
			"name":     item.Name,
		}
		if item.Fenhong != nil {
			itemData["fenhong"] = *item.Fenhong
		}
		items[i] = itemData
	}
	writeJSON(w, map[string]interface{}{
		"code": reply.Code, "market": int(reply.Market),
		"count": int(reply.Count), "data": items,
	})
}

func (s *Server) handleFundFlow(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	tcp := s.getTCPClient()
	if tcp == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	code = normalizeCode(code)
	reply, err := tcp.GetMACCapitalFlow(code, market)
	if err != nil {
		writeError(w, 500, "获取当日资金流向失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{
		"code":            code,
		"market":          market,
		"today_main_in":   reply.TodayMainIn,
		"today_main_out":  reply.TodayMainOut,
		"today_retail_in":  reply.TodayRetailIn,
		"today_retail_out": reply.TodayRetailOut,
		"today_main_net_in":   reply.TodayMainNetIn,
		"today_retail_net_in": reply.TodayRetailNetIn,
		"five_day_main_buy":   reply.FiveDayMainBuy,
		"five_day_main_sell":  reply.FiveDayMainSell,
		"five_day_super_net":  reply.FiveDaySuperNet,
		"five_day_large_net":  reply.FiveDayLargeNet,
		"five_day_medium_net": reply.FiveDayMediumNet,
		"five_day_small_net":  reply.FiveDaySmallNet,
		"five_day_main_net_in": reply.FiveDayMainNetIn,
		"today_array":    reply.Today,
		"five_days_array": reply.FiveDays,
	})
}

func (s *Server) handleMinuteMulti(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	datesStr := r.URL.Query().Get("dates")
	if code == "" || !ok || datesStr == "" {
		writeError(w, 400, "code, market, dates 参数必填 (dates 逗号分隔 YYYYMMDD)")
		return
	}
	if s.client == nil {
		writeError(w, 503, "TDX服务不可用")
		return
	}
	dates := strings.Split(datesStr, ",")
	results := make(map[string]interface{})
	for _, d := range dates {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		body := map[string]interface{}{
			"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
			"Code": code, "Setcode": market, "Period": 104,
			"Startxh": 0, "WantNum": 240, "TQFlag": 0, "Date": d,
		}
		rawResp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
		if err != nil {
			results[d] = map[string]string{"error": err.Error()}
			continue
		}
		if rawResp == nil {
			results[d] = map[string]string{"error": "empty response"}
			continue
		}
		results[d] = rawResp.Data
	}
	writeJSON(w, map[string]interface{}{"code": code, "market": market, "dates": dates, "data": results})
}
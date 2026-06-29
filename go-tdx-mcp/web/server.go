package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/finance"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/offline"
	"github.com/tdx/go-tdx-mcp/scraper"
	"github.com/tdx/go-tdx-mcp/tdx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	client tdx.Client
	addr   string
	mux    *http.ServeMux
	wsHub  *wsHub
}

type wsHub struct {
	clients map[*wsConn]bool
	mu      sync.RWMutex
}

type wsConn struct {
	conn   *websocket.Conn
	symbol string
	stop   chan struct{}
}

func newWSHub() *wsHub {
	return &wsHub{clients: make(map[*wsConn]bool)}
}

func (h *wsHub) add(c *wsConn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *wsHub) remove(c *wsConn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *wsHub) broadcastFor(symbol string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.symbol == symbol {
			select {
			case <-c.stop:
				continue
			default:
				c.conn.WriteJSON(data)
			}
		}
	}
}

func NewServer(client tdx.Client, addr string) *Server {
	s := &Server{
		client: client,
		addr:   addr,
		mux:    http.NewServeMux(),
		wsHub:  newWSHub(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) Start() error {
	log.Printf("TDX Web API 已启动: http://%s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/api/v1/quotes", s.handleQuotes)
	s.mux.HandleFunc("/api/v1/bars", s.handleBars)
	s.mux.HandleFunc("/api/v1/indicator/list", s.handleIndicatorList)
	s.mux.HandleFunc("/api/v1/indicator/compute", s.handleIndicatorCompute)
	s.mux.HandleFunc("/api/v1/indicator/compute_all", s.handleIndicatorComputeAll)
	s.mux.HandleFunc("/api/v1/chanlun/analyze", s.handleChanlun)
	s.mux.HandleFunc("/api/v1/backtest/run", s.handleBacktest)
	s.mux.HandleFunc("/api/v1/financial/report", s.handleFinancial)
	s.mux.HandleFunc("/api/v1/announcements", s.handleAnnouncements)
	s.mux.HandleFunc("/api/v1/ex/markets", s.handleExMarkets)
	s.mux.HandleFunc("/api/v1/ex/bars", s.handleExBars)
	s.mux.HandleFunc("/api/v1/ex/quote", s.handleExQuote)
	// Offline data (requires local TDX data files)
	s.mux.HandleFunc("/api/v1/offline/daily", s.handleOfflineDaily)
	s.mux.HandleFunc("/api/v1/offline/min", s.handleOfflineMin)
	// Market overview & macro
	s.mux.HandleFunc("/api/v1/market-overview", s.handleMarketOverview)
	s.mux.HandleFunc("/api/v1/macro-data", s.handleMacroData)
	s.mux.HandleFunc("/api/v1/news-sentiment", s.handleNewsSentiment)
	// Share board operations
	s.mux.HandleFunc("/api/v1/board/list", s.handleBoardList)
	s.mux.HandleFunc("/api/v1/board/members", s.handleBoardMembers)
	s.mux.HandleFunc("/api/v1/board/ranking", s.handleBoardRanking)
	s.mux.HandleFunc("/api/v1/capital-flow", s.handleCapitalFlow)
	s.mux.HandleFunc("/api/v1/auction", s.handleAuction)
	s.mux.HandleFunc("/api/v1/unusual", s.handleUnusual)
	s.mux.HandleFunc("/api/v1/market-stat", s.handleMarketStat)
	s.mux.HandleFunc("/api/v1/server-info", s.handleServerInfo)
	s.mux.HandleFunc("/api/v1/symbol-info", s.handleSymbolInfo)
	s.mux.HandleFunc("/api/v1/quote-list", s.handleQuoteList)
	s.mux.HandleFunc("/api/v1/security-count", s.handleSecurityCount)
	s.mux.HandleFunc("/api/v1/belong-board", s.handleBelongBoard)
	s.mux.HandleFunc("/api/v1/block", s.handleBlock)
	s.mux.HandleFunc("/api/v1/scraper", s.handleScraper)
	// Scraper module endpoints
	s.mux.HandleFunc("/api/v1/scraper/sector-boards", s.handleScraperSectorBoards)
	s.mux.HandleFunc("/api/v1/scraper/northbound-flow", s.handleScraperNorthboundFlow)
	s.mux.HandleFunc("/api/v1/scraper/northbound-stocks", s.handleScraperNorthboundStocks)
	s.mux.HandleFunc("/api/v1/scraper/northbound-holders", s.handleScraperNorthboundHolders)
	s.mux.HandleFunc("/api/v1/scraper/fund-nav", s.handleScraperFundNav)
	s.mux.HandleFunc("/api/v1/scraper/margin-trade", s.handleScraperMarginTrade)
	s.mux.HandleFunc("/api/v1/scraper/fund-holding", s.handleScraperFundHolding)
	s.mux.HandleFunc("/api/v1/scraper/fund-search", s.handleScraperFundSearch)
	s.mux.HandleFunc("/api/v1/scraper/hkus-quote", s.handleScraperHKUSQuote)
	s.mux.HandleFunc("/api/v1/scraper/crypto", s.handleScraperCrypto)
	// Offline extras
	s.mux.HandleFunc("/api/v1/offline/gbbq", s.handleOfflineGBBQ)
	s.mux.HandleFunc("/api/v1/offline/blocks", s.handleOfflineBlocks)
	s.mux.HandleFunc("/api/v1/offline/home", s.handleOfflineHome)
	s.mux.HandleFunc("/api/v1/offline/ex-files", s.handleOfflineExFiles)
	s.mux.HandleFunc("/api/v1/offline/ex-daily", s.handleOfflineExDaily)
	// WebSocket real-time feed
	s.mux.HandleFunc("/ws/realtime/", s.handleWebSocket)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func queryParam(r *http.Request, key, def string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	return v
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryFloat(r *http.Request, key string, def float64) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

func parseMarket(r *http.Request) (int, bool) {
	marketStr := r.URL.Query().Get("market")
	if marketStr == "" {
		return -1, false
	}
	marketStr = strings.ToLower(strings.TrimSpace(marketStr))
	switch marketStr {
	case "sz", "0":
		return 0, true
	case "sh", "1":
		return 1, true
	case "bj", "2":
		return 2, true
	default:
		n, err := strconv.Atoi(marketStr)
		if err == nil {
			return n, true
		}
		return -1, false
	}
}

func normalizeCode(rawCode string) string {
	c := strings.TrimSpace(rawCode)
	c = strings.TrimPrefix(c, "SZ")
	c = strings.TrimPrefix(c, "sz")
	c = strings.TrimPrefix(c, "SH")
	c = strings.TrimPrefix(c, "sh")
	c = strings.TrimPrefix(c, "BJ")
	c = strings.TrimPrefix(c, "bj")
	return c
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, 404, "page not found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="zh-CN">
<head><title>TDX Finance MCP v1.0.0</title>
<style>
body{font-family:system-ui,sans-serif;max-width:800px;margin:40px auto;padding:0 20px;color:#333}
h1{color:#1a73e8}h2{border-bottom:2px solid #eee;padding-bottom:8px}
table{width:100%;border-collapse:collapse}
td,th{border:1px solid #ddd;padding:8px 12px;text-align:left}
th{background:#f5f5f5}code{background:#f0f0f0;padding:2px 6px;border-radius:3px;font-size:13px}
a{color:#1a73e8}
</style></head>
<body>
<h1>TDX Finance MCP v1.0.0</h1>
<p>通达信 A 股金融数据 Web API · REST + WebSocket</p>
<h2>核心接口</h2>
<table>
<tr><th>端点</th><th>方法</th><th>说明</th></tr>
<tr><td><a href="/api/v1/health">/api/v1/health</a></td><td>GET</td><td>健康检查</td></tr>
<tr><td><a href="/api/v1/quotes?codes=000001">/api/v1/quotes</a></td><td>GET/POST</td><td>实时行情</td></tr>
<tr><td><a href="/api/v1/bars?code=000001&market=0&period=day&count=30">/api/v1/bars</a></td><td>GET</td><td>K线数据</td></tr>
<tr><td><a href="/api/v1/financial/report?code=000001&type=lrb">/api/v1/financial/report</a></td><td>GET</td><td>财务报表</td></tr>
<tr><td><a href="/api/v1/announcements?code=000001">/api/v1/announcements</a></td><td>GET</td><td>公司公告</td></tr>
<tr><td><a href="/api/v1/macro-data?indicator=LPR">/api/v1/macro-data</a></td><td>GET</td><td>宏观经济(CPI/GDP/PMI/LPR/SHIBOR/M2)</td></tr>
<tr><td><a href="/api/v1/market-overview">/api/v1/market-overview</a></td><td>GET</td><td>全市场概览</td></tr>
</table>
<h2>分析工具</h2>
<table>
<tr><td><a href="/api/v1/indicator/compute_all?code=000001&market=0&indicators=MACD,KDJ">/api/v1/indicator/compute_all</a></td><td>GET</td><td>技术指标计算(34种)</td></tr>
<tr><td><a href="/api/v1/chanlun/analyze?code=000001&market=0">/api/v1/chanlun/analyze</a></td><td>GET</td><td>缠论分析</td></tr>
<tr><td><a href="/api/v1/backtest/run?code=000001&market=0&strategy=MA_CROSS">/api/v1/backtest/run</a></td><td>GET</td><td>策略回测(7种)</td></tr>
<tr><td><a href="/api/v1/news-sentiment?code=000001">/api/v1/news-sentiment</a></td><td>GET</td><td>新闻情感分析</td></tr>
</table>
<h2>板块与资金</h2>
<table>
<tr><td><a href="/api/v1/board/list?board_type=HY&count=50">/api/v1/board/list</a></td><td>GET</td><td>板块列表(HY/CN/DY)</td></tr>
<tr><td><a href="/api/v1/board/members?board_symbol=HY001004&count=10">/api/v1/board/members</a></td><td>GET</td><td>板块成分股</td></tr>
<tr><td><a href="/api/v1/board/ranking?board_type=HY&top_n=10">/api/v1/board/ranking</a></td><td>GET</td><td>板块排行</td></tr>
<tr><td><a href="/api/v1/capital-flow?code=000001&market=0">/api/v1/capital-flow</a></td><td>GET</td><td>资金流向</td></tr>
<tr><td><a href="/api/v1/auction?code=000001&market=0">/api/v1/auction</a></td><td>GET</td><td>集合竞价</td></tr>
<tr><td><a href="/api/v1/unusual?market=0&count=50">/api/v1/unusual</a></td><td>GET</td><td>异常波动</td></tr>
<tr><td><a href="/api/v1/market-stat">/api/v1/market-stat</a></td><td>GET</td><td>市场统计</td></tr>
<tr><td><a href="/api/v1/server-info">/api/v1/server-info</a></td><td>GET</td><td>服务器信息</td></tr>
</table>
<h2>证券信息</h2>
<table>
<tr><td><a href="/api/v1/symbol-info?code=000001&market=0">/api/v1/symbol-info</a></td><td>GET</td><td>证券信息</td></tr>
<tr><td><a href="/api/v1/quote-list?count=20&sort_type=CHANGE_PCT">/api/v1/quote-list</a></td><td>GET</td><td>行情列表</td></tr>
<tr><td><a href="/api/v1/security-count?market=SZ">/api/v1/security-count</a></td><td>GET</td><td>证券数量</td></tr>
<tr><td><a href="/api/v1/belong-board?code=000001&market=0">/api/v1/belong-board</a></td><td>GET</td><td>所属板块</td></tr>
<tr><td><a href="/api/v1/block?filename=block_gy.dat">/api/v1/block</a></td><td>GET</td><td>板块文件</td></tr>
</table>
<h2>Scraper 模块</h2>
<table>
<tr><td><a href="/api/v1/scraper/sector-boards?board_type=HY">/api/v1/scraper/sector-boards</a></td><td>GET</td><td>板块数据(行业/概念/地域)</td></tr>
<tr><td><a href="/api/v1/scraper/northbound-flow?days=5">/api/v1/scraper/northbound-flow</a></td><td>GET</td><td>北向资金流向</td></tr>
<tr><td><a href="/api/v1/scraper/northbound-stocks?count=10&market=all">/api/v1/scraper/northbound-stocks</a></td><td>GET</td><td>北向资金持仓(沪/深/全部)</td></tr>
<tr><td><a href="/api/v1/scraper/northbound-holders?count=10">/api/v1/scraper/northbound-holders</a></td><td>GET</td><td>北向资金机构持仓排名</td></tr>
<tr><td><a href="/api/v1/scraper/fund-nav?code=110011">/api/v1/scraper/fund-nav</a></td><td>GET</td><td>基金净值(goquery)</td></tr>
<tr><td><a href="/api/v1/scraper/margin-trade">/api/v1/scraper/margin-trade</a></td><td>GET</td><td>融资融券(东财datacenter)</td></tr>
<tr><td><a href="/api/v1/scraper/fund-holding?code=000001">/api/v1/scraper/fund-holding</a></td><td>GET</td><td>基金持仓</td></tr>
<tr><td><a href="/api/v1/scraper/fund-search?keyword=科技&page_size=10">/api/v1/scraper/fund-search</a></td><td>GET</td><td>基金搜索</td></tr>
<tr><td><a href="/api/v1/scraper/hkus-quote?code=00700&market=hk">/api/v1/scraper/hkus-quote</a></td><td>GET</td><td>港股/美股报价</td></tr>
<tr><td><a href="/api/v1/scraper/crypto?symbols=bitcoin,ethereum">/api/v1/scraper/crypto</a></td><td>GET</td><td>加密货币(Binance)</td></tr>
<tr><td><a href="/api/v1/scraper?query=贵州茅台&source=all">/api/v1/scraper</a></td><td>GET</td><td>综合爬虫(iwcy/xiaoda/eastmoney)</td></tr>
</table>
<h2>扩展市场</h2>
<table>
<tr><td><a href="/api/v1/ex/markets">/api/v1/ex/markets</a></td><td>GET</td><td>扩展市场列表</td></tr>
<tr><td><a href="/api/v1/ex/bars?ex_market=HK_MAIN_BOARD&code=00700">/api/v1/ex/bars</a></td><td>GET</td><td>扩展市场K线</td></tr>
<tr><td><a href="/api/v1/ex/quote?ex_market=HK_MAIN_BOARD&code=00700">/api/v1/ex/quote</a></td><td>GET</td><td>扩展市场报价</td></tr>
</table>
<h2>离线数据</h2>
<table>
<tr><td>/api/v1/offline/daily?market=sz&code=000001</td><td>GET</td><td>离线日线数据</td></tr>
<tr><td>/api/v1/offline/min?market=sz&code=000001</td><td>GET</td><td>离线分钟线</td></tr>
<tr><td>/api/v1/offline/gbbq?path=/path/to/gbbq.dat</td><td>GET</td><td>股本变迁</td></tr>
<tr><td>/api/v1/offline/blocks?path=/path/to</td><td>GET</td><td>板块文件</td></tr>
<tr><td>/api/v1/offline/home</td><td>GET</td><td>检测通达信目录</td></tr>
<tr><td>/api/v1/offline/ex-files?vipdoc=/path/to</td><td>GET</td><td>扩展市场文件</td></tr>
<tr><td>/api/v1/offline/ex-daily?code=38#2_CL</td><td>GET</td><td>扩展市场日线</td></tr>
</table>
<h2>WebSocket</h2>
<p><code>ws://host/ws/realtime/000001</code> — 实时行情推送(3秒轮询)</p>
<p style="margin-top:30px;color:#666;font-size:13px">开源免费 · MIT License · 45+ MCP 工具 + 50+ 投资技能</p>
</body></html>`))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "version": "1.0.0"})
}

func (s *Server) handleQuotes(w http.ResponseWriter, r *http.Request) {
	var stocks []struct {
		Market string `json:"market"`
		Code   string `json:"code"`
	}
	if r.Method == http.MethodPost {
		var req struct {
			Stocks []struct {
				Market string `json:"market"`
				Code   string `json:"code"`
			} `json:"stocks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "无效的请求体")
			return
		}
		stocks = req.Stocks
	} else {
		codesParam := r.URL.Query().Get("codes")
		if codesParam == "" {
			writeError(w, 400, "codes 参数必填")
			return
		}
		for _, c := range strings.Split(codesParam, ",") {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			setcode := "0"
			if strings.HasPrefix(c, "6") || strings.HasPrefix(c, "SH") {
				setcode = "1"
			} else if strings.HasPrefix(c, "4") || strings.HasPrefix(c, "8") {
				setcode = "2"
			}
			code := strings.TrimPrefix(strings.TrimPrefix(c, "SZ"), "SH")
			stocks = append(stocks, struct {
				Market string `json:"market"`
				Code   string `json:"code"`
			}{Market: setcode, Code: code})
		}
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	results := make([]interface{}, 0, len(stocks))
	for _, st := range stocks {
		setcodeInt := 0
		switch st.Market {
		case "1":
			setcodeInt = 1
		case "2":
			setcodeInt = 2
		default:
			setcodeInt = 0
		}
		// Use eastmoney push2 API (TDX PBHQInfo returns 503)
		setcodeStr := fmt.Sprintf("%d.%s", setcodeInt, st.Code)
		url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f168,f169,f170,f171", setcodeStr)
		respHTTP, err := hc.Get(url)
		if err != nil {
			results = append(results, map[string]interface{}{"code": st.Code, "error": err.Error()})
			continue
		}
		var data interface{}
		json.NewDecoder(respHTTP.Body).Decode(&data)
		respHTTP.Body.Close()
		results = append(results, data)
	}
	writeJSON(w, results)
}

func (s *Server) handleBars(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	code = normalizeCode(code)
	fqType := queryInt(r, "fq_type", 0)
	count := queryInt(r, "count", 200)
	period := queryParam(r, "period", "day")

	bars, err := s.fetchKlines(code, market, period, count, fqType)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取K线失败: %v", err))
		return
	}
	writeJSON(w, bars)
}

func klinePeriodToCode(period string) int {
	switch period {
	case "1min":
		return 1
	case "5min":
		return 2
	case "15min":
		return 3
	case "30min":
		return 4
	case "60min":
		return 5
	case "week":
		return 6
	case "month":
		return 7
	case "quarter":
		return 8
	case "year":
		return 9
	default:
		return 0
	}
}

func (s *Server) fetchKlinesFromOffline(code string, market int, period string, count int) ([]indicator.Bar, error) {
	marketPrefix := "sh"
	if market == 0 {
		marketPrefix = "sz"
	}
	codeZeroPadded := code
	if len(code) < 6 {
		codeZeroPadded = fmt.Sprintf("%06s", code)
	}
	filename := marketPrefix + codeZeroPadded

	var dir string
	switch period {
	case "day":
		dir = "day"
	case "5m", "15m", "30m", "60m":
		dir = "minline"
	case "week":
		dir = "weekline"
	case "month":
		dir = "monthline"
	default:
		dir = "day"
	}

	baseDir := offline.DetectHome()
	if baseDir == "" {
		return nil, fmt.Errorf("未找到TDX数据目录")
	}

	filePath := filepath.Join(baseDir, "vipdoc", marketPrefix, dir, filename+".day")
	bars, err := offline.ReadDaily(filePath)
	if err != nil {
		return nil, err
	}

	result := make([]indicator.Bar, 0, len(bars))
	for _, b := range bars {
		result = append(result, indicator.Bar{
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Vol:    b.Vol,
			Amount: b.Amount,
		})
	}
	if len(result) > count {
		result = result[len(result)-count:]
	}
	return result, nil
}

func (s *Server) handleIndicatorList(w http.ResponseWriter, r *http.Request) {
	names := []string{
		"MACD", "KDJ", "RSI", "BOLL", "DMI", "ATR", "WR", "CCI", "BIAS", "BIAS_SIGNAL",
		"OBV", "VR", "EMV", "MFI", "BRAR", "ASI", "TRIX", "DPO", "MTM", "ROC",
		"EXPMA", "BBI", "PSY", "DFMA", "CR", "KTN", "XSII", "MASS", "TAQ",
		"ZHUOYAO", "SAR", "VWAP", "AROON", "FK",
	}
	writeJSON(w, map[string]interface{}{"count": len(names), "indicators": names})
}

func (s *Server) handleIndicatorCompute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data       []indicator.Bar         `json:"data"`
		Indicators []string                `json:"indicators"`
		Params     map[string]float64      `json:"params,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效的请求体: "+err.Error())
		return
	}
	if len(req.Data) == 0 {
		writeError(w, 400, "data 不能为空")
		return
	}
	result, err := indicator.ComputeAll(req.Data, req.Indicators, req.Params)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("计算失败: %v", err))
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleIndicatorComputeAll(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	code = normalizeCode(code)
	fqType := queryInt(r, "fq_type", 0)
	count := queryInt(r, "count", 200)
	period := queryParam(r, "period", "day")

	bars, err := s.fetchKlines(code, market, period, count, fqType)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取K线失败: %v", err))
		return
	}
	indicators := strings.Split(queryParam(r, "indicators", "MACD,KDJ,RSI,BOLL,EXPMA"), ",")
	for i := range indicators {
		indicators[i] = strings.ToUpper(strings.TrimSpace(indicators[i]))
	}
	result, err := indicator.ComputeAll(bars, indicators, nil)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("计算失败: %v", err))
		return
	}
	writeJSON(w, result)
}

func (s *Server) fetchKlines(code string, market int, period string, count, fqType int) ([]indicator.Bar, error) {
	if s.client != nil {
		periodCode := klinePeriodToCode(period)
		body := map[string]interface{}{
			"Head": map[string]string{"Target": "0", "CharSet": "UTF8"},
			"Code": code, "Setcode": market, "Period": periodCode,
			"Startxh": 0, "WantNum": count, "TQFlag": fqType, "MPData": 1,
			"HasAttachInfo": 0, "HasLtgb": 0, "ForRefresh": 0, "HasIpoPrice": 0,
		}
		resp, err := s.client.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
		if err == nil && resp != nil && resp.Data != nil {
			dataBytes, _ := json.Marshal(resp.Data)
			var arrs [][]float64
			if err := json.Unmarshal(dataBytes, &arrs); err == nil && len(arrs) > 0 {
				bars := make([]indicator.Bar, 0, len(arrs))
				for _, row := range arrs {
					if len(row) >= 6 {
						bars = append(bars, indicator.Bar{Open: row[0], Close: row[1], High: row[2], Low: row[3], Vol: row[4], Amount: row[5]})
					}
				}
				if len(bars) > 0 {
					return bars, nil
				}
			}
		}
	}

	if period == "day" || period == "week" || period == "month" {
		if bars, err := s.fetchKlinesFromOffline(code, market, period, count); err == nil && len(bars) > 0 {
			return bars, nil
		}
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	periodMap := map[string]string{"day": "101", "week": "102", "month": "103", "5m": "104", "15m": "105", "30m": "106", "60m": "107"}
	klt := periodMap[period]
	if klt == "" {
		klt = "101"
	}
	fqParam := "2"
	if fqType == 1 {
		fqParam = "2"
	} else if fqType == 2 {
		fqParam = "3"
	}
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?secid=%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=%s&fqt=%s&beg=0&end=20500000&scf=&count=%d", setcodeStr, klt, fqParam, count)
	respHTTP, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	var rawData interface{}
	json.NewDecoder(respHTTP.Body).Decode(&rawData)
	return parseKlineBars(rawData)
}

func (s *Server) handleChanlun(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	code = normalizeCode(code)
	fqType := queryInt(r, "fq_type", 0)
	count := queryInt(r, "count", 200)
	period := queryParam(r, "period", "day")

	bars, err := s.fetchKlines(code, market, period, count, fqType)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取K线失败: %v", err))
		return
	}
	klines := make([]chanlun.Kline, 0, len(bars))
	for _, b := range bars {
		klines = append(klines, chanlun.Kline{
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Vol:    b.Vol,
			Amount: b.Amount,
		})
	}
	result := chanlun.Analyze(klines)
	writeJSON(w, result)
}

func (s *Server) handleBacktest(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	strategy := r.URL.Query().Get("strategy")
	if code == "" || !ok || strategy == "" {
		writeError(w, 400, "code, market, strategy 参数必填")
		return
	}
	strategy = strings.ToLower(strings.TrimSpace(strategy))
	st := backtest.NewStrategy(strategy)
	if st == nil {
		writeError(w, 400, "不支持的策略: "+strategy+" (可用策略: "+strings.Join(backtest.AvailableStrategies(), ", ")+")")
		return
	}
	fqType := queryInt(r, "fq_type", 0)
	count := queryInt(r, "count", 2000)
	period := queryParam(r, "period", "day")

	bars, err := s.fetchKlines(code, market, period, count, fqType)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取K线失败: %v", err))
		return
	}
	cash := queryFloat(r, "cash", 1000000)
	engine := backtest.NewEngine(cash)
	btResult := engine.Run(st, bars)
	btResult.Code = code
	btResult.Market = market
	btResult.Period = period
	writeJSON(w, btResult)
}

func (s *Server) handleFinancial(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	reportType := r.URL.Query().Get("type")
	if code == "" || reportType == "" {
		writeError(w, 400, "code 和 type 参数必填 (type: lrb/fzb/llb)")
		return
	}
	// Try eastmoney datacenter API first (reliable, no network issues)
	financeURL := ""
	switch strings.ToLower(reportType) {
	case "lrb":
		financeURL = fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_FN_INCOME&columns=ALL&filter=(SECURITY_CODE=%s)&pageSize=10&pageNumber=1&sortBy=REPORT_DATE&sortType=desc", code)
	case "fzb":
		financeURL = fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_FN_BALANCE&columns=ALL&filter=(SECURITY_CODE=%s)&pageSize=10&pageNumber=1&sortBy=REPORT_DATE&sortType=desc", code)
	case "llb":
		financeURL = fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DMSK_FN_CASHFLOW&columns=ALL&filter=(SECURITY_CODE=%s)&pageSize=10&pageNumber=1&sortBy=REPORT_DATE&sortType=desc", code)
	default:
		writeError(w, 400, "type 必须为: lrb(利润表)/fzb(资产负债表)/llb(现金流量表)")
		return
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	respHTTP, err := hc.Get(financeURL)
	if err != nil {
		// Fallback to sina if eastmoney fails
		report, ferr := finance.FetchReport(code, reportType)
		if ferr != nil {
			writeError(w, 500, fmt.Sprintf("获取财务数据失败: eastmoney 和 sina 均不可用: %v / %v", err, ferr))
			return
		}
		writeJSON(w, report)
		return
	}
	defer respHTTP.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleAnnouncements(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	count := queryInt(r, "count", 30)
	page := queryInt(r, "page", 1)
	// Use eastmoney announcement API (more reliable than cninfo)
	apiURL := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/ann?page_size=%d&page_index=%d&stock_list=%s", count, page, code)
	hc := &http.Client{Timeout: 10 * time.Second}
	respHTTP, err := hc.Get(apiURL)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("请求公告失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var body bytes.Buffer
	_, err = body.ReadFrom(respHTTP.Body)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取响应失败: %v", err))
		return
	}
	var result interface{}
	json.Unmarshal(body.Bytes(), &result)
	writeJSON(w, result)
}

func (s *Server) handleExMarkets(w http.ResponseWriter, r *http.Request) {
	type ExMarket struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Category    string `json:"category"`
	}
	markets := []ExMarket{
		{Code: "HK_MAIN_BOARD", Name: "港股主板", Category: "stock"},
		{Code: "HK_GEM_BOARD", Name: "港股创业板", Category: "stock"},
		{Code: "US_STOCK", Name: "美股", Category: "stock"},
		{Code: "FT_FUTURES", Name: "国内期货", Category: "futures"},
		{Code: "IP_STOCK", Name: "外盘股票", Category: "stock"},
		{Code: "IP_FUTURES", Name: "外盘期货", Category: "futures"},
		{Code: "IP_FOREX", Name: "外汇", Category: "forex"},
		{Code: "IP_INDEX", Name: "国际指数", Category: "index"},
	}
	writeJSON(w, markets)
}

func (s *Server) handleExBars(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	code := r.URL.Query().Get("code")
	if exMarket == "" || code == "" {
		writeError(w, 400, "ex_market 和 code 参数必填")
		return
	}
	// TDX TdxEx.PBFXT not available — return quote data as fallback
	// push2his does not support HK(116.)/US(117.) secid format for K-lines
	secid := ""
	switch strings.ToLower(exMarket) {
	case "hk", "h":
		secid = "116." + code
	case "us", "u":
		secid = "117." + code
	default:
		secid = "116." + code
	}
	hc := &http.Client{Timeout: 5 * time.Second}
	quoteUrl := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f49,f50,f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f71", secid)
	respHTTP, err := hc.Get(quoteUrl)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取数据失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var quoteResult interface{}
	if err := json.NewDecoder(respHTTP.Body).Decode(&quoteResult); err != nil {
		writeError(w, 500, fmt.Sprintf("解析失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{
		"warning":     "K-line data not available for " + exMarket + " stocks via EastMoney API, returning quote data as fallback",
		"ex_market":   exMarket,
		"code":        code,
		"quote_data":  quoteResult,
	})
}

func (s *Server) handleExQuote(w http.ResponseWriter, r *http.Request) {
	exMarket := r.URL.Query().Get("ex_market")
	code := r.URL.Query().Get("code")
	if exMarket == "" || code == "" {
		writeError(w, 400, "ex_market 和 code 参数必填")
		return
	}
	// TDX TdxEx.PBHQInfo not available — use EastMoney push2 for HK/US quotes
	secid := ""
	switch strings.ToLower(exMarket) {
	case "hk", "h":
		secid = "116." + code
	case "us", "u":
		secid = "117." + code
	default:
		secid = "116." + code
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f49,f50,f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f71", secid)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取行情失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleOfflineDaily(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	code := r.URL.Query().Get("code")
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	market = strings.ToLower(market)
	vipdoc := r.URL.Query().Get("vipdoc")
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			writeError(w, 404, "未找到通达信目录")
			return
		}
		vipdoc = home + "/vipdoc"
	}
	filePath := fmt.Sprintf("%s/%s/day/%s%s.day", vipdoc, market, market, code)
	bars, err := offline.ReadDaily(filePath)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取失败: %v", err))
		return
	}
	count := queryInt(r, "count", 0)
	if count > 0 && count < len(bars) {
		bars = bars[len(bars)-count:]
	}
	writeJSON(w, map[string]interface{}{
		"market": market,
		"code":   code,
		"count":  len(bars),
		"bars":   bars,
	})
}

func (s *Server) handleOfflineMin(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	code := r.URL.Query().Get("code")
	if market == "" || code == "" {
		writeError(w, 400, "market 和 code 参数必填")
		return
	}
	market = strings.ToLower(market)
	minType := queryParam(r, "min_type", "lc5")
	vipdoc := r.URL.Query().Get("vipdoc")
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			writeError(w, 404, "未找到通达信目录")
			return
		}
		vipdoc = home + "/vipdoc"
	}
	filePath := fmt.Sprintf("%s/%s/minline/%s%s.%s", vipdoc, market, market, code, minType)
	bars, err := offline.ReadMin(filePath)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取失败: %v", err))
		return
	}
	count := queryInt(r, "count", 0)
	if count > 0 && count < len(bars) {
		bars = bars[len(bars)-count:]
	}
	writeJSON(w, map[string]interface{}{
		"market":   market,
		"code":     code,
		"min_type": minType,
		"count":    len(bars),
		"bars":     bars,
	})
}

func (s *Server) handleOfflineGBBQ(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, 400, "path 参数必填")
		return
	}
	records, err := offline.ReadGBBQ(path)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{
		"path":    path,
		"count":   len(records),
		"records": records,
	})
}

func (s *Server) handleOfflineBlocks(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, 400, "path 参数必填")
		return
	}
	blocks, err := offline.ReadBlocks(path)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{
		"path":   path,
		"count":  len(blocks),
		"blocks": blocks,
	})
}

func (s *Server) handleOfflineHome(w http.ResponseWriter, r *http.Request) {
	home := offline.DetectHome()
	writeJSON(w, map[string]interface{}{
		"found": home != "",
		"home":  home,
	})
}

func (s *Server) handleOfflineExFiles(w http.ResponseWriter, r *http.Request) {
	vipdoc := r.URL.Query().Get("vipdoc")
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			writeError(w, 404, "未找到通达信目录")
			return
		}
		vipdoc = home + "/vipdoc"
	}
	type exFile struct {
		Market string `json:"market"`
		Code   string `json:"code"`
		Name   string `json:"name"`
	}
	known := []exFile{
		{Market: "38", Code: "38#2_CPI", Name: "美元指数"},
		{Market: "38", Code: "38#2_CL", Name: "美原油"},
		{Market: "71", Code: "71#2_HSI", Name: "恒生指数"},
	}
	writeJSON(w, map[string]interface{}{
		"vipdoc": vipdoc,
		"files":  known,
	})
}

func (s *Server) handleOfflineExDaily(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	vipdoc := r.URL.Query().Get("vipdoc")
	if vipdoc == "" {
		home := offline.DetectHome()
		if home == "" {
			writeError(w, 404, "未找到通达信目录")
			return
		}
		vipdoc = home + "/vipdoc"
	}
	parts := strings.SplitN(code, "#", 2)
	if len(parts) != 2 {
		writeError(w, 400, "code 格式应为 '市场#代码'")
		return
	}
	filePath := fmt.Sprintf("%s/ds/%s/day/%s.day", vipdoc, parts[0], code)
	bars, err := offline.ReadDaily(filePath)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("读取失败: %v", err))
		return
	}
	count := queryInt(r, "count", 0)
	if count > 0 && count < len(bars) {
		bars = bars[len(bars)-count:]
	}
	writeJSON(w, map[string]interface{}{
		"code":  code,
		"count": len(bars),
		"bars":  bars,
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/ws/realtime/")
	if symbol == "" {
		writeError(w, 400, "symbol required in path: /ws/realtime/SZ000001")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wsc := &wsConn{conn: conn, symbol: symbol, stop: make(chan struct{})}
	s.wsHub.add(wsc)
	defer s.wsHub.remove(wsc)

	// Poll real-time quotes every 3 seconds using eastmoney API
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	hc := &http.Client{Timeout: 5 * time.Second}

	for {
		select {
		case <-wsc.stop:
			return
		case <-ticker.C:
			// Parse symbol: SZ000001 or SH600000
			market := 0
			code := symbol
			if strings.HasPrefix(symbol, "SH") {
				market = 1
				code = strings.TrimPrefix(symbol, "SH")
			} else if strings.HasPrefix(symbol, "SZ") {
				market = 0
				code = strings.TrimPrefix(symbol, "SZ")
			}
			setcodeStr := fmt.Sprintf("%d.%s", market, code)
			url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f168,f169,f170,f171", setcodeStr)
			respHTTP, err := hc.Get(url)
			if err != nil {
				continue
			}
			var data interface{}
			json.NewDecoder(respHTTP.Body).Decode(&data)
			respHTTP.Body.Close()
			conn.WriteJSON(data)
		}
	}
}

func (s *Server) handleBoardList(w http.ResponseWriter, r *http.Request) {
	boardType := queryParam(r, "board_type", "HY")
	count := queryInt(r, "count", 50)
	// Use EastMoneyEnhanced scraper (TDX PBBoardList returns 503)
	es := scraper.NewEastMoneyScraper()
	boards, err := es.SectorBoards(boardType)
	if err == nil && len(boards) > 0 {
		if len(boards) > count {
			boards = boards[:count]
		}
		writeJSON(w, map[string]interface{}{"board_type": boardType, "count": len(boards), "data": boards})
		return
	}
	// Fallback to old sector scraper
	ss := scraper.NewSectorScraper()
	oldBoards, err2 := ss.FetchSectorBoards(boardType)
	if err2 == nil && len(oldBoards) > 0 {
		if len(oldBoards) > count {
			oldBoards = oldBoards[:count]
		}
		writeJSON(w, map[string]interface{}{"board_type": boardType, "count": len(oldBoards), "data": oldBoards})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取板块列表失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleBoardMembers(w http.ResponseWriter, r *http.Request) {
	boardSymbol := r.URL.Query().Get("board_symbol")
	count := queryInt(r, "count", 50)
	if boardSymbol == "" {
		writeError(w, 400, "board_symbol 参数必填")
		return
	}
	// Use EastMoneyEnhanced scraper (TDX PBBoardMembers returns 503)
	es := scraper.NewEastMoneyScraper()
	stocks, err := es.SectorStocks(boardSymbol)
	if err == nil {
		if len(stocks) > count {
			stocks = stocks[:count]
		}
		writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "count": len(stocks), "data": stocks})
		return
	}
	// Fallback to old sector scraper
	ss := scraper.NewSectorScraper()
	oldStocks, err2 := ss.FetchBoardStocks(boardSymbol)
	if err2 == nil {
		if len(oldStocks) > count {
			oldStocks = oldStocks[:count]
		}
		writeJSON(w, map[string]interface{}{"board_symbol": boardSymbol, "count": len(oldStocks), "data": oldStocks})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取板块成分股失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleBoardRanking(w http.ResponseWriter, r *http.Request) {
	boardType := queryParam(r, "board_type", "HY")
	topN := queryInt(r, "top_n", 10)
	sortBy := queryParam(r, "sort_by", "change_pct")
	// Use EastMoneyEnhanced scraper (TDX PBBoardRanking returns 503)
	es := scraper.NewEastMoneyScraper()
	boards, err := es.SectorBoards(boardType)
	if err == nil && len(boards) > 0 {
		if sortBy == "change_pct" {
			// Already sorted by scraper (descending by change)
		}
		if len(boards) > topN {
			boards = boards[:topN]
		}
		writeJSON(w, map[string]interface{}{"board_type": boardType, "top_n": topN, "sort_by": sortBy, "data": boards})
		return
	}
	// Fallback to old sector scraper
	ss := scraper.NewSectorScraper()
	oldBoards, err2 := ss.FetchSectorBoards(boardType)
	if err2 == nil && len(oldBoards) > 0 {
		if len(oldBoards) > topN {
			oldBoards = oldBoards[:topN]
		}
		writeJSON(w, map[string]interface{}{"board_type": boardType, "top_n": topN, "sort_by": sortBy, "data": oldBoards})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取板块排名失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleCapitalFlow(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	// Use old northbound scraper (EastMoneyEnhanced NorthBoundDaily uses datacenter API that returns empty)
	ns := scraper.NewNorthboundScraper()
	flow, err := ns.GetDailyFlow(5)
	if err == nil {
		writeJSON(w, map[string]interface{}{"code": code, "market": market, "count": len(flow), "data": flow})
		return
	}
	// Fallback to EastMoneyEnhanced
	es := scraper.NewEastMoneyScraper()
	newFlow, err2 := es.NorthBoundDaily("")
	if err2 == nil {
		writeJSON(w, map[string]interface{}{"code": code, "market": market, "count": len(newFlow), "data": newFlow})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取资金流向失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleAuction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	// Use eastmoney push2 API for auction data (TDX PBAuction may not be available)
	hc := &http.Client{Timeout: 10 * time.Second}
	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f19,f20,f21,f22,f23,f24,f25,f32,f33,f34,f35,f36,f37,f38,f39,f40,f41,f42,f43,f44,f45,f46,f47,f48,f49", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取数据失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleUnusual(w http.ResponseWriter, r *http.Request) {
	market := queryInt(r, "market", -1)
	count := queryInt(r, "count", 50)
	// Use EastMoneyEnhanced scraper for hot stocks (TDX PBUnusual returns 503)
	es := scraper.NewEastMoneyScraper()
	hot, err := es.HotRank(50)
	if err == nil && len(hot) > 0 {
		if len(hot) > count {
			hot = hot[:count]
		}
		writeJSON(w, map[string]interface{}{"market": market, "count": len(hot), "data": hot})
		return
	}
	// Fallback to old sector scraper
	ss := scraper.NewSectorScraper()
	boards, err2 := ss.FetchSectorBoards("HY")
	if err2 == nil && len(boards) > 0 {
		if len(boards) > count {
			boards = boards[:count]
		}
		writeJSON(w, map[string]interface{}{"market": market, "count": len(boards), "data": boards})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取异动数据失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleMarketStat(w http.ResponseWriter, r *http.Request) {
	// Use eastmoney push2 clist API for market statistics (TDX PBMarketStat returns 503)
	hc := &http.Client{Timeout: 10 * time.Second}
	// Query total A-share count — the API returns total in data.total
	url := "https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=1&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81&fields=f3"
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取市场统计失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	// TDX PBServerInfo returns 503 — return static server info
	writeJSON(w, map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
		"data_sources": []string{"TDX TCP", "EastMoney API", "Sina Finance"},
		"status":      "running",
	})
}

func (s *Server) handleSymbolInfo(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	// Use eastmoney push2 API for stock info
	hc := &http.Client{Timeout: 10 * time.Second}
	setcode := ""
	if market == 0 {
		setcode = "0." + code
	} else if market == 1 {
		setcode = "1." + code
	} else {
		setcode = "0." + code
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f168,f169,f170,f171", setcode)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取证券信息失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleQuoteList(w http.ResponseWriter, r *http.Request) {
	count := queryInt(r, "count", 20)
	sortType := queryParam(r, "sort_type", "CHANGE_PCT")
	// Use eastmoney push2 API for quote list
	hc := &http.Client{Timeout: 10 * time.Second}
	sortField := "f151" // default: change pct
	switch sortType {
	case "VOLUME_RATIO":
		sortField = "f160"
	case "AMPLITUDE":
		sortField = "f147"
	case "TURNOVER":
		sortField = "f163"
	}
	order := "desc"
	if sortType == "CHANGE_PCT" {
		order = "desc"
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=%d&po=1&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f12,f14,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f14,f15,f16,f17,f18,f20,f21,f22,f23,f24,f25,f26,f30,f31,f32,f33,f34,f35,f36,f37,f38,f39,f40,f41,f45,f46,f48,f50,f51,f52,f55,f57,f58,f60,f71&sort=%s&order=%s", count, sortField, order)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取行情列表失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleSecurityCount(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	if market == "" {
		market = "SZ"
	}
	// Use clist API with A-share filter to get total count
	hc := &http.Client{Timeout: 10 * time.Second}
	var fs string
	if strings.ToUpper(market) == "SH" {
		fs = "m:1+t:2,m:1+t:23"
	} else {
		fs = "m:0+t:6,m:0+t:80"
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?pn=1&pz=1&po=1&np=1&fltt=2&invt=2&fid=f3&fs=%s&fields=f3", fs)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取证券数量失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleBelongBoard(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market, ok := parseMarket(r)
	if code == "" || !ok {
		writeError(w, 400, "code 和 market 参数必填")
		return
	}
	// Use eastmoney push2 API for belong boards
	hc := &http.Client{Timeout: 10 * time.Second}
	setcode := ""
	if market == 0 {
		setcode = "0." + code
	} else if market == 1 {
		setcode = "1." + code
	} else {
		setcode = "0." + code
	}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/slist/get?pn=1&pz=50&po=1&np=1&fltt=2&invt=2&fs=%s&fields=f12,f14,f23", setcode)
	respHTTP, err := hc.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取所属板块失败: %v", err))
		return
	}
	defer respHTTP.Body.Close()
	var result interface{}
	json.NewDecoder(respHTTP.Body).Decode(&result)
	writeJSON(w, result)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeError(w, 400, "filename 参数必填")
		return
	}
	// TDX PBBlock not available — use EastMoney sector boards as replacement
	// Map filename to board type: gy.dat=industry, gn.dat=concept, dy.dat=region
	boardType := "industry"
	switch strings.ToLower(filename) {
	case "gn.dat", "block_gn.dat":
		boardType = "concept"
	case "dy.dat", "block_dy.dat":
		boardType = "region"
	case "zs.dat", "block_zs.dat":
		boardType = "index"
	case "zc.dat", "block_zc.dat":
		boardType = "policy"
	}
	// Use EastMoney scraper for sector boards
	es := scraper.NewEastMoneyScraper()
	boards, err := es.SectorBoards(boardType)
	if err == nil {
		writeJSON(w, map[string]interface{}{"filename": filename, "board_type": boardType, "count": len(boards), "data": boards})
		return
	}
	// Fallback to old sector scraper
	ss := scraper.NewSectorScraper()
	oldBoards, err2 := ss.FetchSectorBoards(boardType)
	if err2 == nil {
		writeJSON(w, map[string]interface{}{"filename": filename, "board_type": boardType, "count": len(oldBoards), "data": oldBoards})
		return
	}
	writeError(w, 500, fmt.Sprintf("获取板块文件失败: %v, fallback: %v", err, err2))
}

func (s *Server) handleMarketOverview(w http.ResponseWriter, r *http.Request) {
	// Use eastmoney APIs for market overview (TDX PBMarketStat returns 503)
	hc := &http.Client{Timeout: 10 * time.Second}

	// Fetch market stats from eastmoney upndown API
	statURL := "https://push2delay.eastmoney.com/api/qt/upndown/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60&ut=b2884a393a59ad64002292a3e90d46a5"
	respHTTP, err := hc.Get(statURL)
	var statData interface{}
	if err == nil {
		defer respHTTP.Body.Close()
		json.NewDecoder(respHTTP.Body).Decode(&statData)
	}

	// Fetch top boards from EastMoneyEnhanced scraper
	var boardData interface{}
	es := scraper.NewEastMoneyScraper()
	boards, err := es.SectorBoards("industry")
	if err == nil && len(boards) > 0 {
		boardData = boards
	}

	writeJSON(w, map[string]interface{}{
		"market_stat": statData,
		"board_data":  boardData,
	})
}

func (s *Server) handleMacroData(w http.ResponseWriter, r *http.Request) {
	indicator := queryParam(r, "indicator", "CPI")
	count := queryInt(r, "count", 12)

	// Use MacroScraper for supported indicators
	ms := scraper.NewMacroScraper("")
	switch indicator {
	case "CPI":
		data, err := ms.GetCPI(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取CPI失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	case "GDP":
		data, err := ms.GetGDP(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取GDP失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	case "PMI":
		data, err := ms.GetPMI(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取PMI失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	case "M2":
		data, err := ms.GetMoneySupply(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取M2失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	case "LPR":
		data, err := ms.GetLPRDirect(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取LPR失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	case "SHIBOR":
		data, err := ms.GetShibor(count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取SHIBOR失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"indicator": indicator, "count": len(data), "data": data})
	default:
		// Fallback: raw HTTP request to eastmoney datacenter
		macroURLs := map[string]string{
			"CPI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_CPI&columns=TRADE_DATE,NATIONAL_SAME,NATIONAL_BASE&pageSize=%d&pageNumber=1",
			"PMI": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_PMI&columns=TRADE_DATE,MAKE_INDEX,NONMANU_INDEX&pageSize=%d&pageNumber=1",
			"GDP": "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ECONOMY_GDP&columns=TRADE_DATE,GDP,CUM_GDP,CUM_GDP_SAME&pageSize=%d&pageNumber=1",
		}
		urlFmt, ok := macroURLs[indicator]
		if !ok {
			writeError(w, 400, "indicator 必须为: CPI, PMI, GDP, M2, LPR, SHIBOR")
			return
		}
		resp, err := http.Get(fmt.Sprintf(urlFmt, count))
		if err != nil {
			writeError(w, 500, fmt.Sprintf("请求失败: %v", err))
			return
		}
		defer resp.Body.Close()
		var data interface{}
		json.NewDecoder(resp.Body).Decode(&data)
		writeJSON(w, map[string]interface{}{"indicator": indicator, "data": data})
	}
}

func (s *Server) handleNewsSentiment(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	count := queryInt(r, "count", 10)
	url := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/ann?page_size=%d&page_index=1&stock_list=%s", count, code)
	resp, err := http.Get(url)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("请求失败: %v", err))
		return
	}
	defer resp.Body.Close()
	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	writeJSON(w, map[string]interface{}{"code": code, "count": count, "data": data})
}

func (s *Server) handleScraper(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		writeError(w, 400, "query 参数必填")
		return
	}
	source := queryParam(r, "source", "all")
	scrpr, err := scraper.NewScraper(30 * time.Second)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("初始化爬虫失败: %v", err))
		return
	}
	sources := []string{}
	switch source {
	case "iwcy":
		sources = []string{"iwcy"}
	case "xiaoda":
		sources = []string{"xiaoda"}
	case "eastmoney":
		sources = []string{"eastmoney"}
	default:
		sources = []string{"iwcy", "xiaoda", "eastmoney"}
	}
	result := scrpr.ScrapeAll(sources, query)
	writeJSON(w, result)
}

// Scraper endpoints for new modules
func (s *Server) handleScraperSectorBoards(w http.ResponseWriter, r *http.Request) {
	boardType := queryParam(r, "board_type", "HY")
	ss := scraper.NewSectorScraper()
	boards, err := ss.FetchSectorBoards(boardType)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取板块失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{"board_type": boardType, "count": len(boards), "data": boards})
}

func (s *Server) handleScraperNorthboundFlow(w http.ResponseWriter, r *http.Request) {
	ns := scraper.NewNorthboundScraper()
	days := queryInt(r, "days", 5)
	flow, err := ns.GetDailyFlow(days)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取北向资金流向失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{"count": len(flow), "data": flow})
}

func (s *Server) handleScraperNorthboundStocks(w http.ResponseWriter, r *http.Request) {
	ns := scraper.NewNorthboundScraper()
	count := queryInt(r, "count", 10)
	market := queryParam(r, "market", "all")
	switch market {
	case "sh":
		stocks, err := ns.GetTopShanghaiNorthbound("", count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取沪股通持仓失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"market": "SH", "count": len(stocks), "data": stocks})
	case "sz":
		stocks, err := ns.GetTopShenzhenNorthbound("", count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取深股通持仓失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"market": "SZ", "count": len(stocks), "data": stocks})
	default:
		stocks, err := ns.GetTopNorthboundStocks("", count)
		if err != nil {
			writeError(w, 500, fmt.Sprintf("获取北向资金持仓失败: %v", err))
			return
		}
		writeJSON(w, map[string]interface{}{"count": len(stocks), "data": stocks})
	}
}

func (s *Server) handleScraperNorthboundHolders(w http.ResponseWriter, r *http.Request) {
	ns := scraper.NewNorthboundScraper()
	count := queryInt(r, "count", 10)
	mutualType := queryParam(r, "mutual_type", "")
	holders, err := ns.GetNorthboundHolders(mutualType, count)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取北向资金机构持仓排名失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{"count": len(holders), "data": holders})
}

func (s *Server) handleScraperFundNav(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	fc := scraper.NewFundNavClient()
	nav, err := fc.GetLatestNAV(code)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取基金净值失败: %v", err))
		return
	}
	writeJSON(w, nav)
}

func (s *Server) handleScraperMarginTrade(w http.ResponseWriter, r *http.Request) {
	mc := scraper.NewMarginTradeWebClient()
	data, err := mc.GetSummary()
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取融资融券数据失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{"count": len(data), "data": data})
}

func (s *Server) handleScraperFundHolding(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		code = r.URL.Query().Get("fund_code")
	}
	period := r.URL.Query().Get("period")
	if code == "" {
		writeError(w, 400, "code 参数必填（支持 code 或 fund_code）")
		return
	}
	fh := scraper.NewFundHoldingClient()
	report, err := fh.GetHoldingReport(code, period)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取基金持仓失败: %v", err))
		return
	}
	writeJSON(w, report)
}

func (s *Server) handleScraperFundSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	pageSize := queryInt(r, "page_size", 10)
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	fh := scraper.NewFundHoldingClient()
	funds, err := fh.SearchFunds(keyword, pageSize)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("基金搜索失败: %v", err))
		return
	}
	writeJSON(w, map[string]interface{}{"count": len(funds), "data": funds})
}

func (s *Server) handleScraperHKUSQuote(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	market := queryParam(r, "market", "hk")
	if code == "" {
		writeError(w, 400, "code 参数必填")
		return
	}
	hc := scraper.NewHKUSFinancialClient()
	info, err := hc.GetStockQuote(code, market)
	if err != nil {
		writeError(w, 500, fmt.Sprintf("获取报价失败: %v", err))
		return
	}
	writeJSON(w, info)
}

func (s *Server) handleScraperCrypto(w http.ResponseWriter, r *http.Request) {
	symbolsParam := r.URL.Query().Get("symbols")
	if symbolsParam == "" {
		symbolsParam = "bitcoin,ethereum"
	}
	symbols := strings.Split(symbolsParam, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}
	data := scraper.GetCachedCryptoData(symbols)
	writeJSON(w, map[string]interface{}{"count": len(data), "data": data, "source": "cached"})
}

// Helper functions for parsing TQLEX response data.

func parseKlineBars(data interface{}) ([]indicator.Bar, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Try array of objects (TCP format)
	type klineObj struct {
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Vol    float64 `json:"vol"`
		Amount float64 `json:"amount"`
	}
	var objs []klineObj
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		bars := make([]indicator.Bar, len(objs))
		for i, o := range objs {
			bars[i] = indicator.Bar{Open: o.Open, Close: o.Close, High: o.High, Low: o.Low, Vol: o.Vol, Amount: o.Amount}
		}
		return bars, nil
	}

	// Try eastmoney klines string array format: {"data":{"klines":["date,open,close,...",...]}}
	type emKlinesWrapper struct {
		Data struct{ Klines []string `json:"klines"` } `json:"data"`
	}
	var emk emKlinesWrapper
	if err := json.Unmarshal(raw, &emk); err == nil && len(emk.Data.Klines) > 0 {
		return parseEastMoneyKlines(emk.Data.Klines)
	}

	// Try array of arrays (HTTP TQLEX ListItem.Item format)
	var arrs [][]float64
	if err := json.Unmarshal(raw, &arrs); err == nil && len(arrs) > 0 {
		bars := make([]indicator.Bar, len(arrs))
		for i, row := range arrs {
			if len(row) >= 6 {
				bars[i] = indicator.Bar{Open: row[0], Close: row[1], High: row[2], Low: row[3], Vol: row[4], Amount: row[5]}
			}
		}
		if len(bars) > 0 {
			return bars, nil
		}
	}
	// Try ListItem format
	type listItemWrapper struct {
		ListItem []struct{ Item []interface{} } `json:"ListItem"`
	}
	var lw listItemWrapper
	if err := json.Unmarshal(raw, &lw); err == nil && len(lw.ListItem) > 0 {
		return extractBarsFromListItem(lw.ListItem)
	}
	// Try nested Data.ListItem
	type nestedWrapper struct {
		Data struct{ ListItem []struct{ Item []interface{} } } `json:"Data"`
	}
	var nw nestedWrapper
	if err := json.Unmarshal(raw, &nw); err == nil && len(nw.Data.ListItem) > 0 {
		return extractBarsFromListItem(nw.Data.ListItem)
	}
	return nil, fmt.Errorf("unsupported kline data format")
}

func parseEastMoneyKlines(klines []string) ([]indicator.Bar, error) {
	var bars []indicator.Bar
	for _, kl := range klines {
		parts := strings.Split(kl, ",")
		if len(parts) < 7 {
			continue
		}
		b := indicator.Bar{
			Open:   toFloat64(parts[1]),
			Close:  toFloat64(parts[2]),
			High:   toFloat64(parts[3]),
			Low:    toFloat64(parts[4]),
			Vol:    toFloat64(parts[5]),
			Amount: toFloat64(parts[6]),
		}
		bars = append(bars, b)
	}
	if len(bars) > 0 {
		return bars, nil
	}
	return nil, fmt.Errorf("no valid kline data parsed")
}

func extractBarsFromListItem(listItem []struct{ Item []interface{} }) ([]indicator.Bar, error) {
	var bars []indicator.Bar
	for _, li := range listItem {
		if len(li.Item) < 6 {
			continue
		}
		b := indicator.Bar{
			Open: toFloat64(li.Item[2]), Close: toFloat64(li.Item[3]),
			High: toFloat64(li.Item[4]), Low: toFloat64(li.Item[5]),
			Vol: toFloat64(li.Item[6]), Amount: toFloat64(li.Item[7]),
		}
		bars = append(bars, b)
	}
	if len(bars) > 0 {
		return bars, nil
	}
	return nil, fmt.Errorf("unsupported kline data format")
}

func parseChanlunKlines(data interface{}) ([]chanlun.Kline, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Try eastmoney klines string array format first
	type emKlinesWrapper struct {
		Data struct{ Klines []string } `json:"data"`
	}
	var emk emKlinesWrapper
	if err := json.Unmarshal(raw, &emk); err == nil && len(emk.Data.Klines) > 0 {
		return parseEastMoneyChanlunKlines(emk.Data.Klines)
	}

	// Try array of objects (TCP format)
	type klineObj struct {
		Date   string  `json:"date"`
		Time   string  `json:"time"`
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Vol    float64 `json:"vol"`
		Amount float64 `json:"amount"`
	}
	var objs []klineObj
	if err := json.Unmarshal(raw, &objs); err == nil && len(objs) > 0 {
		klines := make([]chanlun.Kline, len(objs))
		for i, o := range objs {
			dt := o.Date
			if dt == "" {
				dt = o.Time
			}
			klines[i] = chanlun.Kline{Date: dt, Open: o.Open, Close: o.Close, High: o.High, Low: o.Low, Vol: o.Vol, Amount: o.Amount}
		}
		return klines, nil
	}

	// Try array of arrays (HTTP TQLEX ListItem.Item format)
	var arrItems [][]interface{}
	if err := json.Unmarshal(raw, &arrItems); err == nil && len(arrItems) > 0 {
		klines := make([]chanlun.Kline, len(arrItems))
		for i, item := range arrItems {
			if len(item) < 8 {
				continue
			}
			date := fmt.Sprintf("%v", item[0])
			klines[i] = chanlun.Kline{Date: date, Open: toFloat64(item[2]), Close: toFloat64(item[3]), High: toFloat64(item[4]), Low: toFloat64(item[5]), Vol: toFloat64(item[6]), Amount: toFloat64(item[7])}
		}
		if len(klines) > 0 {
			return klines, nil
		}
	}

	// Try map with ListItem (HTTP TQLEX response)
	type listItemWrapper struct {
		ListItem []struct{ Item []interface{} } `json:"ListItem"`
	}
	var lw listItemWrapper
	if err := json.Unmarshal(raw, &lw); err == nil && len(lw.ListItem) > 0 {
		return extractKlinesFromListItem(lw.ListItem)
	}

	// Try nested: {"Data": {"ListItem": [...]}}
	type nestedWrapper struct {
		Data struct{ ListItem []struct{ Item []interface{} } } `json:"Data"`
	}
	var nw nestedWrapper
	if err := json.Unmarshal(raw, &nw); err == nil && len(nw.Data.ListItem) > 0 {
		return extractKlinesFromListItem(nw.Data.ListItem)
	}

	return nil, fmt.Errorf("unsupported kline data format")
}

func extractKlinesFromListItem(listItem []struct{ Item []interface{} }) ([]chanlun.Kline, error) {
	var allKlines []chanlun.Kline
	for _, li := range listItem {
		if len(li.Item) < 8 {
			continue
		}
		date := fmt.Sprintf("%v", li.Item[0])
		k := chanlun.Kline{Date: date, Open: toFloat64(li.Item[2]), Close: toFloat64(li.Item[3]), High: toFloat64(li.Item[4]), Low: toFloat64(li.Item[5]), Vol: toFloat64(li.Item[6]), Amount: toFloat64(li.Item[7])}
		allKlines = append(allKlines, k)
	}
	if len(allKlines) > 0 {
		return allKlines, nil
	}
	return nil, fmt.Errorf("unsupported kline data format")
}

func parseEastMoneyChanlunKlines(klines []string) ([]chanlun.Kline, error) {
	var allKlines []chanlun.Kline
	for _, kl := range klines {
		parts := strings.Split(kl, ",")
		if len(parts) < 7 {
			continue
		}
		k := chanlun.Kline{
			Date:   parts[0],
			Open:   toFloat64(parts[1]),
			Close:  toFloat64(parts[2]),
			High:   toFloat64(parts[3]),
			Low:    toFloat64(parts[4]),
			Vol:    toFloat64(parts[5]),
			Amount: toFloat64(parts[6]),
		}
		allKlines = append(allKlines, k)
	}
	if len(allKlines) > 0 {
		return allKlines, nil
	}
	return nil, fmt.Errorf("no valid kline data parsed")
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case nil:
		return 0
	default:
		return 0
	}
}

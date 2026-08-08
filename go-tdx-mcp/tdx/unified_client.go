package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bensema/gotdx/proto"
	"github.com/bensema/gotdx/types"
	"github.com/tdx/go-tdx-mcp/backtest"
	"github.com/tdx/go-tdx-mcp/chanlun"
	"github.com/tdx/go-tdx-mcp/factor"
	"github.com/tdx/go-tdx-mcp/indicator"
	"github.com/tdx/go-tdx-mcp/scraper"
)

// normalizeBody converts a typed request body (struct) into map[string]interface{}.
// This fixes the mismatch between handlers passing typed structs (QuoteRequest,
// KlineRequest, etc.) and query* methods asserting map[string]interface{}.
func normalizeBody(body interface{}) (map[string]interface{}, bool) {
	if m, ok := body.(map[string]interface{}); ok {
		return m, true
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}

// toInt converts a numeric value read from a normalized body map to int.
// After json.Marshal/Unmarshal round-trip, numeric fields become float64.
func toInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	case string:
		var n int
		fmt.Sscanf(t, "%d", &n)
		return n
	default:
		return 0
	}
}

// UnifiedClient wraps both HTTP (TQLEX) and TCP clients with automatic fallback.
type UnifiedClient struct {
	httpClient        *HTTPClient
	tcpClient         *TDXTCPClient
	collector         *MultiHostCollector
	sectorScraper     *scraper.SectorScraper
	macroScraper      *scraper.MacroScraper
	northScraper      *scraper.NorthboundScraper
	fundNavClient     *scraper.FundNavClient
	marginTradeClient *scraper.MarginTradeWebClient
	eastMoneyScraper  *scraper.EastMoneyScraper
	blockTradeClient  *scraper.BlockTradeClient
	fundHoldingClient *scraper.FundHoldingClient
	sinaClient        *scraper.SinaClient
	tableParser       *scraper.TableParser
	ocrClient         *scraper.OCRClient
	webScraper        *scraper.Scraper
	backtestEngine    *backtest.Engine
	initOnce          sync.Once
	initErr           error
	useCollector      bool
	useScraper        bool
	useMacroScraper   bool
	useNorthScraper   bool
	scrapersInitOnce  sync.Once
}

// NewUnifiedClient creates a unified client with both HTTP and TCP backends.
// HTTP client is only created when token is non-empty; when absent, all data
// flows through TCP (auto-select TDX servers) and web scrapers.
func NewUnifiedClient(token string, timeoutSec int, tdxHost string, tdxPort int, opts ...UnifiedClientOption) *UnifiedClient {
	uc := &UnifiedClient{
		tcpClient: NewTDXTCPClientWithHost(timeoutSec, tdxHost, tdxPort),
	}
	if token != "" {
		uc.httpClient = NewHTTPClient(token, 0)
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// UnifiedClientOption configures a UnifiedClient.
type UnifiedClientOption func(*UnifiedClient)

// WithMultiHostCollector enables MultiHostCollector as the primary TCP backend.
func WithMultiHostCollector(collector *MultiHostCollector) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.collector = collector
		uc.useCollector = true
	}
}

// WithSectorScraper enables web scraping for sector board data (industry/concept/region).
func WithSectorScraper(sectorScraper *scraper.SectorScraper) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.sectorScraper = sectorScraper
		uc.useScraper = true
	}
}

// CollectorEnabled returns true if MultiHostCollector is configured.
func (uc *UnifiedClient) CollectorEnabled() bool {
	return uc.collector != nil
}

// ScraperEnabled returns true if SectorScraper is configured.
func (uc *UnifiedClient) ScraperEnabled() bool {
	return uc.sectorScraper != nil
}

// WithMacroScraper enables macro data scraping from EastMoney.
func WithMacroScraper(macroScraper *scraper.MacroScraper) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.macroScraper = macroScraper
		uc.useMacroScraper = true
	}
}

// MacroScraperEnabled returns true if MacroScraper is configured.
func (uc *UnifiedClient) MacroScraperEnabled() bool {
	return uc.macroScraper != nil
}

// WithNorthboundScraper enables northbound capital flow scraping.
func WithNorthboundScraper(northScraper *scraper.NorthboundScraper) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.northScraper = northScraper
		uc.useNorthScraper = true
	}
}

// NorthboundScraperEnabled returns true if NorthboundScraper is configured.
func (uc *UnifiedClient) NorthboundScraperEnabled() bool {
	return uc.northScraper != nil
}

// WithFundNavClient enables fund NAV web scraping via goquery.
func WithFundNavClient(fundNavClient *scraper.FundNavClient) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.fundNavClient = fundNavClient
	}
}

// WithMarginTradeClient enables margin trade data via eastmoney datacenter API.
func WithMarginTradeClient(marginTradeClient *scraper.MarginTradeWebClient) UnifiedClientOption {
	return func(uc *UnifiedClient) {
		uc.marginTradeClient = marginTradeClient
	}
}

// initTCP lazily initializes TCP connections.
func (uc *UnifiedClient) initTCP(ctx context.Context) error {
	// Only connect main A-share server. ConnectEx/ConnectMAC for HK/US/futures
	// are not needed for A-share queries and cause a gotdx race condition when
	// initialized concurrently with ConnectMain (see gotdx StockKLineOffset panic
	// with "slice bounds out of range" after parallel init).
	uc.initOnce.Do(func() {
		if err := uc.tcpClient.ConnectMain(ctx); err != nil {
			uc.initErr = fmt.Errorf("tcp init failed: %w", err)
		}
	})
	if uc.initErr != nil {
		return uc.initErr
	}
	return nil
}

// hasHTTP returns true when an HTTP TQLEX client is configured (token provided).
func (uc *UnifiedClient) hasHTTP() bool {
	return uc.httpClient != nil
}

// TQLEXQuery implements the Client interface with TCP-first fallback to TQLEX.
func (uc *UnifiedClient) TQLEXQuery(ctx context.Context, entry string, body interface{}) (*TQLEXResponse, error) {
	// Map TQLEX entries to TCP methods
	switch entry {
	case "TdxShare.PBHQInfo":
		return uc.queryQuotes(body)
	case "TdxShare.PBFXT":
		return uc.queryKline(body)
	case "TdxShare.wendaQuery":
		return uc.queryScreener(body)
	case "TdxShare.InfoSelectV2":
		return uc.queryIndicator(body)
	case "TdxShare.PBAuction":
		return uc.queryAuction(body)
	case "TdxShare.TdxSharePCCW":
		return uc.queryF10(body)
	case "TdxShare.PBCapitalFlow":
		return uc.queryCapitalFlow(body)
	case "TdxShare.PBBoardList":
		return uc.queryBoardList(body)
	case "TdxShare.PBBoardMembers":
		return uc.queryBoardMembers(body)
	case "TdxShare.PBBoardRanking":
		return uc.queryBoardRanking(body)
	case "TdxShare.PBServerInfo":
		return uc.queryServerInfo(body)
	case "TdxShare.PBSymbolInfo":
		return uc.querySymbolInfo(body)
	case "TdxShare.PBBelongBoard":
		return uc.queryBelongBoard(body)
	case "TdxShare.PBUnusual":
		return uc.queryUnusual(body)
	case "TdxShare.PBMarketStat":
		return uc.queryMarketStat(body)
	case "TdxShare.PBSecurityList":
		return uc.querySecurityList(body)
	case "TdxShare.PBGetFinanceInfo":
		return uc.queryFinanceInfo(body)
	case "TdxShare.PBQuoteList":
		return uc.queryQuoteList(body)
	case "TdxShare.PBFSTick":
		return uc.queryFSTick(body)
	case "TdxShare.PBTrans":
		return uc.queryTrans(body)
	}

	// For unimplemented entries, try TCP first, then HTTP if available
	if err := uc.initTCP(ctx); err == nil && uc.tcpClient.IsConnected() {
		// TCP connected but entry not implemented via TCP — try HTTP if available
		if !uc.hasHTTP() {
			return nil, fmt.Errorf("entry %s not supported via TCP and no HTTP client (missing TDX_TOKEN)", entry)
		}
		return uc.httpClient.TQLEXQuery(ctx, entry, body)
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(ctx, entry, body)
	}
	return nil, fmt.Errorf("entry %s not supported via TCP and no HTTP client configured", entry)
}

// queryQuotes handles PBHQInfo via TCP or MultiHostCollector.
func (uc *UnifiedClient) queryQuotes(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid quotes request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}

	if uc.useCollector && uc.collector != nil {
		quotes, _, err := uc.collector.CollectDetailedQuotes([]string{code}, market, 1)
		if err == nil && len(quotes) > 0 {
			result, _ := json.Marshal(&quotes[0])
			return &TQLEXResponse{Data: json.RawMessage(result)}, nil
		}
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPQuotes(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBHQInfo", body)
	}
	return nil, fmt.Errorf("quotes: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

// tryTCPQuotes attempts to get quotes via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPQuotes(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		quote, e := uc.tcpClient.GetQuote(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(quote)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryKline handles PBFXT via TCP or MultiHostCollector.
func (uc *UnifiedClient) queryKline(body interface{}) (*TQLEXResponse, error) {
	// Ensure TCP is initialized before querying
	_ = uc.initTCP(context.Background())
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in queryKline: %v\n", r)
		}
	}()
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid kline request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	period := PeriodCodeToString(data["Period"])
	count := 100
	if w := toInt(data["WantNum"]); w > 0 {
		count = w
	}
	fq := 0
	if t := toInt(data["TQFlag"]); t != 0 {
		if t&0x01 != 0 {
			fq = 1
		} else if t&0x02 != 0 {
			fq = 2
		}
	}

	if uc.useCollector && uc.collector != nil {
		periodCat := tcpPeriodToCategory(period)
		bars, _, err := uc.collector.CollectKLines([]string{code}, market, periodCat, uint16(count), uint16(fq))
		if err == nil {
			if bl, ok := bars[code]; ok && len(bl) > 0 {
				result, _ := json.Marshal(bl)
				return &TQLEXResponse{Data: json.RawMessage(result)}, nil
			}
		}
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPKline(code, market, period, count, fq)
		if result != nil {
			// Check if data is empty/unusable
			if raw, ok := result.Data.(json.RawMessage); ok && len(raw) > 0 {
				isEmpty := (raw[0] == '[' && len(raw) <= 2) || // "[]"
					(raw[0] == 'n' && len(raw) == 4)           // "null"
				if !isEmpty {
					return result, nil
				}
			}
		}
		_ = err
	}

	// Fall back to Sina HTTP API for Kline (no TDX_TOKEN needed)
	uc.initScrapers()
	return uc.queryKlineHTTP(code, market, period, count)
}

// queryKlineHTTP fetches K-line via Sina VIP HTTP API (no token required).
// push2his.eastmoney.com is not reachable from this environment, so Sina is used.
func (uc *UnifiedClient) queryKlineHTTP(code string, market int, period string, count int) (*TQLEXResponse, error) {
	if uc.sinaClient == nil {
		return nil, fmt.Errorf("kline: TCP returned empty and no HTTP client configured")
	}
	// Map period to Sina scale parameter
	scale := 240 // 日K
	switch period {
	case "5min":
		scale = 5
	case "15min":
		scale = 15
	case "30min":
		scale = 30
	case "60min":
		scale = 60
	case "week":
		scale = 1000
	case "month":
		scale = 5000
	case "day":
		scale = 240
	}
	// Build symbol: sz000001 (SZ) or sh000001 (SH)
	prefix := "sz"
	if market == 1 {
		prefix = "sh"
	}
	symbol := prefix + code
	klines, err := uc.sinaClient.GetKlineHistory(symbol, scale, count)
	if err != nil {
		return nil, fmt.Errorf("kline Sina fallback: %w", err)
	}
	if len(klines) == 0 {
		return nil, fmt.Errorf("kline: no data from Sina for %s", symbol)
	}
	result, _ := json.Marshal(klines)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// tryTCPKline attempts to get kline data via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPKline(code string, market int, period string, count, fq int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in GetKLineWithAdjust: %v", r)
		}
	}()
	err = uc.tcpClient.withRetry(func() error {
		bars, e := uc.tcpClient.GetKLineWithAdjust(code, market, period, count, fq)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(bars)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryScreener handles wendaQuery via TQLEX HTTP.
// wendaQuery is NLP-based, no direct TCP equivalent exists.
func (uc *UnifiedClient) queryScreener(body interface{}) (*TQLEXResponse, error) {
	if !uc.hasHTTP() {
		return nil, fmt.Errorf("screener requires HTTP client (missing TDX_TOKEN); TCP auto-select can be used as alternative")
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "JNLPSE:wendaQuery", body)
}

// queryIndicator handles InfoSelectV2 via TQLEX HTTP.
// InfoSelectV2 is NLP-based, no direct TCP equivalent exists.
func (uc *UnifiedClient) queryIndicator(body interface{}) (*TQLEXResponse, error) {
	if !uc.hasHTTP() {
		return nil, fmt.Errorf("indicator select requires HTTP client (missing TDX_TOKEN); use formula engine or factor engine as alternative")
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "NLPSE:InfoSelectV2", body)
}

// queryAuction handles PBAuction via TCP.
func (uc *UnifiedClient) queryAuction(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid auction request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPAuction(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBAuction", body)
	}
	return nil, fmt.Errorf("auction: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

// tryTCPAuction attempts to get auction data via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPAuction(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		auctions, e := uc.tcpClient.GetAuction(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(auctions)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryF10 handles TdxSharePCCW via TCP.
func (uc *UnifiedClient) queryF10(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid F10 request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPF10(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.TdxSharePCCW", body)
	}
	return nil, fmt.Errorf("F10: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

// tryTCPF10 attempts to get F10 info via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPF10(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		info, e := uc.tcpClient.GetF10(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(info)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryCapitalFlow handles PBCapitalFlow via TCP.
func (uc *UnifiedClient) queryCapitalFlow(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid capital flow request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPCapitalFlow(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBCapitalFlow", body)
	}
	return nil, fmt.Errorf("capital flow: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPCapitalFlow(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		reply, e := uc.tcpClient.GetCapitalFlow(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(reply)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryBoardList handles PBBoardList via TCP.
func (uc *UnifiedClient) queryBoardList(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid board list request body type")
	}
	boardType := fmt.Sprintf("%v", data["BoardType"])
	count := 50
	if c := toInt(data["Count"]); c != 0 {
		count = c
	}
	var bt BlockType
	switch strings.ToUpper(boardType) {
	case "HY":
		bt = BlockIndustry
	case "GN":
		bt = BlockConcept
	case "DY":
		bt = BlockConcept
	default:
		bt = BlockIndustry
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPBoardList(bt, count)
		if err == nil && result != nil {
			return result, nil
		}
	}

	// Fallback: East Money scraper (no token required)
	uc.initScrapers()
	if uc.eastMoneyScraper != nil {
		boardTypeStr := "industry"
		if bt == BlockConcept {
			boardTypeStr = "concept"
		}
		boards, err := uc.eastMoneyScraper.SectorBoards(boardTypeStr)
		if err == nil && len(boards) > 0 {
			if count > 0 && len(boards) > count {
				boards = boards[:count]
			}
			result, _ := json.Marshal(boards)
			return &TQLEXResponse{Data: json.RawMessage(result)}, nil
		}
	}

	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardList", body)
	}
	return nil, fmt.Errorf("board list: TCP failed and no HTTP client and scraper unavailable")
}

func (uc *UnifiedClient) tryTCPBoardList(bt BlockType, count int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		boards, e := uc.tcpClient.GetSectorBoards(bt)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(boards)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryBoardMembers handles PBBoardMembers via TCP.
func (uc *UnifiedClient) queryBoardMembers(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid board members request body type")
	}
	code := fmt.Sprintf("%v", data["BoardCode"])
	count := 50
	if c := toInt(data["Count"]); c != 0 {
		count = c
	}
	_ = count
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPBoardMembers(code)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardMembers", body)
	}
	return nil, fmt.Errorf("board members: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPBoardMembers(boardCode string) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		stocks, e := uc.tcpClient.GetSectorBoardStocks(boardCode)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(stocks)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryBoardRanking handles PBBoardRanking via TCP (falls back to scraper).
func (uc *UnifiedClient) queryBoardRanking(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid board ranking request body type")
	}
	boardType := fmt.Sprintf("%v", data["BoardType"])
	topN := 10
	if t := toInt(data["TopN"]); t != 0 {
		topN = t
	}
	sortBy := fmt.Sprintf("%v", data["SortBy"])
	_ = boardType
	_ = topN
	_ = sortBy
	// Board ranking not available via TCP — fall back to HTTP or scraper
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardRanking", body)
	}
	return nil, fmt.Errorf("board ranking not available without HTTP client (missing TDX_TOKEN)")
}

// queryServerInfo handles PBServerInfo via TCP.
func (uc *UnifiedClient) queryServerInfo(body interface{}) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPServerInfo()
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBServerInfo", body)
	}
	return nil, fmt.Errorf("server info: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPServerInfo() (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		info := uc.tcpClient.GetServerInfo()
		r, _ := json.Marshal(map[string]interface{}{"server_info": info, "status": "connected"})
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// querySymbolInfo handles PBSymbolInfo via MAC TCP.
func (uc *UnifiedClient) querySymbolInfo(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid symbol info request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPSymbolInfo(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBSymbolInfo", body)
	}
	return nil, fmt.Errorf("symbol info: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPSymbolInfo(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		reply, e := uc.tcpClient.GetSymbolInfo(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(reply)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryBelongBoard handles PBBelongBoard via TCP.
func (uc *UnifiedClient) queryBelongBoard(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid belong board request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	_ = code
	_ = market
	// Belong board not directly available via TCP — fall back to HTTP
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBelongBoard", body)
	}
	return nil, fmt.Errorf("belong board not available without HTTP client (missing TDX_TOKEN)")
}

// queryUnusual handles PBUnusual via TCP.
func (uc *UnifiedClient) queryUnusual(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid unusual request body type")
	}
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	count := 50
	if c := toInt(data["WantNum"]); c != 0 {
		count = c
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPUnusual(market, count)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBUnusual", body)
	}
	return nil, fmt.Errorf("unusual: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPUnusual(market, count int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		data, e := uc.tcpClient.GetUnusual(market, count)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(data)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryMarketStat handles PBMarketStat via TCP.
func (uc *UnifiedClient) queryMarketStat(body interface{}) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPMarketStat()
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBMarketStat", body)
	}
	return nil, fmt.Errorf("market stat: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPMarketStat() (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		count, e := uc.tcpClient.GetSecurityCount(0)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(map[string]interface{}{"sh_count": count, "sz_count": 0, "status": "ok"})
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// querySecurityList handles PBSecurityList via TCP.
func (uc *UnifiedClient) querySecurityList(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid security list request body type")
	}
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	start := uint16(0)
	if sv := toInt(data["Start"]); sv != 0 {
		start = uint16(sv)
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPSecurityList(market, start)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBSecurityList", body)
	}
	return nil, fmt.Errorf("security list: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPSecurityList(market int, start uint16) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		reply, e := uc.tcpClient.GetSecurityList(market, start)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(reply)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryFinanceInfo handles PBGetFinanceInfo via TCP.
func (uc *UnifiedClient) queryFinanceInfo(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid finance info request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if v := toInt(data["Setcode"]); v != 0 {
		market = v
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPFinanceInfo(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBGetFinanceInfo", body)
	}
	return nil, fmt.Errorf("finance info: TCP failed and no HTTP client (missing TDX_TOKEN)")
}

func (uc *UnifiedClient) tryTCPFinanceInfo(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		reply, e := uc.tcpClient.GetFinanceInfo(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(reply)
		// Extract only the stock_basic field which contains the finance info
		var data map[string]interface{}
		if r != nil {
			json.Unmarshal(r, &data)
		}
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// RAGQuery implements the Client interface.
func (uc *UnifiedClient) RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error) {
	if !uc.hasHTTP() {
		return nil, fmt.Errorf("RAG query requires HTTP client (missing TDX_TOKEN)")
	}
	return uc.httpClient.RAGQuery(ctx, query, topK)
}

// queryQuoteList handles PBQuoteList via EastMoney push2 clist (HTTP TQLEX returns 503).
func (uc *UnifiedClient) queryQuoteList(body interface{}) (*TQLEXResponse, error) {
	data, ok := normalizeBody(body)
	if !ok {
		return nil, fmt.Errorf("invalid quote list request body type")
	}
	count := 100
	if c := toInt(data["count"]); c != 0 {
		count = c
	}
	sortType := "f2"
	if s, ok := data["sort_type"].(string); ok {
		sortType = s
	}
	order := "desc"
	if o, ok := data["order"].(string); ok {
		order = o
	}
	pn := 1
	if p, ok := data["page"].(float64); ok {
		pn = int(p)
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	urlStr := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/clist/get?pn=%d&pz=%d&po=%s&np=1&fltt=2&invt=2&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f2,f3,f12,f14", pn, count, order)
	_ = sortType
	respHTTP, err := hc.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	bodyBytes, _ := io.ReadAll(respHTTP.Body)
	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("解析行情列表失败: %w", err)
	}
	encoded, _ := json.Marshal(result)
	return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
}

// queryFSTick handles PBFSTick via TCP (preferred) or EastMoney fallback.
func (uc *UnifiedClient) queryFSTick(body interface{}) (*TQLEXResponse, error) {
	code, market, err := extractCodeMarket(body)
	if err != nil {
		return nil, err
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPFSTick(code, market)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBFSTick", body)
	}

	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	bodyBytes, _ := io.ReadAll(respHTTP.Body)
	var result interface{}
	json.Unmarshal(bodyBytes, &result)
	encoded, _ := json.Marshal(result)
	return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
}

func (uc *UnifiedClient) tryTCPFSTick(code string, market int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		data, e := uc.tcpClient.GetTickChart(code, market)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(data)
		// Convert to array format expected by client
		var tickData interface{}
		if r != nil {
			json.Unmarshal(r, &tickData)
		}
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// queryTrans handles PBTrans via TCP (preferred) or EastMoney fallback.
func (uc *UnifiedClient) queryTrans(body interface{}) (*TQLEXResponse, error) {
	code, market, err := extractCodeMarket(body)
	if err != nil {
		return nil, err
	}
	count := 100
	if m, ok := normalizeBody(body); ok {
		if c := toInt(m["count"]); c != 0 {
			count = c
		}
	} else if tp, ok := body.(TransRequestParams); ok && tp.Count > 0 {
		count = tp.Count
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		result, err := uc.tryTCPTrans(code, market, count)
		if err == nil && result != nil {
			return result, nil
		}
	}
	if uc.hasHTTP() {
		return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBTrans", body)
	}

	setcodeStr := fmt.Sprintf("%d.%s", market, code)
	hc := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://push2delay.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71", setcodeStr)
	respHTTP, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	bodyBytes, _ := io.ReadAll(respHTTP.Body)
	var result interface{}
	json.Unmarshal(bodyBytes, &result)
	encoded, _ := json.Marshal(result)
	return &TQLEXResponse{Data: json.RawMessage(encoded)}, nil
}

func (uc *UnifiedClient) tryTCPTrans(code string, market, count int) (*TQLEXResponse, error) {
	var result *TQLEXResponse
	err := uc.tcpClient.withRetry(func() error {
		defer func() { recover() }()
		data, e := uc.tcpClient.GetTransaction(code, market, count)
		if e != nil {
			return e
		}
		r, _ := json.Marshal(data)
		result = &TQLEXResponse{Data: json.RawMessage(r)}
		return nil
	})
	return result, err
}

// extractCodeMarket extracts the Code and Setcode/market fields from a
// request body that may be a typed struct or a normalized map.
func extractCodeMarket(body interface{}) (string, int, error) {
	if m, ok := normalizeBody(body); ok {
		return fmt.Sprintf("%v", m["Code"]), toInt(m["Setcode"]), nil
	}
	switch b := body.(type) {
	case TickRequestParams:
		return b.Code, b.Market, nil
	case TransRequestParams:
		return b.Code, b.Market, nil
	default:
		return "", 0, fmt.Errorf("invalid request body type: %T", body)
	}
}

// Close releases all resources.
func (uc *UnifiedClient) Close() error {
	return uc.tcpClient.Disconnect()
}

// blockTypeToScraper maps BlockType to the SectorScraper filter type.
func blockTypeToScraper(bt BlockType) string {
	switch bt {
	case BlockIndustry:
		return "industry"
	case BlockConcept:
		return "concept"
	case BlockRegion:
		return "region"
	default:
		return ""
	}
}

// GetSectorBoards fetches sector boards. Uses web scraper for industry/concept/region,
// falls back to TCP for index/policy/custom and as a safety net.
func (uc *UnifiedClient) GetSectorBoards(bt BlockType) ([]SectorBoard, error) {
	scraperType := blockTypeToScraper(bt)

	// Try web scraper first for scrapable types
	if uc.useScraper && uc.sectorScraper != nil && scraperType != "" {
		webBoards, err := uc.sectorScraper.FetchSectorBoards(scraperType)
		if err == nil && len(webBoards) > 0 {
			boards := make([]SectorBoard, len(webBoards))
			for i, wb := range webBoards {
				boards[i] = SectorBoard{
					Code:     wb.Code,
					Name:     wb.Name,
					Type:     bt.BlockName(),
					StockCnt: 0,
				}
			}
			return boards, nil
		}
	}

	// Try MultiHostCollector if configured
	if uc.useCollector && uc.collector != nil {
		boards, _, err := uc.collector.CollectSectorBoards(bt)
		if err == nil {
			return boards, nil
		}
	}

	return uc.tcpClient.GetSectorBoards(bt)
}

// GetSectorBoardStocks fetches constituent stocks of a specific board.
// Uses web scraper for BK prefixed codes, falls back to TCP.
func (uc *UnifiedClient) GetSectorBoardStocks(boardCode string) ([]string, error) {
	// Try web scraper for BK-prefixed board codes
	if uc.useScraper && uc.sectorScraper != nil {
		stocks, err := uc.sectorScraper.FetchBoardStocks(boardCode)
		if err == nil && len(stocks) > 0 {
			return stocks, nil
		}
	}

	// Try MultiHostCollector
	if uc.useCollector && uc.collector != nil {
		boards, _, err := uc.collector.CollectSectorBoards(BlockIndustry)
		if err == nil {
			var matched []SectorBoard
			for _, b := range boards {
				if b.Code == boardCode || b.Name == boardCode {
					matched = append(matched, b)
				}
			}
			if len(matched) > 0 {
				stocks, _, err := uc.collector.CollectSectorBoardStocks(matched)
				if err == nil {
					if s, ok := stocks[matched[0].Code]; ok {
						return s, nil
					}
				}
			}
		}
	}

	return uc.tcpClient.GetSectorBoardStocks(boardCode)
}

// PeriodCodeToString converts TQLEX period code to TCP period string.
func PeriodCodeToString(v interface{}) string {
	switch val := v.(type) {
	case int:
		switch val {
		case 3:
			return "60min"
		case 4:
			return "day"
		case 5:
			return "week"
		case 6:
			return "month"
		case 9:
			return "1min"
		case 10:
			return "5min"
		case 11:
			return "15min"
		case 12:
			return "30min"
		case 13:
			return "quarter"
		case 14:
			return "year"
		default:
			return "day"
		}
	case float64:
		return PeriodCodeToString(int(val))
	case json.Number:
		n, _ := val.Int64()
		return PeriodCodeToString(int(n))
	default:
		return "day"
	}
}

// tcpPeriodToCategory converts a TCP period string to gotdx period category.
// Categories must match GotKLine's periodMap: 1min=1, 5min=2, ..., day=6, week=7, month=8.
func tcpPeriodToCategory(period string) uint16 {
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
	case "day":
		return 6
	case "week":
		return 7
	case "month":
		return 8
	case "quarter":
		return 9
	case "halfyear":
		return 10
	case "year":
		return 11
	default:
		return 6 // default to day
	}
}

// GetMacroCPI fetches China CPI data.
func (uc *UnifiedClient) GetMacroCPI(count int) ([]scraper.MacroIndicator, error) {
	if uc.useMacroScraper && uc.macroScraper != nil {
		return uc.macroScraper.GetCPI(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetMacroGDP fetches China GDP data.
func (uc *UnifiedClient) GetMacroGDP(count int) ([]scraper.MacroIndicator, error) {
	if uc.useMacroScraper && uc.macroScraper != nil {
		return uc.macroScraper.GetGDP(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetMacroPMI fetches China PMI data.
func (uc *UnifiedClient) GetMacroPMI(count int) ([]scraper.MacroIndicator, error) {
	if uc.useMacroScraper && uc.macroScraper != nil {
		return uc.macroScraper.GetPMI(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetMacroLPR fetches China LPR data.
func (uc *UnifiedClient) GetMacroLPR(count int) ([]scraper.MacroIndicator, error) {
	if uc.useMacroScraper && uc.macroScraper != nil {
		return uc.macroScraper.GetLPR(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetMacroShibor fetches China SHIBOR data.
func (uc *UnifiedClient) GetMacroShibor(count int) ([]scraper.MacroIndicator, error) {
	if uc.useMacroScraper && uc.macroScraper != nil {
		return uc.macroScraper.GetShibor(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetNorthboundFlow fetches intraday northbound capital flow.
func (uc *UnifiedClient) GetNorthboundFlow() ([]scraper.NorthboundFlow, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetFlowMinute()
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetNorthboundDaily fetches daily northbound capital flow history.
func (uc *UnifiedClient) GetNorthboundDaily(days int) ([]scraper.NorthboundFlow, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetDailyFlow(days)
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetTopNorthboundStocks fetches top stocks by northbound holding.
func (uc *UnifiedClient) GetTopNorthboundStocks(sortField string, count int) ([]scraper.NorthboundStock, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetTopNorthboundStocks(sortField, count)
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetTopShanghaiNorthbound fetches top stocks held by Shanghai northbound (沪股通).
func (uc *UnifiedClient) GetTopShanghaiNorthbound(sortField string, count int) ([]scraper.NorthboundStock, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetTopShanghaiNorthbound(sortField, count)
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetTopShenzhenNorthbound fetches top stocks held by Shenzhen northbound (深股通).
func (uc *UnifiedClient) GetTopShenzhenNorthbound(sortField string, count int) ([]scraper.NorthboundStock, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetTopShenzhenNorthbound(sortField, count)
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetNorthboundHolders fetches institutional holding rankings from northbound trading.
func (uc *UnifiedClient) GetNorthboundHolders(mutualType string, pageSize int) ([]*scraper.NorthboundHolder, error) {
	if uc.useNorthScraper && uc.northScraper != nil {
		return uc.northScraper.GetNorthboundHolders(mutualType, pageSize)
	}
	return nil, fmt.Errorf("northbound scraper not configured")
}

// GetFundNav fetches latest fund net asset value via goquery web parser.
func (uc *UnifiedClient) GetFundNav(fundCode string) (*scraper.FundNav, error) {
	if uc.fundNavClient != nil {
		return uc.fundNavClient.GetLatestNAV(fundCode)
	}
	return nil, fmt.Errorf("fund nav client not configured")
}

// GetFundNavHistory fetches fund NAV history via goquery web parser.
func (uc *UnifiedClient) GetFundNavHistory(fundCode string, limit int) ([]*scraper.FundNav, error) {
	if uc.fundNavClient != nil {
		return uc.fundNavClient.GetNAVRHistory(fundCode, limit)
	}
	return nil, fmt.Errorf("fund nav client not configured")
}

// GetMarginTrade fetches margin trading data via eastmoney datacenter API.
func (uc *UnifiedClient) GetMarginTrade() ([]*scraper.MarginTradeData, error) {
	if uc.marginTradeClient != nil {
		return uc.marginTradeClient.GetSummary()
	}
	return nil, fmt.Errorf("margin trade client not configured")
}

func (uc *UnifiedClient) initScrapers() {
	uc.scrapersInitOnce.Do(func() {
		uc.eastMoneyScraper = scraper.NewEastMoneyScraper()
		uc.blockTradeClient = scraper.NewBlockTradeClient()
		uc.fundHoldingClient = scraper.NewFundHoldingClient()
		uc.sinaClient = scraper.NewSinaClient()
		uc.tableParser = scraper.NewTableParser()
		uc.ocrClient = scraper.NewOCRClient()
		uc.northScraper = scraper.NewNorthboundScraper()
		uc.useNorthScraper = true
		uc.macroScraper = scraper.NewMacroScraper("")
		uc.fundNavClient = scraper.NewFundNavClient()
		uc.marginTradeClient = scraper.NewMarginTradeWebClient()
		if s, err := scraper.NewScraper(30 * time.Second); err == nil {
			uc.webScraper = s
		}
	})
}

func (uc *UnifiedClient) initBacktest() {
	if uc.backtestEngine == nil {
		uc.backtestEngine = backtest.NewEngine(1000000)
	}
}

// LimitUpPool returns the daily limit-up stock pool.
func (uc *UnifiedClient) LimitUpPool(date string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.LimitUpPool(date)
}

// LimitDownPool returns the daily limit-down stock pool.
func (uc *UnifiedClient) LimitDownPool(date string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.LimitDownPool(date)
}

// YesterdayLimitUp returns yesterday's limit-up stocks' today performance.
func (uc *UnifiedClient) YesterdayLimitUp(date string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.YesterdayLimitUp(date)
}

// HotRank returns the hot search ranking from THS.
func (uc *UnifiedClient) HotRank(limit int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.HotRank(limit)
}

// NorthBoundTop10 returns northbound top 10 traded stocks.
func (uc *UnifiedClient) NorthBoundTop10(date string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.NorthBoundTop10(date)
}

// MarketIndices returns the market index list.
func (uc *UnifiedClient) MarketIndices() ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.MarketIndices()
}

// SecurityList returns a generic security list query.
func (uc *UnifiedClient) SecurityList(fs, fields string, pn, pz int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.SecurityList(fs, fields, pn, pz)
}

// SecurityCount returns security count statistics.
func (uc *UnifiedClient) SecurityCount(secid string) (map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.SecurityCount(secid)
}

// StockBelongSector queries sectors for multiple stocks.
func (uc *UnifiedClient) StockBelongSector(codes []string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	return uc.eastMoneyScraper.StockBelongSector(codes)
}

// GetBlockTradesByDate queries block trades by date.
func (uc *UnifiedClient) GetBlockTradesByDate(date string, limit int) ([]*scraper.BlockTradeData, error) {
	uc.initScrapers()
	return uc.blockTradeClient.GetBlockTradesByDate(date, limit)
}

// SearchBlockTrades searches block trades by keyword.
func (uc *UnifiedClient) SearchBlockTrades(keyword string, limit int) ([]*scraper.BlockTradeData, error) {
	uc.initScrapers()
	return uc.blockTradeClient.SearchBlockTrades(keyword, limit)
}

// GetFundCompanies returns the list of all fund companies.
func (uc *UnifiedClient) GetFundCompanies() ([]*scraper.FundCompanyInfo, error) {
	uc.initScrapers()
	return uc.fundHoldingClient.GetFundCompanies()
}

// GetMoneySupply returns M2 money supply data.
func (uc *UnifiedClient) GetMoneySupply(count int) ([]scraper.MacroIndicator, error) {
	if uc.macroScraper != nil {
		return uc.macroScraper.GetMoneySupply(count)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetGlobalIndicator returns global macroeconomic indicators.
func (uc *UnifiedClient) GetGlobalIndicator(country, indicator string) (*scraper.MacroIndicator, error) {
	if uc.macroScraper != nil {
		return uc.macroScraper.GetGlobalIndicator(country, indicator)
	}
	return nil, fmt.Errorf("macro scraper not configured")
}

// GetMarginTrade (Sina) fetches margin trading data from Sina Finance.
func (uc *UnifiedClient) GetSinaMarginTrade(limit int) ([]*scraper.SinaMarginData, error) {
	uc.initScrapers()
	return uc.sinaClient.GetMarginTrade(limit)
}

// GetBlockTrades (Sina) fetches block trades from Sina Finance.
func (uc *UnifiedClient) GetSinaBlockTrades(limit int) ([]*scraper.SinaBlockTradeData, error) {
	uc.initScrapers()
	return uc.sinaClient.GetBlockTrades(limit)
}

// EastMoney Realtime Quote delegation
func (uc *UnifiedClient) EastMoneyRealtimeQuote(codes []string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.RealtimeQuote(codes)
}

// EastMoney Kline History delegation
func (uc *UnifiedClient) EastMoneyKlineHistory(secid, klt string, count int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.KlineHistory(secid, klt, count)
}

// EastMoney Stock Changes delegation
func (uc *UnifiedClient) EastMoneyStockChanges(changeType string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.StockChanges(changeType)
}

// EastMoney Symbol Info delegation
func (uc *UnifiedClient) EastMoneySymbolInfo(secid string) (map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.SymbolInfo(secid)
}

// EastMoney Sector Boards delegation
func (uc *UnifiedClient) EastMoneySectorBoards(boardType string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.SectorBoards(boardType)
}

// EastMoney Sector Stocks delegation
func (uc *UnifiedClient) EastMoneySectorStocks(boardCode string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.SectorStocks(boardCode)
}

// EastMoney UpDown Count delegation
func (uc *UnifiedClient) EastMoneyUpDownCount(date string) (map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.UpDownCount(date)
}

// EastMoney Belong Board delegation
func (uc *UnifiedClient) EastMoneyBelongBoard(secid string) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.BelongBoard(secid)
}

// Fund Nav Latest delegation
func (uc *UnifiedClient) FundNavLatest(fundCode string) (*scraper.FundNav, error) {
	uc.initScrapers()
	return uc.fundNavClient.GetLatestNAV(fundCode)
}

// Fund Nav History delegation
func (uc *UnifiedClient) FundNavHistory(fundCode string, limit int) ([]*scraper.FundNav, error) {
	uc.initScrapers()
	return uc.fundNavClient.GetNAVRHistory(fundCode, limit)
}

// Margin Trade Summary delegation
func (uc *UnifiedClient) MarginTradeSummary(limit int) ([]*scraper.SinaMarginData, error) {
	uc.initScrapers()
	return uc.sinaClient.GetMarginTrade(limit)
}

// Table Parser from URL delegation
func (uc *UnifiedClient) TableParserURL(url string) ([]scraper.Table, error) {
	uc.initScrapers()
	return uc.tableParser.ParseFromURL(url)
}

// Table Parser from HTML delegation
func (uc *UnifiedClient) TableParserHTML(html string) ([]scraper.Table, error) {
	uc.initScrapers()
	return uc.tableParser.ParseFromString(html)
}

// Table Parser Find Keyword delegation
func (uc *UnifiedClient) TableParserFindKeyword(tables []scraper.Table, keyword string) (*scraper.Table, error) {
	uc.initScrapers()
	return uc.tableParser.FindTableByKeyword(tables, keyword)
}

// Backtest Available Strategies delegation
func (uc *UnifiedClient) BacktestAvailableStrategies() []string {
	uc.initBacktest()
	return backtest.AvailableStrategies()
}

// Backtest Run delegation
func (uc *UnifiedClient) BacktestRun(strategy backtest.Strategy, bars []indicator.Bar) *backtest.Result {
	uc.initBacktest()
	return uc.backtestEngine.Run(strategy, bars)
}

// Backtest Combo delegation
func (uc *UnifiedClient) BacktestCombo(strategies []backtest.Strategy, bars []indicator.Bar, mode backtest.ComboMode) *backtest.ComboResult {
	uc.initBacktest()
	return backtest.RunCombo(uc.backtestEngine, strategies, bars, mode)
}

// Factor Get Info delegation
func (uc *UnifiedClient) FactorGetInfo(name string) *factor.FactorMeta {
	return factor.Get(name)
}

// Factor Analysis Report delegation
func (uc *UnifiedClient) FactorAnalysisReport(factorName string, codes []string, period, nQuantiles int) (*factor.FactorReport, error) {
	return nil, fmt.Errorf("factor analysis not yet implemented - requires factor values and returns computation")
}

// Factor Forward Returns delegation
func (uc *UnifiedClient) FactorForwardReturns(factorName string, codes []string, period, market, count int) (map[string][]float64, error) {
	return nil, fmt.Errorf("factor forward returns not yet implemented - requires factor values computation")
}

// Chanlun Merge Klines delegation
func (uc *UnifiedClient) ChanlunMergeKlines(code string, market, count int) ([]chanlun.Kline, error) {
	return nil, fmt.Errorf("chanlun merge klines requires kline data - use TCP client or collector with KlineQuery")
}

// Chanlun Find FenXing delegation
func (uc *UnifiedClient) ChanlunFindFenXing(code string, market, count int) ([]chanlun.FenXing, error) {
	return nil, fmt.Errorf("chanlun find fenxing requires kline data - use TCP client or collector with KlineQuery")
}

// Chanlun Build Bi delegation
func (uc *UnifiedClient) ChanlunBuildBi(code string, market, count int) ([]chanlun.Bi, error) {
	return nil, fmt.Errorf("chanlun build bi requires kline data - use TCP client or collector with KlineQuery")
}

// Chanlun Build ZhongShu delegation
func (uc *UnifiedClient) ChanlunBuildZhongShu(code string, market, count int) ([]chanlun.ZhongShu, error) {
	return nil, fmt.Errorf("chanlun build zhongshu requires kline data - use TCP client or collector with KlineQuery")
}

// Chanlun Find MaiMaiDian delegation
func (uc *UnifiedClient) ChanlunFindMaiMaiDian(code string, market, count int) ([]chanlun.MaiMaiDian, error) {
	return nil, fmt.Errorf("chanlun find maimaidian requires kline data - use TCP client or collector with KlineQuery")
}

// IPOCalendar delegation
func (uc *UnifiedClient) IPOCalendar(date string, limit int) ([]map[string]interface{}, error) {
	uc.initScrapers()
	if uc.eastMoneyScraper == nil {
		return nil, fmt.Errorf("east money scraper not configured")
	}
	return uc.eastMoneyScraper.IPOCalendar(date, limit)
}

// Call Auction delegation
func (uc *UnifiedClient) GetCallAuction(code string, market int) ([]proto.AuctionData, error) {
	return uc.tcpClient.GetAuction(code, market)
}

// History Minute delegation
func (uc *UnifiedClient) GetHistoryMinute(code string, market int, date string) ([]proto.HistoryMinuteTimeData, error) {
	dateNum, err := strconv.ParseUint(date, 10, 32)
	if err != nil || len(date) != 8 {
		return nil, fmt.Errorf("invalid date format: %s (use YYYYMMDD)", date)
	}
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	reply, err := uc.tcpClient.mainClient.GetHistoryMinuteTimeData(uint32(dateNum), m.Uint8(), code)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

// History Trade delegation
func (uc *UnifiedClient) GetHistoryTrade(code string, market int, date string, count int) ([]proto.HistoryTransactionData, error) {
	dateNum, err := strconv.ParseUint(date, 10, 32)
	if err != nil || len(date) != 8 {
		return nil, fmt.Errorf("invalid date format: %s (use YYYYMMDD)", date)
	}
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	if count > 500 {
		count = 500
	}
	reply, err := uc.tcpClient.mainClient.GetHistoryTransactionData(uint32(dateNum), m.Uint8(), code, 0, uint16(count))
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

// Symbol Info delegation
func (uc *UnifiedClient) GetSymbolInfo(code string, market int) (*proto.MACSymbolInfoReply, error) {
	return uc.tcpClient.GetSymbolInfo(code, market)
}

// Finance Info delegation
func (uc *UnifiedClient) GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error) {
	return uc.tcpClient.GetFinanceInfo(code, market)
}

// Index KLine delegation
func (uc *UnifiedClient) GetIndexKLine(code string, market int, period int, count int) ([]proto.IndexBar, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		m = types.MarketSH
	}
	reply, err := uc.tcpClient.mainClient.GetIndexBars(uint16(period), m.Uint8(), code, 0, uint16(count))
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

// Board Members delegation
func (uc *UnifiedClient) GetBoardMembers(boardSymbol string, count int) ([]proto.MACBoardMemberItem, error) {
	return uc.tcpClient.macClient.MACBoardMembers(boardSymbol, uint32(count))
}

// Belong Board delegation
func (uc *UnifiedClient) GetSymbolBelongBoard(code string, market uint8) ([]proto.MACBelongBoardItem, error) {
	return uc.tcpClient.macClient.MACSymbolBelongBoard(code, market)
}

// Realtime Quote delegation
func (uc *UnifiedClient) GetRealtimeQuote(code string, market int) (*proto.SecurityQuote, error) {
	return uc.tcpClient.GetQuote(code, market)
}

// MAC Board List delegation
func (uc *UnifiedClient) GetMACBoardList(boardType uint16, count uint32) ([]proto.MACBoardListItem, error) {
	return uc.tcpClient.macClient.MACBoardList(boardType, count)
}

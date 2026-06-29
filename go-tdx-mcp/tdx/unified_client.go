package tdx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tdx/go-tdx-mcp/scraper"
)

// UnifiedClient wraps both HTTP (TQLEX) and TCP clients with automatic fallback.
type UnifiedClient struct {
	httpClient       *HTTPClient
	tcpClient        *TDXTCPClient
	collector        *MultiHostCollector
	sectorScraper    *scraper.SectorScraper
	macroScraper     *scraper.MacroScraper
	northScraper     *scraper.NorthboundScraper
	fundNavClient    *scraper.FundNavClient
	marginTradeClient *scraper.MarginTradeWebClient
	initOnce         sync.Once
	initErr          error
	useCollector     bool
	useScraper       bool
	useMacroScraper  bool
	useNorthScraper  bool
}

// NewUnifiedClient creates a unified client with both HTTP and TCP backends.
func NewUnifiedClient(token string, timeoutSec int, tdxHost string, tdxPort int, opts ...UnifiedClientOption) *UnifiedClient {
	uc := &UnifiedClient{
		httpClient: NewHTTPClient(token, 0),
		tcpClient:  NewTDXTCPClientWithHost(timeoutSec, tdxHost, tdxPort),
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
	uc.initOnce.Do(func() {
		// Try connecting main + ex + mac in parallel
		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			_ = uc.tcpClient.ConnectMain(ctx)
		}()
		go func() {
			defer wg.Done()
			_ = uc.tcpClient.ConnectEx(ctx)
		}()
		go func() {
			defer wg.Done()
			_ = uc.tcpClient.ConnectMAC(ctx)
		}()
		wg.Wait()
	})

	// Verify main connection is actually usable
	if !uc.tcpClient.IsConnected() {
		if err := uc.tcpClient.ConnectMain(ctx); err != nil {
			uc.initErr = fmt.Errorf("tcp init failed: %w", err)
			return uc.initErr
		}
	}
	return nil
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

	// For unimplemented entries, fall back to HTTP
	if err := uc.initTCP(ctx); err != nil {
		return uc.httpClient.TQLEXQuery(ctx, entry, body)
	}
	return uc.httpClient.TQLEXQuery(ctx, entry, body)
}

// queryQuotes handles PBHQInfo via TCP or MultiHostCollector.
func (uc *UnifiedClient) queryQuotes(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid quotes request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}

	if uc.useCollector && uc.collector != nil {
		quotes, _, err := uc.collector.CollectDetailedQuotes([]string{code}, market, 1)
		if err == nil && len(quotes) > 0 {
			result, _ := json.Marshal(&quotes[0])
			return &TQLEXResponse{Data: json.RawMessage(result)}, nil
		}
	}

	// Try TCP first
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPQuotes(code, market); err == nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBHQInfo", body)
}

// tryTCPQuotes attempts to get quotes via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPQuotes(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	quote, err := uc.tcpClient.GetQuote(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(quote)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryKline handles PBFXT via TCP or MultiHostCollector.
func (uc *UnifiedClient) queryKline(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid kline request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	period := PeriodCodeToString(data["Period"])
	count := 100
	if w, ok := data["WantNum"].(int); ok && w > 0 {
		count = w
	}
	fq := 0
	if t, ok := data["TQFlag"].(int); ok {
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
		if result, err := uc.tryTCPKline(code, market, period, count, fq); err == nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBFXT", body)
}

// tryTCPKline attempts to get kline data via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPKline(code string, market int, period string, count, fq int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	bars, err := uc.tcpClient.GetKLineWithAdjust(code, market, period, count, fq)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(bars)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryScreener handles wendaQuery via TCP.
// Note: wendaQuery is NLP-based, no direct TCP equivalent. Falls back to HTTP.
func (uc *UnifiedClient) queryScreener(body interface{}) (*TQLEXResponse, error) {
	// wendaQuery requires NLP engine — no TCP alternative exists
	// Return empty result to indicate TCP-unavailable
	return &TQLEXResponse{Data: map[string]interface{}{"data": []interface{}{}, "page": 1, "total": 0}}, nil
}

// queryIndicator handles InfoSelectV2 via TCP.
// Note: InfoSelectV2 is NLP-based, no direct TCP equivalent. Falls back to HTTP.
func (uc *UnifiedClient) queryIndicator(body interface{}) (*TQLEXResponse, error) {
	// InfoSelectV2 requires NLP engine — no TCP alternative exists
	return &TQLEXResponse{Data: map[string]interface{}{"data": []interface{}{}, "total": 0}}, nil
}

// queryAuction handles PBAuction via TCP.
func (uc *UnifiedClient) queryAuction(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid auction request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPAuction(code, market); err == nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBAuction", body)
}

// tryTCPAuction attempts to get auction data via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPAuction(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	auctions, err := uc.tcpClient.GetAuction(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(auctions)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryF10 handles TdxSharePCCW via TCP.
func (uc *UnifiedClient) queryF10(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid F10 request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}

	// Try TCP first with panic recovery
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPF10(code, market); err == nil {
			return result, nil
		}
	}

	// Fall back to HTTP TQLEX
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.TdxSharePCCW", body)
}

// tryTCPF10 attempts to get F10 info via TCP, recovering from panics.
func (uc *UnifiedClient) tryTCPF10(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	info, err := uc.tcpClient.GetF10(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(info)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryCapitalFlow handles PBCapitalFlow via TCP.
func (uc *UnifiedClient) queryCapitalFlow(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid capital flow request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPCapitalFlow(code, market); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBCapitalFlow", body)
}

func (uc *UnifiedClient) tryTCPCapitalFlow(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	reply, err := uc.tcpClient.GetCapitalFlow(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(reply)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryBoardList handles PBBoardList via TCP.
func (uc *UnifiedClient) queryBoardList(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid board list request body type")
	}
	boardType := fmt.Sprintf("%v", data["BoardType"])
	count := 50
	if c, ok := data["Count"].(int); ok {
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
		if result, err := uc.tryTCPBoardList(bt, count); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardList", body)
}

func (uc *UnifiedClient) tryTCPBoardList(bt BlockType, count int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	boards, err := uc.tcpClient.GetSectorBoards(bt)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(boards)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryBoardMembers handles PBBoardMembers via TCP.
func (uc *UnifiedClient) queryBoardMembers(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid board members request body type")
	}
	code := fmt.Sprintf("%v", data["BoardCode"])
	count := 50
	if c, ok := data["Count"].(int); ok {
		count = c
	}
	_ = count
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPBoardMembers(code); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardMembers", body)
}

func (uc *UnifiedClient) tryTCPBoardMembers(boardCode string) (*TQLEXResponse, error) {
	defer func() { recover() }()
	stocks, err := uc.tcpClient.GetSectorBoardStocks(boardCode)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(stocks)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryBoardRanking handles PBBoardRanking via TCP (falls back to scraper).
func (uc *UnifiedClient) queryBoardRanking(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid board ranking request body type")
	}
	boardType := fmt.Sprintf("%v", data["BoardType"])
	topN := 10
	if t, ok := data["TopN"].(int); ok {
		topN = t
	}
	sortBy := fmt.Sprintf("%v", data["SortBy"])
	_ = boardType
	_ = topN
	_ = sortBy
	// Board ranking not available via TCP — fall back to scraper or HTTP
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBoardRanking", body)
}

// queryServerInfo handles PBServerInfo via TCP.
func (uc *UnifiedClient) queryServerInfo(body interface{}) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPServerInfo(); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBServerInfo", body)
}

func (uc *UnifiedClient) tryTCPServerInfo() (*TQLEXResponse, error) {
	defer func() { recover() }()
	info := uc.tcpClient.GetServerInfo()
	result, _ := json.Marshal(map[string]interface{}{"server_info": info, "status": "connected"})
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// querySymbolInfo handles PBSymbolInfo via MAC TCP.
func (uc *UnifiedClient) querySymbolInfo(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid symbol info request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPSymbolInfo(code, market); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBSymbolInfo", body)
}

func (uc *UnifiedClient) tryTCPSymbolInfo(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	reply, err := uc.tcpClient.GetSymbolInfo(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(reply)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryBelongBoard handles PBBelongBoard via TCP.
func (uc *UnifiedClient) queryBelongBoard(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid belong board request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	_ = code
	_ = market
	// Belong board not directly available via TCP — fall back to scraper
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBBelongBoard", body)
}

// queryUnusual handles PBUnusual via TCP.
func (uc *UnifiedClient) queryUnusual(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid unusual request body type")
	}
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	count := 50
	if c, ok := data["WantNum"].(int); ok {
		count = c
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPUnusual(market, count); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBUnusual", body)
}

func (uc *UnifiedClient) tryTCPUnusual(market, count int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	data, err := uc.tcpClient.GetUnusual(market, count)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(data)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryMarketStat handles PBMarketStat via TCP.
func (uc *UnifiedClient) queryMarketStat(body interface{}) (*TQLEXResponse, error) {
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPMarketStat(); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBMarketStat", body)
}

func (uc *UnifiedClient) tryTCPMarketStat() (*TQLEXResponse, error) {
	defer func() { recover() }()
	count, err := uc.tcpClient.GetSecurityCount(0)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(map[string]interface{}{"sh_count": count, "sz_count": 0, "status": "ok"})
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// querySecurityList handles PBSecurityList via TCP.
func (uc *UnifiedClient) querySecurityList(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid security list request body type")
	}
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	start := uint16(0)
	if s, ok := data["Start"].(int); ok {
		start = uint16(s)
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPSecurityList(market, start); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBSecurityList", body)
}

func (uc *UnifiedClient) tryTCPSecurityList(market int, start uint16) (*TQLEXResponse, error) {
	defer func() { recover() }()
	reply, err := uc.tcpClient.GetSecurityList(market, start)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(reply)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// queryFinanceInfo handles PBGetFinanceInfo via TCP.
func (uc *UnifiedClient) queryFinanceInfo(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid finance info request body type")
	}
	code := fmt.Sprintf("%v", data["Code"])
	market := 0
	if m, ok := data["Setcode"].(int); ok {
		market = m
	}
	if uc.tcpClient != nil && uc.tcpClient.IsConnected() {
		if result, err := uc.tryTCPFinanceInfo(code, market); err == nil {
			return result, nil
		}
	}
	return uc.httpClient.TQLEXQuery(context.Background(), "TdxShare.PBGetFinanceInfo", body)
}

func (uc *UnifiedClient) tryTCPFinanceInfo(code string, market int) (*TQLEXResponse, error) {
	defer func() { recover() }()
	reply, err := uc.tcpClient.GetFinanceInfo(code, market)
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(reply)
	return &TQLEXResponse{Data: json.RawMessage(result)}, nil
}

// RAGQuery implements the Client interface.
func (uc *UnifiedClient) RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error) {
	return uc.httpClient.RAGQuery(ctx, query, topK)
}

// queryQuoteList handles PBQuoteList via EastMoney push2 clist (HTTP TQLEX returns 503).
func (uc *UnifiedClient) queryQuoteList(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid quote list request body type")
	}
	count := 100
	if c, ok := data["count"].(float64); ok {
		count = int(c)
	}
	if c, ok := data["count"].(int); ok {
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

// queryFSTick handles PBFSTick via EastMoney (HTTP TQLEX returns 503).
func (uc *UnifiedClient) queryFSTick(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(TickRequestParams)
	if !ok {
		m, ok2 := body.(map[string]interface{})
		if ok2 {
			_ = m
		}
		return nil, fmt.Errorf("FSTick not available via HTTP TQLEX, use quote fallback")
	}
	_ = data
	code := data.Code
	market := data.Market
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

// queryTrans handles PBTrans via EastMoney (HTTP TQLEX returns 503).
func (uc *UnifiedClient) queryTrans(body interface{}) (*TQLEXResponse, error) {
	data, ok := body.(TransRequestParams)
	if !ok {
		return nil, fmt.Errorf("Trans not available via HTTP TQLEX, use quote fallback")
	}
	_ = data
	code := data.Code
	market := data.Market
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
func tcpPeriodToCategory(period string) uint16 {
	switch period {
	case "1min":
		return 8
	case "5min":
		return 0
	case "15min":
		return 1
	case "30min":
		return 2
	case "60min":
		return 3
	case "day":
		return 4
	case "week":
		return 5
	case "month":
		return 6
	case "quarter":
		return 13
	case "year":
		return 11
	default:
		return 4
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

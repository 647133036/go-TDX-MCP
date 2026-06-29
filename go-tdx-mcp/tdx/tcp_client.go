package tdx

import (
	"context"
	"fmt"

	"github.com/bensema/gotdx"
	"github.com/bensema/gotdx/proto"
	"github.com/bensema/gotdx/types"
)

// TDXTCPClient wraps gotdx for our MCP server.
type TDXTCPClient struct {
	mainClient  *gotdx.Client
	exClient    *gotdx.Client
	macClient   *gotdx.Client
	mainAddr    string
	exAddr      string
	macAddr     string
	customHost  string
	customPort  int
}

// NewTDXTCPClient creates a new TDX TCP client with auto-select of fastest hosts.
func NewTDXTCPClient(timeoutSec int) *TDXTCPClient {
	return newTDXTCPClientWithOptions(timeoutSec, "", 0)
}

// NewTDXTCPClientWithHost creates a TCP client with a custom host and port.
func NewTDXTCPClientWithHost(timeoutSec int, host string, port int) *TDXTCPClient {
	return newTDXTCPClientWithOptions(timeoutSec, host, port)
}

func newTDXTCPClientWithOptions(timeoutSec int, customHost string, customPort int) *TDXTCPClient {
	if timeoutSec <= 0 {
		timeoutSec = 6
	}
	baseOpts := []gotdx.Option{
		gotdx.WithAutoSelectFastest(true),
		gotdx.WithTimeoutSec(timeoutSec),
	}
	var mainOpts, exOpts, macOpts []gotdx.Option
	mainOpts = append(mainOpts, baseOpts...)
	exOpts = append(exOpts, baseOpts...)
	macOpts = append(macOpts, baseOpts...)

	if customHost != "" && customPort > 0 {
		addr := fmt.Sprintf("%s:%d", customHost, customPort)
		mainOpts = append(mainOpts, gotdx.WithTCPAddress(addr))
		exOpts = append(exOpts, gotdx.WithExTCPAddress(addr))
		macOpts = append(macOpts, gotdx.WithMacTCPAddress(addr))
	}

	return &TDXTCPClient{
		mainClient: gotdx.New(mainOpts...),
		exClient:   gotdx.NewEx(exOpts...),
		macClient:  gotdx.NewMAC(macOpts...),
		customHost: customHost,
		customPort: customPort,
	}
}

// ConnectMain connects to the main A-share server (port 7709).
func (c *TDXTCPClient) ConnectMain(ctx context.Context) error {
	reply, err := c.mainClient.Connect()
	if err != nil {
		return fmt.Errorf("connect main server: %w", err)
	}
	c.mainAddr = c.mainClient.CurrentAddress()
	_ = reply
	return nil
}

// ConnectEx connects to the extension market server (HK/US/futures).
func (c *TDXTCPClient) ConnectEx(ctx context.Context) error {
	reply, err := c.exClient.ConnectEx()
	if err != nil {
		return fmt.Errorf("connect ex server: %w", err)
	}
	c.exAddr = c.exClient.CurrentAddress()
	_ = reply
	return nil
}

// ConnectMAC connects to the MAC (market monitor) server.
func (c *TDXTCPClient) ConnectMAC(ctx context.Context) error {
	err := c.macClient.ConnectMAC()
	if err != nil {
		return fmt.Errorf("connect mac server: %w", err)
	}
	c.macAddr = c.macClient.CurrentAddress()
	return nil
}

// Disconnect closes all connections.
func (c *TDXTCPClient) Disconnect() error {
	var errs []error
	if c.mainClient != nil {
		if err := c.mainClient.Disconnect(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.exClient != nil {
		if err := c.exClient.Disconnect(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.macClient != nil {
		if err := c.macClient.Disconnect(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}
	return nil
}

// GetQuote fetches real-time quote for a stock.
func (c *TDXTCPClient) GetQuote(code string, market int) (*proto.SecurityQuote, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	result, err := c.mainClient.StockQuotesDetail([]uint8{m.Uint8()}, []string{code})
	if err != nil {
		return nil, fmt.Errorf("quote query: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no quote data for %s", code)
	}
	return &result[0], nil
}

// GetKLine fetches K-line data.
func (c *TDXTCPClient) GetKLine(code string, market int, period string, count int, adjust int) ([]proto.SecurityBar, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	category := uint16(proto.KLINE_TYPE_RI_K) // daily
	switch period {
	case "1min":
		category = uint16(proto.KLINE_TYPE_1MIN)
	case "5min":
		category = uint16(proto.KLINE_TYPE_5MIN)
	case "15min":
		category = uint16(proto.KLINE_TYPE_15MIN)
	case "30min":
		category = uint16(proto.KLINE_TYPE_30MIN)
	case "60min":
		category = uint16(proto.KLINE_TYPE_1HOUR)
	case "day":
		category = uint16(proto.KLINE_TYPE_RI_K)
	case "week":
		category = uint16(proto.KLINE_TYPE_WEEKLY)
	case "month":
		category = uint16(proto.KLINE_TYPE_MONTHLY)
	case "quarter":
		category = uint16(proto.KLINE_TYPE_3MONTH)
	case "year":
		category = uint16(proto.KLINE_TYPE_YEARLY)
	}

	adjustType := uint16(types.AdjustNone)
	switch adjust {
	case 1:
		adjustType = types.AdjustQFQ
	case 2:
		adjustType = types.AdjustHFQ
	}

	bars, err := c.mainClient.StockFullKLine(category, m.Uint8(), code, 1, adjustType, nil)
	if err != nil {
		return nil, fmt.Errorf("kline query: %w", err)
	}
	if len(bars) > count {
		bars = bars[len(bars)-count:]
	}
	return bars, nil
}

// GetTickChart fetches intraday tick chart data.
func (c *TDXTCPClient) GetTickChart(code string, market int) ([]proto.MinuteTimeData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	reply, err := c.mainClient.GetMinuteTimeData(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("tick chart query: %w", err)
	}
	return reply.List, nil
}

// GetTransaction fetches tick-by-tick transaction data.
func (c *TDXTCPClient) GetTransaction(code string, market int, count int) ([]proto.TransactionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	reply, err := c.mainClient.GetTransactionData(m.Uint8(), code, 0, uint16(count))
	if err != nil {
		return nil, fmt.Errorf("transaction query: %w", err)
	}
	return reply.List, nil
}

// GetAuction fetches call auction data.
func (c *TDXTCPClient) GetAuction(code string, market int) ([]proto.AuctionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	reply, err := c.mainClient.GetAuction(m.Uint8(), code, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("auction query: %w", err)
	}
	return reply.List, nil
}

// GetCapitalFlow fetches capital flow data via MAC.
func (c *TDXTCPClient) GetCapitalFlow(code string, market int) (*proto.MACCapitalFlowReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	data, err := c.macClient.MACCapitalFlow(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("capital flow query: %w", err)
	}
	return data, nil
}

// GetUnusual fetches unusual movement data.
func (c *TDXTCPClient) GetUnusual(market int, count int) ([]proto.UnusualData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		m = types.MarketSZ
	}

	reply, err := c.mainClient.GetUnusual(m.Uint8(), 0, uint32(count))
	if err != nil {
		return nil, fmt.Errorf("unusual query: %w", err)
	}
	return reply.List, nil
}

// GetServerInfo fetches server status info.
func (c *TDXTCPClient) GetServerInfo() string {
	if c.mainAddr != "" {
		return c.mainAddr
	}
	return c.mainClient.CurrentAddress()
}

// GetSymbolInfo fetches stock basic info via MAC.
func (c *TDXTCPClient) GetSymbolInfo(code string, market int) (*proto.MACSymbolInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	data, err := c.macClient.MACSymbolInfo(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("symbol info query: %w", err)
	}
	return data, nil
}

// GetF10 fetches F10 company info.
func (c *TDXTCPClient) GetF10(code string, market int) (*gotdx.CompanyInfoBundle, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	data, err := c.mainClient.GetCompanyInfo(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("F10 query: %w", err)
	}
	return data, nil
}

// GetFinanceInfo fetches financial info.
func (c *TDXTCPClient) GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	data, err := c.mainClient.GetFinanceInfo(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("finance info query: %w", err)
	}
	return data, nil
}

// GetXDXRInfo fetches dividend/ex-rights info.
func (c *TDXTCPClient) GetXDXRInfo(code string, market int) (*proto.GetXDXRInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	data, err := c.mainClient.GetXDXRInfo(m.Uint8(), code)
	if err != nil {
		return nil, fmt.Errorf("xdxr info query: %w", err)
	}
	return data, nil
}

// ExGetQuote fetches extension market (HK/US/futures) quote.
func (c *TDXTCPClient) ExGetQuote(code string, category uint8) (*proto.ExQuoteItem, error) {
	data, err := c.exClient.ExQuote(category, code)
	if err != nil {
		return nil, fmt.Errorf("ex quote query: %w", err)
	}
	return data, nil
}

// ExGetKLine fetches extension market K-line.
func (c *TDXTCPClient) ExGetKLine(category uint8, code string, period uint16, count int) ([]proto.ExKLineItem, error) {
	bars, err := c.exClient.ExKLine(category, code, period, 0, uint16(count), 1)
	if err != nil {
		return nil, fmt.Errorf("ex kline query: %w", err)
	}
	return bars, nil
}

// ExGetQuotes batch fetches extension market quotes.
func (c *TDXTCPClient) ExGetQuotes(categories []uint8, codes []string) ([]proto.ExQuoteItem, error) {
	data, err := c.exClient.ExQuotes(categories, codes)
	if err != nil {
		return nil, fmt.Errorf("ex quotes query: %w", err)
	}
	return data, nil
}

// GetKLineWithAdjust fetches K-line with context-aware client-side adjustment.
// This replicates the tdxrs adjuster algorithm: fetches raw bars + context bars,
// fetches xdxr events, computes adjustment factors, and applies them locally.
// adjust=0: none, 1: qfq, 2: hfq
func (c *TDXTCPClient) GetKLineWithAdjust(code string, market int, period string, count int, adjust int) ([]proto.SecurityBar, error) {
	// 1. Fetch raw bars (no server-side adjust)
	rawBars, err := c.GetKLine(code, market, period, count+500, 0)
	if err != nil {
		return nil, fmt.Errorf("raw kline: %w", err)
	}

	// 2. Fetch xdxr info
	xdxrReply, err := c.GetXDXRInfo(code, market)
	if err != nil {
		// No xdxr data means no dividends — return raw bars as-is
		if len(rawBars) > count {
			rawBars = rawBars[len(rawBars)-count:]
		}
		return rawBars, nil
	}

	// 3. Build XdXrEvent list from xdxrReply
	events := buildXdXrEvents(xdxrReply)
	if len(events) == 0 {
		if len(rawBars) > count {
			rawBars = rawBars[len(rawBars)-count:]
		}
		return rawBars, nil
	}

	// 4. Determine context bars needed
	needsContext := false
	for _, evt := range events {
		evtDate := evt.Year*10000 + evt.Month*100 + evt.Day
		for _, bar := range rawBars {
			barDate := bar.Year*10000 + bar.Month*100 + bar.Day
			if barDate > evtDate {
				needsContext = true
				break
			}
		}
		if needsContext {
			break
		}
	}

	var contextBars []proto.SecurityBar
	if needsContext {
		contextBars, err = c.GetKLine(code, market, period, 1000, 0)
		if err != nil {
			contextBars = nil
		}
	}

	// 5. Apply client-side adjustment
	adjusted := applyAdjustToBars(rawBars, contextBars, events, adjust)

	// 6. Trim to requested count
	if len(adjusted) > count {
		adjusted = adjusted[len(adjusted)-count:]
	}
	return adjusted, nil
}

// GetBatchQuotes fetches real-time quotes for multiple stocks in one call.
// Returns []proto.SecurityQuote for the given (market, code) pairs.
func (c *TDXTCPClient) GetBatchQuotes(pairs []struct{ Market int; Code string }) ([]proto.SecurityQuote, error) {
	if len(pairs) == 0 {
		return nil, fmt.Errorf("empty quote pairs")
	}
	markets := make([]uint8, len(pairs))
	codes := make([]string, len(pairs))
	for i, p := range pairs {
		m := types.MarketSZ
		if p.Market == 1 {
			m = types.MarketSH
		}
		markets[i] = m.Uint8()
		codes[i] = p.Code
	}
	return c.mainClient.StockQuotesDetail(markets, codes)
}

// GetSecurityList fetches the full stock list for a market (with caching).
func (c *TDXTCPClient) GetSecurityList(market int, start uint16) (*proto.GetSecurityListReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return c.mainClient.GetSecurityList(m.Uint8(), start)
}

// GetSecurityCount fetches the total number of securities in a market (with caching).
func (c *TDXTCPClient) GetSecurityCount(market int) (uint16, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return 0, fmt.Errorf("unknown market: %d", market)
	}
	reply, err := c.mainClient.GetSecurityCount(m.Uint8())
	if err != nil {
		return 0, err
	}
	return reply.Count, nil
}

// PoolStats returns connection pool statistics.
func (c *TDXTCPClient) PoolStats() string {
	return fmt.Sprintf("main: addr=%s, ex=%s, mac=%s", c.GetServerInfo(), c.exAddr, c.macAddr)
}

// IsConnected returns whether the main TCP connection is alive.
func (c *TDXTCPClient) IsConnected() bool {
	return c.mainAddr != ""
}

// =============================================================================
// Sector/Board Data (板块数据)
// =============================================================================

// BlockType identifies the type of sector board.
type BlockType int

const (
	BlockIndustry BlockType = iota // 行业板块
	BlockConcept                   // 概念板块
	BlockRegion                    // 地域板块
	BlockIndex                     // 指数板块
	BlockPolicy                    // 政策板块
	BlockCustom                    // 自定义板块
)

// BlockFilename returns the TDX protocol filename for this block type.
func (bt BlockType) BlockFilename() string {
	switch bt {
	case BlockIndustry:
		return "block_gy.dat"
	case BlockConcept:
		return "block_gn.dat"
	case BlockRegion:
		return "block_dy.dat"
	case BlockIndex:
		return "block_zs.dat"
	case BlockPolicy:
		return "block_zc.dat"
	case BlockCustom:
		return "block_zdy.dat"
	default:
		return "block_gy.dat"
	}
}

// BlockName returns the human-readable name for this block type.
func (bt BlockType) BlockName() string {
	switch bt {
	case BlockIndustry:
		return "行业板块"
	case BlockConcept:
		return "概念板块"
	case BlockRegion:
		return "地域板块"
	case BlockIndex:
		return "指数板块"
	case BlockPolicy:
		return "政策板块"
	case BlockCustom:
		return "自定义板块"
	default:
		return "未知板块"
	}
}

// SectorBoard represents a single sector board entry.
type SectorBoard struct {
	Code     string // 板块代码
	Name     string // 板块名称
	Type     string // 板块类型 (行业/概念/地域)
	StockCnt int    // 成分股数量
}

// GetSectorBoards fetches all boards of a given type.
func (c *TDXTCPClient) GetSectorBoards(bt BlockType) ([]SectorBoard, error) {
	data, err := c.mainClient.GetBlockFile(bt.BlockFilename())
	if err != nil {
		return nil, fmt.Errorf("get block file %s: %w", bt.BlockFilename(), err)
	}
	groups, err := gotdx.ParseBlockGroups(data)
	if err != nil {
		return nil, fmt.Errorf("parse block groups: %w", err)
	}
	boards := make([]SectorBoard, len(groups))
	for i, g := range groups {
		boards[i] = SectorBoard{
			Code:     g.BlockName,
			Name:     g.BlockName,
			Type:     bt.BlockName(),
			StockCnt: g.StockCount,
		}
	}
	return boards, nil
}

// GetSectorBoardStocks fetches constituent stocks of a specific board.
func (c *TDXTCPClient) GetSectorBoardStocks(boardCode string) ([]string, error) {
	// Try to find the board file by scanning common filenames
	files := []string{"block_gy.dat", "block_gn.dat", "block_dy.dat", "block_zs.dat", "block_zc.dat", "block_zdy.dat"}
	for _, filename := range files {
		data, err := c.mainClient.GetBlockFile(filename)
		if err != nil {
			continue
		}
		groups, err := gotdx.ParseBlockGroups(data)
		if err != nil {
			continue
		}
		for _, g := range groups {
			if g.BlockName == boardCode {
				return g.Codes, nil
			}
		}
	}
	return nil, fmt.Errorf("board %s not found", boardCode)
}

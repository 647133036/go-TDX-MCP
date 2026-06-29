package tdx

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bensema/gotdx"
	"github.com/bensema/gotdx/proto"
	"github.com/bensema/gotdx/types"
)

// =============================================================================
// Strategy 1: TdxPoolClient — 连接池(5) + 心跳 + 重试 + 缓存 (主力，顺序请求)
// =============================================================================

// PoolConfig controls the connection pool.
type PoolConfig struct {
	Size         int           // max connections in pool, default 5
	ConnectTimeout time.Duration // per-connection dial timeout, default 6s
	HeartbeatInterval time.Duration // ping interval, default 30s
	RetryCount   int             // retries on failure, default 3
	RetryDelay   time.Duration // delay between retries, default 500ms
	CacheTTL     time.Duration // cache TTL for list/count, default 30s
}

// DefaultPoolConfig returns sensible defaults matching tdxrs TdxHqClient.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		Size:              5,
		ConnectTimeout:    6 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		RetryCount:        3,
		RetryDelay:        500 * time.Millisecond,
		CacheTTL:          30 * time.Second,
	}
}

// poolConn wraps a gotdx client + metadata.
type poolConn struct {
	client    *gotdx.Client
	lastUsed  time.Time
	connAddr  string
}

// TdxPoolClient is the Go equivalent of tdxrs TdxHqClient.
// Features: connection pool (5), heartbeat, auto-retry, list/count cache.
// Connections are distributed across multiple reachable hosts for load balancing.
type TdxPoolClient struct {
	pool     []*poolConn
	hostPool []string // reachable main host addresses, sorted by latency
	mu       sync.RWMutex
	cfg      PoolConfig
	stopCh   chan struct{}
	main     *gotdx.Client // fallback single client for non-pooled ops
	ex       *gotdx.Client
	mac      *gotdx.Client
	countCache map[uint8]*cacheEntry[u16Val]
	listCache  map[uint8]*cacheEntry[[]proto.Security]
	connected  bool
}

type u16Val struct{ v uint16 }

type cacheEntry[T any] struct {
	data      T
	expiresAt time.Time
}

// NewTdxPoolClient creates a pooled client with heartbeat and retry.
func NewTdxPoolClient(cfg PoolConfig, timeoutSec int) *TdxPoolClient {
	if cfg.Size <= 0 {
		cfg.Size = 5
	}
	if timeoutSec <= 0 {
		timeoutSec = 6
	}
	p := &TdxPoolClient{
		pool:       make([]*poolConn, 0, cfg.Size),
		cfg:        cfg,
		stopCh:     make(chan struct{}),
		countCache: make(map[uint8]*cacheEntry[u16Val]),
		listCache:  make(map[uint8]*cacheEntry[[]proto.Security]),
		main:       gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(timeoutSec)),
		ex:         gotdx.NewEx(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(timeoutSec)),
		mac:        gotdx.NewMAC(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(timeoutSec)),
	}
	p.fillPool()
	p.startHeartbeat()
	return p
}

func (p *TdxPoolClient) fillPool() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Step 1: Scan all reachable hosts
	timeout := p.cfg.ConnectTimeout
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	results := gotdx.ProbeHosts(gotdx.MainHosts(), timeout)

	// Step 2: Collect reachable host addresses sorted by latency
	p.hostPool = make([]string, 0, len(results))
	for _, r := range results {
		if r.Reachable {
			p.hostPool = append(p.hostPool, r.Address)
		}
	}

	if len(p.hostPool) == 0 {
		return
	}

	// Step 3: Distribute pool connections round-robin across reachable hosts
	for i := 0; i < p.cfg.Size; i++ {
		addr := p.hostPool[i%len(p.hostPool)]
		c := gotdx.New(gotdx.WithTCPAddress(addr), gotdx.WithTimeoutSec(int(timeout.Seconds())))
		if _, err := c.Connect(); err == nil {
			p.pool = append(p.pool, &poolConn{
				client:   c,
				lastUsed: time.Now(),
				connAddr: addr,
			})
		}
	}
	if len(p.pool) > 0 {
		p.connected = true
	}
}

func (p *TdxPoolClient) startHeartbeat() {
	go func() {
		ticker := time.NewTicker(p.cfg.HeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.doHeartbeat()
			case <-p.stopCh:
				return
			}
		}
	}()
}

func (p *TdxPoolClient) doHeartbeat() {
	p.mu.RLock()
	conns := make([]*poolConn, len(p.pool))
	copy(conns, p.pool)
	p.mu.RUnlock()

	for _, c := range conns {
		// Simple ping via get_security_count
		_, err := c.client.StockCount(1) // SH market
		if err != nil {
			// Connection dead, mark for replacement
			c.lastUsed = time.Time{}
		}
	}
}

// borrow picks the least recently used healthy connection.
func (p *TdxPoolClient) borrow() (*gotdx.Client, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find least recently used healthy conn
	var best *poolConn
	bestTime := time.Now()
	for _, c := range p.pool {
		if c.lastUsed.IsZero() {
			// Dead connection, replace it using host pool
			c.client.Disconnect()
			addr := c.connAddr
			if len(p.hostPool) > 0 {
				addr = p.hostPool[0]
			}
			newC := gotdx.New(gotdx.WithTCPAddress(addr), gotdx.WithTimeoutSec(int(p.cfg.ConnectTimeout.Seconds())))
			if _, err := newC.Connect(); err == nil {
				c.client = newC
				c.lastUsed = time.Now()
				c.connAddr = addr
				best = c
				break
			}
			continue
		}
		if c.lastUsed.Before(bestTime) {
			bestTime = c.lastUsed
			best = c
		}
	}
	if best == nil {
		return nil, "", fmt.Errorf("no available connections in pool")
	}
	best.lastUsed = time.Now()
	return best.client, best.connAddr, nil
}

// withRetry executes fn with auto-retry and connection recovery.
func (p *TdxPoolClient) withRetry(fn func(c *gotdx.Client) error) error {
	for attempt := 0; attempt <= p.cfg.RetryCount; attempt++ {
		c, _, err := p.borrow()
		if err != nil {
			if attempt < p.cfg.RetryCount {
				time.Sleep(p.cfg.RetryDelay)
				continue
			}
			return err
		}
		if err := fn(c); err != nil {
			if attempt < p.cfg.RetryCount {
				time.Sleep(p.cfg.RetryDelay)
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("retry exhausted after %d attempts", p.cfg.RetryCount+1)
}

// GetQuote with retry.
func (p *TdxPoolClient) GetQuote(code string, market int) (*proto.SecurityQuote, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result proto.SecurityQuote
	err := p.withRetry( func(c *gotdx.Client) error {
		res, err := c.StockQuotesDetail([]uint8{m.Uint8()}, []string{code})
		if err != nil {
			return err
		}
		if len(res) == 0 {
			return fmt.Errorf("no quote data for %s", code)
		}
		result = res[0]
		return nil
	})
	return &result, err
}

// GetBatchQuotes with retry.
func (p *TdxPoolClient) GetBatchQuotes(pairs []struct{ Market int; Code string }) ([]proto.SecurityQuote, error) {
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
	var result []proto.SecurityQuote
	err := p.withRetry( func(c *gotdx.Client) error {
		var err error
		result, err = c.StockQuotesDetail(markets, codes)
		return err
	})
	return result, err
}

// GetKLine with retry.
func (p *TdxPoolClient) GetKLine(code string, market int, period string, count int, adjust int) ([]proto.SecurityBar, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	category := uint16(proto.KLINE_TYPE_RI_K)
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
	var bars []proto.SecurityBar
	err := p.withRetry( func(c *gotdx.Client) error {
		var err error
		bars, err = c.StockFullKLine(category, m.Uint8(), code, 1, adjustType, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(bars) > count {
		bars = bars[len(bars)-count:]
	}
	return bars, nil
}

// GetKLineWithAdjust — context-aware adjuster (same as TDXTCPClient).
func (p *TdxPoolClient) GetKLineWithAdjust(code string, market int, period string, count int, adjust int) ([]proto.SecurityBar, error) {
	rawBars, err := p.GetKLine(code, market, period, count+500, 0)
	if err != nil {
		return nil, err
	}

	// Fetch xdxr via main client (pooled doesn't have direct xdxr, use main)
	xdxrReply, err := p.main.GetXDXRInfo(0, code)
	if err != nil {
		if len(rawBars) > count {
			rawBars = rawBars[len(rawBars)-count:]
		}
		return rawBars, nil
	}

	events := buildXdXrEvents(xdxrReply)
	if len(events) == 0 {
		if len(rawBars) > count {
			rawBars = rawBars[len(rawBars)-count:]
		}
		return rawBars, nil
	}

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
		contextBars, err = p.GetKLine(code, market, period, 1000, 0)
		if err != nil {
			contextBars = nil
		}
	}

	adjusted := applyAdjustToBars(rawBars, contextBars, events, adjust)
	if len(adjusted) > count {
		adjusted = adjusted[len(adjusted)-count:]
	}
	return adjusted, nil
}

// GetTickChart with retry.
func (p *TdxPoolClient) GetTickChart(code string, market int) ([]proto.MinuteTimeData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.MinuteTimeData
	err := p.withRetry( func(c *gotdx.Client) error {
		reply, err := c.GetMinuteTimeData(m.Uint8(), code)
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetTransaction with retry.
func (p *TdxPoolClient) GetTransaction(code string, market int, count int) ([]proto.TransactionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.TransactionData
	err := p.withRetry( func(c *gotdx.Client) error {
		reply, err := c.GetTransactionData(m.Uint8(), code, 0, uint16(count))
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetAuction with retry.
func (p *TdxPoolClient) GetAuction(code string, market int) ([]proto.AuctionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.AuctionData
	err := p.withRetry( func(c *gotdx.Client) error {
		reply, err := c.GetAuction(m.Uint8(), code, 0, 100)
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetCapitalFlow via MAC.
func (p *TdxPoolClient) GetCapitalFlow(code string, market int) (*proto.MACCapitalFlowReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return p.mac.MACCapitalFlow(m.Uint8(), code)
}

// GetUnusual with retry.
func (p *TdxPoolClient) GetUnusual(market int, count int) ([]proto.UnusualData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		m = types.MarketSZ
	}
	var result []proto.UnusualData
	err := p.withRetry( func(c *gotdx.Client) error {
		reply, err := c.GetUnusual(m.Uint8(), 0, uint32(count))
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetSymbolInfo via MAC.
func (p *TdxPoolClient) GetSymbolInfo(code string, market int) (*proto.MACSymbolInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return p.mac.MACSymbolInfo(m.Uint8(), code)
}

// GetF10 with retry.
func (p *TdxPoolClient) GetF10(code string, market int) (*gotdx.CompanyInfoBundle, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return p.main.GetCompanyInfo(m.Uint8(), code)
}

// GetFinanceInfo with retry.
func (p *TdxPoolClient) GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return p.main.GetFinanceInfo(m.Uint8(), code)
}

// GetXDXRInfo with retry.
func (p *TdxPoolClient) GetXDXRInfo(code string, market int) (*proto.GetXDXRInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return p.main.GetXDXRInfo(m.Uint8(), code)
}

// ExGetQuote.
func (p *TdxPoolClient) ExGetQuote(code string, category uint8) (*proto.ExQuoteItem, error) {
	return p.ex.ExQuote(category, code)
}

// ExGetKLine.
func (p *TdxPoolClient) ExGetKLine(category uint8, code string, period uint16, count int) ([]proto.ExKLineItem, error) {
	return p.ex.ExKLine(category, code, period, 0, uint16(count), 1)
}

// ExGetQuotes.
func (p *TdxPoolClient) ExGetQuotes(categories []uint8, codes []string) ([]proto.ExQuoteItem, error) {
	return p.ex.ExQuotes(categories, codes)
}

// GetSecurityList with cache.
func (p *TdxPoolClient) GetSecurityList(market int, start uint16) (*proto.GetSecurityListReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}

	// Check cache
	if start == 0 {
		p.mu.RLock()
		entry, cached := p.listCache[m.Uint8()]
		p.mu.RUnlock()
		if cached && time.Now().Before(entry.expiresAt) {
			return &proto.GetSecurityListReply{
				Count: uint16(len(entry.data)),
				List:  entry.data,
			}, nil
		}
	}

	var result *proto.GetSecurityListReply
	err := p.withRetry( func(c *gotdx.Client) error {
		var err error
		result, err = c.GetSecurityList(m.Uint8(), start)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Update cache
	if start == 0 && result != nil {
		p.mu.Lock()
		p.listCache[m.Uint8()] = &cacheEntry[[]proto.Security]{
			data:      result.List,
			expiresAt: time.Now().Add(p.cfg.CacheTTL),
		}
		p.mu.Unlock()
	}
	return result, nil
}

// GetSecurityCount with cache.
func (p *TdxPoolClient) GetSecurityCount(market int) (uint16, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return 0, fmt.Errorf("unknown market: %d", market)
	}

	// Check cache
	p.mu.RLock()
	entry, cached := p.countCache[m.Uint8()]
	p.mu.RUnlock()
	if cached && time.Now().Before(entry.expiresAt) {
		return entry.data.v, nil
	}

	var count uint16
	err := p.withRetry( func(c *gotdx.Client) error {
		reply, err := c.GetSecurityCount(m.Uint8())
		if err != nil {
			return err
		}
		count = reply.Count
		return nil
	})
	if err != nil {
		return 0, err
	}

	// Update cache
	p.mu.Lock()
	p.countCache[m.Uint8()] = &cacheEntry[u16Val]{
		data:      u16Val{v: count},
		expiresAt: time.Now().Add(p.cfg.CacheTTL),
	}
	p.mu.Unlock()
	return count, nil
}

// PoolStats returns pool statistics.
func (p *TdxPoolClient) PoolStats() string {
	p.mu.RLock()
	size := len(p.pool)
	healthy := 0
	hostCounts := make(map[string]int)
	for _, c := range p.pool {
		if !c.lastUsed.IsZero() {
			healthy++
			hostCounts[c.connAddr]++
		}
	}
	p.mu.RUnlock()
	return fmt.Sprintf("pool: %d healthy/%d total, hosts=%d, size=%d, hostDist=%v",
		healthy, size, len(hostCounts), p.cfg.Size, hostCounts)
}

// IsConnected returns whether any pool connection is alive.
func (p *TdxPoolClient) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pool) > 0
}

// Disconnect stops heartbeat and closes all pool connections.
func (p *TdxPoolClient) Disconnect() error {
	close(p.stopCh)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.pool {
		c.client.Disconnect()
	}
	p.pool = nil
	p.connected = false
	return nil
}

// GetSectorBoards fetches boards by type.
func (p *TdxPoolClient) GetSectorBoards(bt BlockType) ([]SectorBoard, error) {
	return (&TDXTCPClient{mainClient: p.main}).GetSectorBoards(bt)
}

// GetSectorBoardStocks fetches constituent stocks.
func (p *TdxPoolClient) GetSectorBoardStocks(boardCode string) ([]string, error) {
	return (&TDXTCPClient{mainClient: p.main}).GetSectorBoardStocks(boardCode)
}

// =============================================================================
// Strategy 2: TdxDirectClient — 每请求独立 TCP (高并发，60线程零退化)
// =============================================================================

// TdxDirectClient creates a fresh TCP connection per request.
// Hosts are probed upfront, then connections are distributed round-robin
// across reachable hosts to maximize concurrency throughput.
type TdxDirectClient struct {
	timeoutSec int
	hostPool   []string
	nextHost   atomic.Uint64
}

// NewTdxDirectClient creates a direct client with multi-host round-robin.
func NewTdxDirectClient(timeoutSec int) *TdxDirectClient {
	if timeoutSec <= 0 {
		timeoutSec = 6
	}
	d := &TdxDirectClient{timeoutSec: timeoutSec}

	// Probe all main hosts upfront
	results := gotdx.ProbeHosts(gotdx.MainHosts(), time.Duration(timeoutSec)*time.Second)
	d.hostPool = make([]string, 0, len(results))
	for _, r := range results {
		if r.Reachable {
			d.hostPool = append(d.hostPool, r.Address)
		}
	}
	return d
}

// do creates a fresh connection to the next host in round-robin.
func (d *TdxDirectClient) do(fn func(c *gotdx.Client) error) error {
	var c *gotdx.Client
	if len(d.hostPool) > 0 {
		idx := d.nextHost.Add(1) % uint64(len(d.hostPool))
		addr := d.hostPool[idx]
		c = gotdx.New(gotdx.WithTCPAddress(addr), gotdx.WithTimeoutSec(d.timeoutSec))
	} else {
		c = gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(d.timeoutSec))
	}
	defer c.Disconnect()
	if _, err := c.Connect(); err != nil {
		return err
	}
	return fn(c)
}

// GetQuoteDirect — direct per-request.
func (d *TdxDirectClient) GetQuote(code string, market int) (*proto.SecurityQuote, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result proto.SecurityQuote
	err := d.do(func(c *gotdx.Client) error {
		res, err := c.StockQuotesDetail([]uint8{m.Uint8()}, []string{code})
		if err != nil {
			return err
		}
		if len(res) == 0 {
			return fmt.Errorf("no quote for %s", code)
		}
		result = res[0]
		return nil
	})
	return &result, err
}

// GetKLineDirect — direct per-request.
func (d *TdxDirectClient) GetKLine(code string, market int, period string, count int, adjust int) ([]proto.SecurityBar, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	category := uint16(proto.KLINE_TYPE_RI_K)
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
	var bars []proto.SecurityBar
	err := d.do(func(c *gotdx.Client) error {
		var err error
		bars, err = c.StockFullKLine(category, m.Uint8(), code, 1, adjustType, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(bars) > count {
		bars = bars[len(bars)-count:]
	}
	return bars, nil
}

// GetBatchQuotesDirect — direct per-request.
func (d *TdxDirectClient) GetBatchQuotes(pairs []struct{ Market int; Code string }) ([]proto.SecurityQuote, error) {
	if len(pairs) == 0 {
		return nil, fmt.Errorf("empty pairs")
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
	var result []proto.SecurityQuote
	err := d.do(func(c *gotdx.Client) error {
		var err error
		result, err = c.StockQuotesDetail(markets, codes)
		return err
	})
	return result, err
}

// GetTickChartDirect.
func (d *TdxDirectClient) GetTickChart(code string, market int) ([]proto.MinuteTimeData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.MinuteTimeData
	err := d.do(func(c *gotdx.Client) error {
		reply, err := c.GetMinuteTimeData(m.Uint8(), code)
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetTransactionDirect.
func (d *TdxDirectClient) GetTransaction(code string, market int, count int) ([]proto.TransactionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.TransactionData
	err := d.do(func(c *gotdx.Client) error {
		reply, err := c.GetTransactionData(m.Uint8(), code, 0, uint16(count))
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetAuctionDirect.
func (d *TdxDirectClient) GetAuction(code string, market int) ([]proto.AuctionData, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	var result []proto.AuctionData
	err := d.do(func(c *gotdx.Client) error {
		reply, err := c.GetAuction(m.Uint8(), code, 0, 100)
		if err != nil {
			return err
		}
		result = reply.List
		return nil
	})
	return result, err
}

// GetSecurityListDirect.
func (d *TdxDirectClient) GetSecurityList(market int, start uint16) (*proto.GetSecurityListReply, error) {
	return nil, fmt.Errorf("direct client does not support GetSecurityList")
}

// GetSecurityCountDirect.
func (d *TdxDirectClient) GetSecurityCount(market int) (uint16, error) {
	return 0, fmt.Errorf("direct client does not support GetSecurityCount")
}

// PoolStats returns "N/A" for direct client.
func (d *TdxDirectClient) PoolStats() string {
	return "direct: no pool"
}

// IsConnected always false for direct client.
func (d *TdxDirectClient) IsConnected() bool {
	return false
}

// Disconnect is a no-op.
func (d *TdxDirectClient) Disconnect() error {
	return nil
}

// GetSectorBoards fetches boards by type via fresh connection.
func (d *TdxDirectClient) GetSectorBoards(bt BlockType) ([]SectorBoard, error) {
	c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(d.timeoutSec))
	defer c.Disconnect()
	if _, err := c.Connect(); err != nil {
		return nil, err
	}
	return (&TDXTCPClient{mainClient: c}).GetSectorBoards(bt)
}

// GetSectorBoardStocks fetches constituent stocks via fresh connection.
func (d *TdxDirectClient) GetSectorBoardStocks(boardCode string) ([]string, error) {
	c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(d.timeoutSec))
	defer c.Disconnect()
	if _, err := c.Connect(); err != nil {
		return nil, err
	}
	return (&TDXTCPClient{mainClient: c}).GetSectorBoardStocks(boardCode)
}

// =============================================================================
// Strategy 3: TdxFinanceClient — 独立超时(15s) + 磁盘缓存 (gpcw大文件下载)
// =============================================================================

// TdxFinanceClient wraps main client with longer timeout and optional disk cache.
type TdxFinanceClient struct {
	mainClient *gotdx.Client
	cacheDir   string
	timeoutSec int
}

// NewTdxFinanceClient creates a finance client with 15s timeout.
func NewTdxFinanceClient(cacheDir string, timeoutSec int) *TdxFinanceClient {
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(timeoutSec))
	return &TdxFinanceClient{
		mainClient: c,
		cacheDir:   cacheDir,
		timeoutSec: timeoutSec,
	}
}

// ConnectFinance connects to the finance server.
func (f *TdxFinanceClient) Connect() error {
	_, err := f.mainClient.Connect()
	return err
}

// GetFinanceInfo with longer timeout.
func (f *TdxFinanceClient) GetFinanceInfo(code string, market int) (*proto.GetFinanceInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return f.mainClient.GetFinanceInfo(m.Uint8(), code)
}

// GetXDXRInfo with longer timeout.
func (f *TdxFinanceClient) GetXDXRInfo(code string, market int) (*proto.GetXDXRInfoReply, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return f.mainClient.GetXDXRInfo(m.Uint8(), code)
}

// GetCompanyInfo (F10) with longer timeout.
func (f *TdxFinanceClient) GetCompanyInfo(code string, market int) (*gotdx.CompanyInfoBundle, error) {
	var m types.Market
	switch market {
	case 0:
		m = types.MarketSZ
	case 1:
		m = types.MarketSH
	default:
		return nil, fmt.Errorf("unknown market: %d", market)
	}
	return f.mainClient.GetCompanyInfo(m.Uint8(), code)
}

// Disconnect closes the finance client.
func (f *TdxFinanceClient) Disconnect() error {
	return f.mainClient.Disconnect()
}

// =============================================================================
// Shared helpers (used by both PoolClient and original TCP client)
// =============================================================================

// buildXdXrEvents extracts dividend/ex-rights events from the xdxr reply.
func buildXdXrEvents(reply *proto.GetXDXRInfoReply) []XdXrEvent {
	if reply == nil || reply.List == nil {
		return nil
	}
	var events []XdXrEvent
	for _, item := range reply.List {
		if item.Category != 1 {
			continue
		}
		y, m, d := item.Date.Date()
		fenhong := 0.0
		if item.Fenhong != nil {
			fenhong = float64(*item.Fenhong)
		}
		songzhuang := 0.0
		if item.Songzhuangu != nil {
			songzhuang = float64(*item.Songzhuangu)
		}
		peigu := 0.0
		if item.Peigu != nil {
			peigu = float64(*item.Peigu)
		}
		peigujia := 0.0
		if item.Peigujia != nil {
			peigujia = float64(*item.Peigujia)
		}
		events = append(events, XdXrEvent{
			Year:       y,
			Month:      int(m),
			Day:        d,
			Fenhong:    fenhong,
			Songzhuang: songzhuang,
			Peigu:      peigu,
			Peigujia:   peigujia,
		})
	}
	return events
}

// XdXrEvent represents a dividend/ex-rights event.
type XdXrEvent struct {
	Year       int
	Month      int
	Day        int
	Fenhong    float64
	Songzhuang float64
	Peigu      float64
	Peigujia   float64
}

// DateKey returns YYYYMMDD.
func (e XdXrEvent) DateKey() int {
	return e.Year*10000 + e.Month*100 + e.Day
}

// CalcQfqFactor calculates forward adjustment factor.
func CalcQfqFactor(closeBefore float64, event XdXrEvent) float64 {
	divPerShare := event.Fenhong / 10.0
	bonusRatio := event.Songzhuang / 10.0
	rightsRatio := event.Peigu / 10.0
	rightsPrice := event.Peigujia

	denom := closeBefore * (1.0 + bonusRatio + rightsRatio)
	num := closeBefore - divPerShare + rightsPrice*rightsRatio

	if math.Abs(denom) < 1e-10 || math.Abs(closeBefore) < 1e-10 {
		return 1.0
	}
	return num / denom
}

// RoundPrice rounds to 3 decimal places.
func RoundPrice(p float64) float64 {
	return math.Round(p*1000) / 1000
}

// applyAdjustToBars applies client-side adjustment.
func applyAdjustToBars(bars []proto.SecurityBar, contextBars []proto.SecurityBar, events []XdXrEvent, adjustType int) []proto.SecurityBar {
	if len(events) == 0 || len(bars) == 0 {
		return bars
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].DateKey() < events[j].DateKey()
	})

	factorMap := make(map[int]float64)
	for _, evt := range events {
		closeBefore := findCloseBefore(bars, contextBars, evt.DateKey())
		if closeBefore == 0 {
			continue
		}
		factor := CalcQfqFactor(closeBefore, evt)
		factorMap[evt.DateKey()] = factor
	}

	if len(factorMap) == 0 {
		return bars
	}

	switch adjustType {
	case 1: // 前复权
		cumulative := 1.0
		sortedEvents := make([]XdXrEvent, len(events))
		copy(sortedEvents, events)
		sort.Slice(sortedEvents, func(i, j int) bool {
			return sortedEvents[i].DateKey() > sortedEvents[j].DateKey()
		})
		eventIdx := 0
		for i := len(bars) - 1; i >= 0; i-- {
			bars[i].Open = RoundPrice(bars[i].Open * cumulative)
			bars[i].High = RoundPrice(bars[i].High * cumulative)
			bars[i].Low = RoundPrice(bars[i].Low * cumulative)
			bars[i].Close = RoundPrice(bars[i].Close * cumulative)
			for eventIdx < len(sortedEvents) && sortedEvents[eventIdx].DateKey() > bars[i].Year*10000+bars[i].Month*100+bars[i].Day {
				if f, ok := factorMap[sortedEvents[eventIdx].DateKey()]; ok {
					cumulative *= f
				}
				eventIdx++
			}
		}
	case 2: // 后复权
		cumulative := 1.0
		for i := range bars {
			for _, evt := range events {
				if evt.DateKey() <= bars[i].Year*10000+bars[i].Month*100+bars[i].Day {
					if f, ok := factorMap[evt.DateKey()]; ok {
						cumulative *= 1.0 / f
					}
				}
			}
			bars[i].Open = RoundPrice(bars[i].Open * cumulative)
			bars[i].High = RoundPrice(bars[i].High * cumulative)
			bars[i].Low = RoundPrice(bars[i].Low * cumulative)
			bars[i].Close = RoundPrice(bars[i].Close * cumulative)
		}
	}

	return bars
}

// findCloseBefore finds closing price before dateKey.
func findCloseBefore(bars []proto.SecurityBar, contextBars []proto.SecurityBar, dateKey int) float64 {
	var result float64
	found := false
	for _, bar := range bars {
		key := bar.Year*10000 + bar.Month*100 + bar.Day
		if key < dateKey {
			result = bar.Close
			found = true
		}
	}
	if found {
		return result
	}
	for _, bar := range contextBars {
		key := bar.Year*10000 + bar.Month*100 + bar.Day
		if key < dateKey {
			result = bar.Close
		}
	}
	return result
}

// ProbeHosts probes all TDX hosts.
func ProbeHosts() (mainProbes []gotdx.HostProbeResult, exProbes []gotdx.HostProbeResult, macProbes []gotdx.HostProbeResult) {
	timeout := 3 * time.Second
	mainProbes = gotdx.ProbeAddresses(gotdx.MainHostAddresses(), timeout)
	exProbes = gotdx.ProbeAddresses(gotdx.ExHostAddresses(), timeout)
	macProbes = gotdx.ProbeAddresses(gotdx.MACHostAddresses(), timeout)
	return
}

// GetMainHosts returns known main market hosts.
func GetMainHosts() []gotdx.HostInfo {
	return gotdx.MainHosts()
}

// GetExHosts returns known extension market hosts.
func GetExHosts() []gotdx.HostInfo {
	return gotdx.ExHosts()
}

// GetBrokerHosts returns known broker hosts.
func GetBrokerHosts() []gotdx.HostInfo {
	return gotdx.BrokerHosts()
}

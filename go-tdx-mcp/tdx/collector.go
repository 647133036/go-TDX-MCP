package tdx

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bensema/gotdx"
	"github.com/bensema/gotdx/proto"
)

// =============================================================================
// MultiHostCollector — 多主机并发数据采集器
// =============================================================================

// hostConn wraps a gotdx client with mutex for goroutine safety.
type hostConn struct {
	addr   string
	client *gotdx.Client
	mu     sync.Mutex
	alive  atomic.Bool
	reqs   atomic.Int64
}

// CollectStats tracks collection progress and results.
type CollectStats struct {
	Total     int64
	Success   atomic.Int64
	Fail      atomic.Int64
	Elapsed   time.Duration   // set after collection
	Latencies []time.Duration // set after collection
}

// Summary returns a formatted stats string.
func (s *CollectStats) Summary() string {
	success := s.Success.Load()
	fail := s.Fail.Load()
	total := success + fail
	if total == 0 {
		return "no data"
	}
	rate := float64(success) / float64(total) * 100
	throughput := float64(success) / s.Elapsed.Seconds()
	return fmt.Sprintf("%d/%d (%.1f%%), %.1f req/s, elapsed=%v",
		success, total, rate, throughput, s.Elapsed.Round(time.Millisecond))
}

// Percentile returns the nth percentile from sorted latencies.
func (s *CollectStats) Percentile(p int) time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	n := len(s.Latencies) * p / 100
	if n >= len(s.Latencies) {
		n = len(s.Latencies) - 1
	}
	return s.Latencies[n]
}

// MultiHostCollector manages connections to multiple TDX main hosts.
type MultiHostCollector struct {
	conns   []*hostConn
	nextIdx atomic.Uint64
	cfg     CollectorConfig
	stopCh  chan struct{}
	healWg  sync.WaitGroup
}

// CollectorConfig configures the collector.
type CollectorConfig struct {
	HostTimeout     time.Duration // per-host connect timeout, default 6s
	MaxConnsPerHost int           // connections per host, default 1
	MaxHosts        int           // max number of hosts to connect, 0 = all
	RetryCount      int           // retries per request, default 2
	RetryDelay      time.Duration // delay between retries, default 200ms
	DetailedBatchSize int         // batch size for detailed quotes, default 20 (32KB limit safe)
}

// DefaultCollectorConfig returns sensible defaults.
func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		HostTimeout:      6 * time.Second,
		MaxConnsPerHost:  1,
		MaxHosts:         0,
		RetryCount:       2,
		RetryDelay:       200 * time.Millisecond,
		DetailedBatchSize: 20,
	}
}

// NewMultiHostCollector creates a collector connected to all reachable TDX hosts.
func NewMultiHostCollector(cfg CollectorConfig) (*MultiHostCollector, error) {
	if cfg.HostTimeout <= 0 {
		cfg.HostTimeout = 6 * time.Second
	}
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}

	c := &MultiHostCollector{
		conns: make([]*hostConn, 0),
		cfg:   cfg,
	}

	results := gotdx.ProbeHosts(gotdx.MainHosts(), cfg.HostTimeout)
	reachable := make([]gotdx.HostProbeResult, 0)
	for _, r := range results {
		if r.Reachable {
			reachable = append(reachable, r)
		}
	}

	if len(reachable) == 0 {
		return nil, fmt.Errorf("no reachable hosts")
	}

	maxHosts := cfg.MaxHosts
	if maxHosts <= 0 || maxHosts > len(reachable) {
		maxHosts = len(reachable)
	}
	reachable = reachable[:maxHosts]

	connsPerHost := cfg.MaxConnsPerHost
	if connsPerHost <= 0 {
		connsPerHost = 1
	}

	// Concurrent connection establishment
	type connResult struct {
		conn *hostConn
		err  error
	}
	resultsCh := make(chan connResult, maxHosts*connsPerHost)

	for _, host := range reachable {
		for j := 0; j < connsPerHost; j++ {
			go func(addr string) {
				gc := gotdx.New(gotdx.WithTCPAddress(addr), gotdx.WithTimeoutSec(int(cfg.HostTimeout.Seconds())))
				if _, err := gc.Connect(); err != nil {
					resultsCh <- connResult{err: fmt.Errorf("connect %s: %w", addr, err)}
					return
				}
				hc := &hostConn{addr: addr, client: gc}
				hc.alive.Store(true)
				resultsCh <- connResult{conn: hc}
			}(host.Address)
		}
	}

	total := maxHosts * connsPerHost
	for i := 0; i < total; i++ {
		res := <-resultsCh
		if res.conn != nil {
			c.conns = append(c.conns, res.conn)
		} else {
			log.Printf("collector: %v", res.err)
		}
	}

	if len(c.conns) == 0 {
		return nil, fmt.Errorf("failed to connect to any host")
	}

	log.Printf("collector: %d hosts x %d conns = %d total connections",
		len(reachable), connsPerHost, len(c.conns))

	// Start connection healing goroutine
	c.stopCh = make(chan struct{})
	c.healWg.Add(1)
	go c.healLoop()

	return c, nil
}

// healLoop periodically reconnects dead connections.
func (c *MultiHostCollector) healLoop() {
	defer c.healWg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
		}
		c.heal()
	}
}

// heal attempts to reconnect all dead connections.
func (c *MultiHostCollector) heal() {
	for _, conn := range c.conns {
		if conn.alive.Load() {
			continue
		}
		conn.mu.Lock()
		if conn.alive.Load() {
			conn.mu.Unlock()
			continue
		}
		newGC := gotdx.New(gotdx.WithTCPAddress(conn.addr), gotdx.WithTimeoutSec(int(c.cfg.HostTimeout.Seconds())))
		if _, err := newGC.Connect(); err != nil {
			conn.mu.Unlock()
			continue
		}
		conn.client.Disconnect()
		conn.client = newGC
		conn.alive.Store(true)
		conn.mu.Unlock()
		log.Printf("collector: healed connection to %s", conn.addr)
	}
}

func (c *MultiHostCollector) borrow() *hostConn {
	if len(c.conns) == 0 {
		return nil
	}
	idx := int(c.nextIdx.Add(1) % uint64(len(c.conns)))
	conn := c.conns[idx]
	conn.mu.Lock()
	return conn
}

func (c *MultiHostCollector) release(conn *hostConn) {
	conn.mu.Unlock()
}

func (c *MultiHostCollector) do(fn func(gc *gotdx.Client) error) error {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.RetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(c.cfg.RetryDelay)
		}
		conn := c.borrow()
		if conn == nil {
			return fmt.Errorf("no connections available")
		}
		if !conn.alive.Load() {
			// Dead connection: reconnect without releasing lock
			newGC := gotdx.New(gotdx.WithTCPAddress(conn.addr), gotdx.WithTimeoutSec(int(c.cfg.HostTimeout.Seconds())))
			if _, err := newGC.Connect(); err != nil {
				c.release(conn)
				lastErr = err
				continue
			}
			conn.client.Disconnect()
			conn.client = newGC
			conn.alive.Store(true)
		}
		err := fn(conn.client)
		if err == nil {
			conn.reqs.Add(1)
			c.release(conn)
			return nil
		}
		if isDataError(err) {
			c.release(conn)
			return err
		}
		conn.alive.Store(false)
		lastErr = err
		c.release(conn)
	}
	return fmt.Errorf("failed after %d retries: %w", c.cfg.RetryCount+1, lastErr)
}

// isDataError returns true for errors that are NOT transient network issues.
func isDataError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch {
	case len(msg) > 7 && msg[:7] == "invalid":
		return true // invalid block data length, invalid response, etc.
	case len(msg) >= 5 && (msg[:5] == "parse" || msg[:5] == "unmar"):
		return true // parse errors
	case len(msg) >= 4 && msg[:4] == "zero":
		return true // zero-length data
	case msg == "more than 8M data":
		return true // gotdx internal 32KB packet limit exceeded
	}
	return false
}

// =============================================================================
// Data Collection Methods
// =============================================================================

const (
	MarketSZ = 0
	MarketSH = 1
)

// CollectAllQuotes fetches quotes for all securities in a market concurrently.
func (c *MultiHostCollector) CollectAllQuotes(market int) ([]proto.QuoteListItem, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}

	// Use optimized batch method — StockQuotesList fetches top N in one request
	mkt := marketToUint8(market)

	stats := &CollectStats{}
	start := time.Now()

	var quotes []proto.QuoteListItem
	err := c.do(func(gc *gotdx.Client) error {
		var e error
		quotes, e = gc.StockQuotesList(mkt, 0, 5000, 0, false, 0)
		return e
	})

	stats.Elapsed = time.Since(start)
	if err != nil {
		stats.Fail.Store(1)
		return nil, stats, err
	}
	stats.Total = int64(len(quotes))
	stats.Success.Store(int64(len(quotes)))
	return quotes, stats, nil
}

// CollectTopQuotes fetches top N quotes sorted by a given field (optimized batch).
func (c *MultiHostCollector) CollectTopQuotes(market int, count uint16, sortType uint16, reverse bool) ([]proto.QuoteListItem, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}

	mkt := marketToUint8(market)
	stats := &CollectStats{Total: int64(count)}
	start := time.Now()

	var quotes []proto.QuoteListItem
	err := c.do(func(gc *gotdx.Client) error {
		var e error
		quotes, e = gc.StockQuotesList(mkt, 0, count, sortType, reverse, 0)
		return e
	})

	stats.Elapsed = time.Since(start)
	if err != nil {
		stats.Fail.Store(int64(count))
		return nil, stats, err
	}
	stats.Success.Store(int64(len(quotes)))
	return quotes, stats, nil
}

// CollectDetailedQuotes fetches detailed quotes for specific codes.
func (c *MultiHostCollector) CollectDetailedQuotes(codes []string, market int, batchSize int) ([]proto.SecurityQuote, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}
	if batchSize <= 0 {
		batchSize = c.cfg.DetailedBatchSize
		if batchSize <= 0 {
			batchSize = 20
		}
	}

	stats := &CollectStats{Total: int64(len(codes))}
	start := time.Now()

	var mu sync.Mutex
	results := make([]proto.SecurityQuote, 0, len(codes))
	var latencies []time.Duration

	currentBatch := batchSize
	for i := 0; i < len(codes); i += currentBatch {
		end := i + currentBatch
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]
		markets := make([]uint8, len(batch))
		mkt := marketToUint8(market)
		for j := range markets {
			markets[j] = mkt
		}

		var batchQuotes []proto.SecurityQuote
		err := c.do(func(gc *gotdx.Client) error {
			var e error
			batchQuotes, e = gc.StockQuotesDetail(markets, batch)
			return e
		})
		mu.Lock()
		if err != nil {
			// Auto-fallback: if batch failed, try with half size
			if currentBatch > 1 && len(batch) > 1 {
				currentBatch = currentBatch / 2
				if currentBatch < 1 {
					currentBatch = 1
				}
				i -= batchSize - currentBatch
				mu.Unlock()
				continue
			}
			stats.Fail.Add(int64(len(batch)))
		} else {
			results = append(results, batchQuotes...)
			for k := 0; k < len(batchQuotes); k++ {
				latencies = append(latencies, time.Since(start))
			}
			stats.Success.Add(int64(len(batchQuotes)))
		}
		mu.Unlock()
	}

	stats.Elapsed = time.Since(start)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	stats.Latencies = latencies

	return results, stats, nil
}

// CollectKLines fetches K-line data for multiple codes concurrently.
func (c *MultiHostCollector) CollectKLines(codes []string, market int, periodCategory uint16, count uint16, adjust uint16) (map[string][]proto.SecurityBar, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}

	stats := &CollectStats{Total: int64(len(codes))}
	start := time.Now()

	var mu sync.Mutex
	result := make(map[string][]proto.SecurityBar, len(codes))
	var latencies []time.Duration
	var wg sync.WaitGroup

	mkt := marketToUint8(market)

	for _, code := range codes {
		wg.Add(1)
		go func(cd string) {
			defer wg.Done()
			var bars []proto.SecurityBar
			err := c.do(func(gc *gotdx.Client) error {
				var e error
				bars, e = gc.StockKLine(periodCategory, mkt, cd, 0, count, 1, adjust)
				return e
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				stats.Fail.Add(1)
				return
			}
			result[cd] = bars
			latencies = append(latencies, time.Since(start))
			stats.Success.Add(1)
		}(code)
	}
	wg.Wait()

	stats.Elapsed = time.Since(start)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	stats.Latencies = latencies

	return result, stats, nil
}

// CollectSectorBoards fetches sector boards for a given block type.
func (c *MultiHostCollector) CollectSectorBoards(bt BlockType) ([]SectorBoard, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}

	stats := &CollectStats{Total: 1}
	start := time.Now()

	var boards []SectorBoard
	err := c.do(func(gc *gotdx.Client) error {
		data, e := gc.GetBlockFile(bt.BlockFilename())
		if e != nil {
			return e
		}
		if len(data) == 0 {
			return fmt.Errorf("empty block file %s (server returned no data)", bt.BlockFilename())
		}
		groups, e := gotdx.ParseBlockGroups(data)
		if e != nil {
			return e
		}
		boards = make([]SectorBoard, 0, len(groups))
		for _, g := range groups {
			boards = append(boards, SectorBoard{
				Code:     g.BlockName,
				Name:     g.BlockName,
				Type:     bt.BlockName(),
				StockCnt: g.StockCount,
			})
		}
		return nil
	})

	stats.Elapsed = time.Since(start)
	if err != nil {
		stats.Fail.Add(1)
		return nil, stats, err
	}
	stats.Total = int64(len(boards))
	stats.Success.Store(int64(len(boards)))
	return boards, stats, nil
}

// CollectSectorBoardStocks fetches constituent stocks for boards concurrently.
func (c *MultiHostCollector) CollectSectorBoardStocks(boards []SectorBoard) (map[string][]string, *CollectStats, error) {
	if c == nil || len(c.conns) == 0 {
		return nil, nil, fmt.Errorf("collector not initialized")
	}

	stats := &CollectStats{Total: int64(len(boards))}
	start := time.Now()

	var mu sync.Mutex
	result := make(map[string][]string, len(boards))
	var latencies []time.Duration
	var wg sync.WaitGroup

	// block filenames to scan (same as GetSectorBoardStocks in tcp_client.go)
	blockFiles := []string{"block_gy.dat", "block_gn.dat", "block_dy.dat", "block_zs.dat", "block_zc.dat", "block_zdy.dat"}

	for _, board := range boards {
		wg.Add(1)
		go func(b SectorBoard) {
			defer wg.Done()
			var stocks []string

			for _, filename := range blockFiles {
				err := c.do(func(gc *gotdx.Client) error {
					data, e := gc.GetBlockFile(filename)
					if e != nil {
						return e
					}
					groups, e := gotdx.ParseBlockGroups(data)
					if e != nil {
						return e
					}
					for _, g := range groups {
						if g.BlockName == b.Code || g.BlockName == b.Name {
							stocks = g.Codes
							return nil
						}
					}
					return fmt.Errorf("board %s not found in %s", b.Code, filename)
				})
				if err == nil && len(stocks) > 0 {
					break
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if len(stocks) > 0 {
				result[b.Code] = stocks
				latencies = append(latencies, time.Since(start))
				stats.Success.Add(1)
			} else {
				stats.Fail.Add(1)
			}
		}(board)
	}
	wg.Wait()

	stats.Elapsed = time.Since(start)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	stats.Latencies = latencies

	return result, stats, nil
}

func marketToUint8(market int) uint8 {
	switch market {
	case MarketSZ:
		return 0 // proto.MarketSZ
	case MarketSH:
		return 1 // proto.MarketSH
	default:
		return 1
	}
}

// PoolStats returns connection pool statistics.
func (c *MultiHostCollector) PoolStats() string {
	if c == nil || len(c.conns) == 0 {
		return "collector: no connections"
	}
	alive := 0
	totalReqs := int64(0)
	for _, conn := range c.conns {
		if conn.alive.Load() {
			alive++
		}
		totalReqs += conn.reqs.Load()
	}
	return fmt.Sprintf("collector: %d alive/%d total conns, %d total requests",
		alive, len(c.conns), totalReqs)
}

// Disconnect closes all connections.
func (c *MultiHostCollector) Disconnect() error {
	if c == nil {
		return nil
	}
	// Stop healing goroutine
	if c.stopCh != nil {
		close(c.stopCh)
		c.healWg.Wait()
	}
	for _, conn := range c.conns {
		conn.mu.Lock()
		conn.client.Disconnect()
		conn.alive.Store(false)
		conn.mu.Unlock()
	}
	return nil
}

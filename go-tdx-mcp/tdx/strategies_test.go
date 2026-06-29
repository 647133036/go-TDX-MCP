package tdx

import (
	"context"
	"testing"
	"time"

	"github.com/bensema/gotdx"
	"github.com/bensema/gotdx/proto"
)

// TestCalcQfqFactor verifies forward adjustment factor calculation.
func TestCalcQfqFactor(t *testing.T) {
	tests := []struct {
		name      string
		close     float64
		event     XdXrEvent
		expectMin float64
		expectMax float64
	}{
		{
			name:      "pure_dividend",
			close:     10.0,
			event:     XdXrEvent{Fenhong: 0.5}, // 每10股派0.5元
			expectMin: 0.99,
			expectMax: 1.0,
		},
		{
			name:      "pure_bonus",
			close:     10.0,
			event:     XdXrEvent{Songzhuang: 3.0}, // 每10股送3股
			expectMin: 0.75,
			expectMax: 0.8,
		},
		{
			name:      "rights_issue",
			close:     10.0,
			event:     XdXrEvent{Peigu: 3.0, Peigujia: 5.0}, // 每10股配3股,配股价5元
			expectMin: 0.85,
			expectMax: 0.95,
		},
		{
			name:      "combined",
			close:     20.0,
			event:     XdXrEvent{Fenhong: 1.0, Songzhuang: 5.0, Peigu: 2.0, Peigujia: 8.0},
			expectMin: 0.5,
			expectMax: 0.7,
		},
		{
			name:      "zero_close_returns_1",
			close:     0.0,
			event:     XdXrEvent{Fenhong: 1.0},
			expectMin: 1.0,
			expectMax: 1.0,
		},
		{
			name:      "tiny_close_returns_1",
			close:     1e-11,
			event:     XdXrEvent{Fenhong: 1.0},
			expectMin: 1.0,
			expectMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factor := CalcQfqFactor(tt.close, tt.event)
			if factor < tt.expectMin || factor > tt.expectMax {
				t.Errorf("CalcQfqFactor(%v, %+v) = %v, expected [%v, %v]",
					tt.close, tt.event, factor, tt.expectMin, tt.expectMax)
			}
		})
	}
}

// TestRoundPrice verifies price rounding.
func TestRoundPrice(t *testing.T) {
	tests := []struct {
		input  float64
		expect float64
	}{
		{10.1234, 10.123},
		{10.1235, 10.124},
		{0.0, 0.0},
		{-5.6789, -5.679},
	}

	for _, tt := range tests {
		result := RoundPrice(tt.input)
		if result != tt.expect {
			t.Errorf("RoundPrice(%v) = %v, expected %v", tt.input, result, tt.expect)
		}
	}
}

// TestApplyAdjustToBars_empty_events_returns_original.
func TestApplyAdjustToBars_empty_events(t *testing.T) {
	bars := []proto.SecurityBar{
		{Open: 10, High: 12, Low: 9, Close: 11, Year: 2024, Month: 1, Day: 1},
		{Open: 11, High: 13, Low: 10, Close: 12, Year: 2024, Month: 1, Day: 2},
	}
	result := applyAdjustToBars(bars, nil, nil, 1)
	if len(result) != len(bars) {
		t.Errorf("expected %d bars, got %d", len(bars), len(result))
	}
}

// TestApplyAdjustToBars_no_context_needed.
func TestApplyAdjustToBars_no_context(t *testing.T) {
	bars := []proto.SecurityBar{
		{Open: 10, High: 12, Low: 9, Close: 11, Year: 2024, Month: 1, Day: 1},
		{Open: 11, High: 13, Low: 10, Close: 12, Year: 2024, Month: 1, Day: 2},
	}
	events := []XdXrEvent{
		{Year: 2020, Month: 1, Day: 1, Fenhong: 0.5}, // Before all bars
	}
	result := applyAdjustToBars(bars, nil, events, 1)
	// Should have same length
	if len(result) != len(bars) {
		t.Errorf("expected %d bars, got %d", len(bars), len(result))
	}
}

// TestFindCloseBefore verifies date-based close price lookup.
func TestFindCloseBefore(t *testing.T) {
	bars := []proto.SecurityBar{
		{Close: 10.0, Year: 2024, Month: 1, Day: 1},
		{Close: 11.0, Year: 2024, Month: 1, Day: 2},
		{Close: 12.0, Year: 2024, Month: 1, Day: 3},
	}
	contextBars := []proto.SecurityBar{
		{Close: 9.0, Year: 2023, Month: 12, Day: 31},
	}

	tests := []struct {
		dateKey int
		expect  float64
	}{
		{20240102, 10.0}, // Close before Jan 2 is Jan 1
		{20240103, 11.0}, // Close before Jan 3 is Jan 2
		{20240104, 12.0}, // Close before Jan 4 is Jan 3
		{20240105, 12.0}, // Close before Jan 5 is Jan 3 (last available)
	}

	for _, tt := range tests {
		result := findCloseBefore(bars, contextBars, tt.dateKey)
		if result != tt.expect {
			t.Errorf("findCloseBefore(%d) = %v, expected %v", tt.dateKey, result, tt.expect)
		}
	}
}

// TestFindCloseBefore_fallback_to_context.
func TestFindCloseBefore_fallback(t *testing.T) {
	bars := []proto.SecurityBar{
		{Close: 10.0, Year: 2024, Month: 1, Day: 10},
	}
	contextBars := []proto.SecurityBar{
		{Close: 8.0, Year: 2024, Month: 1, Day: 1},
		{Close: 9.0, Year: 2024, Month: 1, Day: 5},
	}

	// Looking for close before Jan 5, not in bars but in context
	// findCloseBefore returns the last bar's close before dateKey
	result := findCloseBefore(bars, contextBars, 20240105)
	if result != 8.0 {
		t.Errorf("expected fallback close 8.0, got %v", result)
	}
}

// TestXdXrEvent_DateKey.
func TestXdXrEvent_DateKey(t *testing.T) {
	event := XdXrEvent{Year: 2024, Month: 6, Day: 15}
	key := event.DateKey()
	if key != 20240615 {
		t.Errorf("DateKey() = %d, expected 20240615", key)
	}
}

// BenchmarkCalcQfqFactor measures adjustment factor computation speed.
func BenchmarkCalcQfqFactor(b *testing.B) {
	event := XdXrEvent{Fenhong: 0.5, Songzhuang: 1.0, Peigu: 2.0, Peigujia: 8.0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcQfqFactor(15.0, event)
	}
}

// BenchmarkRoundPrice measures price rounding speed.
func BenchmarkRoundPrice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RoundPrice(float64(i) * 1.234567)
	}
}

// TestTdxPoolClient_ConfigDefaults verifies default config values.
func TestTdxPoolClient_ConfigDefaults(t *testing.T) {
	cfg := DefaultPoolConfig()
	if cfg.Size != 5 {
		t.Errorf("default Size = %d, expected 5", cfg.Size)
	}
	if cfg.ConnectTimeout != 6*time.Second {
		t.Errorf("default ConnectTimeout = %v, expected 6s", cfg.ConnectTimeout)
	}
	if cfg.HeartbeatInterval != 30*time.Second {
		t.Errorf("default HeartbeatInterval = %v, expected 30s", cfg.HeartbeatInterval)
	}
	if cfg.RetryCount != 3 {
		t.Errorf("default RetryCount = %d, expected 3", cfg.RetryCount)
	}
	if cfg.RetryDelay != 500*time.Millisecond {
		t.Errorf("default RetryDelay = %v, expected 500ms", cfg.RetryDelay)
	}
	if cfg.CacheTTL != 30*time.Second {
		t.Errorf("default CacheTTL = %v, expected 30s", cfg.CacheTTL)
	}
}

// TestTdxDirectClient_Creation verifies direct client creation.
func TestTdxDirectClient_Creation(t *testing.T) {
	c := NewTdxDirectClient(0)
	if c.timeoutSec != 6 {
		t.Errorf("default timeout = %d, expected 6", c.timeoutSec)
	}

	c2 := NewTdxDirectClient(10)
	if c2.timeoutSec != 10 {
		t.Errorf("custom timeout = %d, expected 10", c2.timeoutSec)
	}
}

// TestTdxFinanceClient_Creation verifies finance client creation.
func TestTdxFinanceClient_Creation(t *testing.T) {
	c := NewTdxFinanceClient("/tmp/tdx_cache", 0)
	if c.timeoutSec != 15 {
		t.Errorf("default finance timeout = %d, expected 15", c.timeoutSec)
	}

	c2 := NewTdxFinanceClient("/tmp/tdx_cache", 30)
	if c2.timeoutSec != 30 {
		t.Errorf("custom finance timeout = %d, expected 30", c2.timeoutSec)
	}
}

// TestAsyncClient_Creation verifies async client creation.
func TestAsyncClient_Creation(t *testing.T) {
	// We can't actually connect for this test, so just verify struct creation
	poolCfg := DefaultPoolConfig()
	poolCfg.Size = 1
	pool := &TdxPoolClient{
		cfg:        poolCfg,
		pool:       make([]*poolConn, 0, 1),
		countCache: make(map[uint8]*cacheEntry[u16Val]),
		listCache:  make(map[uint8]*cacheEntry[[]proto.Security]),
		stopCh:     make(chan struct{}),
	}
	direct := NewTdxDirectClient(6)

	async := NewTdxAsyncClient(pool, direct, 5)
	if async.maxActive != 5 {
		t.Errorf("maxActive = %d, expected 5", async.maxActive)
	}
	if cap(async.reqChan) != 100 {
		t.Errorf("reqChan capacity = %d, expected 100", cap(async.reqChan))
	}
	async.Close()
}

// TestBatchProcessor_Creation verifies batch processor creation.
func TestBatchProcessor_Creation(t *testing.T) {
	bp := NewBatchProcessor(10)
	if bp.total != 10 {
		t.Errorf("total = %d, expected 10", bp.total)
	}
	if bp.Completed() != 0 {
		t.Errorf("initial completed = %d, expected 0", bp.Completed())
	}
}

// BenchmarkAsyncClient_SubmitAndAwait measures async client throughput.
func BenchmarkAsyncClient_SubmitAndAwait(b *testing.B) {
	poolCfg := DefaultPoolConfig()
	poolCfg.Size = 1
	pool := &TdxPoolClient{
		cfg:        poolCfg,
		pool:       make([]*poolConn, 0, 1),
		countCache: make(map[uint8]*cacheEntry[u16Val]),
		listCache:  make(map[uint8]*cacheEntry[[]proto.Security]),
		stopCh:     make(chan struct{}),
		main:       gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6)),
	}
	direct := NewTdxDirectClient(6)
	async := NewTdxAsyncClient(pool, direct, 5)
	defer async.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := AsyncRequest{
			Type:   "quote",
			Code:   "000001",
			Market: 0,
			Result: make(chan AsyncResult, 1),
		}
		async.SubmitAndAwait(context.Background(), req)
	}
	b.StopTimer()
}

// BenchmarkBatchProcessor_Wait measures batch processor throughput.
func BenchmarkBatchProcessor_Wait(b *testing.B) {
	poolCfg := DefaultPoolConfig()
	poolCfg.Size = 1
	pool := &TdxPoolClient{
		cfg:        poolCfg,
		pool:       make([]*poolConn, 0, 1),
		countCache: make(map[uint8]*cacheEntry[u16Val]),
		listCache:  make(map[uint8]*cacheEntry[[]proto.Security]),
		stopCh:     make(chan struct{}),
		main:       gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6)),
	}
	direct := NewTdxDirectClient(6)
	async := NewTdxAsyncClient(pool, direct, 5)
	defer async.Close()

	batchSize := 10
	bp := NewBatchProcessor(batchSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			ch := make(chan AsyncResult, 1)
			req := AsyncRequest{
				Type:   "quote",
				Code:   "000001",
				Market: 0,
				Result: ch,
			}
			async.Submit(req)
			bp.AddRequest(j, req)
		}
		bp.Wait(context.Background())
	}
	b.StopTimer()
}

// BenchmarkHandleRequestRouting_switch measures pure switch-case dispatch overhead.
func BenchmarkHandleRequestRouting_switch(b *testing.B) {
	types := []string{"quote", "kline", "tick", "transaction", "auction", "f10", "finance", "xdxr"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(types)
		switch types[idx] {
		case "quote":
		case "kline":
		case "tick":
		case "transaction":
		case "auction":
		case "f10":
		case "finance":
		case "xdxr":
		}
	}
}

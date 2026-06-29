package tdx

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/bensema/gotdx"
)

// =============================================================================
// 1. 连接握手耗时分布测试
// =============================================================================

func BenchmarkConnectHandshake_100(b *testing.B) {
	benchmarkConnectDistribution(b, 100)
}

func BenchmarkConnectHandshake_1000(b *testing.B) {
	benchmarkConnectDistribution(b, 1000)
}

func benchmarkConnectDistribution(b *testing.B, iterations int) {
	times := make([]time.Duration, 0, iterations)
	for i := 0; i < iterations; i++ {
		c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
		start := time.Now()
		_, err := c.Connect()
		elapsed := time.Since(start)
		c.Disconnect()
		if err != nil {
			b.Logf("connect failed at iter %d: %v (elapsed=%v)", i, err, elapsed)
			continue
		}
		times = append(times, elapsed)
	}

	if len(times) == 0 {
		b.Skip("no successful connections")
		return
	}

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	n := len(times)
	sum := time.Duration(0)
	for _, t := range times {
		sum += t
	}
	mean := sum / time.Duration(n)

	vars := 0.0
	for _, t := range times {
		diff := float64(t-mean)
		vars += diff * diff
	}
	stddev := time.Duration(math.Sqrt(float64(vars) / float64(n)))

	p50 := times[n*50/100]
	p90 := times[n*90/100]
	p95 := times[n*95/100]
	p99 := times[n*99/100]
	min := times[0]
	max := times[n-1]

	b.Logf("=== 连接握手耗时分布 (%d 次) ===", n)
	b.Logf("均值:   %v", mean)
	b.Logf("标准差: %v", stddev)
	b.Logf("最小:   %v", min)
	b.Logf("P50:    %v", p50)
	b.Logf("P90:    %v", p90)
	b.Logf("P95:    %v", p95)
	b.Logf("P99:    %v", p99)
	b.Logf("最大:   %v", max)
}

// =============================================================================
// 2. 并发连接稳定性测试
// =============================================================================

func BenchmarkConnectConcurrency_50(b *testing.B) {
	benchmarkConnectConcurrency(b, 50)
}

func BenchmarkConnectConcurrency_100(b *testing.B) {
	benchmarkConnectConcurrency(b, 100)
}

func BenchmarkConnectConcurrency_200(b *testing.B) {
	benchmarkConnectConcurrency(b, 200)
}

func benchmarkConnectConcurrency(b *testing.B, goroutines int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	failures := 0
	var latencies []time.Duration
	var latencyMu sync.Mutex

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		successes = 0
		failures = 0
		latencies = latencies[:0]

		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
				start := time.Now()
				_, err := c.Connect()
				elapsed := time.Since(start)
				c.Disconnect()

				latencyMu.Lock()
				if err == nil {
					latencies = append(latencies, elapsed)
				}
				latencyMu.Unlock()

				mu.Lock()
				if err == nil {
					successes++
				} else {
					failures++
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		b.Logf("iter %d: %d/%d 成功, %d 失败", i, successes, goroutines, failures)
		if len(latencies) > 0 {
			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
			n := len(latencies)
			b.Logf("  P50=%v P95=%v P99=%v", latencies[n*50/100], latencies[n*95/100], latencies[n*99/100])
		}
	}
}

// =============================================================================
// 3. 连接池效果对比
// =============================================================================

func BenchmarkPoolVsSingle_100req(b *testing.B) {
	benchmarkPoolVsSingle(b, 100)
}

func BenchmarkPoolVsSingle_1000req(b *testing.B) {
	benchmarkPoolVsSingle(b, 1000)
}

func benchmarkPoolVsSingle(b *testing.B, totalRequests int) {
	// Single connection (reconnect per request)
	b.Run("SinglePerRequest", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < totalRequests; j++ {
				c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
				if _, err := c.Connect(); err != nil {
					b.Skipf("connect failed: %v", err)
				}
				_, _ = c.StockCount(1)
				c.Disconnect()
			}
		}
	})

	// Pool size 1 (reuse one connection)
	b.Run("PoolSize1", func(b *testing.B) {
		b.ReportAllocs()
		c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
		if _, err := c.Connect(); err != nil {
			b.Skipf("connect failed: %v", err)
		}
		defer c.Disconnect()
		for i := 0; i < b.N; i++ {
			for j := 0; j < totalRequests; j++ {
				_, _ = c.StockCount(1)
			}
		}
	})

	// Pool size 5
	b.Run("PoolSize5", func(b *testing.B) {
		b.ReportAllocs()
		pool := make([]*gotdx.Client, 5)
		for i := 0; i < 5; i++ {
			c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
			if _, err := c.Connect(); err != nil {
				b.Skipf("connect failed: %v", err)
			}
			pool[i] = c
		}
		defer func() {
			for _, c := range pool {
				c.Disconnect()
			}
		}()
		var mu sync.Mutex
		idx := 0
		for i := 0; i < b.N; i++ {
			for j := 0; j < totalRequests; j++ {
				mu.Lock()
				c := pool[idx%5]
				idx++
				mu.Unlock()
				_, _ = c.StockCount(1)
			}
		}
	})

	// Pool size 10
	b.Run("PoolSize10", func(b *testing.B) {
		b.ReportAllocs()
		pool := make([]*gotdx.Client, 10)
		for i := 0; i < 10; i++ {
			c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
			if _, err := c.Connect(); err != nil {
				b.Skipf("connect failed: %v", err)
			}
			pool[i] = c
		}
		defer func() {
			for _, c := range pool {
				c.Disconnect()
			}
		}()
		var mu sync.Mutex
		idx := 0
		for i := 0; i < b.N; i++ {
			for j := 0; j < totalRequests; j++ {
				mu.Lock()
				c := pool[idx%10]
				idx++
				mu.Unlock()
				_, _ = c.StockCount(1)
			}
		}
	})
}

// =============================================================================
// 4. 心跳保活效果测试
// =============================================================================

func BenchmarkHeartbeatVsNoHeartbeat(b *testing.B) {
	// No heartbeat: connections may die, causing retry overhead
	b.Run("NoHeartbeat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
			if _, err := c.Connect(); err != nil {
				b.Skipf("connect failed: %v", err)
			}
			// Simulate idle gap (no heartbeat)
			// In reality, connection might timeout here
			for j := 0; j < 10; j++ {
				_, _ = c.StockCount(1)
			}
			c.Disconnect()
		}
	})

	// With heartbeat: connections stay alive
	b.Run("WithHeartbeat", func(b *testing.B) {
		b.ReportAllocs()
		c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
		if _, err := c.Connect(); err != nil {
			b.Skipf("connect failed: %v", err)
		}
		defer c.Disconnect()
		for i := 0; i < b.N; i++ {
			// Simulate heartbeat ping
			_, _ = c.StockCount(1)
			for j := 0; j < 10; j++ {
				_, _ = c.StockCount(1)
			}
		}
	})
}

// =============================================================================
// 5. 真实 TDX 服务器连通性测试
// =============================================================================

func TestConnectRealServers(t *testing.T) {
	servers := []struct {
		name string
		addr string
	}{
		{"Main-1", "221.235.208.33:7709"},
		{"Main-2", "113.108.17.38:7709"},
		{"Main-3", "124.160.89.179:7709"},
		{"Main-4", "59.173.143.83:7709"},
		{"Main-5", "123.125.108.23:7709"},
		{"Main-6", "221.235.210.106:7709"},
		{"Ex-1", "121.14.12.130:7727"},
		{"Ex-2", "117.29.145.209:7727"},
	}

	for _, s := range servers {
		t.Run(s.name, func(t *testing.T) {
			c := gotdx.New(gotdx.WithTCPAddress(s.addr), gotdx.WithTimeoutSec(3))
			start := time.Now()
			_, err := c.Connect()
			elapsed := time.Since(start)
			c.Disconnect()

			if err != nil {
				t.Logf("%s: FAILED after %v — %v", s.name, elapsed, err)
			} else {
				t.Logf("%s: OK after %v (addr=%s)", s.name, elapsed, c.CurrentAddress())
			}
		})
	}
}

// =============================================================================
// 6. 自动选最快服务器测试
// =============================================================================

func BenchmarkAutoSelectFastest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := gotdx.New(gotdx.WithAutoSelectFastest(true), gotdx.WithTimeoutSec(6))
		start := time.Now()
		_, err := c.Connect()
		elapsed := time.Since(start)
		c.Disconnect()

		if err != nil {
			b.Logf("iter %d: connect failed after %v", i, elapsed)
			continue
		}
		b.Logf("iter %d: connected to %s in %v", i, c.CurrentAddress(), elapsed)
	}
}

// =============================================================================
// 7. TDXPoolClient 连接池压力测试
// =============================================================================

func BenchmarkTdxPoolClient_PoolFill(b *testing.B) {
	for poolSize := 1; poolSize <= 10; poolSize *= 2 {
		b.Run(fmt.Sprintf("Size%d", poolSize), func(b *testing.B) {
			cfg := DefaultPoolConfig()
			cfg.Size = poolSize
			cfg.ConnectTimeout = 3 * time.Second

			client := NewTdxPoolClient(cfg, 3)
			defer client.Disconnect()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = client.PoolStats()
			}
		})
	}
}

func BenchmarkTdxPoolClient_QuoteThroughput(b *testing.B) {
	cfg := DefaultPoolConfig()
	cfg.Size = 5
	client := NewTdxPoolClient(cfg, 3)
	defer client.Disconnect()

	if !client.IsConnected() {
		b.Skip("pool not connected")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetQuote("000001", 1)
	}
}

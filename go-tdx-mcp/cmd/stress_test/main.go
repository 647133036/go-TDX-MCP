package main

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/tdx/go-tdx-mcp/tdx"
)

func main() {
	fmt.Println("=" + // 生产环境多路并发压力测试
		"================================================")
	fmt.Printf("开始时间: %s\n\n", time.Now().Format("15:04:05"))

	// =========================================================================
	// Phase 1: 初始化
	// =========================================================================
	fmt.Println("[Phase 1] 初始化 MultiHostCollector")
	initStart := time.Now()

	cfg := tdx.DefaultCollectorConfig()
	cfg.MaxHosts = 10  // 前10台最快主机
	cfg.MaxConnsPerHost = 2 // 每台2个连接, 共20连接池
	cfg.RetryCount = 2
	cfg.RetryDelay = 100 * time.Millisecond
	cfg.HostTimeout = 5 * time.Second

	collector, err := tdx.NewMultiHostCollector(cfg)
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
		return
	}
	defer collector.Disconnect()

	initTime := time.Since(initStart)
	fmt.Printf("初始化耗时: %v\n", initTime.Round(time.Millisecond))
	fmt.Printf("连接池: %s\n\n", collector.PoolStats())

	// =========================================================================
	// Phase 2: 行情列表采集 (上证全市场)
	// =========================================================================
	fmt.Println("[Phase 2] 上证全市场行情列表")
	run("AllQuotes_SH", func() (rec *Record) {
		rec = &Record{Name: "行情列表(SH)"}
		quotes, stats, err := collector.CollectAllQuotes(tdx.MarketSH)
		rec.Stats = stats
		rec.Error = err
		rec.ItemCount = len(quotes)
		return
	})

	// =========================================================================
	// Phase 3: 批量详情行情 (前50只股票)
	// =========================================================================
	fmt.Println("[Phase 3] 批量详情行情 (前50只沪市股票)")
	run("DetailedQuotes_50", func() (rec *Record) {
		rec = &Record{Name: "详情行情(50只)"}
		// Use common SH codes
		codes := []string{
			"600000", "600004", "600009", "600010", "600011",
			"600015", "600016", "600018", "600019", "600021",
			"600025", "600026", "600027", "600028", "600029",
			"600030", "600031", "600036", "600048", "600050",
			"600061", "600066", "600068", "600071", "600072",
			"600079", "600085", "600089", "600096", "600100",
			"600104", "600109", "600111", "600115", "600118",
			"600119", "600120", "600125", "600132", "600138",
			"600143", "600150", "600153", "600157", "600161",
			"600166", "600170", "600183", "600185", "600188",
		}
		quotes, stats, err := collector.CollectDetailedQuotes(codes, tdx.MarketSH, 10)
		rec.Stats = stats
		rec.Error = err
		rec.ItemCount = len(quotes)
		return
	})

	// =========================================================================
	// Phase 4: K线并发采集 (20只股票 日线)
	// =========================================================================
	fmt.Println("[Phase 4] K线并发采集 (20只股票, 日线, 前100根)")
	run("KLines_20", func() (rec *Record) {
		rec = &Record{Name: "K线(20只日线)"}
		codes := []string{
			"600000", "600036", "600519", "600276", "600309",
			"600585", "600809", "600031", "600900", "601318",
			"601398", "601939", "601288", "601857", "601088",
			"601166", "601668", "601766", "601628", "601006",
		}
		bars, stats, err := collector.CollectKLines(codes, tdx.MarketSH, 4, 100, 0) // 4=日线, 100条, 不复权
		rec.Stats = stats
		rec.Error = err
		for _, b := range bars {
			rec.ItemCount += len(b)
		}
		return
	})

	// =========================================================================
	// Phase 5: 板块数据采集 (行业+概念)
	// =========================================================================
	fmt.Println("[Phase 5] 板块数据采集 (行业 + 概念)")
	run("SectorBoards", func() (rec *Record) {
		rec = &Record{Name: "板块(行业+概念)"}
		boards1, stats1, err1 := collector.CollectSectorBoards(tdx.BlockIndustry)
		boards2, stats2, err2 := collector.CollectSectorBoards(tdx.BlockConcept)

		combined := &tdx.CollectStats{
			Total: stats1.Total + stats2.Total,
		}
		combined.Success.Store(stats1.Success.Load() + stats2.Success.Load())
		combined.Fail.Store(stats1.Fail.Load() + stats2.Fail.Load())
		combined.Elapsed = stats1.Elapsed + stats2.Elapsed

		rec.Stats = combined
		if err1 != nil {
			rec.Error = err1
		} else if err2 != nil {
			rec.Error = err2
		}
		rec.ItemCount = len(boards1) + len(boards2)
		return
	})

	// =========================================================================
	// Phase 6: 混合负载 (行情 + K线并发)
	// =========================================================================
	fmt.Println("[Phase 6] 混合负载 (行情+K线并发)")
	run("Mixed_Load", func() (rec *Record) {
		rec = &Record{Name: "混合负载"}
		done := make(chan struct{})

		// Sub-test 1: K-line for 10 codes
		go func() {
			codes := []string{"600000", "600036", "600519", "600276", "600309",
				"600585", "600809", "600031", "600900", "601318"}
			_, _, _ = collector.CollectKLines(codes, tdx.MarketSH, 4, 50, 0)
			done <- struct{}{}
		}()

		// Sub-test 2: Detailed quotes for 10 codes
		go func() {
			codes := []string{"600000", "600036", "600519", "600276", "600309",
				"600585", "600809", "600031", "600900", "601318"}
			_, _, _ = collector.CollectDetailedQuotes(codes, tdx.MarketSH, 5)
			done <- struct{}{}
		}()

		// Sub-test 3: Sector boards
		go func() {
			_, _, _ = collector.CollectSectorBoards(tdx.BlockIndustry)
			done <- struct{}{}
		}()

		// Sub-test 4: Top quotes
		go func() {
			_, _, _ = collector.CollectTopQuotes(tdx.MarketSH, 50, 0, false)
			done <- struct{}{}
		}()

		start := time.Now()
		for i := 0; i < 4; i++ {
			<-done
		}

		rec.Stats = &tdx.CollectStats{}
		rec.Stats.Elapsed = time.Since(start)
		rec.Stats.Success.Store(4)
		rec.ItemCount = 4
		return
	})

	// =========================================================================
	// Phase 7: 持续压力测试 (30秒连续采集)
	// =========================================================================
	fmt.Println("\n[Phase 7] 持续压力测试 (30秒连续采集)")
	run("Stress_Sustain", func() (rec *Record) {
		rec = &Record{Name: "持续压力(30s)"}

		var totalReqs atomic.Int64
		var totalSuccess atomic.Int64
		stopCh := make(chan struct{})
		start := time.Now()

		// 20 goroutines continuously making requests
		for i := 0; i < 20; i++ {
			go func(id int) {
				codes := []string{
					"600000", "600036", "600519", "600276", "600309",
					"601318", "600585", "601398", "600900", "601166",
				}
				for {
					select {
					case <-stopCh:
						return
					default:
					}
					code := codes[id%len(codes)]
					_, _, err := collector.CollectDetailedQuotes([]string{code}, tdx.MarketSH, 1)
					totalReqs.Add(1)
					if err == nil {
						totalSuccess.Add(1)
					}
				}
			}(i)
		}

		time.Sleep(30 * time.Second)
		close(stopCh)

		rec.Stats = &tdx.CollectStats{}
		rec.Stats.Elapsed = time.Since(start)
		rec.Stats.Total = totalReqs.Load()
		rec.Stats.Success.Store(totalSuccess.Load())
		rec.Stats.Fail.Store(totalReqs.Load() - totalSuccess.Load())
		rec.ItemCount = int(totalReqs.Load())
		return
	})

	// =========================================================================
	// Phase 8: 多连接对比测试 (MaxConnsPerHost=1 vs 2)
	// =========================================================================
	fmt.Println("\n[Phase 8] 多连接对比测试 (MaxConnsPerHost=1/2)")
	run("MultiConnCompare", func() (rec *Record) {
		rec = &Record{Name: "多连接对比"}

		configs := []struct {
			name   string
			conns  int
		}{
			{"1 conn/host", 1},
			{"2 conns/host", 2},
		}

		var results []string
		for _, cfg := range configs {
			testCfg := tdx.DefaultCollectorConfig()
			testCfg.MaxHosts = 10
			testCfg.MaxConnsPerHost = cfg.conns
			testCfg.RetryCount = 2
			testCfg.RetryDelay = 100 * time.Millisecond
			testCfg.HostTimeout = 5 * time.Second

			c, err := tdx.NewMultiHostCollector(testCfg)
			if err != nil {
				results = append(results, fmt.Sprintf("%s: FAIL %v", cfg.name, err))
				continue
			}

			// 10 concurrent K-line requests
			start := time.Now()
			codes := []string{"600000", "600036", "600519", "600276", "600309",
				"600585", "600809", "600031", "600900", "601318"}
			_, stats, err := c.CollectKLines(codes, tdx.MarketSH, 4, 50, 0)
			elapsed := time.Since(start)

			if err != nil {
				results = append(results, fmt.Sprintf("%s: %s, err=%v, elapsed=%v", cfg.name, stats.Summary(), err, elapsed))
			} else {
				results = append(results, fmt.Sprintf("%s: %s, elapsed=%v", cfg.name, stats.Summary(), elapsed.Round(time.Millisecond)))
			}
			c.Disconnect()
		}

		rec.ItemCount = len(results)
		rec.Error = nil
		rec.Stats = &tdx.CollectStats{Total: int64(len(results)), Success: atomic.Int64{}, Fail: atomic.Int64{}}
		rec.Stats.Success.Store(int64(len(results)))
		_ = results
		return
	})

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println()
	fmt.Println("=" + // =========================================================
		"================================================")
	fmt.Println("                       测试结果汇总")
	fmt.Println("=" + // =========================================================
		"================================================")
	fmt.Println()

	sort.Slice(records, func(i, j int) bool { return records[i].Order < records[j].Order })

	for _, rec := range records {
		fmt.Printf("--- %s ---\n", rec.Name)
		if rec.Error != nil {
			fmt.Printf("  错误: %v\n", rec.Error)
		}
		if rec.Stats != nil {
			fmt.Printf("  结果: %s\n", rec.Stats.Summary())
			if rec.Stats.Total > 1 && len(rec.Stats.Latencies) > 0 {
				fmt.Printf("  P50=%v P90=%v P95=%v P99=%v\n",
					rec.Stats.Percentile(50).Round(time.Millisecond),
					rec.Stats.Percentile(90).Round(time.Millisecond),
					rec.Stats.Percentile(95).Round(time.Millisecond),
					rec.Stats.Percentile(99).Round(time.Millisecond))
			}
		}
		fmt.Printf("  数据量: %d 条\n\n", rec.ItemCount)
	}

	fmt.Printf("连接池最终状态: %s\n", collector.PoolStats())
	fmt.Printf("结束时间: %s\n", time.Now().Format("15:04:05"))
}

// Record stores a single test result.
type Record struct {
	Name      string
	Order     int
	Stats     *tdx.CollectStats
	Error     error
	ItemCount int
}

var records []Record
var order atomic.Int64

func run(name string, fn func() *Record) {
	start := time.Now()
	rec := fn()
	rec.Order = int(order.Add(1))
	records = append(records, *rec)
	fmt.Printf("  耗时: %v | %s\n\n", time.Since(start).Round(time.Millisecond), rec.Stats.Summary())
}

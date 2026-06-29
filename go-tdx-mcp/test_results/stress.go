package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	baseURL := "http://localhost:8000"
	
	endpoints := []string{
		"/api/v1/health",
		"/api/v1/quotes?codes=SZ000001,SH600000",
		"/api/v1/bars?code=000001&market=sz&period=day&count=5",
		"/api/v1/indicator/list",
		"/api/v1/indicator/compute_all?code=000001&market=sz&indicators=MACD,KDJ",
		"/api/v1/chanlun/analyze?code=000001&market=sz&period=day&count=30",
		"/api/v1/board/list?board_type=HY&top_n=5",
		"/api/v1/market-overview",
	}
	
	numWorkers := 10
	requestsPerWorker := 20
	totalRequests := numWorkers * len(endpoints) * requestsPerWorker
	
	fmt.Println("========================================")
	fmt.Println("Web API Stress Test")
	fmt.Println("========================================")
	fmt.Printf("Workers: %d\n", numWorkers)
	fmt.Printf("Requests per endpoint per worker: %d\n", requestsPerWorker)
	fmt.Printf("Total requests: %d\n\n", totalRequests)
	
	var totalSuccess atomic.Int64
	var totalFail atomic.Int64
	var totalBytes atomic.Int64
	
	start := time.Now()
	
	for _, ep := range endpoints {
		var wg sync.WaitGroup
		var epSuccess atomic.Int64
		var epFail atomic.Int64
		var epBytes atomic.Int64
		var epDuration atomic.Int64
		
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < requestsPerWorker; i++ {
					reqStart := time.Now()
					
					resp, err := http.Get(baseURL + ep)
					if err != nil {
						epFail.Add(1)
						totalFail.Add(1)
						continue
					}
					
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					
					epDuration.Add(time.Since(reqStart).Milliseconds())
					epBytes.Add(int64(len(body)))
					epSuccess.Add(1)
					totalSuccess.Add(1)
					totalBytes.Add(int64(len(body)))
				}
			}()
		}
		
		wg.Wait()
		
		avgMs := int64(0)
		if epSuccess.Load() > 0 {
			avgMs = epDuration.Load() / epSuccess.Load()
		}
		
		fmt.Printf("[%s] SUCCESS=%d FAIL=%d AVG=%dms BYTES=%d\n",
			ep, epSuccess.Load(), epFail.Load(), avgMs, epBytes.Load())
	}
	
	elapsed := time.Since(start)
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("SUMMARY")
	fmt.Println("========================================")
	fmt.Printf("Total: %d requests\n", totalSuccess.Load()+totalFail.Load())
	fmt.Printf("Success: %d (%.1f%%)\n", totalSuccess.Load(), float64(totalSuccess.Load())/float64(totalSuccess.Load()+totalFail.Load())*100)
	fmt.Printf("Fail: %d\n", totalFail.Load())
	fmt.Printf("Throughput: %.0f req/s\n", float64(totalSuccess.Load())/elapsed.Seconds())
	fmt.Printf("Total data: %.2f MB\n", float64(totalBytes.Load())/1024/1024)
	fmt.Printf("Elapsed: %v\n", elapsed.Round(time.Millisecond))
}

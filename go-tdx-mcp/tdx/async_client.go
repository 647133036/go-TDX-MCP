package tdx

import (
	"context"
	"fmt"
	"sync"

	"github.com/bensema/gotdx"
	"github.com/bensema/gotdx/proto"
)

// =============================================================================
// Strategy 4: TdxAsyncClient — Go channels/select for并发异步请求
// =============================================================================

// AsyncRequest represents a single async TDX query.
type AsyncRequest struct {
	Type   string
	Code   string
	Market int
	Period string
	Count  int
	Adjust int
	Result chan AsyncResult
}

// AsyncResult holds the result of an async request.
type AsyncResult struct {
	Quote *SecurityQuoteWrapper
	Error error
}

// SecurityQuoteWrapper bundles all possible return types.
type SecurityQuoteWrapper struct {
	Quote            interface{}
	KLine            []proto.SecurityBar
	TickChart        []proto.MinuteTimeData
	Trans            []proto.TransactionData
	Auction          []proto.AuctionData
	Unusual          []proto.UnusualData
	SymbolInfo       interface{}
	F10              *gotdx.CompanyInfoBundle
	Finance          *proto.GetFinanceInfoReply
	XDXR             *proto.GetXDXRInfoReply
	ExQuote          *proto.ExQuoteItem
	ExKLine          []proto.ExKLineItem
	SecurityList     *proto.GetSecurityListReply
	SecurityCount    uint16
	CapitalFlow      *proto.MACCapitalFlowReply
	Error            error
}

// TdxAsyncClient manages async request queue with worker pool.
type TdxAsyncClient struct {
	pool      *TdxPoolClient
	direct    *TdxDirectClient
	reqChan   chan AsyncRequest
	wg        sync.WaitGroup
	mu        sync.Mutex
	active    int
	maxActive int
}

// NewTdxAsyncClient creates an async client backed by pool + direct strategies.
func NewTdxAsyncClient(pool *TdxPoolClient, direct *TdxDirectClient, maxWorkers int) *TdxAsyncClient {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}
	ac := &TdxAsyncClient{
		pool:      pool,
		direct:    direct,
		reqChan:   make(chan AsyncRequest, 100),
		maxActive: maxWorkers,
	}
	ac.startWorkers(maxWorkers)
	return ac
}

// startWorkers launches goroutine pool to process async requests.
func (ac *TdxAsyncClient) startWorkers(n int) {
	for i := 0; i < n; i++ {
		ac.wg.Add(1)
		go ac.worker()
	}
}

// worker processes requests from the channel.
func (ac *TdxAsyncClient) worker() {
	defer ac.wg.Done()
	for req := range ac.reqChan {
		ac.mu.Lock()
		ac.active++
		ac.mu.Unlock()

		req.Result <- AsyncResult{
			Quote: ac.handleRequest(req),
			Error: nil,
		}

		ac.mu.Lock()
		ac.active--
		ac.mu.Unlock()
	}
}

// handleRequest routes async request to appropriate handler.
func (ac *TdxAsyncClient) handleRequest(req AsyncRequest) *SecurityQuoteWrapper {
	switch req.Type {
	case "quote":
		q, err := ac.pool.GetQuote(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{Quote: q}
	case "batch_quotes":
		// Batch quotes need special handling — use direct client for speed
		pairs := []struct{ Market int; Code string }{{Market: req.Market, Code: req.Code}}
		qs, err := ac.direct.GetBatchQuotes(pairs)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		if len(qs) > 0 {
			return &SecurityQuoteWrapper{Quote: qs[0]}
		}
		return &SecurityQuoteWrapper{Error: fmt.Errorf("empty batch result")}
	case "kline":
		bars, err := ac.pool.GetKLine(req.Code, req.Market, req.Period, req.Count, req.Adjust)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{KLine: bars}
	case "kline_adjust":
		bars, err := ac.pool.GetKLineWithAdjust(req.Code, req.Market, req.Period, req.Count, req.Adjust)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{KLine: bars}
	case "tick":
		data, err := ac.pool.GetTickChart(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{TickChart: data}
	case "transaction":
		data, err := ac.pool.GetTransaction(req.Code, req.Market, req.Count)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{Trans: data}
	case "auction":
		data, err := ac.pool.GetAuction(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{Auction: data}
	case "unusual":
		data, err := ac.pool.GetUnusual(req.Market, req.Count)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{Unusual: data}
	case "symbol_info":
		info, err := ac.pool.GetSymbolInfo(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{SymbolInfo: info}
	case "f10":
		info, err := ac.pool.GetF10(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{F10: info}
	case "finance":
		info, err := ac.pool.GetFinanceInfo(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{Finance: info}
	case "xdxr":
		info, err := ac.pool.GetXDXRInfo(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{XDXR: info}
	case "ex_quote":
		// ExGetQuote requires category — use 0 as default
		q, err := ac.pool.ExGetQuote(req.Code, 0)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{ExQuote: q}
	case "ex_kline":
		k, err := ac.pool.ExGetKLine(0, req.Code, 1, req.Count)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{ExKLine: k}
	case "security_list":
		list, err := ac.pool.GetSecurityList(req.Market, 0)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{SecurityList: list}
	case "security_count":
		count, err := ac.pool.GetSecurityCount(req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{SecurityCount: count}
	case "capital_flow":
		flow, err := ac.pool.GetCapitalFlow(req.Code, req.Market)
		if err != nil {
			return &SecurityQuoteWrapper{Error: err}
		}
		return &SecurityQuoteWrapper{CapitalFlow: flow}
	default:
		return &SecurityQuoteWrapper{Error: fmt.Errorf("unknown request type: %s", req.Type)}
	}
}

// Submit sends an async request and returns a result channel.
func (ac *TdxAsyncClient) Submit(req AsyncRequest) {
	if req.Result == nil {
		req.Result = make(chan AsyncResult, 1)
	}
	ac.reqChan <- req
}

// SubmitAndAwait sends a request and waits for result with context timeout.
func (ac *TdxAsyncClient) SubmitAndAwait(ctx context.Context, req AsyncRequest) (*SecurityQuoteWrapper, error) {
	if req.Result == nil {
		req.Result = make(chan AsyncResult, 1)
	}
	ac.reqChan <- req

	select {
	case result := <-req.Result:
		return result.Quote, result.Error
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ActiveCount returns number of currently processing requests.
func (ac *TdxAsyncClient) ActiveCount() int {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return ac.active
}

// Close stops the async client and drains pending requests.
func (ac *TdxAsyncClient) Close() {
	close(ac.reqChan)
	ac.wg.Wait()
}

// =============================================================================
// Concurrent batch processor — fire-and-forget with result aggregation
// =============================================================================

// BatchProcessor fires multiple async requests and aggregates results.
type BatchProcessor struct {
	clients []*TdxAsyncClient
	results []chan AsyncResult
	mu      sync.Mutex
	done    int
	total   int
}

// NewBatchProcessor creates a processor for N concurrent requests.
func NewBatchProcessor(n int) *BatchProcessor {
	return &BatchProcessor{
		results: make([]chan AsyncResult, n),
		total:   n,
		done:    0,
	}
}

// AddRequest queues a request at index i.
func (bp *BatchProcessor) AddRequest(i int, req AsyncRequest) {
	bp.results[i] = req.Result
}

// Wait blocks until all results are received or context cancels.
func (bp *BatchProcessor) Wait(ctx context.Context) []AsyncResult {
	var results []AsyncResult
	for _, ch := range bp.results {
		if ch == nil {
			results = append(results, AsyncResult{Error: fmt.Errorf("no result channel at index")})
			continue
		}
		select {
		case r := <-ch:
			results = append(results, r)
		case <-ctx.Done():
			results = append(results, AsyncResult{Error: ctx.Err()})
		}
	}
	return results
}

// Completed returns number of completed requests.
func (bp *BatchProcessor) Completed() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.done
}

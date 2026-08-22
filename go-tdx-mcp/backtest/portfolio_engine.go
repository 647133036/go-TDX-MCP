package backtest

import (
	"github.com/tdx/go-tdx-mcp/indicator"
)

type PortfolioItem struct {
	Code   string
	Market int
	Bars   []indicator.Bar
}

type PortfolioResult struct {
	Strategy    string              `json:"strategy"`
	Items       []PortfolioItem     `json:"items"`
	TotalCash   float64             `json:"total_cash"`
	Results     []Result            `json:"results"`
	Performance Performance         `json:"performance"`
}

func RunPortfolio(strategy Strategy, items []PortfolioItem, totalCash float64) *PortfolioResult {
	if len(items) == 0 {
		return &PortfolioResult{Strategy: strategy.Name(), TotalCash: totalCash}
	}

	cashPerItem := totalCash / float64(len(items))
	engine := NewEngine(cashPerItem)
	results := make([]Result, 0, len(items))
	allEquityCurves := make([][]float64, 0, len(items))

	for _, item := range items {
		if len(item.Bars) < 2 {
			continue
		}
		itemEngine := NewEngine(cashPerItem)
		itemEngine.SetCommission(engine.commission)
		itemEngine.SetSlippage(engine.slippage)
		r := itemEngine.Run(strategy, item.Bars)
		r.Code = item.Code
		r.Market = item.Market
		results = append(results, *r)
		allEquityCurves = append(allEquityCurves, equityCurveFromResult(r, cashPerItem))
	}

	portfolio := &PortfolioResult{
		Strategy:    strategy.Name(),
		TotalCash:   totalCash,
		Items:       items,
		Results:     results,
	}
	if len(allEquityCurves) > 0 {
		combined := combineEquityCurves(allEquityCurves, len(items))
		perf := calcPerformance(combined, nil)
		portfolio.Performance = perf
	}
	return portfolio
}

func equityCurveFromResult(r *Result, initial float64) []float64 {
	if r == nil {
		return []float64{initial, initial}
	}
	curve := make([]float64, r.BarCount)
	if len(curve) == 0 {
		return []float64{initial}
	}
	start := initial
	end := r.FinalEquity
	if r.BarCount == 1 {
		curve[0] = start
		return curve
	}
	step := (end - start) / float64(r.BarCount-1)
	for i := 0; i < r.BarCount; i++ {
		curve[i] = start + step*float64(i)
	}
	return curve
}

func combineEquityCurves(curves [][]float64, nItems int) []float64 {
	if len(curves) == 0 {
		return nil
	}
	maxLen := 0
	for _, c := range curves {
		if len(c) > maxLen {
			maxLen = len(c)
		}
	}
	if maxLen == 0 {
		return nil
	}
	combined := make([]float64, maxLen)
	for i := 0; i < maxLen; i++ {
		var sum float64
		for _, c := range curves {
			if i < len(c) {
				sum += c[i]
			}
		}
		combined[i] = sum
	}
	return combined
}

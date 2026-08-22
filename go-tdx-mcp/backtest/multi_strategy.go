package backtest

import (
	"github.com/tdx/go-tdx-mcp/indicator"
)

type MultiStrategyItem struct {
	Strategy  string
	Params    map[string]float64
	Code      string
	Market    int
	Bars      []indicator.Bar
	Allocation float64
}

type MultiStrategyResult struct {
	Items        []MultiStrategyItem     `json:"items"`
	TotalCash    float64                 `json:"total_cash"`
	Results      []Result                `json:"results"`
	Performance  Performance             `json:"performance"`
}

func RunMultiStrategy(items []MultiStrategyItem, totalCash float64) *MultiStrategyResult {
	if len(items) == 0 {
		return &MultiStrategyResult{TotalCash: totalCash}
	}

	result := &MultiStrategyResult{
		Items:     items,
		TotalCash: totalCash,
		Results:   make([]Result, 0, len(items)),
	}

	allocSum := 0.0
	for _, item := range items {
		allocSum += item.Allocation
	}
	if allocSum <= 0 {
		allocSum = float64(len(items))
		for i := range items {
			items[i].Allocation = 1.0
		}
	}

	allCurves := make([][]float64, 0, len(items))
	for _, item := range items {
		cash := totalCash * item.Allocation / allocSum
		st := NewStrategyWithParams(item.Strategy, item.Params)
		if st == nil {
			continue
		}
		if len(item.Bars) < 2 {
			continue
		}
		engine := NewEngine(cash)
		r := engine.Run(st, item.Bars)
		r.Code = item.Code
		r.Market = item.Market
		result.Results = append(result.Results, *r)
		allCurves = append(allCurves, equityCurveFromResult(r, cash))
	}

	if len(allCurves) > 0 {
		combined := combineEquityCurves(allCurves, len(allCurves))
		result.Performance = calcPerformance(combined, nil)
	}
	return result
}

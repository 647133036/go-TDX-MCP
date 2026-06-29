package backtest

type AttributionReport struct {
	TotalReturn      float64            `json:"total_return"`
	AllocationReturn float64            `json:"allocation_return"`
	SelectionReturn  float64            `json:"selection_return"`
	InteractionReturn float64           `json:"interaction_return"`
	FactorReturns    map[string]float64 `json:"factor_returns"`
	SpecificReturn   float64            `json:"specific_return"`
	TotalTradeCost   float64            `json:"total_trade_cost"`
	SlippageCost     float64            `json:"slippage_cost"`
	CommissionCost   float64            `json:"commission_cost"`
	StampTaxCost     float64            `json:"stamp_tax_cost"`
}

type CostAttribution struct {
	trades []Trade
}

func NewCostAttribution(trades []Trade) *CostAttribution {
	return &CostAttribution{trades: trades}
}

func (a *CostAttribution) Analyze() AttributionReport {
	r := AttributionReport{}
	var slippageCost, commissionCost, stampTaxCost float64
	for _, t := range a.trades {
		commissionCost += t.EntryPrice * float64(t.Size) * 0.0003
		if t.ExitPrice > 0 {
			commissionCost += t.ExitPrice * float64(t.Size) * 0.0003
			stampTaxCost += t.ExitPrice * float64(t.Size) * 0.001
		}
	}
	r.SlippageCost = slippageCost
	r.CommissionCost = commissionCost
	r.StampTaxCost = stampTaxCost
	r.TotalTradeCost = slippageCost + commissionCost + stampTaxCost
	return r
}

type BrinsonGroup struct {
	PortfolioWeight  float64
	BenchmarkWeight  float64
	PortfolioReturn  float64
	BenchmarkReturn  float64
}

func BrinsonAttribution(portfolioReturn, benchmarkReturn float64, groups []BrinsonGroup) AttributionReport {
	r := AttributionReport{TotalReturn: portfolioReturn}
	if len(groups) == 0 {
		r.SelectionReturn = portfolioReturn - benchmarkReturn
		return r
	}
	var allocation, selection, interaction float64
	for _, g := range groups {
		wDelta := g.PortfolioWeight - g.BenchmarkWeight
		rDelta := g.PortfolioReturn - g.BenchmarkReturn
		allocation += wDelta * g.BenchmarkReturn
		selection += g.BenchmarkWeight * rDelta
		interaction += wDelta * rDelta
	}
	r.AllocationReturn = allocation
	r.SelectionReturn = selection
	r.InteractionReturn = interaction
	return r
}

func FactorAttribution(totalReturn float64, exposures, factorReturns []float64, factorNames []string) AttributionReport {
	r := AttributionReport{TotalReturn: totalReturn}
	r.FactorReturns = make(map[string]float64)
	var totalFactorReturn float64
	minLen := len(exposures)
	if len(factorReturns) < minLen {
		minLen = len(factorReturns)
	}
	for i := 0; i < minLen; i++ {
		name := "factor"
		if i < len(factorNames) {
			name = factorNames[i]
		}
		contrib := exposures[i] * factorReturns[i]
		r.FactorReturns[name] = contrib
		totalFactorReturn += contrib
	}
	r.SpecificReturn = totalReturn - totalFactorReturn
	return r
}

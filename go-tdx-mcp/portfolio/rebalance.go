package portfolio

import "math"

type PortfolioResult struct {
	TotalReturn    float64   `json:"total_return"`
	AnnualReturn   float64   `json:"annual_return"`
	MaxDrawdown    float64   `json:"max_drawdown"`
	SharpeRatio    float64   `json:"sharpe_ratio"`
	TurnoverRate   float64   `json:"turnover_rate"`
	EquityCurve    []float64 `json:"equity_curve"`
	PositionCounts []int     `json:"position_counts"`
}

type WeightMap map[string]float64

type RebalanceEngine struct {
	initialCapital float64
	cash           float64
	positions      map[string]float64
	weights        WeightMap
	equityCurve    []float64
	commissionRate float64
	slippageRate   float64
	minHolding     int
}

func NewRebalanceEngine(initialCapital float64) *RebalanceEngine {
	return &RebalanceEngine{
		initialCapital: initialCapital,
		cash:           initialCapital,
		positions:      make(map[string]float64),
		weights:        make(WeightMap),
		equityCurve:    make([]float64, 0),
		commissionRate: 0.0003,
		slippageRate:   0.001,
		minHolding:     3,
	}
}

func (e *RebalanceEngine) SetCommission(rate float64) { e.commissionRate = rate }
func (e *RebalanceEngine) SetSlippage(rate float64)    { e.slippageRate = rate }

type RebalanceSignal struct {
	Code     string
	Weight   float64
	Price    float64
	SigType  string
}

func (e *RebalanceEngine) Rebalance(signals []RebalanceSignal, prices map[string]float64) {
	if len(signals) == 0 {
		return
	}

	targetWeights := make(WeightMap)
	var totalWeight float64
	for _, s := range signals {
		if s.Weight > 0 {
			targetWeights[s.Code] = s.Weight
			totalWeight += s.Weight
		}
	}
	if totalWeight == 0 {
		return
	}
	for k, w := range targetWeights {
		targetWeights[k] = w / totalWeight
	}

	totalEquity := e.calcTotalEquity(prices)
	if totalEquity <= 0 {
		return
	}

	for code := range e.positions {
		if _, inTarget := targetWeights[code]; !inTarget {
			price, ok := prices[code]
			if !ok {
				continue
			}
			proceeds := e.positions[code] * price * (1 - e.slippageRate) * (1 - e.commissionRate)
			e.cash += proceeds
			delete(e.positions, code)
		}
	}

	for code, targetW := range targetWeights {
		price, ok := prices[code]
		if !ok || price <= 0 {
			continue
		}
		targetVal := totalEquity * targetW
		currentVal := e.positions[code] * price
		delta := targetVal - currentVal
		if delta > 0 && e.cash >= delta*(1+e.commissionRate) {
			shares := delta / (price * (1 + e.slippageRate))
			cost := shares * price * (1 + e.slippageRate) * (1 + e.commissionRate)
			e.cash -= cost
			e.positions[code] += shares
		} else if delta < 0 && e.positions[code] > 0 {
			shares := math.Min(e.positions[code], -delta/(price*(1-e.slippageRate)))
			proceeds := shares * price * (1 - e.slippageRate) * (1 - e.commissionRate)
			e.cash += proceeds
			e.positions[code] -= shares
			if e.positions[code] < 0.01 {
				delete(e.positions, code)
			}
		}
	}

	equity := e.calcTotalEquity(prices)
	e.equityCurve = append(e.equityCurve, equity)
}

func (e *RebalanceEngine) calcTotalEquity(prices map[string]float64) float64 {
	equity := e.cash
	for code, shares := range e.positions {
		if price, ok := prices[code]; ok {
			equity += shares * price
		}
	}
	return equity
}

func (e *RebalanceEngine) GetResult() *PortfolioResult {
	r := &PortfolioResult{
		EquityCurve: e.equityCurve,
	}
	if len(e.equityCurve) == 0 {
		return r
	}

	init := e.initialCapital
	final := e.equityCurve[len(e.equityCurve)-1]
	r.TotalReturn = final/init - 1

	if len(e.equityCurve) > 0 {
		years := float64(len(e.equityCurve)) / 52
		if years > 0 && final > init {
			r.AnnualReturn = math.Pow(final/init, 1/years) - 1
		}
	}

	peak := init
	var maxDD float64
	for _, eq := range e.equityCurve {
		if eq > peak {
			peak = eq
		}
		dd := peak - eq
		if dd > maxDD {
			maxDD = dd
		}
	}
	r.MaxDrawdown = maxDD / init

	if len(e.equityCurve) > 1 {
		returns := make([]float64, len(e.equityCurve)-1)
		for i := 1; i < len(e.equityCurve); i++ {
			if e.equityCurve[i-1] > 0 {
				returns[i-1] = e.equityCurve[i]/e.equityCurve[i-1] - 1
			}
		}
		var sum, sum2 float64
		for _, ret := range returns {
			sum += ret
			sum2 += ret * ret
		}
		mean := sum / float64(len(returns))
		variance := sum2/float64(len(returns)) - mean*mean
		if variance > 0 {
			r.SharpeRatio = mean / math.Sqrt(variance) * math.Sqrt(52)
		}
	}

	return r
}

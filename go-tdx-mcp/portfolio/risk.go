package portfolio

import "math"

type RiskModel struct {
	volatilities    map[string]float64
	riskContribution map[string]float64
	covariance      map[string]map[string]float64
}

func NewRiskModel() *RiskModel {
	return &RiskModel{
		volatilities:     make(map[string]float64),
		riskContribution: make(map[string]float64),
		covariance:       make(map[string]map[string]float64),
	}
}

func (r *RiskModel) SetVolatility(code string, vol float64) {
	r.volatilities[code] = vol
}

func (r *RiskModel) GetVolatility(code string) float64 {
	if vol, ok := r.volatilities[code]; ok {
		return vol
	}
	return 0.1
}

func (r *RiskModel) SetRiskContribution(code string, contrib float64) {
	r.riskContribution[code] = contrib
}

func (r *RiskModel) GetRiskContribution(code string) float64 {
	if contrib, ok := r.riskContribution[code]; ok {
		return contrib
	}
	return 0.1
}

func (r *RiskModel) SetCovariance(code1, code2 string, cov float64) {
	if _, ok := r.covariance[code1]; !ok {
		r.covariance[code1] = make(map[string]float64)
	}
	r.covariance[code1][code2] = cov
}

func (r *RiskModel) GetCovariance(code1, code2 string) float64 {
	if m, ok := r.covariance[code1]; ok {
		if cov, ok2 := m[code2]; ok2 {
			return cov
		}
	}
	return 0
}

func (r *RiskModel) EstimateFromReturns(returns map[string][]float64) {
	for code, rets := range returns {
		vol := calcVolatility(rets)
		r.volatilities[code] = vol
		r.riskContribution[code] = vol
	}

	codes := make([]string, 0, len(returns))
	for code := range returns {
		codes = append(codes, code)
	}
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			c1, c2 := codes[i], codes[j]
			cov := calcCovariance(returns[c1], returns[c2])
			r.SetCovariance(c1, c2, cov)
			r.SetCovariance(c2, c1, cov)
		}
	}
}

func (r *RiskModel) PortfolioRisk(weights WeightMap) float64 {
	var totalRisk float64
	codes := make([]string, 0, len(weights))
	for code := range weights {
		codes = append(codes, code)
	}
	for _, c1 := range codes {
		for _, c2 := range codes {
			cov := 0.0
			if c1 == c2 {
				v := r.GetVolatility(c1)
				cov = v * v
			} else {
				cov = r.GetCovariance(c1, c2)
			}
			totalRisk += weights[c1] * weights[c2] * cov
		}
	}
	return math.Sqrt(math.Max(totalRisk, 0))
}

func (r *RiskModel) PositionCount() int {
	return len(r.volatilities)
}

func calcVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))
	var ss float64
	for _, r := range returns {
		d := r - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(returns)-1))
}

func calcCovariance(x, y []float64) float64 {
	n := len(x)
	if n < 2 || len(y) != n {
		return 0
	}
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)
	var sumXY float64
	for i := 0; i < n; i++ {
		sumXY += (x[i] - meanX) * (y[i] - meanY)
	}
	return sumXY / float64(n-1)
}

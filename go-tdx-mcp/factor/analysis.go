package factor

import (
	"math"
	"sort"
)

type FactorReport struct {
	Name             string             `json:"name"`
	ICMean           float64            `json:"ic_mean"`
	ICStd            float64            `json:"ic_std"`
	IR               float64            `json:"ir"`
	ICPositiveRate   float64            `json:"ic_positive_rate"`
	QuantileReturns  map[string]float64 `json:"quantile_returns"`
	TopMinusBottom   float64            `json:"top_minus_bottom"`
	TurnoverRate     float64            `json:"turnover_rate"`
	AutoCorr         float64            `json:"autocorr"`
	CoverageRatio    float64            `json:"coverage_ratio"`
}

type Analyzer struct {
	factorValues []float64
	returns      []float64
	factorName   string
	nQuantiles   int
}

func NewAnalyzer(factorValues, returns []float64, factorName string, nQuantiles int) *Analyzer {
	if nQuantiles <= 0 {
		nQuantiles = 5
	}
	validFactor := make([]float64, 0)
	validReturns := make([]float64, 0)
	for i := 0; i < len(factorValues) && i < len(returns); i++ {
		if !math.IsNaN(factorValues[i]) && !math.IsInf(factorValues[i], 0) &&
			!math.IsNaN(returns[i]) && !math.IsInf(returns[i], 0) {
			validFactor = append(validFactor, factorValues[i])
			validReturns = append(validReturns, returns[i])
		}
	}
	return &Analyzer{
		factorValues: validFactor,
		returns:      validReturns,
		factorName:   factorName,
		nQuantiles:   nQuantiles,
	}
}

func (a *Analyzer) ComputeIC(method string) float64 {
	n := len(a.factorValues)
	if n < 5 {
		return 0
	}
	if method == "spearman" {
		return rankCorrelation(a.factorValues, a.returns)
	}
	return pearsonCorrelation(a.factorValues, a.returns)
}

func (a *Analyzer) ComputeQuantileReturns() map[string]float64 {
	n := len(a.factorValues)
	if n < a.nQuantiles {
		return nil
	}
	type pair struct {
		factor  float64
		ret     float64
	}
	pairs := make([]pair, n)
	for i := 0; i < n; i++ {
		pairs[i] = pair{a.factorValues[i], a.returns[i]}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].factor < pairs[j].factor })

	result := make(map[string]float64)
	bucketSize := n / a.nQuantiles
	for q := 0; q < a.nQuantiles; q++ {
		start := q * bucketSize
		end := start + bucketSize
		if q == a.nQuantiles-1 {
			end = n
		}
		if start >= end {
			continue
		}
		var sum float64
		for j := start; j < end; j++ {
			sum += pairs[j].ret
		}
		name := "Q" + itoa(q+1)
		result[name] = sum / float64(end-start)
	}
	return result
}

func (a *Analyzer) TopMinusBottom() float64 {
	qr := a.ComputeQuantileReturns()
	if len(qr) < 2 {
		return 0
	}
	top := qr["Q"+itoa(a.nQuantiles)]
	bottom := qr["Q1"]
	return top - bottom
}

func (a *Analyzer) FullReport() FactorReport {
	ic := a.ComputeIC("spearman")
	r := FactorReport{
		Name:    a.factorName,
		ICMean:  ic,
		ICStd:   0,
		IR:      0,
	}

	n := len(a.factorValues)
	coverage := 0
	if n > 0 {
		coverage = n
	}
	_ = coverage

	if ic > 0 {
		r.ICPositiveRate = 1
	}

	qr := a.ComputeQuantileReturns()
	r.QuantileReturns = qr
	r.TopMinusBottom = a.TopMinusBottom()

	return r
}

func pearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}
	num := float64(n)*sumXY - sumX*sumY
	denom := math.Sqrt(float64(n)*sumX2-sumX*sumX) * math.Sqrt(float64(n)*sumY2-sumY*sumY)
	if denom == 0 {
		return 0
	}
	return num / denom
}

func rankCorrelation(x, y []float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}
	rankX := rankSlice(x)
	rankY := rankSlice(y)
	return pearsonCorrelation(rankX, rankY)
}

func rankSlice(data []float64) []float64 {
	n := len(data)
	type indexed struct {
		val float64
		idx int
	}
	items := make([]indexed, n)
	for i, v := range data {
		items[i] = indexed{v, i}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].val < items[j].val })

	ranks := make([]float64, n)
	for i := 0; i < n; {
		j := i
		for j < n && items[j].val == items[i].val {
			j++
		}
		avgRank := float64(i+j-1)/2.0 + 1
		for k := i; k < j; k++ {
			ranks[items[k].idx] = avgRank
		}
		i = j
	}
	return ranks
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

package factor

import (
	"github.com/tdx/go-tdx-mcp/indicator"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) ComputeSingle(bars []indicator.Bar, factorNames []string) (map[string][]float64, error) {
	result := make(map[string][]float64)
	for _, name := range factorNames {
		f := Get(name)
		if f == nil {
			continue
		}
		values := make([]float64, len(bars))
		copy(values, f.Compute(bars))
		result[name] = values
	}
	return result, nil
}

type CrossSectionRow struct {
	Date    string
	Code    string
	Factors map[string]float64
}

func (e *Engine) ComputeCrossSection(data map[string][]indicator.Bar, factorNames []string) []CrossSectionRow {
	var rows []CrossSectionRow
	for code, bars := range data {
		if len(bars) == 0 {
			continue
		}
		factorResults, _ := e.ComputeSingle(bars, factorNames)
		factorVals := make(map[string]float64)
		for name, vals := range factorResults {
			if len(vals) > 0 {
				factorVals[name] = vals[len(vals)-1]
			}
		}
		rows = append(rows, CrossSectionRow{
			Code:    code,
			Factors: factorVals,
		})
	}
	return rows
}

func (e *Engine) ComputeForwardReturns(data map[string][]indicator.Bar, period int) map[string][]float64 {
	result := make(map[string][]float64)
	colName := "forward"
	for code, bars := range data {
		n := len(bars)
		forward := make([]float64, n)
		for i := 0; i < n-period; i++ {
			if bars[i].Close > 0 {
				forward[i] = bars[i+period].Close/bars[i].Close - 1
			}
		}
		result[code] = forward
		_ = colName
	}
	return result
}

package backtest

import (
	"sort"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type RunAllItem struct {
	Code       string
	Market     int
	Bars       []indicator.Bar
	InitialCash float64
	Period     string
}

type RunAllRankItem struct {
	Rank          int     `json:"rank"`
	Strategy      string  `json:"strategy"`
	Code          string  `json:"code"`
	Market        int     `json:"market"`
	TotalTrades   int     `json:"total_trades"`
	TotalReturn   float64 `json:"total_return"`
	AnnualReturn  float64 `json:"annual_return"`
	MaxDrawdown   float64 `json:"max_drawdown"`
	Sharpe        float64 `json:"sharpe"`
	WinRate       float64 `json:"win_rate"`
	ProfitFactor  float64 `json:"profit_factor"`
	BarCount      int     `json:"bar_count"`
	LastClose     float64 `json:"last_close"`
	LastBarDate   string  `json:"last_bar_date"`
}

type RunAllResult struct {
	Code        string           `json:"code"`
	Market      int              `json:"market"`
	InitialCash float64          `json:"initial_cash"`
	BarCount    int              `json:"bar_count"`
	LastClose   float64          `json:"last_close"`
	LastBarDate string           `json:"last_bar_date"`
	StrategyCount int            `json:"strategy_count"`
	SortBy      string           `json:"sort_by"`
	SortReverse bool             `json:"sort_reverse"`
	Results     []RunAllRankItem `json:"results"`
}

func RunAll(item RunAllItem) *RunAllResult {
	if len(item.Bars) < 3 {
		return &RunAllResult{Code: item.Code, Market: item.Market, InitialCash: item.InitialCash}
	}
	if item.InitialCash <= 0 {
		item.InitialCash = 1000000
	}

	avail := AvailableStrategies()
	results := make([]RunAllRankItem, 0, len(avail))

	for _, sName := range avail {
		st := NewStrategyWithParams(sName, nil)
		if st == nil {
			continue
		}
		engine := NewEngine(item.InitialCash)
		res := engine.Run(st, item.Bars)

		lastClose := item.Bars[len(item.Bars)-1].Close
		lastDate := item.Bars[len(item.Bars)-1].Date

		results = append(results, RunAllRankItem{
			Strategy:      st.Name(),
			Code:          item.Code,
			Market:        item.Market,
			TotalTrades:   res.Performance.TotalTrades,
			TotalReturn:   res.Performance.TotalReturn,
			AnnualReturn:  res.Performance.AnnualReturn,
			MaxDrawdown:   res.Performance.MaxDrawdown,
			Sharpe:        res.Performance.Sharpe,
			WinRate:       res.Performance.WinRate,
			ProfitFactor:  res.Performance.ProfitFactor,
			BarCount:      res.BarCount,
			LastClose:     lastClose,
			LastBarDate:   lastDate,
		})
	}

	sortBy := "sharpe"
	sortReverse := true

	sort.Slice(results, func(i, j int) bool {
		var a, b float64
		switch sortBy {
		case "sharpe":
			a, b = results[i].Sharpe, results[j].Sharpe
		case "total_return":
			a, b = results[i].TotalReturn, results[j].TotalReturn
		case "annual_return":
			a, b = results[i].AnnualReturn, results[j].AnnualReturn
		case "max_drawdown":
			a, b = results[i].MaxDrawdown, results[j].MaxDrawdown
		case "win_rate":
			a, b = results[i].WinRate, results[j].WinRate
		case "profit_factor":
			a, b = results[i].ProfitFactor, results[j].ProfitFactor
		default:
			a, b = results[i].Sharpe, results[j].Sharpe
		}
		if sortReverse {
			return a > b
		}
		return a < b
	})

	for i := range results {
		results[i].Rank = i + 1
	}

	return &RunAllResult{
		Code:          item.Code,
		Market:        item.Market,
		InitialCash:   item.InitialCash,
		BarCount:      len(item.Bars),
		LastClose:     item.Bars[len(item.Bars)-1].Close,
		LastBarDate:   item.Bars[len(item.Bars)-1].Date,
		StrategyCount: len(results),
		SortBy:        sortBy,
		SortReverse:   sortReverse,
		Results:       results,
	}
}

func RunAllWithConfig(item RunAllItem, sortBy string, sortReverse bool) *RunAllResult {
	result := RunAll(item)
	result.SortBy = sortBy
	result.SortReverse = sortReverse
	sort.Slice(result.Results, func(i, j int) bool {
		var a, b float64
		switch sortBy {
		case "sharpe":
			a, b = result.Results[i].Sharpe, result.Results[j].Sharpe
		case "total_return":
			a, b = result.Results[i].TotalReturn, result.Results[j].TotalReturn
		case "annual_return":
			a, b = result.Results[i].AnnualReturn, result.Results[j].AnnualReturn
		case "max_drawdown":
			a, b = result.Results[i].MaxDrawdown, result.Results[j].MaxDrawdown
		case "win_rate":
			a, b = result.Results[i].WinRate, result.Results[j].WinRate
		case "profit_factor":
			a, b = result.Results[i].ProfitFactor, result.Results[j].ProfitFactor
		default:
			a, b = result.Results[i].Sharpe, result.Results[j].Sharpe
		}
		if sortReverse {
			return a > b
		}
		return a < b
	})
	for i := range result.Results {
		result.Results[i].Rank = i + 1
	}
	return result
}

package backtest

import (
	"sort"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type SignalRankItem struct {
	Rank           int               `json:"rank"`
	Code           string            `json:"code"`
	Market         int               `json:"market"`
	Name           *string           `json:"name"`
	Strategy       string            `json:"strategy"`
	StrategyName   string            `json:"strategy_name"`
	Params         map[string]float64 `json:"params"`
	Period         string            `json:"period"`
	BarCount       int               `json:"bar_count"`
	TotalTrades    int               `json:"total_trades"`
	TotalReturn    float64           `json:"total_return"`
	AnnualReturn   float64           `json:"annual_return"`
	MaxDrawdown    float64           `json:"max_drawdown"`
	Sharpe         float64           `json:"sharpe"`
	WinRate        float64           `json:"win_rate"`
	ProfitFactor   float64           `json:"profit_factor"`
	LastClose      *float64          `json:"last_close"`
	LastBarDate    *string           `json:"last_bar_date"`
	Error          *string           `json:"error"`
}

type SignalRankResult struct {
	Strategy     string           `json:"strategy"`
	StrategyName string           `json:"strategy_name"`
	Params       map[string]float64 `json:"params"`
	Period       string           `json:"period"`
	SortBy       string           `json:"sort_by"`
	SortReverse  bool             `json:"sort_reverse"`
	Total        int              `json:"total"`
	SuccessCount int              `json:"success_count"`
	ErrorCount   int              `json:"error_count"`
	Results      []SignalRankItem `json:"results"`
}

type ScanTarget struct {
	Strategy      string
	StrategyLabel string
	Params        map[string]float64
	Code          string
	Market        int
	Bars          []indicator.Bar
	Window        int
	StrategyID    string
	StrategyName  string
	Kind          string
	Category      string
	Error         string
}

type ScanSignal struct {
	Date      string `json:"date"`
	Direction string `json:"direction"`
}

type ScanRow struct {
	StrategyID    string              `json:"strategy_id"`
	StrategyName  string              `json:"strategy_name"`
	Kind          string              `json:"kind"`
	Strategy      string              `json:"strategy"`
	StrategyLabel string              `json:"strategy_label"`
	Params        map[string]float64  `json:"params"`
	Code          string              `json:"code"`
	Market        int                 `json:"market"`
	Category      string              `json:"category"`
	LatestSignal  *string             `json:"latest_signal"`
	SignalDate    *string             `json:"signal_date"`
	RecentSignals []ScanSignal        `json:"recent_signals"`
	Position      *string             `json:"position"`
	LastClose     *float64            `json:"last_close"`
	LastBarDate   *string             `json:"last_bar_date"`
	Error         *string             `json:"error"`
}

type SignalScanResult struct {
	Rows       []ScanRow `json:"rows"`
	Total      int       `json:"total"`
	BuyCount   int       `json:"buy_count"`
	SellCount  int       `json:"sell_count"`
	ErrorCount int       `json:"error_count"`
}

const scanCommission = 0.0003

type RankInput struct {
	Code   string
	Market int
	Name   *string
	Bars   []indicator.Bar
}

func RunSignalRank(items []RankInput, strategyName string, params map[string]float64, period string) *SignalRankResult {
	result := &SignalRankResult{
		Strategy:     strategyName,
		Params:       params,
		Period:       period,
		SortBy:       "sharpe",
		SortReverse:  false,
		Results:      make([]SignalRankItem, 0),
	}

	rankItems := make([]SignalRankItem, 0, len(items))
	for _, item := range items {
		bars := item.Bars
		if len(bars) < 3 {
			ri := SignalRankItem{
				Code:    item.Code,
				Market:  item.Market,
				Name:    item.Name,
				Strategy: strategyName,
				Period:   period,
				BarCount: len(bars),
				Error:    strPtr("K线数据不足"),
			}
			rankItems = append(rankItems, ri)
			result.ErrorCount++
			continue
		}

		st := NewStrategyWithParams(strategyName, params)
		if st == nil {
			ri := SignalRankItem{
				Code:    item.Code,
				Market:  item.Market,
				Name:    item.Name,
				Strategy: strategyName,
				Period:   period,
				BarCount: len(bars),
				Error:    strPtr("未知策略"),
			}
			rankItems = append(rankItems, ri)
			result.ErrorCount++
			continue
		}
		result.StrategyName = st.Name()

		engine := NewEngine(1000000)
		bt := engine.Run(st, bars)

		ri := SignalRankItem{
			Code:           item.Code,
			Market:         item.Market,
			Name:           item.Name,
			Strategy:       strategyName,
			Params:         params,
			Period:         period,
			BarCount:       bt.BarCount,
			TotalTrades:    bt.Performance.TotalTrades,
			TotalReturn:    bt.Performance.TotalReturn,
			AnnualReturn:   bt.Performance.AnnualReturn,
			MaxDrawdown:    bt.Performance.MaxDrawdown,
			Sharpe:         bt.Performance.Sharpe,
			WinRate:        bt.Performance.WinRate,
			ProfitFactor:   bt.Performance.ProfitFactor,
		}
		if len(bars) > 0 {
			ri.LastClose = &bars[len(bars)-1].Close
			ri.LastBarDate = strPtr(bars[len(bars)-1].Date)
		}
		if bt.Performance.TotalTrades == 0 {
			ri.Error = strPtr("无交易信号")
			result.ErrorCount++
		} else {
			result.SuccessCount++
		}
		rankItems = append(rankItems, ri)
	}

	result.Results = rankItems
	result.Total = len(rankItems)
	result.SortBy = "sharpe"

	sort.Slice(rankItems, func(i, j int) bool {
		var a, b float64
		switch result.SortBy {
		case "sharpe":
			a = rankItems[i].Sharpe
			b = rankItems[j].Sharpe
		case "total_return":
			a = rankItems[i].TotalReturn
			b = rankItems[j].TotalReturn
		case "annual_return":
			a = rankItems[i].AnnualReturn
			b = rankItems[j].AnnualReturn
		case "max_drawdown":
			a = rankItems[i].MaxDrawdown
			b = rankItems[j].MaxDrawdown
		case "win_rate":
			a = rankItems[i].WinRate
			b = rankItems[j].WinRate
		case "profit_factor":
			a = rankItems[i].ProfitFactor
			b = rankItems[j].ProfitFactor
		default:
			a = rankItems[i].Sharpe
			b = rankItems[j].Sharpe
		}
		if result.SortReverse {
			return a > b
		}
		return a < b
	})

	for i := range rankItems {
		rankItems[i].Rank = i + 1
	}
	result.Results = rankItems
	return result
}

func RunSignalScan(targets []ScanTarget, defaultWindow int) *SignalScanResult {
	if defaultWindow <= 0 {
		defaultWindow = 30
	}
	rows := make([]ScanRow, 0, len(targets))
	buyCount, sellCount, errorCount := 0, 0, 0

	for _, t := range targets {
		row := ScanRow{
			StrategyID:    t.StrategyID,
			StrategyName:  t.StrategyName,
			Kind:          t.Kind,
			Strategy:      t.Strategy,
			StrategyLabel: t.StrategyLabel,
			Params:        t.Params,
			Code:          t.Code,
			Market:        t.Market,
			Category:      t.Category,
		}
		if t.Error != "" {
			row.Error = strPtr(t.Error)
			errorCount++
			rows = append(rows, row)
			continue
		}

		bars := t.Bars
		w := t.Window
		if w <= 0 {
			w = defaultWindow
		}
		if len(bars) < 2 {
			row.Error = strPtr("K线数据不足")
			errorCount++
			rows = append(rows, row)
			continue
		}

		st := NewStrategyWithParams(t.Strategy, t.Params)
		if st == nil {
			row.Error = strPtr("未知策略: " + t.Strategy)
			errorCount++
			rows = append(rows, row)
			continue
		}

		recent := make([]ScanSignal, 0)
		var position float64
		startIdx := len(bars) - w
		if startIdx < 0 {
			startIdx = 0
		}

		for i := 1; i < len(bars); i++ {
			sig := st.Next(i, bars)
			if i >= startIdx && sig != Hold {
				recent = append(recent, ScanSignal{
					Date:      bars[i].Date,
					Direction: signalDir(sig),
				})
			}
			if sig == Buy && position == 0 {
				price := bars[i].Close * (1 + scanCommission)
				if price > 0 {
					position = 100000 / price
				}
			} else if sig == Sell && position > 0 {
				position = 0
			}
		}

		latestSig := ""
		sigDate := ""
		if len(recent) > 0 {
			latestSig = recent[len(recent)-1].Direction
			sigDate = recent[len(recent)-1].Date
		}
		pos := "flat"
		if position > 0 {
			pos = "holding"
		}
		lastClose := bars[len(bars)-1].Close
		lastDate := bars[len(bars)-1].Date

		row.LatestSignal = strPtr(latestSig)
		row.SignalDate = strPtr(sigDate)
		row.RecentSignals = recent
		row.Position = strPtr(pos)
		row.LastClose = &lastClose
		row.LastBarDate = strPtr(lastDate)

		for _, s := range recent {
			if s.Direction == "BUY" {
				buyCount++
			} else {
				sellCount++
			}
		}
		rows = append(rows, row)
	}

	return &SignalScanResult{
		Rows:       rows,
		Total:      len(rows),
		BuyCount:   buyCount,
		SellCount:  sellCount,
		ErrorCount: errorCount,
	}
}

func signalDir(s Signal) string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}

func strPtr(s string) *string {
	return &s
}

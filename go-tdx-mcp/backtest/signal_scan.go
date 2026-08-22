package backtest

import "github.com/tdx/go-tdx-mcp/indicator"

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

package backtest

import (
	"github.com/tdx/go-tdx-mcp/indicator"
)

type PerfMetrics struct {
	TotalReturn      float64 `json:"total_return"`
	TotalReturnPct   float64 `json:"total_return_pct"`
	CAGR             float64 `json:"cagr"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	MaxDrawdownPct   float64 `json:"max_drawdown_pct"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	SortinoRatio     float64 `json:"sortino_ratio"`
	CalmarRatio      float64 `json:"calmar_ratio"`
	AnnualVolatility float64 `json:"annual_volatility"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	AvgWin           float64 `json:"avg_win"`
	AvgLoss          float64 `json:"avg_loss"`
	MaxConsecutiveWins int   `json:"max_consecutive_wins"`
	MaxConsecutiveLosses int `json:"max_consecutive_losses"`
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
	EquityCurve      []float64 `json:"equity_curve"`
}

func CalcExtendedPerformance(initial, final float64, trades []Trade, bars int, equityCurve []float64) PerfMetrics {
	p := PerfMetrics{}
	p.EquityCurve = equityCurve
	// 复用 engine 绩效计算（基于逐 bar 资金曲线，对齐 easy_tdx 语义：比例口径、日收益率）
	perf := calcPerformance(equityCurve, trades)
	p.TotalReturn = perf.TotalReturn
	p.TotalReturnPct = perf.TotalReturnPct
	p.CAGR = perf.CAGR
	p.MaxDrawdown = perf.MaxDrawdown
	p.MaxDrawdownPct = perf.MaxDrawdownPct
	p.SharpeRatio = perf.SharpeRatio
	p.SortinoRatio = perf.Sortino
	p.CalmarRatio = perf.Calmar
	p.AnnualVolatility = perf.AnnualVolatility
	p.WinRate = perf.WinRate
	p.ProfitFactor = perf.ProfitFactor
	p.AvgWin = perf.AvgWin
	p.AvgLoss = perf.AvgLoss
	p.TotalTrades = perf.TotalTrades
	p.WinningTrades = perf.WinningTrades
	p.LosingTrades = perf.LosingTrades

	// 连续盈亏统计（PerfMetrics 特有）
	var maxConWins, maxConLosses, conWins, conLosses int
	for _, t := range trades {
		if t.Profit > 0 {
			conWins++
			conLosses = 0
			if conWins > maxConWins {
				maxConWins = conWins
			}
		} else {
			conLosses++
			conWins = 0
			if conLosses > maxConLosses {
				maxConLosses = conLosses
			}
		}
	}
	p.MaxConsecutiveWins = maxConWins
	p.MaxConsecutiveLosses = maxConLosses
	return p
}

type ComboResult struct {
	Mode     string   `json:"mode"`
	Results  []Result `json:"results"`
	Signals  []string `json:"signals"`
}

type ComboMode string

const (
	ComboAnd       ComboMode = "and"
	ComboOr        ComboMode = "or"
	ComboMajority  ComboMode = "majority"
)

func RunCombo(engine *Engine, strategies []Strategy, bars []indicator.Bar, mode ComboMode) *ComboResult {
	if len(strategies) == 0 || len(bars) < 2 {
		return &ComboResult{Mode: string(mode)}
	}

	signals := make([][]Signal, len(strategies))
	signalNames := make([]string, len(strategies))
	for sIdx, s := range strategies {
		signals[sIdx] = make([]Signal, len(bars))
		signalNames[sIdx] = s.Name()
		for i := 1; i < len(bars); i++ {
			signals[sIdx][i] = s.Next(i, bars)
		}
	}

	merged := make([]Signal, len(bars))
	for i := 1; i < len(bars); i++ {
		switch mode {
		case ComboAnd:
			merged[i] = mergeAnd(signals, i)
		case ComboOr:
			merged[i] = mergeOr(signals, i)
		case ComboMajority:
			merged[i] = mergeMajority(signals, i)
		}
	}

	comboStrategy := &comboStrategy{name: "combo_" + string(mode), signals: merged}
	result := engine.Run(comboStrategy, bars)

	r := &ComboResult{
		Mode:    string(mode),
		Results: []Result{*result},
		Signals: signalNames,
	}
	return r
}

func mergeAnd(signals [][]Signal, i int) Signal {
	buyCount, sellCount := 0, 0
	for _, s := range signals {
		if i < len(s) {
			switch s[i] {
			case Buy:
				buyCount++
			case Sell:
				sellCount++
			}
		}
	}
	total := len(signals)
	if buyCount == total {
		return Buy
	}
	if sellCount == total {
		return Sell
	}
	return Hold
}

func mergeOr(signals [][]Signal, i int) Signal {
	buyCount, sellCount := 0, 0
	for _, s := range signals {
		if i < len(s) {
			switch s[i] {
			case Buy:
				buyCount++
			case Sell:
				sellCount++
			}
		}
	}
	if buyCount > 0 {
		return Buy
	}
	if sellCount > 0 {
		return Sell
	}
	return Hold
}

func mergeMajority(signals [][]Signal, i int) Signal {
	buyCount, sellCount := 0, 0
	for _, s := range signals {
		if i < len(s) {
			switch s[i] {
			case Buy:
				buyCount++
			case Sell:
				sellCount++
			}
		}
	}
	threshold := len(signals) / 2
	if buyCount > threshold {
		return Buy
	}
	if sellCount > threshold {
		return Sell
	}
	return Hold
}

type comboStrategy struct {
	name    string
	signals []Signal
}

func (s *comboStrategy) Name() string                         { return s.name }
func (s *comboStrategy) Next(i int, bars []indicator.Bar) Signal { return s.signals[i] }

package backtest

import (
	"math"

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
	p.TotalReturn = final - initial
	p.TotalReturnPct = p.TotalReturn / initial * 100
	p.TotalTrades = len(trades)
	p.EquityCurve = equityCurve

	if bars > 0 {
		years := float64(bars) / 252
		if years > 0 && final > initial {
			p.CAGR = (math.Pow(final/initial, 1/years) - 1) * 100
		}
	}

	var wins, losses int
	var totalWin, totalLoss, totalReturnVal float64
	var returns []float64
	var consecutiveWins, consecutiveLosses, maxConWins, maxConLosses int

	for _, t := range trades {
		returns = append(returns, t.ReturnPct)
		returns = append(returns, t.ReturnPct)
		totalReturnVal += t.ReturnPct
		if t.Profit > 0 {
			wins++
			totalWin += t.Profit
			consecutiveWins++
			consecutiveLosses = 0
			if consecutiveWins > maxConWins {
				maxConWins = consecutiveWins
			}
		} else {
			losses++
			totalLoss += math.Abs(t.Profit)
			consecutiveLosses++
			consecutiveWins = 0
			if consecutiveLosses > maxConLosses {
				maxConLosses = consecutiveLosses
			}
		}
	}
	p.WinningTrades = wins
	p.LosingTrades = losses
	p.MaxConsecutiveWins = maxConWins
	p.MaxConsecutiveLosses = maxConLosses
	if p.TotalTrades > 0 {
		p.WinRate = float64(wins) / float64(p.TotalTrades) * 100
	}
	if wins > 0 {
		p.AvgWin = totalWin / float64(wins)
	}
	if losses > 0 {
		p.AvgLoss = totalLoss / float64(losses)
	}
	if totalLoss > 0 {
		p.ProfitFactor = totalWin / totalLoss
	} else if totalWin > 0 {
		p.ProfitFactor = 999
	}

	if len(returns) > 1 {
		mean := meanFloat(returns)
		std := stdFloat(returns, mean)
		if std > 0 {
			p.SharpeRatio = mean / std * math.Sqrt(252)
		}
		p.AnnualVolatility = std * math.Sqrt(252)

		downsideReturns := make([]float64, 0)
		for _, r := range returns {
			if r < 0 {
				downsideReturns = append(downsideReturns, r)
			}
		}
		if len(downsideReturns) > 0 {
			downMean := meanFloat(downsideReturns)
			downStd := stdFloat(downsideReturns, downMean)
			if downStd > 0 {
				p.SortinoRatio = mean / downStd * math.Sqrt(252)
			}
		}
	}

	if p.MaxDrawdownPct > 0 {
		p.CalmarRatio = p.TotalReturnPct / p.MaxDrawdownPct
	}

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

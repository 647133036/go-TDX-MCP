package backtest

import (
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type Signal int

const (
	Hold Signal = 0
	Buy  Signal = 1
	Sell Signal = -1
)

type Strategy interface {
	Name() string
	Next(i int, bars []indicator.Bar) Signal
}

type Trade struct {
	EntryPrice float64 `json:"entry_price"`
	ExitPrice  float64 `json:"exit_price"`
	Size       int     `json:"size"`
	Profit     float64 `json:"profit"`
	ReturnPct  float64 `json:"return_pct"`
	Side       string  `json:"side"`
}

type Performance struct {
	TotalReturn      float64 `json:"total_return"`
	TotalReturnPct   float64 `json:"total_return_pct"`
	CAGR             float64 `json:"cagr"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	MaxDrawdownPct   float64 `json:"max_drawdown_pct"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	AnnualVolatility float64 `json:"annual_volatility"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	AvgWin           float64 `json:"avg_win"`
	AvgLoss          float64 `json:"avg_loss"`
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
}

type Result struct {
	Strategy    string        `json:"strategy"`
	Code        string        `json:"code"`
	Market      int           `json:"market"`
	Period      string        `json:"period"`
	InitialCash float64       `json:"initial_cash"`
	FinalEquity float64       `json:"final_equity"`
	BarCount    int           `json:"bar_count"`
	Performance Performance   `json:"performance"`
	Trades      []Trade       `json:"trades"`
}

type Engine struct {
	cash       float64
	commission float64
	slippage   float64
}

func NewEngine(initialCash float64) *Engine {
	return &Engine{
		cash:       initialCash,
		commission: 0.0003,
		slippage:   0.001,
	}
}

func (e *Engine) SetCommission(v float64) { e.commission = v }
func (e *Engine) SetSlippage(v float64)   { e.slippage = v }

func (e *Engine) Run(strategy Strategy, bars []indicator.Bar) *Result {
	if len(bars) < 2 {
		return &Result{Strategy: strategy.Name(), InitialCash: e.cash, FinalEquity: e.cash}
	}

	position := 0
	cash := e.cash
	var trades []Trade
	var openTrade *Trade

	for i := 1; i < len(bars); i++ {
		sig := strategy.Next(i, bars)

		switch sig {
		case Buy:
			if position == 0 && cash > 0 {
				price := bars[i].Close * (1 + e.slippage)
				shares := int(cash * (1 - e.commission) / price)
				if shares >= 100 {
					cost := float64(shares)*price*(1+e.commission) + 5
					cash -= cost
					position = shares
					openTrade = &Trade{
						EntryPrice: price,
						Size:       shares,
						Side:       "long",
					}
				}
			}
		case Sell:
			if position > 0 && openTrade != nil {
				price := bars[i].Close * (1 - e.slippage)
				proceeds := float64(position)*price*(1-e.commission) - 5
				profit := proceeds - float64(position)*openTrade.EntryPrice*(1+e.commission)
				cash += proceeds
				openTrade.ExitPrice = price
				openTrade.Profit = profit
				openTrade.ReturnPct = profit / (float64(position)*openTrade.EntryPrice*(1+e.commission)) * 100
				trades = append(trades, *openTrade)
				openTrade = nil
				position = 0
			}
		}
	}

	if position > 0 && openTrade != nil {
		price := bars[len(bars)-1].Close
		proceeds := float64(position) * price
		cash += proceeds
		openTrade.ExitPrice = price
		openTrade.Profit = proceeds - float64(position)*openTrade.EntryPrice*(1+e.commission)
		openTrade.ReturnPct = openTrade.Profit / (float64(position)*openTrade.EntryPrice*(1+e.commission)) * 100
		trades = append(trades, *openTrade)
	}

	finalEquity := cash
	perf := calcPerformance(e.cash, finalEquity, trades, len(bars))

	return &Result{
		Strategy:    strategy.Name(),
		InitialCash: e.cash,
		FinalEquity: finalEquity,
		BarCount:    len(bars),
		Performance: perf,
		Trades:      trades,
	}
}

func calcPerformance(initial, final float64, trades []Trade, bars int) Performance {
	p := Performance{}
	p.TotalReturn = final - initial
	p.TotalReturnPct = p.TotalReturn / initial * 100
	p.TotalTrades = len(trades)

	if bars > 0 {
		years := float64(bars) / 252
		if years > 0 && final > initial {
			p.CAGR = (math.Pow(final/initial, 1/years) - 1) * 100
		}
	}

	var wins, losses int
	var totalWin, totalLoss float64
	var returns []float64
	for _, t := range trades {
		returns = append(returns, t.ReturnPct)
		if t.Profit > 0 {
			wins++
			totalWin += t.Profit
		} else {
			losses++
			totalLoss += math.Abs(t.Profit)
		}
	}
	p.WinningTrades = wins
	p.LosingTrades = losses
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
	}

	peak := initial
	maxDD := 0.0
	equity := initial
	for _, t := range trades {
		equity += t.Profit
		if equity > peak {
			peak = equity
		}
		dd := peak - equity
		if dd > maxDD {
			maxDD = dd
		}
	}
	if final < peak {
		dd := peak - final
		if dd > maxDD {
			maxDD = dd
		}
	}
	p.MaxDrawdown = maxDD
	p.MaxDrawdownPct = maxDD / initial * 100

	return p
}

func meanFloat(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func stdFloat(v []float64, mean float64) float64 {
	if len(v) <= 1 {
		return 0
	}
	sum := 0.0
	for _, x := range v {
		d := x - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(v)-1))
}

// --- Built-in Strategies ---

type MACrossStrategy struct {
	FastPeriod int
	SlowPeriod int
}

func (s *MACrossStrategy) Name() string { return "ma_cross" }

func (s *MACrossStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.SlowPeriod+1 {
		return Hold
	}
	closes := extractClose(bars)
	fast := indicator.MA(closes, s.FastPeriod)
	slow := indicator.MA(closes, s.SlowPeriod)
	if fast[i] <= 0 || slow[i] <= 0 || fast[i-1] <= 0 || slow[i-1] <= 0 {
		return Hold
	}
	if fast[i-1] <= slow[i-1] && fast[i] > slow[i] {
		return Buy
	}
	if fast[i-1] >= slow[i-1] && fast[i] < slow[i] {
		return Sell
	}
	return Hold
}

type MACDCrossStrategy struct {
	Fast   int
	Slow   int
	Signal int
}

func (s *MACDCrossStrategy) Name() string { return "macd_cross" }

func (s *MACDCrossStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Slow+s.Signal {
		return Hold
	}
	r := indicator.MACD(bars, s.Fast, s.Slow, s.Signal)
	dif := r.Values
	dea := r.Line2
	if i >= len(dif) || i >= len(dea) {
		return Hold
	}
	if dif[i] <= 0 || dea[i] <= 0 {
		return Hold
	}
	if dif[i-1] <= dea[i-1] && dif[i] > dea[i] {
		return Buy
	}
	if dif[i-1] >= dea[i-1] && dif[i] < dea[i] {
		return Sell
	}
	return Hold
}

type RSIReversalStrategy struct {
	Period     int
	Oversold   float64
	Overbought float64
}

func (s *RSIReversalStrategy) Name() string { return "rsi_reversal" }

func (s *RSIReversalStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period {
		return Hold
	}
	r := indicator.RSI(bars, s.Period)
	rsi := r.Values
	if i >= len(rsi) {
		return Hold
	}
	if rsi[i] <= 0 {
		return Hold
	}
	if rsi[i-1] <= s.Oversold && rsi[i] > s.Oversold {
		return Buy
	}
	if rsi[i-1] >= s.Overbought && rsi[i] < s.Overbought {
		return Sell
	}
	return Hold
}

type BollingerBreakoutStrategy struct {
	Period     int
	Multiplier float64
}

func (s *BollingerBreakoutStrategy) Name() string { return "bollinger_breakout" }

func (s *BollingerBreakoutStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period {
		return Hold
	}
	r := indicator.BOLL(bars, s.Period, s.Multiplier)
	mid := r.Values
	upper := r.Line2
	lower := r.Line3
	if i >= len(mid) || i >= len(upper) || i >= len(lower) {
		return Hold
	}
	if mid[i] <= 0 || upper[i] <= 0 || lower[i] <= 0 {
		return Hold
	}
	if bars[i-1].Close <= upper[i-1] && bars[i].Close > upper[i] {
		return Buy
	}
	if bars[i-1].Close >= mid[i-1] && bars[i].Close < mid[i] {
		return Sell
	}
	return Hold
}

type EXPMAStrategy struct {
	FastPeriod int
	SlowPeriod int
}

func (s *EXPMAStrategy) Name() string { return "expma_cross" }

func (s *EXPMAStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.SlowPeriod {
		return Hold
	}
	fastR := indicator.EXPMA(bars, s.FastPeriod)
	slowR := indicator.EXPMA(bars, s.SlowPeriod)
	fast := fastR.Values
	slow := slowR.Values
	if i >= len(fast) || i >= len(slow) {
		return Hold
	}
	if fast[i] <= 0 || slow[i] <= 0 || fast[i-1] <= 0 || slow[i-1] <= 0 {
		return Hold
	}
	if fast[i-1] <= slow[i-1] && fast[i] > slow[i] {
		return Buy
	}
	if fast[i-1] >= slow[i-1] && fast[i] < slow[i] {
		return Sell
	}
	return Hold
}

type KDJGoldenStrategy struct {
	Period  int
	KPeriod int
	DPeriod int
}

func (s *KDJGoldenStrategy) Name() string { return "kdj_golden" }

func (s *KDJGoldenStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+s.DPeriod {
		return Hold
	}
	r := indicator.KDJ(bars, s.Period, s.KPeriod, s.DPeriod)
	k := r.Values
	d := r.Line2
	if i >= len(k) || i >= len(d) {
		return Hold
	}
	if k[i] <= 0 || d[i] <= 0 || k[i-1] <= 0 || d[i-1] <= 0 {
		return Hold
	}
	if k[i-1] <= d[i-1] && k[i] > d[i] && k[i] < 30 {
		return Buy
	}
	if k[i-1] >= d[i-1] && k[i] < d[i] && k[i] > 70 {
		return Sell
	}
	return Hold
}

type TurtleBreakoutStrategy struct {
	EntryPeriod int
	ExitPeriod  int
}

func (s *TurtleBreakoutStrategy) Name() string { return "turtle_breakout" }

func (s *TurtleBreakoutStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.EntryPeriod || i < 2 {
		return Hold
	}
	highEntry := maxHigh(bars, i-1, s.EntryPeriod)
	lowExit := minLow(bars, i-1, s.ExitPeriod)
	if bars[i].Close > highEntry && bars[i-1].Close <= highEntry {
		return Buy
	}
	if bars[i].Close < lowExit && bars[i-1].Close >= lowExit {
		return Sell
	}
	return Hold
}

func maxHigh(bars []indicator.Bar, end, period int) float64 {
	start := end - period + 1
	if start < 0 {
		start = 0
	}
	m := bars[start].High
	for j := start + 1; j <= end && j < len(bars); j++ {
		if bars[j].High > m {
			m = bars[j].High
		}
	}
	return m
}

func minLow(bars []indicator.Bar, end, period int) float64 {
	start := end - period + 1
	if start < 0 {
		start = 0
	}
	m := bars[start].Low
	for j := start + 1; j <= end && j < len(bars); j++ {
		if bars[j].Low < m {
			m = bars[j].Low
		}
	}
	return m
}

func extractClose(bars []indicator.Bar) []float64 {
	r := make([]float64, len(bars))
	for i, b := range bars {
		r[i] = b.Close
	}
	return r
}

func NewStrategy(name string) Strategy {
	switch name {
	case "ma_cross":
		return &MACrossStrategy{FastPeriod: 5, SlowPeriod: 20}
	case "macd_cross":
		return &MACDCrossStrategy{Fast: 12, Slow: 26, Signal: 9}
	case "rsi_reversal":
		return &RSIReversalStrategy{Period: 14, Oversold: 30, Overbought: 70}
	case "bollinger_breakout":
		return &BollingerBreakoutStrategy{Period: 20, Multiplier: 2}
	case "expma_cross":
		return &EXPMAStrategy{FastPeriod: 5, SlowPeriod: 20}
	case "kdj_golden":
		return &KDJGoldenStrategy{Period: 9, KPeriod: 3, DPeriod: 3}
	case "turtle_breakout":
		return &TurtleBreakoutStrategy{EntryPeriod: 20, ExitPeriod: 10}
	case "bias_reversal":
		return &BiasReversalStrategy{Period: 6, Oversold: -3, Overbought: 3}
	case "volume_price":
		return &VolumePriceStrategy{}
	case "dmi_trend":
		return &DMITrendStrategy{Period: 14, SignalPeriod: 6, ADXThreshold: 25}
	case "cci_breakout":
		return &CCIBreakoutStrategy{Period: 14, Overbought: 100, Oversold: -100}
	case "mfi_volume":
		return &MFIVolumeStrategy{Period: 14, Oversold: 20, Overbought: 80}
	case "zhuoyao_momentum":
		return &ZhuoYaoMomentumStrategy{}
	case "trix_cross":
		return &TRIXCrossStrategy{Period: 12, SignalPeriod: 9}
	case "mtm_momentum":
		return &MTMMomentumStrategy{Period: 6}
	case "obv_trend":
		return &OBVTrendStrategy{Period: 20}
	default:
		return nil
	}
}

func AvailableStrategies() []string {
	return []string{
		"ma_cross", "macd_cross", "rsi_reversal", "bollinger_breakout",
		"expma_cross", "kdj_golden", "turtle_breakout",
		"bias_reversal", "volume_price", "dmi_trend", "cci_breakout",
		"mfi_volume", "zhuoyao_momentum", "trix_cross", "mtm_momentum", "obv_trend",
	}
}

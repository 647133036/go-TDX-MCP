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
	// 核心字段（对齐 easy_tdx 语义：比例口径，0~1）
	TotalReturn   float64 `json:"total_return"`   // 总收益率 (final/initial - 1)
	AnnualReturn  float64 `json:"annual_return"`  // 年化收益率
	MaxDrawdown   float64 `json:"max_drawdown"`   // 最大回撤 (0~1 正值)
	MaxDDDuration int     `json:"max_dd_duration"` // 最大回撤持续 bar 数
	Sharpe        float64 `json:"sharpe"`         // 夏普比率（基于日收益率）
	Sortino       float64 `json:"sortino"`        // 索提诺比率
	Calmar        float64 `json:"calmar"`          // 卡玛比率
	TotalTrades   int     `json:"total_trades"`    // 总交易次数（卖出次数）
	WinTrades     int     `json:"win_trades"`     // 盈利交易次数
	LoseTrades    int     `json:"lose_trades"`     // 亏损交易次数
	WinRate       float64 `json:"win_rate"`       // 胜率 (0~1)
	ProfitFactor  float64 `json:"profit_factor"`  // 盈亏比
	AvgWin        float64 `json:"avg_win"`        // 平均盈利（单笔收益率口径）
	AvgLoss       float64 `json:"avg_loss"`       // 平均亏损（单笔收益率口径）
	MaxWin        float64 `json:"max_win"`        // 最大盈利（单笔收益率）
	MaxLoss       float64 `json:"max_loss"`       // 最大亏损（单笔收益率）
	Volatility    float64 `json:"volatility"`     // 年化波动率 (0~1)
	StartCash     float64 `json:"start_cash"`     // 初始资金
	EndValue      float64 `json:"end_value"`      // 期末净值
	// 兼容别名（百分比/旧名，供展示与历史调用方使用）
	TotalReturnPct   float64 `json:"total_return_pct"`
	MaxDrawdownPct   float64 `json:"max_drawdown_pct"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	AnnualVolatility float64 `json:"annual_volatility"`
	CAGR             float64 `json:"cagr"`
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
		eq := []float64{e.cash, e.cash}
		return &Result{Strategy: strategy.Name(), InitialCash: e.cash, FinalEquity: e.cash, Performance: calcPerformance(eq, nil)}
	}

	position := 0
	cash := e.cash
	var trades []Trade
	var openTrade *Trade
	equityCurve := make([]float64, len(bars))
	equityCurve[0] = cash

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
				costBasis := float64(position) * openTrade.EntryPrice * (1 + e.commission)
				profit := proceeds - costBasis
				cash += proceeds
				openTrade.ExitPrice = price
				openTrade.Profit = profit
				openTrade.ReturnPct = profit / costBasis * 100
				trades = append(trades, *openTrade)
				openTrade = nil
				position = 0
			}
		}
		// 逐 bar 记录总权益（现金 + 持仓市值），用于计算日收益率/回撤/夏普
		equityCurve[i] = cash + float64(position)*bars[i].Close
	}

	// 末尾仍有持仓：按最后收盘价虚拟平仓，计入交易统计
	finalEquity := cash
	if position > 0 && openTrade != nil {
		price := bars[len(bars)-1].Close
		proceeds := float64(position) * price
		costBasis := float64(position) * openTrade.EntryPrice * (1 + e.commission)
		profit := proceeds - costBasis
		finalEquity = cash + proceeds
		openTrade.ExitPrice = price
		openTrade.Profit = profit
		openTrade.ReturnPct = profit / costBasis * 100
		trades = append(trades, *openTrade)
		equityCurve[len(bars)-1] = finalEquity
	}

	perf := calcPerformance(equityCurve, trades)

	return &Result{
		Strategy:    strategy.Name(),
		InitialCash: e.cash,
		FinalEquity: finalEquity,
		BarCount:    len(bars),
		Performance: perf,
		Trades:      trades,
	}
}

// calcPerformance 基于逐 bar 资金曲线计算绩效，口径对齐 easy_tdx PerformanceAnalyzer：
// total_return/annual_return/max_drawdown/win_rate/volatility 均为 0~1 比例，
// sharpe/sortino 基于日收益率，avg_win/avg_loss/max_win/max_loss 为单笔收益率口径。
func calcPerformance(equityCurve []float64, trades []Trade) Performance {
	p := Performance{}
	if len(equityCurve) == 0 {
		return p
	}
	initial := equityCurve[0]
	final := equityCurve[len(equityCurve)-1]
	p.StartCash = initial
	p.EndValue = final

	// 边界：资金曲线不足 2 根
	if len(equityCurve) < 2 {
		return p
	}

	// 日收益率（除零保护：前值为 0 跳过）
	var dailyRets []float64
	for i := 1; i < len(equityCurve); i++ {
		prev := equityCurve[i-1]
		if prev != 0 {
			dailyRets = append(dailyRets, (equityCurve[i]-prev)/prev)
		}
	}
	if len(dailyRets) < 2 {
		return p
	}

	// 1. 总收益率（比例）
	if initial != 0 {
		p.TotalReturn = final/initial - 1
	}
	p.TotalReturnPct = p.TotalReturn * 100

	// 2. 年化收益率
	n := len(dailyRets)
	p.AnnualReturn = math.Pow(1+p.TotalReturn, 252.0/float64(n)) - 1
	p.CAGR = p.AnnualReturn * 100

	// 3. 最大回撤（0~1 正值，逐 bar 净值口径）
	peak := equityCurve[0]
	maxDD := 0.0
	maxDDIdx := 0
	peakIdx := 0
	curPeakIdx := 0
	for i := 0; i < len(equityCurve); i++ {
		if equityCurve[i] > peak {
			peak = equityCurve[i]
			curPeakIdx = i
		}
		if peak > 0 {
			dd := (peak - equityCurve[i]) / peak
			if dd > maxDD {
				maxDD = dd
				maxDDIdx = i
				peakIdx = curPeakIdx
			}
		}
	}
	p.MaxDrawdown = maxDD
	p.MaxDrawdownPct = maxDD * 100
	p.MaxDDDuration = maxDDIdx - peakIdx

	// 5. 夏普比率（基于日收益率）
	rfDaily := 0.03 / 252
	meanRet := meanFloat(dailyRets)
	stdRet := stdFloat(dailyRets, meanRet)
	if stdRet > 0 {
		var sumExcess float64
		for _, r := range dailyRets {
			sumExcess += r - rfDaily
		}
		p.Sharpe = (sumExcess / float64(len(dailyRets))) / stdRet * math.Sqrt(252)
	}
	p.SharpeRatio = p.Sharpe

	// 6. 索提诺比率（分母只用负收益标准差）
	var negRets []float64
	for _, r := range dailyRets {
		if r-rfDaily < 0 {
			negRets = append(negRets, r-rfDaily)
		}
	}
	if len(negRets) > 0 {
		negStd := stdFloat(negRets, meanFloat(negRets))
		if negStd > 0 {
			var sumExcess float64
			for _, r := range dailyRets {
				sumExcess += r - rfDaily
			}
			p.Sortino = (sumExcess / float64(len(dailyRets))) / negStd * math.Sqrt(252)
		} else if meanRet-rfDaily > 0 {
			p.Sortino = 999
		}
	} else if meanRet-rfDaily > 0 {
		p.Sortino = 999
	}

	// 7. 卡玛比率
	if p.MaxDrawdown > 1e-10 {
		p.Calmar = p.AnnualReturn / p.MaxDrawdown
	} else if p.AnnualReturn > 0 {
		p.Calmar = 999
	}

	// 年化波动率
	p.Volatility = stdRet * math.Sqrt(252)
	p.AnnualVolatility = p.Volatility

	// 交易统计（单笔收益率口径：ReturnPct 已是百分比，/100 转比例）
	p.TotalTrades = len(trades)
	var wins, losses int
	var winReturns, loseReturns []float64
	var totalWinPnl, totalLossPnl float64
	for _, t := range trades {
		ret := t.ReturnPct / 100
		if t.Profit > 0 {
			wins++
			totalWinPnl += t.Profit
			winReturns = append(winReturns, ret)
		} else {
			losses++
			totalLossPnl += math.Abs(t.Profit)
			loseReturns = append(loseReturns, ret)
		}
	}
	p.WinTrades = wins
	p.LoseTrades = losses
	p.WinningTrades = wins
	p.LosingTrades = losses
	if p.TotalTrades > 0 {
		p.WinRate = float64(wins) / float64(p.TotalTrades)
	}
	if len(winReturns) > 0 {
		p.AvgWin = meanFloat(winReturns)
		p.MaxWin = maxFloat(winReturns)
	}
	if len(loseReturns) > 0 {
		p.AvgLoss = meanFloat(loseReturns)
		p.MaxLoss = minFloat(loseReturns)
	}
	if totalLossPnl > 0 {
		p.ProfitFactor = totalWinPnl / totalLossPnl
	} else if totalWinPnl > 0 {
		p.ProfitFactor = 999
	}

	return p
}

func maxFloat(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func minFloat(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x < m {
			m = x
		}
	}
	return m
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

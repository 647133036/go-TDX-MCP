package factor

import (
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

func init() {
	registerMomentum()
	registerTechnical()
	registerVolume()
	registerVolatility()
	registerChanlun()
	registerValue()
	registerQuality()
}

func registerMomentum() {
	Register("momentum_20d", "momentum", "20-day price momentum", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				if i < 20 {
					out[i] = math.NaN()
					continue
				}
				if bars[i-20].Close > 0 {
					out[i] = bars[i].Close/bars[i-20].Close - 1
				} else {
					out[i] = math.NaN()
				}
			}
			return out
		})

	Register("momentum_60d", "momentum", "60-day price momentum", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				if i < 60 {
					out[i] = math.NaN()
					continue
				}
				if bars[i-60].Close > 0 {
					out[i] = bars[i].Close/bars[i-60].Close - 1
				} else {
					out[i] = math.NaN()
				}
			}
			return out
		})

	Register("rsi_14", "momentum", "14-period RSI as factor", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.RSI(bars, 14)
			return r.Values
		})
}

func registerTechnical() {
	Register("macd_hist", "technical", "MACD histogram signal", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.MACD(bars, 12, 26, 9)
			hist := make([]float64, len(r.Values))
			for i := 0; i < len(r.Values); i++ {
				if math.IsNaN(r.Values[i]) || math.IsNaN(r.Line2[i]) {
					hist[i] = math.NaN()
				} else {
					hist[i] = (r.Values[i] - r.Line2[i]) * 2
				}
			}
			return hist
		})

	Register("kdj_k", "technical", "KDJ K-line signal", []string{"close", "high", "low"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.KDJ(bars, 9, 3, 3)
			return r.Values
		})

	Register("boll_pct_b", "technical", "Bollinger Band %B position", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.BOLL(bars, 20, 2)
			mid := r.Values
			upper := r.Line2
			lower := r.Line3
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				if i >= len(mid) || i >= len(upper) || i >= len(lower) {
					out[i] = math.NaN()
					continue
				}
				if upper[i]-lower[i] > 0 {
					out[i] = (bars[i].Close - lower[i]) / (upper[i] - lower[i])
				} else {
					out[i] = math.NaN()
				}
			}
			return out
		})

	Register("ema_cross_signal", "technical", "EMA cross signal (5/20)", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			fast := indicator.EXPMA(bars, 5).Values
			slow := indicator.EXPMA(bars, 20).Values
			n := len(bars)
			out := make([]float64, n)
			for i := 1; i < n; i++ {
				if math.IsNaN(fast[i]) || math.IsNaN(slow[i]) {
					out[i] = math.NaN()
				} else if fast[i] > 0 && slow[i] > 0 {
					out[i] = fast[i] - slow[i]
				} else {
					out[i] = math.NaN()
				}
			}
			return out
		})

	Register("dmi_adx", "technical", "DMI ADX trend strength", []string{"close", "high", "low"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.DMI(bars, 14, 6)
			return r.Line3
		})

	Register("cci_14", "technical", "CCI commodity channel index", []string{"close", "high", "low"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.CCI(bars, 14)
			return r.Values
		})

	Register("bias_30", "technical", "30-day bias deviation", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.BIAS(bars, 30)
			return r.Values
		})
}

func registerVolume() {
	Register("obv_trend", "volume", "OBV trend slope", []string{"close", "vol"},
		func(bars []indicator.Bar) []float64 {
			obv := indicator.OBV(bars).Values
			n := len(bars)
			out := make([]float64, n)
			period := 5
			for i := period; i < n; i++ {
				out[i] = linearSlope(obv[i-period:i+1], period+1)
			}
			return out
		})

	Register("volume_ratio_5d", "volume", "5-day volume ratio", []string{"vol"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.VR(bars, 5)
			return r.Values
		})

	Register("turnover_ratio", "volume", "Turnover rate proxy via amount/vol ratio", []string{"vol", "amount"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				if bars[i].Vol > 0 {
					out[i] = bars[i].Amount / bars[i].Vol
				} else {
					out[i] = math.NaN()
				}
			}
			return out
		})
}

func registerVolatility() {
	Register("volatility_20d", "volatility", "20-day return volatility", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			period := 20
			for i := period; i < n; i++ {
				var sum float64
				for j := i - period + 1; j <= i; j++ {
					if bars[j-1].Close > 0 {
						r := bars[j].Close/bars[j-1].Close - 1
						sum += r
					}
				}
				mean := sum / float64(period)
				var ss float64
				for j := i - period + 1; j <= i; j++ {
					if bars[j-1].Close > 0 {
						r := bars[j].Close/bars[j-1].Close - 1
						d := r - mean
						ss += d * d
					}
				}
				out[i] = math.Sqrt(ss / float64(period-1))
			}
			return out
		})

	Register("atr_14", "volatility", "14-day Average True Range", []string{"close", "high", "low"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.ATR(bars, 14)
			return r.Values
		})
}

func registerChanlun() {
	Register("cl_bi_direction", "chanlun", "Chanlun Bi direction encoding (1=up, -1=down)", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				out[i] = math.NaN()
				if i >= 5 {
					trend := bars[i].Close - bars[i-5].Close
					if trend > 0 {
						out[i] = 1
					} else if trend < 0 {
						out[i] = -1
					} else {
						out[i] = 0
					}
				}
			}
			return out
		})

	Register("cl_buy_sell_signal", "chanlun", "Chanlun buy/sell signal encoding", []string{"close", "high", "low"},
		func(bars []indicator.Bar) []float64 {
			r := indicator.MACD(bars, 12, 26, 9)
			dif := r.Values
			dea := r.Line2
			n := len(bars)
			out := make([]float64, n)
			for i := 1; i < n; i++ {
				if i >= len(dif) || i >= len(dea) {
					out[i] = math.NaN()
					continue
				}
				if math.IsNaN(dif[i]) || math.IsNaN(dea[i]) {
					out[i] = math.NaN()
				} else if dif[i] > dea[i] && dif[i-1] <= dea[i-1] {
					out[i] = 1
				} else if dif[i] < dea[i] && dif[i-1] >= dea[i-1] {
					out[i] = -1
				} else {
					out[i] = 0
				}
			}
			return out
		})
}

func registerValue() {
	Register("pe_ratio", "value", "PE ratio placeholder", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				out[i] = math.NaN()
			}
			return out
		})
}

func registerQuality() {
	Register("roe", "quality", "ROE placeholder", []string{"close"},
		func(bars []indicator.Bar) []float64 {
			n := len(bars)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				out[i] = math.NaN()
			}
			return out
		})
}

func linearSlope(series []float64, length int) float64 {
	if length < 2 {
		return 0
	}
	n := float64(length)
	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < length; i++ {
		x := float64(i)
		y := series[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

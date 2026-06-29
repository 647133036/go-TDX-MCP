package backtest

import (
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type BiasReversalStrategy struct {
	Period     int
	Oversold   float64
	Overbought float64
}

func (s *BiasReversalStrategy) Name() string { return "bias_reversal" }

func (s *BiasReversalStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period {
		return Hold
	}
	r := indicator.BIAS(bars, s.Period)
	bias := r.Values
	if i >= len(bias) {
		return Hold
	}
	if bias[i] <= 0 || math.IsNaN(bias[i]) {
		return Hold
	}
	if bias[i-1] <= s.Oversold && bias[i] > s.Oversold {
		return Buy
	}
	if bias[i-1] >= s.Overbought && bias[i] < s.Overbought {
		return Sell
	}
	return Hold
}

type VolumePriceStrategy struct{}

func (s *VolumePriceStrategy) Name() string { return "volume_price" }

func (s *VolumePriceStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < 5 {
		return Hold
	}
	avgVol := avgVol(bars, i, 5)
	if avgVol <= 0 {
		return Hold
	}
	volRatio := bars[i].Vol / avgVol
	isBull := bars[i].Close > bars[i-1].Close
	isBear := bars[i].Close < bars[i-1].Close
	if i == 0 {
		return Hold
	}
	if volRatio > 1.5 && isBull {
		return Buy
	}
	if volRatio < 0.5 && isBear {
		return Sell
	}
	return Hold
}

type DMITrendStrategy struct {
	Period       int
	SignalPeriod int
	ADXThreshold float64
}

func (s *DMITrendStrategy) Name() string { return "dmi_trend" }

func (s *DMITrendStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+s.SignalPeriod {
		return Hold
	}
	r := indicator.DMI(bars, s.Period, s.SignalPeriod)
	pdi := r.Values
	mdi := r.Line2
	adx := r.Line3
	if i >= len(pdi) || i >= len(mdi) || i >= len(adx) {
		return Hold
	}
	if pdi[i] <= 0 || mdi[i] <= 0 || adx[i] <= 0 {
		return Hold
	}
	if adx[i] > s.ADXThreshold && pdi[i] > mdi[i] && pdi[i-1] <= mdi[i-1] {
		return Buy
	}
	if adx[i] > s.ADXThreshold && mdi[i] > pdi[i] && mdi[i-1] <= pdi[i-1] {
		return Sell
	}
	return Hold
}

type CCIBreakoutStrategy struct {
	Period     int
	Overbought float64
	Oversold   float64
}

func (s *CCIBreakoutStrategy) Name() string { return "cci_breakout" }

func (s *CCIBreakoutStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period {
		return Hold
	}
	r := indicator.CCI(bars, s.Period)
	cci := r.Values
	if i >= len(cci) {
		return Hold
	}
	if cci[i] <= 0 || math.IsNaN(cci[i]) {
		return Hold
	}
	if cci[i-1] <= s.Oversold && cci[i] > s.Oversold {
		return Buy
	}
	if cci[i-1] >= s.Overbought && cci[i] < s.Overbought {
		return Sell
	}
	return Hold
}

type MFIVolumeStrategy struct {
	Period    int
	Oversold  float64
	Overbought float64
}

func (s *MFIVolumeStrategy) Name() string { return "mfi_volume" }

func (s *MFIVolumeStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+1 {
		return Hold
	}
	r := indicator.MFI(bars, s.Period)
	mfi := r.Values
	if i >= len(mfi) {
		return Hold
	}
	if mfi[i] <= 0 || math.IsNaN(mfi[i]) {
		return Hold
	}
	if mfi[i-1] <= s.Oversold && mfi[i] > s.Oversold {
		return Buy
	}
	if mfi[i-1] >= s.Overbought && mfi[i] < s.Overbought {
		return Sell
	}
	return Hold
}

type ZhuoYaoMomentumStrategy struct{}

func (s *ZhuoYaoMomentumStrategy) Name() string { return "zhuoyao_momentum" }

func (s *ZhuoYaoMomentumStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < 30 {
		return Hold
	}
	r := indicator.ZHUOYAO(bars)
	short := r.Values
	mid := r.Line2
	trend := r.Line3
	if i >= len(short) || i >= len(mid) || i >= len(trend) {
		return Hold
	}
	if short[i] <= 0 || mid[i] <= 0 || trend[i] <= 0 {
		return Hold
	}
	if short[i] > mid[i] && short[i-1] <= mid[i-1] && trend[i] > 0 {
		return Buy
	}
	if short[i] < mid[i] && short[i-1] >= mid[i-1] {
		return Sell
	}
	return Hold
}

type TRIXCrossStrategy struct {
	Period       int
	SignalPeriod int
}

func (s *TRIXCrossStrategy) Name() string { return "trix_cross" }

func (s *TRIXCrossStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+s.SignalPeriod+1 {
		return Hold
	}
	r := indicator.TRIX(bars, s.Period, s.SignalPeriod)
	trix := r.Values
	matrix := r.Line2
	if i >= len(trix) || i >= len(matrix) {
		return Hold
	}
	if trix[i] <= 0 || matrix[i] <= 0 {
		return Hold
	}
	if trix[i-1] <= matrix[i-1] && trix[i] > matrix[i] {
		return Buy
	}
	if trix[i-1] >= matrix[i-1] && trix[i] < matrix[i] {
		return Sell
	}
	return Hold
}

type MTMMomentumStrategy struct {
	Period int
}

func (s *MTMMomentumStrategy) Name() string { return "mtm_momentum" }

func (s *MTMMomentumStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+1 {
		return Hold
	}
	r := indicator.MTM(bars, s.Period)
	mtm := r.Values
	if i >= len(mtm) {
		return Hold
	}
	if math.IsNaN(mtm[i]) || math.IsNaN(mtm[i-1]) {
		return Hold
	}
	if mtm[i-1] <= 0 && mtm[i] > 0 {
		return Buy
	}
	if mtm[i-1] >= 0 && mtm[i] < 0 {
		return Sell
	}
	return Hold
}

type OBVTrendStrategy struct {
	Period int
}

func (s *OBVTrendStrategy) Name() string { return "obv_trend" }

func (s *OBVTrendStrategy) Next(i int, bars []indicator.Bar) Signal {
	if i < s.Period+1 {
		return Hold
	}
	r := indicator.OBV(bars)
	obv := r.Values
	maobv := indicator.MA(obv, s.Period)
	if i >= len(obv) || i >= len(maobv) {
		return Hold
	}
	if maobv[i] <= 0 || maobv[i-1] <= 0 || maobv[i-5] <= 0 {
		return Hold
	}
	if obv[i] > maobv[i]*1.02 && obv[i-1] <= maobv[i-1]*1.02 && maobv[i] > maobv[i-5] {
		return Buy
	}
	if obv[i] < maobv[i]*0.98 && obv[i-1] >= maobv[i-1]*0.98 {
		return Sell
	}
	return Hold
}

func avgVol(bars []indicator.Bar, end, period int) float64 {
	start := end - period + 1
	if start < 0 {
		start = 0
	}
	if start >= len(bars) {
		return 0
	}
	sum := 0.0
	count := 0
	for j := start; j <= end && j < len(bars); j++ {
		sum += bars[j].Vol
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

package indicator

import "math"

type Bar struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Vol    float64
	Amount float64
}

type IndicatorResult struct {
	Values []float64
	Line2  []float64
	Line3  []float64
	Data   map[string][]float64
}

func nan() float64 {
	return math.NaN()
}

func MA(data []float64, period int) []float64 {
	return sma(data, period)
}

func resultLen(bars []Bar) int {
	return len(bars)
}

func sma(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += data[i]
		if i >= period {
			sum -= data[i-period]
		}
		if i >= period-1 {
			out[i] = sum / float64(period)
		}
	}
	return out
}

func ema(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	alpha := 2.0 / float64(period+1)
	var seedSum float64
	for i := 0; i < period && i < n; i++ {
		seedSum += data[i]
	}
	if n >= period {
		out[period-1] = seedSum / float64(period)
	}
	for i := period; i < n; i++ {
		out[i] = alpha*data[i] + (1-alpha)*out[i-1]
	}
	return out
}

func wilderSmooth(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	var seedSum float64
	for i := 0; i < period && i < n; i++ {
		seedSum += data[i]
	}
	if n >= period {
		out[period-1] = seedSum / float64(period)
	}
	m := float64(period)
	for i := period; i < n; i++ {
		out[i] = (out[i-1]*(m-1) + data[i]) / m
	}
	return out
}

func hhv(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	for i := 0; i < n; i++ {
		start := i - period + 1
		if start < 0 {
			start = 0
		}
		mx := data[start]
		for j := start + 1; j <= i; j++ {
			if data[j] > mx {
				mx = data[j]
			}
		}
		if i >= period-1 {
			out[i] = mx
		}
	}
	return out
}

func llv(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	for i := 0; i < n; i++ {
		start := i - period + 1
		if start < 0 {
			start = 0
		}
		mn := data[start]
		for j := start + 1; j <= i; j++ {
			if data[j] < mn {
				mn = data[j]
			}
		}
		if i >= period-1 {
			out[i] = mn
		}
	}
	return out
}

func hhvIndex(data []float64, period int) []int {
	n := len(data)
	out := make([]int, n)
	if period <= 0 || n == 0 {
		return out
	}
	for i := period - 1; i < n; i++ {
		mx := data[i-period+1]
		mxIdx := i - period + 1
		for j := i - period + 2; j <= i; j++ {
			if data[j] > mx {
				mx = data[j]
				mxIdx = j
			}
		}
		out[i] = mxIdx
	}
	return out
}

func llvIndex(data []float64, period int) []int {
	n := len(data)
	out := make([]int, n)
	if period <= 0 || n == 0 {
		return out
	}
	for i := period - 1; i < n; i++ {
		mn := data[i-period+1]
		mnIdx := i - period + 1
		for j := i - period + 2; j <= i; j++ {
			if data[j] < mn {
				mn = data[j]
				mnIdx = j
			}
		}
		out[i] = mnIdx
	}
	return out
}

func std(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 1 || n == 0 {
		return out
	}
	ma := sma(data, period)
	for i := period - 1; i < n; i++ {
		var sumSq float64
		for j := i - period + 1; j <= i; j++ {
			diff := data[j] - ma[i]
			sumSq += diff * diff
		}
		out[i] = math.Sqrt(sumSq / float64(period))
	}
	return out
}

func rollingSum(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	if period <= 0 || n == 0 {
		return out
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += data[i]
		if i >= period {
			sum -= data[i-period]
		}
		if i >= period-1 {
			out[i] = sum
		}
	}
	return out
}

func ref(data []float64, offset int) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		idx := i - offset
		if idx >= 0 && idx < n {
			out[i] = data[idx]
		}
	}
	return out
}

func absSlice(data []float64) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = math.Abs(data[i])
	}
	return out
}

func subSlice(a, b []float64) []float64 {
	n := len(a)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] - b[i]
	}
	return out
}

func addSlice(a, b []float64) []float64 {
	n := len(a)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] + b[i]
	}
	return out
}

func mulSlice(a, b []float64) []float64 {
	n := len(a)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] * b[i]
	}
	return out
}

func divSlice(a, b []float64) []float64 {
	n := len(a)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		if b[i] != 0 {
			out[i] = a[i] / b[i]
		} else {
			out[i] = 0
		}
	}
	return out
}

func scalarMul(data []float64, s float64) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = data[i] * s
	}
	return out
}

func scalarAdd(data []float64, s float64) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = data[i] + s
	}
	return out
}

func extractClose(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].Close
	}
	return out
}

func extractHigh(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].High
	}
	return out
}

func extractLow(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].Low
	}
	return out
}

func extractOpen(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].Open
	}
	return out
}

func extractVol(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].Vol
	}
	return out
}

func extractAmount(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = bars[i].Amount
	}
	return out
}

func typicalPrice(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = (bars[i].High + bars[i].Low + bars[i].Close) / 3.0
	}
	return out
}

func trueRange(bars []Bar) []float64 {
	n := len(bars)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		highLow := bars[i].High - bars[i].Low
		if i == 0 {
			out[i] = highLow
			continue
		}
		highClose := math.Abs(bars[i].High - bars[i-1].Close)
		lowClose := math.Abs(bars[i].Low - bars[i-1].Close)
		out[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}
	return out
}

func max2(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max3(a, b, c float64) float64 {
	return max2(a, max2(b, c))
}

func min3(a, b, c float64) float64 {
	return min2(a, min2(b, c))
}

// === 1. MACD ===
func MACD(bars []Bar, fastEMA, slowEMA, signalEMA int) IndicatorResult {
	n := resultLen(bars)
	closes := extractClose(bars)
	ema12 := ema(closes, fastEMA)
	ema26 := ema(closes, slowEMA)
	dif := subSlice(ema12, ema26)
	dea := ema(dif, signalEMA)
	macdBar := make([]float64, n)
	for i := 0; i < n; i++ {
		macdBar[i] = 2.0 * (dif[i] - dea[i])
	}
	return IndicatorResult{Values: dif, Line2: dea, Line3: macdBar}
}

// === 2. KDJ ===
func KDJ(bars []Bar, n, m1, m2 int) IndicatorResult {
	size := resultLen(bars)
	highs := extractHigh(bars)
	lows := extractLow(bars)
	closes := extractClose(bars)
	k := make([]float64, size)
	d := make([]float64, size)
	j := make([]float64, size)
	a1 := 1.0 / float64(m1)
	a2 := 1.0 / float64(m2)
	rsv := make([]float64, size)
	hh := hhv(highs, n)
	ll := llv(lows, n)
	for i := 0; i < size; i++ {
		denom := hh[i] - ll[i]
		if denom != 0 {
			rsv[i] = (closes[i] - ll[i]) / denom * 100.0
		}
	}
	for i := 0; i < size; i++ {
		if i == 0 {
			k[i] = 50
			d[i] = 50
		} else {
			k[i] = (1-a1)*k[i-1] + a1*rsv[i]
			d[i] = (1-a2)*d[i-1] + a2*k[i]
		}
		j[i] = 3.0*k[i] - 2.0*d[i]
	}
	return IndicatorResult{Values: k, Line2: d, Line3: j}
}

// === 3. RSI ===
func RSI(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	closes := extractClose(bars)
	rsiVals := make([]float64, size)
	if size < n+1 {
		return IndicatorResult{Values: rsiVals}
	}
	gains := make([]float64, size)
	losses := make([]float64, size)
	for i := 1; i < size; i++ {
		diff := closes[i] - closes[i-1]
		if diff > 0 {
			gains[i] = diff
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -diff
		}
	}
	avgGain := wilderSmooth(gains, n)
	avgLoss := wilderSmooth(losses, n)
	for i := 0; i < size; i++ {
		if avgLoss[i] == 0 {
			rsiVals[i] = 100
		} else {
			rs := avgGain[i] / avgLoss[i]
			rsiVals[i] = 100.0 - 100.0/(1.0+rs)
		}
	}
	return IndicatorResult{Values: rsiVals}
}

// === 4. BOLL ===
func BOLL(bars []Bar, n int, p float64) IndicatorResult {
	closes := extractClose(bars)
	mid := sma(closes, n)
	stdev := std(closes, n)
	size := resultLen(bars)
	upper := make([]float64, size)
	lower := make([]float64, size)
	for i := 0; i < size; i++ {
		upper[i] = mid[i] + p*stdev[i]
		lower[i] = mid[i] - p*stdev[i]
	}
	return IndicatorResult{Values: mid, Line2: upper, Line3: lower}
}

// === 5. DMI ===
func DMI(bars []Bar, n, m int) IndicatorResult {
	size := resultLen(bars)
	highs := extractHigh(bars)
	lows := extractLow(bars)
	tr := trueRange(bars)
	hd := make([]float64, size)
	ld := make([]float64, size)
	dmp := make([]float64, size)
	dmm := make([]float64, size)
	for i := 1; i < size; i++ {
		hd[i] = highs[i] - highs[i-1]
		ld[i] = lows[i-1] - lows[i]
		if hd[i] > 0 && hd[i] > ld[i] {
			dmp[i] = hd[i]
		} else {
			dmp[i] = 0
		}
		if ld[i] > 0 && ld[i] > hd[i] {
			dmm[i] = ld[i]
		} else {
			dmm[i] = 0
		}
	}
	trN := wilderSmooth(tr, n)
	dmpN := wilderSmooth(dmp, n)
	dmmN := wilderSmooth(dmm, n)
	pdi := make([]float64, size)
	mdi := make([]float64, size)
	dx := make([]float64, size)
	for i := 0; i < size; i++ {
		if trN[i] != 0 {
			pdi[i] = dmpN[i] / trN[i] * 100.0
			mdi[i] = dmmN[i] / trN[i] * 100.0
			sumPDMDI := pdi[i] + mdi[i]
			if sumPDMDI != 0 {
				dx[i] = math.Abs(pdi[i]-mdi[i]) / sumPDMDI * 100.0
			}
		}
	}
	adx := sma(dx, m)
	adxr := make([]float64, size)
	refADX := ref(adx, m)
	for i := 0; i < size; i++ {
		adxr[i] = (adx[i] + refADX[i]) / 2.0
	}
	return IndicatorResult{
		Values: pdi,
		Line2:  mdi,
		Line3:  adx,
		Data:   map[string][]float64{"ADXR": adxr},
	}
}

// === 6. ATR ===
func ATR(bars []Bar, n int) IndicatorResult {
	return IndicatorResult{Values: wilderSmooth(trueRange(bars), n)}
}

// === 7. WR ===
func WR(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	highs := extractHigh(bars)
	lows := extractLow(bars)
	closes := extractClose(bars)
	hh := hhv(highs, n)
	ll := llv(lows, n)
	wrVals := make([]float64, size)
	for i := 0; i < size; i++ {
		denom := hh[i] - ll[i]
		if denom != 0 {
			wrVals[i] = (hh[i] - closes[i]) / denom * 100.0
		}
	}
	return IndicatorResult{Values: wrVals}
}

// === 8. CCI ===
func CCI(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	tp := typicalPrice(bars)
	tpMA := sma(tp, n)
	cciVals := make([]float64, size)
	for i := n - 1; i < size; i++ {
		var madSum float64
		for j := i - n + 1; j <= i; j++ {
			madSum += math.Abs(tp[j] - tpMA[i])
		}
		mad := madSum / float64(n) * 0.015
		if mad != 0 {
			cciVals[i] = (tp[i] - tpMA[i]) / mad
		}
	}
	return IndicatorResult{Values: cciVals}
}

// === 9. BIAS ===
func BIAS(bars []Bar, n int) IndicatorResult {
	closes := extractClose(bars)
	maN := sma(closes, n)
	size := resultLen(bars)
	biasVals := make([]float64, size)
	for i := 0; i < size; i++ {
		if maN[i] != 0 {
			biasVals[i] = (closes[i] - maN[i]) / maN[i] * 100.0
		}
	}
	return IndicatorResult{Values: biasVals}
}

// === 10. BIAS_SIGNAL ===
func BIAS_SIGNAL(bars []Bar, p, m int) IndicatorResult {
	size := resultLen(bars)
	closes := extractClose(bars)
	ma30 := sma(closes, 30)
	bias30 := make([]float64, size)
	for i := 0; i < size; i++ {
		if ma30[i] != 0 {
			bias30[i] = (closes[i] - ma30[i]) / ma30[i] * 100.0
		}
	}
	sSMA := sma(bias30, p)
	lLMA := sma(bias30, m)
	return IndicatorResult{Values: bias30, Line2: sSMA, Line3: lLMA}
}

// === 11. OBV ===
func OBV(bars []Bar) IndicatorResult {
	size := resultLen(bars)
	obvVals := make([]float64, size)
	for i := 1; i < size; i++ {
		obvVals[i] = obvVals[i-1]
		if bars[i].Close > bars[i-1].Close {
			obvVals[i] += bars[i].Vol
		} else if bars[i].Close < bars[i-1].Close {
			obvVals[i] -= bars[i].Vol
		}
	}
	return IndicatorResult{Values: obvVals}
}

// === 12. VR ===
func VR(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	vrVals := make([]float64, size)
	if size < n {
		return IndicatorResult{Values: vrVals}
	}
	avs := make([]float64, size)
	bvs := make([]float64, size)
	cvs := make([]float64, size)
	for i := 1; i < size; i++ {
		if bars[i].Close > bars[i-1].Close {
			avs[i] = bars[i].Vol
		} else if bars[i].Close < bars[i-1].Close {
			bvs[i] = bars[i].Vol
		} else {
			cvs[i] = bars[i].Vol
		}
	}
	avsSum := rollingSum(avs, n)
	bvsSum := rollingSum(bvs, n)
	cvsSum := rollingSum(cvs, n)
	for i := 0; i < size; i++ {
		denom := bvsSum[i] + cvsSum[i]/2.0
		if denom != 0 {
			vrVals[i] = (avsSum[i] + cvsSum[i]/2.0) / denom * 100.0
		}
	}
	return IndicatorResult{Values: vrVals}
}

// === 13. EMV ===
func EMV(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	emvRaw := make([]float64, size)
	for i := 1; i < size; i++ {
		midPrev := (bars[i-1].High + bars[i-1].Low) / 2.0
		midCurr := (bars[i].High + bars[i].Low) / 2.0
		hlRange := bars[i].High - bars[i].Low
		if hlRange == 0 {
			continue
		}
		boxRatio := (bars[i].Vol / 100.0) / hlRange
		if boxRatio != 0 {
			emvRaw[i] = (midCurr - midPrev) / boxRatio
		}
	}
	return IndicatorResult{Values: sma(emvRaw, n)}
}

// === 14. MFI ===
func MFI(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	posMF := make([]float64, size)
	negMF := make([]float64, size)
	for i := 1; i < size; i++ {
		tpCurr := (bars[i].High + bars[i].Low + bars[i].Close) / 3.0
		tpPrev := (bars[i-1].High + bars[i-1].Low + bars[i-1].Close) / 3.0
		mf := tpCurr * bars[i].Vol
		if tpCurr > tpPrev {
			posMF[i] = mf
		} else if tpCurr < tpPrev {
			negMF[i] = mf
		}
	}
	posSum := rollingSum(posMF, n)
	negSum := rollingSum(negMF, n)
	mfiVals := make([]float64, size)
	for i := 0; i < size; i++ {
		if negSum[i] != 0 {
			mr := posSum[i] / negSum[i]
			mfiVals[i] = 100.0 - 100.0/(1.0+mr)
		} else if posSum[i] != 0 {
			mfiVals[i] = 100.0
		}
	}
	return IndicatorResult{Values: mfiVals}
}

// === 15. BRAR ===
func BRAR(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	ars := make([]float64, size)
	brs := make([]float64, size)
	arH := make([]float64, size)
	arL := make([]float64, size)
	brH := make([]float64, size)
	brL := make([]float64, size)
	for i := 0; i < size; i++ {
		arH[i] = bars[i].High - bars[i].Open
		arL[i] = bars[i].Open - bars[i].Low
		if i > 0 {
			brH[i] = max2(0, bars[i].High-bars[i-1].Close)
			brL[i] = max2(0, bars[i-1].Close-bars[i].Low)
		}
	}
	arHSum := rollingSum(arH, n)
	arLSum := rollingSum(arL, n)
	brHSum := rollingSum(brH, n)
	brLSum := rollingSum(brL, n)
	for i := 0; i < size; i++ {
		if arLSum[i] != 0 {
			ars[i] = arHSum[i] / arLSum[i] * 100.0
		}
		if brLSum[i] != 0 {
			brs[i] = brHSum[i] / brLSum[i] * 100.0
		}
	}
	return IndicatorResult{Values: brs, Line2: ars}
}

// === 16. ASI ===
func ASI(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	si := make([]float64, size)
	for i := 1; i < size; i++ {
		h := bars[i].High
		l := bars[i].Low
		c := bars[i].Close
		o := bars[i].Open
		pc := bars[i-1].Close
		po := bars[i-1].Open
		ph := bars[i-1].High
		a := math.Abs(h - pc)
		b := math.Abs(l - pc)
		e := math.Abs(h - l)
		f := math.Abs(ph - pc)
		var r float64
		if a >= b && a >= e {
			r = a + b/2.0 + f/4.0
		} else if b >= a && b >= e {
			r = b + a/2.0 + f/4.0
		} else {
			r = e + f/4.0
		}
		k := max2(a, b)
		t := 3.0
		if r != 0 {
			si[i] = 50.0 * (c-pc + (c-o)/2.0 + (pc-po)/4.0) / r * k / t
		}
	}
	asiVals := make([]float64, size)
	var cum float64
	for i := 0; i < size; i++ {
		cum += si[i]
		asiVals[i] = cum
	}
	return IndicatorResult{Values: sma(asiVals, n)}
}

// === 17. TRIX ===
func TRIX(bars []Bar, n, m int) IndicatorResult {
	closes := extractClose(bars)
	trixVals := ema(ema(ema(closes, n), n), n)
	trixRate := make([]float64, len(trixVals))
	for i := 1; i < len(trixVals); i++ {
		if trixVals[i-1] != 0 {
			trixRate[i] = (trixVals[i] - trixVals[i-1]) / trixVals[i-1] * 100.0
		}
	}
	matrix := sma(trixRate, m)
	return IndicatorResult{Values: trixRate, Line2: matrix}
}

// === 18. DPO ===
func DPO(bars []Bar, n int) IndicatorResult {
	closes := extractClose(bars)
	maN := sma(closes, n)
	size := resultLen(bars)
	dpoVals := make([]float64, size)
	shift := n / 2
	for i := 0; i < size; i++ {
		idx := i - shift
		if idx >= 0 && idx < size && i < size {
			dpoVals[i] = closes[idx] - maN[idx]
		}
	}
	return IndicatorResult{Values: dpoVals}
}

// === 19. MTM ===
func MTM(bars []Bar, n int) IndicatorResult {
	closes := extractClose(bars)
	refClose := ref(closes, n)
	return IndicatorResult{Values: subSlice(closes, refClose)}
}

// === 20. ROC ===
func ROC(bars []Bar, n int) IndicatorResult {
	closes := extractClose(bars)
	refClose := ref(closes, n)
	size := resultLen(bars)
	rocVals := make([]float64, size)
	for i := 0; i < size; i++ {
		if refClose[i] != 0 {
			rocVals[i] = (closes[i] - refClose[i]) / refClose[i] * 100.0
		}
	}
	return IndicatorResult{Values: rocVals}
}

// === 21. EXPMA ===
func EXPMA(bars []Bar, n int) IndicatorResult {
	return IndicatorResult{Values: ema(extractClose(bars), n)}
}

// === 22. BBI ===
func BBI(bars []Bar, n1, n2, n3, n4 int) IndicatorResult {
	closes := extractClose(bars)
	size := resultLen(bars)
	ma1 := sma(closes, n1)
	ma2 := sma(closes, n2)
	ma3 := sma(closes, n3)
	ma4 := sma(closes, n4)
	bbiVals := make([]float64, size)
	for i := 0; i < size; i++ {
		bbiVals[i] = (ma1[i] + ma2[i] + ma3[i] + ma4[i]) / 4.0
	}
	return IndicatorResult{Values: bbiVals}
}

// === 23. PSY ===
func PSY(bars []Bar, n int) IndicatorResult {
	closes := extractClose(bars)
	size := resultLen(bars)
	psyVals := make([]float64, size)
	for i := n - 1; i < size; i++ {
		count := 0
		for j := i - n + 1; j <= i; j++ {
			if j > 0 && closes[j] > closes[j-1] {
				count++
			}
		}
		psyVals[i] = float64(count) / float64(n) * 100.0
	}
	return IndicatorResult{Values: psyVals}
}

// === 24. DFMA ===
func DFMA(bars []Bar) IndicatorResult {
	size := resultLen(bars)
	hlRange := make([]float64, size)
	for i := 0; i < size; i++ {
		hlRange[i] = bars[i].High - bars[i].Low
	}
	dfmaMid := sma(hlRange, 20)
	dfmaUp := sma(hlRange, 10)
	dfmaLow := sma(hlRange, 30)
	return IndicatorResult{Values: dfmaMid, Line2: dfmaUp, Line3: dfmaLow}
}

// === 25. CR ===
func CR(bars []Bar, n, m1, m2 int) IndicatorResult {
	size := resultLen(bars)
	crVals := make([]float64, size)
	for i := 1; i < size; i++ {
		mid := (bars[i-1].High + bars[i-1].Low + bars[i-1].Close) / 3.0
		posSum := max2(0, bars[i].High-mid)
		negSum := max2(0, mid-bars[i].Low)
		if negSum != 0 {
			crVals[i] = posSum / negSum * 100.0
		}
	}
	crMA := sma(crVals, n)
	ma1 := sma(crMA, m1)
	ma2 := sma(crMA, m2)
	return IndicatorResult{Values: crMA, Line2: ma1, Line3: ma2}
}

// === 26. KTN (Keltner Channel) ===
func KTN(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	closes := extractClose(bars)
	mid := ema(closes, n)
	atr := ATR(bars, n).Values
	upper := make([]float64, size)
	lower := make([]float64, size)
	for i := 0; i < size; i++ {
		upper[i] = mid[i] + 2.0*atr[i]
		lower[i] = mid[i] - 2.0*atr[i]
	}
	return IndicatorResult{Values: mid, Line2: upper, Line3: lower}
}

// === 27. XSII (薛斯通道II) ===
func XSII(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	closes := extractClose(bars)
	shortPeriod := n / 2
	if shortPeriod < 2 {
		shortPeriod = 2
	}
	shortMA := sma(closes, shortPeriod)
	shortStd := std(closes, shortPeriod)
	longMA := sma(closes, n)
	longStd := std(closes, n)
	shortTop := make([]float64, size)
	shortBot := make([]float64, size)
	longTop := make([]float64, size)
	longBot := make([]float64, size)
	for i := 0; i < size; i++ {
		shortTop[i] = shortMA[i] + 2.0*shortStd[i]
		shortBot[i] = shortMA[i] - 2.0*shortStd[i]
		longTop[i] = longMA[i] + 2.0*longStd[i]
		longBot[i] = longMA[i] - 2.0*longStd[i]
	}
	mid := sma(closes, n)
	return IndicatorResult{
		Values: shortTop,
		Line2:  shortBot,
		Line3:  mid,
		Data:   map[string][]float64{"LONG_TOP": longTop, "LONG_BOT": longBot},
	}
}

// === 28. MASS ===
func MASS(bars []Bar, n, m int) IndicatorResult {
	size := resultLen(bars)
	hml := make([]float64, size)
	for i := 0; i < size; i++ {
		hml[i] = bars[i].High - bars[i].Low
	}
	ema1 := ema(hml, m)
	ema2 := ema(ema1, m)
	ratio := make([]float64, size)
	for i := 0; i < size; i++ {
		if ema2[i] != 0 {
			ratio[i] = ema1[i] / ema2[i]
		}
	}
	return IndicatorResult{Values: rollingSum(ratio, n)}
}

// === 29. TAQ (唐安奇通道) ===
func TAQ(bars []Bar, n int) IndicatorResult {
	highs := extractHigh(bars)
	lows := extractLow(bars)
	upper := hhv(highs, n)
	lower := llv(lows, n)
	size := resultLen(bars)
	mid := make([]float64, size)
	for i := 0; i < size; i++ {
		mid[i] = (upper[i] + lower[i]) / 2.0
	}
	return IndicatorResult{Values: upper, Line2: lower, Line3: mid}
}

// === 30. ZHUOYAO ===
func ZHUOYAO(bars []Bar) IndicatorResult {
	size := resultLen(bars)
	closes := extractClose(bars)
	chg120 := make([]float64, size)
	chg60 := make([]float64, size)
	chg20 := make([]float64, size)
	for i := 0; i < size; i++ {
		if i >= 120 {
			if closes[i-120] != 0 {
				chg120[i] = (closes[i] - closes[i-120]) / closes[i-120] * 100.0
			}
		} else {
			chg120[i] = 0
		}
		if i >= 60 {
			if closes[i-60] != 0 {
				chg60[i] = (closes[i] - closes[i-60]) / closes[i-60] * 100.0
			}
		}
		if i >= 20 {
			if closes[i-20] != 0 {
				chg20[i] = (closes[i] - closes[i-20]) / closes[i-20] * 100.0
			}
		}
	}
	zyLong := ema(chg120, 10)
	zyTrend := ema(chg60, 10)
	return IndicatorResult{
		Values: zyLong,
		Line2:  chg60,
		Line3:  chg20,
		Data:   map[string][]float64{"TREND": zyTrend},
	}
}

// === 31. SAR ===
func SAR(bars []Bar, afStep, afMax float64) IndicatorResult {
	size := resultLen(bars)
	sarVals := make([]float64, size)
	if size < 2 {
		return IndicatorResult{Values: sarVals}
	}

	highs := extractHigh(bars)
	lows := extractLow(bars)

	bullish := true
	af := afStep
	ep := highs[0]
	sarVals[0] = lows[0]
	if bars[0].Close > bars[0].Open {
		sarVals[0] = lows[0]
	} else {
		sarVals[0] = highs[0]
	}

	for i := 1; i < size; i++ {
		prevSAR := sarVals[i-1]
		if bullish {
			sarVals[i] = prevSAR + af*(ep-prevSAR)
			if sarVals[i] > lows[i] {
				sarVals[i] = lows[i]
			}
			if i >= 1 {
				if sarVals[i] > lows[i-1] {
					sarVals[i] = lows[i-1]
				}
			}
			if highs[i] > ep {
				ep = highs[i]
				af += afStep
				if af > afMax {
					af = afMax
				}
			}
			if bars[i].Low < sarVals[i] {
				bullish = false
				sarVals[i] = ep
				af = afStep
				ep = lows[i]
			}
		} else {
			sarVals[i] = prevSAR + af*(ep-prevSAR)
			if sarVals[i] < highs[i] {
				sarVals[i] = highs[i]
			}
			if i >= 1 {
				if sarVals[i] < highs[i-1] {
					sarVals[i] = highs[i-1]
				}
			}
			if lows[i] < ep {
				ep = lows[i]
				af += afStep
				if af > afMax {
					af = afMax
				}
			}
			if bars[i].High > sarVals[i] {
				bullish = true
				sarVals[i] = ep
				af = afStep
				ep = highs[i]
			}
		}
	}
	return IndicatorResult{Values: sarVals}
}

// === 32. VWAP ===
func VWAP(bars []Bar) IndicatorResult {
	size := resultLen(bars)
	vwapVals := make([]float64, size)
	var cumPV, cumV float64
	for i := 0; i < size; i++ {
		tp := (bars[i].High + bars[i].Low + bars[i].Close) / 3.0
		cumPV += tp * bars[i].Vol
		cumV += bars[i].Vol
		if cumV != 0 {
			vwapVals[i] = cumPV / cumV
		}
	}
	return IndicatorResult{Values: vwapVals}
}

// === 33. AROON ===
func AROON(bars []Bar, n int) IndicatorResult {
	size := resultLen(bars)
	highs := extractHigh(bars)
	lows := extractLow(bars)
	hhIdx := hhvIndex(highs, n)
	llIdx := llvIndex(lows, n)
	aroonUp := make([]float64, size)
	aroonDown := make([]float64, size)
	for i := n - 1; i < size; i++ {
		periodsSinceHigh := i - hhIdx[i]
		aroonUp[i] = (float64(n) - float64(periodsSinceHigh)) / float64(n) * 100.0
		periodsSinceLow := i - llIdx[i]
		aroonDown[i] = (float64(n) - float64(periodsSinceLow)) / float64(n) * 100.0
	}
	return IndicatorResult{Values: aroonUp, Line2: aroonDown}
}

// === 34. FK 反K指标 ===
func FK(bars []Bar) IndicatorResult {
	size := resultLen(bars)
	n := 20
	fkVals := make([]float64, size)
	for i := n - 1; i < size; i++ {
		count := 0
		for j := i - n + 1; j <= i; j++ {
			if j > 0 {
				prevDirection := bars[j-1].Close - bars[j-1].Open
				currDirection := bars[j].Close - bars[j].Open
				if prevDirection*currDirection < 0 {
					count++
				}
			}
		}
		fkVals[i] = float64(count) / float64(n) * 100.0
	}
	return IndicatorResult{Values: fkVals}
}

// === ComputeAll ===
func ComputeAll(bars []Bar, indicators []string, params map[string]float64) (map[string]IndicatorResult, error) {
	result := make(map[string]IndicatorResult)
	for _, name := range indicators {
		switch name {
		case "MACD":
			fastEMA := intParam(params, "FastEMA", 12)
			slowEMA := intParam(params, "SlowEMA", 26)
			signalEMA := intParam(params, "SignalEMA", 9)
			result[name] = MACD(bars, fastEMA, slowEMA, signalEMA)
		case "MA":
			n := intParam(params, "N", 5)
			result[name] = IndicatorResult{Values: MA(extractClose(bars), n)}
		case "KDJ":
			n := intParam(params, "N", 9)
			m1 := intParam(params, "M1", 3)
			m2 := intParam(params, "M2", 3)
			result[name] = KDJ(bars, n, m1, m2)
		case "RSI":
			n := intParam(params, "N", 6)
			result[name] = RSI(bars, n)
		case "BOLL":
			n := intParam(params, "N", 20)
			p := params["P"]
			if p == 0 {
				p = 2
			}
			result[name] = BOLL(bars, n, p)
		case "DMI":
			n := intParam(params, "N", 14)
			m := intParam(params, "M", 6)
			result[name] = DMI(bars, n, m)
		case "ATR":
			n := intParam(params, "N", 14)
			result[name] = ATR(bars, n)
		case "WR":
			n := intParam(params, "N", 10)
			result[name] = WR(bars, n)
		case "CCI":
			n := intParam(params, "N", 14)
			result[name] = CCI(bars, n)
		case "BIAS":
			n := intParam(params, "N", 6)
			result[name] = BIAS(bars, n)
		case "BIAS_SIGNAL":
			p := intParam(params, "P", 5)
			m := intParam(params, "M", 20)
			result[name] = BIAS_SIGNAL(bars, p, m)
		case "OBV":
			result[name] = OBV(bars)
		case "VR":
			n := intParam(params, "N", 26)
			result[name] = VR(bars, n)
		case "EMV":
			n := intParam(params, "N", 14)
			result[name] = EMV(bars, n)
		case "MFI":
			n := intParam(params, "N", 14)
			result[name] = MFI(bars, n)
		case "BRAR":
			n := intParam(params, "N", 26)
			result[name] = BRAR(bars, n)
		case "ASI":
			n := intParam(params, "N", 6)
			result[name] = ASI(bars, n)
		case "TRIX":
			n := intParam(params, "N", 12)
			m := intParam(params, "M", 9)
			result[name] = TRIX(bars, n, m)
		case "DPO":
			n := intParam(params, "N", 20)
			result[name] = DPO(bars, n)
		case "MTM":
			n := intParam(params, "N", 6)
			result[name] = MTM(bars, n)
		case "ROC":
			n := intParam(params, "N", 12)
			result[name] = ROC(bars, n)
		case "EXPMA":
			n := intParam(params, "N", 12)
			result[name] = EXPMA(bars, n)
		case "BBI":
			n1 := intParam(params, "N1", 3)
			n2 := intParam(params, "N2", 6)
			n3 := intParam(params, "N3", 12)
			n4 := intParam(params, "N4", 24)
			result[name] = BBI(bars, n1, n2, n3, n4)
		case "PSY":
			n := intParam(params, "N", 12)
			result[name] = PSY(bars, n)
		case "DFMA":
			result[name] = DFMA(bars)
		case "CR":
			n := intParam(params, "N", 26)
			m1 := intParam(params, "M1", 10)
			m2 := intParam(params, "M2", 20)
			result[name] = CR(bars, n, m1, m2)
		case "KTN":
			n := intParam(params, "N", 20)
			result[name] = KTN(bars, n)
		case "XSII":
			n := intParam(params, "N", 20)
			result[name] = XSII(bars, n)
		case "MASS":
			n := intParam(params, "N", 25)
			m := intParam(params, "M", 9)
			result[name] = MASS(bars, n, m)
		case "TAQ":
			n := intParam(params, "N", 20)
			result[name] = TAQ(bars, n)
		case "ZHUOYAO":
			result[name] = ZHUOYAO(bars)
		case "SAR":
			afStep := params["AFStep"]
			if afStep == 0 {
				afStep = 0.02
			}
			afMax := params["AFMax"]
			if afMax == 0 {
				afMax = 0.2
			}
			result[name] = SAR(bars, afStep, afMax)
		case "VWAP":
			result[name] = VWAP(bars)
		case "AROON":
			n := intParam(params, "N", 25)
			result[name] = AROON(bars, n)
		case "FK":
			result[name] = FK(bars)
		}
	}
	return result, nil
}

func intParam(params map[string]float64, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		return int(v)
	}
	return defaultVal
}

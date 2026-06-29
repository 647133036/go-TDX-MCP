package chanlun

import (
	"fmt"
	"math"
)

// ---- 数据结构定义 ----

type Kline struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Vol    float64
	Amount float64
}

type FenXing struct {
	Index int
	Type  string // "top"(顶分型) or "bottom"(底分型)
	Date  string
	High  float64
	Low   float64
}

type Bi struct {
	Index     int
	Direction string // "up" or "down"
	StartDate string
	EndDate   string
	StartIdx  int
	EndIdx    int
	High      float64
	Low       float64
	Confirmed bool
}

type ZhongShu struct {
	Index     int
	StartDate string
	EndDate   string
	ZG        float64 // 中枢上沿
	ZD        float64 // 中枢下沿
	GG        float64 // 最高点
	DD        float64 // 最低点
	LineCount int    // 构成笔数
	Direction string // "up" or "down"
	Confirmed bool
}

type XianDuan struct {
	Index     int
	Direction string
	StartDate string
	EndDate   string
	High      float64
	Low       float64
}

type MaiMaiDian struct {
	Index  int
	Type   string // "1buy"/"2buy"/"3buy"/"1sell"/"2sell"/"3sell"
	Date   string
	Price  float64
	Reason string
}

type BeiChi struct {
	Index int
	Type  string // "bi"(笔背驰) / "pz"(盘整背驰) / "qs"(趋势背驰)
	Desc  string
}

type ChanLunResult struct {
	Symbol         string
	Period         string
	OrigCount      int
	MergedCount    int
	FenXingCount   int
	BiList         []Bi
	ZhongShuList   []ZhongShu
	XianDuanList   []XianDuan
	MaiMaiDianList []MaiMaiDian
	BeiChiList     []BeiChi
}

// ---- 通用辅助函数 ----

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func containsKline(a, b Kline) bool {
	return (a.High >= b.High && a.Low <= b.Low) || (b.High >= a.High && b.Low <= a.Low)
}

func getMergeDirection(klines []Kline) int {
	n := len(klines)
	if n < 2 {
		return 1
	}
	for j := n - 1; j >= 1; j-- {
		if klines[j].High > klines[j-1].High {
			return 1
		}
		if klines[j].High < klines[j-1].High {
			return -1
		}
	}
	return 1
}

// ---- EMA 计算 ----

func calcEMA(data []float64, period int) []float64 {
	if len(data) == 0 || period <= 0 {
		return nil
	}
	result := make([]float64, len(data))
	k := 2.0 / float64(period+1)
	result[0] = data[0]
	for i := 1; i < len(data); i++ {
		result[i] = data[i]*k + result[i-1]*(1-k)
	}
	return result
}

// ---- MACD面积计算 ----

func CalcMACDArea(klines []Kline, startIdx, endIdx int) float64 {
	if startIdx < 0 || endIdx >= len(klines) || startIdx >= endIdx {
		return 0
	}
	n := endIdx - startIdx + 1
	closes := make([]float64, n)
	for i := 0; i < n; i++ {
		closes[i] = klines[startIdx+i].Close
	}

	ema12 := calcEMA(closes, 12)
	ema26 := calcEMA(closes, 26)
	if ema12 == nil || ema26 == nil {
		return 0
	}

	dif := make([]float64, n)
	for i := 0; i < n; i++ {
		dif[i] = ema12[i] - ema26[i]
	}

	dea := calcEMA(dif, 9)
	if dea == nil {
		return 0
	}

	area := 0.0
	for i := 0; i < n; i++ {
		area += math.Abs(dif[i] - dea[i])
	}
	return area
}

// ---- 1. K线合并（包含处理） ----

func MergeKlines(klines []Kline) []Kline {
	if len(klines) < 2 {
		return klines
	}

	result := make([]Kline, len(klines))
	copy(result, klines)

	for {
		tmp := make([]Kline, 0, len(result))
		tmp = append(tmp, result[0])
		merged := false

		for i := 1; i < len(result); i++ {
			cur := result[i]
			prev := &tmp[len(tmp)-1]

			if containsKline(cur, *prev) {
				dir := getMergeDirection(tmp)
				if dir >= 0 {
					if cur.High > prev.High {
						prev.High = cur.High
					}
					if cur.Low > prev.Low {
						prev.Low = cur.Low
					}
				} else {
					if cur.High < prev.High {
						prev.High = cur.High
					}
					if cur.Low < prev.Low {
						prev.Low = cur.Low
					}
				}
				merged = true
			} else {
				tmp = append(tmp, cur)
			}
		}

		if !merged {
			break
		}
		result = tmp
	}

	return result
}

// ---- 2. 分型识别 ----

func FindFenXing(mergedKlines []Kline) []FenXing {
	if len(mergedKlines) < 3 {
		return nil
	}

	var rawFx []FenXing
	for i := 1; i < len(mergedKlines)-1; i++ {
		a, b, c := mergedKlines[i-1], mergedKlines[i], mergedKlines[i+1]

		if b.High > a.High && b.High > c.High &&
			b.Low >= a.Low && b.Low >= c.Low {
			rawFx = append(rawFx, FenXing{
				Index: i,
				Type:  "top",
				Date:  b.Date,
				High:  b.High,
				Low:   b.Low,
			})
		}

		if b.Low < a.Low && b.Low < c.Low &&
			b.High <= a.High && b.High <= c.High {
			rawFx = append(rawFx, FenXing{
				Index: i,
				Type:  "bottom",
				Date:  b.Date,
				High:  b.High,
				Low:   b.Low,
			})
		}
	}

	return filterFenXing(rawFx)
}

func filterFenXing(fx []FenXing) []FenXing {
	if len(fx) < 2 {
		return fx
	}

	clean := make([]FenXing, 0, len(fx))
	clean = append(clean, fx[0])

	for i := 1; i < len(fx); i++ {
		last := &clean[len(clean)-1]
		if fx[i].Type == last.Type {
			if fx[i].Type == "top" && fx[i].High > last.High {
				*last = fx[i]
			} else if fx[i].Type == "bottom" && fx[i].Low < last.Low {
				*last = fx[i]
			}
		} else {
			clean = append(clean, fx[i])
		}
	}

	return clean
}

// ---- 3. 笔 ----

func BuildBi(fenxings []FenXing, mergedKlines []Kline) []Bi {
	if len(fenxings) < 2 {
		return nil
	}

	cleanFx := filterFenXing(fenxings)
	if len(cleanFx) < 2 {
		return nil
	}

	var biList []Bi
	idx := 0

	for i := 0; i < len(cleanFx)-1; {
		fx1 := cleanFx[i]
		found := false
		for j := i + 1; j < len(cleanFx); j++ {
			fx2 := cleanFx[j]
			if fx1.Type != fx2.Type && fx2.Index-fx1.Index >= 4 {
				direction := "up"
				if fx1.Type == "top" {
					direction = "down"
				}

				high := fx1.High
				if fx2.High > high {
					high = fx2.High
				}
				low := fx1.Low
				if fx2.Low < low {
					low = fx2.Low
				}

				biList = append(biList, Bi{
					Index:     idx,
					Direction: direction,
					StartDate: fx1.Date,
					EndDate:   fx2.Date,
					StartIdx:  fx1.Index,
					EndIdx:    fx2.Index,
					High:      high,
					Low:       low,
				})
				idx++
				i = j
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	for i := 0; i < len(biList)-1; i++ {
		if biList[i].Direction != biList[i+1].Direction {
			biList[i].Confirmed = true
		}
	}

	return biList
}

// ---- 4. 中枢 ----

func BuildZhongShu(biList []Bi, mergedKlines []Kline) []ZhongShu {
	if len(biList) < 3 {
		return nil
	}

	var zsList []ZhongShu
	idx := 0

	for i := 0; i < len(biList)-2; {
		overlapLow := math.MaxFloat64
		overlapHigh := -math.MaxFloat64
		gg := -math.MaxFloat64
		dd := math.MaxFloat64

		lows := make([]float64, 3)
		highs := make([]float64, 3)
		for k := 0; k < 3; k++ {
			bi := biList[i+k]
			lows[k] = bi.Low
			highs[k] = bi.High
			if bi.High > gg {
				gg = bi.High
			}
			if bi.Low < dd {
				dd = bi.Low
			}
		}

		overlapLow = math.Max(lows[0], math.Max(lows[1], lows[2]))
		overlapHigh = math.Min(highs[0], math.Min(highs[1], highs[2]))

		if overlapLow >= overlapHigh {
			i++
			continue
		}

		lineCount := 3
		endJ := i + 2

		for j := i + 3; j < len(biList); j++ {
			bi := biList[j]
			if bi.Low < overlapHigh && bi.High > overlapLow {
				lineCount++
				endJ = j
				if bi.Low > overlapLow {
					overlapLow = bi.Low
				}
				if bi.High < overlapHigh {
					overlapHigh = bi.High
				}
				if bi.High > gg {
					gg = bi.High
				}
				if bi.Low < dd {
					dd = bi.Low
				}
			} else {
				break
			}
		}

		confirmed := false
		if endJ < len(biList)-1 {
			nextBi := biList[endJ+1]
			if nextBi.High < overlapLow || nextBi.Low > overlapHigh {
				confirmed = true
			}
		}

		zs := ZhongShu{
			Index:     idx,
			StartDate: biList[i].StartDate,
			EndDate:   biList[endJ].EndDate,
			ZG:        overlapHigh,
			ZD:        overlapLow,
			GG:        gg,
			DD:        dd,
			LineCount: lineCount,
			Direction: biList[i].Direction,
			Confirmed: confirmed,
		}
		zsList = append(zsList, zs)
		idx++
		i = endJ + 1
	}

	return zsList
}

// ---- 5. 线段 ----

func BuildXianDuan(biList []Bi, mergedKlines []Kline) []XianDuan {
	if len(biList) == 0 {
		return nil
	}

	var xdList []XianDuan
	idx := 0

	type segment struct {
		direction string
		startDate string
		endDate   string
		startIdx  int
		endIdx    int
		high      float64
		low       float64
		bis       []Bi
	}

	var segments []segment
	segments = append(segments, segment{
		direction: biList[0].Direction,
		startDate: biList[0].StartDate,
		endDate:   biList[0].EndDate,
		startIdx:  biList[0].StartIdx,
		endIdx:    biList[0].EndIdx,
		high:      biList[0].High,
		low:       biList[0].Low,
		bis:       []Bi{biList[0]},
	})

	for i := 1; i < len(biList); i++ {
		bi := biList[i]
		last := &segments[len(segments)-1]

		if bi.Direction == last.direction {
			last.endDate = bi.EndDate
			last.endIdx = bi.EndIdx
			if bi.High > last.high {
				last.high = bi.High
			}
			if bi.Low < last.low {
				last.low = bi.Low
			}
			last.bis = append(last.bis, bi)
		} else {
			newHigh := last.high
			if bi.High > newHigh {
				newHigh = bi.High
			}
			newLow := last.low
			if bi.Low < newLow {
				newLow = bi.Low
			}

			if len(last.bis) >= 2 {
				significant := false
				if bi.Direction == "up" {
					if bi.High > last.high {
						significant = true
					}
				} else {
					if bi.Low < last.low {
						significant = true
					}
				}
				if significant {
					segments = append(segments, segment{
						direction: bi.Direction,
						startDate: bi.StartDate,
						endDate:   bi.EndDate,
						startIdx:  bi.StartIdx,
						endIdx:    bi.EndIdx,
						high:      bi.High,
						low:       bi.Low,
						bis:       []Bi{bi},
					})
					continue
				}
			}

			last.endDate = bi.EndDate
			last.endIdx = bi.EndIdx
			last.high = newHigh
			last.low = newLow
			last.bis = append(last.bis, bi)
		}
	}

	for _, seg := range segments {
		if len(seg.bis) >= 1 {
			xdList = append(xdList, XianDuan{
				Index:     idx,
				Direction: seg.direction,
				StartDate: seg.startDate,
				EndDate:   seg.endDate,
				High:      seg.high,
				Low:       seg.low,
			})
			idx++
		}
	}

	return xdList
}

// ---- 6. 买卖点 ----

func FindMaiMaiDian(biList []Bi, zhongShuList []ZhongShu, mergedKlines []Kline) []MaiMaiDian {
	if len(biList) == 0 {
		return nil
	}

	var mmList []MaiMaiDian
	idx := 0

	// 一类买点: 下跌趋势末端，中枢下方出现力度衰减的低点
	// 检查每个底分型对应的 "up" 笔起点
	for i := 0; i < len(biList); i++ {
		bi := biList[i]

		if i == 0 {
			continue
		}

		if bi.Direction == "up" && biList[i-1].Direction == "down" {
			prevBi := biList[i-1]
			for _, zs := range zhongShuList {
				if prevBi.Low < zs.ZD && bi.StartIdx > 0 {
					area1 := CalcMACDArea(mergedKlines, prevBi.StartIdx, prevBi.EndIdx)
					// 查找更早期的同向笔来比较衰减
					earlyArea := 0.0
					for k := i - 3; k >= 0; k -= 2 {
						if biList[k].Direction == "down" {
							earlyArea = CalcMACDArea(mergedKlines, biList[k].StartIdx, biList[k].EndIdx)
							break
						}
					}
					if earlyArea > 0 && area1 < earlyArea {
						mmList = append(mmList, MaiMaiDian{
							Index:  idx,
							Type:   "1buy",
							Date:   bi.StartDate,
							Price:  prevBi.Low,
							Reason: "下跌趋势末端，中枢下方力度衰减",
						})
						idx++
						break
					}
				}
			}
		}

		// 一类卖点: 上涨趋势末端，中枢上方力度衰减的高点
		if bi.Direction == "down" && biList[i-1].Direction == "up" {
			prevBi := biList[i-1]
			for _, zs := range zhongShuList {
				if prevBi.High > zs.ZG {
					area1 := CalcMACDArea(mergedKlines, prevBi.StartIdx, prevBi.EndIdx)
					earlyArea := 0.0
					for k := i - 3; k >= 0; k -= 2 {
						if biList[k].Direction == "up" {
							earlyArea = CalcMACDArea(mergedKlines, biList[k].StartIdx, biList[k].EndIdx)
							break
						}
					}
					if earlyArea > 0 && area1 < earlyArea {
						mmList = append(mmList, MaiMaiDian{
							Index:  idx,
							Type:   "1sell",
							Date:   bi.StartDate,
							Price:  prevBi.High,
							Reason: "上涨趋势末端，中枢上方力度衰减",
						})
						idx++
						break
					}
				}
			}
		}
	}

	// 二类买点: 一类买点后的回调不创新低
	for i := 0; i < len(mmList); i++ {
		if mmList[i].Type == "1buy" {
			for _, bi := range biList {
				if bi.Direction == "down" && bi.StartDate > mmList[i].Date {
					if bi.Low > mmList[i].Price {
						mmList = append(mmList, MaiMaiDian{
							Index:  idx,
							Type:   "2buy",
							Date:   bi.EndDate,
							Price:  bi.Low,
							Reason: "一类买点后回调不创新低",
						})
						idx++
					}
					break
				}
			}
		}
		if mmList[i].Type == "1sell" {
			for _, bi := range biList {
				if bi.Direction == "up" && bi.StartDate > mmList[i].Date {
					if bi.High < mmList[i].Price {
						mmList = append(mmList, MaiMaiDian{
							Index:  idx,
							Type:   "2sell",
							Date:   bi.EndDate,
							Price:  bi.High,
							Reason: "一类卖点后反弹不创新高",
						})
						idx++
					}
					break
				}
			}
		}
	}

	// 三类买点: 回调不进入中枢上沿 (ZG)
	// 三类卖点: 反弹不进入中枢下沿 (ZD)
	for _, zs := range zhongShuList {
		if !zs.Confirmed {
			continue
		}
		for _, bi := range biList {
			if bi.Direction == "down" && bi.StartDate > zs.EndDate {
				if bi.Low > zs.ZG {
					mmList = append(mmList, MaiMaiDian{
						Index:  idx,
						Type:   "3buy",
						Date:   bi.EndDate,
						Price:  bi.Low,
						Reason: fmt.Sprintf("回调不进入中枢%d上沿(ZG=%.2f)", zs.Index, zs.ZG),
					})
					idx++
					break
				}
			}
			if bi.Direction == "up" && bi.StartDate > zs.EndDate {
				if bi.High < zs.ZD {
					mmList = append(mmList, MaiMaiDian{
						Index:  idx,
						Type:   "3sell",
						Date:   bi.EndDate,
						Price:  bi.High,
						Reason: fmt.Sprintf("反弹不进入中枢%d下沿(ZD=%.2f)", zs.Index, zs.ZD),
					})
					idx++
					break
				}
			}
		}
	}

	return mmList
}

// ---- 7. 背驰 ----

func FindBeiChi(biList []Bi, zhongShuList []ZhongShu, mergedKlines []Kline) []BeiChi {
	var bcList []BeiChi
	idx := 0

	// 笔背驰: 同向相邻两笔比较，后一笔MACD面积 < 前一笔
	for i := 0; i < len(biList)-2; i++ {
		for j := i + 1; j < len(biList); j++ {
			if biList[i].Direction == biList[j].Direction {
				area1 := CalcMACDArea(mergedKlines, biList[i].StartIdx, biList[i].EndIdx)
				area2 := CalcMACDArea(mergedKlines, biList[j].StartIdx, biList[j].EndIdx)
				if area2 > 0 && area2 < area1 {
					bcList = append(bcList, BeiChi{
						Index: idx,
						Type:  "bi",
						Desc: fmt.Sprintf("%s笔%d(%s-%s)MACD面积%.2f < %s笔%d(%s-%s)面积%.2f",
							biList[j].Direction, biList[j].Index,
							biList[j].StartDate, biList[j].EndDate, area2,
							biList[i].Direction, biList[i].Index,
							biList[i].StartDate, biList[i].EndDate, area1),
					})
					idx++
				}
				break
			}
		}
	}

	// 盘整背驰: 同一中枢内，末笔MACD面积 < 首笔
	for _, zs := range zhongShuList {
		var biInZS []Bi
		for _, bi := range biList {
			if bi.StartDate >= zs.StartDate && bi.EndDate <= zs.EndDate {
				biInZS = append(biInZS, bi)
			}
		}
		if len(biInZS) >= 2 {
			first := biInZS[0]
			last := biInZS[len(biInZS)-1]
			if first.Direction == last.Direction {
				area1 := CalcMACDArea(mergedKlines, first.StartIdx, first.EndIdx)
				area2 := CalcMACDArea(mergedKlines, last.StartIdx, last.EndIdx)
				if area2 > 0 && area2 < area1 {
					bcList = append(bcList, BeiChi{
						Index: idx,
						Type:  "pz",
						Desc: fmt.Sprintf("中枢%d内盘整背驰：末笔MACD面积%.2f < 首笔%.2f",
							zs.Index, area2, area1),
					})
					idx++
				}
			}
		}
	}

	// 趋势背驰: 两个同向中枢之间，后一中枢离开力度 < 前一中枢离开力度
	for i := 0; i < len(zhongShuList)-1; i++ {
		zs1 := zhongShuList[i]
		zs2 := zhongShuList[i+1]
		if zs1.Direction != zs2.Direction {
			continue
		}

		var leaveBi1, leaveBi2 *Bi
		for k := range biList {
			if biList[k].StartDate > zs1.EndDate && biList[k].Direction == zs1.Direction {
				if leaveBi1 == nil {
					leaveBi1 = &biList[k]
				}
			}
			if biList[k].StartDate > zs2.EndDate && biList[k].Direction == zs2.Direction {
				if leaveBi2 == nil {
					leaveBi2 = &biList[k]
				}
			}
		}

		if leaveBi1 != nil && leaveBi2 != nil {
			area1 := CalcMACDArea(mergedKlines, leaveBi1.StartIdx, leaveBi1.EndIdx)
			area2 := CalcMACDArea(mergedKlines, leaveBi2.StartIdx, leaveBi2.EndIdx)
			if area2 > 0 && area2 < area1 {
				bcList = append(bcList, BeiChi{
					Index: idx,
					Type:  "qs",
					Desc: fmt.Sprintf("趋势背驰：中枢%d离开力度(%.2f) < 中枢%d离开力度(%.2f)",
						zs2.Index, area2, zs1.Index, area1),
				})
				idx++
			}
		}
	}

	return bcList
}

// ---- 8. 主分析函数 ----

func Analyze(klines []Kline) *ChanLunResult {
	result := &ChanLunResult{
		OrigCount: len(klines),
	}

	if len(klines) == 0 {
		return result
	}

	// 步骤1: K线合并
	merged := MergeKlines(klines)
	result.MergedCount = len(merged)

	// 步骤2: 分型识别
	fenxings := FindFenXing(merged)
	result.FenXingCount = len(fenxings)

	// 步骤3: 笔
	biList := BuildBi(fenxings, merged)
	result.BiList = biList

	// 步骤4: 中枢
	zhongShuList := BuildZhongShu(biList, merged)
	result.ZhongShuList = zhongShuList

	// 步骤5: 线段
	xianDuanList := BuildXianDuan(biList, merged)
	result.XianDuanList = xianDuanList

	// 步骤6: 买卖点
	maiMaiDianList := FindMaiMaiDian(biList, zhongShuList, merged)
	result.MaiMaiDianList = maiMaiDianList

	// 步骤7: 背驰
	beiChiList := FindBeiChi(biList, zhongShuList, merged)
	result.BeiChiList = beiChiList

	return result
}

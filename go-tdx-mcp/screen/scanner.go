package screen

import (
	"github.com/tdx/go-tdx-mcp/indicator"
)

type ScanResult struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	SignalDate string  `json:"signal_date"`
	LastClose  float64 `json:"last_close"`
	SignalType string  `json:"signal_type"`
}

type Scanner struct {
	cache map[string]float64
}

func NewScanner() *Scanner {
	return &Scanner{
		cache: make(map[string]float64),
	}
}

type ScanPredicate func(bars []indicator.Bar) (bool, string)

func (s *Scanner) ScanBars(code string, bars []indicator.Bar, predicates []ScanPredicate) []ScanResult {
	var results []ScanResult
	if len(bars) < 2 {
		return results
	}
	for _, pred := range predicates {
		if hit, sigType := pred(bars); hit {
			results = append(results, ScanResult{
				Code:       code,
				LastClose:  bars[len(bars)-1].Close,
				SignalType: sigType,
			})
		}
	}
	return results
}

func (s *Scanner) CacheResult(code string, score float64) {
	s.cache[code] = score
}

func (s *Scanner) GetCache(code string) (float64, bool) {
	v, ok := s.cache[code]
	return v, ok
}

func (s *Scanner) ClearCache() {
	s.cache = make(map[string]float64)
}

func MACDGoldenPredicate(bars []indicator.Bar) (bool, string) {
	if len(bars) < 30 {
		return false, ""
	}
	r := indicator.MACD(bars, 12, 26, 9)
	n := len(r.Values)
	if n < 2 {
		return false, ""
	}
	if r.Values[n-1] <= 0 || r.Line2[n-1] <= 0 {
		return false, ""
	}
	if r.Values[n-2] <= r.Line2[n-2] && r.Values[n-1] > r.Line2[n-1] {
		return true, "macd_golden_cross"
	}
	return false, ""
}

func KDJGoldenPredicate(bars []indicator.Bar) (bool, string) {
	if len(bars) < 15 {
		return false, ""
	}
	r := indicator.KDJ(bars, 9, 3, 3)
	n := len(r.Values)
	if n < 2 {
		return false, ""
	}
	if r.Values[n-1] <= 0 || r.Line2[n-1] <= 0 {
		return false, ""
	}
	if r.Values[n-2] <= r.Line2[n-2] && r.Values[n-1] > r.Line2[n-1] {
		return true, "kdj_golden_cross"
	}
	return false, ""
}

func VolumeBreakoutPredicate(bars []indicator.Bar) (bool, string) {
	if len(bars) < 6 {
		return false, ""
	}
	n := len(bars)
	avgVol := avgVol(bars, n-2, 5)
	if avgVol <= 0 {
		return false, ""
	}
	if bars[n-1].Vol > avgVol*2 && bars[n-1].Close > bars[n-2].Close {
		return true, "volume_breakout"
	}
	return false, ""
}

func MAConvergencePredicate(bars []indicator.Bar) (bool, string) {
	if len(bars) < 20 {
		return false, ""
	}
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	ma5 := indicator.MA(closes, 5)
	ma10 := indicator.MA(closes, 10)
	n := len(bars)
	if ma5[n-1] <= 0 || ma10[n-1] <= 0 {
		return false, ""
	}
	if ma5[n-2] <= ma10[n-2] && ma5[n-1] > ma10[n-1] {
		return true, "ma_convergence"
	}
	return false, ""
}

func DefaultScanPredicates() []ScanPredicate {
	return []ScanPredicate{
		MACDGoldenPredicate,
		KDJGoldenPredicate,
		VolumeBreakoutPredicate,
		MAConvergencePredicate,
	}
}

func avgVol(bars []indicator.Bar, end, period int) float64 {
	start := end - period + 1
	if start < 0 {
		start = 0
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

package screen

import (
	"math"
	"sort"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type StrengthResult struct {
	Code      string  `json:"code"`
	Strength  float64 `json:"strength"`
	Ret5D     float64 `json:"ret_5d"`
	Ret20D    float64 `json:"ret_20d"`
	Ret60D    float64 `json:"ret_60d"`
	Vol20D    float64 `json:"vol_20d"`
}

type StrengthPreset struct {
	Name     string
	W5       float64
	W20      float64
	W60      float64
	Penalize bool
}

var StrengthPresets = map[string]StrengthPreset{
	"steady":   {Name: "steady", W5: 0.2, W20: 0.3, W60: 0.5, Penalize: true},
	"breakout": {Name: "breakout", W5: 0.6, W20: 0.3, W60: 0.1, Penalize: false},
	"balanced": {Name: "balanced", W5: 0.34, W20: 0.33, W60: 0.33, Penalize: true},
}

type StrengthRanker struct {
	bars    map[string][]indicator.Bar
	preset  StrengthPreset
}

func NewStrengthRanker(presetName string) *StrengthRanker {
	preset, ok := StrengthPresets[presetName]
	if !ok {
		preset = StrengthPresets["balanced"]
	}
	return &StrengthRanker{
		bars:   make(map[string][]indicator.Bar),
		preset: preset,
	}
}

func (s *StrengthRanker) AddBars(code string, bars []indicator.Bar) {
	s.bars[code] = bars
}

func (s *StrengthRanker) Rank(topN int) []StrengthResult {
	var results []StrengthResult
	for code, bars := range s.bars {
		sr := s.calcStrength(code, bars)
		results = append(results, sr)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Strength > results[j].Strength
	})

	if topN > 0 && topN < len(results) {
		results = results[:topN]
	}
	return results
}

func (s *StrengthRanker) calcStrength(code string, bars []indicator.Bar) StrengthResult {
	n := len(bars)
	r := StrengthResult{Code: code}

	if n < 60 {
		return r
	}

	if bars[n-1].Close > 0 && bars[n-6].Close > 0 {
		r.Ret5D = bars[n-1].Close/bars[n-6].Close - 1
	}
	if bars[n-1].Close > 0 && bars[n-21].Close > 0 {
		r.Ret20D = bars[n-1].Close/bars[n-21].Close - 1
	}
	if bars[n-1].Close > 0 && bars[n-61].Close > 0 {
		r.Ret60D = bars[n-1].Close/bars[n-61].Close - 1
	}

	var volSum float64
	start := n - 20
	if start < 0 {
		start = 0
	}
	for i := start; i < n; i++ {
		ret := 0.0
		if i > 0 && bars[i-1].Close > 0 {
			ret = bars[i].Close/bars[i-1].Close - 1
		}
		volSum += ret * ret
	}
	if n-start > 0 {
		r.Vol20D = math.Sqrt(volSum / float64(n-start))
	}

	p := s.preset
	strength := p.W5*r.Ret5D + p.W20*r.Ret20D + p.W60*r.Ret60D
	if p.Penalize && r.Vol20D > 0 {
		strength /= r.Vol20D
	}
	r.Strength = strength

	return r
}

func CalcSimpleStrength(bars []indicator.Bar) float64 {
	n := len(bars)
	if n < 60 {
		return 0
	}
	ret5 := 0.0
	if bars[n-6].Close > 0 {
		ret5 = bars[n-1].Close/bars[n-6].Close - 1
	}
	ret20 := 0.0
	if bars[n-21].Close > 0 {
		ret20 = bars[n-1].Close/bars[n-21].Close - 1
	}
	ret60 := 0.0
	if bars[n-61].Close > 0 {
		ret60 = bars[n-1].Close/bars[n-61].Close - 1
	}
	return 0.34*ret5 + 0.33*ret20 + 0.33*ret60
}

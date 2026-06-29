package backtest

import "math"

type SlippageModel interface {
	Apply(price, volume float64, isBuy bool) float64
}

type FixedSlippage struct {
	Value float64
}

func (s *FixedSlippage) Apply(price, _ float64, _ bool) float64 {
	if s.Value >= 0 {
		return price + s.Value
	}
	return price
}

type PercentSlippage struct {
	Rate float64
}

func (s *PercentSlippage) Apply(price, _ float64, isBuy bool) float64 {
	if isBuy {
		return price * (1 + s.Rate)
	}
	return price * (1 - s.Rate)
}

type SquareRootSlippage struct {
	Base  float64
	Scale float64
}

func (s *SquareRootSlippage) Apply(price, volume float64, isBuy bool) float64 {
	if volume <= 0 {
		return price
	}
	delta := s.Base + s.Scale*math.Sqrt(volume)
	if isBuy {
		return price * (1 + delta)
	}
	return price * (1 - delta)
}

type VolumeRatioSlippage struct {
	Base float64
}

func (s *VolumeRatioSlippage) Apply(price, volume float64, isBuy bool) float64 {
	if volume <= 0 {
		return price
	}
	impact := s.Base * math.Log1p(volume)
	if isBuy {
		return price * (1 + impact)
	}
	return price * (1 - impact)
}

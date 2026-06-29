package portfolio

import (
	"math"
	"sort"
)

type Optimizer interface {
	Name() string
	Optimize(factorScores map[string]float64, riskModel *RiskModel, maxPositions int) WeightMap
}

type EqualWeightOptimizer struct{}

func (e *EqualWeightOptimizer) Name() string { return "equal_weight" }

func (e *EqualWeightOptimizer) Optimize(factorScores map[string]float64, _ *RiskModel, maxPositions int) WeightMap {
	if len(factorScores) == 0 {
		return nil
	}

	type codeScore struct {
		code  string
		score float64
	}
	list := make([]codeScore, 0, len(factorScores))
	for code, score := range factorScores {
		list = append(list, codeScore{code, score})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })

	if maxPositions <= 0 || maxPositions > len(list) {
		maxPositions = len(list)
	}
	selected := list[:maxPositions]

	weights := make(WeightMap)
	each := 1.0 / float64(len(selected))
	for _, item := range selected {
		weights[item.code] = each
	}
	return weights
}

type FactorWeightedOptimizer struct{}

func (f *FactorWeightedOptimizer) Name() string { return "factor_weighted" }

func (f *FactorWeightedOptimizer) Optimize(factorScores map[string]float64, _ *RiskModel, maxPositions int) WeightMap {
	if len(factorScores) == 0 {
		return nil
	}

	type codeScore struct {
		code  string
		score float64
	}
	list := make([]codeScore, 0, len(factorScores))
	for code, score := range factorScores {
		list = append(list, codeScore{code, score})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })

	if maxPositions <= 0 || maxPositions > len(list) {
		maxPositions = len(list)
	}
	selected := list[:maxPositions]

	var totalScore float64
	for _, item := range selected {
		totalScore += item.score
	}

	weights := make(WeightMap)
	if totalScore > 0 {
		for _, item := range selected {
			weights[item.code] = item.score / totalScore
		}
	} else {
		each := 1.0 / float64(len(selected))
		for _, item := range selected {
			weights[item.code] = each
		}
	}
	return weights
}

type RiskParityOptimizer struct{}

func (r *RiskParityOptimizer) Name() string { return "risk_parity" }

func (r *RiskParityOptimizer) Optimize(factorScores map[string]float64, riskModel *RiskModel, maxPositions int) WeightMap {
	type codeScore struct {
		code  string
		score float64
	}
	list := make([]codeScore, 0, len(factorScores))
	for code, score := range factorScores {
		list = append(list, codeScore{code, score})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })

	if maxPositions <= 0 || maxPositions > len(list) {
		maxPositions = len(list)
	}
	selected := list[:maxPositions]

	weights := make(WeightMap)
	if riskModel != nil && len(selected) > 0 {
		for _, item := range selected {
			contrib := riskModel.GetRiskContribution(item.code)
			weights[item.code] = 1.0 / math.Max(contrib, 0.001)
		}
	} else {
		each := 1.0 / float64(len(selected))
		for _, item := range selected {
			weights[item.code] = each
		}
	}

	var total float64
	for _, w := range weights {
		total += w
	}
	if total > 0 {
		for k, w := range weights {
			weights[k] = w / total
		}
	}

	return weights
}

type MeanVarianceOptimizer struct{}

func (m *MeanVarianceOptimizer) Name() string { return "mean_variance" }

func (m *MeanVarianceOptimizer) Optimize(factorScores map[string]float64, riskModel *RiskModel, maxPositions int) WeightMap {
	type codeScore struct {
		code  string
		score float64
	}
	list := make([]codeScore, 0, len(factorScores))
	for code, score := range factorScores {
		list = append(list, codeScore{code, score})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })

	if maxPositions <= 0 || maxPositions > len(list) {
		maxPositions = len(list)
	}
	selected := list[:maxPositions]

	weights := make(WeightMap)
	if riskModel != nil && len(selected) > 1 {
		var sumInvRisk float64
		for _, item := range selected {
			risk := riskModel.GetVolatility(item.code)
			if risk > 0 {
				weights[item.code] = item.score / risk
			} else {
				weights[item.code] = item.score / 0.01
			}
			sumInvRisk += weights[item.code]
		}
		if sumInvRisk > 0 {
			for k := range weights {
				weights[k] /= sumInvRisk
			}
		}
	} else {
		each := 1.0 / float64(len(selected))
		for _, item := range selected {
			weights[item.code] = each
		}
	}

	return weights
}

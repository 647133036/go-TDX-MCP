package factor

import (
	"math"
	"sort"
)

func Winsorize(values []float64, method string, threshold float64) []float64 {
	n := len(values)
	result := make([]float64, n)
	copy(result, values)

	valid := make([]float64, 0)
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			valid = append(valid, v)
		}
	}
	if len(valid) < 3 {
		return result
	}

	sort.Float64s(valid)

	var lower, upper float64
	switch method {
	case "mad":
		median := valid[len(valid)/2]
		deviations := make([]float64, len(valid))
		for i, v := range valid {
			deviations[i] = math.Abs(v - median)
		}
		sort.Float64s(deviations)
		mad := deviations[len(deviations)/2] * 1.4826
		lower = median - threshold*mad
		upper = median + threshold*mad
	case "sigma":
		mean := meanFloat(valid)
		std := stdFloat(valid, mean)
		lower = mean - threshold*std
		upper = mean + threshold*std
	case "percentile":
		lower = valid[len(valid)*25/1000]
		upper = valid[len(valid)*975/1000]
	default:
		return result
	}

	for i, v := range result {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if v < lower {
			result[i] = lower
		} else if v > upper {
			result[i] = upper
		}
	}
	return result
}

func ZScore(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	validVals := make([]float64, 0)
	validIdx := make([]int, 0)
	for i, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			validVals = append(validVals, v)
			validIdx = append(validIdx, i)
		} else {
			result[i] = math.NaN()
		}
	}
	if len(validVals) < 2 {
		return result
	}

	mean := meanFloat(validVals)
	std := stdFloat(validVals, mean)
	if std == 0 {
		for _, idx := range validIdx {
			result[idx] = 0
		}
		return result
	}
	for j, idx := range validIdx {
		result[idx] = (validVals[j] - mean) / std
	}
	return result
}

func RankNormalize(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	type indexed struct {
		val float64
		idx int
	}
	items := make([]indexed, 0, n)
	for i, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			items = append(items, indexed{v, i})
		}
	}
	if len(items) == 0 {
		return result
	}
	sort.Slice(items, func(i, j int) bool { return items[i].val < items[j].val })

	m := len(items)
	for rank := 0; rank < m; rank++ {
		result[items[rank].idx] = float64(rank+1) / float64(m)
	}
	return result
}

func FillMissing(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)
	copy(result, values)

	var lastValid float64
	found := false
	for i := 0; i < n; i++ {
		if math.IsNaN(result[i]) || math.IsInf(result[i], 0) {
			if found {
				result[i] = lastValid
			}
		} else {
			lastValid = result[i]
			found = true
		}
	}
	return result
}

func Orthogonalize(values, reference []float64) []float64 {
	n := len(values)
	if n != len(reference) {
		return values
	}
	result := make([]float64, n)

	validX := make([]float64, 0)
	validY := make([]float64, 0)
	validIdx := make([]int, 0)
	for i := 0; i < n; i++ {
		if !math.IsNaN(values[i]) && !math.IsInf(values[i], 0) &&
			!math.IsNaN(reference[i]) && !math.IsInf(reference[i], 0) {
			validX = append(validX, reference[i])
			validY = append(validY, values[i])
			validIdx = append(validIdx, i)
		} else {
			result[i] = math.NaN()
		}
	}
	if len(validX) < 2 {
		copy(result, values)
		return result
	}

	beta := linearRegressionBeta(validX, validY)
	for j, idx := range validIdx {
		result[idx] = validY[j] - beta*validX[j]
	}
	return result
}

func Preprocess(values []float64) []float64 {
	result := Winsorize(values, "mad", 3)
	result = ZScore(result)
	result = FillMissing(result)
	return result
}

func linearRegressionBeta(x, y []float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}
	denom := float64(n)*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (float64(n)*sumXY - sumX*sumY) / denom
}

func meanFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stdFloat(vals []float64, mean float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)-1))
}

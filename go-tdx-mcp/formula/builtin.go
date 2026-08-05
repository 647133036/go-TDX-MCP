package formula

import (
	"fmt"
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// registerBuiltin registers the base built-in function library
// (ported from formula-go, ISC licensed). Drawing functions are
// registered here as well; their results carry DrawingEvents.
func (r *FunctionRegistry) registerBuiltin() {
	// Mathematical functions
	r.Register("MAX", "math", "2", "返回两数较大值", fnMAX)
	r.Register("MIN", "math", "2", "返回两数较小值", fnMIN)
	r.Register("ABS", "math", "1", "绝对值", fnABS)
	r.Register("SQRT", "math", "1", "平方根", fnSQRT)
	r.Register("POW", "math", "2", "幂运算", fnPOW)
	r.Register("EXP", "math", "1", "e 的幂", fnEXP)
	r.Register("LN", "math", "1", "自然对数", fnLN)
	r.Register("LOG", "math", "1", "以 10 为底的对数", fnLOG)
	r.Register("MOD", "math", "2", "取模", fnMOD)
	r.Register("CEILING", "math", "1", "向上取整", fnCEILING)
	r.Register("FLOOR", "math", "1", "向下取整", fnFLOOR)
	r.Register("INTPART", "math", "1", "取整数部分", fnINTPART)
	r.Register("FRACPART", "math", "1", "取小数部分", fnFRACPART)
	r.Register("ROUND", "math", "1", "四舍五入", fnROUND)
	r.Register("ROUND2", "math", "2", "指定小数位四舍五入", fnROUND2)
	r.Register("SIGN", "math", "1", "符号函数", fnSIGN)
	r.Register("SIN", "math", "1", "正弦", fnSIN)
	r.Register("COS", "math", "1", "余弦", fnCOS)
	r.Register("TAN", "math", "1", "正切", fnTAN)
	r.Register("ASIN", "math", "1", "反正弦", fnASIN)
	r.Register("ACOS", "math", "1", "反余弦", fnACOS)
	r.Register("ATAN", "math", "1", "反正切", fnATAN)

	// Moving average / statistics functions
	r.Register("MA", "math", "2", "简单移动平均", fnMA)
	r.Register("EMA", "math", "2", "指数移动平均", fnEMA)
	r.Register("SMA", "math", "2~3", "通达信递归平滑均线", fnSMA)
	r.Register("WMA", "math", "2", "加权移动平均", fnWMA)
	r.Register("DMA", "math", "2", "动态移动平均", fnDMA)
	r.Register("SUM", "math", "2", "周期求和", fnSUM)
	r.Register("STD", "math", "2", "总体标准差", fnSTD)
	r.Register("STDP", "math", "2", "总体标准差", fnSTDP)
	r.Register("STDDEV", "math", "2", "样本标准差", fnSTDDEV)
	r.Register("VAR", "math", "2", "总体方差", fnVAR)
	r.Register("VARP", "math", "2", "总体方差", fnVARP)
	r.Register("DEVSQ", "math", "2", "偏差平方和", fnDEVSQ)
	r.Register("AVEDEV", "math", "2", "平均绝对偏差", fnAVEDEV)
	r.Register("FORCAST", "math", "2", "线性回归预测", fnFORCAST)
	r.Register("SLOPE", "math", "2", "线性回归斜率", fnSLOPE)
	r.Register("COVAR", "math", "3", "协方差", fnCOVAR)
	r.Register("RELATE", "math", "3", "相关系数", fnRELATE)
	r.Register("BETA", "math", "3", "贝塔系数", fnBETA)

	// Reference functions
	r.Register("REF", "reference", "2", "向前引用 N 周期前数据", fnREF)
	r.Register("REFV", "reference", "2", "向前引用（无未来函数标记）", fnREFV)
	r.Register("REFX", "reference", "2", "向后引用 N 周期后数据", fnREFX)
	r.Register("REFXV", "reference", "2", "向后引用（无未来函数标记）", fnREFXV)
	r.Register("HHV", "reference", "2", "周期内最高值", fnHHV)
	r.Register("LLV", "reference", "2", "周期内最低值", fnLLV)
	r.Register("HHVBARS", "reference", "2", "距周期内最高值周期数", fnHHVBARS)
	r.Register("LLVBARS", "reference", "2", "距周期内最低值周期数", fnLLVBARS)
	r.Register("CURRBARSCOUNT", "reference", "0", "当前 K 线距最后一根 K 线周期数", fnCURRBARSCOUNT)
	r.Register("TOTALBARSCOUNT", "reference", "0", "总 K 线数", fnTOTALBARSCOUNT)
	r.Register("ISLASTBAR", "reference", "0", "是否为最后一根 K 线", fnISLASTBAR)
	r.Register("BARSTATUS", "reference", "0", "K 线位置状态", fnBARSTATUS)
	r.Register("SUMBARS", "reference", "2", "累加到目标值所需周期数", fnSUMBARS)

	// Logical / conditional functions
	r.Register("CROSS", "logical", "2", "上穿判断", fnCROSS)
	r.Register("LONGCROSS", "logical", "3", "持续上穿判断", fnLONGCROSS)
	r.Register("IF", "logical", "3", "条件选择", fnIF)
	r.Register("IFF", "logical", "3", "条件选择（同 IF）", fnIFF)
	r.Register("IFN", "logical", "3", "反向条件选择", fnIFN)
	r.Register("NOT", "logical", "1", "逻辑取反", fnNOT)
	r.Register("BETWEEN", "logical", "3", "区间判断", fnBETWEEN)
	r.Register("RANGE", "logical", "3", "区间判断（同 BETWEEN）", fnRANGE)
	r.Register("CONST", "logical", "1", "以最后一个值填充", fnCONST)
	r.Register("VALUEWHEN", "logical", "2", "条件成立时取值", fnVALUEWHEN)
	r.Register("DRAWNULL", "logical", "0", "返回空值用于断线/隐藏", fnDRAWNULL)
	r.Register("COUNT", "logical", "2", "周期内条件成立次数", fnCOUNT)
	r.Register("EVERY", "logical", "2", "周期内条件全部成立", fnEVERY)
	r.Register("EXIST", "logical", "2", "周期内条件成立过", fnEXIST)
	r.Register("EXISTR", "logical", "3", "区间内条件成立过", fnEXISTR)
	r.Register("BARSLAST", "logical", "1", "距上次条件成立周期数", fnBARSLAST)
	r.Register("BARSCOUNT", "logical", "1", "有效数据周期数", fnBARSCOUNT)
	r.Register("BARSSINCE", "logical", "1", "距首次条件成立周期数", fnBARSSINCE)
	r.Register("BARSLASTCOUNT", "logical", "1", "条件连续成立周期数", fnBARSLASTCOUNT)
	r.Register("UPNDAY", "logical", "2", "连续上涨 N 日", fnUPNDAY)
	r.Register("DOWNNDAY", "logical", "2", "连续下跌 N 日", fnDOWNNDAY)
	r.Register("NDAY", "logical", "3", "连续 N 日 A 大于 B", fnNDAY)
	r.Register("LAST", "logical", "3", "区间内条件持续成立", fnLAST)
	r.Register("FILTER", "logical", "2", "信号过滤", fnFILTER)

	// Drawing functions
	r.Register("DRAWTEXT", "draw", "3", "文字标注", fnDRAWTEXT)
	r.Register("DRAWICON", "draw", "3", "图标标注", fnDRAWICON)
	r.Register("DRAWNUMBER", "draw", "3", "数字标注", fnDRAWNUMBER)
	r.Register("STICKLINE", "draw", "5", "柱线", fnSTICKLINE)
	r.Register("DRAWLINE", "draw", "5", "连接直线", fnDRAWLINE)
	r.Register("POLYLINE", "draw", "2", "折线", fnPOLYLINE)
	r.Register("DRAWBAND", "draw", "4", "带状区域", fnDRAWBAND)
	r.Register("DRAWKLINE", "draw", "4", "自定义 K 线", fnDRAWKLINE)

	r.registerComplexBuiltin()
}

// fnMA implements Moving Average: MA(data, period)
func fnMA(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("MA requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("MA first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("MA second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("MA period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	sum := 0.0
	nanCount := 0
	for i, value := range data.Array {
		if math.IsNaN(value) {
			nanCount++
		} else {
			sum += value
		}
		if i >= n {
			outgoing := data.Array[i-n]
			if math.IsNaN(outgoing) {
				nanCount--
			} else {
				sum -= outgoing
			}
		}
		if i < n-1 || nanCount > 0 {
			result[i] = math.NaN()
			continue
		}
		result[i] = sum / float64(n)
	}
	return NewArrayValue(result), nil
}

// fnEMA implements Exponential Moving Average: EMA(data, period).
// Seeding follows the project indicator package convention (SMA seed over the
// first period values) so that formulas like EMA(C,12)-EMA(C,26) match
// indicator.MACD exactly.
func fnEMA(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("EMA requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("EMA first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("EMA second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("EMA period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	alpha := 2.0 / float64(n+1)
	var seedSum float64
	for i := 0; i < n && i < len(data.Array); i++ {
		seedSum += data.Array[i]
	}
	if len(data.Array) >= n {
		result[n-1] = seedSum / float64(n)
	}
	for i := n; i < len(data.Array); i++ {
		result[i] = alpha*data.Array[i] + (1-alpha)*result[i-1]
	}
	return NewArrayValue(result), nil
}

// fnSUM implements Sum: SUM(data, period)
func fnSUM(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("SUM requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("SUM first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("SUM second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("SUM period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	sum := 0.0
	nanCount := 0
	for i, value := range data.Array {
		if math.IsNaN(value) {
			nanCount++
		} else {
			sum += value
		}
		if i >= n {
			outgoing := data.Array[i-n]
			if math.IsNaN(outgoing) {
				nanCount--
			} else {
				sum -= outgoing
			}
		}
		if i < n-1 || nanCount > 0 {
			result[i] = math.NaN()
			continue
		}
		result[i] = sum
	}
	return NewArrayValue(result), nil
}

// fnMAX implements Max: MAX(a, b)
func fnMAX(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("MAX requires 2 arguments")
	}
	a, b := args[0], args[1]
	if a.IsString || b.IsString || a.IsDraw || b.IsDraw {
		return nil, NewEvalError("MAX arguments must be numeric")
	}
	if !a.IsArray && !b.IsArray {
		return NewSingleValue(math.Max(a.Single, b.Single)), nil
	}
	if a.IsArray && b.IsArray {
		if len(a.Array) != len(b.Array) {
			return nil, NewEvalError("MAX: array length mismatch")
		}
		result := make([]float64, len(a.Array))
		for i := range a.Array {
			result[i] = math.Max(a.Array[i], b.Array[i])
		}
		return NewArrayValue(result), nil
	}
	return nil, NewEvalError("MAX: incompatible argument types")
}

// fnMIN implements Min: MIN(a, b)
func fnMIN(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("MIN requires 2 arguments")
	}
	a, b := args[0], args[1]
	if a.IsString || b.IsString || a.IsDraw || b.IsDraw {
		return nil, NewEvalError("MIN arguments must be numeric")
	}
	if !a.IsArray && !b.IsArray {
		return NewSingleValue(math.Min(a.Single, b.Single)), nil
	}
	if a.IsArray && b.IsArray {
		if len(a.Array) != len(b.Array) {
			return nil, NewEvalError("MIN: array length mismatch")
		}
		result := make([]float64, len(a.Array))
		for i := range a.Array {
			result[i] = math.Min(a.Array[i], b.Array[i])
		}
		return NewArrayValue(result), nil
	}
	return nil, NewEvalError("MIN: incompatible argument types")
}

func fnABS(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "ABS", math.Abs)
}

func fnSQRT(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "SQRT", math.Sqrt)
}

func fnPOW(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericBinaryFunc(args, "POW", math.Pow)
}

func fnEXP(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "EXP", math.Exp)
}

func fnLN(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "LN", math.Log)
}

func fnLOG(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "LOG", math.Log10)
}

func fnMOD(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericBinaryFunc(args, "MOD", math.Mod)
}

func fnCEILING(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "CEILING", math.Ceil)
}

func fnFLOOR(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "FLOOR", math.Floor)
}

func fnINTPART(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "INTPART", math.Trunc)
}

func fnFRACPART(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "FRACPART", func(v float64) float64 {
		return v - math.Trunc(v)
	})
}

func fnROUND(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "ROUND", math.Round)
}

func fnROUND2(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("ROUND2 requires 2 arguments")
	}
	if args[1].IsArray {
		return nil, NewEvalError("ROUND2 second argument must be a number")
	}
	scale := math.Pow(10, args[1].Single)
	return numericUnaryFunc([]*Value{args[0]}, "ROUND2", func(v float64) float64 {
		return math.Round(v*scale) / scale
	})
}

func fnSIGN(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "SIGN", func(v float64) float64 {
		if v > 0 {
			return 1
		}
		if v < 0 {
			return -1
		}
		return 0
	})
}

func fnSIN(args []*Value, _ []indicator.Bar) (*Value, error) { return numericUnaryFunc(args, "SIN", math.Sin) }
func fnCOS(args []*Value, _ []indicator.Bar) (*Value, error) { return numericUnaryFunc(args, "COS", math.Cos) }
func fnTAN(args []*Value, _ []indicator.Bar) (*Value, error) { return numericUnaryFunc(args, "TAN", math.Tan) }
func fnASIN(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "ASIN", math.Asin)
}
func fnACOS(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "ACOS", math.Acos)
}
func fnATAN(args []*Value, _ []indicator.Bar) (*Value, error) {
	return numericUnaryFunc(args, "ATAN", math.Atan)
}

// fnREF implements Reference: REF(data, n)
func fnREF(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("REF requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("REF first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("REF second argument must be a number")
	}
	n := int(period.Single)
	if n < 0 {
		return nil, NewEvalError("REF period must be non-negative")
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n && i < len(result); i++ {
		result[i] = math.NaN()
	}
	for i := n; i < len(data.Array); i++ {
		result[i] = data.Array[i-n]
	}
	return NewArrayValue(result), nil
}

func fnREFV(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("REFV requires 2 arguments")
	}
	return fnREF(args, bars)
}

func fnREFX(args []*Value, _ []indicator.Bar) (*Value, error) {
	return futureReference(args, "REFX")
}

func fnREFXV(args []*Value, _ []indicator.Bar) (*Value, error) {
	return futureReference(args, "REFXV")
}

// fnHHV implements Highest High Value: HHV(data, period)
func fnHHV(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("HHV requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("HHV first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("HHV second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("HHV period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(data.Array); i++ {
		maxVal := data.Array[i]
		for j := 1; j < n; j++ {
			if data.Array[i-j] > maxVal {
				maxVal = data.Array[i-j]
			}
		}
		result[i] = maxVal
	}
	return NewArrayValue(result), nil
}

// fnLLV implements Lowest Low Value: LLV(data, period)
func fnLLV(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("LLV requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("LLV first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("LLV second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("LLV period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(data.Array); i++ {
		minVal := data.Array[i]
		for j := 1; j < n; j++ {
			if data.Array[i-j] < minVal {
				minVal = data.Array[i-j]
			}
		}
		result[i] = minVal
	}
	return NewArrayValue(result), nil
}

// fnIF implements conditional: IF(condition, trueValue, falseValue)
func fnIF(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("IF requires 3 arguments")
	}
	cond, trueVal, falseVal := args[0], args[1], args[2]

	if !cond.IsArray {
		if isTruthy(cond.Single) {
			return trueVal, nil
		}
		return falseVal, nil
	}
	if trueVal.IsString || falseVal.IsString || trueVal.IsDraw || falseVal.IsDraw {
		return nil, NewEvalError("IF: true/false values must be numeric")
	}
	if trueVal.IsArray && len(cond.Array) != len(trueVal.Array) {
		return nil, NewEvalError("IF: true value array length mismatch")
	}
	if falseVal.IsArray && len(cond.Array) != len(falseVal.Array) {
		return nil, NewEvalError("IF: false value array length mismatch")
	}

	result := make([]float64, len(cond.Array))
	for i := range cond.Array {
		if isTruthy(cond.Array[i]) {
			result[i] = scalarOrArrayAt(trueVal, i)
		} else {
			result[i] = scalarOrArrayAt(falseVal, i)
		}
	}
	return NewArrayValue(result), nil
}

// fnCROSS implements cross detection: CROSS(a, b)
func fnCROSS(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("CROSS requires 2 arguments")
	}
	a, b := args[0], args[1]
	if !a.IsArray || !b.IsArray {
		return nil, NewEvalError("CROSS requires array arguments")
	}
	if len(a.Array) != len(b.Array) {
		return nil, NewEvalError("CROSS: array length mismatch")
	}

	result := make([]float64, len(a.Array))
	if len(result) > 0 {
		result[0] = 0
	}
	for i := 1; i < len(a.Array); i++ {
		if a.Array[i-1] <= b.Array[i-1] && a.Array[i] > b.Array[i] {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return NewArrayValue(result), nil
}

// fnSTD implements Standard Deviation: STD(data, period) (population)
func fnSTD(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingStatsFunc(args, "STD", func(values []float64) float64 {
		return math.Sqrt(variance(values))
	})
}

func fnSTDP(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("STDP requires 2 arguments")
	}
	return fnSTD(args, bars)
}

func fnSTDDEV(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingStatsFunc(args, "STDDEV", func(values []float64) float64 {
		if len(values) < 2 {
			return 0
		}
		mean := average(values)
		sum := 0.0
		for _, v := range values {
			diff := v - mean
			sum += diff * diff
		}
		return math.Sqrt(sum / float64(len(values)-1))
	})
}

// fnVAR implements Variance: VAR(data, period) (population)
func fnVAR(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingStatsFunc(args, "VAR", variance)
}

func fnVARP(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("VARP requires 2 arguments")
	}
	return fnVAR(args, bars)
}

func fnDEVSQ(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingStatsFunc(args, "DEVSQ", func(values []float64) float64 {
		mean := average(values)
		sum := 0.0
		for _, v := range values {
			diff := v - mean
			sum += diff * diff
		}
		return sum
	})
}

func fnFORCAST(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingRegressionFunc(args, "FORCAST", func(slope, intercept float64, n int) float64 {
		return intercept + slope*float64(n-1)
	})
}

func fnSLOPE(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingRegressionFunc(args, "SLOPE", func(slope, _ float64, _ int) float64 {
		return slope
	})
}

func fnCOVAR(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingPairStatsFunc(args, "COVAR", covariance)
}

func fnRELATE(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingPairStatsFunc(args, "RELATE", func(a, b []float64) float64 {
		cov := covariance(a, b)
		varA := variance(a)
		varB := variance(b)
		if varA == 0 || varB == 0 {
			return math.NaN()
		}
		return cov / math.Sqrt(varA*varB)
	})
}

func fnBETA(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingPairStatsFunc(args, "BETA", func(a, b []float64) float64 {
		varB := variance(b)
		if varB == 0 {
			return math.NaN()
		}
		return covariance(a, b) / varB
	})
}

// fnSMA implements SMA(data, period) as MA alias, and SMA(data, period, weight)
// as the TDX recursive smoothing formula.
func fnSMA(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) == 2 {
		return fnMA(args, bars)
	}
	if len(args) != 3 {
		return nil, NewEvalError("SMA requires 2 or 3 arguments")
	}
	source, period, weight := args[0], args[1], args[2]
	if !source.IsArray {
		return nil, NewEvalError("SMA first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("SMA second argument must be a number")
	}
	if weight.IsArray {
		return nil, NewEvalError("SMA third argument must be a number")
	}
	n := int(period.Single)
	m := int(weight.Single)
	if n <= 0 {
		return nil, NewEvalError("SMA period must be positive")
	}
	if m <= 0 || m > n {
		return nil, NewEvalError("SMA weight must be between 1 and period")
	}
	if len(source.Array) == 0 {
		return NewArrayValue([]float64{}), nil
	}

	result := make([]float64, len(source.Array))
	result[0] = source.Array[0]
	for i := 1; i < len(source.Array); i++ {
		result[i] = (float64(m)*source.Array[i] + float64(n-m)*result[i-1]) / float64(n)
	}
	return NewArrayValue(result), nil
}

// fnWMA implements Weighted Moving Average: WMA(data, period)
func fnWMA(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("WMA requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("WMA first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("WMA second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("WMA period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	weightSum := float64(n * (n + 1) / 2)
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(data.Array); i++ {
		weightedSum := 0.0
		for j := 0; j < n; j++ {
			weight := float64(n - j)
			weightedSum += data.Array[i-j] * weight
		}
		result[i] = weightedSum / weightSum
	}
	return NewArrayValue(result), nil
}

// fnDMA implements DMA(data, alpha) - dynamic moving average.
func fnDMA(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("DMA requires 2 arguments")
	}
	data, alpha := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("DMA first argument must be an array")
	}
	if alpha.IsArray && len(alpha.Array) != len(data.Array) {
		return nil, NewEvalError("DMA: array length mismatch")
	}
	if len(data.Array) == 0 {
		return NewArrayValue([]float64{}), nil
	}

	result := make([]float64, len(data.Array))
	result[0] = data.Array[0]
	for i := 1; i < len(data.Array); i++ {
		a := alpha.Single
		if alpha.IsArray {
			a = alpha.Array[i]
		}
		result[i] = a*data.Array[i] + (1-a)*result[i-1]
	}
	return NewArrayValue(result), nil
}

func fnCONST(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("CONST requires 1 argument")
	}
	value := args[0]
	if !value.IsArray {
		return NewSingleValue(value.Single), nil
	}
	if len(value.Array) == 0 {
		return NewArrayValue([]float64{}), nil
	}
	lastValue := value.Array[len(value.Array)-1]
	result := make([]float64, len(value.Array))
	for i := range result {
		result[i] = lastValue
	}
	return NewArrayValue(result), nil
}

func fnVALUEWHEN(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("VALUEWHEN requires 2 arguments")
	}
	condition, value := args[0], args[1]
	if !condition.IsArray {
		return nil, NewEvalError("VALUEWHEN first argument must be an array")
	}
	if value.IsArray && len(value.Array) != len(condition.Array) {
		return nil, NewEvalError("VALUEWHEN: array length mismatch")
	}

	result := make([]float64, len(condition.Array))
	lastValue := math.NaN()
	hasValue := false
	for i, cond := range condition.Array {
		if isTruthy(cond) {
			if value.IsArray {
				lastValue = value.Array[i]
			} else {
				lastValue = value.Single
			}
			hasValue = true
		}
		if hasValue {
			result[i] = lastValue
		} else {
			result[i] = math.NaN()
		}
	}
	return NewArrayValue(result), nil
}

func fnDRAWNULL(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 0 {
		return nil, NewEvalError("DRAWNULL requires 0 arguments")
	}
	return NewSingleValue(math.NaN()), nil
}

func fnLONGCROSS(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("LONGCROSS requires 3 arguments")
	}
	a, b, period := args[0], args[1], args[2]
	if !a.IsArray || !b.IsArray {
		return nil, NewEvalError("LONGCROSS first two arguments must be arrays")
	}
	if period.IsArray {
		return nil, NewEvalError("LONGCROSS third argument must be a number")
	}
	if len(a.Array) != len(b.Array) {
		return nil, NewEvalError("LONGCROSS: array length mismatch")
	}
	n := int(period.Single)
	if n <= 0 {
		return nil, NewEvalError("LONGCROSS period must be positive")
	}

	result := make([]float64, len(a.Array))
	for i := 1; i < len(a.Array); i++ {
		if !(a.Array[i-1] <= b.Array[i-1] && a.Array[i] > b.Array[i]) || i < n {
			continue
		}
		ok := true
		for j := 1; j <= n; j++ {
			if a.Array[i-j] >= b.Array[i-j] {
				ok = false
				break
			}
		}
		if ok {
			result[i] = 1
		}
	}
	return NewArrayValue(result), nil
}

func fnUPNDAY(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("UPNDAY requires 2 arguments")
	}
	return compareConsecutive(args, "UPNDAY", func(curr, prev float64) bool { return curr > prev })
}

func fnDOWNNDAY(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("DOWNNDAY requires 2 arguments")
	}
	return compareConsecutive(args, "DOWNNDAY", func(curr, prev float64) bool { return curr < prev })
}

func fnNDAY(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("NDAY requires 3 arguments")
	}
	a, b, period := args[0], args[1], args[2]
	if !a.IsArray || !b.IsArray {
		return nil, NewEvalError("NDAY first two arguments must be arrays")
	}
	if period.IsArray {
		return nil, NewEvalError("NDAY third argument must be a number")
	}
	if len(a.Array) != len(b.Array) {
		return nil, NewEvalError("NDAY: array length mismatch")
	}
	n := int(period.Single)
	if n <= 0 {
		return nil, NewEvalError("NDAY period must be positive")
	}

	result := make([]float64, len(a.Array))
	for i := n - 1; i < len(a.Array); i++ {
		ok := true
		for j := 0; j < n; j++ {
			if !(a.Array[i-j] > b.Array[i-j]) {
				ok = false
				break
			}
		}
		if ok {
			result[i] = 1
		}
	}
	return NewArrayValue(result), nil
}

func fnLAST(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("LAST requires 3 arguments")
	}
	condition, from, to := args[0], args[1], args[2]
	if !condition.IsArray {
		return nil, NewEvalError("LAST first argument must be an array")
	}
	if from.IsArray || to.IsArray {
		return nil, NewEvalError("LAST second and third arguments must be numbers")
	}
	fromN := int(from.Single)
	toN := int(to.Single)
	if fromN < toN || toN < 0 {
		return nil, NewEvalError("LAST requires from >= to >= 0")
	}

	result := make([]float64, len(condition.Array))
	for i := range condition.Array {
		if i < fromN {
			continue
		}
		ok := true
		for j := toN; j <= fromN; j++ {
			if !isTruthy(condition.Array[i-j]) {
				ok = false
				break
			}
		}
		if ok {
			result[i] = 1
		}
	}
	return NewArrayValue(result), nil
}

func fnEXISTR(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("EXISTR requires 3 arguments")
	}
	condition, from, to := args[0], args[1], args[2]
	if !condition.IsArray {
		return nil, NewEvalError("EXISTR first argument must be an array")
	}
	if from.IsArray || to.IsArray {
		return nil, NewEvalError("EXISTR second and third arguments must be numbers")
	}
	fromN := int(from.Single)
	toN := int(to.Single)
	if fromN < toN || toN < 0 {
		return nil, NewEvalError("EXISTR requires from >= to >= 0")
	}

	result := make([]float64, len(condition.Array))
	for i := range condition.Array {
		if i < fromN {
			continue
		}
		for j := toN; j <= fromN; j++ {
			if isTruthy(condition.Array[i-j]) {
				result[i] = 1
				break
			}
		}
	}
	return NewArrayValue(result), nil
}

func fnCOUNT(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("COUNT requires 2 arguments")
	}
	condition, period := args[0], args[1]
	if !condition.IsArray {
		return nil, NewEvalError("COUNT first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("COUNT second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(condition.Array) {
		return nil, NewEvalError(fmt.Sprintf("COUNT period must be between 1 and %d", len(condition.Array)))
	}

	result := make([]float64, len(condition.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(condition.Array); i++ {
		count := 0.0
		for j := 0; j < n; j++ {
			if isTruthy(condition.Array[i-j]) {
				count++
			}
		}
		result[i] = count
	}
	return NewArrayValue(result), nil
}

func fnEVERY(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("EVERY requires 2 arguments")
	}
	condition, period := args[0], args[1]
	if !condition.IsArray {
		return nil, NewEvalError("EVERY first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("EVERY second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(condition.Array) {
		return nil, NewEvalError(fmt.Sprintf("EVERY period must be between 1 and %d", len(condition.Array)))
	}

	result := make([]float64, len(condition.Array))
	for i := 0; i < n-1; i++ {
		result[i] = 0
	}
	for i := n - 1; i < len(condition.Array); i++ {
		everyCond := true
		for j := 0; j < n; j++ {
			if !isTruthy(condition.Array[i-j]) {
				everyCond = false
				break
			}
		}
		if everyCond {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return NewArrayValue(result), nil
}

func fnEXIST(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("EXIST requires 2 arguments")
	}
	condition, period := args[0], args[1]
	if !condition.IsArray {
		return nil, NewEvalError("EXIST first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("EXIST second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(condition.Array) {
		return nil, NewEvalError(fmt.Sprintf("EXIST period must be between 1 and %d", len(condition.Array)))
	}

	result := make([]float64, len(condition.Array))
	for i := 0; i < n-1; i++ {
		result[i] = 0
	}
	for i := n - 1; i < len(condition.Array); i++ {
		exists := false
		for j := 0; j < n; j++ {
			if isTruthy(condition.Array[i-j]) {
				exists = true
				break
			}
		}
		if exists {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return NewArrayValue(result), nil
}

func fnBARSLAST(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("BARSLAST requires 1 argument")
	}
	condition := args[0]
	if !condition.IsArray {
		return nil, NewEvalError("BARSLAST argument must be an array")
	}

	result := make([]float64, len(condition.Array))
	lastTrueIndex := -1
	for i := 0; i < len(condition.Array); i++ {
		if isTruthy(condition.Array[i]) {
			lastTrueIndex = i
			result[i] = 0
		} else if lastTrueIndex >= 0 {
			result[i] = float64(i - lastTrueIndex)
		} else {
			result[i] = math.NaN()
		}
	}
	return NewArrayValue(result), nil
}

func fnHHVBARS(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("HHVBARS requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("HHVBARS first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("HHVBARS second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("HHVBARS period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(data.Array); i++ {
		maxValue := data.Array[i]
		bars := 0
		for j := 1; j < n; j++ {
			if data.Array[i-j] > maxValue {
				maxValue = data.Array[i-j]
				bars = j
			}
		}
		result[i] = float64(bars)
	}
	return NewArrayValue(result), nil
}

func fnLLVBARS(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("LLVBARS requires 2 arguments")
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("LLVBARS first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("LLVBARS second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("LLVBARS period must be between 1 and %d", len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	for i := n - 1; i < len(data.Array); i++ {
		minValue := data.Array[i]
		bars := 0
		for j := 1; j < n; j++ {
			if data.Array[i-j] < minValue {
				minValue = data.Array[i-j]
				bars = j
			}
		}
		result[i] = float64(bars)
	}
	return NewArrayValue(result), nil
}

func fnBARSCOUNT(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("BARSCOUNT requires 1 argument")
	}
	data := args[0]
	if !data.IsArray {
		return nil, NewEvalError("BARSCOUNT argument must be an array")
	}

	result := make([]float64, len(data.Array))
	count := 0
	for i, value := range data.Array {
		if !math.IsNaN(value) {
			count++
		}
		result[i] = float64(count)
	}
	return NewArrayValue(result), nil
}

func fnBARSSINCE(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("BARSSINCE requires 1 argument")
	}
	condition := args[0]
	if !condition.IsArray {
		return nil, NewEvalError("BARSSINCE argument must be an array")
	}

	result := make([]float64, len(condition.Array))
	firstTrueIndex := -1
	for i, value := range condition.Array {
		if firstTrueIndex < 0 {
			if isTruthy(value) {
				firstTrueIndex = i
				result[i] = 0
			} else {
				result[i] = math.NaN()
			}
			continue
		}
		result[i] = float64(i - firstTrueIndex)
	}
	return NewArrayValue(result), nil
}

func fnBARSLASTCOUNT(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("BARSLASTCOUNT requires 1 argument")
	}
	condition := args[0]
	if !condition.IsArray {
		return nil, NewEvalError("BARSLASTCOUNT argument must be an array")
	}

	result := make([]float64, len(condition.Array))
	count := 0
	for i, value := range condition.Array {
		if isTruthy(value) {
			count++
		} else {
			count = 0
		}
		result[i] = float64(count)
	}
	return NewArrayValue(result), nil
}

func fnAVEDEV(args []*Value, _ []indicator.Bar) (*Value, error) {
	return rollingStatsFunc(args, "AVEDEV", func(values []float64) float64 {
		mean := average(values)
		devSum := 0.0
		for _, v := range values {
			devSum += math.Abs(v - mean)
		}
		return devSum / float64(len(values))
	})
}

func fnFILTER(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("FILTER requires 2 arguments")
	}
	condition, period := args[0], args[1]
	if !condition.IsArray {
		return nil, NewEvalError("FILTER first argument must be an array")
	}
	if period.IsArray {
		return nil, NewEvalError("FILTER second argument must be a number")
	}
	n := int(period.Single)
	if n <= 0 {
		return nil, NewEvalError("FILTER period must be positive")
	}

	result := make([]float64, len(condition.Array))
	lastSignal := -n - 1
	for i := 0; i < len(condition.Array); i++ {
		if isTruthy(condition.Array[i]) && (i-lastSignal) >= n {
			result[i] = 1
			lastSignal = i
		} else {
			result[i] = 0
		}
	}
	return NewArrayValue(result), nil
}

func fnBETWEEN(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("BETWEEN requires 3 arguments")
	}
	value, lower, upper := args[0], args[1], args[2]

	if !value.IsArray && !lower.IsArray && !upper.IsArray {
		if value.Single >= lower.Single && value.Single <= upper.Single {
			return NewSingleValue(1), nil
		}
		return NewSingleValue(0), nil
	}
	if !value.IsArray {
		return nil, NewEvalError("BETWEEN: value must be array when using array bounds")
	}

	result := make([]float64, len(value.Array))
	for i := range value.Array {
		lowerBound := lower.Single
		upperBound := upper.Single
		if lower.IsArray {
			lowerBound = lower.Array[i]
		}
		if upper.IsArray {
			upperBound = upper.Array[i]
		}
		if value.Array[i] >= lowerBound && value.Array[i] <= upperBound {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return NewArrayValue(result), nil
}

func fnRANGE(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("RANGE requires 3 arguments")
	}
	return fnBETWEEN(args, bars)
}

func fnNOT(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError("NOT requires 1 argument")
	}
	value := args[0]
	if !value.IsArray {
		if isTruthy(value.Single) {
			return NewSingleValue(0), nil
		}
		return NewSingleValue(1), nil
	}

	result := make([]float64, len(value.Array))
	for i, v := range value.Array {
		if isTruthy(v) {
			result[i] = 0
		} else {
			result[i] = 1
		}
	}
	return NewArrayValue(result), nil
}

func fnIFN(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("IFN requires 3 arguments")
	}
	return fnIF([]*Value{args[0], args[2], args[1]}, bars)
}

func fnIFF(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("IFF requires 3 arguments")
	}
	return fnIF(args, bars)
}

func fnCURRBARSCOUNT(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 0 {
		return nil, NewEvalError("CURRBARSCOUNT requires 0 arguments")
	}
	n := len(bars)
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		result[i] = float64(n - i)
	}
	return NewArrayValue(result), nil
}

func fnTOTALBARSCOUNT(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 0 {
		return nil, NewEvalError("TOTALBARSCOUNT requires 0 arguments")
	}
	n := len(bars)
	result := make([]float64, n)
	for i := range result {
		result[i] = float64(n)
	}
	return NewArrayValue(result), nil
}

func fnISLASTBAR(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 0 {
		return nil, NewEvalError("ISLASTBAR requires 0 arguments")
	}
	result := make([]float64, len(bars))
	if len(result) > 0 {
		result[len(result)-1] = 1
	}
	return NewArrayValue(result), nil
}

func fnBARSTATUS(args []*Value, bars []indicator.Bar) (*Value, error) {
	if len(args) != 0 {
		return nil, NewEvalError("BARSTATUS requires 0 arguments")
	}
	result := make([]float64, len(bars))
	if len(result) > 0 {
		result[0] = 1
		if len(result) > 1 {
			for i := 1; i < len(result)-1; i++ {
				result[i] = 2
			}
			result[len(result)-1] = 3
		}
	}
	return NewArrayValue(result), nil
}

func fnSUMBARS(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("SUMBARS requires 2 arguments")
	}
	data, target := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError("SUMBARS first argument must be an array")
	}
	if target.IsArray {
		return nil, NewEvalError("SUMBARS second argument must be a number")
	}

	result := make([]float64, len(data.Array))
	for i := range data.Array {
		sum := 0.0
		found := false
		for j := i; j >= 0; j-- {
			sum += data.Array[j]
			if sum >= target.Single {
				result[i] = float64(i - j + 1)
				found = true
				break
			}
		}
		if !found {
			result[i] = math.NaN()
		}
	}
	return NewArrayValue(result), nil
}

// Drawing functions

func fnDRAWTEXT(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("DRAWTEXT requires 3 arguments")
	}
	return buildPointDrawings("DRAWTEXT", args[0], args[1], nil, args[2], "price")
}

func fnDRAWICON(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("DRAWICON requires 3 arguments")
	}
	return buildPointDrawings("DRAWICON", args[0], args[1], args[2], nil, "price")
}

func fnDRAWNUMBER(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError("DRAWNUMBER requires 3 arguments")
	}
	return buildPointDrawings("DRAWNUMBER", args[0], args[1], args[2], nil, "price")
}

func fnSTICKLINE(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 5 {
		return nil, NewEvalError("STICKLINE requires 5 arguments")
	}
	condition, price1, price2, width, empty := args[0], args[1], args[2], args[3], args[4]
	if !condition.IsArray {
		return nil, NewEvalError("STICKLINE first argument must be an array")
	}
	if err := validateDrawingNumericArgs("STICKLINE", len(condition.Array), price1, price2, width, empty); err != nil {
		return nil, err
	}

	drawings := make([]DrawingEvent, 0, truthyCount(condition.Array))
	for i, cond := range condition.Array {
		if !isTruthy(cond) {
			continue
		}
		event := DrawingEvent{
			Function: "STICKLINE",
			BarIndex: i,
			Values: map[string]float64{
				"price1": scalarOrArrayAt(price1, i),
				"price2": scalarOrArrayAt(price2, i),
				"width":  scalarOrArrayAt(width, i),
				"empty":  scalarOrArrayAt(empty, i),
			},
		}
		drawings = append(drawings, event)
	}
	return NewDrawingValue(drawings), nil
}

func fnDRAWLINE(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 5 {
		return nil, NewEvalError("DRAWLINE requires 5 arguments")
	}
	cond1, price1, cond2, price2, expand := args[0], args[1], args[2], args[3], args[4]
	if !cond1.IsArray {
		return nil, NewEvalError("DRAWLINE first argument must be an array")
	}
	if !cond2.IsArray {
		return nil, NewEvalError("DRAWLINE third argument must be an array")
	}
	if len(cond1.Array) != len(cond2.Array) {
		return nil, NewEvalError("DRAWLINE: condition array length mismatch")
	}
	if err := validateDrawingNumericArgs("DRAWLINE", len(cond1.Array), price1, price2, expand); err != nil {
		return nil, err
	}

	drawings := make([]DrawingEvent, 0)
	startIndex := -1
	startPrice := math.NaN()
	for i := range cond1.Array {
		if isTruthy(cond1.Array[i]) {
			startIndex = i
			startPrice = scalarOrArrayAt(price1, i)
		}
		if startIndex < 0 || !isTruthy(cond2.Array[i]) || i < startIndex {
			continue
		}
		endPrice := scalarOrArrayAt(price2, i)
		drawings = append(drawings, DrawingEvent{
			Function: "DRAWLINE",
			BarIndex: startIndex,
			Values: map[string]float64{
				"startBar":   float64(startIndex),
				"startPrice": startPrice,
				"endBar":     float64(i),
				"endPrice":   endPrice,
				"expand":     scalarOrArrayAt(expand, i),
			},
		})
		startIndex = -1
		startPrice = math.NaN()
	}
	return NewDrawingValue(drawings), nil
}

func fnPOLYLINE(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError("POLYLINE requires 2 arguments")
	}
	return buildPointDrawings("POLYLINE", args[0], args[1], nil, nil, "price")
}

func fnDRAWBAND(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 4 {
		return nil, NewEvalError("DRAWBAND requires 4 arguments")
	}
	upper, upperColor, lower, lowerColor := args[0], args[1], args[2], args[3]
	length, err := drawingValueLength("DRAWBAND", upper, upperColor, lower, lowerColor)
	if err != nil {
		return nil, err
	}

	drawings := make([]DrawingEvent, 0, length)
	for i := 0; i < length; i++ {
		drawings = append(drawings, DrawingEvent{
			Function: "DRAWBAND",
			BarIndex: i,
			Values: map[string]float64{
				"upper":      scalarOrArrayAt(upper, i),
				"upperColor": scalarOrArrayAt(upperColor, i),
				"lower":      scalarOrArrayAt(lower, i),
				"lowerColor": scalarOrArrayAt(lowerColor, i),
			},
		})
	}
	return NewDrawingValue(drawings), nil
}

func fnDRAWKLINE(args []*Value, _ []indicator.Bar) (*Value, error) {
	if len(args) != 4 {
		return nil, NewEvalError("DRAWKLINE requires 4 arguments")
	}
	high, open, low, close := args[0], args[1], args[2], args[3]
	length, err := drawingValueLength("DRAWKLINE", high, open, low, close)
	if err != nil {
		return nil, err
	}

	drawings := make([]DrawingEvent, 0, length)
	for i := 0; i < length; i++ {
		drawings = append(drawings, DrawingEvent{
			Function: "DRAWKLINE",
			BarIndex: i,
			Values: map[string]float64{
				"high":  scalarOrArrayAt(high, i),
				"open":  scalarOrArrayAt(open, i),
				"low":   scalarOrArrayAt(low, i),
				"close": scalarOrArrayAt(close, i),
			},
		})
	}
	return NewDrawingValue(drawings), nil
}

func buildPointDrawings(function string, condition, price, numeric *Value, text *Value, priceKey string) (*Value, error) {
	if !condition.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s first argument must be an array", function))
	}
	if err := validateDrawingNumericArgs(function, len(condition.Array), price); err != nil {
		return nil, err
	}
	if numeric != nil {
		if err := validateDrawingNumericArgs(function, len(condition.Array), numeric); err != nil {
			return nil, err
		}
	}

	drawings := make([]DrawingEvent, 0, truthyCount(condition.Array))
	for i, cond := range condition.Array {
		if !isTruthy(cond) {
			continue
		}
		event := DrawingEvent{
			Function: function,
			BarIndex: i,
			Values:   make(map[string]float64, 2),
		}
		event.Values[priceKey] = scalarOrArrayAt(price, i)
		if numeric != nil {
			event.Values["value"] = scalarOrArrayAt(numeric, i)
		}
		if text != nil {
			event.Text = textValueAt(text, i)
		}
		drawings = append(drawings, event)
	}
	return NewDrawingValue(drawings), nil
}

func validateDrawingNumericArgs(function string, length int, values ...*Value) error {
	for _, value := range values {
		if value == nil || value.IsString || value.IsDraw {
			return NewEvalError(fmt.Sprintf("%s arguments must be numeric", function))
		}
		if value.IsArray && len(value.Array) != length {
			return NewEvalError(fmt.Sprintf("%s: array length mismatch", function))
		}
	}
	return nil
}

func drawingValueLength(function string, values ...*Value) (int, error) {
	length := 0
	for _, value := range values {
		if value == nil || value.IsString || value.IsDraw {
			return 0, NewEvalError(fmt.Sprintf("%s arguments must be numeric", function))
		}
		if !value.IsArray {
			continue
		}
		if length == 0 {
			length = len(value.Array)
			continue
		}
		if len(value.Array) != length {
			return 0, NewEvalError(fmt.Sprintf("%s: array length mismatch", function))
		}
	}
	if length == 0 {
		length = 1
	}
	return length, nil
}

func scalarOrArrayAt(value *Value, index int) float64 {
	if value == nil {
		return math.NaN()
	}
	if value.IsArray {
		if index >= len(value.Array) {
			return math.NaN()
		}
		return value.Array[index]
	}
	return value.Single
}

func textValueAt(value *Value, index int) string {
	if value == nil {
		return ""
	}
	if value.IsString {
		return value.Text
	}
	if value.IsArray {
		return fmt.Sprintf("%g", scalarOrArrayAt(value, index))
	}
	return fmt.Sprintf("%g", value.Single)
}

func numericUnaryFunc(args []*Value, name string, fn func(float64) float64) (*Value, error) {
	if len(args) != 1 {
		return nil, NewEvalError(fmt.Sprintf("%s requires 1 argument", name))
	}
	value := args[0]
	if value.IsString || value.IsDraw {
		return nil, NewEvalError(fmt.Sprintf("%s argument must be numeric", name))
	}
	if !value.IsArray {
		return NewSingleValue(fn(value.Single)), nil
	}

	result := make([]float64, len(value.Array))
	for i, v := range value.Array {
		result[i] = fn(v)
	}
	return NewArrayValue(result), nil
}

func numericBinaryFunc(args []*Value, name string, fn func(float64, float64) float64) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError(fmt.Sprintf("%s requires 2 arguments", name))
	}
	a, b := args[0], args[1]
	if a.IsString || b.IsString || a.IsDraw || b.IsDraw {
		return nil, NewEvalError(fmt.Sprintf("%s arguments must be numeric", name))
	}
	if !a.IsArray && !b.IsArray {
		return NewSingleValue(fn(a.Single, b.Single)), nil
	}

	length := valueLength(a)
	if length == 0 {
		length = valueLength(b)
	}
	if a.IsArray && b.IsArray && len(a.Array) != len(b.Array) {
		return nil, NewEvalError(fmt.Sprintf("%s: array length mismatch", name))
	}

	result := make([]float64, length)
	for i := range result {
		result[i] = fn(scalarOrArrayAt(a, i), scalarOrArrayAt(b, i))
	}
	return NewArrayValue(result), nil
}

func rollingStatsFunc(args []*Value, name string, fn func([]float64) float64) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError(fmt.Sprintf("%s requires 2 arguments", name))
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s first argument must be an array", name))
	}
	if period.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s second argument must be a number", name))
	}
	n := int(period.Single)
	if n <= 0 || n > len(data.Array) {
		return nil, NewEvalError(fmt.Sprintf("%s period must be between 1 and %d", name, len(data.Array)))
	}

	result := make([]float64, len(data.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	window := make([]float64, n)
	for i := n - 1; i < len(data.Array); i++ {
		copy(window, data.Array[i-n+1:i+1])
		result[i] = fn(window)
	}
	return NewArrayValue(result), nil
}

func rollingRegressionFunc(args []*Value, name string, fn func(float64, float64, int) float64) (*Value, error) {
	return rollingStatsFunc(args, name, func(values []float64) float64 {
		slope, intercept := linearRegression(values)
		return fn(slope, intercept, len(values))
	})
}

func rollingPairStatsFunc(args []*Value, name string, fn func([]float64, []float64) float64) (*Value, error) {
	if len(args) != 3 {
		return nil, NewEvalError(fmt.Sprintf("%s requires 3 arguments", name))
	}
	a, b, period := args[0], args[1], args[2]
	if !a.IsArray || !b.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s first two arguments must be arrays", name))
	}
	if len(a.Array) != len(b.Array) {
		return nil, NewEvalError(fmt.Sprintf("%s: array length mismatch", name))
	}
	if period.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s third argument must be a number", name))
	}
	n := int(period.Single)
	if n <= 0 || n > len(a.Array) {
		return nil, NewEvalError(fmt.Sprintf("%s period must be between 1 and %d", name, len(a.Array)))
	}

	result := make([]float64, len(a.Array))
	for i := 0; i < n-1; i++ {
		result[i] = math.NaN()
	}
	windowA := make([]float64, n)
	windowB := make([]float64, n)
	for i := n - 1; i < len(a.Array); i++ {
		copy(windowA, a.Array[i-n+1:i+1])
		copy(windowB, b.Array[i-n+1:i+1])
		result[i] = fn(windowA, windowB)
	}
	return NewArrayValue(result), nil
}

func futureReference(args []*Value, name string) (*Value, error) {
	if len(args) != 2 {
		return nil, NewEvalError(fmt.Sprintf("%s requires 2 arguments", name))
	}
	data, period := args[0], args[1]
	if !data.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s first argument must be an array", name))
	}
	if period.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s second argument must be a number", name))
	}
	n := int(period.Single)
	if n < 0 {
		return nil, NewEvalError(fmt.Sprintf("%s period must be non-negative", name))
	}

	result := make([]float64, len(data.Array))
	for i := range data.Array {
		futureIndex := i + n
		if futureIndex >= len(data.Array) {
			result[i] = math.NaN()
			continue
		}
		result[i] = data.Array[futureIndex]
	}
	return NewArrayValue(result), nil
}

func average(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func variance(values []float64) float64 {
	mean := average(values)
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

func covariance(a, b []float64) float64 {
	meanA := average(a)
	meanB := average(b)
	sum := 0.0
	for i := range a {
		sum += (a[i] - meanA) * (b[i] - meanB)
	}
	return sum / float64(len(a))
}

func linearRegression(values []float64) (float64, float64) {
	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return 0, values[len(values)-1]
	}
	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n
	return slope, intercept
}

func valueLength(value *Value) int {
	if value != nil && value.IsArray {
		return len(value.Array)
	}
	return 0
}

func compareConsecutive(args []*Value, name string, cmp func(curr, prev float64) bool) (*Value, error) {
	data := args[0]
	period := args[1]
	if !data.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s first argument must be an array", name))
	}
	if period.IsArray {
		return nil, NewEvalError(fmt.Sprintf("%s second argument must be a number", name))
	}
	n := int(period.Single)
	if n <= 0 {
		return nil, NewEvalError(fmt.Sprintf("%s period must be positive", name))
	}

	result := make([]float64, len(data.Array))
	for i := n; i < len(data.Array); i++ {
		ok := true
		for j := 0; j < n; j++ {
			if !cmp(data.Array[i-j], data.Array[i-j-1]) {
				ok = false
				break
			}
		}
		if ok {
			result[i] = 1
		}
	}
	return NewArrayValue(result), nil
}

func truthyCount(values []float64) int {
	count := 0
	for _, value := range values {
		if isTruthy(value) {
			count++
		}
	}
	return count
}

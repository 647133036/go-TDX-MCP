package backtest

import (
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/tdx/go-tdx-mcp/indicator"
)

const maxGridPoints = 200

type GridPointResult struct {
	Params       map[string]float64 `json:"params"`
	TotalReturn  float64            `json:"total_return"`
	Sharpe       float64            `json:"sharpe"`
	MaxDrawdown  float64            `json:"max_drawdown"`
	TotalTrades  int                `json:"total_trades"`
	WinRate      float64            `json:"win_rate"`
	ProfitFactor float64            `json:"profit_factor"`
}

type Heatmap struct {
	XName string      `json:"x_name"`
	YName string      `json:"y_name"`
	X     []float64   `json:"x"`
	Y     []float64   `json:"y"`
	Data  [][]float64 `json:"data"`
}

type OptimizeResult struct {
	Strategy   string           `json:"strategy"`
	ParamNames []string         `json:"param_names"`
	Results    []GridPointResult `json:"results"`
	Best       *GridPointResult `json:"best"`
	Heatmap    *Heatmap         `json:"heatmap,omitempty"`
}

func applyParams(strategy Strategy, params map[string]float64) {
	if strategy == nil || len(params) == 0 {
		return
	}
	v := reflect.ValueOf(strategy)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.CanAddr() {
		return
	}
	for name, val := range params {
		field := v.FieldByName(name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			field.SetInt(int64(val))
		case reflect.Float32, reflect.Float64:
			field.SetFloat(val)
		case reflect.Bool:
			if val > 0 {
				field.SetBool(true)
			}
		}
	}
}

func NewStrategyWithParams(name string, params map[string]float64) Strategy {
	st := NewStrategy(strings.ToLower(name))
	if st == nil {
		return nil
	}
	applyParams(st, params)
	return st
}

type OptimizeRequest struct {
	Strategy   string
	Params     map[string][]float64
	Bars       []indicator.Bar
	Cash       float64
	Commission float64
	Slippage   float64
}

func RunOptimizer(req *OptimizeRequest) *OptimizeResult {
	if req == nil {
		return nil
	}
	paramNames := make([]string, 0, len(req.Params))
	valueLists := make([][]float64, 0)
	size := 1
	for name, vals := range req.Params {
		if len(vals) == 0 {
			continue
		}
		paramNames = append(paramNames, name)
		valueLists = append(valueLists, vals)
		size *= len(vals)
	}
	if size == 0 {
		return &OptimizeResult{Strategy: req.Strategy, ParamNames: paramNames}
	}
	if size > maxGridPoints {
		return nil
	}

	strategyBase := NewStrategy(strings.ToLower(req.Strategy))
	if strategyBase == nil {
		return nil
	}

	engine := NewEngine(req.Cash)
	if req.Commission > 0 {
		engine.SetCommission(req.Commission)
	}
	if req.Slippage > 0 {
		engine.SetSlippage(req.Slippage)
	}

	results := make([]GridPointResult, 0)
	prod := make([]float64, len(paramNames))
	prodIndices := make([]int, len(paramNames))
	total := size
	for iter := 0; iter < total; iter++ {
		params := make(map[string]float64)
		for i, name := range paramNames {
			idx := prodIndices[i]
			val := valueLists[i][idx]
			params[name] = val
			prod[i] = val
		}

		st := NewStrategyWithParams(req.Strategy, params)
		if st == nil {
			continue
		}
		result := engine.Run(st, req.Bars)
		if result == nil {
			continue
		}
		p := result.Performance
		gr := GridPointResult{
			Params:       params,
			TotalReturn:  p.TotalReturn,
			Sharpe:       p.Sharpe,
			MaxDrawdown:  p.MaxDrawdown,
			TotalTrades:  p.TotalTrades,
			WinRate:      p.WinRate,
			ProfitFactor: p.ProfitFactor,
		}
		results = append(results, gr)

		prodIndices[len(prodIndices)-1]++
		for idx := len(prodIndices) - 1; idx >= 0; idx-- {
			if prodIndices[idx] < len(valueLists[idx]) {
				break
			}
			prodIndices[idx] = 0
			if idx > 0 {
				prodIndices[idx-1]++
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalReturn > results[j].TotalReturn
	})

	opResult := &OptimizeResult{
		Strategy:   req.Strategy,
		ParamNames: paramNames,
		Results:    results,
	}
	if len(results) > 0 {
		opResult.Best = &results[0]
	}
	if len(paramNames) == 2 {
		opResult.Heatmap = buildHeatmap(results, paramNames, req.Params)
	}
	return opResult
}

func buildHeatmap(results []GridPointResult, paramNames []string, grid map[string][]float64) *Heatmap {
	xName, yName := paramNames[0], paramNames[1]
	xVals := sortedUnique(grid[xName])
	yVals := sortedUnique(grid[yName])
	xIdx := make(map[float64]int)
	for i, v := range xVals {
		xIdx[v] = i
	}
	yIdx := make(map[float64]int)
	for i, v := range yVals {
		yIdx[v] = i
	}
	data := make([][]float64, 0)
	for _, r := range results {
		x := r.Params[xName]
		y := r.Params[yName]
		xi, okX := xIdx[x]
		yi, okY := yIdx[y]
		if okX && okY {
			data = append(data, []float64{float64(xi), float64(yi), r.TotalReturn})
		}
	}
	return &Heatmap{XName: xName, YName: yName, X: xVals, Y: yVals, Data: data}
}

func sortedUnique(vals []float64) []float64 {
	if len(vals) == 0 {
		return nil
	}
	seen := make(map[float64]bool)
	result := make([]float64, 0)
	for _, v := range vals {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	sort.Float64s(result)
	return result
}

func cleanFloat(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

type ParamSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     float64  `json:"default"`
	Description string   `json:"description"`
}

type StrategySchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Params      []ParamSchema  `json:"params"`
}

func getStrategySchemas() []StrategySchema {
	return []StrategySchema{
		{
			Name:        "ma_cross",
			Description: "双均线交叉策略",
			Params: []ParamSchema{
				{Name: "FastPeriod", Type: "int", Default: 5, Description: "短期均线周期"},
				{Name: "SlowPeriod", Type: "int", Default: 20, Description: "长期均线周期"},
			},
		},
		{
			Name:        "macd_cross",
			Description: "MACD交叉策略",
			Params: []ParamSchema{
				{Name: "Fast", Type: "int", Default: 12, Description: "快线EMA周期"},
				{Name: "Slow", Type: "int", Default: 26, Description: "慢线EMA周期"},
				{Name: "Signal", Type: "int", Default: 9, Description: "信号线周期"},
			},
		},
		{
			Name:        "rsi_reversal",
			Description: "RSI均值回归策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 14, Description: "RSI周期"},
				{Name: "Oversold", Type: "int", Default: 30, Description: "超卖阈值"},
				{Name: "Overbought", Type: "int", Default: 70, Description: "超买阈值"},
			},
		},
		{
			Name:        "bollinger_breakout",
			Description: "布林带突破策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 20, Description: "周期"},
				{Name: "Multiplier", Type: "float", Default: 2, Description: "标准差倍数"},
			},
		},
		{
			Name:        "expma_cross",
			Description: "EXPMA交叉策略",
			Params: []ParamSchema{
				{Name: "FastPeriod", Type: "int", Default: 5, Description: "短期EXPMA周期"},
				{Name: "SlowPeriod", Type: "int", Default: 20, Description: "长期EXPMA周期"},
			},
		},
		{
			Name:        "kdj_golden",
			Description: "KDJ金叉策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 9, Description: "周期"},
				{Name: "KPeriod", Type: "int", Default: 3, Description: "K值周期"},
				{Name: "DPeriod", Type: "int", Default: 3, Description: "D值周期"},
			},
		},
		{
			Name:        "turtle_breakout",
			Description: "海龟突破策略",
			Params: []ParamSchema{
				{Name: "EntryPeriod", Type: "int", Default: 20, Description: "入场突破周期"},
				{Name: "ExitPeriod", Type: "int", Default: 10, Description: "出场周期"},
			},
		},
		{
			Name:        "bias_reversal",
			Description: "乖离率回归策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 6, Description: "周期"},
				{Name: "Oversold", Type: "float", Default: -3, Description: "超卖阈值"},
				{Name: "Overbought", Type: "float", Default: 3, Description: "超买阈值"},
			},
		},
		{
			Name:        "dmi_trend",
			Description: "DMI趋势策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 14, Description: "周期"},
				{Name: "SignalPeriod", Type: "int", Default: 6, Description: "信号周期"},
				{Name: "ADXThreshold", Type: "float", Default: 25, Description: "ADX阈值"},
			},
		},
		{
			Name:        "cci_breakout",
			Description: "CCI突破策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 14, Description: "周期"},
				{Name: "Overbought", Type: "float", Default: 100, Description: "超买阈值"},
				{Name: "Oversold", Type: "float", Default: -100, Description: "超卖阈值"},
			},
		},
		{
			Name:        "mfi_volume",
			Description: "MFI资金流量策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 14, Description: "周期"},
				{Name: "Oversold", Type: "float", Default: 20, Description: "超卖阈值"},
				{Name: "Overbought", Type: "float", Default: 80, Description: "超买阈值"},
			},
		},
		{
			Name:        "trix_cross",
			Description: "TRIX交叉策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 12, Description: "周期"},
				{Name: "SignalPeriod", Type: "int", Default: 9, Description: "信号周期"},
			},
		},
		{
			Name:        "mtm_momentum",
			Description: "MTM动量策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 6, Description: "动量周期"},
			},
		},
		{
			Name:        "obv_trend",
			Description: "OBV趋势策略",
			Params: []ParamSchema{
				{Name: "Period", Type: "int", Default: 20, Description: "OBV均线周期"},
			},
		},
	}
}

func AvailableStrategySchemas() []StrategySchema {
	return getStrategySchemas()
}


package formula

import (
	"fmt"
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

type Value struct {
	Single   float64
	Array    []float64
	Text     string
	Draws    []DrawingEvent
	IsArray  bool
	IsString bool
	IsDraw   bool
}

func NewSingleValue(v float64) *Value {
	return &Value{Single: v}
}

func NewArrayValue(arr []float64) *Value {
	return &Value{Array: arr, IsArray: true}
}

func NewStringValue(text string) *Value {
	return &Value{Text: text, IsString: true}
}

func NewDrawingValue(drawings []DrawingEvent) *Value {
	return &Value{Draws: drawings, IsDraw: true}
}

type DrawingEvent struct {
	Function string             `json:"function"`
	BarIndex int                `json:"bar_index"`
	Values   map[string]float64 `json:"values"`
	Text     string             `json:"text,omitempty"`
	Meta     map[string]string  `json:"meta,omitempty"`
}

type LineStyle struct {
	Color      string `json:"color,omitempty"`
	LineWidth  int    `json:"line_width"`
	LineStyle  string `json:"line_style,omitempty"`
	DrawMethod string `json:"draw_method,omitempty"`
	Hidden     bool   `json:"hidden"`
}

type OutputLine struct {
	Name  string     `json:"name"`
	Data  []float64  `json:"data"`
	Style *LineStyle `json:"style,omitempty"`
}

type FormulaResult struct {
	Outputs   []*OutputLine
	Variables map[string]float64
	Drawings  []DrawingEvent
}

type Function func(args []*Value, bars []indicator.Bar) (*Value, error)

type FunctionInfo struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Arity       string `json:"arity"`
	Description string `json:"description"`
}

type builtinFn struct {
	fn          Function
	category    string
	arity       string
	description string
}

type FunctionRegistry struct {
	functions map[string]*builtinFn
}

func NewFunctionRegistry() *FunctionRegistry {
	r := &FunctionRegistry{functions: make(map[string]*builtinFn)}
	r.registerBuiltin()
	return r
}

// registerBuiltin is implemented in builtin.go and builtin_complex.go.

func (r *FunctionRegistry) Register(name, category, arity, description string, fn Function) {
	r.functions[normalizeName(name)] = &builtinFn{
		fn: fn, category: category, arity: arity, description: description,
	}
}

func (r *FunctionRegistry) Lookup(name string) (*builtinFn, bool) {
	fn, ok := r.functions[normalizeName(name)]
	return fn, ok
}

func (r *FunctionRegistry) Names() []string {
	names := make([]string, 0, len(r.functions))
	for n := range r.functions {
		names = append(names, n)
	}
	return names
}

func (r *FunctionRegistry) Call(name string, args []*Value, bars []indicator.Bar) (*Value, error) {
	fn, ok := r.Lookup(name)
	if !ok {
		return nil, NewEvalError(fmt.Sprintf("undefined function: %s", name))
	}
	return fn.fn(args, bars)
}

func normalizeName(name string) string {
	b := []byte(name)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] = b[i] - 'a' + 'A'
		}
	}
	return string(b)
}

type outputDeclaration struct {
	Name  string
	Value *Value
	Style *DrawingStyle
}

type Interpreter struct {
	Bars         []indicator.Bar
	Variables    map[string]*Value
	Outputs      []outputDeclaration
	Functions    *FunctionRegistry
	TradeSignals map[string][]float64
	BKColor      []float64
	ExprCount    int
}

func NewInterpreter(bars []indicator.Bar) *Interpreter {
	return &Interpreter{
		Bars:         bars,
		Variables:    make(map[string]*Value),
		Functions:    NewFunctionRegistry(),
		TradeSignals: make(map[string][]float64),
	}
}

func (interp *Interpreter) Execute(program *Program) (*FormulaResult, error) {
	interp.initMarketDataVariables()

	for _, stmt := range program.Body {
		if err := interp.executeStatement(stmt); err != nil {
			return nil, err
		}
	}

	return interp.buildResult(), nil
}

func (interp *Interpreter) initMarketDataVariables() {
	n := len(interp.Bars)

	open := make([]float64, n)
	close := make([]float64, n)
	high := make([]float64, n)
	low := make([]float64, n)
	vol := make([]float64, n)
	amount := make([]float64, n)

	for i, b := range interp.Bars {
		open[i] = b.Open
		close[i] = b.Close
		high[i] = b.High
		low[i] = b.Low
		vol[i] = b.Vol
		amount[i] = b.Amount
	}

	interp.Variables["OPEN"] = NewArrayValue(open)
	interp.Variables["CLOSE"] = NewArrayValue(close)
	interp.Variables["HIGH"] = NewArrayValue(high)
	interp.Variables["LOW"] = NewArrayValue(low)
	interp.Variables["VOLUME"] = NewArrayValue(vol)
	interp.Variables["VOL"] = interp.Variables["VOLUME"]
	interp.Variables["V"] = interp.Variables["VOLUME"]
	interp.Variables["AMOUNT"] = NewArrayValue(amount)
	interp.Variables["AMO"] = interp.Variables["AMOUNT"]
	interp.Variables["O"] = interp.Variables["OPEN"]
	interp.Variables["C"] = interp.Variables["CLOSE"]
	interp.Variables["H"] = interp.Variables["HIGH"]
	interp.Variables["L"] = interp.Variables["LOW"]
}

func (interp *Interpreter) executeStatement(stmt Stmt) error {
	switch s := stmt.(type) {
	case *AssignStmt:
		value, err := interp.evaluateExpression(s.Expr)
		if err != nil {
			return err
		}
		interp.Variables[normalizeName(s.Name)] = value
		return nil
	case *OutputStmt:
		value, err := interp.evaluateExpression(s.Expr)
		if err != nil {
			return err
		}
		interp.Variables[normalizeName(s.Name)] = value
		interp.Outputs = append(interp.Outputs, outputDeclaration{
			Name: s.Name, Value: value, Style: s.Style,
		})
		return nil
	case *TradeSignalStmt:
		value, err := interp.evaluateExpression(s.Expr)
		if err != nil {
			return err
		}
		interp.Variables[normalizeName(s.Kind)] = value
		interp.TradeSignals[s.Kind] = valueToArray(value)
		return nil
	case *BKColorStmt:
		value, err := interp.evaluateExpression(s.Expr)
		if err != nil {
			return err
		}
		interp.BKColor = valueToArray(value)
		return nil
	case *PlotCallStmt:
		value, err := interp.evaluatePlotCall(s)
		if err != nil {
			return err
		}
		interp.addDrawings(value)
		return nil
	case *ExpressionStmt:
		value, err := interp.evaluateExpression(s.Expr)
		if err != nil {
			return err
		}
		if value.IsDraw {
			interp.addDrawings(value)
			return nil
		}
		name := fmt.Sprintf("__expr__%d", interp.ExprCount)
		interp.ExprCount++
		if id, ok := s.Expr.(*IdentExpr); ok {
			name = id.Name
		}
		interp.Variables[normalizeName(name)] = value
		return nil
	default:
		return NewEvalError(fmt.Sprintf("unknown statement type: %T", stmt))
	}
}

func (interp *Interpreter) evaluatePlotCall(stmt *PlotCallStmt) (*Value, error) {
	call := &CallExpr{Type: "CallExpr", Func: stmt.Func, Args: stmt.Args}
	return interp.evaluateCallExpr(call)
}

func (interp *Interpreter) evaluateExpression(expr Expr) (*Value, error) {
	switch e := expr.(type) {
	case *NumberExpr:
		return NewSingleValue(e.Value), nil
	case *StringExpr:
		if e.External {
			if v, ok := interp.Variables[normalizeName(e.Value)]; ok {
				return v, nil
			}
			return nil, NewEvalError(fmt.Sprintf("undefined external reference: %s", e.Value))
		}
		return NewStringValue(e.Value), nil
	case *IdentExpr:
		v, ok := interp.Variables[normalizeName(e.Name)]
		if !ok {
			return nil, NewEvalError(fmt.Sprintf("undefined variable: %s", e.Name))
		}
		return v, nil
	case *UnaryExpr:
		return interp.evaluateUnaryExpr(e)
	case *BinaryExpr:
		return interp.evaluateBinaryExpr(e)
	case *CallExpr:
		return interp.evaluateCallExpr(e)
	default:
		return nil, NewEvalError(fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func (interp *Interpreter) evaluateCallExpr(call *CallExpr) (*Value, error) {
	args := make([]*Value, len(call.Args))
	for i, arg := range call.Args {
		v, err := interp.evaluateExpression(arg)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return interp.Functions.Call(call.Func, args, interp.Bars)
}

func (interp *Interpreter) evaluateUnaryExpr(expr *UnaryExpr) (*Value, error) {
	operand, err := interp.evaluateExpression(expr.X)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "-":
		if operand.IsArray {
			out := make([]float64, len(operand.Array))
			for i, v := range operand.Array {
				out[i] = -v
			}
			return NewArrayValue(out), nil
		}
		return NewSingleValue(-operand.Single), nil
	case "NOT":
		if operand.IsArray {
			out := make([]float64, len(operand.Array))
			for i, v := range operand.Array {
				out[i] = boolToFloat(!isTruthy(v))
			}
			return NewArrayValue(out), nil
		}
		return NewSingleValue(boolToFloat(!isTruthy(operand.Single))), nil
	default:
		return nil, NewEvalError(fmt.Sprintf("unknown unary operator: %s", expr.Op))
	}
}

func (interp *Interpreter) evaluateBinaryExpr(expr *BinaryExpr) (*Value, error) {
	left, err := interp.evaluateExpression(expr.L)
	if err != nil {
		return nil, err
	}
	right, err := interp.evaluateExpression(expr.R)
	if err != nil {
		return nil, err
	}

	switch {
	case left.IsArray && right.IsArray:
		return binaryOpArrays(expr.Op, left.Array, right.Array)
	case left.IsArray:
		return binaryOpArrayScalar(expr.Op, left.Array, right.Single)
	case right.IsArray:
		return binaryOpScalarArray(expr.Op, left.Single, right.Array)
	default:
		v, err := binaryOpScalar(expr.Op, left.Single, right.Single)
		if err != nil {
			return nil, err
		}
		return NewSingleValue(v), nil
	}
}

func (interp *Interpreter) buildResult() *FormulaResult {
	result := &FormulaResult{
		Outputs:   make([]*OutputLine, 0),
		Variables: make(map[string]float64),
		Drawings:  make([]DrawingEvent, 0),
	}

	for _, output := range interp.Outputs {
		if output.Value.IsDraw {
			result.Drawings = append(result.Drawings, output.Value.Draws...)
		} else if output.Value.IsArray {
			result.Outputs = append(result.Outputs, &OutputLine{
				Name: output.Name, Data: output.Value.Array, Style: lineStyleFromAST(output.Style),
			})
		} else if !output.Value.IsString {
			data := make([]float64, len(interp.Bars))
			for i := range data {
				data[i] = output.Value.Single
			}
			result.Outputs = append(result.Outputs, &OutputLine{
				Name: output.Name, Data: data, Style: lineStyleFromAST(output.Style),
			})
		}
	}

	for _, v := range interp.Variables {
		if v.IsDraw {
			result.Drawings = append(result.Drawings, v.Draws...)
		}
	}

	return result
}

func (interp *Interpreter) addDrawings(value *Value) {
	if value.IsDraw {
		interp.Variables[fmt.Sprintf("__draw__%d", len(interp.Outputs)+interp.ExprCount)] = value
	}
}

func lineStyleFromAST(style *DrawingStyle) *LineStyle {
	if style == nil {
		return nil
	}
	ls := &LineStyle{Hidden: style.Hidden}
	if style.Color != nil {
		ls.Color = *style.Color
	}
	if style.LineWidth != nil {
		ls.LineWidth = *style.LineWidth
	}
	if style.LineStyle != nil {
		ls.LineStyle = *style.LineStyle
	}
	if style.DrawMethod != nil {
		ls.DrawMethod = *style.DrawMethod
	}
	return ls
}

func valueToArray(v *Value) []float64 {
	if v == nil {
		return nil
	}
	if v.IsArray {
		return v.Array
	}
	return []float64{v.Single}
}

func binaryOpScalar(op string, a, b float64) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, NewEvalError("division by zero")
		}
		return a / b, nil
	case "^":
		return math.Pow(a, b), nil
	case ">":
		return boolToFloat(a > b), nil
	case "<":
		return boolToFloat(a < b), nil
	case ">=":
		return boolToFloat(a >= b), nil
	case "<=":
		return boolToFloat(a <= b), nil
	case "=":
		return boolToFloat(math.Abs(a-b) < 1e-10), nil
	case "<>":
		return boolToFloat(math.Abs(a-b) >= 1e-10), nil
	case "AND":
		return boolToFloat(isTruthy(a) && isTruthy(b)), nil
	case "OR":
		return boolToFloat(isTruthy(a) || isTruthy(b)), nil
	default:
		return 0, NewEvalError(fmt.Sprintf("unknown binary operator: %s", op))
	}
}

func binaryOpArrays(op string, a, b []float64) (*Value, error) {
	if len(a) != len(b) {
		return nil, NewEvalError("array length mismatch")
	}
	out := make([]float64, len(a))
	for i := range a {
		v, err := binaryOpScalar(op, a[i], b[i])
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return NewArrayValue(out), nil
}

func binaryOpArrayScalar(op string, arr []float64, s float64) (*Value, error) {
	out := make([]float64, len(arr))
	for i, v := range arr {
		r, err := binaryOpScalar(op, v, s)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return NewArrayValue(out), nil
}

func binaryOpScalarArray(op string, s float64, arr []float64) (*Value, error) {
	out := make([]float64, len(arr))
	for i, v := range arr {
		r, err := binaryOpScalar(op, s, v)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return NewArrayValue(out), nil
}

func isTruthy(v float64) bool {
	return v != 0 && !math.IsNaN(v)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

package formula

import (
	"math"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// ParseResult is the result of parsing a formula source.
type ParseResult struct {
	Type     string         `json:"type"`
	Outputs  []*OutputInfo  `json:"outputs"`
	Drawings []*DrawingCall `json:"drawings"`
	Body     []Stmt         `json:"body"`
}

// OutputInfo describes a declared output line found during parsing.
type OutputInfo struct {
	Name string `json:"name"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

// DrawingCall describes a drawing function call found during parsing.
type DrawingCall struct {
	Func string `json:"func"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

// ExecuteResult is the result of executing a formula against K-line data.
type ExecuteResult struct {
	Type          string                `json:"type"`
	Outputs       []*OutputLine         `json:"outputs"`
	Variables     map[string]float64    `json:"variables"`
	Drawings      []DrawingEvent        `json:"drawings"`
	TradeSignals  map[string][]float64  `json:"trade_signals,omitempty"`
	BKColor       []float64             `json:"bkcolor,omitempty"`
	NanCounts     []int                 `json:"nan_counts"`
	FunctionCount int                   `json:"function_count"`
}

// Engine is the formula engine facade.
type Engine struct {
	registry *FunctionRegistry
}

// NewEngine creates a formula engine with the full built-in function library.
func NewEngine() *Engine {
	return &Engine{registry: NewFunctionRegistry()}
}

// Parse parses formula source text and returns the AST, formula type,
// declared output lines and drawing calls.
func (e *Engine) Parse(src string) (*ParseResult, error) {
	tokens, err := NewLexer(src).Tokenize()
	if err != nil {
		return nil, err
	}
	program, err := NewParser(tokens).Parse()
	if err != nil {
		return nil, err
	}

	result := &ParseResult{
		Type:     program.Type,
		Outputs:  make([]*OutputInfo, 0),
		Drawings: make([]*DrawingCall, 0),
		Body:     program.Body,
	}

	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *OutputStmt:
			result.Outputs = append(result.Outputs, &OutputInfo{
				Name: s.Name, Line: s.Line, Col: s.Col,
			})
		case *PlotCallStmt:
			result.Drawings = append(result.Drawings, &DrawingCall{
				Func: s.Func, Line: s.Line, Col: s.Col,
			})
		case *TradeSignalStmt:
			result.Outputs = append(result.Outputs, &OutputInfo{
				Name: s.Kind, Line: s.Line, Col: s.Col,
			})
		case *BKColorStmt:
			result.Outputs = append(result.Outputs, &OutputInfo{
				Name: "BKCOLOR", Line: s.Line, Col: s.Col,
			})
		}
	}
	return result, nil
}

// Execute parses and executes a formula against the given bars.
func (e *Engine) Execute(src string, bars []indicator.Bar) (*ExecuteResult, error) {
	tokens, err := NewLexer(src).Tokenize()
	if err != nil {
		return nil, err
	}
	program, err := NewParser(tokens).Parse()
	if err != nil {
		return nil, err
	}

	interp := NewInterpreter(bars)
	result, err := interp.Execute(program)
	if err != nil {
		return nil, err
	}

	exec := &ExecuteResult{
		Type:          program.Type,
		Outputs:       result.Outputs,
		Variables:     result.Variables,
		Drawings:      result.Drawings,
		TradeSignals:  interp.TradeSignals,
		FunctionCount: len(interp.Functions.Names()),
	}

	if interp.BKColor != nil {
		exec.BKColor = interp.BKColor
	}

	for _, output := range result.Outputs {
		exec.NanCounts = append(exec.NanCounts, nanCount(output.Data))
	}
	return exec, nil
}

// ListFunctions returns metadata for all registered built-in functions.
func (e *Engine) ListFunctions() []FunctionInfo {
	names := e.registry.Names()
	infos := make([]FunctionInfo, 0, len(names))
	for _, name := range names {
		fn, ok := e.registry.Lookup(name)
		if !ok {
			continue
		}
		infos = append(infos, FunctionInfo{
			Name:        name,
			Category:    fn.category,
			Arity:       fn.arity,
			Description: fn.description,
		})
	}
	return infos
}

func nanCount(data []float64) int {
	count := 0
	for _, v := range data {
		if math.IsNaN(v) {
			count++
		}
	}
	return count
}

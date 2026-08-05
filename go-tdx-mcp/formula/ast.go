package formula

type FormulaType int

const (
	FormulaIndicator FormulaType = iota
	FormulaSelection
	FormulaTrade
	FormulaColorful
)

func (t FormulaType) String() string {
	switch t {
	case FormulaIndicator:
		return "indicator"
	case FormulaSelection:
		return "selection"
	case FormulaTrade:
		return "trade"
	case FormulaColorful:
		return "colorful"
	}
	return "unknown"
}

type Expr interface {
	exprNode()
}

type Stmt interface {
	stmtNode()
}

type Program struct {
	Type string `json:"type"`
	Body []Stmt `json:"body"`
}

type AssignStmt struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Expr Expr   `json:"expr"`
}

type OutputStmt struct {
	Type  string        `json:"type"`
	Name  string        `json:"name"`
	Expr  Expr          `json:"expr"`
	Style *DrawingStyle `json:"style,omitempty"`
	Line  int           `json:"line,omitempty"`
	Col   int           `json:"col,omitempty"`
}

type TradeSignalStmt struct {
	Type string `json:"type"`
	Kind string `json:"kind"`
	Expr Expr   `json:"expr"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

type BKColorStmt struct {
	Type string `json:"type"`
	Expr Expr   `json:"expr"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

type PlotCallStmt struct {
	Type string `json:"type"`
	Func string `json:"func"`
	Args []Expr `json:"args"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

type ExpressionStmt struct {
	Type string `json:"type"`
	Expr Expr   `json:"expr"`
}

type NumberExpr struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

type StringExpr struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	External bool   `json:"external,omitempty"`
}

type IdentExpr struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type UnaryExpr struct {
	Type string `json:"type"`
	Op   string `json:"op"`
	X    Expr   `json:"x"`
}

type BinaryExpr struct {
	Type string `json:"type"`
	Op   string `json:"op"`
	L    Expr   `json:"left"`
	R    Expr   `json:"right"`
}

type CallExpr struct {
	Type string `json:"type"`
	Func string `json:"func"`
	Args []Expr `json:"args"`
}

type DrawingStyle struct {
	Color      *string `json:"color,omitempty"`
	LineWidth  *int    `json:"line_width,omitempty"`
	LineStyle  *string `json:"line_style,omitempty"`
	DrawMethod *string `json:"draw_method,omitempty"`
	Hidden     bool    `json:"hidden"`
}

func (*Program) stmtNode()      {}
func (*AssignStmt) stmtNode()   {}
func (*OutputStmt) stmtNode()   {}
func (*TradeSignalStmt) stmtNode() {}
func (*BKColorStmt) stmtNode()  {}
func (*PlotCallStmt) stmtNode() {}
func (*ExpressionStmt) stmtNode() {}

func (*NumberExpr) exprNode()  {}
func (*StringExpr) exprNode()  {}
func (*IdentExpr) exprNode()   {}
func (*UnaryExpr) exprNode()   {}
func (*BinaryExpr) exprNode()  {}
func (*CallExpr) exprNode()    {}

func inferFormulaType(p *Program) FormulaType {
	hasTrade := false
	hasBKColor := false
	hasOutput := false
	for _, s := range p.Body {
		switch s.(type) {
		case *TradeSignalStmt:
			hasTrade = true
		case *BKColorStmt:
			hasBKColor = true
		case *OutputStmt:
			hasOutput = true
		}
	}
	switch {
	case hasTrade:
		return FormulaTrade
	case hasBKColor:
		return FormulaColorful
	case hasOutput:
		return FormulaIndicator
	default:
		return FormulaSelection
	}
}

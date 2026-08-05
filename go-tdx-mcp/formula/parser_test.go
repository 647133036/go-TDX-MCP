package formula

import (
	"strings"
	"testing"
)

func parseFormula(t *testing.T, src string) *Program {
	t.Helper()
	toks, err := NewLexer(src).Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	prog, err := NewParser(toks).Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	return prog
}

func TestParseAssignStmt(t *testing.T) {
	prog := parseFormula(t, "MA5:=MA(CLOSE,5);")
	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Body))
	}
	stmt, ok := prog.Body[0].(*AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", prog.Body[0])
	}
	if stmt.Name != "MA5" {
		t.Errorf("expected name MA5, got %s", stmt.Name)
	}
	call, ok := stmt.Expr.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", stmt.Expr)
	}
	if call.Func != "MA" {
		t.Errorf("expected func MA, got %s", call.Func)
	}
	if len(call.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(call.Args))
	}
}

func TestParseOutputStmt(t *testing.T) {
	prog := parseFormula(t, "DIF:EMA(CLOSE,12),COLORWHITE,LINETHICK2;")
	stmt := prog.Body[0].(*OutputStmt)
	if stmt.Name != "DIF" {
		t.Errorf("expected name DIF, got %s", stmt.Name)
	}
	if stmt.Style == nil {
		t.Fatal("expected style to be parsed")
	}
	if stmt.Style.Color == nil || *stmt.Style.Color != "WHITE" {
		t.Errorf("expected color WHITE, got %v", stmt.Style.Color)
	}
	if stmt.Style.LineWidth == nil || *stmt.Style.LineWidth != 2 {
		t.Errorf("expected line width 2, got %v", stmt.Style.LineWidth)
	}
}

func TestParseStyleVariants(t *testing.T) {
	prog := parseFormula(t, "A:B,STICK,NODRAW,DOTLINE,COLORSTICK,POINTDOT;")
	stmt := prog.Body[0].(*OutputStmt)
	if stmt.Style == nil {
		t.Fatal("expected style")
	}
	if stmt.Style.DrawMethod == nil || *stmt.Style.DrawMethod != "pointdot" {
		t.Errorf("expected draw method pointdot (last wins), got %v", stmt.Style.DrawMethod)
	}
	if !stmt.Style.Hidden {
		t.Error("expected hidden=true")
	}
	if stmt.Style.LineStyle == nil || *stmt.Style.LineStyle != "dotted" {
		t.Errorf("expected line style dotted, got %v", stmt.Style.LineStyle)
	}
}

func TestParsePrecedence(t *testing.T) {
	prog := parseFormula(t, "A:=1+2*3;")
	stmt := prog.Body[0].(*AssignStmt)
	bin, ok := stmt.Expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}
	if bin.Op != "+" {
		t.Errorf("expected top op +, got %s", bin.Op)
	}
	right, ok := bin.R.(*BinaryExpr)
	if !ok || right.Op != "*" {
		t.Errorf("expected right subtree *, got %T %v", bin.R, bin.R)
	}
}

func TestParseComparisonChain(t *testing.T) {
	prog := parseFormula(t, "A:=C>O AND H>=L;")
	stmt := prog.Body[0].(*AssignStmt)
	top, ok := stmt.Expr.(*BinaryExpr)
	if !ok || top.Op != "AND" {
		t.Fatalf("expected AND at top, got %v", stmt.Expr)
	}
}

func TestParsePowerRightAssoc(t *testing.T) {
	prog := parseFormula(t, "A:=2^3^2;")
	stmt := prog.Body[0].(*AssignStmt)
	top, ok := stmt.Expr.(*BinaryExpr)
	if !ok || top.Op != "^" {
		t.Fatalf("expected ^ at top, got %v", stmt.Expr)
	}
	if _, ok := top.R.(*BinaryExpr); !ok {
		t.Errorf("expected right-assoc power, got %T", top.R)
	}
}

func TestParseTradeSignal(t *testing.T) {
	prog := parseFormula(t, "ENTERLONG:CROSS(MA5,MA10);EXITLONG:CROSS(MA10,MA5);")
	if len(prog.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Body))
	}
	first, ok := prog.Body[0].(*TradeSignalStmt)
	if !ok {
		t.Fatalf("expected TradeSignalStmt, got %T", prog.Body[0])
	}
	if first.Kind != "ENTERLONG" {
		t.Errorf("expected ENTERLONG, got %s", first.Kind)
	}
}

func TestParseBKColor(t *testing.T) {
	prog := parseFormula(t, "BKCOLOR:IF(C>O,1,0);")
	if _, ok := prog.Body[0].(*BKColorStmt); !ok {
		t.Fatalf("expected BKColorStmt, got %T", prog.Body[0])
	}
}

func TestParsePlotCall(t *testing.T) {
	prog := parseFormula(t, "DRAWTEXT(C>O,LOW,'UP');")
	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Body))
	}
	stmt, ok := prog.Body[0].(*PlotCallStmt)
	if !ok {
		t.Fatalf("expected PlotCallStmt, got %T", prog.Body[0])
	}
	if stmt.Func != "DRAWTEXT" {
		t.Errorf("expected DRAWTEXT, got %s", stmt.Func)
	}
	if len(stmt.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(stmt.Args))
	}
}

func TestParseExpressionStmt(t *testing.T) {
	prog := parseFormula(t, "CROSS(MA5,MA10);")
	if _, ok := prog.Body[0].(*ExpressionStmt); !ok {
		t.Fatalf("expected ExpressionStmt, got %T", prog.Body[0])
	}
}

func TestParseTypeInference(t *testing.T) {
	cases := []struct {
		src  string
		want FormulaType
	}{
		{"DIF:EMA(C,12);DEA:EMA(DIF,9);", FormulaIndicator},
		{"MA5:=MA(C,5);MA10:=MA(C,10);CROSS(MA5,MA10);", FormulaSelection},
		{"ENTERLONG:CROSS(MA5,MA10);EXITLONG:CROSS(MA10,MA5);", FormulaTrade},
		{"BKCOLOR:IF(C>O,1,0);", FormulaColorful},
	}
	for _, c := range cases {
		prog := parseFormula(t, c.src)
		if got := inferFormulaType(prog); got != c.want {
			t.Errorf("for %q: expected %s, got %s", c.src, c.want, got)
		}
	}
}

func TestParseUnmatchedParen(t *testing.T) {
	_, err := NewParser(mustTokenize(t, "A:=(1+2;")).Parse()
	if err == nil {
		t.Fatal("expected parse error for unmatched paren")
	}
	if fe, ok := err.(*FormulaError); ok {
		if fe.Kind != KindParse {
			t.Errorf("expected parse error kind, got %s", fe.Kind)
		}
	} else {
		t.Errorf("expected *FormulaError, got %T", err)
	}
}

func TestParseUnexpectedToken(t *testing.T) {
	_, err := NewParser(mustTokenize(t, "A:=;")).Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("expected unexpected token message, got %v", err)
	}
}

func mustTokenize(t *testing.T, src string) []Token {
	t.Helper()
	toks, err := NewLexer(src).Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	return toks
}

func TestParseMultiLine(t *testing.T) {
	prog := parseFormula(t, "MA5:=MA(C,5)\nMA10:=MA(C,10)\n")
	if len(prog.Body) != 2 {
		t.Fatalf("expected 2 statements across newlines, got %d", len(prog.Body))
	}
}

func TestParseChineseIdent(t *testing.T) {
	prog := parseFormula(t, "阻力1:MA(C,5);")
	stmt := prog.Body[0].(*OutputStmt)
	if stmt.Name != "阻力1" {
		t.Errorf("expected Chinese identifier 阻力1, got %s", stmt.Name)
	}
}

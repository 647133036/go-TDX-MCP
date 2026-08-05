package formula

import (
	"fmt"
	"strconv"
	"strings"
)

var plotFunctionNames = map[string]bool{
	"DRAWTEXT":   true,
	"DRAWICON":   true,
	"DRAWNUMBER": true,
	"STICKLINE":  true,
	"DRAWLINE":   true,
	"POLYLINE":   true,
	"DRAWBAND":   true,
	"DRAWKLINE":  true,
}

var tradeSignalNames = map[string]bool{
	"ENTERLONG":  true,
	"EXITLONG":   true,
	"ENTERSHORT": true,
	"EXITSHORT":  true,
}

type Parser struct {
	tokens  []Token
	pos     int
	current Token
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens, pos: 0}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

func (p *Parser) Parse() (*Program, error) {
	body := make([]Stmt, 0)
	for !p.isAtEnd() {
		if p.current.Type == TokenNewline {
			p.advance()
			continue
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	program := &Program{Body: body}
	program.Type = inferFormulaType(program).String()
	return program, nil
}

func (p *Parser) parseStatement() (Stmt, error) {
	if p.current.Type == TokenIdent {
		upper := strings.ToUpper(p.current.Value)
		if p.peek().Type == TokenColon {
			if tradeSignalNames[upper] {
				return p.parseTradeSignalStmt()
			}
			if upper == "BKCOLOR" {
				return p.parseBKColorStmt()
			}
			return p.parseOutputStmt()
		}
		if p.peek().Type == TokenAssign {
			return p.parseAssignStmt()
		}
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consumeStatementEnd(); err != nil {
		return nil, err
	}

	if call, ok := expr.(*CallExpr); ok && plotFunctionNames[strings.ToUpper(call.Func)] {
		return &PlotCallStmt{
			Type: "PlotCallStmt", Func: strings.ToUpper(call.Func), Args: call.Args,
			Line: p.current.Line, Col: p.current.Col,
		}, nil
	}
	return &ExpressionStmt{Type: "ExpressionStmt", Expr: expr}, nil
}

func (p *Parser) parseTradeSignalStmt() (Stmt, error) {
	line, col := p.current.Line, p.current.Col
	name := p.current.Value
	p.advance()
	p.advance() // consume ':'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeStatementEnd(); err != nil {
		return nil, err
	}
	return &TradeSignalStmt{Type: "TradeSignalStmt", Kind: strings.ToUpper(name), Expr: expr, Line: line, Col: col}, nil
}

func (p *Parser) parseBKColorStmt() (Stmt, error) {
	line, col := p.current.Line, p.current.Col
	p.advance()
	p.advance() // consume ':'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeStatementEnd(); err != nil {
		return nil, err
	}
	return &BKColorStmt{Type: "BKColorStmt", Expr: expr, Line: line, Col: col}, nil
}

func (p *Parser) parseOutputStmt() (Stmt, error) {
	line, col := p.current.Line, p.current.Col
	name := p.current.Value
	p.advance()
	p.advance() // consume ':'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	style, err := p.parseStyleSuffixes()
	if err != nil {
		return nil, err
	}
	if err := p.consumeStatementEnd(); err != nil {
		return nil, err
	}
	return &OutputStmt{Type: "OutputStmt", Name: name, Expr: expr, Style: style, Line: line, Col: col}, nil
}

func (p *Parser) parseAssignStmt() (Stmt, error) {
	name := p.current.Value
	p.advance()
	p.advance() // consume ':='
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.consumeStatementEnd(); err != nil {
		return nil, err
	}
	return &AssignStmt{Type: "AssignStmt", Name: name, Expr: expr}, nil
}

func (p *Parser) parseExpression() (Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for !p.isAtEnd() && p.current.Type == TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: "OR", L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for !p.isAtEnd() && p.current.Type == TokenAnd {
		p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: "AND", L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for !p.isAtEnd() {
		var op string
		switch p.current.Type {
		case TokenGT:
			op = ">"
		case TokenLT:
			op = "<"
		case TokenGTE:
			op = ">="
		case TokenLTE:
			op = "<="
		case TokenEQ:
			op = "="
		case TokenNEQ:
			op = "<>"
		default:
			return left, nil
		}
		p.advance()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parseAdditive() (Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for !p.isAtEnd() && (p.current.Type == TokenPlus || p.current.Type == TokenMinus) {
		op := "+"
		if p.current.Type == TokenMinus {
			op = "-"
		}
		p.advance()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parseMultiplicative() (Expr, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}
	for !p.isAtEnd() && (p.current.Type == TokenMultiply || p.current.Type == TokenDivide) {
		op := "*"
		if p.current.Type == TokenDivide {
			op = "/"
		}
		p.advance()
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: op, L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parsePower() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	if !p.isAtEnd() && p.current.Type == TokenPower {
		p.advance()
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Type: "BinaryExpr", Op: "^", L: left, R: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if !p.isAtEnd() && (p.current.Type == TokenMinus || p.current.Type == TokenNot) {
		op := "-"
		if p.current.Type == TokenNot {
			op = "NOT"
		}
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Type: "UnaryExpr", Op: op, X: operand}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.isAtEnd() {
		return nil, p.error("unexpected end of input")
	}

	switch p.current.Type {
	case TokenNumber:
		return p.parseNumber()
	case TokenString:
		return p.parseString(false)
	case TokenExternalRef:
		return p.parseString(true)
	case TokenIdent:
		return p.parseIdentOrCall()
	case TokenLParen:
		return p.parseGrouped()
	default:
		return nil, p.error(fmt.Sprintf("unexpected token: %s", p.current.Type))
	}
}

func (p *Parser) parseNumber() (Expr, error) {
	value, err := strconv.ParseFloat(p.current.Value, 64)
	if err != nil {
		return nil, p.error(fmt.Sprintf("invalid number: %s", p.current.Value))
	}
	p.advance()
	return &NumberExpr{Type: "NumberExpr", Value: value}, nil
}

func (p *Parser) parseString(external bool) (Expr, error) {
	value := p.current.Value
	p.advance()
	return &StringExpr{Type: "StringExpr", Value: value, External: external}, nil
}

func (p *Parser) parseIdentOrCall() (Expr, error) {
	name := p.current.Value
	p.advance()

	if !p.isAtEnd() && p.current.Type == TokenLParen {
		return p.parseCallArgs(name)
	}
	return &IdentExpr{Type: "IdentExpr", Name: name}, nil
}

func (p *Parser) parseCallArgs(name string) (Expr, error) {
	p.advance() // consume '('

	args := make([]Expr, 0)
	for !p.isAtEnd() && p.current.Type != TokenRParen {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.current.Type == TokenComma {
			p.advance()
		} else if p.current.Type != TokenRParen {
			return nil, p.error("expected ',' or ')' in function call")
		}
	}

	if p.current.Type != TokenRParen {
		return nil, p.error("expected ')' after function arguments")
	}
	p.advance()

	return &CallExpr{Type: "CallExpr", Func: name, Args: args}, nil
}

func (p *Parser) parseGrouped() (Expr, error) {
	p.advance() // consume '('

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.current.Type != TokenRParen {
		return nil, p.error("expected ')' after expression")
	}
	p.advance()
	return expr, nil
}

func (p *Parser) parseStyleSuffixes() (*DrawingStyle, error) {
	var style *DrawingStyle

	for !p.isAtEnd() && p.current.Type == TokenComma {
		next := p.peek()
		if !p.isStyleToken(next) {
			return style, nil
		}
		p.advance() // consume ','
		if style == nil {
			style = &DrawingStyle{}
		}
		if err := p.applyStyleToken(style, p.current); err != nil {
			return nil, err
		}
		p.advance()
	}

	return style, nil
}

func (p *Parser) isStyleToken(tok Token) bool {
	switch tok.Type {
	case TokenColor, TokenLineThick, TokenDotLine, TokenStick, TokenColorStick,
		TokenVolStick, TokenNoDraw, TokenPointDot, TokenCircleDot, TokenCrossDot:
		return true
	default:
		return false
	}
}

func (p *Parser) applyStyleToken(style *DrawingStyle, tok Token) error {
	switch tok.Type {
	case TokenColor:
		if strings.EqualFold(tok.Value, "COLORSTICK") {
			m := "colorstick"
			style.DrawMethod = &m
			return nil
		}
		upper := strings.ToUpper(tok.Value)
		if strings.HasPrefix(upper, "COLOR") && len(upper) > 5 {
			c := upper[5:]
			style.Color = &c
		} else {
			style.Color = &upper
		}
	case TokenLineThick:
		w := 1
		if text := strings.TrimPrefix(strings.ToUpper(tok.Value), "LINETHICK"); text != "" {
			if v, err := strconv.Atoi(text); err == nil {
				w = v
			}
		}
		style.LineWidth = &w
	case TokenDotLine:
		ls := "dotted"
		style.LineStyle = &ls
	case TokenStick:
		m := "stick"
		style.DrawMethod = &m
	case TokenColorStick:
		m := "colorstick"
		style.DrawMethod = &m
	case TokenVolStick:
		m := "volstick"
		style.DrawMethod = &m
	case TokenNoDraw:
		style.Hidden = true
	case TokenPointDot:
		m := "pointdot"
		style.DrawMethod = &m
	case TokenCircleDot:
		m := "circledot"
		style.DrawMethod = &m
	case TokenCrossDot:
		m := "crossdot"
		style.DrawMethod = &m
	}
	return nil
}

func (p *Parser) consumeStatementEnd() error {
	if !p.isAtEnd() && p.current.Type == TokenSemicolon {
		p.advance()
		return nil
	}
	if !p.isAtEnd() && p.current.Type == TokenNewline {
		p.advance()
		return nil
	}
	return nil
}

func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.pos++
		if p.pos < len(p.tokens) {
			p.current = p.tokens[p.pos]
		}
	}
}

func (p *Parser) peek() Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) isAtEnd() bool {
	return p.current.Type == TokenEOF
}

func (p *Parser) error(message string) error {
	return NewParseError(message, p.current.Line, p.current.Col)
}

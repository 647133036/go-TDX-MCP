package formula

import "fmt"

type TokenType string

const (
	TokenNumber      TokenType = "NUMBER"
	TokenString      TokenType = "STRING"
	TokenExternalRef TokenType = "EXTERNAL_REFERENCE"
	TokenIdent       TokenType = "IDENT"
	TokenEOF         TokenType = "EOF"
	TokenNewline     TokenType = "NEWLINE"

	TokenPlus     TokenType = "PLUS"
	TokenMinus    TokenType = "MINUS"
	TokenMultiply TokenType = "MULTIPLY"
	TokenDivide   TokenType = "DIVIDE"
	TokenPower    TokenType = "POWER"

	TokenGT  TokenType = "GT"
	TokenLT  TokenType = "LT"
	TokenGTE TokenType = "GTE"
	TokenLTE TokenType = "LTE"
	TokenEQ  TokenType = "EQ"
	TokenNEQ TokenType = "NEQ"

	TokenAnd TokenType = "AND"
	TokenOr  TokenType = "OR"
	TokenNot TokenType = "NOT"

	TokenLParen    TokenType = "LPAREN"
	TokenRParen    TokenType = "RPAREN"
	TokenComma     TokenType = "COMMA"
	TokenSemicolon TokenType = "SEMICOLON"
	TokenColon     TokenType = "COLON"
	TokenAssign    TokenType = "ASSIGN"
	TokenHash      TokenType = "HASH"

	TokenColor      TokenType = "COLOR"
	TokenLineThick  TokenType = "LINETHICK"
	TokenDotLine    TokenType = "DOTLINE"
	TokenStick      TokenType = "STICK"
	TokenColorStick TokenType = "COLORSTICK"
	TokenVolStick   TokenType = "VOLSTICK"
	TokenNoDraw     TokenType = "NODRAW"
	TokenPointDot   TokenType = "POINTDOT"
	TokenCircleDot  TokenType = "CIRCLEDOT"
	TokenCrossDot   TokenType = "CROSSDOT"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func NewToken(t TokenType, value string, line, col int) Token {
	return Token{Type: t, Value: value, Line: line, Col: col}
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Type, t.Value, t.Line, t.Col)
}

func (t Token) Is(tt TokenType) bool {
	return t.Type == tt
}

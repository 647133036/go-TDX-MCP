package formula

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
	tokens []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
		tokens: make([]Token, 0),
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for !l.isAtEnd() {
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}
	l.tokens = append(l.tokens, NewToken(TokenEOF, "", l.line, l.column))
	return l.tokens, nil
}

func (l *Lexer) scanToken() error {
	for !l.isAtEnd() && l.isSpace(l.peek()) {
		l.advance()
	}
	if l.isAtEnd() {
		return nil
	}

	ch := l.peek()

	if ch == '\n' {
		l.tokens = append(l.tokens, NewToken(TokenNewline, "\n", l.line, l.column))
		l.advance()
		l.line++
		l.column = 1
		return nil
	}

	if ch == '/' && l.peekNext() == '/' {
		l.skipLineComment()
		return nil
	}

	if ch == '{' {
		return l.skipBlockComment()
	}

	if unicode.IsDigit(ch) {
		return l.scanNumber()
	}

	if ch == '\'' || ch == '"' {
		return l.scanString()
	}

	if unicode.IsLetter(ch) || ch == '_' {
		return l.scanIdentifier()
	}

	return l.scanOperator()
}

func (l *Lexer) skipLineComment() {
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() error {
	startLine, startCol := l.line, l.column
	l.advance()
	depth := 1
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				l.advance()
				return nil
			}
		} else if ch == '\n' {
			l.advance()
			l.line++
			l.column = 1
			continue
		}
		l.advance()
	}
	return NewLexError("unterminated block comment", startLine, startCol)
}

func (l *Lexer) scanNumber() error {
	start := l.pos
	startCol := l.column

	// Check for hex number: 0xHHHH or 0XHHHH
	if l.peek() == '0' {
		next := l.peekNext()
		if next == 'x' || next == 'X' {
			l.advance() // consume '0'
			l.advance() // consume 'x'/'X'
			for !l.isAtEnd() {
				ch := l.peek()
				if isHexDigit(ch) {
					l.advance()
				} else {
					break
				}
			}
			l.tokens = append(l.tokens, NewToken(TokenNumber, l.input[start:l.pos], l.line, startCol))
			return nil
		}
	}

	for !l.isAtEnd() && unicode.IsDigit(l.peek()) {
		l.advance()
	}

	if !l.isAtEnd() && l.peek() == '.' {
		l.advance()
		for !l.isAtEnd() && unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}

	if !l.isAtEnd() && (l.peek() == 'e' || l.peek() == 'E') {
		l.advance()
		if !l.isAtEnd() && (l.peek() == '+' || l.peek() == '-') {
			l.advance()
		}
		for !l.isAtEnd() && unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}

	l.tokens = append(l.tokens, NewToken(TokenNumber, l.input[start:l.pos], l.line, startCol))
	return nil
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func (l *Lexer) scanString() error {
	quote := l.advance()
	start := l.pos
	startCol := l.column - 1

	for !l.isAtEnd() && l.peek() != quote {
		if l.peek() == '\n' {
			return NewLexError("unterminated string", l.line, startCol)
		}
		l.advance()
	}

	if l.isAtEnd() {
		return NewLexError("unterminated string", l.line, startCol)
	}

	value := l.input[start:l.pos]
	l.advance()

	tt := TokenString
	if quote == '"' {
		tt = TokenExternalRef
	}
	l.tokens = append(l.tokens, NewToken(tt, value, l.line, startCol))
	return nil
}

func (l *Lexer) scanIdentifier() error {
	start := l.pos
	startCol := l.column

	for !l.isAtEnd() {
		r := l.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			l.advance()
		} else {
			break
		}
	}

	value := l.input[start:l.pos]
	upper := strings.ToUpper(value)

	tt := TokenIdent
	switch {
	case upper == "AND":
		tt = TokenAnd
	case upper == "OR":
		tt = TokenOr
	case upper == "NOT":
		tt = TokenNot
	case upper == "DOTLINE":
		tt = TokenDotLine
	case upper == "STICK":
		tt = TokenStick
	case upper == "COLORSTICK":
		tt = TokenColorStick
	case upper == "VOLSTICK":
		tt = TokenVolStick
	case upper == "NODRAW":
		tt = TokenNoDraw
	case upper == "POINTDOT":
		tt = TokenPointDot
	case upper == "CIRCLEDOT":
		tt = TokenCircleDot
	case upper == "CROSSDOT":
		tt = TokenCrossDot
	case strings.HasPrefix(upper, "COLOR"):
		tt = TokenColor
	case strings.HasPrefix(upper, "LINETHICK"):
		tt = TokenLineThick
	}

	l.tokens = append(l.tokens, NewToken(tt, value, l.line, startCol))
	return nil
}

func (l *Lexer) scanOperator() error {
	startCol := l.column
	ch := l.advance()

	switch ch {
	case '+':
		l.addToken(TokenPlus, "+", startCol)
	case '-':
		l.addToken(TokenMinus, "-", startCol)
	case '*':
		l.addToken(TokenMultiply, "*", startCol)
	case '/':
		l.addToken(TokenDivide, "/", startCol)
	case '^':
		l.addToken(TokenPower, "^", startCol)
	case '(':
		l.addToken(TokenLParen, "(", startCol)
	case ')':
		l.addToken(TokenRParen, ")", startCol)
	case ',':
		l.addToken(TokenComma, ",", startCol)
	case ';':
		l.addToken(TokenSemicolon, ";", startCol)
	case ':':
		if !l.isAtEnd() && l.peek() == '=' {
			l.advance()
			l.addToken(TokenAssign, ":=", startCol)
		} else {
			l.addToken(TokenColon, ":", startCol)
		}
	case '>':
		if !l.isAtEnd() && l.peek() == '=' {
			l.advance()
			l.addToken(TokenGTE, ">=", startCol)
		} else {
			l.addToken(TokenGT, ">", startCol)
		}
	case '<':
		if !l.isAtEnd() && l.peek() == '=' {
			l.advance()
			l.addToken(TokenLTE, "<=", startCol)
		} else if !l.isAtEnd() && l.peek() == '>' {
			l.advance()
			l.addToken(TokenNEQ, "<>", startCol)
		} else {
			l.addToken(TokenLT, "<", startCol)
		}
	case '=':
		if !l.isAtEnd() && l.peek() == '=' {
			l.advance()
			l.addToken(TokenEQ, "==", startCol)
		} else {
			l.addToken(TokenEQ, "=", startCol)
		}
	case '!':
		if !l.isAtEnd() && l.peek() == '=' {
			l.advance()
			l.addToken(TokenNEQ, "!=", startCol)
		} else {
			return NewLexError("unexpected character", l.line, startCol)
		}
	case '#':
		l.addToken(TokenHash, "#", startCol)
	default:
		return NewLexError(fmt.Sprintf("unexpected character: %c", ch), l.line, startCol)
	}

	return nil
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return ch
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	_, size := utf8.DecodeRuneInString(l.input[l.pos:])
	if l.pos+size >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.pos+size:])
	return ch
}

func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0
	}
	ch, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	l.column++
	return ch
}

func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func (l *Lexer) addToken(tt TokenType, value string, col int) {
	l.tokens = append(l.tokens, NewToken(tt, value, l.line, col))
}

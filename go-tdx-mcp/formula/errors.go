package formula

import "fmt"

type ErrorKind string

const (
	KindLex    ErrorKind = "lex"
	KindParse  ErrorKind = "parse"
	KindEval   ErrorKind = "eval"
	KindVerify ErrorKind = "verify"
)

type FormulaError struct {
	Kind    ErrorKind `json:"type"`
	Message string    `json:"message"`
	Line    int       `json:"line,omitempty"`
	Col     int       `json:"col,omitempty"`
}

func (e *FormulaError) Error() string {
	switch {
	case e.Line > 0 && e.Col > 0:
		return fmt.Sprintf("%s error at line %d, column %d: %s", e.Kind, e.Line, e.Col, e.Message)
	case e.Line > 0:
		return fmt.Sprintf("%s error at line %d: %s", e.Kind, e.Line, e.Message)
	default:
		return fmt.Sprintf("%s error: %s", e.Kind, e.Message)
	}
}

func NewLexError(message string, line, col int) *FormulaError {
	return &FormulaError{Kind: KindLex, Message: message, Line: line, Col: col}
}

func NewParseError(message string, line, col int) *FormulaError {
	return &FormulaError{Kind: KindParse, Message: message, Line: line, Col: col}
}

func NewEvalError(message string) *FormulaError {
	return &FormulaError{Kind: KindEval, Message: message}
}

func NewVerifyError(message string) *FormulaError {
	return &FormulaError{Kind: KindVerify, Message: message}
}

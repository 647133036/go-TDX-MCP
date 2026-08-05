package formula

import (
	"strings"
	"testing"
)

func TestLexerSimpleExpression(t *testing.T) {
	toks, err := NewLexer("MA5:MA(CLOSE,5);").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []TokenType{
		TokenIdent, TokenColon, TokenIdent, TokenLParen,
		TokenIdent, TokenComma, TokenNumber, TokenRParen,
		TokenSemicolon, TokenEOF,
	}
	if len(toks) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(toks), toks)
	}
	for i, tt := range want {
		if toks[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s (%q)", i, tt, toks[i].Type, toks[i].Value)
		}
	}
}

func TestLexerComments(t *testing.T) {
	src := "A:=1; // line comment\n{block\ncomment}B:=2;"
	toks, err := NewLexer(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var idents []string
	for _, tok := range toks {
		if tok.Type == TokenIdent {
			idents = append(idents, tok.Value)
		}
	}
	want := []string{"A", "B"}
	if len(idents) != len(want) {
		t.Fatalf("expected idents %v, got %v", want, idents)
	}
	for i, v := range want {
		if idents[i] != v {
			t.Errorf("ident[%d]: expected %s, got %s", i, v, idents[i])
		}
	}
}

func TestLexerUnterminatedBlockComment(t *testing.T) {
	_, err := NewLexer("{abc").Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated block comment")
	}
	if fe, ok := err.(*FormulaError); ok {
		if fe.Kind != KindLex {
			t.Errorf("expected lex error, got %s", fe.Kind)
		}
	} else {
		t.Errorf("expected *FormulaError, got %T", err)
	}
}

func TestLexerUnterminatedString(t *testing.T) {
	_, err := NewLexer("'abc").Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
	if !strings.Contains(err.Error(), "unterminated string") {
		t.Errorf("expected unterminated string message, got %v", err)
	}
}

func TestLexerStrings(t *testing.T) {
	toks, err := NewLexer("DRAWTEXT(C>O,L,'UP');").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundString := false
	foundExternal := false
	for _, tok := range toks {
		switch tok.Type {
		case TokenString:
			foundString = true
			if tok.Value != "UP" {
				t.Errorf("expected string value UP, got %q", tok.Value)
			}
		case TokenExternalRef:
			foundExternal = true
		}
	}
	if !foundString {
		t.Error("expected single-quoted STRING token")
	}
	if foundExternal {
		t.Error("did not expect EXTERNAL_REFERENCE token")
	}
}

func TestLexerExternalReference(t *testing.T) {
	toks, err := NewLexer(`"MACD.DIF#WEEK";`).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, tok := range toks {
		if tok.Type == TokenExternalRef {
			found = true
			if tok.Value != "MACD.DIF#WEEK" {
				t.Errorf("expected value MACD.DIF#WEEK, got %q", tok.Value)
			}
		}
	}
	if !found {
		t.Error("expected EXTERNAL_REFERENCE token for double-quoted string")
	}
}

func TestLexerOperators(t *testing.T) {
	toks, err := NewLexer("A:=(B+C)*D>=E;F<G<>H;I<=J;K=L;M!=N;O^P;").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []TokenType{
		TokenIdent, TokenAssign, TokenLParen, TokenIdent, TokenPlus, TokenIdent,
		TokenRParen, TokenMultiply, TokenIdent, TokenGTE, TokenIdent, TokenSemicolon,
		TokenIdent, TokenLT, TokenIdent, TokenNEQ, TokenIdent, TokenSemicolon,
		TokenIdent, TokenLTE, TokenIdent, TokenSemicolon,
		TokenIdent, TokenEQ, TokenIdent, TokenSemicolon,
		TokenIdent, TokenNEQ, TokenIdent, TokenSemicolon,
		TokenIdent, TokenPower, TokenIdent, TokenSemicolon,
		TokenEOF,
	}
	if len(toks) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(toks), toks)
	}
	for i, tt := range want {
		if toks[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s (%q)", i, tt, toks[i].Type, toks[i].Value)
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	toks, err := NewLexer("AND OR NOT IF").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []TokenType{TokenAnd, TokenOr, TokenNot, TokenIdent}
	if len(toks) != len(want)+1 {
		t.Fatalf("expected %d tokens, got %d", len(want)+1, len(toks))
	}
	for i, tt := range want {
		if toks[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s", i, tt, toks[i].Type)
		}
	}
}

func TestLexerStyleAttributes(t *testing.T) {
	toks, err := NewLexer("DIF:EMA(C,12),COLORWHITE,LINETHICK2,STICK,NODRAW,DOTLINE;").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var types []TokenType
	for _, tok := range toks {
		if tok.Type != TokenEOF && tok.Type != TokenSemicolon {
			types = append(types, tok.Type)
		}
	}
	want := []TokenType{
		TokenIdent, TokenColon, TokenIdent, TokenLParen, TokenIdent, TokenComma,
		TokenNumber, TokenRParen, TokenComma, TokenColor, TokenComma, TokenLineThick,
		TokenComma, TokenStick, TokenComma, TokenNoDraw, TokenComma, TokenDotLine,
	}
	if len(types) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(types), types)
	}
	for i, tt := range want {
		if types[i] != tt {
			t.Errorf("token[%d]: expected %s, got %s", i, tt, types[i])
		}
	}
}

func TestLexerLineColumn(t *testing.T) {
	toks, err := NewLexer("A;\nB;").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var bPos Token
	found := false
	for _, tok := range toks {
		if tok.Type == TokenIdent && tok.Value == "B" {
			bPos = tok
			found = true
		}
	}
	if !found {
		t.Fatal("expected identifier B")
	}
	if bPos.Line != 2 {
		t.Errorf("expected B on line 2, got %d", bPos.Line)
	}
	if bPos.Col != 1 {
		t.Errorf("expected B on col 1, got %d", bPos.Col)
	}
}

func TestLexerUnexpectedChar(t *testing.T) {
	_, err := NewLexer("A$B;").Tokenize()
	if err == nil {
		t.Fatal("expected error for unexpected character")
	}
	if !strings.Contains(err.Error(), "unexpected character") {
		t.Errorf("expected unexpected character message, got %v", err)
	}
}

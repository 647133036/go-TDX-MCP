package formula

import "testing"

func TestTokenNew(t *testing.T) {
	tok := NewToken(TokenIdent, "MA", 2, 5)
	if tok.Type != TokenIdent {
		t.Errorf("expected type %s, got %s", TokenIdent, tok.Type)
	}
	if tok.Value != "MA" {
		t.Errorf("expected value %q, got %q", "MA", tok.Value)
	}
	if tok.Line != 2 || tok.Col != 5 {
		t.Errorf("expected position 2:5, got %d:%d", tok.Line, tok.Col)
	}
}

func TestTokenIs(t *testing.T) {
	tok := NewToken(TokenPlus, "+", 1, 1)
	if !tok.Is(TokenPlus) {
		t.Error("expected Is(TokenPlus) to be true")
	}
	if tok.Is(TokenMinus) {
		t.Error("expected Is(TokenMinus) to be false")
	}
}

func TestTokenString(t *testing.T) {
	tok := NewToken(TokenNumber, "12.5", 1, 3)
	want := `Token(NUMBER, "12.5", 1:3)`
	if got := tok.String(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestErrorPosition(t *testing.T) {
	e := NewParseError("unexpected token", 3, 7)
	if e.Kind != KindParse {
		t.Errorf("expected kind %s, got %s", KindParse, e.Kind)
	}
	want := "parse error at line 3, column 7: unexpected token"
	if got := e.Error(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestErrorNoPosition(t *testing.T) {
	e := NewEvalError("division by zero")
	if e.Line != 0 || e.Col != 0 {
		t.Errorf("expected no position fields, got %d:%d", e.Line, e.Col)
	}
	want := "eval error: division by zero"
	if got := e.Error(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

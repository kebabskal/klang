package lexer

import "testing"

func TestBasicTokens(t *testing.T) {
	src := `Entity:class
health :int = 10
cooldown := 5.0`

	lex := New([]byte(src))
	tokens := lex.Tokenize()

	// Find key tokens
	found := map[string]bool{}
	for _, tok := range tokens {
		switch {
		case tok.Type == TOKEN_IDENT && tok.Value == "Entity":
			found["Entity"] = true
		case tok.Type == TOKEN_CLASS:
			found["class"] = true
		case tok.Type == TOKEN_COLON:
			found[":"] = true
		case tok.Type == TOKEN_COLON_EQ:
			found[":="] = true
		case tok.Type == TOKEN_INT_LIT && tok.Value == "10":
			found["10"] = true
		case tok.Type == TOKEN_FLOAT_LIT && tok.Value == "5.0":
			found["5.0"] = true
		}
	}

	for _, key := range []string{"Entity", "class", ":", ":=", "10", "5.0"} {
		if !found[key] {
			t.Errorf("missing expected token: %s", key)
		}
	}
}

func TestOperators(t *testing.T) {
	src := `a?.b ?? c`
	lex := New([]byte(src))
	tokens := lex.Tokenize()

	types := []TokenType{}
	for _, tok := range tokens {
		if tok.Type != TOKEN_EOF {
			types = append(types, tok.Type)
		}
	}

	expected := []TokenType{TOKEN_IDENT, TOKEN_QUESTION_DOT, TOKEN_IDENT, TOKEN_QUESTION_QUESTION, TOKEN_IDENT}
	if len(types) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(types))
	}
	for i := range expected {
		if types[i] != expected[i] {
			t.Errorf("token %d: expected %d, got %d", i, expected[i], types[i])
		}
	}
}

func TestInterpString(t *testing.T) {
	src := `$"{action.name} failed"`
	lex := New([]byte(src))
	tokens := lex.Tokenize()

	found := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_INTERP_STRING {
			found = true
			if tok.Value != "{action.name} failed" {
				t.Errorf("unexpected interp string value: %q", tok.Value)
			}
		}
	}
	if !found {
		t.Error("missing interp string token")
	}
}

func TestEllipsis(t *testing.T) {
	src := `act(...)`
	lex := New([]byte(src))
	tokens := lex.Tokenize()

	found := false
	for _, tok := range tokens {
		if tok.Type == TOKEN_ELLIPSIS {
			found = true
		}
	}
	if !found {
		t.Error("missing ellipsis token")
	}
}

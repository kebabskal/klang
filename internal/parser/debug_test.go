package parser

import (
	"os"
	"testing"

	"github.com/klang-lang/klang/internal/lexer"
)

func TestDebugTokens(t *testing.T) {
	src, err := os.ReadFile("../../example.k")
	if err != nil {
		t.Skip(err)
	}
	lex := lexer.New(src)
	tokens := lex.Tokenize()

	for i, tok := range tokens {
		if tok.Line >= 49 && tok.Line <= 55 {
			t.Logf("[%d] line=%d type=%d value=%q indent=%d", i, tok.Line, tok.Type, tok.Value, tok.Indent)
		}
	}
}

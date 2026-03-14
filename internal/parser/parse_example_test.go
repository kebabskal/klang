package parser

import (
	"os"
	"testing"

	"github.com/klang-lang/klang/internal/lexer"
)

func TestParseExampleK(t *testing.T) {
	src, err := os.ReadFile("../../example.k")
	if err != nil {
		t.Skip("example.k not found:", err)
	}

	lex := lexer.New(src)
	tokens := lex.Tokenize()
	p := New(tokens)
	file, err := p.Parse()
	if err != nil {
		t.Fatal("parse error:", err)
	}

	if file.Namespace != "Game" {
		t.Errorf("expected namespace 'Game', got %q", file.Namespace)
	}
	if len(file.Classes) < 1 {
		t.Fatal("expected at least 1 class")
	}

	entity := file.Classes[0]
	if entity.Name != "Entity" {
		t.Errorf("expected class 'Entity', got %q", entity.Name)
	}

	t.Logf("Entity fields: %d", len(entity.Fields))
	for _, f := range entity.Fields {
		t.Logf("  field: %s (inferred=%v)", f.Name, f.Inferred)
	}

	t.Logf("Entity methods: %d", len(entity.Methods))
	for _, m := range entity.Methods {
		t.Logf("  method: %s (spread=%v, params=%d)", m.Name, m.IsSpread, len(m.Params))
	}

	t.Logf("Entity nested classes: %d", len(entity.Classes))
	for _, c := range entity.Classes {
		t.Logf("  class: %s (parent=%s)", c.Name, c.Parent)
	}

	t.Logf("Entity enums: %d", len(entity.Enums))
	t.Logf("Entity properties: %d", len(entity.Properties))
	t.Logf("Entity events: %d", len(entity.Events))
	if entity.Constructor != nil {
		t.Logf("Entity constructor: %d params", len(entity.Constructor.Params))
	}
}

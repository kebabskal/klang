package parser

import (
	"testing"

	"github.com/klang-lang/klang/internal/lexer"
)

func parse(src string) (*File, error) {
	lex := lexer.New([]byte(src))
	tokens := lex.Tokenize()
	p := New(tokens)
	return p.Parse()
}

func TestParseFileScoped(t *testing.T) {
	src := `Entity:class
health :int = 10
cooldown := 5.0`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(file.Classes))
	}
	cls := file.Classes[0]
	if cls.Name != "Entity" {
		t.Errorf("expected class name 'Entity', got '%s'", cls.Name)
	}
	if len(cls.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(cls.Fields))
	}
}

func TestParseNamespace(t *testing.T) {
	src := `namespace Game
Entity:class
health :int = 10`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if file.Namespace != "Game" {
		t.Errorf("expected namespace 'Game', got '%s'", file.Namespace)
	}
}

func TestParseNestedClass(t *testing.T) {
	src := `Entity:class
	Action:class {
		name:string
	}`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	cls := file.Classes[0]
	if len(cls.Classes) != 1 {
		t.Fatalf("expected 1 nested class, got %d", len(cls.Classes))
	}
	if cls.Classes[0].Name != "Action" {
		t.Errorf("expected nested class 'Action', got '%s'", cls.Classes[0].Name)
	}
}

func TestParseMethod(t *testing.T) {
	src := `Entity:class
	update(delta_time:float) {
		return
	}`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	cls := file.Classes[0]
	if len(cls.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Methods))
	}
	m := cls.Methods[0]
	if m.Name != "update" {
		t.Errorf("expected method name 'update', got '%s'", m.Name)
	}
	if len(m.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(m.Params))
	}
}

func TestParseConstructor(t *testing.T) {
	src := `Entity:class
	constructor(health:int) {
		this.health = health
	}`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	cls := file.Classes[0]
	if cls.Constructor == nil {
		t.Fatal("expected constructor")
	}
	if len(cls.Constructor.Params) != 1 {
		t.Errorf("expected 1 constructor param, got %d", len(cls.Constructor.Params))
	}
}

func TestParseEnum(t *testing.T) {
	src := `Entity:class
	Type:enum = {
		Regular = 0,
		Charged,
	}`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	cls := file.Classes[0]
	if len(cls.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(cls.Enums))
	}
	if len(cls.Enums[0].Members) != 2 {
		t.Errorf("expected 2 enum members, got %d", len(cls.Enums[0].Members))
	}
}

func TestParseSpreadMethod(t *testing.T) {
	src := `Entity:class
	act(...) {
		return
	}`

	file, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	cls := file.Classes[0]
	if len(cls.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cls.Methods))
	}
	if !cls.Methods[0].IsSpread {
		t.Error("expected method to have spread params")
	}
}

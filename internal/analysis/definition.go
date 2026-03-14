package analysis

import (
	"strings"

	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

// DefinitionResult holds the location of a symbol's definition.
type DefinitionResult struct {
	URI    string
	Line   int
	Col    int
	EndCol int
}

// Definition returns the definition location for the symbol at the given position.
func (d *Document) Definition(line, col int) *DefinitionResult {
	if d.AST == nil {
		return nil
	}

	tok := d.TokenAtPosition(line, col)
	if tok == nil || tok.Type != lexer.TOKEN_IDENT {
		return nil
	}
	name := tok.Value

	// Check: is there a dot before this token?
	prevTok := d.findPrevMeaningfulToken(line, col)

	if prevTok != nil && (prevTok.Type == lexer.TOKEN_DOT || prevTok.Type == lexer.TOKEN_QUESTION_DOT) {
		objTok := d.findPrevMeaningfulTokenBefore(prevTok)
		if objTok != nil && objTok.Type == lexer.TOKEN_IDENT {
			return d.definitionMember(objTok.Value, name, line)
		}
		return nil
	}

	return d.definitionBare(name, line)
}

func (d *Document) definitionMember(objName, member string, line int) *DefinitionResult {
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	typeName := d.resolveIdentType(objName, line)
	if typeName == "" {
		// objName might be "this"
		if objName == "this" {
			cls, fullName := d.FindEnclosingClass(line)
			if cls != nil {
				typeName = fullName
			}
		}
	}
	if typeName == "" {
		return nil
	}

	return d.findMemberDefinition(typeName, member, classes)
}

func (d *Document) findMemberDefinition(typeName, member string, classes map[string]*parser.ClassDecl) *DefinitionResult {
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}

	for _, f := range cls.Fields {
		if f.Name == member && f.Pos.Line > 0 {
			return &DefinitionResult{URI: d.URI, Line: f.Pos.Line, Col: f.Pos.Col, EndCol: f.Pos.EndCol}
		}
	}
	for _, m := range cls.Methods {
		if m.Name == member && m.Pos.Line > 0 {
			return &DefinitionResult{URI: d.URI, Line: m.Pos.Line, Col: m.Pos.Col, EndCol: m.Pos.EndCol}
		}
	}
	for _, p := range cls.Properties {
		if p.Name == member && p.Pos.Line > 0 {
			return &DefinitionResult{URI: d.URI, Line: p.Pos.Line, Col: p.Pos.Col, EndCol: p.Pos.EndCol}
		}
	}

	if cls.Parent != "" {
		return d.findMemberDefinition(cls.Parent, member, classes)
	}
	return nil
}

func (d *Document) definitionBare(name string, line int) *DefinitionResult {
	classes := d.GetClasses()

	// Check if it's a class name
	if classes != nil {
		for fullName, cls := range classes {
			if fullName == name || cls.Name == name {
				if cls.Pos.Line > 0 {
					return &DefinitionResult{URI: d.URI, Line: cls.Pos.Line, Col: cls.Pos.Col, EndCol: cls.Pos.EndCol}
				}
			}
		}
		// Also check suffix match (e.g., "Ball" matching "Main_Ball")
		for fullName, cls := range classes {
			if strings.HasSuffix(fullName, "_"+name) {
				if cls.Pos.Line > 0 {
					return &DefinitionResult{URI: d.URI, Line: cls.Pos.Line, Col: cls.Pos.Col, EndCol: cls.Pos.EndCol}
				}
			}
		}
	}

	// Check enclosing class
	cls, _ := d.FindEnclosingClass(line)
	if cls == nil {
		return nil
	}

	method := d.FindEnclosingMethod(cls, line)
	if method != nil {
		// Check locals (reverse order — most recent declaration wins)
		locals := d.CollectLocalsBeforeLine(method, line)
		for i := len(locals) - 1; i >= 0; i-- {
			if locals[i].Name == name && locals[i].Pos.Line > 0 {
				return &DefinitionResult{URI: d.URI, Line: locals[i].Pos.Line, Col: locals[i].Pos.Col, EndCol: locals[i].Pos.EndCol}
			}
		}

		// Check params
		for _, p := range method.Params {
			if p.Name == name && p.Pos.Line > 0 {
				return &DefinitionResult{URI: d.URI, Line: p.Pos.Line, Col: p.Pos.Col, EndCol: p.Pos.EndCol}
			}
		}
	}

	// Check fields
	for _, f := range cls.Fields {
		if f.Name == name && f.Pos.Line > 0 {
			return &DefinitionResult{URI: d.URI, Line: f.Pos.Line, Col: f.Pos.Col, EndCol: f.Pos.EndCol}
		}
	}

	// Check methods
	for _, m := range cls.Methods {
		if m.Name == name && m.Pos.Line > 0 {
			return &DefinitionResult{URI: d.URI, Line: m.Pos.Line, Col: m.Pos.Col, EndCol: m.Pos.EndCol}
		}
	}

	return nil
}

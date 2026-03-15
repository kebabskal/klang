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

	// Check: is there a dot before this token? Walk back to collect the full chain.
	prevTok := d.findPrevMeaningfulToken(line, col)

	if prevTok != nil && (prevTok.Type == lexer.TOKEN_DOT || prevTok.Type == lexer.TOKEN_QUESTION_DOT) {
		// Collect the chain of identifiers: a.b.c.name → chain = [a, b, c]
		var chain []string
		dotTok := prevTok
		for dotTok != nil && (dotTok.Type == lexer.TOKEN_DOT || dotTok.Type == lexer.TOKEN_QUESTION_DOT) {
			identTok := d.findPrevMeaningfulTokenBefore(dotTok)
			if identTok == nil || identTok.Type != lexer.TOKEN_IDENT {
				break
			}
			chain = append([]string{identTok.Value}, chain...)
			dotTok = d.findPrevMeaningfulTokenBefore(identTok)
		}
		if len(chain) > 0 {
			return d.definitionChained(chain, name, line)
		}
		return nil
	}

	return d.definitionBare(name, line)
}

// definitionChained resolves go-to-definition for chained member access: a.b.c.member
// chain contains all identifiers before the final member (e.g., [a, b, c]).
func (d *Document) definitionChained(chain []string, member string, line int) *DefinitionResult {
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	// Resolve the type of the first identifier in the chain
	typeName := d.resolveIdentType(chain[0], line)
	if typeName == "" && chain[0] == "this" {
		_, fullName := d.FindEnclosingClass(line)
		typeName = fullName
	}
	if typeName == "" {
		return nil
	}

	// Walk the rest of the chain, resolving each field's type
	for _, fieldName := range chain[1:] {
		typeName = d.resolveFieldKlangType(typeName, fieldName)
		if typeName == "" {
			return nil
		}
	}

	return d.findMemberDefinition(typeName, member, classes)
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

	uri := d.classURI(cls)
	if f := cls.FindField(member); f != nil && f.Pos.Line > 0 {
		return &DefinitionResult{URI: uri, Line: f.Pos.Line, Col: f.Pos.Col, EndCol: f.Pos.EndCol}
	}
	if m := cls.FindMethod(member); m != nil && m.Pos.Line > 0 {
		return &DefinitionResult{URI: uri, Line: m.Pos.Line, Col: m.Pos.Col, EndCol: m.Pos.EndCol}
	}
	if p := cls.FindProperty(member); p != nil && p.Pos.Line > 0 {
		return &DefinitionResult{URI: uri, Line: p.Pos.Line, Col: p.Pos.Col, EndCol: p.Pos.EndCol}
	}
	if ev := cls.FindEvent(member); ev != nil && ev.Pos.Line > 0 {
		return &DefinitionResult{URI: uri, Line: ev.Pos.Line, Col: ev.Pos.Col, EndCol: ev.Pos.EndCol}
	}

	if cls.Parent != "" {
		return d.findMemberDefinition(cls.Parent, member, classes)
	}
	return nil
}

func (d *Document) definitionBare(name string, line int) *DefinitionResult {
	classes := d.GetClasses()

	// Check if it's a class name (current file first)
	if classes != nil {
		for fullName, cls := range classes {
			if fullName == name || cls.Name == name {
				if cls.Pos.Line > 0 {
					uri := d.classURI(cls)
					return &DefinitionResult{URI: uri, Line: cls.Pos.Line, Col: cls.Pos.Col, EndCol: cls.Pos.EndCol}
				}
			}
		}
		// Also check suffix match (e.g., "Ball" matching "Main_Ball")
		for fullName, cls := range classes {
			if strings.HasSuffix(fullName, "_"+name) {
				if cls.Pos.Line > 0 {
					uri := d.classURI(cls)
					return &DefinitionResult{URI: uri, Line: cls.Pos.Line, Col: cls.Pos.Col, EndCol: cls.Pos.EndCol}
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
	if f := cls.FindField(name); f != nil && f.Pos.Line > 0 {
		return &DefinitionResult{URI: d.URI, Line: f.Pos.Line, Col: f.Pos.Col, EndCol: f.Pos.EndCol}
	}
	// Check methods
	if m := cls.FindMethod(name); m != nil && m.Pos.Line > 0 {
		return &DefinitionResult{URI: d.URI, Line: m.Pos.Line, Col: m.Pos.Col, EndCol: m.Pos.EndCol}
	}
	// Check events
	if ev := cls.FindEvent(name); ev != nil && ev.Pos.Line > 0 {
		return &DefinitionResult{URI: d.URI, Line: ev.Pos.Line, Col: ev.Pos.Col, EndCol: ev.Pos.EndCol}
	}
	// Check properties
	if p := cls.FindProperty(name); p != nil && p.Pos.Line > 0 {
		return &DefinitionResult{URI: d.URI, Line: p.Pos.Line, Col: p.Pos.Col, EndCol: p.Pos.EndCol}
	}

	return nil
}

// classURI returns the file URI where a class is defined.
// Checks the current document first, then sibling files.
func (d *Document) classURI(cls *parser.ClassDecl) string {
	// Check if it's in the current file
	if d.AST != nil {
		if classInFile(cls, d.AST.Classes) {
			return d.URI
		}
	}
	// Check sibling files
	for uri, file := range d.SiblingFiles {
		if classInFile(cls, file.Classes) {
			return uri
		}
	}
	return d.URI
}

// classInFile checks if a ClassDecl pointer exists in a list of classes (by identity).
func classInFile(target *parser.ClassDecl, classes []*parser.ClassDecl) bool {
	for _, cls := range classes {
		if cls == target {
			return true
		}
		if classInFile(target, cls.Classes) {
			return true
		}
	}
	return false
}

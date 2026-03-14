package analysis

import (
	"strings"

	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

// HoverResult holds the hover information for a symbol.
type HoverResult struct {
	Content string // markdown content
	Line    int    // range start
	Col     int
	EndCol  int
}

// Hover returns hover information at the given position (1-based).
func (d *Document) Hover(line, col int) *HoverResult {
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
		// Member access: find the object
		objTok := d.findPrevMeaningfulTokenBefore(prevTok)
		if objTok != nil && objTok.Type == lexer.TOKEN_IDENT {
			return d.hoverMember(objTok.Value, name, line, tok)
		}
		return nil
	}

	// Bare identifier
	return d.hoverBare(name, line, tok)
}

func (d *Document) hoverMember(objName, member string, line int, tok *lexer.Token) *HoverResult {
	// Check modules
	if sigs, ok := StdlibModuleSignatures[objName]; ok {
		for _, sig := range sigs {
			if sig.Name == member {
				return &HoverResult{
					Content: "```klang\n" + objName + "." + sig.Detail + "\n```",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}
	}
	if consts, ok := StdlibModuleConstantNames[objName]; ok {
		for _, c := range consts {
			if c.Name == member {
				return &HoverResult{
					Content: "```klang\n" + objName + "." + c.Name + " — " + c.Detail + "\n```",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}
	}
	if members, ok := StdlibNamespaces[objName]; ok {
		for _, m := range members {
			if m.Name == member {
				return &HoverResult{
					Content: "```klang\n" + objName + "." + m.Name + ":" + m.Detail + "\n```",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}
	}

	// Check "this" member access
	if objName == "this" {
		cls, fullName := d.FindEnclosingClass(line)
		if cls != nil {
			classes := d.GetClasses()
			return d.hoverClassMember(fullName, member, classes, tok)
		}
	}

	// Check class member access
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	typeName := d.resolveIdentType(objName, line)
	if typeName == "" {
		return nil
	}
	return d.hoverClassMember(typeName, member, classes, tok)
}

func (d *Document) hoverClassMember(typeName, member string, classes map[string]*parser.ClassDecl, tok *lexer.Token) *HoverResult {
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}

	for _, f := range cls.Fields {
		if f.Name == member {
			ktype := typeExprToString(f.TypeExpr)
			if ktype == "" {
				ktype = "(inferred)"
			}
			return &HoverResult{
				Content: "```klang\n" + member + ":" + ktype + "\n```\nField on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}
	for _, m := range cls.Methods {
		if m.Name == member {
			sig := formatMethodSignature(m)
			return &HoverResult{
				Content: "```klang\n" + sig + "\n```\nMethod on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}
	for _, p := range cls.Properties {
		if p.Name == member {
			ktype := typeExprToString(p.TypeExpr)
			return &HoverResult{
				Content: "```klang\n" + member + ":" + ktype + "\n```\nProperty on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}

	// Check parent
	if cls.Parent != "" {
		return d.hoverClassMember(cls.Parent, member, classes, tok)
	}
	return nil
}

func (d *Document) hoverBare(name string, line int, tok *lexer.Token) *HoverResult {
	classes := d.GetClasses()

	// Check enclosing class context FIRST (locals/params take priority over class names)
	cls, fullName := d.FindEnclosingClass(line)
	if cls != nil {
		method := d.FindEnclosingMethod(cls, line)
		if method != nil {
			// Check locals (with inferred type resolution)
			locals := d.CollectLocalsBeforeLine(method, line)
			for i := len(locals) - 1; i >= 0; i-- {
				lv := locals[i]
				if lv.Name == name {
					ktype := lv.KType
					if ktype == "" {
						ktype = "(inferred)"
					}
					return &HoverResult{
						Content: "```klang\n" + name + ":" + ktype + "\n```\nLocal variable",
						Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
					}
				}
			}

			// Check params
			for _, p := range method.Params {
				if p.Name == name {
					ktype := typeExprToString(p.TypeExpr)
					if ktype == "" {
						ktype = "(inferred)"
					}
					return &HoverResult{
						Content: "```klang\n" + name + ":" + ktype + "\n```\nParameter",
						Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
					}
				}
			}

			// Check with-module functions
			mods := CollectWithModulesAtLine(method, line)
			for _, mod := range mods {
				if sigs, ok := StdlibModuleSignatures[mod]; ok {
					for _, sig := range sigs {
						if sig.Name == name {
							return &HoverResult{
								Content: "```klang\n" + mod + "." + sig.Detail + "\n```\nvia `with " + mod + "`",
								Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
							}
						}
					}
				}
			}
		}

		// Check fields on current class
		for _, f := range cls.Fields {
			if f.Name == name {
				ktype := d.ResolveFieldType(f, fullName)
				if ktype == "" {
					ktype = "(inferred)"
				}
				return &HoverResult{
					Content: "```klang\n" + name + ":" + ktype + "\n```\nField on `" + cls.Name + "`",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}

		// Check methods on current class
		for _, m := range cls.Methods {
			if m.Name == name {
				sig := formatMethodSignature(m)
				return &HoverResult{
					Content: "```klang\n" + sig + "\n```\nMethod on `" + cls.Name + "`",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}
	}

	// Check if it's a class name
	if classes != nil {
		if foundCls, ok := classes[name]; ok {
			info := name + ":class"
			if foundCls.Parent != "" {
				info += ":" + foundCls.Parent
			}
			return &HoverResult{
				Content: "```klang\n" + info + "\n```",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
		// suffix match
		for fullName, foundCls := range classes {
			if strings.HasSuffix(fullName, "_"+name) || foundCls.Name == name {
				info := name + ":class"
				if foundCls.Parent != "" {
					info += ":" + foundCls.Parent
				}
				return &HoverResult{
					Content: "```klang\n" + info + "\n```",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
		}
	}

	// Check built-in constructors
	builtinCtors := map[string]string{
		"vec2": "vec2(x:float, y:float):vec2",
		"vec3": "vec3(x:float, y:float, z:float):vec3",
		"vec4": "vec4(x:float, y:float, z:float, w:float):vec4",
		"quat": "quat(x:float, y:float, z:float, w:float):quat",
	}
	if sig, ok := builtinCtors[name]; ok {
		return &HoverResult{
			Content: "```klang\n" + sig + "\n```\nBuilt-in constructor",
			Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
		}
	}

	return nil
}

// resolveIdentType resolves the Klang type name of an identifier at the given line.
func (d *Document) resolveIdentType(name string, line int) string {
	cls, fullName := d.FindEnclosingClass(line)
	if cls == nil {
		return ""
	}

	method := d.FindEnclosingMethod(cls, line)
	if method != nil {
		// Use codegen-backed type resolution for locals
		locals := d.CollectLocalsBeforeLine(method, line)
		for i := len(locals) - 1; i >= 0; i-- {
			if locals[i].Name == name && locals[i].KType != "" {
				return locals[i].KType
			}
		}
	}

	// Check fields (use codegen for accurate type resolution)
	for _, f := range cls.Fields {
		if f.Name == name {
			resolved := d.ResolveFieldType(f, fullName)
			if resolved != "" {
				return resolved
			}
			// Fallback to type expression
			if f.TypeExpr != nil {
				return typeExprToString(f.TypeExpr)
			}
		}
	}
	return ""
}

func (d *Document) findClass(name string, classes map[string]*parser.ClassDecl) *parser.ClassDecl {
	if cls, ok := classes[name]; ok {
		return cls
	}
	// Try suffix match
	for fullName, cls := range classes {
		if strings.HasSuffix(fullName, "_"+name) {
			return cls
		}
	}
	return nil
}

func (d *Document) findPrevMeaningfulToken(line, col int) *lexer.Token {
	// Find the token immediately before the one at (line, col)
	tok := d.TokenAtPosition(line, col)
	if tok == nil {
		return nil
	}
	var prev *lexer.Token
	for i := range d.Tokens {
		t := &d.Tokens[i]
		if t.Type == lexer.TOKEN_NEWLINE || t.Type == lexer.TOKEN_INDENT || t.Type == lexer.TOKEN_EOF {
			continue
		}
		if t.Line == tok.Line && t.Col == tok.Col {
			return prev
		}
		prev = t
	}
	return prev
}

func (d *Document) findPrevMeaningfulTokenBefore(target *lexer.Token) *lexer.Token {
	var prev *lexer.Token
	for i := range d.Tokens {
		t := &d.Tokens[i]
		if t.Type == lexer.TOKEN_NEWLINE || t.Type == lexer.TOKEN_INDENT || t.Type == lexer.TOKEN_EOF {
			continue
		}
		if t == target || (t.Line == target.Line && t.Col == target.Col) {
			return prev
		}
		prev = t
	}
	return prev
}

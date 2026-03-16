package analysis

import (
	"strings"

	"github.com/klang-lang/klang/internal/codegen"
	"github.com/klang-lang/klang/internal/errs"
	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

// vendorCTypes is populated lazily by ensureVendorsMerged().
var vendorCTypes = map[string]string{}

// Document represents a single analyzed .k file.
type Document struct {
	URI    string
	Source []byte
	Tokens []lexer.Token
	AST    *parser.File
	Diags      []errs.Diagnostic
	ParseDiags []errs.Diagnostic  // diagnostics from lexing/parsing only (preserved across re-checks)
	Gen    *codegen.Generator // for type resolution
	// Cross-file support: sibling .k files in the same project
	SiblingFiles map[string]*parser.File // URI -> parsed AST of sibling files
}

// ClassInfo holds resolved information about a class.
type ClassInfo struct {
	Decl     *parser.ClassDecl
	FullName string // prefix_Name for nested
	Fields   []FieldInfo
	Methods  []MethodInfo
}

// FieldInfo holds resolved field information.
type FieldInfo struct {
	Name  string
	KType string // Klang-level display type
	Pos   parser.Pos
}

// MethodInfo holds resolved method information.
type MethodInfo struct {
	Name       string
	Params     []ParamInfo
	ReturnType string // Klang-level display type
	Pos        parser.Pos
}

// ParamInfo holds parameter information.
type ParamInfo struct {
	Name  string
	KType string
}

// addError appends an error diagnostic at the given position.
func (d *Document) addError(pos parser.Pos, message string) {
	d.Diags = append(d.Diags, errs.Diagnostic{
		File:    d.URI,
		Line:    pos.Line,
		Col:     pos.Col,
		EndCol:  pos.EndCol,
		Kind:    errs.Error,
		Message: message,
		Source:  errs.GetSourceLine(d.Source, pos.Line),
	})
}

// Analyze parses and analyzes a source file, returning a Document with diagnostics and symbol info.
func Analyze(uri string, src []byte) *Document {
	doc := &Document{URI: uri, Source: src}

	// Lex
	l := lexer.New(src)
	tokens := l.Tokenize()
	doc.Tokens = tokens

	// Convert lex errors to diagnostics
	for _, le := range l.Errors() {
		endCol := le.Col + 1
		doc.Diags = append(doc.Diags, errs.Diagnostic{
			File:    uri,
			Line:    le.Line,
			Col:     le.Col,
			EndCol:  endCol,
			Kind:    errs.Error,
			Message: le.Message,
			Source:  errs.GetSourceLine(src, le.Line),
		})
	}

	// Parse
	p := parser.New(tokens)
	p.SetSource(src, uri)
	file, _ := p.Parse()
	doc.AST = file
	doc.Diags = append(doc.Diags, p.Diagnostics()...)

	// Preserve parse-phase diagnostics (before semantic checks add more)
	doc.ParseDiags = append([]errs.Diagnostic{}, doc.Diags...)

	// Create codegen Generator for type resolution (doesn't generate code)
	if file != nil {
		doc.Gen = codegen.New(file)
	}

	// Semantic checks (only if no parse errors)
	if file != nil && len(doc.Diags) == 0 {
		doc.Check()
	}

	return doc
}

// AddSiblingFile registers a sibling .k file for cross-file definition lookups.
func (d *Document) AddSiblingFile(uri string, file *parser.File) {
	if d.SiblingFiles == nil {
		d.SiblingFiles = make(map[string]*parser.File)
	}
	d.SiblingFiles[uri] = file
}

// GetClasses returns all registered classes from the codegen.
func (d *Document) GetClasses() map[string]*parser.ClassDecl {
	if d.Gen == nil {
		return nil
	}
	return d.Gen.GetClasses()
}

// ResolveFieldType resolves the display type string for a field.
func (d *Document) ResolveFieldType(field *parser.FieldDecl, className string) string {
	if d.Gen == nil {
		return ""
	}
	cType := d.Gen.FieldCType(field, className)
	return CTypeToKlang(cType)
}

// ResolveParamType resolves the display type string for a parameter.
func (d *Document) ResolveParamType(param *parser.Param, className string) string {
	if param.TypeExpr == nil {
		return ""
	}
	if d.Gen == nil {
		return ""
	}
	cType := d.Gen.TypeToC(param.TypeExpr, className)
	return CTypeToKlang(cType)
}

// ResolveReturnType resolves the display type for a method return.
func (d *Document) ResolveReturnType(method *parser.MethodDecl, className string) string {
	if method.ReturnType == nil {
		return ""
	}
	if d.Gen == nil {
		return ""
	}
	cType := d.Gen.TypeToC(method.ReturnType, className)
	return CTypeToKlang(cType)
}

// ResolveExprType resolves the display type for an expression.
func (d *Document) ResolveExprType(expr parser.Expr) string {
	if d.Gen == nil {
		return ""
	}
	cType := d.Gen.InferCType(expr)
	return CTypeToKlang(cType)
}

// CTypeToKlang converts a C type string back to Klang display type.
func CTypeToKlang(cType string) string {
	ensureVendorsMerged()
	switch cType {
	case "int":
		return "int"
	case "float":
		return "float"
	case "bool":
		return "bool"
	case "const char*":
		return "string"
	case "void":
		return "void"
	case "KlList*":
		return "List"
	case "KlDict*":
		return "Dictionary"
	case "KlClosure*":
		return "fn"
	case "vec2":
		return "vec2"
	case "vec3":
		return "vec3"
	case "vec4":
		return "vec4"
	case "mat4":
		return "mat4"
	case "quat":
		return "quat"
	}
	// Check vendor-contributed type mappings
	if klType, ok := vendorCTypes[cType]; ok {
		return klType
	}
	// ClassName* → ClassName
	if strings.HasSuffix(cType, "*") {
		return strings.TrimSuffix(cType, "*")
	}
	return cType
}

// TokenAtPosition finds the token at the given line/col (1-based).
func (d *Document) TokenAtPosition(line, col int) *lexer.Token {
	for i := range d.Tokens {
		t := &d.Tokens[i]
		if t.Line == line && col >= t.Col && col < t.Col+len(t.Value) {
			return t
		}
	}
	return nil
}

// TokenBeforePosition finds the last meaningful token before or at the given position.
func (d *Document) TokenBeforePosition(line, col int) *lexer.Token {
	var best *lexer.Token
	for i := range d.Tokens {
		t := &d.Tokens[i]
		if t.Type == lexer.TOKEN_NEWLINE || t.Type == lexer.TOKEN_INDENT || t.Type == lexer.TOKEN_EOF {
			continue
		}
		if t.Line < line || (t.Line == line && t.Col <= col) {
			best = t
		}
	}
	return best
}

// FindEnclosingClass returns the class and its full name that contains the given position.
func (d *Document) FindEnclosingClass(line int) (cls *parser.ClassDecl, fullName string) {
	if d.AST == nil {
		return nil, ""
	}
	for _, c := range d.AST.Classes {
		if found, name := findClassAtLine("", c, line); found != nil {
			cls = found
			fullName = name
		}
	}
	return
}

func findClassAtLine(prefix string, cls *parser.ClassDecl, line int) (*parser.ClassDecl, string) {
	fullName := cls.Name
	if prefix != "" {
		fullName = prefix + "_" + cls.Name
	}

	// For braced classes, check if line is within the class bounds
	if cls.EndLine > 0 && line > cls.EndLine {
		return nil, "" // line is past this class's closing brace
	}

	// Check nested classes first (more specific)
	for _, nested := range cls.Classes {
		if found, name := findClassAtLine(fullName, nested, line); found != nil {
			return found, name
		}
	}
	// Check if any method in this class contains the line
	for _, m := range cls.Methods {
		if m.Pos.Line > 0 && m.Pos.Line <= line {
			return cls, fullName
		}
	}
	// Fallback: if this class has a position and the line is after it
	if cls.Pos.Line > 0 && cls.Pos.Line <= line {
		return cls, fullName
	}
	return nil, ""
}

// FindEnclosingMethod returns the method at the given line.
func (d *Document) FindEnclosingMethod(cls *parser.ClassDecl, line int) *parser.MethodDecl {
	if cls == nil {
		return nil
	}
	var best *parser.MethodDecl
	for _, m := range cls.Methods {
		if m.Pos.Line > 0 && m.Pos.Line <= line {
			best = m
		}
	}
	return best
}

// CollectLocalsBeforeLine walks a method body and collects variable declarations before the given line.
// Uses the Document's codegen to resolve inferred types.
func (d *Document) CollectLocalsBeforeLine(method *parser.MethodDecl, line int) []LocalVar {
	var locals []LocalVar
	// Add parameters
	for _, p := range method.Params {
		locals = append(locals, LocalVar{
			Name:  p.Name,
			KType: typeExprToString(p.TypeExpr),
			Pos:   p.Pos,
		})
	}
	// Walk body
	if method.Body != nil {
		d.collectLocalsFromBlock(method.Body, line, &locals)
	}
	return locals
}

// LocalVar represents a local variable in scope.
type LocalVar struct {
	Name  string
	KType string
	Pos   parser.Pos
}

func (d *Document) collectLocalsFromBlock(block *parser.Block, line int, locals *[]LocalVar) {
	for _, stmt := range block.Stmts {
		switch s := stmt.(type) {
		case *parser.VarDecl:
			if s.Pos.Line > 0 && s.Pos.Line <= line {
				ktype := typeExprToString(s.TypeExpr)
				// For inferred types (:=), use codegen to resolve the type from the value
				if ktype == "" && s.Value != nil && d.Gen != nil {
					// Try Klang-level inference first (preserves generics like Stack<string>)
					ktype = d.Gen.InferKlangType(s.Value)
					if ktype == "" {
						cType := d.Gen.InferCType(s.Value)
						ktype = CTypeToKlang(cType)
					}
				}
				*locals = append(*locals, LocalVar{Name: s.Name, KType: ktype, Pos: s.Pos})
			}
		case *parser.ForStmt:
			if s.Pos.Line > 0 && s.Pos.Line <= line && s.VarName != "" {
				if s.ValueVar != "" {
					keyType, valType := d.inferForDictTypes(s.Iterable)
					*locals = append(*locals, LocalVar{Name: s.VarName, KType: keyType, Pos: s.Pos})
					*locals = append(*locals, LocalVar{Name: s.ValueVar, KType: valType, Pos: s.Pos})
				} else {
					ktype := d.inferForVarType(s.Iterable)
					*locals = append(*locals, LocalVar{Name: s.VarName, KType: ktype, Pos: s.Pos})
				}
			}
			if s.Body != nil {
				d.collectLocalsFromBlock(s.Body, line, locals)
			}
		case *parser.IfStmt:
			if s.Then != nil {
				d.collectLocalsFromBlock(s.Then, line, locals)
			}
			if blk, ok := s.Else.(*parser.Block); ok {
				d.collectLocalsFromBlock(blk, line, locals)
			}
		case *parser.WhileStmt:
			if s.Body != nil {
				d.collectLocalsFromBlock(s.Body, line, locals)
			}
		case *parser.WithStmt:
			if s.Body != nil {
				d.collectLocalsFromBlock(s.Body, line, locals)
			}
		}
	}
}

// inferForDictTypes resolves the key and value types for a for-loop over a Dictionary.
func (d *Document) inferForDictTypes(iterable parser.Expr) (string, string) {
	if iterable == nil {
		return "", ""
	}
	ident, ok := iterable.(*parser.Ident)
	if !ok {
		return "", ""
	}

	// Check local variable type expressions in the AST
	if d.AST != nil {
		for _, cls := range d.AST.Classes {
			for _, m := range cls.Methods {
				if m.Body != nil {
					if k, v := d.findDictTypeInBlock(ident.Name, m.Body); k != "" {
						return k, v
					}
				}
			}
		}
	}

	// Check class fields
	classes := d.GetClasses()
	if classes == nil {
		return "", ""
	}
	for _, cls := range classes {
		for _, f := range cls.Fields {
			if f.Name == ident.Name {
				if gt, ok := f.TypeExpr.(*parser.GenericType); ok && gt.Name == "Dictionary" && len(gt.TypeArgs) >= 2 {
					return typeExprToString(gt.TypeArgs[0]), typeExprToString(gt.TypeArgs[1])
				}
			}
		}
	}
	return "", ""
}

// findDictTypeInBlock searches for a local VarDecl with Dictionary<K,V> type.
func (d *Document) findDictTypeInBlock(name string, block *parser.Block) (string, string) {
	for _, stmt := range block.Stmts {
		if vd, ok := stmt.(*parser.VarDecl); ok && vd.Name == name {
			if gt, ok := vd.TypeExpr.(*parser.GenericType); ok && gt.Name == "Dictionary" && len(gt.TypeArgs) >= 2 {
				return typeExprToString(gt.TypeArgs[0]), typeExprToString(gt.TypeArgs[1])
			}
		}
	}
	return "", ""
}

// inferForVarType resolves the element type for a for-loop iterable expression.
// e.g., for "for ball in balls" where balls:List<Ball>, returns "Ball".
func (d *Document) inferForVarType(iterable parser.Expr) string {
	if iterable == nil {
		return ""
	}

	// Range expression: for i in start..end → int
	if _, ok := iterable.(*parser.RangeExpr); ok {
		return "int"
	}

	// Handle dict.keys() and dict.values() calls
	if call, ok := iterable.(*parser.CallExpr); ok {
		if mem, ok := call.Callee.(*parser.MemberExpr); ok {
			if mem.Field == "keys" || mem.Field == "values" {
				k, v := d.inferForDictTypes(mem.Object)
				if mem.Field == "keys" && k != "" {
					return k
				}
				if mem.Field == "values" && v != "" {
					return v
				}
			}
		}
	}

	// If iterable is an identifier, look up its type
	ident, ok := iterable.(*parser.Ident)
	if !ok {
		return ""
	}

	// Check local variable type expressions in the AST
	if d.AST != nil {
		for _, cls := range d.AST.Classes {
			for _, m := range cls.Methods {
				if m.Body != nil {
					if t := d.findListTypeInBlock(ident.Name, m.Body); t != "" {
						return t
					}
				}
			}
		}
	}

	// Find the enclosing class and check fields
	classes := d.GetClasses()
	if classes == nil {
		return ""
	}

	// Check all classes for a field with this name that has List<T> type
	for fullName, cls := range classes {
		for _, f := range cls.Fields {
			if f.Name == ident.Name {
				// Check if it's List<T>
				if gt, ok := f.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
					return typeExprToString(gt.TypeArgs[0])
				}
				// Fallback: resolve via codegen
				cType := d.ResolveFieldType(f, fullName)
				if cType == "List" {
					// Can't determine element type
					return ""
				}
				return ""
			}
		}
	}

	return ""
}

// findListTypeInBlock searches for a local VarDecl with List<T> type.
func (d *Document) findListTypeInBlock(name string, block *parser.Block) string {
	for _, stmt := range block.Stmts {
		if vd, ok := stmt.(*parser.VarDecl); ok && vd.Name == name {
			if gt, ok := vd.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
				return typeExprToString(gt.TypeArgs[0])
			}
		}
	}
	return ""
}

// CollectWithModulesAtLine returns which modules are in scope via "with" at the given line.
func CollectWithModulesAtLine(method *parser.MethodDecl, line int) []string {
	if method == nil || method.Body == nil {
		return nil
	}
	var mods []string
	collectWithFromBlock(method.Body, line, &mods)
	return mods
}

func collectWithFromBlock(block *parser.Block, line int, mods *[]string) {
	for _, stmt := range block.Stmts {
		switch s := stmt.(type) {
		case *parser.WithStmt:
			if s.Pos.Line > 0 && s.Pos.Line <= line {
				*mods = append(*mods, s.Module)
			}
			// Recurse into the with body for nested with statements
			if s.Body != nil {
				collectWithFromBlock(s.Body, line, mods)
			}
		case *parser.IfStmt:
			if s.Then != nil {
				collectWithFromBlock(s.Then, line, mods)
			}
			if blk, ok := s.Else.(*parser.Block); ok {
				collectWithFromBlock(blk, line, mods)
			}
		case *parser.WhileStmt:
			if s.Body != nil {
				collectWithFromBlock(s.Body, line, mods)
			}
		case *parser.ForStmt:
			if s.Body != nil {
				collectWithFromBlock(s.Body, line, mods)
			}
		}
	}
}

func typeExprToString(t parser.TypeExpr) string {
	if t == nil {
		return ""
	}
	switch te := t.(type) {
	case *parser.SimpleType:
		return te.Name
	case *parser.GenericType:
		args := make([]string, len(te.TypeArgs))
		for i, a := range te.TypeArgs {
			args[i] = typeExprToString(a)
		}
		return te.Name + "<" + strings.Join(args, ", ") + ">"
	case *parser.FnType:
		params := make([]string, len(te.ParamTypes))
		for i, p := range te.ParamTypes {
			params[i] = typeExprToString(p)
		}
		ret := ""
		if te.ReturnType != nil {
			ret = ":" + typeExprToString(te.ReturnType)
		}
		return "fn(" + strings.Join(params, ", ") + ")" + ret
	case *parser.UnionType:
		types := make([]string, len(te.Types))
		for i, t := range te.Types {
			types[i] = typeExprToString(t)
		}
		return strings.Join(types, "|")
	}
	return ""
}

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

// hoverMarkdown formats a klang code block with optional context text.
func hoverMarkdown(signature, context string) string {
	md := "```klang\n" + signature + "\n```"
	if context != "" {
		md += "\n" + context
	}
	return md
}

// makeHover creates a HoverResult with formatted markdown content.
func makeHover(signature, context string, tok *lexer.Token) *HoverResult {
	return &HoverResult{
		Content: hoverMarkdown(signature, context),
		Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
	}
}

// Hover returns hover information at the given position (1-based).
func (d *Document) Hover(line, col int) *HoverResult {
	ensureVendorsMerged()
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
		// Collect the full chain of identifiers: a.b.c.name → chain = [a, b, c]
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
			return d.hoverChained(chain, name, line, tok)
		}
		return nil
	}

	// Bare identifier
	return d.hoverBare(name, line, tok)
}

func (d *Document) hoverChained(chain []string, member string, line int, tok *lexer.Token) *HoverResult {
	// Simple case: single-element chain (e.g., obj.member)
	objName := chain[0]

	if len(chain) == 1 {
		// Check modules
		if sigs, ok := StdlibModuleSignatures[objName]; ok {
			for _, sig := range sigs {
				if sig.Name == member {
					return makeHover(objName+"."+sig.Detail, "", tok)
				}
			}
		}
		if consts, ok := StdlibModuleConstantNames[objName]; ok {
			for _, c := range consts {
				if c.Name == member {
					return makeHover(objName+"."+c.Name+" — "+c.Detail, "", tok)
				}
			}
		}
		if members, ok := StdlibNamespaces[objName]; ok {
			for _, m := range members {
				if m.Name == member {
					return makeHover(objName+"."+m.Name+":"+m.Detail, "", tok)
				}
			}
		}

		// Check event member access (bare event name: event_name.emit/connect)
		cls, _ := d.FindEnclosingClass(line)
		if cls != nil {
			if ev := cls.FindEvent(objName); ev != nil {
				return d.hoverEventMember(ev, member, tok)
			}
		}
	}

	// Resolve the type through the chain
	chainStr := strings.Join(chain, ".")
	typeName := d.resolveChainedType(chainStr, line)
	if typeName == "" {
		// Try resolving the first element as a lambda parameter
		cls, _ := d.FindEnclosingClass(line)
		if cls != nil {
			method := d.FindEnclosingMethod(cls, line)
			if method != nil {
				typeName = d.resolveLambdaParamType(chain[0], method, line)
				// Walk the rest of the chain
				for _, fieldName := range chain[1:] {
					if typeName == "" {
						break
					}
					typeName = d.resolveFieldKlangType(typeName, fieldName)
				}
			}
		}
	}
	if typeName == "" {
		return nil
	}

	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	// Check if member is an event on the resolved type
	if resolvedCls := d.findClass(typeName, classes); resolvedCls != nil {
		if ev := resolvedCls.FindEvent(member); ev != nil {
			return makeHover(member+":event("+formatEventParams(ev)+")", "Event on `"+typeName+"`", tok)
		}
	}

	// Check if member is an event method (e.g., entity.health.after_change.emit)
	// The "member" might be emit/connect/disconnect on an event found in the chain
	// This is handled by checking if the last chain element is an event

	// Check user-defined class members
	if result := d.hoverClassMember(typeName, member, classes, tok); result != nil {
		return result
	}

	// Check built-in type members (List, Dictionary, vec2, etc.)
	return d.hoverBuiltinMember(typeName, member, tok)
}

func (d *Document) hoverClassMember(typeName, member string, classes map[string]*parser.ClassDecl, tok *lexer.Token) *HoverResult {
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}

	if f := cls.FindField(member); f != nil {
		ktype := typeExprToString(f.TypeExpr)
		if ktype == "" {
			ktype = "(inferred)"
		}
		ktype = resolveTypeWithParams(typeName, cls, ktype)
		return makeHover(member+":"+ktype, "Field on `"+typeName+"`", tok)
	}
	if m := cls.FindMethod(member); m != nil {
		sig := resolveTypeWithParams(typeName, cls, formatMethodSignature(m))
		return makeHover(sig, "Method on `"+typeName+"`", tok)
	}
	if p := cls.FindProperty(member); p != nil {
		ktype := resolveTypeWithParams(typeName, cls, typeExprToString(p.TypeExpr))
		return makeHover(member+":"+ktype, "Property on `"+typeName+"`", tok)
	}
	if ev := cls.FindEvent(member); ev != nil {
		return makeHover(member+":event("+formatEventParams(ev)+")", "Event on `"+typeName+"`", tok)
	}

	// Check parent
	if cls.Parent != "" {
		return d.hoverClassMember(cls.Parent, member, classes, tok)
	}
	return nil
}

// buildTypeParamSub builds a substitution map from a concrete type name and its class declaration.
// e.g. typeName="Stack<string>", cls.TypeParams=["T"] → map{"T": "string"}
// e.g. typeName="Pair<string, int>", cls.TypeParams=["A","B"] → map{"A": "string", "B": "int"}
func buildTypeParamSub(typeName string, cls *parser.ClassDecl) map[string]string {
	if len(cls.TypeParams) == 0 {
		return nil
	}
	idx := strings.Index(typeName, "<")
	if idx < 0 {
		return nil
	}
	inner := typeName[idx+1 : len(typeName)-1]
	args := splitTypeArgs(inner)
	if len(args) != len(cls.TypeParams) {
		return nil
	}
	sub := make(map[string]string, len(cls.TypeParams))
	for i, tp := range cls.TypeParams {
		sub[tp] = args[i]
	}
	return sub
}

// applyTypeParamSub replaces type parameter names in a signature string with concrete types.
// It matches type params at type positions: after ':', after ')', in '<>', and standalone.
func applyTypeParamSub(sig string, sub map[string]string) string {
	if len(sub) == 0 {
		return sig
	}
	var oldnew []string
	for tp, concrete := range sub {
		// Replace in various positions where type names appear
		oldnew = append(oldnew,
			":"+tp, ":"+concrete, // param types: "item:T" → "item:string"
			"("+tp+")", "("+concrete+")", // wrapped: "(T)" → "(string)"
			"<"+tp+">", "<"+concrete+">", // generic args: "<T>" → "<string>"
			"):"+tp, "):"+concrete, // return types: "):T" → "):string"
			", "+tp+",", ", "+concrete+",", // in lists
			", "+tp+")", ", "+concrete+")", // last in list
		)
	}
	return strings.NewReplacer(oldnew...).Replace(sig)
}

// resolveTypeWithParams combines building and applying type parameter substitution.
func resolveTypeWithParams(typeName string, cls *parser.ClassDecl, ktype string) string {
	return applyTypeParamSub(ktype, buildTypeParamSub(typeName, cls))
}

func (d *Document) hoverBuiltinMember(typeName, member string, tok *lexer.Token) *HoverResult {
	// Try exact match first, then generic base name
	lookups := []string{typeName}
	if idx := strings.Index(typeName, "<"); idx > 0 {
		lookups = append(lookups, typeName[:idx])
	}
	for _, name := range lookups {
		if members, ok := BuiltinTypeMembers[name]; ok {
			for _, m := range members {
				if m.Label == member {
					detail := d.expandGenericPlaceholders(typeName, m.Detail)
					return makeHover(detail, "Method on `"+typeName+"`", tok)
				}
			}
		}
	}
	return nil
}

// expandGenericPlaceholders replaces K, V, T etc. in a detail string with the
// actual type arguments from the concrete type name.
// e.g. typeName="Dictionary<string, int>", detail="get(key:K):V" → "get(key:string):int"
// e.g. typeName="List<Ball>", detail="append(item:T)" → "append(item:Ball)"
func (d *Document) expandGenericPlaceholders(typeName, detail string) string {
	idx := strings.Index(typeName, "<")
	if idx < 0 {
		return detail
	}
	baseName := typeName[:idx]
	inner := typeName[idx+1 : len(typeName)-1] // strip < and >
	args := splitTypeArgs(inner)

	switch baseName {
	case "Dictionary":
		if len(args) >= 2 {
			r := strings.NewReplacer(":K", ":"+args[0], "(K)", "("+args[0]+")",
				":V", ":"+args[1], "):V", "):"+args[1],
				"<K>", "<"+args[0]+">", "<V>", "<"+args[1]+">")
			return r.Replace(detail)
		}
	case "List":
		if len(args) >= 1 {
			r := strings.NewReplacer(":T", ":"+args[0], "(T)", "("+args[0]+")",
				"):T", "):"+args[0], "<T>", "<"+args[0]+">")
			return r.Replace(detail)
		}
	}
	return detail
}

// splitTypeArgs splits "string, int" into ["string", "int"], respecting nested <>.
func splitTypeArgs(s string) []string {
	var args []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	args = append(args, strings.TrimSpace(s[start:]))
	return args
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
					return makeHover(name+":"+ktype, "Local variable", tok)
				}
			}

			// Check params
			for _, p := range method.Params {
				if p.Name == name {
					ktype := typeExprToString(p.TypeExpr)
					if ktype == "" {
						ktype = "(inferred)"
					}
					return makeHover(name+":"+ktype, "Parameter", tok)
				}
			}

			// Check lambda parameters (e.g., (info) => ... in event.connect)
			if result := d.hoverLambdaParam(name, method, line, tok); result != nil {
				return result
			}

			// Check with-module functions
			mods := CollectWithModulesAtLine(method, line)
			for _, mod := range mods {
				if sigs, ok := StdlibModuleSignatures[mod]; ok {
					for _, sig := range sigs {
						if sig.Name == name {
							return makeHover(mod+"."+sig.Detail, "via `with "+mod+"`", tok)
						}
					}
				}
			}
		}

		// Check fields on current class
		if f := cls.FindField(name); f != nil {
			ktype := d.ResolveFieldType(f, fullName)
			if ktype == "" {
				ktype = "(inferred)"
			}
			return makeHover(name+":"+ktype, "Field on `"+cls.Name+"`", tok)
		}

		// Check methods on current class
		if m := cls.FindMethod(name); m != nil {
			return makeHover(formatMethodSignature(m), "Method on `"+cls.Name+"`", tok)
		}

		// Check events on current class
		if ev := cls.FindEvent(name); ev != nil {
			return makeHover(name+":event("+formatEventParams(ev)+")", "Event on `"+cls.Name+"`", tok)
		}
	}

	// Check if it's a class name
	if classes != nil {
		if foundCls, ok := classes[name]; ok {
			info := name + ":class"
			if foundCls.Parent != "" {
				info += ":" + foundCls.Parent
			}
			return makeHover(info, "", tok)
		}
		// suffix match
		for fullName, foundCls := range classes {
			if strings.HasSuffix(fullName, "_"+name) || foundCls.Name == name {
				info := name + ":class"
				if foundCls.Parent != "" {
					info += ":" + foundCls.Parent
				}
				return makeHover(info, "", tok)
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
		return makeHover(sig, "Built-in constructor", tok)
	}

	return nil
}

func (d *Document) hoverEventMember(ev *parser.EventDecl, member string, tok *lexer.Token) *HoverResult {
	paramStr := formatEventParams(ev)
	switch member {
	case "connect":
		handlerSig := "(" + paramStr + ") => void"
		return makeHover(ev.Name+".connect(handler: "+handlerSig+")", "Connect a handler to this event", tok)
	case "disconnect":
		return makeHover(ev.Name+".disconnect(handler)", "Disconnect a handler from this event", tok)
	case "emit":
		sig := formatEventSignature(ev)
		return makeHover(ev.Name+"."+sig, "Emit this event, calling all connected handlers", tok)
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
	if f := cls.FindField(name); f != nil {
		resolved := d.ResolveFieldType(f, fullName)
		if resolved != "" {
			return resolved
		}
		if f.TypeExpr != nil {
			return typeExprToString(f.TypeExpr)
		}
	}
	return ""
}

func (d *Document) findClass(name string, classes map[string]*parser.ClassDecl) *parser.ClassDecl {
	if cls, ok := classes[name]; ok {
		return cls
	}
	// Strip generic type args: "Stack<string>" → "Stack"
	if idx := strings.Index(name, "<"); idx > 0 {
		baseName := name[:idx]
		if cls, ok := classes[baseName]; ok {
			return cls
		}
		for fullName, cls := range classes {
			if strings.HasSuffix(fullName, "_"+baseName) {
				return cls
			}
		}
	}
	// Try suffix match
	for fullName, cls := range classes {
		if strings.HasSuffix(fullName, "_"+name) {
			return cls
		}
	}
	return nil
}

// lambdaParamMatch holds the result of finding a lambda parameter in the AST.
type lambdaParamMatch struct {
	Param      *parser.Param
	Index      int
	ParamHints []*parser.Param
}

// findLambdaParam walks the method body to find a lambda parameter by name.
// Returns the matched param, its index, and the call context's parameter hints.
func (d *Document) findLambdaParam(name string, method *parser.MethodDecl, line int) *lambdaParamMatch {
	if method.Body == nil {
		return nil
	}
	return d.findLambdaParamInBlock(name, method.Body, line)
}

func (d *Document) findLambdaParamInBlock(name string, block *parser.Block, line int) *lambdaParamMatch {
	for _, stmt := range block.Stmts {
		if m := d.findLambdaParamInStmt(name, stmt, line); m != nil {
			return m
		}
	}
	return nil
}

func (d *Document) findLambdaParamInStmt(name string, stmt parser.Stmt, line int) *lambdaParamMatch {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		return d.findLambdaParamInExpr(name, s.Expr, line, nil)
	case *parser.VarDecl:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, nil)
		}
	case *parser.AssignStmt:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, nil)
		}
	case *parser.ForStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line)
		}
	case *parser.IfStmt:
		if s.Then != nil {
			if m := d.findLambdaParamInBlock(name, s.Then, line); m != nil {
				return m
			}
		}
		if blk, ok := s.Else.(*parser.Block); ok {
			if m := d.findLambdaParamInBlock(name, blk, line); m != nil {
				return m
			}
		}
	case *parser.WhileStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line)
		}
	case *parser.WithStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, nil)
		}
	}
	return nil
}

func (d *Document) findLambdaParamInExpr(name string, expr parser.Expr, line int, paramHints []*parser.Param) *lambdaParamMatch {
	switch e := expr.(type) {
	case *parser.CallExpr:
		hints := d.resolveCallParamHints(e)
		for i, arg := range e.Args {
			var argHints []*parser.Param
			if hints != nil && i < len(hints) {
				argHints = hints
			}
			if m := d.findLambdaParamInExpr(name, arg, line, argHints); m != nil {
				return m
			}
		}
	case *parser.LambdaExpr:
		lambdaStartLine := 0
		if len(e.Params) > 0 && e.Params[0].Pos.Line > 0 {
			lambdaStartLine = e.Params[0].Pos.Line
		}
		if lambdaStartLine > 0 && line >= lambdaStartLine {
			for i, p := range e.Params {
				if p.Name == name {
					return &lambdaParamMatch{Param: p, Index: i, ParamHints: paramHints}
				}
			}
			if e.Body != nil {
				return d.findLambdaParamInBlock(name, e.Body, line)
			}
		}
	case *parser.MemberExpr:
		return d.findLambdaParamInExpr(name, e.Object, line, nil)
	case *parser.BinaryExpr:
		if m := d.findLambdaParamInExpr(name, e.Left, line, nil); m != nil {
			return m
		}
		return d.findLambdaParamInExpr(name, e.Right, line, nil)
	case *parser.UnaryExpr:
		return d.findLambdaParamInExpr(name, e.Operand, line, nil)
	}
	return nil
}

// resolveMatchType extracts the type string from a lambda param match.
func resolveMatchType(m *lambdaParamMatch) string {
	ktype := typeExprToString(m.Param.TypeExpr)
	if ktype == "" && m.ParamHints != nil && m.Index < len(m.ParamHints) {
		ktype = typeExprToString(m.ParamHints[m.Index].TypeExpr)
	}
	return ktype
}

// hoverLambdaParam checks if the cursor is on a lambda parameter and resolves its type.
func (d *Document) hoverLambdaParam(name string, method *parser.MethodDecl, line int, tok *lexer.Token) *HoverResult {
	m := d.findLambdaParam(name, method, line)
	if m == nil {
		return nil
	}
	ktype := resolveMatchType(m)
	if ktype == "" {
		ktype = "(inferred)"
	}
	return makeHover(name+":"+ktype, "Lambda parameter", tok)
}

// resolveCallParamHints resolves the parameter types for a call expression,
// so we can provide type hints for lambda arguments.
func (d *Document) resolveCallParamHints(call *parser.CallExpr) []*parser.Param {
	switch callee := call.Callee.(type) {
	case *parser.MemberExpr:
		methodName := callee.Field
		// Resolve the object type
		typeName := d.resolveExprKlangType(callee.Object)
		if typeName == "" {
			return nil
		}
		classes := d.GetClasses()
		if classes == nil {
			return nil
		}
		cls := d.findClass(typeName, classes)
		if cls == nil {
			return nil
		}
		// Check if it's an event (connect pattern: event(handler) or event.connect(handler))
		if ev := cls.FindEvent(methodName); ev != nil {
			return ev.Params
		}
		// Check methods
		if m := cls.FindMethod(methodName); m != nil {
			return m.Params
		}
		if cls.Parent != "" {
			if parentCls := d.findClass(cls.Parent, classes); parentCls != nil {
				if m := parentCls.FindMethod(methodName); m != nil {
					return m.Params
				}
			}
		}
	case *parser.Ident:
		// Could be a bare event name used as connect: event_name(handler)
		// Or a method call
	}
	return nil
}

// resolveExprKlangType resolves the Klang type of an expression.
func (d *Document) resolveExprKlangType(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.Ident:
		// Look up all possible lines — use 0 as a fallback to search broadly
		// We need the line context, but the ident has position info
		if e.Pos.Line > 0 {
			return d.resolveIdentType(e.Name, e.Pos.Line)
		}
		return ""
	case *parser.MemberExpr:
		objType := d.resolveExprKlangType(e.Object)
		if objType == "" {
			return ""
		}
		return d.resolveFieldKlangType(objType, e.Field)
	case *parser.ThisExpr:
		// ThisExpr has no position — caller should use resolveChainedType for "this" chains
		return ""
	}
	return ""
}

// resolveLambdaParamType walks the method body to find a lambda parameter's type.
func (d *Document) resolveLambdaParamType(name string, method *parser.MethodDecl, line int) string {
	m := d.findLambdaParam(name, method, line)
	if m == nil {
		return ""
	}
	return resolveMatchType(m)
}

func (d *Document) findPrevMeaningfulToken(line, col int) *lexer.Token {
	tok := d.TokenAtPosition(line, col)
	if tok == nil {
		return nil
	}
	return d.findPrevMeaningfulTokenBefore(tok)
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

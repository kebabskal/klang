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

		// Check event member access (bare event name: event_name.emit/connect)
		cls, _ := d.FindEnclosingClass(line)
		if cls != nil {
			if ev := d.findEventByName(cls, objName); ev != nil {
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
		for _, ev := range resolvedCls.Events {
			if ev.Name == member {
				return &HoverResult{
					Content: "```klang\n" + member + ":event(" + formatEventParams(ev) + ")\n```\nEvent on `" + typeName + "`",
					Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
				}
			}
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

	// Build type parameter substitution map from typeName (e.g. "Stack<string>" → {T: string})
	sub := buildTypeParamSub(typeName, cls)

	for _, f := range cls.Fields {
		if f.Name == member {
			ktype := typeExprToString(f.TypeExpr)
			if ktype == "" {
				ktype = "(inferred)"
			}
			ktype = applyTypeParamSub(ktype, sub)
			return &HoverResult{
				Content: "```klang\n" + member + ":" + ktype + "\n```\nField on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}
	for _, m := range cls.Methods {
		if m.Name == member {
			sig := applyTypeParamSub(formatMethodSignature(m), sub)
			return &HoverResult{
				Content: "```klang\n" + sig + "\n```\nMethod on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}
	for _, p := range cls.Properties {
		if p.Name == member {
			ktype := applyTypeParamSub(typeExprToString(p.TypeExpr), sub)
			return &HoverResult{
				Content: "```klang\n" + member + ":" + ktype + "\n```\nProperty on `" + typeName + "`",
				Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
			}
		}
	}
	for _, ev := range cls.Events {
		if ev.Name == member {
			return &HoverResult{
				Content: "```klang\n" + member + ":event(" + formatEventParams(ev) + ")\n```\nEvent on `" + typeName + "`",
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
					return &HoverResult{
						Content: "```klang\n" + detail + "\n```\nMethod on `" + typeName + "`",
						Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
					}
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

		// Check events on current class
		for _, ev := range cls.Events {
			if ev.Name == name {
				return &HoverResult{
					Content: "```klang\n" + name + ":event(" + formatEventParams(ev) + ")\n```\nEvent on `" + cls.Name + "`",
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

func (d *Document) hoverEventMember(ev *parser.EventDecl, member string, tok *lexer.Token) *HoverResult {
	paramStr := formatEventParams(ev)
	switch member {
	case "connect":
		handlerSig := "(" + paramStr + ") => void"
		return &HoverResult{
			Content: "```klang\n" + ev.Name + ".connect(handler: " + handlerSig + ")\n```\nConnect a handler to this event",
			Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
		}
	case "disconnect":
		return &HoverResult{
			Content: "```klang\n" + ev.Name + ".disconnect(handler)\n```\nDisconnect a handler from this event",
			Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
		}
	case "emit":
		sig := formatEventSignature(ev)
		return &HoverResult{
			Content: "```klang\n" + ev.Name + "." + sig + "\n```\nEmit this event, calling all connected handlers",
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

// hoverLambdaParam checks if the cursor is on a lambda parameter and resolves its type
// from the enclosing call context (e.g., event connect handler params).
func (d *Document) hoverLambdaParam(name string, method *parser.MethodDecl, line int, tok *lexer.Token) *HoverResult {
	if method.Body == nil {
		return nil
	}
	return d.findLambdaParamInBlock(name, method.Body, line, tok)
}

func (d *Document) findLambdaParamInBlock(name string, block *parser.Block, line int, tok *lexer.Token) *HoverResult {
	for _, stmt := range block.Stmts {
		if result := d.findLambdaParamInStmt(name, stmt, line, tok); result != nil {
			return result
		}
	}
	return nil
}

func (d *Document) findLambdaParamInStmt(name string, stmt parser.Stmt, line int, tok *lexer.Token) *HoverResult {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		return d.findLambdaParamInExpr(name, s.Expr, line, tok, nil)
	case *parser.VarDecl:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, tok, nil)
		}
	case *parser.AssignStmt:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, tok, nil)
		}
	case *parser.ForStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line, tok)
		}
	case *parser.IfStmt:
		if s.Then != nil {
			if r := d.findLambdaParamInBlock(name, s.Then, line, tok); r != nil {
				return r
			}
		}
		if blk, ok := s.Else.(*parser.Block); ok {
			if r := d.findLambdaParamInBlock(name, blk, line, tok); r != nil {
				return r
			}
		}
	case *parser.WhileStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line, tok)
		}
	case *parser.WithStmt:
		if s.Body != nil {
			return d.findLambdaParamInBlock(name, s.Body, line, tok)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			return d.findLambdaParamInExpr(name, s.Value, line, tok, nil)
		}
	}
	return nil
}

// findLambdaParamInExpr walks expressions to find lambda params.
// paramHints are the expected types from the enclosing call context.
func (d *Document) findLambdaParamInExpr(name string, expr parser.Expr, line int, tok *lexer.Token, paramHints []*parser.Param) *HoverResult {
	switch e := expr.(type) {
	case *parser.CallExpr:
		// Resolve what's being called to get parameter type hints for lambda args
		hints := d.resolveCallParamHints(e)
		for i, arg := range e.Args {
			var argHints []*parser.Param
			if hints != nil && i < len(hints) {
				argHints = hints
			}
			if r := d.findLambdaParamInExpr(name, arg, line, tok, argHints); r != nil {
				return r
			}
		}
	case *parser.LambdaExpr:
		// Check if the cursor line falls within this lambda's scope
		lambdaStartLine := 0
		if len(e.Params) > 0 && e.Params[0].Pos.Line > 0 {
			lambdaStartLine = e.Params[0].Pos.Line
		}
		if lambdaStartLine > 0 && line >= lambdaStartLine {
			// Check if any of the lambda's params match our name
			for i, p := range e.Params {
				if p.Name == name {
					ktype := typeExprToString(p.TypeExpr)
					if ktype == "" && paramHints != nil && i < len(paramHints) {
						ktype = typeExprToString(paramHints[i].TypeExpr)
					}
					if ktype == "" {
						ktype = "(inferred)"
					}
					return &HoverResult{
						Content: "```klang\n" + name + ":" + ktype + "\n```\nLambda parameter",
						Line:    tok.Line, Col: tok.Col, EndCol: tok.Col + len(tok.Value),
					}
				}
			}
			// Also check inside the lambda body for nested lambdas
			if e.Body != nil {
				return d.findLambdaParamInBlock(name, e.Body, line, tok)
			}
		}
	case *parser.MemberExpr:
		return d.findLambdaParamInExpr(name, e.Object, line, tok, nil)
	case *parser.BinaryExpr:
		if r := d.findLambdaParamInExpr(name, e.Left, line, tok, nil); r != nil {
			return r
		}
		return d.findLambdaParamInExpr(name, e.Right, line, tok, nil)
	case *parser.UnaryExpr:
		return d.findLambdaParamInExpr(name, e.Operand, line, tok, nil)
	}
	return nil
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
		for _, ev := range cls.Events {
			if ev.Name == methodName {
				// The lambda should receive the event's params
				return ev.Params
			}
		}
		// Check methods
		for _, m := range cls.Methods {
			if m.Name == methodName {
				return m.Params
			}
		}
		if cls.Parent != "" {
			parentCls := d.findClass(cls.Parent, classes)
			if parentCls != nil {
				for _, m := range parentCls.Methods {
					if m.Name == methodName {
						return m.Params
					}
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
	if method.Body == nil {
		return ""
	}
	return d.findLambdaParamTypeInBlock(name, method.Body, line)
}

func (d *Document) findLambdaParamTypeInBlock(name string, block *parser.Block, line int) string {
	for _, stmt := range block.Stmts {
		if t := d.findLambdaParamTypeInStmt(name, stmt, line); t != "" {
			return t
		}
	}
	return ""
}

func (d *Document) findLambdaParamTypeInStmt(name string, stmt parser.Stmt, line int) string {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		return d.findLambdaParamTypeInExpr(name, s.Expr, line, nil)
	case *parser.VarDecl:
		if s.Value != nil {
			return d.findLambdaParamTypeInExpr(name, s.Value, line, nil)
		}
	case *parser.AssignStmt:
		if s.Value != nil {
			return d.findLambdaParamTypeInExpr(name, s.Value, line, nil)
		}
	case *parser.ForStmt:
		if s.Body != nil {
			return d.findLambdaParamTypeInBlock(name, s.Body, line)
		}
	case *parser.IfStmt:
		if s.Then != nil {
			if t := d.findLambdaParamTypeInBlock(name, s.Then, line); t != "" {
				return t
			}
		}
		if blk, ok := s.Else.(*parser.Block); ok {
			return d.findLambdaParamTypeInBlock(name, blk, line)
		}
	case *parser.WhileStmt:
		if s.Body != nil {
			return d.findLambdaParamTypeInBlock(name, s.Body, line)
		}
	case *parser.WithStmt:
		if s.Body != nil {
			return d.findLambdaParamTypeInBlock(name, s.Body, line)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			return d.findLambdaParamTypeInExpr(name, s.Value, line, nil)
		}
	}
	return ""
}

func (d *Document) findLambdaParamTypeInExpr(name string, expr parser.Expr, line int, paramHints []*parser.Param) string {
	switch e := expr.(type) {
	case *parser.CallExpr:
		hints := d.resolveCallParamHints(e)
		for i, arg := range e.Args {
			var argHints []*parser.Param
			if hints != nil && i < len(hints) {
				argHints = hints
			}
			if t := d.findLambdaParamTypeInExpr(name, arg, line, argHints); t != "" {
				return t
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
					ktype := typeExprToString(p.TypeExpr)
					if ktype == "" && paramHints != nil && i < len(paramHints) {
						ktype = typeExprToString(paramHints[i].TypeExpr)
					}
					return ktype
				}
			}
			if e.Body != nil {
				return d.findLambdaParamTypeInBlock(name, e.Body, line)
			}
		}
	case *parser.MemberExpr:
		return d.findLambdaParamTypeInExpr(name, e.Object, line, nil)
	case *parser.BinaryExpr:
		if t := d.findLambdaParamTypeInExpr(name, e.Left, line, nil); t != "" {
			return t
		}
		return d.findLambdaParamTypeInExpr(name, e.Right, line, nil)
	case *parser.UnaryExpr:
		return d.findLambdaParamTypeInExpr(name, e.Operand, line, nil)
	}
	return ""
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

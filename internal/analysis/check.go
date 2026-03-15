package analysis

import (
	"fmt"
	"strings"

	"github.com/klang-lang/klang/internal/parser"
)

// checkScope tracks identifiers and their types during semantic checking.
type checkScope struct {
	vars          map[string]string // name -> klang type ("" = known but type unknown)
	hasWithModule bool
	className     string // enclosing class name (for resolving bare method calls)
}

func newScope() *checkScope {
	return &checkScope{vars: make(map[string]string)}
}

func (s *checkScope) copy() *checkScope {
	ns := &checkScope{vars: make(map[string]string), hasWithModule: s.hasWithModule, className: s.className}
	for k, v := range s.vars {
		ns.vars[k] = v
	}
	return ns
}

func (s *checkScope) set(name, ktype string) {
	s.vars[name] = ktype
}

func (s *checkScope) has(name string) bool {
	_, ok := s.vars[name]
	return ok
}

func (s *checkScope) typeOf(name string) string {
	return s.vars[name]
}

// Check performs semantic validation on the document and appends diagnostics.
func (d *Document) Check() {
	ensureVendorsMerged()
	if d.AST == nil {
		return
	}
	for _, cls := range d.AST.Classes {
		d.checkClass("", cls)
	}
}

func (d *Document) checkClass(prefix string, cls *parser.ClassDecl) {
	fullName := cls.Name
	if prefix != "" {
		fullName = prefix + "_" + cls.Name
	}
	for _, nested := range cls.Classes {
		d.checkClass(fullName, nested)
	}
	for _, method := range cls.Methods {
		d.checkMethod(cls, fullName, method)
	}
	for _, prop := range cls.Properties {
		d.checkProperty(cls, fullName, prop)
	}
}

func (d *Document) checkMethod(cls *parser.ClassDecl, className string, method *parser.MethodDecl) {
	scope := d.buildClassScope(cls, className)

	// Method parameters (with types)
	for _, p := range method.Params {
		ktype := typeExprToString(p.TypeExpr)
		scope.set(p.Name, ktype)
	}

	if method.Body != nil {
		d.checkBlock(method.Body, className, scope)
	}
}

func (d *Document) checkProperty(cls *parser.ClassDecl, className string, prop *parser.PropertyDecl) {
	scope := d.buildClassScope(cls, className)

	// Check getter expression
	if prop.Getter != nil {
		d.checkExpr(prop.Getter, scope)
	}

	// Check setter block (add setter parameter to scope)
	if prop.Setter != nil {
		setterScope := scope.copy()
		paramName := prop.SetParam
		if paramName == "" {
			paramName = "value"
		}
		ktype := typeExprToString(prop.TypeExpr)
		setterScope.set(paramName, ktype)
		d.checkBlock(prop.Setter, className, setterScope)
	}
}

// buildClassScope creates a scope with class fields, methods, events, properties, and builtins.
func (d *Document) buildClassScope(cls *parser.ClassDecl, className string) *checkScope {
	scope := newScope()

	for _, f := range cls.Fields {
		ktype := typeExprToString(f.TypeExpr)
		if ktype == "" {
			ktype = d.ResolveFieldType(f, className)
		}
		scope.set(f.Name, ktype)
	}
	for _, m := range cls.Methods {
		scope.set(m.Name, "method")
	}
	for _, ev := range cls.Events {
		scope.set(ev.Name, "event")
	}
	for _, prop := range cls.Properties {
		ktype := typeExprToString(prop.TypeExpr)
		scope.set(prop.Name, ktype)
	}
	if cls.Parent != "" {
		d.addParentScope(cls.Parent, scope)
	}
	classes := d.GetClasses()
	for _, c := range classes {
		scope.set(c.Name, "class")
	}
	scope.className = className
	for _, name := range builtinIdents {
		scope.set(name, "")
	}
	scope.set("this", cls.Name)
	scope.set("self", cls.Name)

	return scope
}

func (d *Document) addParentScope(parentName string, scope *checkScope) {
	classes := d.GetClasses()
	if classes == nil {
		return
	}
	parent := d.findClass(parentName, classes)
	if parent == nil {
		return
	}
	for _, f := range parent.Fields {
		ktype := typeExprToString(f.TypeExpr)
		scope.set(f.Name, ktype)
	}
	for _, m := range parent.Methods {
		scope.set(m.Name, "method")
	}
	for _, ev := range parent.Events {
		scope.set(ev.Name, "event")
	}
	for _, prop := range parent.Properties {
		ktype := typeExprToString(prop.TypeExpr)
		scope.set(prop.Name, ktype)
	}
	if parent.Parent != "" {
		d.addParentScope(parent.Parent, scope)
	}
}

func (d *Document) checkBlock(block *parser.Block, className string, scope *checkScope) {
	blockScope := scope.copy()
	for _, stmt := range block.Stmts {
		d.checkStmt(stmt, className, blockScope)
	}
}

func (d *Document) checkStmt(stmt parser.Stmt, className string, scope *checkScope) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		if s.Value != nil {
			d.checkExpr(s.Value, scope)
		}
		// Infer type for the variable
		ktype := typeExprToString(s.TypeExpr)
		if ktype == "" && s.Value != nil {
			ktype = d.inferExprType(s.Value, scope)
		}
		scope.set(s.Name, ktype)

	case *parser.AssignStmt:
		d.checkExpr(s.Target, scope)
		d.checkExpr(s.Value, scope)
		// Type-check assignment: target type must be compatible with value type
		if s.Op == "=" {
			targetType, targetName := d.resolveAssignTarget(s.Target, scope)
			if targetType != "" {
				valueType := d.inferExprType(s.Value, scope)
				if valueType != "" && !typesCompatible(valueType, targetType) {
					pos := d.exprPos(s.Value)
					if pos.Line <= 0 {
						pos = d.exprPos(s.Target)
					}
					d.addError(pos, fmt.Sprintf("cannot assign type '%s' to '%s' of type '%s'", valueType, targetName, targetType))
				}
			}
		}

	case *parser.ExprStmt:
		d.checkExpr(s.Expr, scope)

	case *parser.ReturnStmt:
		if s.Value != nil {
			d.checkExpr(s.Value, scope)
		}

	case *parser.IfStmt:
		d.checkExpr(s.Condition, scope)
		if s.Then != nil {
			d.checkBlock(s.Then, className, scope)
		}
		if s.ThenStmt != nil {
			d.checkStmt(s.ThenStmt, className, scope)
		}
		if blk, ok := s.Else.(*parser.Block); ok {
			d.checkBlock(blk, className, scope)
		} else if elif, ok := s.Else.(*parser.IfStmt); ok {
			d.checkStmt(elif, className, scope)
		}

	case *parser.WhileStmt:
		d.checkExpr(s.Condition, scope)
		if s.Body != nil {
			d.checkBlock(s.Body, className, scope)
		}

	case *parser.ForStmt:
		if s.Iterable != nil {
			d.checkExpr(s.Iterable, scope)
		}
		innerScope := scope.copy()
		if s.VarName != "" {
			if s.ValueVar != "" {
				// for key, value in dict
				keyType, valType := d.inferForDictTypes(s.Iterable)
				innerScope.set(s.VarName, keyType)
				innerScope.set(s.ValueVar, valType)
			} else {
				ktype := d.inferForVarType(s.Iterable)
				innerScope.set(s.VarName, ktype)
			}
		}
		if s.Body != nil {
			d.checkBlock(s.Body, className, innerScope)
		}

	case *parser.WithStmt:
		if s.Module != "" {
			d.addWithModuleScope(s.Module, scope)
		}
		if s.Body != nil {
			d.checkBlock(s.Body, className, scope)
		}

	case *parser.InlineCStmt:
		// skip
	}
}

func (d *Document) checkExpr(expr parser.Expr, scope *checkScope) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *parser.Ident:
		if !scope.has(e.Name) && e.Pos.Line > 0 {
			d.addError(e.Pos, "undeclared identifier '"+e.Name+"'")
		}

	case *parser.MemberExpr:
		d.checkExpr(e.Object, scope)
		d.checkMemberExists(e, scope)

	case *parser.CallExpr:
		// Check callee (allow unknown bare function calls when with-module active)
		if ident, ok := e.Callee.(*parser.Ident); ok {
			if !scope.has(ident.Name) && scope.hasWithModule {
				// Likely a with-module function — skip
			} else {
				d.checkExpr(e.Callee, scope)
			}
		} else {
			d.checkExpr(e.Callee, scope)
		}
		// Check arguments
		for _, arg := range e.Args {
			d.checkExpr(arg, scope)
		}
		// Type-check arguments against parameter types
		d.checkCallArgs(e, scope)

	case *parser.BinaryExpr:
		d.checkExpr(e.Left, scope)
		d.checkExpr(e.Right, scope)

	case *parser.UnaryExpr:
		d.checkExpr(e.Operand, scope)

	case *parser.IndexExpr:
		d.checkExpr(e.Object, scope)
		d.checkExpr(e.Index, scope)

	case *parser.ArrayLit:
		for _, elem := range e.Elements {
			d.checkExpr(elem, scope)
		}

	case *parser.LambdaExpr:
		lambdaScope := scope.copy()
		for _, p := range e.Params {
			ktype := typeExprToString(p.TypeExpr)
			lambdaScope.set(p.Name, ktype)
		}
		if e.Body != nil {
			d.checkBlock(e.Body, "", lambdaScope)
		}

	case *parser.StructLit:
		for _, f := range e.Fields {
			d.checkExpr(f.Value, scope)
		}

	case *parser.IsExpr:
		d.checkExpr(e.Expr, scope)

	case *parser.SpreadExpr, *parser.ThisExpr:
		// ok

	case *parser.IntLit, *parser.FloatLit, *parser.StringLit, *parser.BoolLit:
		// ok
	}
}

// checkCallArgs checks that argument types match parameter types for a call.
func (d *Document) checkCallArgs(call *parser.CallExpr, scope *checkScope) {
	if d.Gen == nil {
		return
	}

	// Resolve parameter types for the callee
	paramTypes := d.resolveCallParamTypes(call, scope)
	if paramTypes == nil {
		return
	}

	// Check argument count
	if len(call.Args) != len(paramTypes) {
		// Find position for the error — use the callee position
		pos := d.exprPos(call.Callee)
		if pos.Line > 0 {
			d.addError(pos, fmt.Sprintf("expected %d argument(s), got %d", len(paramTypes), len(call.Args)))
		}
		return
	}

	// Check each argument type
	for i, arg := range call.Args {
		if i >= len(paramTypes) {
			break
		}
		expectedType := paramTypes[i]
		if expectedType == "" {
			continue // unknown expected type
		}

		actualType := d.inferExprType(arg, scope)
		if actualType == "" {
			continue // can't infer argument type
		}

		if !typesCompatible(actualType, expectedType) {
			pos := d.exprPos(arg)
			if pos.Line <= 0 {
				pos = d.exprPos(call.Callee)
			}
			if pos.Line > 0 {
				d.addError(pos, fmt.Sprintf("type '%s' is not assignable to parameter type '%s'", actualType, expectedType))
			}
		}
	}
}

// checkMemberExists verifies that a member access (obj.field or obj.method) is valid.
func (d *Document) checkMemberExists(member *parser.MemberExpr, scope *checkScope) {
	// Skip if no position info
	if member.Pos.Line <= 0 {
		return
	}

	// Resolve the object's type
	typeName := ""
	if ident, ok := member.Object.(*parser.Ident); ok {
		typeName = scope.typeOf(ident.Name)

		// Skip module access (math.sin, rl.draw_circle_v, etc.)
		if _, isModule := StdlibModuleSignatures[ident.Name]; isModule {
			return
		}
		if _, isModule := StdlibModuleConstantNames[ident.Name]; isModule {
			return
		}
		// Known module names that might not have signatures registered
		switch ident.Name {
		case "math", "io", "rl", "os", "Colors", "Key", "Mouse", "Gamepad":
			return
		}
	}

	if typeName == "" || typeName == "method" || typeName == "func" || typeName == "class" {
		return // can't resolve type, skip
	}

	// Event members: connect, emit, disconnect are always valid
	if typeName == "event" {
		switch member.Field {
		case "connect", "emit", "disconnect":
			return
		}
		pos := parser.Pos{Line: member.Pos.Line, Col: member.Pos.Col, EndCol: member.Pos.Col + len(member.Field)}
		d.addError(pos, fmt.Sprintf("event has no member '%s' (use connect, emit, or disconnect)", member.Field))
		return
	}

	// Built-in value types have known fields — don't check those
	switch typeName {
	case "vec2", "vec3", "vec4", "mat4", "quat",
		"int", "float", "string", "bool", "List", "Dictionary":
		return
	}
	if _, ok := BuiltinTypeMembers[typeName]; ok {
		return
	}

	// Look up the class
	classes := d.GetClasses()
	if classes == nil {
		return
	}
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return
	}

	// Check if the field or method exists
	if d.classHasMember(cls, member.Field, classes) {
		return
	}

	// Also check List built-in methods if the type is List<T>
	// (already handled above via typeName == "List")

	pos := parser.Pos{Line: member.Pos.Line, Col: member.Pos.Col, EndCol: member.Pos.Col + len(member.Field)}
	d.addError(pos, fmt.Sprintf("'%s' has no member '%s'", typeName, member.Field))
}

// classHasMember checks if a class (or its parents) has a field or method with the given name.
func (d *Document) classHasMember(cls *parser.ClassDecl, name string, classes map[string]*parser.ClassDecl) bool {
	if cls.HasMember(name) {
		return true
	}
	if cls.Parent != "" {
		parent := d.findClass(cls.Parent, classes)
		if parent != nil {
			return d.classHasMember(parent, name, classes)
		}
	}
	return false
}

// resolveCallParamTypes returns the Klang parameter types for a call expression.
func (d *Document) resolveCallParamTypes(call *parser.CallExpr, scope *checkScope) []string {
	classes := d.GetClasses()

	// Member call: obj.method(...)
	if member, ok := call.Callee.(*parser.MemberExpr); ok {
		if objIdent, ok := member.Object.(*parser.Ident); ok {
			// Check if it's an event.emit() call — validate args against event params
			if scope.typeOf(objIdent.Name) == "event" && member.Field == "emit" {
				return d.resolveEventEmitParamTypes(objIdent.Name)
			}

			// Check if it's a module call
			if sigs, ok := StdlibModuleSignatures[objIdent.Name]; ok {
				for _, sig := range sigs {
					if sig.Name == member.Field {
						return parseParamTypes(sig.Detail)
					}
				}
			}

			// Resolve object type → find method
			typeName := scope.typeOf(objIdent.Name)
			if typeName == "" {
				return nil
			}
			if classes != nil {
				return d.findMethodParamTypes(typeName, member.Field, classes)
			}
		}
		return nil
	}

	// Bare call: func(...)
	if ident, ok := call.Callee.(*parser.Ident); ok {
		if classes != nil {
			// Check enclosing class methods first
			if scope.className != "" {
				if enclosing := d.findClass(scope.className, classes); enclosing != nil {
					if m := enclosing.FindMethod(ident.Name); m != nil {
						params := methodParamTypes(m)
						return substituteTypeParams(params, m.TypeParams, call.TypeArgs)
					}
				}
			}
			// Check other class methods
			for _, cls := range classes {
				if m := cls.FindMethod(ident.Name); m != nil {
					params := methodParamTypes(m)
					return substituteTypeParams(params, m.TypeParams, call.TypeArgs)
				}
			}
			// Check constructors
			for _, cls := range classes {
				if cls.Name == ident.Name && cls.Constructor != nil {
					params := constructorParamTypes(cls.Constructor)
					return substituteTypeParams(params, cls.TypeParams, call.TypeArgs)
				}
			}
		}

		// Check built-in constructors (core + vendor)
		builtinParams := map[string][]string{
			"vec2": {"float", "float"},
			"vec3": {"float", "float", "float"},
			"vec4": {"float", "float", "float", "float"},
			"quat": {"float", "float", "float", "float"},
		}
		for k, v := range VendorBuiltinConstructorParams() {
			builtinParams[k] = v
		}
		if params, ok := builtinParams[ident.Name]; ok {
			return params
		}
	}

	return nil
}

func (d *Document) findMethodParamTypes(typeName, methodName string, classes map[string]*parser.ClassDecl) []string {
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}
	if m := cls.FindMethod(methodName); m != nil {
		params := methodParamTypes(m)
		if len(cls.TypeParams) > 0 {
			tpSet := make(map[string]bool)
			for _, tp := range cls.TypeParams {
				tpSet[tp] = true
			}
			for i, p := range params {
				if tpSet[p] {
					params[i] = ""
				}
			}
		}
		return params
	}
	if cls.Parent != "" {
		return d.findMethodParamTypes(cls.Parent, methodName, classes)
	}
	return nil
}

func methodParamTypes(m *parser.MethodDecl) []string {
	types := make([]string, len(m.Params))
	for i, p := range m.Params {
		types[i] = typeExprToString(p.TypeExpr)
	}
	return types
}

func constructorParamTypes(c *parser.ConstructorDecl) []string {
	types := make([]string, len(c.Params))
	for i, p := range c.Params {
		types[i] = typeExprToString(p.TypeExpr)
	}
	return types
}

// substituteTypeParams replaces generic type parameter names with concrete types.
// typeParams are the declared names (e.g. ["T", "U"]) and typeArgs are the concrete
// types from the call site (e.g. [SimpleType("int"), SimpleType("string")]).
// If typeArgs is empty or doesn't match, the params are returned as-is, which
// effectively skips type checking for unresolved generics.
func substituteTypeParams(params []string, typeParams []string, typeArgs []parser.TypeExpr) []string {
	if len(typeParams) == 0 {
		return params
	}
	// Build substitution map
	sub := make(map[string]string)
	if len(typeArgs) == len(typeParams) {
		// Explicit type args provided: <int, string>
		for i, tp := range typeParams {
			sub[tp] = typeExprToString(typeArgs[i])
		}
	} else {
		// No explicit type args — skip checking by mapping params to ""
		// (empty type is treated as compatible with anything)
		for _, tp := range typeParams {
			sub[tp] = ""
		}
	}
	result := make([]string, len(params))
	for i, p := range params {
		if concrete, ok := sub[p]; ok {
			result[i] = concrete
		} else {
			result[i] = p
		}
	}
	return result
}

// inferExprType infers the Klang type of an expression using scope and codegen.
func (d *Document) inferExprType(expr parser.Expr, scope *checkScope) string {
	switch e := expr.(type) {
	case *parser.IntLit:
		return "int"
	case *parser.FloatLit:
		return "float"
	case *parser.StringLit:
		return "string"
	case *parser.BoolLit:
		return "bool"
	case *parser.Ident:
		return scope.typeOf(e.Name)
	case *parser.ThisExpr:
		return scope.typeOf("this")
	case *parser.CallExpr:
		// Constructor call: ClassName() → type is the class name
		if ident, ok := e.Callee.(*parser.Ident); ok {
			if scope.typeOf(ident.Name) == "class" {
				return ident.Name
			}
		}
		// Use codegen to infer return type
		if d.Gen != nil {
			cType := d.Gen.InferCType(expr)
			return CTypeToKlang(cType)
		}
	case *parser.MemberExpr:
		// Try scope-based resolution first
		objType := d.inferExprType(e.Object, scope)
		if objType != "" {
			fieldType := d.resolveFieldKlangType(objType, e.Field)
			if fieldType != "" {
				return fieldType
			}
		}
		// Fallback to codegen
		if d.Gen != nil {
			cType := d.Gen.InferCType(expr)
			return CTypeToKlang(cType)
		}
	case *parser.IndexExpr:
		objType := d.inferExprType(e.Object, scope)
		elemType := extractListElementType(objType)
		if elemType != "" {
			return elemType
		}
		_, valType := extractDictTypes(objType)
		if valType != "" {
			return valType
		}
	case *parser.BinaryExpr:
		switch e.Op {
		case "==", "!=", "<", ">", "<=", ">=", "and", "or":
			return "bool"
		}
		leftType := d.inferExprType(e.Left, scope)
		rightType := d.inferExprType(e.Right, scope)
		// vec types propagate
		for _, t := range []string{"vec2", "vec3", "vec4", "quat"} {
			if leftType == t || rightType == t {
				return t
			}
		}
		// float promotion
		if leftType == "float" || rightType == "float" {
			return "float"
		}
		if leftType != "" {
			return leftType
		}
		return rightType
	case *parser.UnaryExpr:
		return d.inferExprType(e.Operand, scope)
	}
	// Fallback to codegen
	if d.Gen != nil {
		cType := d.Gen.InferCType(expr)
		return CTypeToKlang(cType)
	}
	return ""
}

// resolveAssignTarget returns the Klang type and display name of an assignment target.
func (d *Document) resolveAssignTarget(target parser.Expr, scope *checkScope) (string, string) {
	switch t := target.(type) {
	case *parser.Ident:
		return scope.typeOf(t.Name), t.Name
	case *parser.MemberExpr:
		objType := d.inferExprType(t.Object, scope)
		if objType != "" {
			fieldType := d.resolveFieldKlangType(objType, t.Field)
			if fieldType != "" {
				return fieldType, t.Field
			}
		}
	case *parser.IndexExpr:
		elemType := d.inferExprType(target, scope)
		if elemType != "" {
			return elemType, "element"
		}
	}
	return "", ""
}

// extractListElementType extracts the element type from "List<T>" → "T".
func extractListElementType(listType string) string {
	if !strings.HasPrefix(listType, "List<") || !strings.HasSuffix(listType, ">") {
		return ""
	}
	return listType[5 : len(listType)-1]
}

// extractDictTypes extracts key and value types from "Dictionary<K, V>".
func extractDictTypes(dictType string) (string, string) {
	if !strings.HasPrefix(dictType, "Dictionary<") || !strings.HasSuffix(dictType, ">") {
		return "", ""
	}
	inner := dictType[11 : len(dictType)-1]
	// Split on ", " (handling nested generics)
	depth := 0
	for i, c := range inner {
		switch c {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(inner[:i]), strings.TrimSpace(inner[i+1:])
			}
		}
	}
	return "", ""
}

// resolveFieldKlangType looks up the Klang type of a field on a class by name.
func (d *Document) resolveFieldKlangType(className, fieldName string) string {
	// Built-in value type fields (core + vendor via BuiltinTypeFieldTypes)
	if fields, ok := BuiltinTypeFieldTypes[className]; ok {
		if ft, ok := fields[fieldName]; ok {
			return ft
		}
	}
	// Look up in class definitions
	classes := d.GetClasses()
	if classes == nil {
		return ""
	}
	cls := d.findClass(className, classes)
	if cls == nil {
		return ""
	}
	if f := cls.FindField(fieldName); f != nil {
		ktype := typeExprToString(f.TypeExpr)
		if ktype != "" {
			return resolveTypeWithParams(className, cls, ktype)
		}
		return d.ResolveFieldType(f, className)
	}
	if prop := cls.FindProperty(fieldName); prop != nil {
		return resolveTypeWithParams(className, cls, typeExprToString(prop.TypeExpr))
	}
	// Check parent
	if cls.Parent != "" {
		return d.resolveFieldKlangType(cls.Parent, fieldName)
	}
	return ""
}

// exprPos returns the position of an expression.
func (d *Document) exprPos(expr parser.Expr) parser.Pos {
	switch e := expr.(type) {
	case *parser.Ident:
		return e.Pos
	case *parser.MemberExpr:
		return e.Pos
	case *parser.CallExpr:
		return d.exprPos(e.Callee)
	case *parser.IndexExpr:
		return e.Pos
	case *parser.BinaryExpr:
		return d.exprPos(e.Left)
	case *parser.UnaryExpr:
		return d.exprPos(e.Operand)
	}
	return parser.Pos{}
}

// typesCompatible checks if actualType can be used where expectedType is required.
func typesCompatible(actual, expected string) bool {
	if actual == expected {
		return true
	}
	// int and float are interchangeable in many contexts
	if (actual == "int" && expected == "float") || (actual == "float" && expected == "int") {
		return true
	}
	// Any class pointer is compatible with void (generic/unknown)
	if expected == "" || actual == "" {
		return true
	}
	return false
}

// parseParamTypes extracts parameter types from a signature string like "func(x:int, y:float):ret"
func parseParamTypes(detail string) []string {
	parenStart := -1
	parenEnd := -1
	for i, c := range detail {
		if c == '(' {
			parenStart = i
		}
		if c == ')' {
			parenEnd = i
			break
		}
	}
	if parenStart < 0 || parenEnd < 0 || parenEnd <= parenStart+1 {
		return []string{} // no params
	}

	paramStr := detail[parenStart+1 : parenEnd]
	var types []string
	for _, p := range splitParams(paramStr) {
		colonIdx := -1
		for i, c := range p {
			if c == ':' {
				colonIdx = i
				break
			}
		}
		if colonIdx >= 0 {
			types = append(types, p[colonIdx+1:])
		} else {
			types = append(types, "")
		}
	}
	return types
}

// splitParams splits "x:int, y:float" handling nested generics
func splitParams(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, c := range s {
		switch c {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				part := trimSpace(s[start:i])
				if part != "" {
					parts = append(parts, part)
				}
				start = i + 1
			}
		}
	}
	part := trimSpace(s[start:])
	if part != "" {
		parts = append(parts, part)
	}
	return parts
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}

func (d *Document) addWithModuleScope(module string, scope *checkScope) {
	if sigs, ok := StdlibModuleSignatures[module]; ok {
		for _, sig := range sigs {
			scope.set(sig.Name, "func")
		}
	}
	if consts, ok := StdlibModuleConstantNames[module]; ok {
		for _, c := range consts {
			scope.set(c.Name, c.Detail)
		}
	}
	scope.hasWithModule = true
}

// resolveEventEmitParamTypes finds the event declaration and returns its parameter types.
func (d *Document) resolveEventEmitParamTypes(eventName string) []string {
	if d.AST == nil {
		return nil
	}
	for _, cls := range d.AST.Classes {
		if ev := cls.FindEvent(eventName); ev != nil {
			types := make([]string, len(ev.Params))
			for i, p := range ev.Params {
				types[i] = typeExprToString(p.TypeExpr)
			}
			return types
		}
	}
	return nil
}

// builtinIdents are identifiers that are always valid (constructors, globals, etc.)
// Vendor-contributed identifiers are appended by ensureVendorsMerged().
var builtinIdents = []string{
	"int", "float", "string", "bool", "void",
	"vec2", "vec3", "vec4", "mat4", "quat",
	"List", "Dictionary", "Random",
	"print", "println", "str", "len", "append", "remove",
	"true", "false", "nil", "null", "this", "self",
	"math", "io", "os",
	"not",
}

package analysis

import (
	"strings"

	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label      string
	Detail     string
	Kind       CompletionKind
	InsertText string // if different from Label
}

// CompletionKind maps to LSP CompletionItemKind.
type CompletionKind int

const (
	CompletionKindKeyword   CompletionKind = 14
	CompletionKindClass     CompletionKind = 7
	CompletionKindField     CompletionKind = 5
	CompletionKindMethod    CompletionKind = 2
	CompletionKindFunction  CompletionKind = 3
	CompletionKindVariable  CompletionKind = 6
	CompletionKindConstant  CompletionKind = 21
	CompletionKindModule    CompletionKind = 9
	CompletionKindEnum      CompletionKind = 13
	CompletionKindProperty  CompletionKind = 10
	CompletionKindType      CompletionKind = 25 // TypeParameter
)

// Complete returns completion items at the given cursor position (1-based line/col).
func (d *Document) Complete(line, col int) []CompletionItem {
	ensureVendorsMerged()
	if d.AST == nil {
		return nil
	}

	// Find the context: is the cursor after a dot?
	trigger := d.findCompletionTrigger(line, col)

	switch trigger.kind {
	case triggerDot:
		return d.completeDot(trigger, line)
	case triggerColon:
		return d.completeType()
	default:
		return d.completeBare(line)
	}
}

type triggerKind int

const (
	triggerBare  triggerKind = iota
	triggerDot               // after "expr."
	triggerColon             // after "name:"
)

type completionTrigger struct {
	kind     triggerKind
	objName  string   // for dot: the identifier before the dot (simple case)
	chain    []string // for dot: full member chain e.g. ["ball", "position"] for ball.position.
}

func (d *Document) findCompletionTrigger(line, col int) completionTrigger {
	// Walk backward through tokens to find the trigger
	var prevTokens []lexer.Token
	for _, t := range d.Tokens {
		if t.Type == lexer.TOKEN_NEWLINE || t.Type == lexer.TOKEN_INDENT || t.Type == lexer.TOKEN_EOF {
			continue
		}
		if t.Line > line || (t.Line == line && t.Col >= col) {
			break
		}
		prevTokens = append(prevTokens, t)
	}

	if len(prevTokens) == 0 {
		return completionTrigger{kind: triggerBare}
	}

	last := prevTokens[len(prevTokens)-1]

	// If the last token is a dot, walk backward to collect the full chain (e.g. ball.position.)
	if last.Type == lexer.TOKEN_DOT || last.Type == lexer.TOKEN_QUESTION_DOT {
		var chain []string
		i := len(prevTokens) - 2 // skip the trailing dot
		for i >= 0 {
			if prevTokens[i].Type == lexer.TOKEN_IDENT {
				chain = append([]string{prevTokens[i].Value}, chain...)
				i--
				// Check for another dot before this ident
				if i >= 0 && (prevTokens[i].Type == lexer.TOKEN_DOT || prevTokens[i].Type == lexer.TOKEN_QUESTION_DOT) {
					i-- // skip the dot, continue collecting
				} else {
					break
				}
			} else {
				break
			}
		}
		objName := ""
		if len(chain) > 0 {
			objName = chain[0]
		}
		return completionTrigger{kind: triggerDot, objName: objName, chain: chain}
	}

	// If the last token is a colon and the one before is an ident, type completion
	if last.Type == lexer.TOKEN_COLON {
		return completionTrigger{kind: triggerColon}
	}

	return completionTrigger{kind: triggerBare}
}

func (d *Document) completeDot(trigger completionTrigger, line int) []CompletionItem {
	var items []CompletionItem

	// Check chained module.namespace first: rl.CameraMode. → show ONLY enum values
	if len(trigger.chain) == 2 {
		nsName := trigger.chain[1]
		if members, ok := StdlibNamespaces[nsName]; ok {
			for _, m := range members {
				items = append(items, CompletionItem{
					Label:  m.Name,
					Detail: m.Detail,
					Kind:   CompletionKindConstant,
				})
			}
			return items
		}
	}

	// Check if objName is a module
	if sigs, ok := StdlibModuleSignatures[trigger.objName]; ok {
		for _, sig := range sigs {
			items = append(items, CompletionItem{
				Label:  sig.Name,
				Detail: sig.Detail,
				Kind:   CompletionKindFunction,
			})
		}
	}
	// Check module constants
	if consts, ok := StdlibModuleConstantNames[trigger.objName]; ok {
		for _, c := range consts {
			items = append(items, CompletionItem{
				Label:  c.Name,
				Detail: c.Detail,
				Kind:   CompletionKindConstant,
			})
		}
	}
	// Check namespace constants (Colors, Key, Mouse, Gamepad)
	if members, ok := StdlibNamespaces[trigger.objName]; ok {
		for _, m := range members {
			items = append(items, CompletionItem{
				Label:  m.Name,
				Detail: m.Detail,
				Kind:   CompletionKindConstant,
			})
		}
	}

	// Check if a module has sub-namespaces: rl. → show CameraMode, Flag, etc.
	if nsNames, ok := ModuleNamespaceMap[trigger.objName]; ok && len(trigger.chain) <= 1 {
		for _, ns := range nsNames {
			items = append(items, CompletionItem{
				Label:  ns,
				Detail: "enum",
				Kind:   CompletionKindEnum,
			})
		}
	}

	// Check if objName is a variable/field whose type has members
	if d.Gen != nil {
		if len(trigger.chain) > 1 {
			items = append(items, d.completeMemberChain(trigger.chain, line)...)
		} else {
			items = append(items, d.completeMemberAccess(trigger.objName, line)...)
		}
	}

	return items
}

func (d *Document) completeMemberAccess(objName string, line int) []CompletionItem {
	var items []CompletionItem
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	// Find the type of objName
	cls, fullName := d.FindEnclosingClass(line)
	if cls == nil {
		return nil
	}
	method := d.FindEnclosingMethod(cls, line)

	// Check locals (including inferred types)
	if method != nil {
		locals := d.CollectLocalsBeforeLine(method, line)
		for _, lv := range locals {
			if lv.Name == objName && lv.KType != "" {
				items = append(items, d.completeTypeMembers(lv.KType, classes)...)
				return items
			}
		}
	}

	// Check if objName is "this" — complete current class
	if objName == "this" {
		items = append(items, d.completeTypeMembers(fullName, classes)...)
		return items
	}

	// Check if objName is an event — complete with connect/emit/disconnect
	if ev := cls.FindEvent(objName); ev != nil {
		items = append(items, d.completeEventMembers(ev)...)
		return items
	}

	// Check fields on current class (resolve type via codegen for accuracy)
	if f := cls.FindField(objName); f != nil {
		typeName := d.ResolveFieldType(f, fullName)
		if typeName != "" {
			items = append(items, d.completeTypeMembers(typeName, classes)...)
			return items
		}
	}

	return items
}

func (d *Document) completeMemberChain(chain []string, line int) []CompletionItem {
	classes := d.GetClasses()
	if classes == nil {
		return nil
	}

	// Resolve the type of the first identifier in the chain
	typeName := d.resolveIdentType(chain[0], line)
	if typeName == "" {
		return nil
	}

	// Walk the chain, resolving each member's type
	for i := 1; i < len(chain); i++ {
		fieldName := chain[i]
		nextType := d.resolveFieldType(typeName, fieldName, classes)
		if nextType == "" {
			return nil
		}
		typeName = nextType
	}

	// Now complete members of the final resolved type
	return d.completeTypeMembers(typeName, classes)
}

// resolveFieldType resolves the type of a field on a given type.
func (d *Document) resolveFieldType(typeName, fieldName string, classes map[string]*parser.ClassDecl) string {
	// Check built-in types first
	if fields, ok := BuiltinTypeFieldTypes[typeName]; ok {
		if ft, ok := fields[fieldName]; ok {
			return ft
		}
	}

	// Check user classes
	cls := d.findClass(typeName, classes)
	if cls == nil {
		return ""
	}
	if f := cls.FindField(fieldName); f != nil {
		ktype := typeExprToString(f.TypeExpr)
		if ktype == "" {
			ktype = d.ResolveFieldType(f, typeName)
		}
		return ktype
	}
	// Check parent
	if cls.Parent != "" {
		return d.resolveFieldType(cls.Parent, fieldName, classes)
	}
	return ""
}

// completeTypeMembers returns completions for any type (user class or built-in).
func (d *Document) completeTypeMembers(typeName string, classes map[string]*parser.ClassDecl) []CompletionItem {
	// Check built-in types (exact match)
	if members, ok := BuiltinTypeMembers[typeName]; ok {
		return members
	}
	// Check built-in generic types: "List<Ball>" → "List"
	if idx := strings.Index(typeName, "<"); idx > 0 {
		baseName := typeName[:idx]
		if members, ok := BuiltinTypeMembers[baseName]; ok {
			return members
		}
	}
	// Fall through to class members
	return d.completeClassMembers(typeName, classes)
}

func (d *Document) completeClassMembers(typeName string, classes map[string]*parser.ClassDecl) []CompletionItem {
	var items []CompletionItem

	cls := d.findClass(typeName, classes)
	if cls == nil {
		return nil
	}

	for _, f := range cls.Fields {
		ktype := typeExprToString(f.TypeExpr)
		if ktype == "" && f.Inferred {
			ktype = "(inferred)"
		}
		ktype = resolveTypeWithParams(typeName, cls, ktype)
		items = append(items, CompletionItem{
			Label:  f.Name,
			Detail: ktype,
			Kind:   CompletionKindField,
		})
	}
	for _, m := range cls.Methods {
		detail := resolveTypeWithParams(typeName, cls, formatMethodSignature(m))
		items = append(items, CompletionItem{
			Label:  m.Name,
			Detail: detail,
			Kind:   CompletionKindMethod,
		})
	}
	for _, p := range cls.Properties {
		ktype := resolveTypeWithParams(typeName, cls, typeExprToString(p.TypeExpr))
		items = append(items, CompletionItem{
			Label:  p.Name,
			Detail: ktype,
			Kind:   CompletionKindProperty,
		})
	}
	for _, ev := range cls.Events {
		items = append(items, CompletionItem{
			Label:  ev.Name,
			Detail: "event(" + formatEventParams(ev) + ")",
			Kind:   CompletionKindField,
		})
	}

	// Include parent class members
	if cls.Parent != "" {
		items = append(items, d.completeClassMembers(cls.Parent, classes)...)
	}

	return items
}

func (d *Document) completeType() []CompletionItem {
	var items []CompletionItem

	// Built-in types
	for _, t := range BuiltinTypes {
		items = append(items, CompletionItem{
			Label: t,
			Kind:  CompletionKindType,
		})
	}

	// User-defined classes
	classes := d.GetClasses()
	for name := range classes {
		items = append(items, CompletionItem{
			Label: name,
			Kind:  CompletionKindClass,
		})
	}

	return items
}

func (d *Document) completeBare(line int) []CompletionItem {
	var items []CompletionItem

	// Keywords
	for _, kw := range Keywords {
		items = append(items, CompletionItem{
			Label: kw,
			Kind:  CompletionKindKeyword,
		})
	}

	// Module names
	for _, mod := range ModuleNames {
		items = append(items, CompletionItem{
			Label:  mod,
			Detail: "module",
			Kind:   CompletionKindModule,
		})
	}

	// Namespace names
	for _, ns := range NamespaceNames {
		items = append(items, CompletionItem{
			Label:  ns,
			Detail: "namespace",
			Kind:   CompletionKindModule,
		})
	}

	// User-defined class names
	classes := d.GetClasses()
	for name := range classes {
		items = append(items, CompletionItem{
			Label: name,
			Kind:  CompletionKindClass,
		})
	}

	// Context-sensitive completions
	cls, fullName := d.FindEnclosingClass(line)
	if cls != nil {
		// Current class fields and methods
		for _, f := range cls.Fields {
			ktype := d.ResolveFieldType(f, fullName)
			items = append(items, CompletionItem{
				Label:  f.Name,
				Detail: ktype,
				Kind:   CompletionKindField,
			})
		}
		for _, m := range cls.Methods {
			detail := formatMethodSignature(m)
			items = append(items, CompletionItem{
				Label:  m.Name,
				Detail: detail,
				Kind:   CompletionKindMethod,
			})
		}
		for _, ev := range cls.Events {
			items = append(items, CompletionItem{
				Label:  ev.Name,
				Detail: "event(" + formatEventParams(ev) + ")",
				Kind:   CompletionKindField,
			})
		}

		// Local variables (with resolved inferred types)
		method := d.FindEnclosingMethod(cls, line)
		if method != nil {
			locals := d.CollectLocalsBeforeLine(method, line)
			for _, lv := range locals {
				items = append(items, CompletionItem{
					Label:  lv.Name,
					Detail: lv.KType,
					Kind:   CompletionKindVariable,
				})
			}

			// With-module bare functions
			mods := CollectWithModulesAtLine(method, line)
			for _, mod := range mods {
				if sigs, ok := StdlibModuleSignatures[mod]; ok {
					for _, sig := range sigs {
						items = append(items, CompletionItem{
							Label:  sig.Name,
							Detail: sig.Detail,
							Kind:   CompletionKindFunction,
						})
					}
				}
				// Also add module constants in with scope
				if consts, ok := StdlibModuleConstantNames[mod]; ok {
					for _, c := range consts {
						items = append(items, CompletionItem{
							Label:  c.Name,
							Detail: c.Detail,
							Kind:   CompletionKindConstant,
						})
					}
				}
			}
		}
	}

	// Built-in type cast functions
	castFuncs := []struct{ name, detail string }{
		{"int", "int(value):int"},
		{"float", "float(value):float"},
		{"bool", "bool(value):bool"},
		{"string", "string(value):string"},
	}
	for _, c := range castFuncs {
		items = append(items, CompletionItem{
			Label:  c.name,
			Detail: c.detail,
			Kind:   CompletionKindFunction,
		})
	}

	return items
}

// formatEventParams returns a comma-separated parameter list for an event.
func formatEventParams(ev *parser.EventDecl) string {
	var params []string
	for _, p := range ev.Params {
		s := p.Name
		if p.TypeExpr != nil {
			s += ":" + typeExprToString(p.TypeExpr)
		}
		params = append(params, s)
	}
	return strings.Join(params, ", ")
}

// completeEventMembers returns completions for event methods.
func (d *Document) completeEventMembers(ev *parser.EventDecl) []CompletionItem {
	emitDetail := formatEventSignature(ev)
	return []CompletionItem{
		{Label: "connect", Detail: "connect(handler)", Kind: CompletionKindMethod},
		{Label: "disconnect", Detail: "disconnect(handler)", Kind: CompletionKindMethod},
		{Label: "emit", Detail: emitDetail, Kind: CompletionKindMethod},
	}
}

// formatEventSignature returns a display string for an event's emit signature.
func formatEventSignature(ev *parser.EventDecl) string {
	var params []string
	for _, p := range ev.Params {
		s := p.Name
		if p.TypeExpr != nil {
			s += ":" + typeExprToString(p.TypeExpr)
		}
		params = append(params, s)
	}
	return "emit(" + strings.Join(params, ", ") + ")"
}

func formatMethodSignature(m *parser.MethodDecl) string {
	var params []string
	for _, p := range m.Params {
		s := p.Name
		if p.TypeExpr != nil {
			s += ":" + typeExprToString(p.TypeExpr)
		}
		params = append(params, s)
	}
	sig := m.Name + "(" + strings.Join(params, ", ") + ")"
	if m.ReturnType != nil {
		sig += ":" + typeExprToString(m.ReturnType)
	}
	return sig
}

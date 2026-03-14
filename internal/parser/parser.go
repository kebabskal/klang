package parser

import (
	"fmt"
	"strings"

	"github.com/klang-lang/klang/internal/errs"
	"github.com/klang-lang/klang/internal/lexer"
)

type Parser struct {
	tokens      []lexer.Token
	pos         int
	errors      []string
	diagnostics []errs.Diagnostic
	source      []byte // raw source for error context
	file        string // filename for error context
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// SetSource provides source code and filename for error reporting.
func (p *Parser) SetSource(src []byte, file string) {
	p.source = src
	p.file = file
}

// Diagnostics returns all accumulated diagnostics.
func (p *Parser) Diagnostics() []errs.Diagnostic {
	return p.diagnostics
}

func (p *Parser) Parse() (*File, error) {
	file := &File{}
	p.skipNewlines()

	// Parse namespace if present
	if p.check(lexer.TOKEN_NAMESPACE) {
		p.advance()
		if p.check(lexer.TOKEN_IDENT) {
			file.Namespace = p.current().Value
			p.advance()
		}
		p.skipNewlines()
	}

	// If the file doesn't start with a class declaration (Ident:class),
	// wrap everything in an implicit Main class (for single-file programs).
	if !p.isAtEnd() && !p.looksLikeClassDecl() {
		cls := p.parseClassBody("Main", "", true)
		file.Classes = append(file.Classes, cls)
	}

	// Parse file-scoped and top-level declarations
	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}
		decl := p.parseTopLevel()
		if decl != nil {
			file.Classes = append(file.Classes, decl)
		}
	}

	if len(p.diagnostics) > 0 {
		// Build combined error message from diagnostics
		var msgs []string
		for _, d := range p.diagnostics {
			msgs = append(msgs, d.Format())
		}
		return file, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}
	if len(p.errors) > 0 {
		return file, fmt.Errorf("parse errors:\n%s", p.errors[0])
	}
	return file, nil
}

// looksLikeClassDecl checks if the current position starts with Ident:class
func (p *Parser) looksLikeClassDecl() bool {
	return p.check(lexer.TOKEN_IDENT) &&
		p.peekIs(1, lexer.TOKEN_COLON) &&
		p.peekIs(2, lexer.TOKEN_CLASS)
}

// parseTopLevel handles: Name:class (file-scoped) or Name:class { } (braced)
func (p *Parser) parseTopLevel() *ClassDecl {
	if !p.check(lexer.TOKEN_IDENT) {
		p.error("expected identifier at top level")
		p.advance()
		return nil
	}

	name := p.current().Value
	p.advance()

	if !p.check(lexer.TOKEN_COLON) {
		p.error(fmt.Sprintf("expected ':' after '%s'", name))
		p.advance()
		return nil
	}
	p.advance() // skip ':'

	if p.check(lexer.TOKEN_CLASS) {
		p.advance()
		var typeParams []string
		if p.check(lexer.TOKEN_LT) {
			typeParams = p.parseTypeParamList()
		}
		// If followed by {, it's a braced class; otherwise file-scoped
		fileScope := !p.check(lexer.TOKEN_LBRACE)
		p.skipNewlines()
		cls := p.parseClassBody(name, "", fileScope)
		cls.TypeParams = typeParams
		return cls
	}

	p.error(fmt.Sprintf("expected 'class' after '%s:'", name))
	return nil
}

// parseClassBody parses class members. fileScope=true means no braces, members until next top-level decl.
func (p *Parser) parseClassBody(name, parent string, fileScope bool) *ClassDecl {
	cls := &ClassDecl{
		Name:        name,
		Parent:      parent,
		IsFileScope: fileScope,
	}

	if !fileScope {
		// Expect opening brace
		if !p.expect(lexer.TOKEN_LBRACE) {
			return cls
		}
		p.skipNewlines()
	} else {
		p.skipNewlines()
	}

	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}

		// For braced class, stop at '}'
		if !fileScope && p.check(lexer.TOKEN_RBRACE) {
			p.advance()
			return cls
		}

		// For file-scoped class, stop if we see an unindented identifier followed by ':class'
		// (that would be a new top-level class)
		if fileScope && p.isNewTopLevelDecl() {
			return cls
		}

		// Skip indent tokens
		if p.check(lexer.TOKEN_INDENT) {
			p.advance()
			continue
		}

		p.parseMember(cls)
	}

	return cls
}

// isNewTopLevelDecl checks if current position looks like a new file-scoped class: Ident:class at indent 0
// Only matches if the class is NOT followed by '{' (braced classes are nested members)
func (p *Parser) isNewTopLevelDecl() bool {
	// Check if we're at a non-indented identifier
	if p.pos > 0 {
		prev := p.tokens[p.pos-1]
		if prev.Type == lexer.TOKEN_INDENT {
			return false // indented = not a new top-level decl
		}
	}

	if !p.check(lexer.TOKEN_IDENT) {
		return false
	}

	// Look ahead: Ident : class (NOT followed by '{')
	if p.pos+2 < len(p.tokens) &&
		p.tokens[p.pos+1].Type == lexer.TOKEN_COLON &&
		p.tokens[p.pos+2].Type == lexer.TOKEN_CLASS {
		// If followed by '{', it's a braced nested class, not a new top-level decl
		if p.pos+3 < len(p.tokens) && p.tokens[p.pos+3].Type == lexer.TOKEN_LBRACE {
			return false
		}
		// If followed by '<', it's a generic class (braced), not a new top-level decl
		if p.pos+3 < len(p.tokens) && p.tokens[p.pos+3].Type == lexer.TOKEN_LT {
			return false
		}
		return true
	}
	return false
}

// parseMember parses a single class member (field, method, nested class, enum, property, event, constructor)
func (p *Parser) parseMember(cls *ClassDecl) {
	// Constructor
	if p.check(lexer.TOKEN_CONSTRUCTOR) {
		p.advance()
		ctor := p.parseConstructor()
		cls.Constructor = ctor
		return
	}

	if !p.check(lexer.TOKEN_IDENT) {
		p.error(fmt.Sprintf("expected member declaration, got %s", p.current().Value))
		p.advance()
		return
	}

	name := p.current().Value
	p.advance()

	// name := value (inferred variable/field)
	if p.check(lexer.TOKEN_COLON_EQ) {
		p.advance()
		val := p.parseExpr()
		cls.Fields = append(cls.Fields, &FieldDecl{
			Name:     name,
			Value:    val,
			Inferred: true,
		})
		p.skipNewlines()
		return
	}

	// name = value (override default in subclass)
	if p.check(lexer.TOKEN_EQ) {
		p.advance()
		val := p.parseExpr()
		cls.Fields = append(cls.Fields, &FieldDecl{
			Name:  name,
			Value: val,
		})
		p.skipNewlines()
		return
	}

	// name:Type ...
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		p.parseMemberAfterColon(cls, name)
		return
	}

	// name<T, U>(...) { ... } — generic method
	if p.check(lexer.TOKEN_LT) {
		typeParams := p.parseTypeParamList()
		method := p.parseMethod(name, nil)
		method.TypeParams = typeParams
		cls.Methods = append(cls.Methods, method)
		return
	}

	// name(...) { ... } — method (no return type)
	if p.check(lexer.TOKEN_LPAREN) {
		method := p.parseMethod(name, nil)
		cls.Methods = append(cls.Methods, method)
		return
	}

	p.error(fmt.Sprintf("unexpected token after '%s'", name))
	p.advance()
}

// parseMemberAfterColon handles Name: <class|enum|event|type> ...
func (p *Parser) parseMemberAfterColon(cls *ClassDecl, name string) {
	// Name:class or Name:class<T> — nested class (braced)
	if p.check(lexer.TOKEN_CLASS) {
		p.advance()
		var typeParams []string
		if p.check(lexer.TOKEN_LT) {
			typeParams = p.parseTypeParamList()
		}
		p.skipNewlines()
		nested := p.parseClassBody(name, "", false)
		nested.TypeParams = typeParams
		cls.Classes = append(cls.Classes, nested)
		return
	}

	// Name:enum = { ... }
	if p.check(lexer.TOKEN_ENUM) {
		p.advance()
		en := p.parseEnum(name)
		cls.Enums = append(cls.Enums, en)
		return
	}

	// Name:event<...>
	if p.check(lexer.TOKEN_EVENT) {
		p.advance()
		ev := p.parseEvent(name)
		cls.Events = append(cls.Events, ev)
		return
	}

	// Everything else starts with a type name (or fn keyword for function types)
	if !p.check(lexer.TOKEN_IDENT) && !p.check(lexer.TOKEN_FN) {
		p.error(fmt.Sprintf("unexpected token after '%s:'", name))
		p.advance()
		return
	}

	// Parse the full type expression (handles generics and unions)
	typeExpr := p.parseTypeExpr()

	// Now decide what this member is based on what follows the type

	// Name:Parent { ... } — inheritance (subclass)
	// Only if typeExpr is a SimpleType (not generic, not union)
	if st, ok := typeExpr.(*SimpleType); ok && p.check(lexer.TOKEN_LBRACE) {
		// Peek inside to see if this looks like a class body or a property
		if p.looksLikeClassBody() {
			nested := p.parseClassBody(name, st.Name, false)
			cls.Classes = append(cls.Classes, nested)
			return
		}
	}

	// Name:type { get => ... set(value) => ... } — property
	if p.check(lexer.TOKEN_LBRACE) {
		prop := p.parseProperty(name, typeExpr)
		cls.Properties = append(cls.Properties, prop)
		return
	}

	// Name:type = value — field with type and default
	if p.check(lexer.TOKEN_EQ) {
		p.advance()
		val := p.parseExpr()
		cls.Fields = append(cls.Fields, &FieldDecl{
			Name:     name,
			TypeExpr: typeExpr,
			Value:    val,
		})
		p.skipNewlines()
		return
	}

	// Name:type (field, no default — end of line)
	cls.Fields = append(cls.Fields, &FieldDecl{
		Name:     name,
		TypeExpr: typeExpr,
	})
	p.skipNewlines()
}

// looksLikeClassBody peeks into a { ... } to see if it contains member declarations (class body)
// vs get/set (property). Does NOT advance the parser.
func (p *Parser) looksLikeClassBody() bool {
	// Save position and scan ahead
	saved := p.pos
	defer func() { p.pos = saved }()

	p.advance() // skip {
	p.skipNewlines()
	p.skipIndent()

	// If first thing is 'get' or 'set', it's a property
	if p.check(lexer.TOKEN_GET) || p.check(lexer.TOKEN_SET) {
		return false
	}
	return true
}

func (p *Parser) parseProperty(name string, typeExpr TypeExpr) *PropertyDecl {
	prop := &PropertyDecl{Name: name, TypeExpr: typeExpr}
	p.expect(lexer.TOKEN_LBRACE)
	p.skipNewlines()

	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_GET) {
			p.advance()
			p.expect(lexer.TOKEN_ARROW) // =>
			prop.Getter = p.parseExpr()
		} else if p.check(lexer.TOKEN_SET) {
			p.advance()
			// set(value) => { ... }
			if p.check(lexer.TOKEN_LPAREN) {
				p.advance()
				// skip param name
				if p.check(lexer.TOKEN_IDENT) {
					p.advance()
				}
				p.expect(lexer.TOKEN_RPAREN)
			}
			p.expect(lexer.TOKEN_ARROW) // =>
			p.skipNewlines()
			prop.Setter = p.parseBlock()
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	p.skipNewlines()
	return prop
}

func (p *Parser) buildTypeExpr(name string) TypeExpr {
	// Check for generics: Type<...>
	if p.check(lexer.TOKEN_LT) {
		p.advance()
		var typeArgs []TypeExpr
		for !p.check(lexer.TOKEN_GT) && !p.isAtEnd() {
			typeArgs = append(typeArgs, p.parseTypeExpr())
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_GT)
		return &GenericType{Name: name, TypeArgs: typeArgs}
	}

	// Check for union type: Type|Type
	if p.check(lexer.TOKEN_PIPE) {
		types := []TypeExpr{&SimpleType{Name: name}}
		for p.check(lexer.TOKEN_PIPE) {
			p.advance()
			types = append(types, p.parseSimpleTypeExpr())
		}
		return &UnionType{Types: types}
	}

	return &SimpleType{Name: name}
}

func (p *Parser) parseTypeExpr() TypeExpr {
	if p.check(lexer.TOKEN_CLASS) {
		// inline class type: class { delta:int }
		p.advance()
		return p.parseInlineClassType()
	}
	if p.check(lexer.TOKEN_FN) {
		return p.parseFnType()
	}
	if !p.check(lexer.TOKEN_IDENT) {
		p.error("expected type name")
		return &SimpleType{Name: "unknown"}
	}
	name := p.current().Value
	p.advance()
	return p.buildTypeExpr(name)
}

func (p *Parser) parseFnType() TypeExpr {
	p.advance() // skip 'fn'
	ft := &FnType{}
	// bare 'fn' without parens = fn() (no params, void return)
	if !p.check(lexer.TOKEN_LPAREN) {
		return ft
	}
	p.advance() // skip '('
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		ft.ParamTypes = append(ft.ParamTypes, p.parseTypeExpr())
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	// Optional return type: fn(int):int
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		ft.ReturnType = p.parseTypeExpr()
	}
	return ft
}

func (p *Parser) parseSimpleTypeExpr() TypeExpr {
	if !p.check(lexer.TOKEN_IDENT) {
		p.error("expected type name")
		return &SimpleType{Name: "unknown"}
	}
	name := p.current().Value
	p.advance()
	// Check for generics on the right side of union too
	if p.check(lexer.TOKEN_LT) {
		p.advance()
		var typeArgs []TypeExpr
		for !p.check(lexer.TOKEN_GT) && !p.isAtEnd() {
			typeArgs = append(typeArgs, p.parseTypeExpr())
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_GT)
		return &GenericType{Name: name, TypeArgs: typeArgs}
	}
	return &SimpleType{Name: name}
}

func (p *Parser) parseInlineClassType() TypeExpr {
	if !p.expect(lexer.TOKEN_LBRACE) {
		return &InlineClassType{}
	}
	p.skipNewlines()
	var fields []*FieldDecl
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_IDENT) {
			name := p.current().Value
			p.advance()
			if p.check(lexer.TOKEN_COLON) {
				p.advance()
				typeExpr := p.parseTypeExpr()
				fields = append(fields, &FieldDecl{Name: name, TypeExpr: typeExpr})
			}
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &InlineClassType{Fields: fields}
}

func (p *Parser) parseConstructor() *ConstructorDecl {
	ctor := &ConstructorDecl{}
	ctor.Params = p.parseParams()
	p.skipNewlines()
	ctor.Body = p.parseBlock()
	return ctor
}

func (p *Parser) parseTypeParamList() []string {
	p.advance() // skip <
	var params []string
	for !p.check(lexer.TOKEN_GT) && !p.isAtEnd() {
		if p.check(lexer.TOKEN_IDENT) {
			params = append(params, p.current().Value)
			p.advance()
		}
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_GT) // skip >
	return params
}

func (p *Parser) parseMethod(name string, returnType TypeExpr) *MethodDecl {
	method := &MethodDecl{
		Name:       name,
		ReturnType: returnType,
	}

	// Check for spread: act(...)
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance()
		if p.check(lexer.TOKEN_ELLIPSIS) {
			method.IsSpread = true
			p.advance()
			p.expect(lexer.TOKEN_RPAREN)
		} else {
			// Parse params
			p.pos-- // back up to re-parse with parseParams
			method.Params = p.parseParams()
		}
	}

	// Check for return type: ):ReturnType
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		method.ReturnType = p.parseTypeExpr()
	}

	p.skipNewlines()
	method.Body = p.parseBlock()
	return method
}

func (p *Parser) parseParams() []*Param {
	var params []*Param
	p.expect(lexer.TOKEN_LPAREN)
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		param := &Param{}
		if p.check(lexer.TOKEN_IDENT) {
			param.Name = p.current().Value
			p.advance()
		}
		if p.check(lexer.TOKEN_COLON) {
			p.advance()
			param.TypeExpr = p.parseTypeExpr()
		}
		params = append(params, param)
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	return params
}

func (p *Parser) parseEnum(name string) *EnumDecl {
	en := &EnumDecl{Name: name}
	p.skipNewlines()
	// = { ... }
	if p.check(lexer.TOKEN_EQ) {
		p.advance()
	}
	p.skipNewlines()
	p.expect(lexer.TOKEN_LBRACE)
	p.skipNewlines()
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_IDENT) {
			member := &EnumMember{Name: p.current().Value}
			p.advance()
			if p.check(lexer.TOKEN_EQ) {
				p.advance()
				member.Value = p.parseExpr()
			}
			en.Members = append(en.Members, member)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	p.skipNewlines()
	return en
}

func (p *Parser) parseEvent(name string) *EventDecl {
	ev := &EventDecl{Name: name}
	// event<Type>
	if p.check(lexer.TOKEN_LT) {
		p.advance()
		ev.TypeExpr = p.parseTypeExpr()
		p.expect(lexer.TOKEN_GT)
	}
	p.skipNewlines()
	return ev
}

// parseBlock parses { stmts }
func (p *Parser) parseBlock() *Block {
	block := &Block{}
	if !p.expect(lexer.TOKEN_LBRACE) {
		return block
	}
	p.skipNewlines()
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_RBRACE) {
			break
		}
		stmt := p.parseStmt()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return block
}

func (p *Parser) parseStmt() Stmt {
	// return
	if p.check(lexer.TOKEN_RETURN) {
		return p.parseReturn()
	}

	// if
	if p.check(lexer.TOKEN_IF) {
		return p.parseIf()
	}

	// for
	if p.check(lexer.TOKEN_FOR) {
		return p.parseFor()
	}

	// while
	if p.check(lexer.TOKEN_WHILE) {
		return p.parseWhile()
	}

	// with module { block } or with module (rest of enclosing block)
	if p.check(lexer.TOKEN_WITH) {
		return p.parseWith()
	}

	// @c { raw C code }
	if p.check(lexer.TOKEN_INLINE_C) {
		code := p.current().Value
		p.advance()
		return &InlineCStmt{Code: code}
	}

	// Variable declaration: name := expr (inferred type)
	if p.check(lexer.TOKEN_IDENT) && p.peekIs(1, lexer.TOKEN_COLON_EQ) {
		name := p.current().Value
		p.advance()
		p.advance() // skip :=
		val := p.parseExpr()
		return &VarDecl{Name: name, Value: val, Inferred: true}
	}

	// Variable declaration: name: = expr (colon space equals, inferred type)
	if p.check(lexer.TOKEN_IDENT) && p.peekIs(1, lexer.TOKEN_COLON) && p.peekIs(2, lexer.TOKEN_EQ) {
		name := p.current().Value
		p.advance()
		p.advance() // skip :
		p.advance() // skip =
		val := p.parseExpr()
		return &VarDecl{Name: name, Value: val, Inferred: true}
	}

	// Typed variable declaration: name:Type = expr
	if p.check(lexer.TOKEN_IDENT) && p.peekIs(1, lexer.TOKEN_COLON) && (p.peekIs(2, lexer.TOKEN_IDENT) || p.peekIs(2, lexer.TOKEN_FN)) {
		name := p.current().Value
		p.advance()
		p.advance() // skip :
		typeExpr := p.parseTypeExpr()
		if p.check(lexer.TOKEN_EQ) {
			p.advance()
			val := p.parseExpr()
			return &VarDecl{Name: name, TypeExpr: typeExpr, Value: val}
		}
		// Declaration without initializer
		return &VarDecl{Name: name, TypeExpr: typeExpr}
	}

	// expression statement (which may become an assignment)
	expr := p.parseExpr()

	// Check for assignment: expr = value, expr += value, etc.
	if p.check(lexer.TOKEN_EQ) || p.check(lexer.TOKEN_PLUS_EQ) || p.check(lexer.TOKEN_MINUS_EQ) ||
		p.check(lexer.TOKEN_STAR_EQ) || p.check(lexer.TOKEN_SLASH_EQ) {
		op := p.current().Value
		p.advance()
		val := p.parseExpr()
		return &AssignStmt{Target: expr, Op: op, Value: val}
	}

	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseReturn() Stmt {
	p.advance() // skip 'return'

	// Check for bare return (next is newline, } or EOF)
	if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_RBRACE) || p.isAtEnd() {
		return &ReturnStmt{}
	}

	val := p.parseExpr()
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIf() Stmt {
	p.advance() // skip 'if'

	cond := p.parseExpr()
	ifStmt := &IfStmt{Condition: cond}

	// Single-statement if: "if not is_alive return"
	if p.check(lexer.TOKEN_RETURN) {
		ifStmt.ThenStmt = p.parseReturn()
		return ifStmt
	}

	p.skipNewlines()
	ifStmt.Then = p.parseBlock()
	p.skipNewlines()

	// else
	if p.check(lexer.TOKEN_INDENT) {
		p.advance()
	}
	if p.check(lexer.TOKEN_ELSE) {
		p.advance()
		p.skipNewlines()
		if p.check(lexer.TOKEN_IF) {
			ifStmt.Else = p.parseIf().(*IfStmt)
		} else {
			ifStmt.Else = p.parseBlock()
		}
	}

	return ifStmt
}

func (p *Parser) parseFor() Stmt {
	p.advance() // skip 'for'
	varName := ""
	if p.check(lexer.TOKEN_IDENT) {
		varName = p.current().Value
		p.advance()
	}
	p.expect(lexer.TOKEN_IN)
	iter := p.parseExpr()
	p.skipNewlines()
	body := p.parseBlock()
	return &ForStmt{VarName: varName, Iterable: iter, Body: body}
}

func (p *Parser) parseWhile() Stmt {
	p.advance() // skip 'while'
	cond := p.parseExpr()
	p.skipNewlines()
	body := p.parseBlock()
	return &WhileStmt{Condition: cond, Body: body}
}

func (p *Parser) parseWith() Stmt {
	p.advance() // skip 'with'
	if !p.check(lexer.TOKEN_IDENT) {
		p.error("expected module name after 'with'")
		return &ExprStmt{Expr: &Ident{Name: "with"}}
	}
	module := p.current().Value
	p.advance()
	p.skipNewlines()

	// Block form: with module { ... }
	if p.check(lexer.TOKEN_LBRACE) {
		body := p.parseBlock()
		return &WithStmt{Module: module, Body: body}
	}

	// Bare form: with module — applies to rest of enclosing block
	var stmts []Stmt
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_RBRACE) {
			break
		}
		stmt := p.parseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipNewlines()
	}
	return &WithStmt{Module: module, Body: &Block{Stmts: stmts}}
}

// --- Expression parsing (precedence climbing) ---

func (p *Parser) parseExpr() Expr {
	return p.parseOr()
}

func (p *Parser) parseOr() Expr {
	left := p.parseAnd()
	for p.check(lexer.TOKEN_OR) || p.check(lexer.TOKEN_PIPEPIPE) {
		op := p.current().Value
		p.advance()
		right := p.parseAnd()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Expr {
	left := p.parseEquality()
	for p.check(lexer.TOKEN_AND) || p.check(lexer.TOKEN_AMPAMP) {
		op := p.current().Value
		p.advance()
		right := p.parseEquality()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() Expr {
	left := p.parseComparison()
	for p.check(lexer.TOKEN_EQEQ) || p.check(lexer.TOKEN_NEQ) {
		op := p.current().Value
		p.advance()
		right := p.parseComparison()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() Expr {
	left := p.parseNullCoalesce()
	for p.check(lexer.TOKEN_LT) || p.check(lexer.TOKEN_GT) || p.check(lexer.TOKEN_LTEQ) || p.check(lexer.TOKEN_GTEQ) {
		op := p.current().Value
		p.advance()
		right := p.parseNullCoalesce()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	// 'is' type check
	if p.check(lexer.TOKEN_IS) {
		p.advance()
		if p.check(lexer.TOKEN_IDENT) {
			typeName := p.current().Value
			p.advance()
			return &IsExpr{Expr: left, TypeName: typeName}
		}
	}
	return left
}

func (p *Parser) parseNullCoalesce() Expr {
	left := p.parseAddition()

	// Handle optional chaining continuation (after newline + indent + ?.)
	for {
		// Check if next meaningful tokens are ?.  (possibly after newline+indent)
		savedPos := p.pos
		p.skipNewlines()
		p.skipIndent()
		if p.check(lexer.TOKEN_QUESTION_DOT) || p.check(lexer.TOKEN_QUESTION_QUESTION) {
			break // let the loop below handle it
		}
		p.pos = savedPos
		break
	}

	for p.check(lexer.TOKEN_QUESTION_QUESTION) {
		p.advance()
		p.skipNewlines()
		p.skipIndent()
		right := p.parseAddition()
		left = &NullCoalesce{Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAddition() Expr {
	left := p.parseMultiplication()
	for p.check(lexer.TOKEN_PLUS) || p.check(lexer.TOKEN_MINUS) {
		op := p.current().Value
		p.advance()
		right := p.parseMultiplication()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplication() Expr {
	left := p.parseUnary()
	for p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) {
		op := p.current().Value
		p.advance()
		right := p.parseUnary()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() Expr {
	if p.check(lexer.TOKEN_NOT) {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: "not", Operand: operand}
	}
	if p.check(lexer.TOKEN_MINUS) {
		p.advance()
		operand := p.parseUnary()
		return &UnaryExpr{Op: "-", Operand: operand}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() Expr {
	expr := p.parsePrimary()

	for {
		// Check for continuation on next line: newline+indent followed by ?. or ??
		if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_INDENT) {
			saved := p.pos
			p.skipNewlines()
			p.skipIndent()
			if p.check(lexer.TOKEN_QUESTION_DOT) || p.check(lexer.TOKEN_QUESTION_QUESTION) {
				// continuation — don't restore, let the loop handle it
			} else {
				p.pos = saved
				break
			}
		}

		if p.check(lexer.TOKEN_DOT) {
			p.advance()
			if p.check(lexer.TOKEN_IDENT) {
				field := p.current().Value
				p.advance()
				expr = &MemberExpr{Object: expr, Field: field}
			}
			continue
		}
		if p.check(lexer.TOKEN_QUESTION_DOT) {
			p.advance()
			if p.check(lexer.TOKEN_IDENT) {
				field := p.current().Value
				p.advance()
				expr = &MemberExpr{Object: expr, Field: field, Optional: true}
			}
			continue
		}
		if p.check(lexer.TOKEN_QUESTION_QUESTION) {
			p.advance()
			p.skipNewlines()
			p.skipIndent()
			right := p.parsePostfix()
			expr = &NullCoalesce{Left: expr, Right: right}
			continue
		}
		// ident<Type, Type>(args) — generic call
		if _, ok := expr.(*Ident); ok && p.check(lexer.TOKEN_LT) {
			if typeArgs, ok := p.tryParseTypeArgList(); ok {
				call := p.parseCall(expr)
				call.(*CallExpr).TypeArgs = typeArgs
				expr = call
				continue
			}
		}
		if p.check(lexer.TOKEN_LPAREN) {
			expr = p.parseCall(expr)
			continue
		}
		break
	}
	return expr
}

// tryParseTypeArgList attempts to parse <Type, Type...> followed by (.
// Returns the type args and true on success, or restores position and returns false.
func (p *Parser) tryParseTypeArgList() ([]TypeExpr, bool) {
	saved := p.pos
	p.advance() // skip <
	var typeArgs []TypeExpr
	for !p.check(lexer.TOKEN_GT) && !p.isAtEnd() {
		if p.check(lexer.TOKEN_IDENT) {
			typeArgs = append(typeArgs, p.parseTypeExpr())
		} else {
			p.pos = saved
			return nil, false
		}
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	if !p.check(lexer.TOKEN_GT) {
		p.pos = saved
		return nil, false
	}
	p.advance() // skip >
	// Must be followed by ( to be a generic call
	if !p.check(lexer.TOKEN_LPAREN) {
		p.pos = saved
		return nil, false
	}
	return typeArgs, true
}

func (p *Parser) parseCall(callee Expr) Expr {
	p.advance() // skip (
	var args []Expr
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		if p.check(lexer.TOKEN_ELLIPSIS) {
			args = append(args, &SpreadExpr{})
			p.advance()
		} else {
			args = append(args, p.parseExpr())
		}
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	return &CallExpr{Callee: callee, Args: args}
}

func (p *Parser) parsePrimary() Expr {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_FN:
		// fn { ... } or fn(params):retType { ... }
		return p.parseFnLambda()
	case lexer.TOKEN_INT_LIT:
		p.advance()
		return &IntLit{Value: tok.Value}
	case lexer.TOKEN_FLOAT_LIT:
		p.advance()
		return &FloatLit{Value: tok.Value}
	case lexer.TOKEN_STRING_LIT:
		p.advance()
		return &StringLit{Value: tok.Value}
	case lexer.TOKEN_INTERP_STRING:
		p.advance()
		return &InterpString{Parts: []Expr{&StringLit{Value: tok.Value}}}
	case lexer.TOKEN_TRUE:
		p.advance()
		return &BoolLit{Value: true}
	case lexer.TOKEN_FALSE:
		p.advance()
		return &BoolLit{Value: false}
	case lexer.TOKEN_THIS:
		p.advance()
		return &ThisExpr{}
	case lexer.TOKEN_IDENT:
		p.advance()
		return &Ident{Name: tok.Value}
	case lexer.TOKEN_LBRACE:
		return p.parseStructLit()
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLit()
	case lexer.TOKEN_LPAREN:
		// Try lambda: () => { ... } or (x, y) => { ... } or (x:int) => { ... }
		if lambda := p.tryParseLambda(); lambda != nil {
			return lambda
		}
		p.advance()
		expr := p.parseExpr()
		p.expect(lexer.TOKEN_RPAREN)
		return expr
	}

	p.error(fmt.Sprintf("unexpected token in expression: %s (%q)", tokenName(tok.Type), tok.Value))
	p.advance()
	return &Ident{Name: "_error_"}
}

func (p *Parser) parseStructLit() Expr {
	p.advance() // skip {
	p.skipNewlines()
	p.skipIndent()

	lit := &StructLit{}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		p.skipIndent()

		field := &StructLitField{}
		// Check if it's named: ident = expr or ident: expr
		if p.check(lexer.TOKEN_IDENT) && (p.peekIs(1, lexer.TOKEN_EQ) || p.peekIs(1, lexer.TOKEN_COLON)) {
			field.Name = p.current().Value
			p.advance()
			p.advance() // skip = or :
			field.Value = p.parseExpr()
		} else {
			field.Value = p.parseExpr()
		}
		lit.Fields = append(lit.Fields, field)

		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return lit
}

func (p *Parser) parseArrayLit() Expr {
	p.advance() // skip [
	p.skipNewlines()
	p.skipIndent()

	lit := &ArrayLit{}
	for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
		p.skipIndent()
		if p.check(lexer.TOKEN_RBRACKET) {
			break
		}
		lit.Elements = append(lit.Elements, p.parseExpr())
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return lit
}

// --- Lambda parsing ---

// parseFnLambda parses fn { ... } or fn(params):retType { ... }
func (p *Parser) parseFnLambda() Expr {
	p.advance() // skip 'fn'
	lambda := &LambdaExpr{}
	// Optional params: fn(x:int, y:int) { ... }
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance()
		for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
			param := &Param{}
			if p.check(lexer.TOKEN_IDENT) {
				param.Name = p.current().Value
				p.advance()
			}
			if p.check(lexer.TOKEN_COLON) {
				p.advance()
				param.TypeExpr = p.parseTypeExpr()
			}
			lambda.Params = append(lambda.Params, param)
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_RPAREN)
	}
	// Optional return type: fn(x:int):int { ... }
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		lambda.ReturnType = p.parseTypeExpr()
	}
	p.skipNewlines()
	lambda.Body = p.parseBlock()
	return lambda
}

// tryParseLambda attempts to parse an arrow lambda: () => { ... } or (x, y) => { ... }
// Returns nil if it's not a lambda (restores position).
func (p *Parser) tryParseLambda() Expr {
	saved := p.pos
	p.advance() // skip (

	// Parse potential parameter list
	var params []*Param
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		param := &Param{}
		if p.check(lexer.TOKEN_IDENT) {
			param.Name = p.current().Value
			p.advance()
		} else {
			p.pos = saved
			return nil
		}
		if p.check(lexer.TOKEN_COLON) {
			p.advance()
			param.TypeExpr = p.parseTypeExpr()
		}
		params = append(params, param)
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
	}
	if !p.check(lexer.TOKEN_RPAREN) {
		p.pos = saved
		return nil
	}
	p.advance() // skip )

	// Must be followed by => to be a lambda
	if !p.check(lexer.TOKEN_ARROW) {
		p.pos = saved
		return nil
	}
	p.advance() // skip =>

	p.skipNewlines()
	body := p.parseBlock()
	return &LambdaExpr{Params: params, Body: body}
}

// --- Helpers ---

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) check(typ lexer.TokenType) bool {
	return p.current().Type == typ
}

func (p *Parser) peekIs(offset int, typ lexer.TokenType) bool {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return false
	}
	return p.tokens[idx].Type == typ
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) expect(typ lexer.TokenType) bool {
	if p.check(typ) {
		p.advance()
		return true
	}
	p.error(fmt.Sprintf("expected %s, got %s (%q)", tokenName(typ), tokenName(p.current().Type), p.current().Value))
	return false
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.current().Type == lexer.TOKEN_EOF
}

func (p *Parser) skipNewlines() {
	for p.check(lexer.TOKEN_NEWLINE) {
		p.advance()
	}
}

func (p *Parser) skipIndent() {
	for p.check(lexer.TOKEN_INDENT) {
		p.advance()
	}
}

func (p *Parser) error(msg string) {
	tok := p.current()
	line := tok.Line
	col := tok.Col

	// For EOF or zero-position tokens, use the last real token position
	if (line == 0 || tok.Type == lexer.TOKEN_EOF) && p.pos > 0 {
		for i := p.pos - 1; i >= 0; i-- {
			if p.tokens[i].Line > 0 {
				line = p.tokens[i].Line
				col = p.tokens[i].Col + len(p.tokens[i].Value)
				break
			}
		}
	}

	sourceLine := ""
	if p.source != nil && line > 0 {
		sourceLine = errs.GetSourceLine(p.source, line)
	}
	fileName := p.file
	if fileName == "" {
		fileName = "<input>"
	}
	endCol := col + len(tok.Value)
	if len(tok.Value) == 0 {
		endCol = col + 1
	}
	p.diagnostics = append(p.diagnostics, errs.Diagnostic{
		File:    fileName,
		Line:    line,
		Col:     col,
		EndCol:  endCol,
		Kind:    errs.Error,
		Message: msg,
		Source:  sourceLine,
	})
}

func tokenName(t lexer.TokenType) string {
	names := map[lexer.TokenType]string{
		lexer.TOKEN_EOF:       "EOF",
		lexer.TOKEN_NEWLINE:   "newline",
		lexer.TOKEN_INDENT:    "indent",
		lexer.TOKEN_IDENT:     "identifier",
		lexer.TOKEN_INT_LIT:   "integer",
		lexer.TOKEN_FLOAT_LIT: "float",
		lexer.TOKEN_STRING_LIT: "string",
		lexer.TOKEN_COLON:     "':'",
		lexer.TOKEN_COLON_EQ:  "':='",
		lexer.TOKEN_EQ:        "'='",
		lexer.TOKEN_LBRACE:    "'{'",
		lexer.TOKEN_RBRACE:    "'}'",
		lexer.TOKEN_LPAREN:    "'('",
		lexer.TOKEN_RPAREN:    "')'",
		lexer.TOKEN_LBRACKET:  "'['",
		lexer.TOKEN_RBRACKET:  "']'",
		lexer.TOKEN_LT:        "'<'",
		lexer.TOKEN_GT:        "'>'",
		lexer.TOKEN_CLASS:     "'class'",
		lexer.TOKEN_ENUM:      "'enum'",
		lexer.TOKEN_EVENT:     "'event'",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("token(%d)", t)
}

package parser

// Node is the interface all AST nodes implement.
type Node interface {
	nodeTag()
}

// Pos represents a source position for LSP support.
type Pos struct {
	Line   int
	Col    int
	EndCol int
}

// --- Top-level ---

type File struct {
	Namespace string
	Classes   []*ClassDecl
}

func (f *File) nodeTag() {}

type ClassDecl struct {
	Name       string
	TypeParams []string // generic type params: <T, U>
	Parent     string   // empty if no parent
	IsFileScope bool    // true = no braces, members at file indent level
	Fields     []*FieldDecl
	Methods    []*MethodDecl
	Constructor *ConstructorDecl
	Classes    []*ClassDecl // nested classes
	Enums      []*EnumDecl
	Properties []*PropertyDecl
	Events     []*EventDecl
	Pos        Pos
	EndLine    int // closing brace line for braced classes (0 = unknown/file-scoped)
}

func (c *ClassDecl) nodeTag() {}

// FindField returns the field with the given name, or nil.
func (c *ClassDecl) FindField(name string) *FieldDecl {
	for _, f := range c.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// FindMethod returns the method with the given name, or nil.
func (c *ClassDecl) FindMethod(name string) *MethodDecl {
	for _, m := range c.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// FindProperty returns the property with the given name, or nil.
func (c *ClassDecl) FindProperty(name string) *PropertyDecl {
	for _, p := range c.Properties {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// FindEvent returns the event with the given name, or nil.
func (c *ClassDecl) FindEvent(name string) *EventDecl {
	for _, ev := range c.Events {
		if ev.Name == name {
			return ev
		}
	}
	return nil
}

// HasMember returns true if the class has a field, method, property, or event with the given name.
func (c *ClassDecl) HasMember(name string) bool {
	return c.FindField(name) != nil || c.FindMethod(name) != nil ||
		c.FindProperty(name) != nil || c.FindEvent(name) != nil
}

type FieldDecl struct {
	Name     string
	TypeExpr TypeExpr // nil if inferred
	Value    Expr     // nil if no default
	Inferred bool     // := syntax
	Pos      Pos
}

func (f *FieldDecl) nodeTag() {}

type MethodDecl struct {
	Name       string
	TypeParams []string // generic type params: <T, U>
	Params     []*Param
	ReturnType TypeExpr // nil = void
	Body       *Block
	IsSpread   bool // act(...) — inherits parent sig
	Pos        Pos
}

func (m *MethodDecl) nodeTag() {}

type ConstructorDecl struct {
	Params []*Param
	Body   *Block
	Pos    Pos
}

func (c *ConstructorDecl) nodeTag() {}

type Param struct {
	Name     string
	TypeExpr TypeExpr
	Pos      Pos
}

type EnumDecl struct {
	Name    string
	Members []*EnumMember
	Pos     Pos
}

func (e *EnumDecl) nodeTag() {}

type EnumMember struct {
	Name  string
	Value Expr // nil = auto
	Pos   Pos
}

type PropertyDecl struct {
	Name      string
	TypeExpr  TypeExpr
	Getter    Expr   // expression for get =>
	Setter    *Block // block for set(value) => { ... }
	SetParam  string // parameter name for setter (default "value")
	Pos       Pos
}

func (p *PropertyDecl) nodeTag() {}

type EventDecl struct {
	Name   string
	Params []*Param // event payload parameters: event(value:int, msg:string)
	Pos    Pos
}

func (e *EventDecl) nodeTag() {}

// --- Type expressions ---

type TypeExpr interface {
	typeExprTag()
}

type SimpleType struct {
	Name string
}

func (s *SimpleType) typeExprTag() {}

type GenericType struct {
	Name     string
	TypeArgs []TypeExpr
}

func (g *GenericType) typeExprTag() {}

type UnionType struct {
	Types []TypeExpr
}

func (u *UnionType) typeExprTag() {}

type InlineClassType struct {
	Fields []*FieldDecl
}

func (i *InlineClassType) typeExprTag() {}

type FnType struct {
	ParamTypes []TypeExpr
	ReturnType TypeExpr // nil = void
}

func (f *FnType) typeExprTag() {}

// --- Statements ---

type Stmt interface {
	Node
	stmtTag()
}

type Block struct {
	Stmts []Stmt
}

func (b *Block) nodeTag() {}
func (b *Block) stmtTag() {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) nodeTag() {}
func (e *ExprStmt) stmtTag() {}

type ReturnStmt struct {
	Value Expr // nil = void return
}

func (r *ReturnStmt) nodeTag() {}
func (r *ReturnStmt) stmtTag() {}

type BreakStmt struct{}

func (b *BreakStmt) nodeTag() {}
func (b *BreakStmt) stmtTag() {}

type ContinueStmt struct{}

func (c *ContinueStmt) nodeTag() {}
func (c *ContinueStmt) stmtTag() {}

type IfStmt struct {
	Condition Expr
	Then      *Block
	Else      Stmt // *Block or *IfStmt or nil
	// Single-line form: "if not is_alive return"
	ThenStmt  Stmt // used for single-statement if (no braces)
}

func (i *IfStmt) nodeTag() {}
func (i *IfStmt) stmtTag() {}

type ForStmt struct {
	VarName    string
	ValueVar   string // second variable for "for key, value in dict"
	Iterable   Expr
	Body       *Block
	Pos        Pos
}

func (f *ForStmt) nodeTag() {}
func (f *ForStmt) stmtTag() {}

type WhileStmt struct {
	Condition Expr
	Body      *Block
}

func (w *WhileStmt) nodeTag() {}
func (w *WhileStmt) stmtTag() {}

type VarDecl struct {
	Name     string
	TypeExpr TypeExpr // nil if inferred
	Value    Expr
	Inferred bool // := syntax
	Pos      Pos
}

func (v *VarDecl) nodeTag() {}
func (v *VarDecl) stmtTag() {}

type AssignStmt struct {
	Target Expr
	Op     string // "=", "+=", "-=", "*=", "/="
	Value  Expr
}

func (a *AssignStmt) nodeTag() {}
func (a *AssignStmt) stmtTag() {}

type WithStmt struct {
	Module string
	Body   *Block
	Pos    Pos
}

func (w *WithStmt) nodeTag() {}
func (w *WithStmt) stmtTag() {}

type InlineCStmt struct {
	Code string
}

func (i *InlineCStmt) nodeTag() {}
func (i *InlineCStmt) stmtTag() {}

// --- Expressions ---

type Expr interface {
	Node
	exprTag()
}

type Ident struct {
	Name string
	Pos  Pos
}

func (i *Ident) nodeTag() {}
func (i *Ident) exprTag() {}

type IntLit struct {
	Value string
}

func (i *IntLit) nodeTag() {}
func (i *IntLit) exprTag() {}

type FloatLit struct {
	Value string
}

func (f *FloatLit) nodeTag() {}
func (f *FloatLit) exprTag() {}

type StringLit struct {
	Value string
}

func (s *StringLit) nodeTag() {}
func (s *StringLit) exprTag() {}

type BoolLit struct {
	Value bool
}

func (b *BoolLit) nodeTag() {}
func (b *BoolLit) exprTag() {}

type InterpString struct {
	Parts []Expr // alternating StringLit and expressions
}

func (i *InterpString) nodeTag() {}
func (i *InterpString) exprTag() {}

type RangeExpr struct {
	Start Expr
	End   Expr
}

func (r *RangeExpr) nodeTag() {}
func (r *RangeExpr) exprTag() {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) nodeTag() {}
func (b *BinaryExpr) exprTag() {}

type UnaryExpr struct {
	Op      string
	Operand Expr
}

func (u *UnaryExpr) nodeTag() {}
func (u *UnaryExpr) exprTag() {}

type CallExpr struct {
	Callee   Expr
	TypeArgs []TypeExpr // explicit generic type args: <int, string>
	Args     []Expr
}

func (c *CallExpr) nodeTag() {}
func (c *CallExpr) exprTag() {}

type MemberExpr struct {
	Object   Expr
	Field    string
	Optional bool // ?.
	Pos      Pos
}

func (m *MemberExpr) nodeTag() {}
func (m *MemberExpr) exprTag() {}

type IndexExpr struct {
	Object Expr
	Index  Expr
	Pos    Pos
}

func (i *IndexExpr) nodeTag() {}
func (i *IndexExpr) exprTag() {}

type IsExpr struct {
	Expr     Expr
	TypeName string
}

func (i *IsExpr) nodeTag() {}
func (i *IsExpr) exprTag() {}

type StructLit struct {
	Fields []*StructLitField
}

func (s *StructLit) nodeTag() {}
func (s *StructLit) exprTag() {}

type StructLitField struct {
	Name  string // empty = positional
	Key   Expr   // for dictionary entries: expr: value
	Value Expr
}

type ArrayLit struct {
	Elements []Expr
}

func (a *ArrayLit) nodeTag() {}
func (a *ArrayLit) exprTag() {}

type NullCoalesce struct {
	Left  Expr
	Right Expr
}

func (n *NullCoalesce) nodeTag() {}
func (n *NullCoalesce) exprTag() {}

type LambdaExpr struct {
	Params     []*Param
	ReturnType TypeExpr // nil = void
	Body       *Block
}

func (l *LambdaExpr) nodeTag() {}
func (l *LambdaExpr) exprTag() {}

type SpreadExpr struct{} // ... in call args or method sig

func (s *SpreadExpr) nodeTag() {}
func (s *SpreadExpr) exprTag() {}

type ThisExpr struct{}

func (t *ThisExpr) nodeTag() {}
func (t *ThisExpr) exprTag() {}

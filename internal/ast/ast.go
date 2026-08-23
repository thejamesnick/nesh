// Package ast defines the Nesh abstract syntax tree.
//
// Contract (MODULES.md): dumb data carrying position info. No behavior
// beyond what debugging needs. Imports token only.
package ast

// Pos is a source position (1-based line and column).
type Pos struct {
	Line   int
	Column int
}

// Node is anything in the tree that knows where it came from.
type Node interface {
	Position() Pos
}

// Script is a parsed .nsh file: a sequence of statements.
type Script struct {
	Stmts []Stmt
}

// Stmt is a single statement.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is a typed expression tree.
type Expr interface {
	Node
	exprNode()
}

// LetStmt is `let name = value`.
type LetStmt struct {
	Pos   Pos
	Name  string
	Value Expr
}

// PrintStmt is `print arg1 arg2 ...` (zero or more space-separated args).
type PrintStmt struct {
	Pos  Pos
	Args []Expr
}

// IfStmt is `if cond then ... elif cond then ... else ... end`.
// An `elif` chain nests: Else holds either zero stmts or a single *IfStmt.
type IfStmt struct {
	Pos  Pos
	Cond Expr
	Then []Stmt
	Else []Stmt
}

// FnStmt is `fn name(params) ... end`. Functions are defined globally.
type FnStmt struct {
	Pos    Pos
	Name   string
	Params []string
	Body   []Stmt
}

// ReturnStmt is `return [expr]`; Value may be nil (implicit false).
type ReturnStmt struct {
	Pos   Pos
	Value Expr
}

// ExprStmt is a bare expression used as a statement — only calls today
// (e.g. `deploy("prod")` for its side effect).
type ExprStmt struct {
	Pos  Pos
	Expr Expr
}

// WhileStmt is `while cond ... end`.
type WhileStmt struct {
	Pos  Pos
	Cond Expr
	Body []Stmt
}

// CmdStmt is a bare system command: `git status --short`. Words are literal
// (no variable expansion yet — Phase 3 decision).
type CmdStmt struct {
	Pos  Pos
	Name string
	Args []string
}

// RunExpr is `run <command>` in expression position; it evaluates to the
// command's exit code as an Int.
type RunExpr struct {
	Pos  Pos
	Name string
	Args []string
}

// Ident is a variable reference.
type Ident struct {
	Pos  Pos
	Name string
}

// IntLit is an integer literal.
type IntLit struct {
	Pos   Pos
	Value int64
}

// FloatLit is a floating-point literal.
type FloatLit struct {
	Pos   Pos
	Value float64
}

// StringLit is a string literal (already unescaped by the lexer).
type StringLit struct {
	Pos   Pos
	Value string
}

// BoolLit is `true` or `false`.
type BoolLit struct {
	Pos   Pos
	Value bool
}

// PrefixExpr is a unary operation (currently: negation).
type PrefixExpr struct {
	Pos   Pos
	Op    string // "-"
	Right Expr
}

// InfixExpr is a binary operation: + - * / == != < > <= >= and or
type InfixExpr struct {
	Pos  Pos
	Op   string
	L, R Expr
}

// CallExpr is a function call: name(arg1, arg2, ...).
type CallExpr struct {
	Pos  Pos
	Name string
	Args []Expr
}

func (s *Script) Position() Pos {
	if len(s.Stmts) == 0 {
		return Pos{}
	}
	return s.Stmts[0].Position()
}

func (s *LetStmt) Position() Pos    { return s.Pos }
func (s *PrintStmt) Position() Pos  { return s.Pos }
func (s *IfStmt) Position() Pos     { return s.Pos }
func (s *FnStmt) Position() Pos     { return s.Pos }
func (s *ReturnStmt) Position() Pos { return s.Pos }
func (s *ExprStmt) Position() Pos   { return s.Expr.Position() }
func (s *WhileStmt) Position() Pos  { return s.Pos }
func (s *CmdStmt) Position() Pos    { return s.Pos }
func (e *RunExpr) Position() Pos    { return e.Pos }
func (e *Ident) Position() Pos      { return e.Pos }
func (e *IntLit) Position() Pos     { return e.Pos }
func (e *FloatLit) Position() Pos   { return e.Pos }
func (e *StringLit) Position() Pos  { return e.Pos }
func (e *BoolLit) Position() Pos    { return e.Pos }
func (e *PrefixExpr) Position() Pos { return e.Pos }
func (e *InfixExpr) Position() Pos  { return e.Pos }
func (e *CallExpr) Position() Pos   { return e.Pos }

func (*LetStmt) stmtNode()    {}
func (*PrintStmt) stmtNode()  {}
func (*IfStmt) stmtNode()     {}
func (*FnStmt) stmtNode()     {}
func (*ReturnStmt) stmtNode() {}
func (*ExprStmt) stmtNode()   {}
func (*WhileStmt) stmtNode()  {}
func (*CmdStmt) stmtNode()    {}

func (*Ident) exprNode()      {}
func (*IntLit) exprNode()     {}
func (*FloatLit) exprNode()   {}
func (*StringLit) exprNode()  {}
func (*BoolLit) exprNode()    {}
func (*PrefixExpr) exprNode() {}
func (*InfixExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*RunExpr) exprNode()    {}

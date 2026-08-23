// Package ast defines the Nesh abstract syntax tree.
//
// Contract (MODULES.md): dumb data carrying position info. No behavior
// beyond what debugging needs. Imports token only.
// JSON encoding for the agent API lives in json.go.
package ast

// Pos is a source position (1-based line and column).
type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Node is anything in the tree that knows where it came from.
type Node interface {
	Position() Pos
}

// Script is a parsed .nsh file: a sequence of statements.
type Script struct {
	Stmts []Stmt `json:"stmts"`
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
	Pos   Pos    `json:"pos"`
	Name  string `json:"name"`
	Value Expr   `json:"value"`
}

// PrintStmt is `print arg1 arg2 ...` (zero or more space-separated args).
type PrintStmt struct {
	Pos  Pos    `json:"pos"`
	Args []Expr `json:"args"`
}

// IfStmt is `if cond then ... elif cond then ... else ... end`.
// An `elif` chain nests: Else holds either zero stmts or a single *IfStmt.
type IfStmt struct {
	Pos  Pos    `json:"pos"`
	Cond Expr   `json:"cond"`
	Then []Stmt `json:"then"`
	Else []Stmt `json:"else"`
}

// FnStmt is `fn name(params) ... end`. Functions are defined globally.
type FnStmt struct {
	Pos    Pos      `json:"pos"`
	Name   string   `json:"name"`
	Params []string `json:"params"`
	Body   []Stmt   `json:"body"`
}

// ReturnStmt is `return [expr]`; Value may be nil (implicit false).
type ReturnStmt struct {
	Pos   Pos  `json:"pos"`
	Value Expr `json:"value,omitempty"`
}

// ExprStmt is a bare expression used as a statement — only calls today
// (e.g. `deploy("prod")` for its side effect).
type ExprStmt struct {
	Pos  Pos  `json:"pos"`
	Expr Expr `json:"expr"`
}

// WhileStmt is `while cond ... end`.
type WhileStmt struct {
	Pos  Pos    `json:"pos"`
	Cond Expr   `json:"cond"`
	Body []Stmt `json:"body"`
}

// CmdStmt is a bare system command: `git status --short`. Words are literal
// (no variable expansion yet — Phase 3 decision).
type CmdStmt struct {
	Pos  Pos      `json:"pos"`
	Name string   `json:"name"`
	Args []string `json:"args,omitempty"`
}

// RunExpr is `run <command>` in expression position; it evaluates to the
// command's exit code as an Int.
type RunExpr struct {
	Pos  Pos      `json:"pos"`
	Name string   `json:"name"`
	Args []string `json:"args,omitempty"`
}

// Ident is a variable reference.
type Ident struct {
	Pos  Pos    `json:"pos"`
	Name string `json:"name"`
}

// IntLit is an integer literal.
type IntLit struct {
	Pos   Pos   `json:"pos"`
	Value int64 `json:"value"`
}

// FloatLit is a floating-point literal.
type FloatLit struct {
	Pos   Pos     `json:"pos"`
	Value float64 `json:"value"`
}

// StringLit is a string literal (already unescaped by the lexer).
type StringLit struct {
	Pos   Pos    `json:"pos"`
	Value string `json:"value"`
}

// BoolLit is `true` or `false`.
type BoolLit struct {
	Pos   Pos  `json:"pos"`
	Value bool `json:"value"`
}

// PrefixExpr is a unary operation (currently: negation, not).
type PrefixExpr struct {
	Pos   Pos    `json:"pos"`
	Op    string `json:"op"`
	Right Expr   `json:"right"`
}

// InfixExpr is a binary operation: + - * / == != < > <= >= and or
type InfixExpr struct {
	Pos Pos    `json:"pos"`
	Op  string `json:"op"`
	L   Expr   `json:"left"`
	R   Expr   `json:"right"`
}

// CallExpr is a function call: name(arg1, arg2, ...).
type CallExpr struct {
	Pos  Pos    `json:"pos"`
	Name string `json:"name"`
	Args []Expr `json:"args,omitempty"`
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

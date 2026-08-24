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

// ForStmt is `for x in <list-expr> ... end`.
type ForStmt struct {
	Pos  Pos    `json:"pos"`
	Var  string `json:"var"`
	Iter Expr   `json:"iter"`
	Body []Stmt `json:"body"`
}

// BreakStmt exits the innermost enclosing while/for loop.
type BreakStmt struct {
	Pos Pos `json:"pos"`
}

// ContinueStmt skips to the next iteration of the innermost enclosing loop.
type ContinueStmt struct {
	Pos Pos `json:"pos"`
}

// TryStmt is `try ... [on failure ...] end`. Runtime errors and `fail`
// statements inside the try block jump to the on-failure block, where the
// `failure` variable holds the message. Without a handler, failures are
// swallowed and execution continues after `end`.
type TryStmt struct {
	Pos      Pos    `json:"pos"`
	Try      []Stmt `json:"try"`
	On       []Stmt `json:"on,omitempty"` // present when HasOn
	HasOn    bool   `json:"has_on"`
}

// FailStmt is `fail ["message"]` — raises a catchable failure. Without a
// message the failure is "failed".
type FailStmt struct {
	Pos Pos  `json:"pos"`
	Msg Expr `json:"msg,omitempty"`
}

// ExitStmt is `exit [code]` — terminates the script with an exit code
// (default 0). Not catchable by try.
type ExitStmt struct {
	Pos  Pos  `json:"pos"`
	Code Expr `json:"code,omitempty"`
}

// ImportStmt is `import "path.nsh" [as alias]`.
// Without an alias, the module's definitions merge into globals; with one,
// they are reachable via dotted access (alias.name).
type ImportStmt struct {
	Pos   Pos    `json:"pos"`
	Path  string `json:"path"`
	Alias string `json:"alias,omitempty"`
}

// Redirect is a stream redirection attached to a command: Op is ">",
// ">>" (append), or "<" (stdin from file); Path is the target file.
type Redirect struct {
	Op   string `json:"op"`
	Path string `json:"path"`
}

// CmdStage is one command segment: a bare command, or one stage of a
// pipeline (`cat log | grep error | wc -l`).
type CmdStage struct {
	Name      string     `json:"name"`
	Args      []string   `json:"args,omitempty"`
	Redirects []Redirect `json:"redirects,omitempty"`
}

// CmdStmt is a bare system command: `git status --short`. Words are literal
// (no variable expansion yet — Phase 3 decision).
type CmdStmt struct {
	Pos       Pos        `json:"pos"`
	Name      string     `json:"name"`
	Args      []string   `json:"args,omitempty"`
	Redirects []Redirect `json:"redirects,omitempty"`
}

// PipelineStmt is a chain of commands: each stage's stdout feeds the next
// stage's stdin; the statement's exit status is the last stage's.
type PipelineStmt struct {
	Pos    Pos        `json:"pos"`
	Stages []CmdStage `json:"stages"`
}

// RunExpr is `run <command>` in expression position; it evaluates to the
// command's exit code as an Int. With `Pipe` stages (`run a | b | c`) it
// is a pipeline expression and still evaluates to the LAST stage's code.
type RunExpr struct {
	Pos       Pos        `json:"pos"`
	Name      string     `json:"name"`
	Args      []string   `json:"args,omitempty"`
	Redirects []Redirect `json:"redirects,omitempty"`
	Pipe      []CmdStage `json:"pipe,omitempty"`
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

func (s *ImportStmt) Position() Pos { return s.Pos }
func (s *LetStmt) Position() Pos    { return s.Pos }
func (s *PrintStmt) Position() Pos  { return s.Pos }
func (s *IfStmt) Position() Pos     { return s.Pos }
func (s *FnStmt) Position() Pos     { return s.Pos }
func (s *ReturnStmt) Position() Pos { return s.Pos }
func (s *ExprStmt) Position() Pos   { return s.Expr.Position() }
func (s *WhileStmt) Position() Pos  { return s.Pos }
func (s *ForStmt) Position() Pos    { return s.Pos }
func (s *BreakStmt) Position() Pos  { return s.Pos }
func (s *ContinueStmt) Position() Pos { return s.Pos }
func (s *TryStmt) Position() Pos    { return s.Pos }
func (s *FailStmt) Position() Pos   { return s.Pos }
func (s *ExitStmt) Position() Pos   { return s.Pos }
func (s *CmdStmt) Position() Pos    { return s.Pos }
func (s *PipelineStmt) Position() Pos { return s.Pos }
func (e *RunExpr) Position() Pos    { return e.Pos }
func (e *Ident) Position() Pos      { return e.Pos }
func (e *IntLit) Position() Pos     { return e.Pos }
func (e *FloatLit) Position() Pos   { return e.Pos }
func (e *StringLit) Position() Pos  { return e.Pos }
func (e *BoolLit) Position() Pos    { return e.Pos }
func (e *PrefixExpr) Position() Pos { return e.Pos }
func (e *InfixExpr) Position() Pos  { return e.Pos }
func (e *CallExpr) Position() Pos   { return e.Pos }

func (*ImportStmt) stmtNode() {}
func (*LetStmt) stmtNode()    {}
func (*PrintStmt) stmtNode()  {}
func (*IfStmt) stmtNode()     {}
func (*FnStmt) stmtNode()     {}
func (*ReturnStmt) stmtNode() {}
func (*ExprStmt) stmtNode()   {}
func (*WhileStmt) stmtNode()  {}
func (*ForStmt) stmtNode()    {}
func (*BreakStmt) stmtNode()  {}
func (*ContinueStmt) stmtNode() {}
func (*TryStmt) stmtNode()    {}
func (*FailStmt) stmtNode()   {}
func (*ExitStmt) stmtNode()   {}
func (*CmdStmt) stmtNode()    {}
func (*PipelineStmt) stmtNode() {}

func (*Ident) exprNode()      {}
func (*IntLit) exprNode()     {}
func (*FloatLit) exprNode()   {}
func (*StringLit) exprNode()  {}
func (*BoolLit) exprNode()    {}
func (*PrefixExpr) exprNode() {}
func (*InfixExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*RunExpr) exprNode()    {}

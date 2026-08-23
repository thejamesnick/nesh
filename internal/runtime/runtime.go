// Package runtime evaluates a Nesh AST.
//
// Contract (MODULES.md): pure Go. No os/exec, no filesystem, no env, no time.
// Output arrives via the Output interface so the whole language is
// unit-testable without touching a syscall.
package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"nesh/internal/ast"
)

// Output is where print writes. cmd/nesh wires it to stdout.
type Output interface {
	WriteString(s string) (int, error)
}

// Error is a runtime failure with the position of the offending node.
type Error struct {
	Line, Column int
	Msg          string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Msg)
}

// Value is a Nesh runtime value.
type Value interface {
	String() string
}

// Int, Float, Str are the Phase 1 value types.
type Int int64
type Float float64
type Str string

func (v Int) String() string   { return strconv.FormatInt(int64(v), 10) }
func (v Float) String() string { return strconv.FormatFloat(float64(v), 'g', -1, 64) }
func (v Str) String() string   { return string(v) }

// Runtime executes scripts against a persistent scope (globals survive
// across Run calls, which is what the REPL needs).
type Runtime struct {
	out     Output
	globals map[string]Value
}

// New builds a Runtime writing to out.
func New(out Output) *Runtime {
	return &Runtime{out: out, globals: make(map[string]Value)}
}

// Run executes every statement in script.
func (r *Runtime) Run(script *ast.Script) *Error {
	for _, stmt := range script.Stmts {
		if err := r.execStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// Global returns a global variable (REPL/tests); ok=false if undefined.
func (r *Runtime) Global(name string) (Value, bool) {
	v, ok := r.globals[name]
	return v, ok
}

func (r *Runtime) execStmt(stmt ast.Stmt) *Error {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		v, err := r.eval(s.Value)
		if err != nil {
			return err
		}
		r.globals[s.Name] = v
		return nil
	case *ast.PrintStmt:
		parts := make([]string, len(s.Args))
		for i, arg := range s.Args {
			v, err := r.eval(arg)
			if err != nil {
				return err
			}
			parts[i] = v.String()
		}
		r.out.WriteString(strings.Join(parts, " ") + "\n")
		return nil
	default:
		return errAt(stmt.Position(), "cannot execute %T", stmt)
	}
}

func (r *Runtime) eval(e ast.Expr) (Value, *Error) {
	switch n := e.(type) {
	case *ast.IntLit:
		return Int(n.Value), nil
	case *ast.FloatLit:
		return Float(n.Value), nil
	case *ast.StringLit:
		return Str(n.Value), nil
	case *ast.Ident:
		if v, ok := r.globals[n.Name]; ok {
			return v, nil
		}
		return nil, errAt(n.Pos, "undefined variable: %s", n.Name)
	case *ast.PrefixExpr:
		v, err := r.eval(n.Right)
		if err != nil {
			return nil, err
		}
		switch x := v.(type) {
		case Int:
			return -x, nil
		case Float:
			return -x, nil
		default:
			return nil, errAt(n.Pos, "cannot negate %s", v)
		}
	case *ast.InfixExpr:
		return r.evalInfix(n)
	default:
		return nil, errAt(e.Position(), "cannot evaluate %T", e)
	}
}

func (r *Runtime) evalInfix(n *ast.InfixExpr) (Value, *Error) {
	l, err := r.eval(n.L)
	if err != nil {
		return nil, err
	}
	rr, err := r.eval(n.R)
	if err != nil {
		return nil, err
	}

	// Strings: concatenation only, with strings — no silent coercion.
	if ls, ok := l.(Str); ok {
		if rs, ok := rr.(Str); ok {
			return ls + rs, nil
		}
		return nil, errAt(n.Pos, "cannot join string and %s with %q", rr, n.Op)
	}

	switch x := l.(type) {
	case Int:
		switch y := rr.(type) {
		case Int:
			return intInfix(n, x, y)
		case Float:
			return floatInfix(n, Float(x), y)
		}
	case Float:
		switch y := rr.(type) {
		case Int:
			return floatInfix(n, x, Float(y))
		case Float:
			return floatInfix(n, x, y)
		}
	}
	return nil, errAt(n.Pos, "cannot use %q on %s and %s", n.Op, l, rr)
}

func intInfix(n *ast.InfixExpr, x, y Int) (Value, *Error) {
	switch n.Op {
	case "+":
		return x + y, nil
	case "-":
		return x - y, nil
	case "*":
		return x * y, nil
	case "/":
		if y == 0 {
			return nil, errAt(n.Pos, "division by zero")
		}
		return x / y, nil // integer division, truncating (documented)
	}
	return nil, errAt(n.Pos, "unknown operator %q", n.Op)
}

func floatInfix(n *ast.InfixExpr, x, y Float) (Value, *Error) {
	switch n.Op {
	case "+":
		return x + y, nil
	case "-":
		return x - y, nil
	case "*":
		return x * y, nil
	case "/":
		if y == 0 {
			return nil, errAt(n.Pos, "division by zero")
		}
		return x / y, nil
	}
	return nil, errAt(n.Pos, "unknown operator %q", n.Op)
}

func errAt(p ast.Pos, format string, args ...any) *Error {
	return &Error{Line: p.Line, Column: p.Column, Msg: fmt.Sprintf(format, args...)}
}

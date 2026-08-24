// Package runtime evaluates a Nesh AST.
//
// Contract (MODULES.md): pure Go. No os/exec, no filesystem, no env, no time.
// Output arrives via the Output interface so the whole language is
// unit-testable without touching a syscall.
package runtime

import (
	"cmp"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"nesh/internal/ast"
	"nesh/internal/parser"
	"nesh/internal/shell"
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

// Int, Float, Str, Bool are the Phase 1–2 value types.
type Int int64
type Float float64
type Str string
type Bool bool

// List is an ordered collection of values, produced by builtins like split.
type List []Value

func (l List) String() string {
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = v.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Builtin is a native function registered from Go. Fn receives already
// evaluated arguments and may return a plain error; the runtime wraps it
// with the call site's position.
type Builtin struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (b *Builtin) String() string { return "<builtin " + b.Name + ">" }

// Module is an imported file's exported definitions.
type Module struct {
	Name    string
	Exports map[string]Value
}

func (m *Module) String() string { return "<module " + m.Name + ">" }

// Func is a user-defined function. It captures its defining scope chain
// (lexical scoping), so module functions see module state, not caller state.
type Func struct {
	Name   string
	Params []string
	Body   []ast.Stmt
	env    []map[string]Value // defining environment (globals + enclosing scopes)
}

func (v Int) String() string   { return strconv.FormatInt(int64(v), 10) }
func (v Float) String() string { return strconv.FormatFloat(float64(v), 'g', -1, 64) }
func (v Str) String() string   { return string(v) }
func (v Bool) String() string {
	if v {
		return "true"
	}
	return "false"
}
func (f *Func) String() string { return "<fn " + f.Name + ">" }

// Truthy reports the Nesh truthiness of a value:
// false for Bool(false), Int(0), Float(0.0), Str(""); true otherwise.
// This is the documented rule (TECHNICAL_SPEC.md — Truthiness).
func Truthy(v Value) bool {
	switch x := v.(type) {
	case Bool:
		return bool(x)
	case Int:
		return x != 0
	case Float:
		return x != 0
	case Str:
		return x != ""
	default:
		return true
	}
}

// Runtime executes scripts against a scope stack: scopes[0] is globals,
// each function call pushes one local scope (which is what recursion and
// the REPL need). System commands run through the injected CommandRunner;
// without one, command syntax is a runtime error (keeps tests OS-free).
type Runtime struct {
	out      Output
	stdin    io.Reader // source for child-process stdin (nil = none)
	scopes   []map[string]Value
	runner   shell.CommandRunner
	events   func(Event)
	fs       shell.FileSystem
	baseDir  string
	factory  func(*Runtime)     // prepares child runtimes (builtins etc.)
	modCache map[string]*Module // resolved absolute path → module
	loading  map[string]bool    // cycle detection
}

// New builds a Runtime writing to out.
func New(out Output) *Runtime {
	return &Runtime{out: out, scopes: []map[string]Value{{}}, loading: map[string]bool{}, modCache: map[string]*Module{}}
}

// SetRunner enables system commands (`git status`, `run ...`).
func (r *Runtime) SetRunner(runner shell.CommandRunner) { r.runner = runner }

// SetStdin sets what child processes read when no `<` redirect applies
// (wire os.Stdin for interactive/script use).
func (r *Runtime) SetStdin(stdin io.Reader) { r.stdin = stdin }

// Define installs a builtin function in the global scope.
func (r *Runtime) Define(name string, fn func(args []Value) (Value, error)) {
	r.scopes[0][name] = &Builtin{Name: name, Fn: fn}
}

// SetFileSystem enables file builtins and imports.
func (r *Runtime) SetFileSystem(fs shell.FileSystem) { r.fs = fs }

// SetBaseDir anchors relative import paths (set to the script's directory).
func (r *Runtime) SetBaseDir(dir string) { r.baseDir = dir }

// SetRuntimeFactory registers a callback preparing child runtimes created
// for module loading (e.g. re-registering builtins).
func (r *Runtime) SetRuntimeFactory(fn func(*Runtime)) { r.factory = fn }

// Event is one structured execution step for the agent API (nesh --json).
type Event struct {
	Type   string   `json:"type"` // "print" | "let" | "command" | "error"
	Line   int      `json:"line"`
	Column int      `json:"column"`
	Text   string   `json:"text,omitempty"`  // print output / error message
	Name   string   `json:"name,omitempty"`  // let target / command name
	Args   []string `json:"args,omitempty"`  // command args
	Code   int      `json:"code,omitempty"`  // command exit code
	Value  string   `json:"value,omitempty"` // let bound value
}

// SetEventSink registers a callback receiving each execution step.
// Nil (the default) disables collection entirely — zero overhead.
func (r *Runtime) SetEventSink(fn func(Event)) { r.events = fn }

func (r *Runtime) emit(e Event) {
	if r.events != nil {
		r.events(e)
	}
}

// Run executes every statement in script. A top-level `return`, `break`,
// or `continue` is an error (the parser rejects break/continue outside
// loops; this is defense in depth).
func (r *Runtime) Run(script *ast.Script) *Error {
	f, err := r.execBlock(script.Stmts)
	if err != nil {
		r.emit(Event{Type: "error", Line: err.Line, Column: err.Column, Text: err.Msg})
		return err
	}
	if f != nil {
		switch f.kind {
		case flowBreak:
			return errAt(f.pos, "break outside loop")
		case flowContinue:
			return errAt(f.pos, "continue outside loop")
		default:
			return errAt(f.pos, "return outside function")
		}
	}
	return nil
}

// Global returns a global variable (REPL/tests); ok=false if undefined.
func (r *Runtime) Global(name string) (Value, bool) {
	return r.lookup(name)
}

func (r *Runtime) lookup(name string) (Value, bool) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if v, ok := r.scopes[i][name]; ok {
			return v, true
		}
	}
	return nil, false
}

// flow signals unwinding through blocks: return to the nearest call,
// break/continue to the nearest enclosing loop.
type flowKind int

const (
	flowNone      flowKind = iota
	flowReturn
	flowBreak
	flowContinue
)

type flow struct {
	kind flowKind
	v    Value   // return payload (flowReturn only)
	pos  ast.Pos // source position of the triggering statement
}

func (r *Runtime) execBlock(stmts []ast.Stmt) (*flow, *Error) {
	for _, stmt := range stmts {
		f, err := r.execStmt(stmt)
		if err != nil || (f != nil && f.kind != flowNone) {
			return f, err
		}
	}
	return nil, nil
}

func (r *Runtime) execStmt(stmt ast.Stmt) (*flow, *Error) {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		v, err := r.eval(s.Value)
		if err != nil {
			return nil, err
		}
		r.scopes[len(r.scopes)-1][s.Name] = v // let always binds in the current scope
		p := s.Position()
		r.emit(Event{Type: "let", Line: p.Line, Column: p.Column, Name: s.Name, Value: v.String()})
		return nil, nil
	case *ast.PrintStmt:
		parts := make([]string, len(s.Args))
		for i, arg := range s.Args {
			v, err := r.eval(arg)
			if err != nil {
				return nil, err
			}
			parts[i] = v.String()
		}
		text := strings.Join(parts, " ")
		r.out.WriteString(text + "\n")
		p := s.Position()
		r.emit(Event{Type: "print", Line: p.Line, Column: p.Column, Text: text})
		return nil, nil
	case *ast.IfStmt:
		cond, err := r.eval(s.Cond)
		if err != nil {
			return nil, err
		}
		if Truthy(cond) {
			return r.execBlock(s.Then)
		}
		return r.execBlock(s.Else)
	case *ast.WhileStmt:
		for {
			cond, err := r.eval(s.Cond)
			if err != nil {
				return nil, err
			}
			if !Truthy(cond) {
				return nil, nil
			}
			f, err := r.execBlock(s.Body)
			if err != nil {
				return nil, err
			}
			switch {
			case f == nil:
			case f.kind == flowBreak:
				return nil, nil
			case f.kind == flowContinue:
				continue
			default:
				return f, nil // return unwinds past the loop
			}
		}
	case *ast.ForStmt:
		it, err := r.eval(s.Iter)
		if err != nil {
			return nil, err
		}
		list, ok := it.(List)
		if !ok {
			return nil, errAt(s.Pos, "cannot iterate %s — for-in needs a list", it)
		}
		// The loop variable binds in the CURRENT scope (same rule as
		// `let`), so accumulators work exactly like in `while`.
		cur := len(r.scopes) - 1
		r.scopes[cur][s.Var] = Bool(false)
		for _, item := range list {
			r.scopes[cur][s.Var] = item
			f, err := r.execBlock(s.Body)
			if err != nil {
				return nil, err
			}
			switch {
			case f == nil:
			case f.kind == flowBreak:
				return nil, nil
			case f.kind == flowContinue:
				continue
			default:
				return f, nil
			}
		}
		return nil, nil
	case *ast.BreakStmt:
		return &flow{kind: flowBreak, pos: s.Pos}, nil
	case *ast.ContinueStmt:
		return &flow{kind: flowContinue, pos: s.Pos}, nil
	case *ast.CmdStmt:
		_, err := r.runCommand(s.Pos, s.Name, s.Args, s.Redirects)
		return nil, err
	case *ast.ImportStmt:
		return nil, r.doImport(s)
	case *ast.FnStmt:
		r.scopes[0][s.Name] = &Func{Name: s.Name, Params: s.Params, Body: s.Body, env: r.scopes}
		return nil, nil
	case *ast.ExprStmt:
		_, err := r.eval(s.Expr) // value discarded; called for side effects
		return nil, err
	case *ast.ReturnStmt:
		if s.Value == nil {
			return &flow{kind: flowReturn, v: Bool(false), pos: s.Pos}, nil
		}
		v, err := r.eval(s.Value)
		if err != nil {
			return nil, err
		}
		return &flow{kind: flowReturn, v: v, pos: s.Pos}, nil
	default:
		return nil, errAt(stmt.Position(), "cannot execute %T", stmt)
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
	case *ast.BoolLit:
		return Bool(n.Value), nil
	case *ast.Ident:
		if strings.Contains(n.Name, ".") {
			return r.resolveDotted(n)
		}
		if v, ok := r.lookup(n.Name); ok {
			return v, nil
		}
		return nil, errAt(n.Pos, "undefined variable: %s", n.Name)
	case *ast.CallExpr:
		return r.call(n)
	case *ast.RunExpr:
		code, err := r.runCommand(n.Pos, n.Name, n.Args, n.Redirects)
		if err != nil {
			return nil, err
		}
		return Int(code), nil
	case *ast.PrefixExpr:
		v, err := r.eval(n.Right)
		if err != nil {
			return nil, err
		}
		switch n.Op {
		case "-":
			switch x := v.(type) {
			case Int:
				return -x, nil
			case Float:
				return -x, nil
			default:
				return nil, errAt(n.Pos, "cannot negate %s", v)
			}
		case "not":
			return Bool(!Truthy(v)), nil
		default:
			return nil, errAt(n.Pos, "unknown unary operator %q", n.Op)
		}
	case *ast.InfixExpr:
		return r.evalInfix(n)
	default:
		return nil, errAt(e.Position(), "cannot evaluate %T", e)
	}
}

// doImport loads a module: resolve path, cache, detect cycles, execute in
// an isolated child runtime, then bind (alias) or merge (no alias).
func (r *Runtime) doImport(s *ast.ImportStmt) *Error {
	if r.fs == nil {
		return errAt(s.Pos, "imports need a filesystem — not available here")
	}
	resolved := s.Path
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(r.baseDir, resolved)
	}

	if mod, cached := r.modCache[resolved]; cached {
		return r.bindModule(s, mod)
	}
	if r.loading[resolved] {
		return errAt(s.Pos, "circular import of %s", filepath.Base(resolved))
	}

	data, ferr := r.fs.ReadFile(resolved)
	if ferr != nil {
		return errAt(s.Pos, "cannot read module %s: %v", s.Path, ferr)
	}
	script, perr := parser.Parse(string(data))
	if perr != nil {
		return errAt(s.Pos, "in %s: %v", filepath.Base(resolved), perr)
	}

	child := &Runtime{
		out:      r.out,
		scopes:   []map[string]Value{{}},
		runner:   r.runner,
		events:   r.events,
		fs:       r.fs,
		baseDir:  filepath.Dir(resolved),
		factory:  r.factory,
		modCache: r.modCache, // shared so cross-module imports deduplicate
		loading:  r.loading,
	}
	if r.factory != nil {
		r.factory(child)
	}

	r.loading[resolved] = true
	f, rerr := child.execBlock(script.Stmts)
	delete(r.loading, resolved)
	if rerr != nil {
		return rerr
	}
	if f != nil {
		switch f.kind {
		case flowBreak:
			return errAt(f.pos, "break outside loop")
		case flowContinue:
			return errAt(f.pos, "continue outside loop")
		default:
			return errAt(f.pos, "return outside function")
		}
	}

	mod := &Module{Name: moduleName(resolved), Exports: child.scopes[0]}
	r.modCache[resolved] = mod
	return r.bindModule(s, mod)
}

func (r *Runtime) bindModule(s *ast.ImportStmt, mod *Module) *Error {
	if s.Alias != "" {
		r.scopes[len(r.scopes)-1][s.Alias] = mod
		return nil
	}
	for k, v := range mod.Exports { // source-style merge; last write wins
		r.scopes[0][k] = v
	}
	return nil
}

func moduleName(path string) string {
	base := filepath.Base(path)
	if ext := filepath.Ext(base); ext != "" {
		return strings.TrimSuffix(base, ext)
	}
	return base
}

// outputWriter adapts the Output sink to io.Writer for command streams.
type outputWriter struct{ o Output }

func (w outputWriter) Write(p []byte) (int, error) { return w.o.WriteString(string(p)) }

// runCommand executes a system command through the injected runner,
// applying redirections: `>` write, `>>` append, `<` read stdin.
func (r *Runtime) runCommand(pos ast.Pos, name string, args []string, redirects []ast.Redirect) (int, *Error) {
	if r.runner == nil {
		return 0, errAt(pos, "system commands are not available here")
	}
	if v, ok := r.lookup(name); ok {
		if _, isFn := v.(*Func); !isFn {
			return 0, errAt(pos, "%s is a variable, not a command — did you mean print %s?", name, name)
		}
	}

	var stdin io.Reader = r.stdin
	var stdout io.Writer = outputWriter{r.out}
	var closers []io.Closer
	defer func() {
		for _, c := range closers {
			c.Close()
		}
	}()

	if len(redirects) > 0 {
		opener, hasOpener := r.fs.(shell.Opener)
		if !hasOpener {
			return 0, errAt(pos, "redirection needs filesystem access — not available here")
		}
		for _, rd := range redirects {
			switch rd.Op {
			case "<":
				f, err := opener.OpenRead(rd.Path)
				if err != nil {
					return 0, errAt(pos, "cannot read %s: %v", rd.Path, err)
				}
				stdin = f
				closers = append(closers, f)
			case ">":
				f, err := opener.OpenWrite(rd.Path, false)
				if err != nil {
					return 0, errAt(pos, "cannot write %s: %v", rd.Path, err)
				}
				stdout = f
				closers = append(closers, f)
			case ">>":
				f, err := opener.OpenWrite(rd.Path, true)
				if err != nil {
					return 0, errAt(pos, "cannot append to %s: %v", rd.Path, err)
				}
				stdout = f
				closers = append(closers, f)
			}
		}
	}

	code := r.runner.Run(name, args, stdin, stdout, stdout)
	r.emit(Event{Type: "command", Line: pos.Line, Column: pos.Column, Name: name, Args: args, Code: code})
	return code, nil
}

// resolveDotted evaluates `alias.name` against an imported module.
func (r *Runtime) resolveDotted(n *ast.Ident) (Value, *Error) {
	parts := strings.SplitN(n.Name, ".", 2)
	base, ok := r.lookup(parts[0])
	if !ok {
		return nil, errAt(n.Pos, "undefined variable: %s", parts[0])
	}
	mod, isMod := base.(*Module)
	if !isMod {
		return nil, errAt(n.Pos, "%s is not a module — dotted access needs an import alias", parts[0])
	}
	v, exists := mod.Exports[parts[1]]
	if !exists {
		return nil, errAt(n.Pos, "module %s has no %s", mod.Name, parts[1])
	}
	return v, nil
}

// call evaluates a function or builtin invocation.
func (r *Runtime) call(n *ast.CallExpr) (Value, *Error) {
	var fnVal Value
	var ok bool
	if strings.Contains(n.Name, ".") {
		v, err := r.resolveDotted(&ast.Ident{Pos: n.Pos, Name: n.Name})
		if err != nil {
			return nil, err
		}
		fnVal, ok = v, true
	} else {
		fnVal, ok = r.lookup(n.Name)
	}
	if !ok {
		return nil, errAt(n.Pos, "undefined function: %s", n.Name)
	}

	args := make([]Value, len(n.Args)) // args evaluated in the caller's scope
	for i, argExpr := range n.Args {
		v, err := r.eval(argExpr)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}

	switch f := fnVal.(type) {
	case *Builtin:
		res, ferr := f.Fn(args)
		if ferr != nil {
			return nil, errAt(n.Pos, "%s: %v", f.Name, ferr)
		}
		return res, nil
	case *Func:
		return r.callFunc(n, f, args)
	default:
		return nil, errAt(n.Pos, "%s is not a function (it is %s)", n.Name, fnVal)
	}
}

// callFunc runs a user-defined function: args bind as locals in a fresh scope.
func (r *Runtime) callFunc(n *ast.CallExpr, f *Func, args []Value) (Value, *Error) {
	if len(args) != len(f.Params) {
		return nil, errAt(n.Pos, "%s expects %d argument(s), got %d", f.Name, len(f.Params), len(args))
	}
	scope := make(map[string]Value, len(f.Params))
	for i, v := range args {
		scope[f.Params[i]] = v
	}
	// Execute in the function's defining environment + one call scope.
	saved := r.scopes
	r.scopes = append(append([]map[string]Value{}, f.env...), scope)
	ret, err := r.execBlock(f.Body)
	r.scopes = saved
	if err != nil {
		return nil, err
	}
	switch {
	case ret == nil:
		return Bool(false), nil // no explicit return → false (documented)
	case ret.kind == flowReturn:
		return ret.v, nil
	default: // break/continue never cross a fn boundary — loops are reset there
		return nil, errAt(ret.pos, "break/continue outside loop")
	}
}

func (r *Runtime) evalInfix(n *ast.InfixExpr) (Value, *Error) {
	// and/or short-circuit: the right side may have side effects later
	// (function calls), so it must not run when the left decides.
	if n.Op == "and" || n.Op == "or" {
		l, err := r.eval(n.L)
		if err != nil {
			return nil, err
		}
		if n.Op == "and" && !Truthy(l) {
			return Bool(false), nil
		}
		if n.Op == "or" && Truthy(l) {
			return Bool(true), nil
		}
		rr, err := r.eval(n.R)
		if err != nil {
			return nil, err
		}
		return Bool(Truthy(rr)), nil
	}

	l, err := r.eval(n.L)
	if err != nil {
		return nil, err
	}
	rr, err := r.eval(n.R)
	if err != nil {
		return nil, err
	}

	// Comparisons: == and != accept any two values of the same type;
	// < > <= >= accept two numbers or two strings.
	switch n.Op {
	case "==", "!=", "<", ">", "<=", ">=":
		return compare(n, l, rr)
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

// compare applies a comparison operator, promoting Int to Float when mixed.
func compare(n *ast.InfixExpr, l, rr Value) (Value, *Error) {
	switch x := l.(type) {
	case Int:
		switch y := rr.(type) {
		case Int:
			return boolResult(n.Op, int64(x), int64(y)), nil
		case Float:
			return boolResult(n.Op, float64(x), float64(y)), nil
		}
	case Float:
		switch y := rr.(type) {
		case Int:
			return boolResult(n.Op, float64(x), float64(y)), nil
		case Float:
			return boolResult(n.Op, float64(x), float64(y)), nil
		}
	case Str:
		if y, ok := rr.(Str); ok {
			return boolResult(n.Op, string(x), string(y)), nil
		}
	case Bool:
		if y, ok := rr.(Bool); ok && (n.Op == "==" || n.Op == "!=") {
			return Bool(bool(n.Op == "==") == (x == y)), nil
		}
	}
	return nil, errAt(n.Pos, "cannot compare %s and %s with %q", l, rr, n.Op)
}

func boolResult[T cmp.Ordered](op string, a, b T) Value {
	var res bool
	switch op {
	case "==":
		res = a == b
	case "!=":
		res = a != b
	case "<":
		res = a < b
	case ">":
		res = a > b
	case "<=":
		res = a <= b
	case ">=":
		res = a >= b
	}
	return Bool(res)
}

func errAt(p ast.Pos, format string, args ...any) *Error {
	return &Error{Line: p.Line, Column: p.Column, Msg: fmt.Sprintf(format, args...)}
}

package parser

import (
	"testing"

	"nesh/internal/ast"
)

func TestIfStatements(t *testing.T) {
	s, err := Parse("if x > 1 then\nprint \"a\"\nelif x > 0 then\nprint \"b\"\nelse\nprint \"c\"\nend\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifs, ok := s.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("got %#v, want IfStmt", s.Stmts[0])
	}
	if len(ifs.Then) != 1 || ifs.Else == nil {
		t.Fatalf("branch shape wrong: then=%d else=%v", len(ifs.Then), ifs.Else)
	}
	// elif nests into Else as a single IfStmt
	elif, ok := ifs.Else[0].(*ast.IfStmt)
	if !ok || len(elif.Else) != 1 {
		t.Fatalf("elif nesting wrong: %#v", elif)
	}
}

func TestFnAndCalls(t *testing.T) {
	s, err := Parse("fn add(a, b)\nreturn a + b\nend\nlet x = add(2, 3)\nadd(1, 1)\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fn, ok := s.Stmts[0].(*ast.FnStmt)
	if !ok || fn.Name != "add" || len(fn.Params) != 2 || len(fn.Body) != 1 {
		t.Fatalf("fn shape wrong: %#v", fn)
	}
	ret, ok := fn.Body[0].(*ast.ReturnStmt)
	if !ok || ret.Value == nil {
		t.Fatalf("return shape wrong: %#v", fn.Body[0])
	}
	let := s.Stmts[1].(*ast.LetStmt)
	call, ok := let.Value.(*ast.CallExpr)
	if !ok || call.Name != "add" || len(call.Args) != 2 {
		t.Fatalf("call shape wrong: %#v", let.Value)
	}
	expr, ok := s.Stmts[2].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("bare call should be ExprStmt, got %#v", s.Stmts[2])
	}
	if _, ok := expr.Expr.(*ast.CallExpr); !ok {
		t.Fatalf("ExprStmt should hold CallExpr, got %#v", expr.Expr)
	}
}

func TestWhileLoops(t *testing.T) {
	s, err := Parse("while i < 3\nprint i\nend\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w, ok := s.Stmts[0].(*ast.WhileStmt)
	if !ok || len(w.Body) != 1 {
		t.Fatalf("while shape wrong: %#v", s.Stmts[0])
	}
}

func TestCommandPhrases(t *testing.T) {
	s, err := Parse("git commit -m \"msg\" --force\nlet c = run deploy prod\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd, ok := s.Stmts[0].(*ast.CmdStmt)
	if !ok || cmd.Name != "git" {
		t.Fatalf("got %#v, want CmdStmt git", s.Stmts[0])
	}
	want := []string{"commit", "-m", "msg", "--force"}
	if len(cmd.Args) != len(want) {
		t.Fatalf("args: got %v, want %v", cmd.Args, want)
	}
	for i := range want {
		if cmd.Args[i] != want[i] {
			t.Fatalf("arg %d: got %q, want %q", i, cmd.Args[i], want[i])
		}
	}
	let := s.Stmts[1].(*ast.LetStmt)
	run, ok := let.Value.(*ast.RunExpr)
	if !ok || run.Name != "deploy" || len(run.Args) != 1 || run.Args[0] != "prod" {
		t.Fatalf("run shape wrong: %#v", let.Value)
	}
}

func TestOpenBlocks(t *testing.T) {
	cases := []struct {
		src  string
		want int
	}{
		{"let x = 1\n", 0},
		{"if true then print 1 end\n", 0},
		{"if x > 1 then\n", 1},
		{"fn f()\n", 1},
		{"while x\nif y then\n", 2},
		{"# if then end inside a comment\nlet s = \"if fake then\"\n", 0},
	}
	for _, c := range cases {
		if got := OpenBlocks(c.src); got != c.want {
			t.Errorf("OpenBlocks(%q) = %d, want %d", c.src, got, c.want)
		}
	}
}

func TestCommandErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"run\n", `1:4: expected command name after run, got "\n"`},
		{"git --\n", `1:7: commands take simple words only (strings, numbers, flags), got "\n" — for expressions use print/if/let`},
	}
	for _, c := range cases {
		_, perr := Parse(c.src)
		if perr == nil {
			t.Errorf("%q: expected error", c.src)
			continue
		}
		if perr.Error() != c.want {
			t.Errorf("%q: got %q, want %q", c.src, perr.Error(), c.want)
		}
	}
}

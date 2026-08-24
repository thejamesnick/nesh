package parser

import (
	"strings"
	"testing"

	"nesh/internal/ast"
)

func TestLetStatements(t *testing.T) {
	s, err := Parse("let x = 5\nlet name = \"Nick\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Stmts) != 2 {
		t.Fatalf("got %d statements, want 2", len(s.Stmts))
	}
	let, ok := s.Stmts[0].(*ast.LetStmt)
	if !ok || let.Name != "x" {
		t.Fatalf("stmt 0: got %#v, want let x", s.Stmts[0])
	}
	if lit, ok := let.Value.(*ast.IntLit); !ok || lit.Value != 5 {
		t.Fatalf("let value: got %#v, want IntLit 5", let.Value)
	}
	name, ok := s.Stmts[1].(*ast.LetStmt)
	if !ok || name.Name != "name" {
		t.Fatalf("stmt 1: got %#v", s.Stmts[1])
	}
	if lit, ok := name.Value.(*ast.StringLit); !ok || lit.Value != "Nick" {
		t.Fatalf("let value: got %#v, want StringLit Nick", name.Value)
	}
}

func TestPrecedence(t *testing.T) {
	// 1 + 2 * 3 must group as 1 + (2 * 3)
	s, err := Parse("let x = 1 + 2 * 3\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	let := s.Stmts[0].(*ast.LetStmt)
	sum, ok := let.Value.(*ast.InfixExpr)
	if !ok || sum.Op != "+" {
		t.Fatalf("root: got %#v, want +", let.Value)
	}
	if l, ok := sum.L.(*ast.IntLit); !ok || l.Value != 1 {
		t.Fatalf("left of +: got %#v, want 1", sum.L)
	}
	prod, ok := sum.R.(*ast.InfixExpr)
	if !ok || prod.Op != "*" {
		t.Fatalf("right of +: got %#v, want *", sum.R)
	}

	// Parens override: (1 + 2) * 3 groups as (1+2) * 3
	s, err = Parse("let x = (1 + 2) * 3\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	let = s.Stmts[0].(*ast.LetStmt)
	prod, ok = let.Value.(*ast.InfixExpr)
	if !ok || prod.Op != "*" {
		t.Fatalf("root: got %#v, want *", let.Value)
	}
	if _, ok := prod.L.(*ast.InfixExpr); !ok {
		t.Fatalf("left of *: got %#v, want infix (1+2)", prod.L)
	}
}

func TestUnaryMinus(t *testing.T) {
	s, err := Parse("let x = -5\nlet y = -x * 2\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	neg, ok := s.Stmts[0].(*ast.LetStmt).Value.(*ast.PrefixExpr)
	if !ok || neg.Op != "-" {
		t.Fatalf("got %#v, want prefix -", s.Stmts[0].(*ast.LetStmt).Value)
	}
	if lit, ok := neg.Right.(*ast.IntLit); !ok || lit.Value != 5 {
		t.Fatalf("negand: got %#v, want 5", neg.Right)
	}
}

func TestPrintArgs(t *testing.T) {
	s, err := Parse("print \"hello\" name\nprint 1 + 2\nprint\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p0 := s.Stmts[0].(*ast.PrintStmt)
	if len(p0.Args) != 2 {
		t.Fatalf("print 0: got %d args, want 2", len(p0.Args))
	}
	p1 := s.Stmts[1].(*ast.PrintStmt)
	if len(p1.Args) != 1 {
		t.Fatalf("print 1: got %d args, want 1 (whole expression)", len(p1.Args))
	}
	if _, ok := p1.Args[0].(*ast.InfixExpr); !ok {
		t.Fatalf("print 1 arg: got %#v, want infix +", p1.Args[0])
	}
	if len(s.Stmts[2].(*ast.PrintStmt).Args) != 0 {
		t.Fatalf("bare print should have zero args")
	}
}

func TestErrorPositions(t *testing.T) {
	cases := []struct {
		src  string
		line int
	}{
		{"let x = 5\nlet = 3\n", 2}, // let without name
		{"let x 5\n", 1},            // missing =
		{"let x = \n", 1},           // missing value
		{"print 1 +\n", 1},          // dangling operator
		{"let x = (1 + 2\n", 1},     // unclosed paren
		{"let\n", 1},                // let without name
	}
	for _, c := range cases {
		_, err := Parse(c.src)
		if err == nil {
			t.Errorf("input %q: expected error, got none", c.src)
			continue
		}
		if err.Line != c.line {
			t.Errorf("input %q: error at line %d, want %d (%s)", c.src, err.Line, c.line, err.Msg)
		}
	}
}

func TestCommentsAndBlankLinesIgnored(t *testing.T) {
	s, err := Parse("# header\n\nlet x = 1\n\n# trailer\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Stmts) != 1 {
		t.Fatalf("got %d statements, want 1", len(s.Stmts))
	}
}

func TestEmptyScript(t *testing.T) {
	s, err := Parse("")
	if err != nil || s == nil || len(s.Stmts) != 0 {
		t.Fatalf("empty input should parse to empty script, got %v, %v", s, err)
	}
}

func TestBreakContinue(t *testing.T) {
	s, err := Parse("while true\nbreak\nend\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := s.Stmts[0].(*ast.WhileStmt)
	if _, ok := w.Body[0].(*ast.BreakStmt); !ok {
		t.Fatalf("got %T, want *ast.BreakStmt", w.Body[0])
	}

	s, err = Parse("for x in items\ncontinue\nend\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := s.Stmts[0].(*ast.ForStmt)
	if _, ok := f.Body[0].(*ast.ContinueStmt); !ok {
		t.Fatalf("got %T, want *ast.ContinueStmt", f.Body[0])
	}
}

func TestBreakContinueOutsideLoop(t *testing.T) {
	cases := []struct {
		src  string
		line int
	}{
		{"break\n", 1},
		{"continue\n", 1},
		{"if true then\nbreak\nend\n", 2},
		// fn bodies execute later — an enclosing loop must not leak in
		{"while true\nfn f()\nbreak\nend\nend\n", 3},
	}
	for _, c := range cases {
		_, err := Parse(c.src)
		if err == nil {
			t.Errorf("input %q: expected error, got none", c.src)
			continue
		}
		if err.Line != c.line {
			t.Errorf("input %q: error at line %d, want %d (%s)", c.src, err.Line, c.line, err.Msg)
		}
	}
}

func TestRedirects(t *testing.T) {
	s, err := Parse("git log > out.txt\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := s.Stmts[0].(*ast.CmdStmt)
	want := []ast.Redirect{{Op: ">", Path: "out.txt"}}
	if len(cmd.Redirects) != 1 || cmd.Redirects[0] != want[0] {
		t.Fatalf("got %+v, want %+v", cmd.Redirects, want)
	}

	// mixed args + multiple redirects, quoted paths, run expression
	s, err = Parse("let n = run cat < \"in file.txt\" > counts.txt >> totals.txt\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	let := s.Stmts[0].(*ast.LetStmt)
	runE := let.Value.(*ast.RunExpr)
	got := runE.Redirects
	wantAll := []ast.Redirect{
		{Op: "<", Path: "in file.txt"},
		{Op: ">", Path: "counts.txt"},
		{Op: ">>", Path: "totals.txt"},
	}
	if len(got) != 3 {
		t.Fatalf("got %d redirects, want 3: %+v", len(got), got)
	}
	for i, w := range wantAll {
		if got[i] != w {
			t.Errorf("redirect %d: got %+v, want %+v", i, got[i], w)
		}
	}
}

func TestRedirectErrors(t *testing.T) {
	cases := []struct {
		src string
		col int
	}{
		{"git log >\n", 10},       // missing path: error points at the newline
		{"cat < # comment\n", 16}, // missing path: comment skipped, error at newline
	}
	for _, c := range cases {
		_, err := Parse(c.src)
		if err == nil {
			t.Errorf("input %q: expected error, got none", c.src)
			continue
		}
		if err.Column != c.col {
			t.Errorf("input %q: error at col %d, want %d (%s)", c.src, err.Column, c.col, err.Msg)
		}
	}
}

func TestPipeline(t *testing.T) {
	s, err := Parse("cat log.txt | grep error | wc -l\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pl := s.Stmts[0].(*ast.PipelineStmt)
	if len(pl.Stages) != 3 {
		t.Fatalf("got %d stages, want 3", len(pl.Stages))
	}
	wantNames := []string{"cat", "grep", "wc"}
	wantArgs := [][]string{{"log.txt"}, {"error"}, {"-l"}}
	for i, st := range pl.Stages {
		if st.Name != wantNames[i] {
			t.Errorf("stage %d name: got %q, want %q", i, st.Name, wantNames[i])
		}
		if strings.Join(st.Args, ",") != strings.Join(wantArgs[i], ",") {
			t.Errorf("stage %d args: got %v, want %v", i, st.Args, wantArgs[i])
		}
	}

	// run expression with pipe stages
	s, err = Parse("let n = run git log | grep fix | wc -l\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	runE := ((s.Stmts[0].(*ast.LetStmt)).Value).(*ast.RunExpr)
	if runE.Name != "git" || len(runE.Pipe) != 2 || runE.Pipe[1].Name != "wc" {
		t.Fatalf("run pipeline wrong: %+v", runE)
	}

	// redirects mix with pipes per stage
	s, err = Parse("git log > all.txt | grep fix < errs.txt\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pl = s.Stmts[0].(*ast.PipelineStmt)
	if len(pl.Stages[0].Redirects) != 1 || pl.Stages[0].Redirects[0].Op != ">" {
		t.Fatalf("stage 0 redirects: %+v", pl.Stages[0].Redirects)
	}
	if len(pl.Stages[1].Redirects) != 1 || pl.Stages[1].Redirects[0].Op != "<" {
		t.Fatalf("stage 1 redirects: %+v", pl.Stages[1].Redirects)
	}
}

func TestPipelineErrors(t *testing.T) {
	for _, src := range []string{"git log |\n", "a | | b\n"} {
		_, err := Parse(src)
		if err == nil {
			t.Errorf("input %q: expected error, got none", src)
		}
	}
}

func TestCaptureExpr(t *testing.T) {
	s, err := Parse("let x = capture git branch --show-current\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	capE := s.Stmts[0].(*ast.LetStmt).Value.(*ast.CaptureExpr)
	if capE.Name != "git" || len(capE.Args) != 2 || capE.Args[0] != "branch" || capE.Args[1] != "--show-current" {
		t.Fatalf("capture expr wrong: %+v", capE)
	}

	// capture with pipeline stages lands in Pipe
	s, err = Parse("let n = capture git log | grep fix | wc -l\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	capE2 := s.Stmts[0].(*ast.LetStmt).Value.(*ast.CaptureExpr)
	if capE2.Name != "git" || len(capE2.Pipe) != 2 || capE2.Pipe[1].Name != "wc" {
		t.Fatalf("capture pipeline wrong: %+v", capE2)
	}

	// capture works in any expression position, not just let
	s, err = Parse("print capture echo hi\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arg, ok := s.Stmts[0].(*ast.PrintStmt).Args[0].(*ast.CaptureExpr)
	if !ok || arg.Name != "echo" || len(arg.Args) != 1 || arg.Args[0] != "hi" {
		t.Fatalf("print arg: got %#v, want CaptureExpr(echo hi)", s.Stmts[0].(*ast.PrintStmt).Args[0])
	}

	// bare capture is a statement (mirrors run) and discards the value
	s, err = Parse("capture git status\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Stmts[0].(*ast.ExprStmt).Expr.(*ast.CaptureExpr); !ok {
		t.Fatalf("bare capture statement: got %#v", s.Stmts[0])
	}

	// capture with nothing after it errors
	if _, err := Parse("let x = capture\n"); err == nil {
		t.Fatal("expected error for bare capture")
	}
	if _, err := Parse("let x = capture 42\n"); err == nil {
		t.Fatal("expected error for numeric command after capture")
	}
}

func TestTryOnFailure(t *testing.T) {
	s, err := Parse("try\ndeploy\non failure\nprint failure\nend\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tr := s.Stmts[0].(*ast.TryStmt)
	if len(tr.Try) != 1 || !tr.HasOn || len(tr.On) != 1 {
		t.Fatalf("shape wrong: %+v", tr)
	}

	// handler optional
	s, err = Parse("try\nrisky\nend\nprint \"after\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Stmts[0].(*ast.TryStmt); !ok {
		t.Fatalf("got %T, want TryStmt", s.Stmts[0])
	}
	if len(s.Stmts) != 2 {
		t.Fatalf("statement after try lost")
	}

	for _, src := range []string{
		"try\nx\non boom\nend\n", // wrong word after on
		"try\nx\non failure\n",   // missing end
	} {
		if _, err := Parse(src); err == nil {
			t.Errorf("input %q: expected error, got none", src)
		}
	}
}

func TestFailStatement(t *testing.T) {
	s, err := Parse("fail \"db down\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fl := s.Stmts[0].(*ast.FailStmt)
	if fl.Msg == nil {
		t.Fatal("message lost")
	}

	// bare fail parses too
	if _, err := Parse("fail\n"); err != nil {
		t.Fatalf("bare fail: %v", err)
	}
}

func TestOpenBlocksCountsTry(t *testing.T) {
	if got := OpenBlocks("try\nx = 1\n"); got != 1 {
		t.Fatalf("open try: got %d, want 1", got)
	}
	if got := OpenBlocks("try\nif true then\ny = 2\nend\n"); got != 1 {
		t.Fatalf("try+if closed once: got %d, want 1", got)
	}
	if got := OpenBlocks("try\nif true then\ny = 2\nend\nend\n"); got != 0 {
		t.Fatalf("fully closed: got %d, want 0", got)
	}
	if got := OpenBlocks("try\nx\non failure\ny\nend\n"); got != 0 {
		t.Fatalf("closed try: got %d, want 0", got)
	}
}

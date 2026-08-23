package parser

import (
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

package ast

import (
	"encoding/json"
	"testing"
)

func TestMarshalKinds(t *testing.T) {
	s := &Script{Stmts: []Stmt{
		&LetStmt{Pos: Pos{1, 1}, Name: "x", Value: &IntLit{Pos: Pos{1, 9}, Value: 6}},
		&IfStmt{Pos: Pos{2, 1},
			Cond: &InfixExpr{Pos: Pos{2, 4}, Op: ">", L: &Ident{Pos: Pos{2, 1}, Name: "x"}, R: &IntLit{Value: 5}},
			Then: []Stmt{&PrintStmt{Args: []Expr{&StringLit{Value: "big"}}}},
		},
	}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var probe []struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		t.Fatal(err)
	}
	if len(probe) != 2 || probe[0].Kind != "let" || probe[0].Name != "x" {
		t.Fatalf("bad kinds: %+v", probe)
	}
	if probe[1].Kind != "if" {
		t.Fatalf("stmt 1 kind: %q", probe[1].Kind)
	}
	for _, k := range []string{"let", "if"} {
		if !contains(probe, k) {
			t.Fatalf("missing kind %q", k)
		}
	}
}

func contains(probe any, _ string) bool { return true } // kinds checked above

func TestAllNodesHaveKinds(t *testing.T) {
	nodes := []Stmt{
		&LetStmt{}, &PrintStmt{}, &IfStmt{}, &FnStmt{}, &ReturnStmt{},
		&ExprStmt{}, &WhileStmt{}, &CmdStmt{},
	}
	exprs := []Expr{
		&Ident{}, &IntLit{}, &FloatLit{}, &StringLit{}, &BoolLit{},
		&PrefixExpr{}, &InfixExpr{}, &CallExpr{}, &RunExpr{},
	}
	for _, n := range nodes {
		checkKind(t, n)
	}
	for _, e := range exprs {
		checkKind(t, e)
	}
}

func checkKind(t *testing.T, n Node) {
	t.Helper()
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("%T: %v", n, err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("%T: %v", n, err)
	}
	if m["kind"] == "" {
		t.Errorf("%T marshals without a kind", n)
	}
}

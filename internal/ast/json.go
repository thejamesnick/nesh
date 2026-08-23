// Package ast — JSON encoding for the agent API (nesh --json).
//
// Every node marshals as its fields plus a "kind" discriminator so
// consumers can walk the tree without Go type knowledge.
package ast

import "encoding/json"

// withKind marshals node, then injects the kind discriminator.
func withKind(kind string, node any) ([]byte, error) {
	b, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m["kind"] = kind
	return json.Marshal(m)
}

// Pos marshals as {"line":L,"column":C}.
func (p Pos) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}{p.Line, p.Column})
}

func (s *Script) MarshalJSON() ([]byte, error) { return json.Marshal(s.Stmts) }

func (s *LetStmt) MarshalJSON() ([]byte, error)    { return withKind("let", *s) }
func (s *PrintStmt) MarshalJSON() ([]byte, error)  { return withKind("print", *s) }
func (s *IfStmt) MarshalJSON() ([]byte, error)     { return withKind("if", *s) }
func (s *FnStmt) MarshalJSON() ([]byte, error)     { return withKind("fn", *s) }
func (s *ReturnStmt) MarshalJSON() ([]byte, error) { return withKind("return", *s) }
func (s *ExprStmt) MarshalJSON() ([]byte, error)   { return withKind("expr", *s) }
func (s *WhileStmt) MarshalJSON() ([]byte, error)  { return withKind("while", *s) }
func (s *ForStmt) MarshalJSON() ([]byte, error)    { return withKind("for", *s) }
func (s *CmdStmt) MarshalJSON() ([]byte, error)    { return withKind("command", *s) }
func (e *Ident) MarshalJSON() ([]byte, error)      { return withKind("ident", *e) }
func (e *IntLit) MarshalJSON() ([]byte, error)     { return withKind("int", *e) }
func (e *FloatLit) MarshalJSON() ([]byte, error)   { return withKind("float", *e) }
func (e *StringLit) MarshalJSON() ([]byte, error)  { return withKind("string", *e) }
func (e *BoolLit) MarshalJSON() ([]byte, error)    { return withKind("bool", *e) }
func (e *PrefixExpr) MarshalJSON() ([]byte, error) { return withKind("prefix", *e) }
func (e *InfixExpr) MarshalJSON() ([]byte, error)  { return withKind("infix", *e) }
func (e *CallExpr) MarshalJSON() ([]byte, error)   { return withKind("call", *e) }
func (e *RunExpr) MarshalJSON() ([]byte, error)    { return withKind("run", *e) }

// Package parser turns Nesh tokens into an AST.
//
// Contract (MODULES.md): never executes, never touches the OS.
// All errors carry source position.
package parser

import (
	"fmt"
	"strconv"

	"nesh/internal/ast"
	"nesh/internal/lexer"
	"nesh/internal/token"
)

// Error is a parse failure with the position where it was detected.
type Error struct {
	Line, Column int
	Msg          string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Msg)
}

// precedence levels; higher binds tighter.
const (
	precLowest  = 0
	precOr      = 1 // or
	precAnd     = 2 // and
	precCompare = 3 // == != < > <= >=
	precSum     = 4 // + -
	precProduct = 5 // * /
)

var precedences = map[token.Type]int{
	token.PLUS:     precSum,
	token.MINUS:    precSum,
	token.ASTERISK: precProduct,
	token.SLASH:    precProduct,
	token.EQ:       precCompare,
	token.NOT_EQ:   precCompare,
	token.LT:       precCompare,
	token.GT:       precCompare,
	token.LTE:      precCompare,
	token.GTE:      precCompare,
	token.AND:      precAnd,
	token.OR:       precOr,
}

type parser struct {
	l    *lexer.Lexer
	cur  token.Token
	peek token.Token
	err  *Error
}

// Parse lexes and parses src into a Script, returning the first parse
// error if any.
func Parse(src string) (*ast.Script, *Error) {
	p := &parser{l: lexer.New(src)}
	p.next() // cur
	p.next() // peek
	s := &ast.Script{}
	for !p.at(token.EOF) {
		if p.at(token.NEWLINE) {
			p.next()
			continue
		}
		stmt, ok := p.parseStmt()
		if !ok {
			return nil, p.err
		}
		s.Stmts = append(s.Stmts, stmt)
		if !p.expectStatementEnd() {
			return nil, p.err
		}
	}
	return s, nil
}

func (p *parser) next() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *parser) at(t token.Type) bool { return p.cur.Type == t }

func (p *parser) fail(format string, args ...any) {
	if p.err == nil {
		p.err = &Error{Line: p.cur.Line, Column: p.cur.Column, Msg: fmt.Sprintf(format, args...)}
	}
}

func (p *parser) parseStmt() (ast.Stmt, bool) {
	switch p.cur.Type {
	case token.LET:
		return p.parseLet()
	case token.PRINT:
		return p.parsePrint()
	case token.IF:
		return p.parseIf()
	default:
		p.fail("expected statement (let, print, or if), got %q", p.cur.Literal)
		return nil, false
	}
}

func (p *parser) parseLet() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next()
	if !p.at(token.IDENT) {
		p.fail("expected variable name after let, got %q", p.cur.Literal)
		return nil, false
	}
	name := p.cur.Literal
	p.next()
	if !p.at(token.ASSIGN) {
		p.fail("expected = after let %s, got %q", name, p.cur.Literal)
		return nil, false
	}
	p.next()
	value, ok := p.parseExpr(precLowest)
	if !ok {
		return nil, false
	}
	return &ast.LetStmt{Pos: pos, Name: name, Value: value}, true
}

func (p *parser) parsePrint() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next()
	stmt := &ast.PrintStmt{Pos: pos}
	for !p.at(token.NEWLINE) && !p.at(token.EOF) {
		arg, ok := p.parseExpr(precLowest)
		if !ok {
			return nil, false
		}
		stmt.Args = append(stmt.Args, arg)
	}
	return stmt, true
}

func (p *parser) expectStatementEnd() bool {
	if p.at(token.NEWLINE) || p.at(token.EOF) {
		if p.at(token.NEWLINE) {
			p.next()
		}
		return true
	}
	// Inside a block, a block keyword (else/elif/end) legally terminates
	// the last statement; the block parser consumes it.
	switch p.cur.Type {
	case token.ELSE, token.ELIF, token.END:
		return true
	}
	p.fail("expected end of statement (newline), got %q", p.cur.Literal)
	return false
}

// parseBlock parses statements until one of the stop tokens (or EOF),
// skipping blank lines. The stop token is left as cur.
func (p *parser) parseBlock(stop ...token.Type) ([]ast.Stmt, bool) {
	var stmts []ast.Stmt
	for {
		for p.at(token.NEWLINE) {
			p.next()
		}
		done := p.at(token.EOF)
		for _, s := range stop {
			if p.at(s) {
				done = true
				break
			}
		}
		if done {
			return stmts, true
		}
		stmt, ok := p.parseStmt()
		if !ok {
			return nil, false
		}
		stmts = append(stmts, stmt)
		if !p.expectStatementEnd() {
			return nil, false
		}
	}
}

func (p *parser) parseIf() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	node := &ast.IfStmt{Pos: pos}
	p.next() // consume 'if'
	if !p.parseIfHead(node) {
		return nil, false
	}
	cur := node
	for {
		thenStmts, ok := p.parseBlock(token.ELSE, token.ELIF, token.END)
		if !ok {
			return nil, false
		}
		cur.Then = thenStmts
		switch {
		case p.at(token.ELIF):
			elifPos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
			nested := &ast.IfStmt{Pos: elifPos}
			p.next() // consume 'elif'
			if !p.parseIfHead(nested) {
				return nil, false
			}
			cur.Else = []ast.Stmt{nested}
			cur = nested
		case p.at(token.ELSE):
			p.next()
			elseStmts, ok := p.parseBlock(token.END)
			if !ok {
				return nil, false
			}
			cur.Else = elseStmts
			if !p.expect(token.END) {
				return nil, false
			}
			return node, true
		default:
			if !p.expect(token.END) {
				return nil, false
			}
			return node, true
		}
	}
}

// parseIfHead parses `if cond then` / `elif cond then`, consuming through
// the `then`.
func (p *parser) parseIfHead(node *ast.IfStmt) bool {
	cond, ok := p.parseExpr(precLowest)
	if !ok {
		return false
	}
	node.Cond = cond
	if !p.expect(token.THEN) {
		return false
	}
	return true
}

func (p *parser) expect(t token.Type) bool {
	if !p.at(t) {
		p.fail("expected %q, got %q", t, p.cur.Literal)
		return false
	}
	p.next()
	return true
}

// parseExpr is precedence climbing: parse a unary operand, then fold
// infix operators while they bind at least as tightly as minPrec.
func (p *parser) parseExpr(minPrec int) (ast.Expr, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return nil, false
	}
	for {
		prec, found := precedences[p.cur.Type]
		if !found || prec < minPrec {
			return left, true
		}
		opPos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
		op := p.cur.Literal
		p.next()
		right, ok := p.parseExpr(prec + 1)
		if !ok {
			return nil, false
		}
		left = &ast.InfixExpr{Pos: opPos, Op: op, L: left, R: right}
	}
}

func (p *parser) parseUnary() (ast.Expr, bool) {
	if p.at(token.MINUS) || p.at(token.NOT) {
		pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
		op := p.cur.Literal
		p.next()
		right, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		return &ast.PrefixExpr{Pos: pos, Op: op, Right: right}, true
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (ast.Expr, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	switch p.cur.Type {
	case token.IDENT:
		name := p.cur.Literal
		p.next()
		return &ast.Ident{Pos: pos, Name: name}, true
	case token.INT:
		v, err := strconv.ParseInt(p.cur.Literal, 10, 64)
		if err != nil {
			p.fail("invalid integer %q", p.cur.Literal)
			return nil, false
		}
		p.next()
		return &ast.IntLit{Pos: pos, Value: v}, true
	case token.FLOAT:
		v, err := strconv.ParseFloat(p.cur.Literal, 64)
		if err != nil {
			p.fail("invalid number %q", p.cur.Literal)
			return nil, false
		}
		p.next()
		return &ast.FloatLit{Pos: pos, Value: v}, true
	case token.STRING:
		v := p.cur.Literal
		p.next()
		return &ast.StringLit{Pos: pos, Value: v}, true
	case token.TRUE, token.FALSE:
		v := p.cur.Type == token.TRUE
		p.next()
		return &ast.BoolLit{Pos: pos, Value: v}, true
	case token.LPAREN:
		p.next()
		inner, ok := p.parseExpr(precLowest)
		if !ok {
			return nil, false
		}
		if !p.at(token.RPAREN) {
			p.fail("expected ) to close (, got %q", p.cur.Literal)
			return nil, false
		}
		p.next()
		return inner, true
	default:
		p.fail("expected expression, got %q", p.cur.Literal)
		return nil, false
	}
}

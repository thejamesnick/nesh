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

// OpenBlocks reports how many if/fn/while blocks are still open at the
// end of src. The REPL uses it to decide whether to keep reading lines.
func OpenBlocks(src string) int {
	l := lexer.New(src)
	depth := 0
	for {
		t := l.NextToken()
		switch t.Type {
		case token.EOF:
			return depth
		case token.IF, token.FN, token.WHILE, token.FOR:
			depth++
		case token.END:
			depth--
		}
	}
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
	case token.FN:
		return p.parseFn()
	case token.WHILE:
		return p.parseWhile()
	case token.FOR:
		return p.parseFor()
	case token.IMPORT:
		return p.parseImport()
	case token.RETURN:
		return p.parseReturn()
	case token.IDENT:
		if p.peek.Type == token.LPAREN {
			call, ok := p.parseCall(p.cur.Literal, ast.Pos{Line: p.cur.Line, Column: p.cur.Column})
			if !ok {
				return nil, false
			}
			return &ast.ExprStmt{Pos: call.Position(), Expr: call}, true
		}
		return p.parseCmdPhrase(p.cur.Literal)
	case token.RUN:
		pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
		runExpr, ok := p.parseRunExpr()
		if !ok {
			return nil, false
		}
		return &ast.ExprStmt{Pos: pos, Expr: runExpr}, true
	default:
		p.fail("expected statement (let, print, if, or fn), got %q", p.cur.Literal)
		return nil, false
	}
}

func (p *parser) parseFn() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume 'fn'
	if !p.at(token.IDENT) {
		p.fail("expected function name after fn, got %q", p.cur.Literal)
		return nil, false
	}
	name := p.cur.Literal
	p.next()
	if !p.expect(token.LPAREN) {
		return nil, false
	}
	var params []string
	if !p.at(token.RPAREN) {
		for {
			if !p.at(token.IDENT) {
				p.fail("expected parameter name, got %q", p.cur.Literal)
				return nil, false
			}
			params = append(params, p.cur.Literal)
			p.next()
			if p.at(token.COMMA) {
				p.next()
				continue
			}
			break
		}
	}
	if !p.expect(token.RPAREN) {
		return nil, false
	}
	body, ok := p.parseBlock(token.END)
	if !ok {
		return nil, false
	}
	if !p.expect(token.END) {
		return nil, false
	}
	return &ast.FnStmt{Pos: pos, Name: name, Params: params, Body: body}, true
}

// parseCmdPhrase parses a bare command: `git commit -m "msg"`.
// cur is the command-name token. Words are literals; leading dashes merge
// into flags (`-m`, `--force`) because the lexer splits them.
func (p *parser) parseCmdPhrase(name string) (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume the name
	stmt := &ast.CmdStmt{Pos: pos, Name: name}
	for !p.atStatementBoundary() {
		word, ok := p.parseCmdWord()
		if !ok {
			return nil, false
		}
		stmt.Args = append(stmt.Args, word)
	}
	return stmt, true
}

// parseRunExpr parses `run <command>` in expression position.
func (p *parser) parseRunExpr() (ast.Expr, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume 'run'
	if p.atStatementBoundary() || p.cur.Type != token.IDENT {
		p.fail("expected command name after run, got %q", p.cur.Literal)
		return nil, false
	}
	name := p.cur.Literal
	p.next()
	expr := &ast.RunExpr{Pos: pos, Name: name}
	for !p.atStatementBoundary() {
		word, ok := p.parseCmdWord()
		if !ok {
			return nil, false
		}
		expr.Args = append(expr.Args, word)
	}
	return expr, true
}

func (p *parser) atStatementBoundary() bool {
	switch p.cur.Type {
	case token.NEWLINE, token.EOF, token.END, token.ELSE, token.ELIF:
		return true
	}
	return false
}

func (p *parser) parseCmdWord() (string, bool) {
	dashes := ""
	for p.at(token.MINUS) {
		dashes += "-"
		p.next()
	}
	switch p.cur.Type {
	case token.IDENT, token.INT, token.FLOAT, token.STRING:
		word := dashes + p.cur.Literal
		p.next()
		return word, true
	default:
		p.fail("commands take simple words only (strings, numbers, flags), got %q — for expressions use print/if/let", p.cur.Literal)
		return "", false
	}
}

func (p *parser) parseImport() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume 'import'
	if p.cur.Type != token.STRING {
		p.fail("expected module path string after import, got %q", p.cur.Literal)
		return nil, false
	}
	path := p.cur.Literal
	p.next()
	alias := ""
	if p.at(token.AS) {
		p.next()
		if !p.at(token.IDENT) {
			p.fail("expected alias name after as, got %q", p.cur.Literal)
			return nil, false
		}
		alias = p.cur.Literal
		p.next()
	}
	return &ast.ImportStmt{Pos: pos, Path: path, Alias: alias}, true
}

func (p *parser) parseWhile() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume 'while'
	cond, ok := p.parseExpr(precLowest)
	if !ok {
		return nil, false
	}
	body, ok := p.parseBlock(token.END)
	if !ok {
		return nil, false
	}
	if !p.expect(token.END) {
		return nil, false
	}
	return &ast.WhileStmt{Pos: pos, Cond: cond, Body: body}, true
}

func (p *parser) parseFor() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	p.next() // consume 'for'
	if !p.at(token.IDENT) {
		p.fail("expected loop variable after for, got %q", p.cur.Literal)
		return nil, false
	}
	varName := p.cur.Literal
	p.next()
	if !p.expect(token.IN) {
		return nil, false
	}
	iter, ok := p.parseExpr(precLowest)
	if !ok {
		return nil, false
	}
	body, ok := p.parseBlock(token.END)
	if !ok {
		return nil, false
	}
	if !p.expect(token.END) {
		return nil, false
	}
	return &ast.ForStmt{Pos: pos, Var: varName, Iter: iter, Body: body}, true
}

func (p *parser) parseReturn() (ast.Stmt, bool) {
	pos := ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
	node := &ast.ReturnStmt{Pos: pos}
	p.next() // consume 'return'
	switch p.cur.Type {
	case token.NEWLINE, token.EOF, token.END, token.ELSE, token.ELIF:
		return node, true
	}
	v, ok := p.parseExpr(precLowest)
	if !ok {
		return nil, false
	}
	node.Value = v
	return node, true
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

// parseCall parses `name(arg1, ...)`; cur is the name token, peek is '('.
func (p *parser) parseCall(name string, pos ast.Pos) (ast.Expr, bool) {
	call := &ast.CallExpr{Pos: pos, Name: name}
	p.next() // now at '('
	p.next() // past '('
	if !p.at(token.RPAREN) {
		for {
			arg, ok := p.parseExpr(precLowest)
			if !ok {
				return nil, false
			}
			call.Args = append(call.Args, arg)
			if p.at(token.COMMA) {
				p.next()
				continue
			}
			break
		}
	}
	if !p.at(token.RPAREN) {
		p.fail("expected ) to close call of %s, got %q", name, p.cur.Literal)
		return nil, false
	}
	p.next()
	return call, true
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
		if p.peek.Type == token.LPAREN {
			return p.parseCall(name, pos)
		}
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
	case token.UNTERM:
		p.fail("unterminated string — missing closing quote")
		return nil, false
	case token.TRUE, token.FALSE:
		v := p.cur.Type == token.TRUE
		p.next()
		return &ast.BoolLit{Pos: pos, Value: v}, true
	case token.RUN:
		return p.parseRunExpr()
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

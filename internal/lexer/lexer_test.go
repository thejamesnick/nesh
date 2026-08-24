package lexer

import (
	"testing"

	"nesh/internal/token"
)

// scan lexes input to completion and returns all tokens (EOF included).
func scan(input string) []token.Token {
	l := New(input)
	var toks []token.Token
	for {
		t := l.NextToken()
		toks = append(toks, t)
		if t.Type == token.EOF {
			return toks
		}
	}
}

type tok struct {
	wantType token.Type
	literal  string
}

func assertTokens(t *testing.T, input string, want []tok) {
	t.Helper()
	got := scan(input)
	if len(got) != len(want) {
		t.Fatalf("input %q: got %d tokens, want %d\ngot: %v", input, len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Type != w.wantType || got[i].Literal != w.literal {
			t.Fatalf("input %q token %d: got %s, want %s %q",
				input, i, got[i], w.wantType, w.literal)
		}
	}
}

func TestKeywordsAndIdentifiers(t *testing.T) {
	assertTokens(t, "let print if then else end fn return for in while break continue",
		[]tok{
			{token.LET, "let"}, {token.PRINT, "print"},
			{token.IF, "if"}, {token.THEN, "then"},
			{token.ELSE, "else"}, {token.END, "end"},
			{token.FN, "fn"}, {token.RETURN, "return"},
			{token.FOR, "for"}, {token.IN, "in"}, {token.WHILE, "while"},
			{token.BREAK, "break"}, {token.CONT, "continue"},
			{token.EOF, ""},
		})
	assertTokens(t, "let name = \"John\"",
		[]tok{
			{token.LET, "let"}, {token.IDENT, "name"}, {token.ASSIGN, "="},
			{token.STRING, "John"}, {token.EOF, ""},
		})
}

func TestOperatorsAndDelimiters(t *testing.T) {
	assertTokens(t, "+ - * /",
		[]tok{
			{token.PLUS, "+"}, {token.MINUS, "-"},
			{token.ASTERISK, "*"}, {token.SLASH, "/"}, {token.EOF, ""},
		})
	assertTokens(t, "== != < > <= >=",
		[]tok{
			{token.EQ, "=="}, {token.NOT_EQ, "!="},
			{token.LT, "<"}, {token.GT, ">"},
			{token.LTE, "<="}, {token.GTE, ">="}, {token.EOF, ""},
		})
	assertTokens(t, "(a, b)",
		[]tok{
			{token.LPAREN, "("}, {token.IDENT, "a"}, {token.COMMA, ","},
			{token.IDENT, "b"}, {token.RPAREN, ")"}, {token.EOF, ""},
		})
	// '=' must not merge with a following token; '==' must.
	assertTokens(t, "a==b = c",
		[]tok{
			{token.IDENT, "a"}, {token.EQ, "=="}, {token.IDENT, "b"},
			{token.ASSIGN, "="}, {token.IDENT, "c"}, {token.EOF, ""},
		})
}

func TestNumbers(t *testing.T) {
	assertTokens(t, "42", []tok{{token.INT, "42"}, {token.EOF, ""}})
	assertTokens(t, "3.14", []tok{{token.FLOAT, "3.14"}, {token.EOF, ""}})
	// Trailing dot is not part of the number.
	assertTokens(t, "3.",
		[]tok{{token.INT, "3"}, {token.Type("."), "."}, {token.EOF, ""}})
	// Negative numbers lex as MINUS INT; sign folding is the parser's job.
	assertTokens(t, "-5",
		[]tok{{token.MINUS, "-"}, {token.INT, "5"}, {token.EOF, ""}})
	assertTokens(t, "1+2.5",
		[]tok{
			{token.INT, "1"}, {token.PLUS, "+"}, {token.FLOAT, "2.5"},
			{token.EOF, ""},
		})
}

func TestStrings(t *testing.T) {
	assertTokens(t, `"hello world"`,
		[]tok{{token.STRING, "hello world"}, {token.EOF, ""}})
	assertTokens(t, `"line1\nline2\tend"`,
		[]tok{{token.STRING, "line1\nline2\tend"}, {token.EOF, ""}})
	assertTokens(t, `"quote:\" back:\\\\"`,
		[]tok{{token.STRING, `quote:" back:\\`}, {token.EOF, ""}})
	// Unknown escapes pass through unchanged.
	assertTokens(t, `"a\qb"`, []tok{{token.STRING, `a\qb`}, {token.EOF, ""}})
	// Unterminated strings get UNTERM; the parser turns it into an error.
	assertTokens(t, `"oops`, []tok{{token.UNTERM, "oops"}, {token.EOF, ""}})
}

func TestNewlinesAreSignificant(t *testing.T) {
	assertTokens(t, "print \"a\"\nprint \"b\"",
		[]tok{
			{token.PRINT, "print"}, {token.STRING, "a"}, {token.NEWLINE, "\n"},
			{token.PRINT, "print"}, {token.STRING, "b"}, {token.EOF, ""},
		})
}

func TestComments(t *testing.T) {
	assertTokens(t, "# whole line\nprint 1",
		[]tok{
			{token.NEWLINE, "\n"}, {token.PRINT, "print"}, {token.INT, "1"},
			{token.EOF, ""},
		})
	assertTokens(t, "print 1 # trailing",
		[]tok{{token.PRINT, "print"}, {token.INT, "1"}, {token.EOF, ""}})
	assertTokens(t, "# just a comment",
		[]tok{{token.EOF, ""}})
}

func TestIdentifierShapes(t *testing.T) {
	// Dotted names and paths fold into one identifier (MODULES.md open item).
	assertTokens(t, "user.name", []tok{{token.IDENT, "user.name"}, {token.EOF, ""}})
	assertTokens(t, "src/main", []tok{{token.IDENT, "src/main"}, {token.EOF, ""}})
	assertTokens(t, "_hidden1", []tok{{token.IDENT, "_hidden1"}, {token.EOF, ""}})
}

func TestUnknownCharactersPassThrough(t *testing.T) {
	// Unknowns become single-char tokens so the parser can report them
	// with position instead of the lexer guessing.
	assertTokens(t, "a ; b",
		[]tok{
			{token.IDENT, "a"}, {token.Type(";"), ";"}, {token.IDENT, "b"},
			{token.EOF, ""},
		})
	assertTokens(t, "!", []tok{{token.Type("!"), "!"}, {token.EOF, ""}})
}

func TestEmptyInput(t *testing.T) {
	assertTokens(t, "", []tok{{token.EOF, ""}})
}

func TestPositions(t *testing.T) {
	toks := scan("let a = 1\nprint a\n  b")

	type posWant struct {
		typ       token.Type
		literal   string
		line, col int
	}
	want := []posWant{
		{token.LET, "let", 1, 1},
		{token.IDENT, "a", 1, 5},
		{token.ASSIGN, "=", 1, 7},
		{token.INT, "1", 1, 9},
		{token.NEWLINE, "\n", 1, 10},
		{token.PRINT, "print", 2, 1},
		{token.IDENT, "a", 2, 7},
		{token.NEWLINE, "\n", 2, 8},
		{token.IDENT, "b", 3, 3},
		{token.EOF, "", 3, 4},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Type != w.typ || toks[i].Literal != w.literal ||
			toks[i].Line != w.line || toks[i].Column != w.col {
			t.Errorf("token %d: got %v, want %s %q at %d:%d",
				i, toks[i], w.typ, w.literal, w.line, w.col)
		}
	}
}

func TestBenchmarkShapedInput(t *testing.T) {
	// A realistic snippet exercising statement separation end to end.
	input := "let x = 5\nlet y = x * 2\nprint y\n"
	toks := scan(input)
	var newlines, lets int
	for _, tk := range toks {
		switch tk.Type {
		case token.NEWLINE:
			newlines++
		case token.LET:
			lets++
		}
	}
	if newlines != 3 || lets != 2 {
		t.Fatalf("got %d newlines / %d lets, want 3/2\n%v", newlines, lets, toks)
	}
}

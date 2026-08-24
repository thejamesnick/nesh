// Package lexer converts Nesh source text into tokens.
//
// Contract (MODULES.md): a pure function of input → tokens. No I/O, no config,
// no globals. Newlines are significant: they terminate statements.
package lexer

import (
	"strings"

	"nesh/internal/token"
)

// Lexer scans Nesh source code and produces a token stream.
type Lexer struct {
	input string
	// position is the byte offset of the last consumed character;
	// readPosition is the byte offset of the next character to consume
	// (peek() returns input[readPosition]).
	position     int
	readPosition int
	line         int
	column       int
}

// New returns a Lexer ready to scan input. No character is consumed:
// the first NextToken call sees input[0].
func New(input string) *Lexer {
	return &Lexer{input: input, position: -1, line: 1, column: 0}
}

// NextToken returns the next token in the input, ending with token.EOF.
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	tok := token.Token{Line: l.line, Column: l.column + 1}

	switch ch := l.peek(); {
	case ch == 0:
		tok.Type = token.EOF
		tok.Literal = ""
	case ch == '"':
		tok.Type, tok.Literal = l.readString()
	case ch == '\n':
		l.readChar()
		tok.Type = token.NEWLINE
		tok.Literal = "\n"
	case isDigit(ch):
		tok.Type, tok.Literal = l.readNumber()
	case isIdentStart(ch):
		lit := l.readIdentifier()
		tok.Type = token.LookupIdent(lit)
		tok.Literal = lit
	default:
		tok.Type, tok.Literal = l.readOperator()
	}
	return tok
}

func (l *Lexer) peek() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func peekAt(input string, i int) byte {
	if i >= len(input) {
		return 0
	}
	return input[i]
}

// readChar advances one byte, tracking line and column.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.position = l.readPosition
		return
	}
	ch := l.input[l.readPosition]
	l.position = l.readPosition
	l.readPosition++
	if ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// skipWhitespaceAndComments consumes spaces, tabs, carriage returns,
// and # comments (to end of line; the newline itself stays a token).
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.readChar()
			continue
		}
		if ch == '#' {
			for l.peek() != '\n' && l.peek() != 0 {
				l.readChar()
			}
			continue
		}
		return
	}
}

func (l *Lexer) readString() (token.Type, string) {
	l.readChar() // opening quote
	var b strings.Builder
	for {
		ch := l.peek()
		if ch == 0 {
			return token.UNTERM, b.String() // parser reports position
		}
		if ch == '"' {
			l.readChar()
			return token.STRING, b.String()
		}
		if ch == '\\' {
			l.readChar()
			switch esc := l.peek(); esc {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte('\\')
				b.WriteByte(esc)
			}
			l.readChar()
			continue
		}
		b.WriteByte(ch)
		l.readChar()
	}
}

func (l *Lexer) readNumber() (token.Type, string) {
	var b strings.Builder
	for isDigit(l.peek()) {
		b.WriteByte(l.peek())
		l.readChar()
	}
	if l.peek() == '.' && isDigit(peekAt(l.input, l.readPosition+1)) {
		b.WriteByte('.')
		l.readChar()
		for isDigit(l.peek()) {
			b.WriteByte(l.peek())
			l.readChar()
		}
		return token.FLOAT, b.String()
	}
	return token.INT, b.String()
}

func (l *Lexer) readIdentifier() string {
	var b strings.Builder
	for isIdentPart(l.peek()) {
		b.WriteByte(l.peek())
		l.readChar()
	}
	return b.String()
}

// readOperator consumes a one- or two-character operator or delimiter.
// Unknown characters pass through as a token of their literal so the
// parser can report a clean syntax error with position.
func (l *Lexer) readOperator() (token.Type, string) {
	ch := l.peek()
	l.readChar()
	next := l.peek()
	switch ch {
	case '=':
		if next == '=' {
			l.readChar()
			return token.EQ, "=="
		}
		return token.ASSIGN, "="
	case '!':
		if next == '=' {
			l.readChar()
			return token.NOT_EQ, "!="
		}
		return token.Type("!"), "!"
	case '<':
		if next == '=' {
			l.readChar()
			return token.LTE, "<="
		}
		return token.LT, "<"
	case '>':
		if next == '>' {
			l.readChar()
			return token.APPEND, ">>"
		}
		if next == '=' {
			l.readChar()
			return token.GTE, ">="
		}
		return token.GT, ">"
	case '+':
		return token.PLUS, "+"
	case '-':
		return token.MINUS, "-"
	case '*':
		return token.ASTERISK, "*"
	case '/':
		return token.SLASH, "/"
	case '(':
		return token.LPAREN, "("
	case ')':
		return token.RPAREN, ")"
	case ',':
		return token.COMMA, ","
	default:
		return token.Type(string(rune(ch))), string(rune(ch))
	}
}

func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isIdentPart also folds '.', '-', and '/' into identifiers so paths and
// dotted names lex as one token (src/main, user.name). Open item in MODULES.md.
func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '.' || ch == '-' || ch == '/'
}

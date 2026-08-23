// Package token defines the lexical tokens of the Nesh language.
package token

import "fmt"

// Type identifies the kind of a token.
type Type string

const (
	// Literals & identifiers
	IDENT  Type = "IDENT"  // variable or command name
	INT    Type = "INT"    // 42
	FLOAT  Type = "FLOAT"  // 3.14
	STRING Type = "STRING" // "hello"

	// Operators
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	EQ       Type = "=="
	NOT_EQ   Type = "!="
	LT       Type = "<"
	GT       Type = ">"
	LTE      Type = "<="
	GTE      Type = ">="

	// UNTERM marks a string literal that never closed before EOF.
	UNTERM Type = "UNTERM"

	// Delimiters
	LPAREN  Type = "("
	RPAREN  Type = ")"
	COMMA   Type = ","
	NEWLINE Type = "NEWLINE"
	EOF     Type = "EOF"

	// Keywords
	LET    Type = "let"
	PRINT  Type = "print"
	IF     Type = "if"
	THEN   Type = "then"
	ELSE   Type = "else"
	END    Type = "end"
	FN     Type = "fn"
	RETURN Type = "return"
	FOR    Type = "for"
	IN     Type = "in"
	WHILE  Type = "while"
	AND    Type = "and"
	OR     Type = "or"
	NOT    Type = "not"
	TRUE   Type = "true"
	FALSE  Type = "false"
	ELIF   Type = "elif"
	RUN    Type = "run"
	IMPORT Type = "import"
	AS     Type = "as"
)

var keywords = map[string]Type{
	"let":    LET,
	"print":  PRINT,
	"if":     IF,
	"then":   THEN,
	"else":   ELSE,
	"elif":   ELIF,
	"run":    RUN,
	"import": IMPORT,
	"as":     AS,
	"end":    END,
	"fn":     FN,
	"return": RETURN,
	"for":    FOR,
	"in":     IN,
	"while":  WHILE,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"true":   TRUE,
	"false":  FALSE,
}

// LookupIdent returns the keyword type for ident, or IDENT if it is not a keyword.
func LookupIdent(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return IDENT
}

// Token is a single lexical unit with its source position.
type Token struct {
	Type    Type
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("%d:%d %s %q", t.Line, t.Column, t.Type, t.Literal)
}

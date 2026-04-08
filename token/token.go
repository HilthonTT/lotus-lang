package token

import "fmt"

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

func (t Token) String() string {
	return fmt.Sprintf("Type: %s, Literal: %s, Line: %d, Col: %d", t.Type, t.Literal, t.Line, t.Col)
}

const (
	// Special
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers + Literals
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	STRING TokenType = "STRING"

	// Operators
	ASSIGN     TokenType = "="
	PLUS       TokenType = "+"
	PLUSPLUS   TokenType = "++"
	MINUS      TokenType = "-"
	MINUSMINUS TokenType = "--"
	BANG       TokenType = "!"
	ASTERISK   TokenType = "*"
	SLASH      TokenType = "/"
	MODULO     TokenType = "%"

	LT    TokenType = "<"
	GT    TokenType = ">"
	EQ    TokenType = "=="
	NOTEQ TokenType = "!="
	LTEQ  TokenType = "<="
	GTEQ  TokenType = ">="

	AND TokenType = "&&"
	OR  TokenType = "||"

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	DOT       TokenType = "."

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"

	// Keywords
	FN       TokenType = "FN"
	LET      TokenType = "LET"
	MUT      TokenType = "MUT"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	WHILE    TokenType = "WHILE"
	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	RETURN   TokenType = "RETURN"
	NIL      TokenType = "NIL"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"

	// OOP keywords
	CLASS   TokenType = "CLASS"
	EXTENDS TokenType = "EXTENDS"
	SELF    TokenType = "SELF"
	SUPER   TokenType = "SUPER"

	QUESTION TokenType = "?"
)

var keywords = map[string]TokenType{
	"fn":       FN,
	"let":      LET,
	"mut":      MUT,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"in":       IN,
	"return":   RETURN,
	"nil":      NIL,
	"break":    BREAK,
	"continue": CONTINUE,
	"class":    CLASS,
	"extends":  EXTENDS,
	"self":     SELF,
	"super":    SUPER,
}

func LookupIdentifier(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}

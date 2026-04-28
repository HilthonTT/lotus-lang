package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/hilthontt/lotus/token"
)

type Lexer struct {
	position     int
	readPosition int
	ch           rune
	characters   []rune
	prevToken    token.Token
	line         int
	col          int
	Comments     []CommentToken
}

type CommentToken struct {
	Line int
	Text string
}

func New(input string) *Lexer {
	l := &Lexer{
		characters: []rune(input),
		line:       1,
		col:        0,
	}
	l.readChar()
	return l
}

func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		t := l.NextToken()
		tokens = append(tokens, t)
		if t.Type == token.EOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.col++
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(string(l.characters[l.readPosition:]))
	return r
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	tok := token.Token{Line: l.line, Col: l.col}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok.Type = token.EQ
			tok.Literal = "=="
			l.readChar()
		} else {
			tok.Type = token.ASSIGN
			tok.Literal = "="
		}

	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok.Type = token.PLUSPLUS
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.PLUS_ASSIGN
			tok.Literal = "+="
		} else {
			tok.Type = token.PLUS
			tok.Literal = "+"
		}

	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok.Type = token.MINUSMINUS
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '>' {
			l.readChar()
			tok.Type = token.ARROW
			tok.Literal = "->"
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.MINUS_ASSIGN
			tok.Literal = "-="
		} else {
			tok.Type = token.MINUS
			tok.Literal = "-"
		}
	case '!':
		if l.peekChar() == '=' {
			tok.Type = token.NOTEQ
			tok.Literal = "!="
			l.readChar()
		} else {
			tok.Type = token.BANG
			tok.Literal = "!"
		}

	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.MUL_ASSIGN
			tok.Literal = "*="
		} else {
			tok.Type = token.ASTERISK
			tok.Literal = "*"
		}

	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.DIV_ASSIGN
			tok.Literal = "/="
		} else {
			tok.Type = token.SLASH
			tok.Literal = "/"
		}

	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.MOD_ASSIGN
			tok.Literal = "%="
		} else {
			tok.Type = token.MODULO
			tok.Literal = "%"
		}

	case '<':
		if l.peekChar() == '<' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok.Type = token.LSHIFT_ASSIGN
				tok.Literal = "<<="
			} else {
				tok.Type = token.LSHIFT
				tok.Literal = "<<"
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.LTEQ
			tok.Literal = "<="
		} else {
			tok.Type = token.LT
			tok.Literal = "<"
		}

	case '>':
		if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok.Type = token.RSHIFT_ASSIGN
				tok.Literal = ">>="
			} else {
				tok.Type = token.RSHIFT
				tok.Literal = ">>"
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.GTEQ
			tok.Literal = ">="
		} else {
			tok.Type = token.GT
			tok.Literal = ">"
		}

	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok.Type = token.AND
			tok.Literal = "&&"
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.AND_ASSIGN
			tok.Literal = "&="
		} else {
			tok.Type = token.BITAND
			tok.Literal = "&"
		}

	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok.Type = token.OR
			tok.Literal = "||"
		} else if l.peekChar() == '>' {
			l.readChar()
			tok.Type = token.PIPE
			tok.Literal = "|>"
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.OR_ASSIGN
			tok.Literal = "|="
		} else {
			tok.Type = token.BITOR
			tok.Literal = "|"
		}

	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.XOR_ASSIGN
			tok.Literal = "^="
		} else {
			tok.Type = token.BITXOR
			tok.Literal = "^"
		}

	case '~':
		tok.Type = token.TILDE
		tok.Literal = "~"

	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			tok.Type = token.NULLCOALESCE
			tok.Literal = "??"
		} else if l.peekChar() == '.' {
			l.readChar()
			tok.Type = token.OPTDOT
			tok.Literal = "?."
		} else {
			tok.Type = token.QUESTION
			tok.Literal = "?"
		}

	case ',':
		tok.Type = token.COMMA
		tok.Literal = ","

	case ';':
		tok.Type = token.SEMICOLON
		tok.Literal = ";"

	case ':':
		tok.Type = token.COLON
		tok.Literal = ":"

	case '.':
		if l.peekChar() == '.' && l.characters[l.readPosition+1] == '.' {
			l.readChar()
			l.readChar()
			tok.Type = token.ELLIPSIS
			tok.Literal = "..."
		} else {
			tok.Type = token.DOT
			tok.Literal = "."
		}

	case '(':
		tok.Type = token.LPAREN
		tok.Literal = "("

	case ')':
		tok.Type = token.RPAREN
		tok.Literal = ")"

	case '{':
		tok.Type = token.LBRACE
		tok.Literal = "{"

	case '}':
		tok.Type = token.RBRACE
		tok.Literal = "}"

	case '[':
		tok.Type = token.LBRACKET
		tok.Literal = "["

	case ']':
		tok.Type = token.RBRACKET
		tok.Literal = "]"

	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()

	case 0:
		tok.Type = token.EOF
		tok.Literal = ""

	default:
		if isDigit(l.ch) {
			return l.readNumber()
		}
		if isLetter(l.ch) {
			lit := l.readIdentifier()
			tok.Type = token.LookupIdentifier(lit)
			tok.Literal = lit
			return tok
		}
		tok.Type = token.ILLEGAL
		tok.Literal = string(l.ch)
	}

	l.readChar()
	l.prevToken = tok
	return tok
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		l.skipWhitespace()
		if l.ch == '/' && l.peekChar() == '/' {
			commentLine := l.line
			var buf []rune

			for l.ch != '\n' && l.ch != 0 {
				buf = append(buf, l.ch)
				l.readChar()
			}

			l.Comments = append(l.Comments, CommentToken{
				Line: commentLine,
				Text: string(buf),
			})

			continue
		}
		break
	}
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return string(l.characters[start:l.position])
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening "
	var buf []byte

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			case '\\':
				buf = append(buf, '\\')
			case '"':
				buf = append(buf, '"')
			default:
				buf = append(buf, '\\')
				buf = append(buf, byte(l.ch))
			}
		} else {
			buf = append(buf, byte(l.ch))
		}
		l.readChar()
	}

	return string(buf)
}

func (l *Lexer) readNumber() token.Token {
	tok := token.Token{Line: l.line, Col: l.col}
	start := l.position
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = string(l.characters[start:l.position])
	if isFloat {
		tok.Type = token.FLOAT
	} else {
		tok.Type = token.INT
	}
	return tok
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func (l *Lexer) Errorf(format string, args ...any) string {
	return fmt.Sprintf("line %d, col %d: %s", l.line, l.col, fmt.Sprintf(format, args...))
}

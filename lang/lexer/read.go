package lexer

import (
	"strings"
	"unicode/utf8"

	"github.com/hilthontt/lotus/token"
)

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
	return l.characters[l.readPosition]
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
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			buf = utf8.AppendRune(buf, l.ch)
		}
		l.readChar()
	}

	return string(buf)
}

func (l *Lexer) readTripleString() string {
	// consume the two extra opening quotes (we already consumed the first)
	l.readChar() // second "
	l.readChar() // third "
	l.readChar() // move past opening """

	var buf []rune

	for {
		// Check for closing """
		if l.ch == '"' && l.peekChar() == '"' &&
			l.readPosition < len(l.characters) &&
			l.characters[l.readPosition] == '"' {
			l.readChar() // second "
			l.readChar() // third "
			l.readChar() // move past closing """
			break
		}

		if l.ch == 0 {
			break // EOF without closing - lexer will emit ILLEGAL later
		}
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		buf = append(buf, l.ch)
		l.readChar()
	}

	raw := string(buf)

	// Strip a single leading newline (right after opening """)
	raw = strings.TrimPrefix(raw, "\n")

	// Dedent: find minimum indentation across all non-empty lines
	lines := strings.Split(raw, "\n")
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		spaces := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || spaces < minIndent {
			minIndent = spaces
		}
	}
	if minIndent > 0 {
		for i, line := range lines {
			if len(line) >= minIndent {
				lines[i] = line[minIndent:]
			}
		}
	}

	// Trim trailing whitespace-only lines
	result := strings.Join(lines, "\n")
	result = strings.TrimRight(result, " \t\n")

	return result
}

func (l *Lexer) readNumber() token.Token {
	if l.ch == 0 {
		switch l.peekChar() {
		case 'x', 'X': // hex
		case 'b', 'B': // binary
		case 'o', 'O': // octal
		}
	}

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

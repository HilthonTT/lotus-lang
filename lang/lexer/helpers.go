package lexer

import (
	"fmt"
	"unicode"
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func (l *Lexer) Errorf(format string, args ...any) error {
	return fmt.Errorf("line %d, col %d: %w", l.line, l.col, fmt.Errorf(format, args...))
}

package lexer

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

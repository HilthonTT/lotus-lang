package lexer

import (
	"fmt"
	"testing"

	"github.com/hilthontt/lotus/token"
)

func TestNextToken(t *testing.T) {
	input := `
let name = "Lotus"
mut counter = 0
print("Language:", name)
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
	}{
		{token.LET, "let", 2},
		{token.IDENT, "name", 2},
		{token.ASSIGN, "=", 2},
		{token.STRING, "Lotus", 2},
		{token.MUT, "mut", 3},
		{token.IDENT, "counter", 3},
		{token.ASSIGN, "=", 3},
		{token.INT, "0", 3},
		{token.IDENT, "print", 4},
		{token.LPAREN, "(", 4},
		{token.STRING, "Language:", 4},
		{token.COMMA, ",", 4},
		{token.IDENT, "name", 4},
		{token.RPAREN, ")", 4},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		// Debug Output
		fmt.Printf("tests[%2d] Line:%2d  %-10s  %q\n",
			i, tok.Line, tok.Type, tok.Literal)

		// Check token type
		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - TOKEN TYPE wrong\n   Expected: %q\n   Got:      %q\n   Literal:  %q\n",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		// Check literal
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - LITERAL wrong\n   Expected: %q\n   Got:      %q\n   Type:     %q\n",
				i, tt.expectedLiteral, tok.Literal, tok.Type)
		}

		// Check line number
		if tok.Line != tt.expectedLine {
			t.Errorf("tests[%d] - LINE wrong\n   Expected: %d\n   Got:      %d\n   Token:    %s %q\n",
				i, tt.expectedLine, tok.Line, tok.Type, tok.Literal)
		}
	}

	fmt.Println("\n=== Test completed. Check above for any errors ===")
}

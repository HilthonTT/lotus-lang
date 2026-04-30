package parser

import "github.com/hilthontt/lotus/token"

// Operator precedences
const (
	_ int = iota
	LOWEST
	PIPE_PREC    // |>
	NULLCOALESCE // ??
	TERNARY      // ?:
	OR_PREC      // ||
	AND_PREC     // &&
	BITWISE_OR   // |
	BITWISE_XOR  // ^
	BITWISE_AND  // &
	EQUALS       // == !=
	LESSGREATER  // < > <= >=
	SHIFT        // << >>
	SUM          // + -
	PRODUCT      // * / %
	PREFIX       // -x !x ~x
	CALL         // fn(x)
	INDEX        // arr[i] obj.field
)

// each token precedence
var precedences = map[token.TokenType]int{
	token.NULLCOALESCE: NULLCOALESCE,
	token.OR:           OR_PREC,
	token.AND:          AND_PREC,
	token.BITOR:        BITWISE_OR,
	token.BITXOR:       BITWISE_XOR,
	token.BITAND:       BITWISE_AND,
	token.EQ:           EQUALS,
	token.NOTEQ:        EQUALS,
	token.IN:           EQUALS,
	token.LT:           LESSGREATER,
	token.GT:           LESSGREATER,
	token.LTEQ:         LESSGREATER,
	token.GTEQ:         LESSGREATER,
	token.LSHIFT:       SHIFT,
	token.RSHIFT:       SHIFT,
	token.PLUS:         SUM,
	token.MINUS:        SUM,
	token.ASTERISK:     PRODUCT,
	token.SLASH:        PRODUCT,
	token.MODULO:       PRODUCT,
	token.LPAREN:       CALL,
	token.LBRACKET:     INDEX,
	token.DOT:          INDEX, // field access binds as tightly as indexing
	token.OPTDOT:       INDEX,
	token.QUESTION:     TERNARY,
	token.PIPE:         PIPE_PREC,
}

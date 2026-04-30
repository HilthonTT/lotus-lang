package parser

import "github.com/hilthontt/lotus/token"

// parseExpressionOrAssignStatement — add compound assignment detection.
// Add this BEFORE the existing p.peekTokenIs(token.ASSIGN) check:
//
// Compound assignment operators:
var compoundOps = map[token.TokenType]string{
	token.PLUS_ASSIGN:   "+=",
	token.MINUS_ASSIGN:  "-=",
	token.MUL_ASSIGN:    "*=",
	token.DIV_ASSIGN:    "/=",
	token.MOD_ASSIGN:    "%=",
	token.AND_ASSIGN:    "&=",
	token.OR_ASSIGN:     "|=",
	token.XOR_ASSIGN:    "^=",
	token.LSHIFT_ASSIGN: "<<=",
	token.RSHIFT_ASSIGN: ">>=",
}

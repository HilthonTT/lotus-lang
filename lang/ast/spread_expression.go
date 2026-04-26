package ast

import "github.com/hilthontt/lotus/token"

// ...expr  — used in call arguments and array literal
type SpreadExpression struct {
	Token token.Token // the '...' token
	Value Expression
}

func (se *SpreadExpression) expressionNode() {}

func (se *SpreadExpression) TokenLiteral() string {
	return se.Token.Literal
}

func (se *SpreadExpression) String() string {
	return "..." + se.Value.String()
}

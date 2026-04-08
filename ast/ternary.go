package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// TernaryExpression: condition ? consequence : alternative
type TernaryExpression struct {
	Token       token.Token // the '?' token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode() {}

func (te *TernaryExpression) TokenLiteral() string {
	return te.Token.Literal
}

func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(te.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(te.Consequence.String())
	out.WriteString(" : ")
	out.WriteString(te.Alternative.String())
	out.WriteString(")")
	return out.String()
}

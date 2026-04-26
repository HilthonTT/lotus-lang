package ast

import "github.com/hilthontt/lotus/token"

// fn sum(...nums)  — the ...nums parameter
type RestParameter struct {
	Token token.Token // the '...' token
	Name  *Identifier
}

func (rp *RestParameter) expressionNode() {}

func (rp *RestParameter) TokenLiteral() string {
	return rp.Token.Literal
}

func (rp *RestParameter) String() string {
	return "..." + rp.Name.String()
}

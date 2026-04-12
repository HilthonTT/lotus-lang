package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// OptionalFieldAccess: obj?.field — returns nil if obj is nil
type OptionalFieldAccess struct {
	Token token.Token // the '?.' token
	Left  Expression
	Field *Identifier
}

func (ofa *OptionalFieldAccess) expressionNode() {}

func (ofa *OptionalFieldAccess) TokenLiteral() string {
	return ofa.Token.Literal
}

func (ofa *OptionalFieldAccess) String() string {
	var out bytes.Buffer
	out.WriteString(ofa.Left.String())
	out.WriteString("?.")
	out.WriteString(ofa.Field.String())
	return out.String()
}

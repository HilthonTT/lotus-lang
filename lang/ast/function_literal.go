package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

type FunctionLiteral struct {
	Token      token.Token
	Name       string // optional, empty for anonymous
	Parameters []*Identifier
	Body       *BlockStatement
	ParamTypes []*TypeAnnotation // parallel to Parameters, may be nil entries
	ReturnType *TypeAnnotation   // optional: -> int
	TypeParams []TypeParam
	IsVariadic bool
}

func (e *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *FunctionLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("fn")
	if e.Name != "" {
		out.WriteString(" ")
		out.WriteString(e.Name)
	}
	out.WriteString("(")
	for i, p := range e.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") { ")
	out.WriteString(e.Body.String())
	out.WriteString(" }")
	return out.String()
}

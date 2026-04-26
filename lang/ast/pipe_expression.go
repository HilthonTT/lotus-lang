package ast

import "github.com/hilthontt/lotus/token"

// value |> fn(args)
// Desugars to: fn(value, args)
type PipeExpression struct {
	Token token.Token // the '|>' token
	Left  Expression  // the value being piped
	Right Expression  // the function call (without the first arg)
}

func (pe *PipeExpression) expressionNode() {}

func (pe *PipeExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PipeExpression) String() string {
	return pe.Left.String() + " |> " + pe.Right.String()
}

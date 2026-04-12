package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// MatchExpression: match x { pattern -> expr ... }
type MatchExpression struct {
	Token   token.Token
	Subject Expression
	Arms    []*MatchArm
}

type MatchArm struct {
	Pattern Expression // nil means wildcard (_)
	Body    Expression
	IsWild  bool
}

func (me *MatchExpression) expressionNode() {}

func (me *MatchExpression) TokenLiteral() string {
	return me.Token.Literal
}

func (me *MatchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("match ")
	out.WriteString(me.Subject.String())
	out.WriteString(" { ")
	for _, arm := range me.Arms {
		if arm.IsWild {
			out.WriteString("_ -> ")
		} else {
			out.WriteString(arm.Pattern.String())
			out.WriteString(" -> ")
		}
		out.WriteString(arm.Body.String())
		out.WriteString(" ")
	}
	out.WriteString("}")
	return out.String()
}

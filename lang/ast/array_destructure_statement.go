package ast

import (
	"strings"

	"github.com/hilthontt/lotus/token"
)

// let [a, b, ...rest] = arr
type ArrayDestructureStatement struct {
	Token   token.Token // let/mut token
	Mutable bool
	Names   []Expression // Identifier or SpreadExpression for rest
	Value   Expression
}

func (ad *ArrayDestructureStatement) statementNode() {}

func (ad *ArrayDestructureStatement) TokenLiteral() string {
	return ad.Token.Literal
}

func (ad *ArrayDestructureStatement) String() string {
	names := make([]string, len(ad.Names))
	for i, n := range ad.Names {
		names[i] = n.String()
	}
	return "let [" + strings.Join(names, ", ") + "] = " + ad.Value.String()
}

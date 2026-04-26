package ast

import (
	"strings"

	"github.com/hilthontt/lotus/token"
)

type MapDestructureStatement struct {
	Token   token.Token
	Mutable bool
	Keys    []*Identifier // keys to extract
	Value   Expression
}

func (md *MapDestructureStatement) statementNode() {}

func (md *MapDestructureStatement) TokenLiteral() string {
	return md.Token.Literal
}

func (md *MapDestructureStatement) String() string {
	keys := make([]string, len(md.Keys))
	for i, k := range md.Keys {
		keys[i] = k.Value
	}
	return "let { " + strings.Join(keys, ", ") + " } = " + md.Value.String()
}

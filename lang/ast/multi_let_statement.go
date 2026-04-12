package ast

import (
	"bytes"
	"strings"

	"github.com/hilthontt/lotus/token"
)

// MultiLetStatement: let a, b = 1, 2  /  mut x, y = 10, 20
type MultiLetStatement struct {
	Token   token.Token // let or mut token
	Mutable bool
	Names   []*Identifier
	Values  []Expression
}

func (mls *MultiLetStatement) statementNode() {}

func (mls *MultiLetStatement) TokenLiteral() string {
	return mls.Token.Literal
}

func (mls *MultiLetStatement) String() string {
	var out bytes.Buffer
	if mls.Mutable {
		out.WriteString("mut ")
	} else {
		out.WriteString("let ")
	}
	names := make([]string, len(mls.Names))
	for i, n := range mls.Names {
		names[i] = n.Value
	}
	out.WriteString(strings.Join(names, ", "))
	out.WriteString(" = ")
	vals := make([]string, len(mls.Values))
	for i, v := range mls.Values {
		vals[i] = v.String()
	}
	out.WriteString(strings.Join(vals, ", "))
	return out.String()
}

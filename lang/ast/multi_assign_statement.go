package ast

import (
	"bytes"
	"strings"

	"github.com/hilthontt/lotus/token"
)

// MultiAssignStatement: a, b = b, a
type MultiAssignStatement struct {
	Token  token.Token
	Names  []Expression // can be Identifier or IndexExpression or FieldAccess
	Values []Expression
}

func (mas *MultiAssignStatement) statementNode() {}

func (mas *MultiAssignStatement) TokenLiteral() string {
	return mas.Token.Literal
}

func (mas *MultiAssignStatement) String() string {
	var out bytes.Buffer
	names := make([]string, len(mas.Names))
	for i, n := range mas.Names {
		names[i] = n.String()
	}
	out.WriteString(strings.Join(names, ", "))
	out.WriteString(" = ")
	vals := make([]string, len(mas.Values))
	for i, v := range mas.Values {
		vals[i] = v.String()
	}
	out.WriteString(strings.Join(vals, ", "))
	return out.String()
}

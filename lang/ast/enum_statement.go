package ast

import (
	"bytes"
	"strings"

	"github.com/hilthontt/lotus/token"
)

// EnumStatement: enum Color { Red, Green, Blue } or enum Shape { Circle(r), Rect(w, h) }
type EnumStatement struct {
	Token    token.Token // the 'enum' token
	Name     *Identifier
	Variants []*EnumVariantDef
}

type EnumVariantDef struct {
	Name   string
	Fields []string // empty for simple variants
}

func (es *EnumStatement) statementNode() {}

func (es *EnumStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *EnumStatement) String() string {
	var out bytes.Buffer
	out.WriteString("enum ")
	out.WriteString(es.Name.String())
	out.WriteString(" { ")
	for i, v := range es.Variants {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(v.Name)
		if len(v.Fields) > 0 {
			out.WriteString("(")
			out.WriteString(strings.Join(v.Fields, ", "))
			out.WriteString(")")
		}
	}
	out.WriteString(" }")
	return out.String()
}

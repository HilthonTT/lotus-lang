package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// LetStatement holds a let-statement
type LetStatement struct {
	// Token holds the token.
	Token token.Token

	// Name is the name of the variable to which we're assigning
	Name *Identifier

	// Mutable holds the value if the variable is mutable or not.
	Mutable bool

	// Value is the thing we're storing in the variable.
	Value Expression

	// Optional type annotation
	TypeAnnot *TypeAnnotation
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String returns this object as a string.
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

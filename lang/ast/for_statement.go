package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

type ForStatement struct {
	Token    token.Token
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (s *ForStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *ForStatement) TokenLiteral() string {
	return s.Token.Literal
}

func (s *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(s.Variable.String())
	out.WriteString(" in ")
	out.WriteString(s.Iterable.String())
	out.WriteString(" { ")
	out.WriteString(s.Body.String())
	out.WriteString(" }")
	return out.String()
}

type ForIndexStatement struct {
	Token    token.Token
	Index    *Identifier // the index variable
	Variable *Identifier // the element variable
	Iterable Expression
	Body     *BlockStatement
}

func (fi *ForIndexStatement) statementNode() {}

func (fi *ForIndexStatement) TokenLiteral() string {
	return fi.Token.Literal
}

func (fi *ForIndexStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fi.Index.Value + ", " + fi.Variable.Value)
	out.WriteString(" in ")
	out.WriteString(fi.Iterable.String())
	out.WriteString(" { ")
	out.WriteString(fi.Body.String())
	out.WriteString(" }")
	return out.String()
}

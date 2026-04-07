package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// ClassStatement: class Foo extends Bar { fn init(self) { ... } }
type ClassStatement struct {
	Token      token.Token // the 'class' token
	Name       *Identifier
	SuperClass *Identifier // nil if no superclass
	Methods    []*FunctionLiteral
}

func (cs *ClassStatement) statementNode() {}

func (cs *ClassStatement) TokenLiteral() string {
	return cs.Token.Literal
}

func (cs *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString("class ")
	out.WriteString(cs.Name.String())
	if cs.SuperClass != nil {
		out.WriteString(" extends ")
		out.WriteString(cs.SuperClass.String())
	}
	out.WriteString(" { ")
	for _, m := range cs.Methods {
		out.WriteString(m.String())
		out.WriteString(" ")
	}
	out.WriteString("}")
	return out.String()
}

// FieldAccessExpression: obj.field
type FieldAccessExpression struct {
	Token token.Token // the '.' token
	Left  Expression
	Field *Identifier
}

func (fa *FieldAccessExpression) expressionNode() {}

func (fa *FieldAccessExpression) TokenLiteral() string {
	return fa.Token.Literal
}

func (fa *FieldAccessExpression) String() string {
	var out bytes.Buffer
	out.WriteString(fa.Left.String())
	out.WriteString(".")
	out.WriteString(fa.Field.String())
	return out.String()
}

// FieldAssignStatement: obj.field = value
type FieldAssignStatement struct {
	Token token.Token
	Left  Expression
	Field *Identifier
	Value Expression
}

func (fa *FieldAssignStatement) statementNode() {}

func (fa *FieldAssignStatement) TokenLiteral() string {
	return fa.Token.Literal
}

func (fa *FieldAssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fa.Left.String())
	out.WriteString(".")
	out.WriteString(fa.Field.String())
	out.WriteString(" = ")
	out.WriteString(fa.Value.String())
	out.WriteString(";")
	return out.String()
}

// SelfExpression: the 'self' keyword inside a method.
type SelfExpression struct {
	Token token.Token
}

func (se *SelfExpression) expressionNode() {}

func (se *SelfExpression) TokenLiteral() string {
	return se.Token.Literal
}

func (se *SelfExpression) String() string {
	return "self"
}

// SuperExpression: the 'super' keyword inside a method.
type SuperExpression struct {
	Token token.Token
}

func (se *SuperExpression) expressionNode() {}

func (se *SuperExpression) TokenLiteral() string {
	return se.Token.Literal
}

func (se *SuperExpression) String() string {
	return "super"
}

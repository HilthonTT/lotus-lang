package ast

import (
	"bytes"

	"github.com/hilthontt/lotus/token"
)

// Node represents a node.
type Node interface {
	// TokenLiteral returns the literal of the token.
	TokenLiteral() string

	// String returns this object as a string.
	String() string
}

// Statement represents a single statement.
type Statement interface {
	// Node is the node holding the actual statement.
	Node
	statementNode()
}

// Expression represents a single expression.
type Expression interface {
	// Node is the node holding the expression.
	Node
	expressionNode()
}

// Program represents a complete program.
type Program struct {
	// Statements is the set of statements which the program is comprised
	// of.
	Statements []Statement
}

// TokenLiteral returns the literal token of our program.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String returns this object as a string.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
	}
	return out.String()
}

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

// Identifier holds a single identifier.
type Identifier struct {
	// Token is the literal token
	Token token.Token

	// Value is the name of the identifier
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the literal token.
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// String returns this object as a string.
func (i *Identifier) String() string {
	return i.Value
}

// ReturnStatement stores a return-statement
type ReturnStatement struct {
	// Token contains the literal token.
	Token token.Token

	// ReturnValue is the value which is to be returned
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

// String returns this object as a string.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.TokenLiteral())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement is an expression.
type ExpressionStatement struct {
	// Token is the literal token.
	Token token.Token

	// Expression holds the expression.
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

// String returns this object as a string.
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (s *WhileStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *WhileStatement) TokenLiteral() string {
	return s.Token.Literal
}

func (s *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(s.Condition.String())
	out.WriteString(" { ")
	out.WriteString(s.Body.String())
	out.WriteString(" }")
	return out.String()
}

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

type BreakStatement struct {
	Token token.Token
}

func (s *BreakStatement) String() string {
	var out bytes.Buffer
	out.WriteString(s.TokenLiteral())
	out.WriteString(";")
	return out.String()
}

func (s *BreakStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *BreakStatement) TokenLiteral() string {
	return s.Token.Literal
}

type ContinueStatement struct {
	Token token.Token
}

func (s *ContinueStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *ContinueStatement) TokenLiteral() string {
	return s.Token.Literal
}

func (s *ContinueStatement) String() string {
	var out bytes.Buffer
	out.WriteString(s.TokenLiteral())
	out.WriteString(";")
	return out.String()
}

type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (s *AssignStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *AssignStatement) TokenLiteral() string {
	return s.Token.Literal
}

// String returns this object as a string.
func (as *AssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	out.WriteString(as.Value.String())
	out.WriteString(";")
	return out.String()
}

type IndexAssignStatement struct {
	Token token.Token
	Left  Expression // the array/map expression
	Index Expression // the index expression
	Value Expression
}

func (s *IndexAssignStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (s *IndexAssignStatement) TokenLiteral() string {
	return s.Token.Literal
}

func (s *IndexAssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(s.Left.String())
	out.WriteString("[")
	out.WriteString(s.Index.String())
	out.WriteString("] = ")
	out.WriteString(s.Value.String())
	out.WriteString(";")
	return out.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (e *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *IntegerLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *IntegerLiteral) String() string {
	return e.Token.Literal
}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (e *FloatLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *FloatLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *FloatLiteral) String() string {
	return e.Token.Literal
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (e *StringLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *StringLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *StringLiteral) String() string {
	return e.Value
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (e *BooleanLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *BooleanLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *BooleanLiteral) String() string {
	return e.Token.Literal
}

type NilLiteral struct {
	Token token.Token
}

func (e *NilLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *NilLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *NilLiteral) String() string {
	return "nil"
}

type ArrayLiteral struct {
	Token    token.Token // the [ token
	Elements []Expression
}

func (e *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *ArrayLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *ArrayLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, el := range e.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(el.String())
	}
	out.WriteString("]")
	return out.String()
}

type MapLiteral struct {
	Token token.Token // the { token
	Pairs map[Expression]Expression
	Keys  []Expression // preserve order
}

func (e *MapLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *MapLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *MapLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	for i, key := range e.Keys {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(key.String())
		out.WriteString(": ")
		out.WriteString(e.Pairs[key].String())
	}
	out.WriteString("}")
	return out.String()
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (e *PrefixExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *PrefixExpression) TokenLiteral() string {
	return e.Token.Literal
}

func (e *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(e.Operator)
	out.WriteString(e.Right.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (e *InfixExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *InfixExpression) TokenLiteral() string {
	return e.Token.Literal
}

func (e *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(e.Left.String())
	out.WriteString(" " + e.Operator + " ")
	out.WriteString(e.Right.String())
	out.WriteString(")")
	return out.String()
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (e *IfExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *IfExpression) TokenLiteral() string {
	return e.Token.Literal
}

func (e *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(e.Condition.String())
	out.WriteString(" { ")
	out.WriteString(e.Consequence.String())
	out.WriteString(" }")
	if e.Alternative != nil {
		out.WriteString(" else { ")
		out.WriteString(e.Alternative.String())
		out.WriteString(" }")
	}
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token
	Name       string // optional, empty for anonymous
	Parameters []*Identifier
	Body       *BlockStatement
}

func (e *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *FunctionLiteral) TokenLiteral() string {
	return e.Token.Literal
}

func (e *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("fn")
	if e.Name != "" {
		out.WriteString(" ")
		out.WriteString(e.Name)
	}
	out.WriteString("(")
	for i, p := range e.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") { ")
	out.WriteString(e.Body.String())
	out.WriteString(" }")
	return out.String()
}

type CallExpression struct {
	Token     token.Token // the ( token
	Function  Expression
	Arguments []Expression
}

func (e *CallExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *CallExpression) TokenLiteral() string {
	return e.Token.Literal
}
func (e *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(e.Function.String())
	out.WriteString("(")
	for i, a := range e.Arguments {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(a.String())
	}
	out.WriteString(")")
	return out.String()
}

type IndexExpression struct {
	Token token.Token // the [ token
	Left  Expression
	Index Expression
}

func (e *IndexExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (e *IndexExpression) TokenLiteral() string {
	return e.Token.Literal
}

// String returns this object as a string.
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

package ast

import "github.com/hilthontt/lotus/token"

// x += 1   x -= 2   x *= 3   etc.
type CompoundAssignStatement struct {
	Token    token.Token // the operator token e.g. +=
	Name     Expression  // the assignment target (Identifier or IndexExpression or FieldAccess)
	Operator string      // "+=" "-=" "*=" etc.
	Value    Expression
}

func (ca *CompoundAssignStatement) statementNode() {}

func (ca *CompoundAssignStatement) TokenLiteral() string {
	return ca.Token.Literal
}

func (ca *CompoundAssignStatement) String() string {
	return ca.Name.String() + " " + ca.Operator + " " + ca.Value.String()
}

package ast

// TypeAnnotation holds an optional type name, e.g. `: int` or `-> string`.
// It is purely informational — the compiler ignores it.
type TypeAnnotation struct {
	Name string // "int", "float", "string", "bool", "array", "map", or a class name
}

func (ta *TypeAnnotation) expressionNode() {}

func (ta *TypeAnnotation) TokenLiteral() string {
	return ta.Name
}

func (ta *TypeAnnotation) String() string {
	return ta.Name
}

package evaluator

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/object"
)

// Sentinel types for break/continue control flow
type BreakSignal struct{}

func (b *BreakSignal) Type() object.ObjectType {
	return "BREAK"
}

func (b *BreakSignal) Inspect() string {
	return "break"
}

type ContinueSignal struct{}

func (c *ContinueSignal) Type() object.ObjectType {
	return "CONTINUE"
}

func (c *ContinueSignal) Inspect() string {
	return "continue"
}

type Error struct {
	Message string
}

func (e *Error) Type() object.ObjectType {
	return "ERROR"
}

func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

// ReturnValue wraps a return value to unwind the call stack
type ReturnValue struct {
	Value object.Object
}

func (rv *ReturnValue) Type() object.ObjectType {
	return "RETURN_VALUE"
}

func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Function struct {
	Name       string
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *object.Environment
}

func (f *Function) Type() object.ObjectType {
	return "FUNCTION"
}

func (f *Function) Inspect() string {
	return fmt.Sprintf("fn<%s>", f.Name)
}

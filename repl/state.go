package repl

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/vm"
)

type vmState struct {
	symbolTable *compiler.SymbolTable
	constants   []object.Object
	globals     []object.Object
}

func newVMState() *vmState {
	st := compiler.NewSymbolTable()
	for i, b := range compiler.Builtins {
		st.DefineBuiltin(i, b.Name)
	}
	return &vmState{
		symbolTable: st,
		constants:   []object.Object{},
		globals:     make([]object.Object, vm.GlobalsSize),
	}
}

func (s vmState) run(program *ast.Program) (object.Object, error) {
	comp := compiler.NewWithState(s.symbolTable, s.constants)
	if err := comp.Compile(program); err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	code := comp.Bytecode()
	s.constants = code.Constants

	machine := vm.NewWithGlobalsState(code, s.globals)
	if err := machine.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %w", err)
	}

	return machine.LastPoppedStackElement(), nil
}

package evaluator

import (
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/object"
)

var builtinFunctions map[string]*object.Builtin

func init() {
	builtinFunctions = make(map[string]*object.Builtin, len(compiler.Builtins))
	for _, b := range compiler.GetBuiltins() {
		builtinFunctions[b.Name] = b
	}
}

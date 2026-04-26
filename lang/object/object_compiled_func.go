package object

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hilthontt/lotus/code"
)

type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
	Name         string
	IsVariadic   bool
}

func (o *CompiledFunction) Type() ObjectType {
	return "COMPILED_FUNCTION"
}

func (o *CompiledFunction) Inspect() string {
	return fmt.Sprintf("fn<%s>", o.Name)
}

func (o *CompiledFunction) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		static := []string{"methods"}
		dynamic := env.Names("function.")

		var names []string
		names = append(names, static...)
		for _, e := range dynamic {
			bits := strings.Split(e, ".")
			names = append(names, bits[1])
		}
		sort.Strings(names)

		result := make([]Object, len(names))
		for i, txt := range names {
			result[i] = &String{Value: txt}
		}
		return &Array{Elements: result}
	}
	return nil
}

func (o *CompiledFunction) ToInterface() any {
	return "<COMPILED_FUNCTION>"
}

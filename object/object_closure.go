package object

import (
	"fmt"
	"sort"
	"strings"
)

type Closure struct {
	Fn            *CompiledFunction
	Free          []Object // captured variables
	DefiningClass *Class   // non-nil when this closure is a class method
}

func (o *Closure) Type() ObjectType {
	return CLOSURE_OBJ
}

func (o *Closure) Inspect() string {
	return fmt.Sprintf("closure[%p]", o.Fn)
}

func (o *Closure) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "methods" {
		static := []string{"methods"}
		dynamic := env.Names("closure.")

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

func (o *Closure) ToInterface() any {
	return "<CLOSURE>"
}

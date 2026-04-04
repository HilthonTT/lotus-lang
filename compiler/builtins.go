package compiler

import (
	"fmt"
	"strings"

	"github.com/hilthontt/lotus/object"
)

type BuiltinDef struct {
	Name string
	Fn   object.BuiltinFunction
}

var Builtins = []BuiltinDef{
	{
		Name: "print",
		Fn: func(args ...object.Object) object.Object {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = a.Inspect()
			}
			fmt.Println(strings.Join(parts, " "))
			return &object.Nil{}
		},
	},
	{
		Name: "len",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			switch a := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(a.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(a.Elements))}
			case *object.Map:
				return &object.Integer{Value: int64(len(a.Pairs))}
			default:
				return &object.Nil{}
			}
		},
	},
	{
		Name: "push",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return &object.Nil{}
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return &object.Nil{}
			}
			newElems := make([]object.Object, len(arr.Elements)+1)
			copy(newElems, arr.Elements)
			newElems[len(arr.Elements)] = args[1]
			return &object.Array{Elements: newElems}
		},
	},
	{
		Name: "pop",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			arr, ok := args[0].(*object.Array)
			if !ok || len(arr.Elements) == 0 {
				return &object.Nil{}
			}
			return arr.Elements[len(arr.Elements)-1]
		},
	},
	{
		Name: "head",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			arr, ok := args[0].(*object.Array)
			if !ok || len(arr.Elements) == 0 {
				return &object.Nil{}
			}
			return arr.Elements[0]
		},
	},
	{
		Name: "tail",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			arr, ok := args[0].(*object.Array)
			if !ok || len(arr.Elements) < 2 {
				return &object.Array{Elements: []object.Object{}}
			}
			newElems := make([]object.Object, len(arr.Elements)-1)
			copy(newElems, arr.Elements[1:])
			return &object.Array{Elements: newElems}
		},
	},
	{
		Name: "type",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	{
		Name: "str",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			return &object.String{Value: args[0].Inspect()}
		},
	},
	{
		Name: "int",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Nil{}
			}
			switch a := args[0].(type) {
			case *object.Integer:
				return a
			case *object.Float:
				return &object.Integer{Value: int64(a.Value)}
			case *object.Boolean:
				if a.Value {
					return &object.Integer{Value: 1}
				}
				return &object.Integer{Value: 0}
			default:
				return &object.Nil{}
			}
		},
	},
	{
		Name: "range",
		Fn: func(args ...object.Object) object.Object {
			var start, end, step int64
			switch len(args) {
			case 1:
				e, ok := args[0].(*object.Integer)
				if !ok {
					return &object.Nil{}
				}
				start, end, step = 0, e.Value, 1
			case 2:
				s, ok1 := args[0].(*object.Integer)
				e, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				start, end, step = s.Value, e.Value, 1
			case 3:
				s, ok1 := args[0].(*object.Integer)
				e, ok2 := args[1].(*object.Integer)
				st, ok3 := args[2].(*object.Integer)
				if !ok1 || !ok2 || !ok3 || st.Value == 0 {
					return &object.Nil{}
				}
				start, end, step = s.Value, e.Value, st.Value
			default:
				return &object.Nil{}
			}

			var elems []object.Object
			if step > 0 {
				for i := start; i < end; i += step {
					elems = append(elems, &object.Integer{Value: i})
				}
			} else {
				for i := start; i > end; i += step {
					elems = append(elems, &object.Integer{Value: i})
				}
			}
			return &object.Array{Elements: elems}
		},
	},
}

func builtinIndex(name string) int {
	for i, b := range Builtins {
		if b.Name == name {
			return i
		}
	}
	return -1
}

func GetBuiltins() []*object.Builtin {
	builtins := make([]*object.Builtin, len(Builtins))
	for i, b := range Builtins {
		builtins[i] = &object.Builtin{Name: b.Name, Fn: b.Fn}
	}
	return builtins
}

func GetBuiltinByName(name string) *BuiltinDef {
	for _, def := range Builtins {
		if def.Name == name {
			return &def
		}
	}
	return nil
}

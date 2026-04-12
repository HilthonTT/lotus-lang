package compiler

import (
	"fmt"
	"sort"

	"github.com/hilthontt/lotus/object"
)

func arrayPackage() *object.Package {
	pkg := &object.Package{
		Name:      "Array",
		Functions: map[string]object.PackageFunction{},
	}

	// Array.filter(arr, fn(elem) -> bool) -> array
	pkg.Functions["filter"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		var result []object.Object
		for _, elem := range arr.Elements {
			ret := pkg.CallVM(fn, []object.Object{elem})
			if object.IsTruthy(ret) {
				result = append(result, elem)
			}
		}
		if result == nil {
			result = []object.Object{}
		}
		return &object.Array{Elements: result}
	}

	// Array.map(arr, fn(elem) -> any) -> array
	pkg.Functions["map"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		result := make([]object.Object, len(arr.Elements))
		for i, elem := range arr.Elements {
			result[i] = pkg.CallVM(fn, []object.Object{elem})
		}
		return &object.Array{Elements: result}
	}

	// Array.reduce(arr, fn(acc, elem) -> any, initial) -> any
	pkg.Functions["reduce"] = func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		acc := args[2]
		for _, elem := range arr.Elements {
			acc = pkg.CallVM(fn, []object.Object{acc, elem})
		}
		return acc
	}

	// Array.find(arr, fn(elem) -> bool) -> elem | nil
	pkg.Functions["find"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		for _, elem := range arr.Elements {
			ret := pkg.CallVM(fn, []object.Object{elem})
			if object.IsTruthy(ret) {
				return elem
			}
		}
		return &object.Nil{}
	}

	// Array.findIndex(arr, fn(elem) -> bool) -> int (-1 if not found)
	pkg.Functions["findIndex"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		for i, elem := range arr.Elements {
			ret := pkg.CallVM(fn, []object.Object{elem})
			if object.IsTruthy(ret) {
				return &object.Integer{Value: int64(i)}
			}
		}
		return &object.Integer{Value: -1}
	}

	// Array.forEach(arr, fn(elem)) — like map but discards result
	pkg.Functions["forEach"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		for _, elem := range arr.Elements {
			pkg.CallVM(fn, []object.Object{elem})
		}
		return &object.Nil{}
	}

	// Array.contains(arr, value) -> bool
	pkg.Functions["contains"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Boolean{Value: false}
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return &object.Boolean{Value: false}
		}
		target := args[1]
		for _, elem := range arr.Elements {
			if elem.Inspect() == target.Inspect() {
				return &object.Boolean{Value: true}
			}
		}
		return &object.Boolean{Value: false}
	}

	// Array.reverse(arr) -> array
	pkg.Functions["reverse"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return &object.Nil{}
		}
		n := len(arr.Elements)
		result := make([]object.Object, n)
		for i, elem := range arr.Elements {
			result[n-1-i] = elem
		}
		return &object.Array{Elements: result}
	}

	// Array.sort(arr) -> sorted array (numbers or strings, ascending)
	pkg.Functions["sort"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		arr, ok := args[0].(*object.Array)
		if !ok || len(arr.Elements) == 0 {
			return arr
		}
		result := make([]object.Object, len(arr.Elements))
		copy(result, arr.Elements)
		sort.Slice(result, func(i, j int) bool {
			a, b := result[i], result[j]
			switch av := a.(type) {
			case *object.Integer:
				if bv, ok := b.(*object.Integer); ok {
					return av.Value < bv.Value
				}
			case *object.Float:
				bv := toFloat64Obj(b)
				return av.Value < bv
			case *object.String:
				if bv, ok := b.(*object.String); ok {
					return av.Value < bv.Value
				}
			}
			return a.Inspect() < b.Inspect()
		})
		return &object.Array{Elements: result}
	}

	// Array.sortBy(arr, fn(elem) -> comparable) -> sorted array
	pkg.Functions["sortBy"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Nil{}
		}
		result := make([]object.Object, len(arr.Elements))
		copy(result, arr.Elements)
		sort.SliceStable(result, func(i, j int) bool {
			ki := pkg.CallVM(fn, []object.Object{result[i]})
			kj := pkg.CallVM(fn, []object.Object{result[j]})
			switch kiv := ki.(type) {
			case *object.Integer:
				if kjv, ok := kj.(*object.Integer); ok {
					return kiv.Value < kjv.Value
				}
			case *object.Float:
				return kiv.Value < toFloat64Obj(kj)
			case *object.String:
				if kjv, ok := kj.(*object.String); ok {
					return kiv.Value < kjv.Value
				}
			}
			return ki.Inspect() < kj.Inspect()
		})
		return &object.Array{Elements: result}
	}

	// Array.flat(arr) -> flattened one level
	pkg.Functions["flat"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return &object.Nil{}
		}
		var result []object.Object
		for _, elem := range arr.Elements {
			if sub, ok := elem.(*object.Array); ok {
				result = append(result, sub.Elements...)
			} else {
				result = append(result, elem)
			}
		}
		if result == nil {
			result = []object.Object{}
		}
		return &object.Array{Elements: result}
	}

	// Array.join(arr, sep) -> string
	pkg.Functions["join"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		sep, ok2 := args[1].(*object.String)
		if !ok1 || !ok2 {
			return &object.Nil{}
		}
		parts := make([]string, len(arr.Elements))
		for i, el := range arr.Elements {
			parts[i] = el.Inspect()
		}
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += sep.Value
			}
			result += p
		}
		return &object.String{Value: result}
	}

	// Array.slice(arr, start, end) -> subarray
	pkg.Functions["slice"] = func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return &object.Nil{}
		}
		arr, ok1 := args[0].(*object.Array)
		start, ok2 := args[1].(*object.Integer)
		end, ok3 := args[2].(*object.Integer)
		if !ok1 || !ok2 || !ok3 {
			return &object.Nil{}
		}
		n := int64(len(arr.Elements))
		st, en := start.Value, end.Value
		if st < 0 {
			st = n + st
		}
		if en < 0 {
			en = n + en
		}
		if st < 0 {
			st = 0
		}
		if en > n {
			en = n
		}
		if st >= en {
			return &object.Array{Elements: []object.Object{}}
		}
		result := make([]object.Object, en-st)
		copy(result, arr.Elements[st:en])
		return &object.Array{Elements: result}
	}

	// Array.unique(arr) -> deduplicated array (preserves order)
	pkg.Functions["unique"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return &object.Nil{}
		}
		seen := map[string]bool{}
		var result []object.Object
		for _, elem := range arr.Elements {
			key := fmt.Sprintf("%s:%s", elem.Type(), elem.Inspect())
			if !seen[key] {
				seen[key] = true
				result = append(result, elem)
			}
		}
		if result == nil {
			result = []object.Object{}
		}
		return &object.Array{Elements: result}
	}

	// Array.len(arr) -> int
	pkg.Functions["len"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		arr, ok := args[0].(*object.Array)
		if !ok {
			return &object.Nil{}
		}
		return &object.Integer{Value: int64(len(arr.Elements))}
	}

	// Array.any(arr, fn(elem) -> bool) -> bool
	pkg.Functions["any"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Boolean{Value: false}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Boolean{Value: false}
		}
		for _, elem := range arr.Elements {
			if object.IsTruthy(pkg.CallVM(fn, []object.Object{elem})) {
				return &object.Boolean{Value: true}
			}
		}
		return &object.Boolean{Value: false}
	}

	// Array.all(arr, fn(elem) -> bool) -> bool
	pkg.Functions["all"] = func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Boolean{Value: false}
		}
		arr, ok1 := args[0].(*object.Array)
		fn, ok2 := args[1].(*object.Closure)
		if !ok1 || !ok2 || pkg.CallVM == nil {
			return &object.Boolean{Value: false}
		}
		for _, elem := range arr.Elements {
			if !object.IsTruthy(pkg.CallVM(fn, []object.Object{elem})) {
				return &object.Boolean{Value: false}
			}
		}
		return &object.Boolean{Value: true}
	}

	return pkg
}

func toFloat64Obj(o object.Object) float64 {
	switch v := o.(type) {
	case *object.Integer:
		return float64(v.Value)
	case *object.Float:
		return v.Value
	}
	return 0
}

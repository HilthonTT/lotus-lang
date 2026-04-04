package object

import (
	"fmt"
	"sort"
	"strings"
)

type MapPair struct {
	Key   Object
	Value Object
}

type Map struct {
	Pairs map[string]MapPair
	Order []string
}

func (m *Map) Type() ObjectType {
	return MAP_OBJ
}

func (m *Map) Inspect() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, k := range m.Order {
		p := m.Pairs[k]
		pairs = append(pairs, fmt.Sprintf("%s: %s", p.Key.Inspect(), p.Value.Inspect()))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (m *Map) InvokeMethod(method string, env Environment, args ...Object) Object {
	if method == "len" {
		return &Integer{Value: int64(len(m.Pairs))}
	}
	if method == "keys" {
		keys := make([]Object, 0, len(m.Order))
		for _, k := range m.Order {
			keys = append(keys, m.Pairs[k].Key)
		}
		return &Array{Elements: keys}
	}
	if method == "values" {
		values := make([]Object, 0, len(m.Order))
		for _, k := range m.Order {
			values = append(values, m.Pairs[k].Value)
		}
		return &Array{Elements: values}
	}
	if method == "methods" {
		static := []string{"keys", "len", "methods", "values"}
		dynamic := env.Names("map.")

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

func (m *Map) ToInterface() any {
	return "<MAP>"
}

package compiler

import (
	"strings"
	"unicode/utf8"

	"github.com/hilthontt/lotus/object"
)

func stringPackage() *object.Package {
	return &object.Package{
		Name: "String",
		Functions: map[string]object.PackageFunction{
			// String.split(str, sep) -> array
			"split": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				sep, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				parts := strings.Split(s.Value, sep.Value)
				elems := make([]object.Object, len(parts))
				for i, p := range parts {
					elems[i] = &object.String{Value: p}
				}
				return &object.Array{Elements: elems}
			},

			// String.trim(str) -> string
			"trim": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimSpace(s.Value)}
			},

			// String.trimLeft(str) -> string
			"trimLeft": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimLeftFunc(s.Value, func(r rune) bool {
					return r == ' ' || r == '\t' || r == '\n' || r == '\r'
				})}
			},

			// String.trimRight(str) -> string
			"trimRight": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimRightFunc(s.Value, func(r rune) bool {
					return r == ' ' || r == '\t' || r == '\n' || r == '\r'
				})}
			},

			// String.upper(str) -> string
			"upper": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.ToUpper(s.Value)}
			},

			// String.lower(str) -> string
			"lower": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.ToLower(s.Value)}
			},

			// String.replace(str, old, new) -> string
			"replace": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				old, ok2 := args[1].(*object.String)
				neu, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, neu.Value)}
			},

			// String.contains(str, substr) -> bool
			"contains": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				sub, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Boolean{Value: strings.Contains(s.Value, sub.Value)}
			},

			// String.startsWith(str, prefix) -> bool
			"startsWith": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				prefix, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Boolean{Value: strings.HasPrefix(s.Value, prefix.Value)}
			},

			// String.endsWith(str, suffix) -> bool
			"endsWith": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				suffix, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Boolean{Value: strings.HasSuffix(s.Value, suffix.Value)}
			},

			// String.indexOf(str, substr) -> int (-1 if not found)
			"indexOf": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				sub, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(strings.Index(s.Value, sub.Value))}
			},

			// String.repeat(str, n) -> string
			"repeat": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 || n.Value < 0 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
			},

			// String.padLeft(str, n, char) -> string
			"padLeft": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				pad, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 || len(pad.Value) == 0 {
					return &object.Nil{}
				}
				cur := utf8.RuneCountInString(s.Value)
				needed := int(n.Value) - cur
				if needed <= 0 {
					return s
				}
				fill := strings.Repeat(pad.Value, needed)
				return &object.String{Value: fill + s.Value}
			},

			// String.padRight(str, n, char) -> string
			"padRight": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				pad, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 || len(pad.Value) == 0 {
					return &object.Nil{}
				}
				cur := utf8.RuneCountInString(s.Value)
				needed := int(n.Value) - cur
				if needed <= 0 {
					return s
				}
				fill := strings.Repeat(pad.Value, needed)
				return &object.String{Value: s.Value + fill}
			},

			// String.chars(str) -> array of single-char strings
			"chars": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				runes := []rune(s.Value)
				elems := make([]object.Object, len(runes))
				for i, r := range runes {
					elems[i] = &object.String{Value: string(r)}
				}
				return &object.Array{Elements: elems}
			},

			// String.len(str) -> int
			"len": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(utf8.RuneCountInString(s.Value))}
			},

			// String.join(array, sep) -> string
			"join": func(args ...object.Object) object.Object {
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
				return &object.String{Value: strings.Join(parts, sep.Value)}
			},

			// String.slice(str, start, end) -> string
			"slice": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				start, ok2 := args[1].(*object.Integer)
				end, ok3 := args[2].(*object.Integer)
				if !ok1 || !ok2 || !ok3 {
					return &object.Nil{}
				}
				runes := []rune(s.Value)
				n := int64(len(runes))
				st := start.Value
				en := end.Value
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
					return &object.String{Value: ""}
				}
				return &object.String{Value: string(runes[st:en])}
			},
		},
	}
}

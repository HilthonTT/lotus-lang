package compiler

import (
	"strings"
	"unicode"

	"github.com/hilthontt/lotus/object"
)

func stringPackage() *object.Package {
	return &object.Package{
		Name: "String",
		Functions: map[string]object.PackageFunction{

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

			"trimLeft": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimLeftFunc(s.Value, unicode.IsSpace)}
			},

			"trimRight": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimRightFunc(s.Value, unicode.IsSpace)}
			},

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

			"replace": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				old, ok2 := args[1].(*object.String)
				new_, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, new_.Value)}
			},

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

			"repeat": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
			},

			"padLeft": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				pad, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 {
					return &object.Nil{}
				}
				result := s.Value
				padChar := pad.Value
				if len(padChar) == 0 {
					padChar = " "
				}
				for len(result) < int(n.Value) {
					result = padChar + result
				}
				return &object.String{Value: result}
			},

			"padRight": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				n, ok2 := args[1].(*object.Integer)
				pad, ok3 := args[2].(*object.String)
				if !ok1 || !ok2 || !ok3 {
					return &object.Nil{}
				}
				result := s.Value
				padChar := pad.Value
				if len(padChar) == 0 {
					padChar = " "
				}
				for len(result) < int(n.Value) {
					result = result + padChar
				}
				return &object.String{Value: result}
			},

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

			"len": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(len([]rune(s.Value)))}
			},

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
				l := int64(len(runes))
				st := start.Value
				en := end.Value
				if st < 0 {
					st = 0
				}
				if en > l {
					en = l
				}
				if st > en {
					return &object.String{Value: ""}
				}
				return &object.String{Value: string(runes[st:en])}
			},

			// ── New ───────────────────────────────────────────────────────────

			// String.trimPrefix(s, prefix) -> string
			"trimPrefix": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				prefix, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimPrefix(s.Value, prefix.Value)}
			},

			// String.trimSuffix(s, suffix) -> string
			"trimSuffix": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				suffix, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.String{Value: strings.TrimSuffix(s.Value, suffix.Value)}
			},

			// String.count(s, substr) -> int
			"count": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				sub, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(strings.Count(s.Value, sub.Value))}
			},

			// String.isDigit(s) -> bool
			"isDigit": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				s, ok := args[0].(*object.String)
				if !ok || len(s.Value) == 0 {
					return &object.Boolean{Value: false}
				}
				for _, r := range s.Value {
					if !unicode.IsDigit(r) {
						return &object.Boolean{Value: false}
					}
				}
				return &object.Boolean{Value: true}
			},

			// String.isAlpha(s) -> bool
			"isAlpha": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				s, ok := args[0].(*object.String)
				if !ok || len(s.Value) == 0 {
					return &object.Boolean{Value: false}
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) {
						return &object.Boolean{Value: false}
					}
				}
				return &object.Boolean{Value: true}
			},

			// String.isAlphaNum(s) -> bool
			"isAlphaNum": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				s, ok := args[0].(*object.String)
				if !ok || len(s.Value) == 0 {
					return &object.Boolean{Value: false}
				}
				for _, r := range s.Value {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						return &object.Boolean{Value: false}
					}
				}
				return &object.Boolean{Value: true}
			},

			// String.reverse(s) -> string
			"reverse": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				runes := []rune(s.Value)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return &object.String{Value: string(runes)}
			},

			// String.lines(s) -> array  (split by \n)
			"lines": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				parts := strings.Split(s.Value, "\n")
				elems := make([]object.Object, len(parts))
				for i, p := range parts {
					elems[i] = &object.String{Value: p}
				}
				return &object.Array{Elements: elems}
			},

			// String.toBytes(s) -> array of ints
			"toBytes": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				bytes := []byte(s.Value)
				elems := make([]object.Object, len(bytes))
				for i, b := range bytes {
					elems[i] = &object.Integer{Value: int64(b)}
				}
				return &object.Array{Elements: elems}
			},

			// String.fromBytes(arr) -> string
			"fromBytes": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				arr, ok := args[0].(*object.Array)
				if !ok {
					return &object.Nil{}
				}
				bytes := make([]byte, len(arr.Elements))
				for i, el := range arr.Elements {
					n, ok := el.(*object.Integer)
					if !ok {
						return &object.Nil{}
					}
					bytes[i] = byte(n.Value)
				}
				return &object.String{Value: string(bytes)}
			},

			// String.format(template, ...args) -> string
			// Simple %s %d %f substitution
			"format": func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return &object.Nil{}
				}
				tmpl, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				result := tmpl.Value
				argIdx := 1
				for strings.Contains(result, "%s") || strings.Contains(result, "%d") || strings.Contains(result, "%f") {
					if argIdx >= len(args) {
						break
					}
					val := args[argIdx].Inspect()
					// Replace first occurrence
					for _, placeholder := range []string{"%s", "%d", "%f"} {
						idx := strings.Index(result, placeholder)
						if idx != -1 {
							result = result[:idx] + val + result[idx+len(placeholder):]
							argIdx++
							break
						}
					}
				}
				return &object.String{Value: result}
			},

			// String.lastIndexOf(s, substr) -> int
			"lastIndexOf": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				s, ok1 := args[0].(*object.String)
				sub, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(strings.LastIndex(s.Value, sub.Value))}
			},

			// String.title(s) -> string  (Title Case)
			"title": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: strings.Title(s.Value)} //nolint:staticcheck
			},
		},
	}
}

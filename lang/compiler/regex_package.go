package compiler

import (
	"errors"
	"regexp"

	"github.com/hilthontt/lotus/object"
)

// regexObj wraps a compiled *regexp.Regexp as a Lotus object.
type regexObj struct {
	Re      *regexp.Regexp
	Pattern string
}

func (r *regexObj) Type() object.ObjectType {
	return "REGEX"
}

func (r *regexObj) Inspect() string {
	return "/" + r.Pattern + "/"
}

func regexPackage() *object.Package {
	return &object.Package{
		Name: "Regex",
		Functions: map[string]object.PackageFunction{

			// Regex.compile(pattern) -> Regex | nil
			"compile": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				re, err := regexp.Compile(s.Value)
				if err != nil {
					return &object.Nil{}
				}
				return &regexObj{Re: re, Pattern: s.Value}
			},

			// Regex.test(pattern, str) -> bool
			// Accepts a compiled Regex or a pattern string.
			"test": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Boolean{Value: false}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: re.MatchString(s.Value)}
			},

			// Regex.find(pattern, str) -> string | nil
			// Returns the first match.
			"find": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				match := re.FindString(s.Value)
				if match == "" {
					return &object.Nil{}
				}
				return &object.String{Value: match}
			},

			// Regex.findAll(pattern, str) -> array
			// Returns all non-overlapping matches.
			"findAll": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				matches := re.FindAllString(s.Value, -1)
				elems := make([]object.Object, len(matches))
				for i, m := range matches {
					elems[i] = &object.String{Value: m}
				}
				return &object.Array{Elements: elems}
			},

			// Regex.replace(pattern, str, replacement) -> string
			"replace": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok1 := args[1].(*object.String)
				repl, ok2 := args[2].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				return &object.String{Value: re.ReplaceAllString(s.Value, repl.Value)}
			},

			// Regex.replaceFirst(pattern, str, replacement) -> string
			"replaceFirst": func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok1 := args[1].(*object.String)
				repl, ok2 := args[2].(*object.String)
				if !ok1 || !ok2 {
					return &object.Nil{}
				}
				found := false
				result := re.ReplaceAllStringFunc(s.Value, func(match string) string {
					if !found {
						found = true
						return repl.Value
					}
					return match
				})
				return &object.String{Value: result}
			},

			// Regex.split(pattern, str) -> array
			"split": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				parts := re.Split(s.Value, -1)
				elems := make([]object.Object, len(parts))
				for i, p := range parts {
					elems[i] = &object.String{Value: p}
				}
				return &object.Array{Elements: elems}
			},

			// Regex.groups(pattern, str) -> array  (captured groups)
			"groups": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				match := re.FindStringSubmatch(s.Value)
				if match == nil {
					return &object.Array{Elements: []object.Object{}}
				}
				// Skip match[0] (full match), return capture groups
				elems := make([]object.Object, len(match)-1)
				for i, g := range match[1:] {
					elems[i] = &object.String{Value: g}
				}
				return &object.Array{Elements: elems}
			},

			// Regex.groupsAll(pattern, str) -> array of arrays
			"groupsAll": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Nil{}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				allMatches := re.FindAllStringSubmatch(s.Value, -1)
				outer := make([]object.Object, len(allMatches))
				for i, match := range allMatches {
					inner := make([]object.Object, len(match)-1)
					for j, g := range match[1:] {
						inner[j] = &object.String{Value: g}
					}
					outer[i] = &object.Array{Elements: inner}
				}
				return &object.Array{Elements: outer}
			},

			// Regex.escape(str) -> string  (escapes special chars)
			"escape": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				return &object.String{Value: regexp.QuoteMeta(s.Value)}
			},

			// Regex.count(pattern, str) -> int  (number of matches)
			"count": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Integer{Value: 0}
				}
				re, err := toRegex(args[0])
				if err != nil {
					return &object.Integer{Value: 0}
				}
				s, ok := args[1].(*object.String)
				if !ok {
					return &object.Integer{Value: 0}
				}
				matches := re.FindAllString(s.Value, -1)
				return &object.Integer{Value: int64(len(matches))}
			},
		},
	}
}

// toRegex accepts either a *regexObj (pre-compiled) or a *object.String (pattern).
func toRegex(obj object.Object) (*regexp.Regexp, error) {
	switch v := obj.(type) {
	case *regexObj:
		return v.Re, nil
	case *object.String:
		return regexp.Compile(v.Value)
	}
	return nil, errors.New("invalid internal order")
}

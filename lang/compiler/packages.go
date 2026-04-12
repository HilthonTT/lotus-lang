package compiler

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"github.com/hilthontt/lotus/object"
)

// BuiltinPackageOrder defines the registration order so global indices
// are identical across every compiler.New() call.
var BuiltinPackageOrder = []string{"Console", "Math", "OS", "Task", "Array", "String"}

// Add new packages here — they are automatically injected as globals.
var BuiltinPackages = map[string]*object.Package{
	"Console": consolePackage(),
	"Math":    mathPackage(),
	"OS":      osPackage(),
	"Task":    taskPackage(),
	"Array":   arrayPackage(),
	"String":  stringPackage(),
}

func consolePackage() *object.Package {
	return &object.Package{
		Name: "Console",
		Functions: map[string]object.PackageFunction{
			// readLine() -> string  — reads a full line from stdin
			"readLine": func(args ...object.Object) object.Object {
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					return &object.String{Value: scanner.Text()}
				}
				return &object.String{Value: ""}
			},

			// readLine with prompt: Console.prompt("Name: ") -> string
			"prompt": func(args ...object.Object) object.Object {
				if len(args) == 1 {
					fmt.Print(args[0].Inspect())
				}
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					return &object.String{Value: scanner.Text()}
				}
				return &object.String{Value: ""}
			},

			// clear() — clears the terminal screen
			"clear": func(args ...object.Object) object.Object {
				fmt.Print("\033[H\033[2J")
				return &object.Nil{}
			},

			// print(...) — same as the global print but namespaced
			"print": func(args ...object.Object) object.Object {
				parts := make([]string, len(args))
				for i, a := range args {
					parts[i] = a.Inspect()
				}
				fmt.Println(strings.Join(parts, " "))
				return &object.Nil{}
			},

			// printErr(...) — writes to stderr
			"printErr": func(args ...object.Object) object.Object {
				parts := make([]string, len(args))
				for i, a := range args {
					parts[i] = a.Inspect()
				}
				fmt.Fprintln(os.Stderr, strings.Join(parts, " "))
				return &object.Nil{}
			},
		},
	}
}

func mathPackage() *object.Package {
	return &object.Package{
		Name: "Math",
		Functions: map[string]object.PackageFunction{
			"sqrt": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				v := toFloat64(args[0])
				return &object.Float{Value: mathSqrt(v)}
			},
			"abs": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				switch a := args[0].(type) {
				case *object.Integer:
					v := a.Value
					if v < 0 {
						v = -v
					}
					return &object.Integer{Value: v}
				case *object.Float:
					v := a.Value
					if v < 0 {
						v = -v
					}
					return &object.Float{Value: v}
				}
				return &object.Nil{}
			},
			"floor": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				return &object.Integer{Value: int64(toFloat64(args[0]))}
			},
			"pow": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}

				base := toFloat64(args[0])
				exp := toFloat64(args[1])
				return &object.Float{Value: mathPow(base, exp)}
			},
			"max": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, b := toFloat64(args[0]), toFloat64(args[1])
				if a > b {
					return args[0] // Return a
				}
				return args[1] // Return b
			},
			"min": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				a, b := toFloat64(args[0]), toFloat64(args[1])
				if a < b {
					return args[0] // Return b
				}
				return args[1] // Return a
			},
			"pi": func(args ...object.Object) object.Object {
				return &object.Float{Value: 3.141592653589793}
			},
			"random": func(args ...object.Object) object.Object {
				randomNum := rand.Float64()
				return &object.Float{Value: randomNum}
			},
		},
	}
}

func osPackage() *object.Package {
	return &object.Package{
		Name: "OS",
		Functions: map[string]object.PackageFunction{
			"exit": func(args ...object.Object) object.Object {
				code := 0
				if len(args) == 1 {
					if i, ok := args[0].(*object.Integer); ok {
						code = int(i.Value)
					}
				}
				os.Exit(code)
				return &object.Nil{}
			},
			"args": func(args ...object.Object) object.Object {
				elems := make([]object.Object, len(os.Args))
				for i, a := range os.Args {
					elems[i] = &object.String{Value: a}
				}
				return &object.Array{Elements: elems}
			},
			"env": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				key := args[0].Inspect()
				val := os.Getenv(key)
				if val == "" {
					return &object.Nil{}
				}
				return &object.String{Value: val}
			},
			"readFile": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				data, err := os.ReadFile(args[0].Inspect())
				if err != nil {
					return &object.Nil{}
				}
				return &object.String{Value: string(data)}
			},
			"writeFile": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				err := os.WriteFile(args[0].Inspect(), []byte(args[1].Inspect()), 0644)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: true}
			},
			"parseInt": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				i, err := strconv.ParseInt(args[0].Inspect(), 10, 64)
				if err != nil {
					return &object.Nil{}
				}
				return &object.Integer{Value: i}
			},
			"parseFloat": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				f, err := strconv.ParseFloat(args[0].Inspect(), 64)
				if err != nil {
					return &object.Nil{}
				}
				return &object.Float{Value: f}
			},
		},
	}
}

func toFloat64(o object.Object) float64 {
	switch v := o.(type) {
	case *object.Integer:
		return float64(v.Value)
	case *object.Float:
		return v.Value
	}
	return 0
}

func mathSqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x / 2
	for range 100 {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func mathPow(base, exp float64) float64 {
	result := 1.0
	for exp > 0 {
		if int(exp)%2 == 0 {
			result *= base
		}
		base *= base
		exp = float64(int(exp) / 2)
	}
	return result
}

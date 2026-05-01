package compiler

import (
	"bufio"
	"os"
	"strings"

	"github.com/hilthontt/lotus/object"
)

func envPackage() *object.Package {
	return &object.Package{
		Name: "Env",
		Functions: map[string]object.PackageFunction{

			// Env.get(key) -> string | nil
			// Returns the value of an environment variable, or nil if not set.
			"get": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				key, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				val, exists := os.LookupEnv(key.Value)
				if !exists {
					return &object.Nil{}
				}
				return &object.String{Value: val}
			},

			// Env.getOr(key, default) -> string
			// Returns the env var, or default if not set.
			"getOr": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Nil{}
				}
				key, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				val, exists := os.LookupEnv(key.Value)
				if !exists {
					return args[1]
				}
				return &object.String{Value: val}
			},

			// Env.require(key) -> string
			// Returns the env var or throws if not set.
			"require": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				key, ok := args[0].(*object.String)
				if !ok {
					return &object.Nil{}
				}
				val, exists := os.LookupEnv(key.Value)
				if !exists {
					return &object.LotusError{
						Message: "required environment variable " + key.Value + " is not set",
					}
				}
				return &object.String{Value: val}
			},

			// Env.set(key, value) -> bool
			// Sets an environment variable for the current process.
			"set": func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Boolean{Value: false}
				}
				key, ok1 := args[0].(*object.String)
				val, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return &object.Boolean{Value: false}
				}
				err := os.Setenv(key.Value, val.Value)
				return &object.Boolean{Value: err == nil}
			},

			// Env.unset(key) -> bool
			"unset": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				key, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				return &object.Boolean{Value: os.Unsetenv(key.Value) == nil}
			},

			// Env.has(key) -> bool
			"has": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Boolean{Value: false}
				}
				key, ok := args[0].(*object.String)
				if !ok {
					return &object.Boolean{Value: false}
				}
				_, exists := os.LookupEnv(key.Value)
				return &object.Boolean{Value: exists}
			},

			// Env.all() -> map
			// Returns all environment variables as a map.
			"all": func(args ...object.Object) object.Object {
				pairs := make(map[object.HashKey]object.HashPair)
				for _, env := range os.Environ() {
					parts := strings.SplitN(env, "=", 2)
					if len(parts) != 2 {
						continue
					}
					key := &object.String{Value: parts[0]}
					val := &object.String{Value: parts[1]}
					pairs[key.HashKey()] = object.HashPair{Key: key, Value: val}
				}
				return &object.Hash{Pairs: pairs}
			},

			// Env.load(path) -> bool
			// Parses a .env file and sets variables into the current environment.
			// Skips blank lines and lines starting with #.
			// Does NOT override existing variables (same behaviour as dotenv).
			"load": func(args ...object.Object) object.Object {
				path := ".env"
				if len(args) >= 1 {
					if p, ok := args[0].(*object.String); ok {
						path = p.Value
					}
				}

				f, err := os.Open(path)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				defer f.Close()

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())

					// Skip blank lines and comments
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}

					// Strip optional "export " prefix
					line = strings.TrimPrefix(line, "export ")

					parts := strings.SplitN(line, "=", 2)
					if len(parts) != 2 {
						continue
					}

					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])

					// Strip surrounding quotes from value
					if len(val) >= 2 {
						if (val[0] == '"' && val[len(val)-1] == '"') ||
							(val[0] == '\'' && val[len(val)-1] == '\'') {
							val = val[1 : len(val)-1]
						}
					}

					// Don't override existing variables
					if _, exists := os.LookupEnv(key); !exists {
						os.Setenv(key, val)
					}
				}
				return &object.Boolean{Value: true}
			},

			// Env.loadOverride(path) -> bool
			// Same as load() but DOES override existing variables.
			"loadOverride": func(args ...object.Object) object.Object {
				path := ".env"
				if len(args) >= 1 {
					if p, ok := args[0].(*object.String); ok {
						path = p.Value
					}
				}

				f, err := os.Open(path)
				if err != nil {
					return &object.Boolean{Value: false}
				}
				defer f.Close()

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					line = strings.TrimPrefix(line, "export ")
					parts := strings.SplitN(line, "=", 2)
					if len(parts) != 2 {
						continue
					}
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if len(val) >= 2 {
						if (val[0] == '"' && val[len(val)-1] == '"') ||
							(val[0] == '\'' && val[len(val)-1] == '\'') {
							val = val[1 : len(val)-1]
						}
					}
					os.Setenv(key, val)
				}
				return &object.Boolean{Value: true}
			},
		},
	}
}

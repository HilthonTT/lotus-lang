package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	kindMethod   = 2
	kindFunction = 3
	kindVariable = 6
	kindModule   = 9
	kindKeyword  = 14
	kindField    = 5
	kindClass    = 7
)

type Analyzer struct {
	keywords  []string
	typeNames []string
	builtins  []builtinDoc
	packages  map[string][]packageMember
}

type builtinDoc struct {
	name, signature, doc string
}

type packageMember struct {
	name, signature, doc string
}

// classInfo holds extracted fields and methods for a class.
type classInfo struct {
	fields  []string
	methods []string
}

var (
	reLetMut    = regexp.MustCompile(`\b(?:let|mut)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	reFn        = regexp.MustCompile(`\bfn\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	reClass     = regexp.MustCompile(`\bclass\s+([A-Z][a-zA-Z0-9_]*)`)
	reVarClass  = regexp.MustCompile(`\b(?:let|mut)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[a-zA-Z_][a-zA-Z0-9_]*)?\s*=\s*([A-Z][a-zA-Z0-9_]*)\s*\(`)
	reSelfField = regexp.MustCompile(`\bself\.([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	reMethod    = regexp.MustCompile(`\bfn\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(\s*self`)
)

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		keywords: []string{
			"let", "mut", "fn", "class", "extends", "if", "else",
			"while", "for", "in", "return", "break", "continue",
			"true", "false", "nil", "self", "super",
			"import", "export", "from",
		},
		typeNames: []string{
			"int", "float", "string", "bool", "array", "map", "nil",
		},
		builtins: []builtinDoc{
			{"print", "print(...values)", "Prints values to stdout separated by spaces."},
			{"len", "len(value) -> int", "Returns the length of a string, array, or map."},
			{"push", "push(array, value) -> array", "Returns a new array with value appended."},
			{"pop", "pop(array) -> value", "Returns the last element of the array."},
			{"head", "head(array) -> value", "Returns the first element of the array."},
			{"tail", "tail(array) -> array", "Returns all elements except the first."},
			{"type", "type(value) -> string", "Returns the type of a value as a string."},
			{"str", "str(value) -> string", "Converts a value to its string representation."},
			{"int", "int(value) -> int", "Converts a value to an integer."},
			{"range", "range([start,] end [,step]) -> array", "Returns an array of integers."},
		},
		packages: map[string][]packageMember{
			"Console": {
				{"readLine", "Console.readLine() -> string", "Reads a line from stdin."},
				{"prompt", "Console.prompt(message) -> string", "Prints message then reads a line from stdin."},
				{"clear", "Console.clear()", "Clears the terminal screen."},
				{"print", "Console.print(...values)", "Prints values to stdout."},
				{"printErr", "Console.printErr(...values)", "Prints values to stderr."},
			},
			"Math": {
				{"sqrt", "Math.sqrt(x) -> float", "Returns the square root of x."},
				{"abs", "Math.abs(x) -> number", "Returns the absolute value of x."},
				{"floor", "Math.floor(x) -> int", "Returns the floor of x as an integer."},
				{"pow", "Math.pow(base, exp) -> float", "Returns base raised to exp."},
				{"max", "Math.max(a, b) -> number", "Returns the larger of two values."},
				{"min", "Math.min(a, b) -> number", "Returns the smaller of two values."},
				{"pi", "Math.pi() -> float", "Returns π (3.141592653589793)."},
			},
			"OS": {
				{"exit", "OS.exit([code])", "Exits the process with an optional exit code."},
				{"args", "OS.args() -> array", "Returns command-line arguments as an array."},
				{"env", "OS.env(key) -> string | nil", "Returns an environment variable by key."},
				{"readFile", "OS.readFile(path) -> string | nil", "Reads a file and returns its contents as a string."},
				{"writeFile", "OS.writeFile(path, content) -> bool", "Writes a string to a file. Returns true on success."},
				{"parseInt", "OS.parseInt(s) -> int | nil", "Parses a string to an integer."},
				{"parseFloat", "OS.parseFloat(s) -> float | nil", "Parses a string to a float."},
			},
			"Task": {
				{"spawn", "Task.spawn(fn())", "Runs a zero-argument Lotus closure in a new goroutine."},
				{"spawnWith", "Task.spawnWith(fn(arg), arg)", "Runs a Lotus closure in a new goroutine, passing one argument."},
				{"wait", "Task.wait()", "Blocks until all spawned tasks have finished."},
				{"sleep", "Task.sleep(ms: int)", "Pauses the current task for the given number of milliseconds."},
				{"mutex", "Task.mutex() -> Mutex", "Creates and returns a new mutex object."},
			},
		},
	}
}

// extractClasses parses class bodies from source and returns a map of
// className -> classInfo (fields set via self.x = ..., and methods).
func extractClasses(source string) map[string]classInfo {
	classes := map[string]classInfo{}

	// Find each class block: class Foo { ... }
	reClassBlock := regexp.MustCompile(`(?s)\bclass\s+([A-Z][a-zA-Z0-9_]*)[^{]*\{(.*?)\n\}`)
	for _, m := range reClassBlock.FindAllStringSubmatch(source, -1) {
		name := m[1]
		body := m[2]

		seen := map[string]bool{}
		info := classInfo{}

		for _, fm := range reSelfField.FindAllStringSubmatch(body, -1) {
			field := fm[1]
			if !seen[field] {
				seen[field] = true
				info.fields = append(info.fields, field)
			}
		}
		for _, mm := range reMethod.FindAllStringSubmatch(body, -1) {
			method := mm[1]
			if method != "init" {
				info.methods = append(info.methods, method)
			}
		}
		classes[name] = info
	}
	return classes
}

// extractVarTypes returns a map of varName -> className for lines like:
// let v = Vector(3.0, 4.0)
func extractVarTypes(source string) map[string]string {
	varTypes := map[string]string{}
	for _, m := range reVarClass.FindAllStringSubmatch(source, -1) {
		varTypes[m[1]] = m[2]
	}
	return varTypes
}

func (a *Analyzer) Complete(source, prefix, receiver string) []CompletionItem {
	var items []CompletionItem

	if receiver != "" {
		// Built-in packages
		if members, ok := a.packages[receiver]; ok {
			for _, m := range members {
				m := m
				items = append(items, CompletionItem{
					Label:         m.name,
					Kind:          new(kindMethod),
					Detail:        new(m.signature),
					Documentation: &MarkupContent{Kind: "markdown", Value: m.doc},
				})
			}
			return items
		}

		// Instance fields and methods
		varTypes := extractVarTypes(source)
		classes := extractClasses(source)
		className, ok := varTypes[receiver]
		if !ok {
			className = receiver
		}
		if info, ok := classes[className]; ok {
			for _, field := range info.fields {
				items = append(items, CompletionItem{
					Label:  field,
					Kind:   new(kindField),
					Detail: new(className + "." + field),
				})
			}
			for _, method := range info.methods {
				items = append(items, CompletionItem{
					Label:  method,
					Kind:   new(kindMethod),
					Detail: new(className + "." + method + "(self, ...)"),
				})
			}
			return items
		}

		return items
	}

	// Keywords
	for _, kw := range a.keywords {
		if strings.HasPrefix(kw, prefix) {
			kw := kw
			items = append(items, CompletionItem{Label: kw, Kind: new(kindKeyword)})
		}
	}

	// Type names (shown as keywords in completion)
	for _, t := range a.typeNames {
		if strings.HasPrefix(t, prefix) {
			t := t
			items = append(items, CompletionItem{
				Label:         t,
				Kind:          new(kindKeyword),
				Detail:        new("type"),
				Documentation: &MarkupContent{Kind: "markdown", Value: fmt.Sprintf("Built-in type: `%s`", t)},
			})
		}
	}

	// Builtins
	for _, b := range a.builtins {
		if strings.HasPrefix(b.name, prefix) {
			b := b
			items = append(items, CompletionItem{
				Label:         b.name,
				Kind:          new(kindFunction),
				Detail:        new(b.signature),
				Documentation: &MarkupContent{Kind: "markdown", Value: b.doc},
			})
		}
	}

	// Packages
	for name := range a.packages {
		if strings.HasPrefix(name, prefix) {
			name := name
			items = append(items, CompletionItem{Label: name, Kind: new(kindModule)})
		}
	}

	// User symbols
	for _, name := range extractUserSymbols(source) {
		if strings.HasPrefix(name, prefix) && name != prefix {
			name := name
			items = append(items, CompletionItem{Label: name, Kind: new(kindVariable)})
		}
	}

	return items
}

func (a *Analyzer) HoverDoc(word string) string {
	// Builtins
	for _, b := range a.builtins {
		if b.name == word {
			return fmt.Sprintf("```\n%s\n```\n\n%s", b.signature, b.doc)
		}
	}

	// Packages
	if members, ok := a.packages[word]; ok {
		var sb strings.Builder
		fmt.Fprintf(&sb, "**%s** — built-in package\n\n", word)
		sb.WriteString("| Member | Signature |\n|--------|----------|\n")
		for _, m := range members {
			fmt.Fprintf(&sb, "| `%s` | `%s` |\n", m.name, m.signature)
		}
		return sb.String()
	}

	// Type names
	typeDoc := map[string]string{
		"int":    "Built-in integer type. Example: `let x: int = 42`",
		"float":  "Built-in float type. Example: `let x: float = 3.14`",
		"string": "Built-in string type. Example: `let s: string = \"hello\"`",
		"bool":   "Built-in boolean type. Values: `true` or `false`",
		"array":  "Built-in array type. Example: `let a: array = [1, 2, 3]`",
		"map":    "Built-in map type. Example: `let m: map = {\"key\": \"value\"}`",
	}
	if doc, ok := typeDoc[word]; ok {
		return doc
	}

	return ""
}

func extractUserSymbols(source string) []string {
	seen := map[string]bool{}
	var result []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	for _, m := range reLetMut.FindAllStringSubmatch(source, -1) {
		add(m[1])
	}
	for _, m := range reFn.FindAllStringSubmatch(source, -1) {
		add(m[1])
	}
	for _, m := range reClass.FindAllStringSubmatch(source, -1) {
		add(m[1])
	}
	return result
}

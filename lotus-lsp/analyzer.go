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
			"import", "export", "from", "match", "enum",
			"defer", "try", "catch", "throw", "interface",
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
			{"implements", "implements(obj, iface) -> bool", "Returns true if obj implements all methods of iface."},
			{"ok", "ok(value) -> Result", "Creates a successful Result wrapping value."},
			{"err", "err(message) -> Result", "Creates a failed Result with the given message."},
			{"isOk", "isOk(result) -> bool", "Returns true if the Result is successful."},
			{"isErr", "isErr(result) -> bool", "Returns true if the Result is an error."},
			{"unwrap", "unwrap(result) -> value", "Returns the value inside ok(), or throws if err()."},
			{"unwrapOr", "unwrapOr(result, default) -> value", "Returns value if ok(), otherwise returns default."},
		},
		packages: map[string][]packageMember{
			"Console": {
				{"readLine", "Console.readLine() -> string", "Reads a line from stdin."},
				{"prompt", "Console.prompt(message) -> string", "Prints message then reads a line."},
				{"clear", "Console.clear()", "Clears the terminal screen."},
				{"print", "Console.print(...values)", "Prints values to stdout."},
				{"printErr", "Console.printErr(...values)", "Prints values to stderr."},
			},
			"Math": {
				{"sqrt", "Math.sqrt(x) -> float", "Returns the square root of x."},
				{"abs", "Math.abs(x) -> number", "Returns the absolute value of x."},
				{"floor", "Math.floor(x) -> int", "Returns the floor of x."},
				{"ceil", "Math.ceil(x) -> int", "Returns the ceiling of x."},
				{"round", "Math.round(x) -> int", "Rounds x to the nearest integer."},
				{"pow", "Math.pow(base, exp) -> float", "Returns base raised to exp."},
				{"max", "Math.max(a, b) -> number", "Returns the larger of two values."},
				{"min", "Math.min(a, b) -> number", "Returns the smaller of two values."},
				{"pi", "Math.pi() -> float", "Returns π."},
				{"e", "Math.e() -> float", "Returns e."},
				{"log", "Math.log(x) -> float", "Returns the natural logarithm of x."},
				{"log2", "Math.log2(x) -> float", "Returns log base 2 of x."},
				{"log10", "Math.log10(x) -> float", "Returns log base 10 of x."},
				{"sin", "Math.sin(x) -> float", "Returns the sine of x (radians)."},
				{"cos", "Math.cos(x) -> float", "Returns the cosine of x (radians)."},
				{"tan", "Math.tan(x) -> float", "Returns the tangent of x (radians)."},
				{"asin", "Math.asin(x) -> float", "Returns the arcsine of x."},
				{"acos", "Math.acos(x) -> float", "Returns the arccosine of x."},
				{"atan", "Math.atan(x) -> float", "Returns the arctangent of x."},
				{"atan2", "Math.atan2(y, x) -> float", "Returns the angle in radians between the positive x-axis and (x,y)."},
				{"random", "Math.random() -> float", "Returns a random float between 0.0 and 1.0."},
				{"randomInt", "Math.randomInt(min, max) -> int", "Returns a random integer in [min, max]."},
				{"clamp", "Math.clamp(x, min, max) -> number", "Clamps x to the range [min, max]."},
				{"isNaN", "Math.isNaN(x) -> bool", "Returns true if x is NaN."},
				{"isInf", "Math.isInf(x) -> bool", "Returns true if x is infinite."},
				{"hypot", "Math.hypot(a, b) -> float", "Returns sqrt(a² + b²)."},
				{"degrees", "Math.degrees(radians) -> float", "Converts radians to degrees."},
				{"radians", "Math.radians(degrees) -> float", "Converts degrees to radians."},
				{"gcd", "Math.gcd(a, b) -> int", "Returns the greatest common divisor."},
				{"lcm", "Math.lcm(a, b) -> int", "Returns the least common multiple."},
				{"inf", "Math.inf() -> float", "Returns positive infinity."},
				{"nan", "Math.nan() -> float", "Returns NaN."},
			},
			"OS": {
				{"exit", "OS.exit([code])", "Exits the process."},
				{"args", "OS.args() -> array", "Returns command-line arguments."},
				{"env", "OS.env(key) -> string | nil", "Returns an environment variable."},
				{"readFile", "OS.readFile(path) -> string | nil", "Reads a file as a string."},
				{"writeFile", "OS.writeFile(path, content) -> bool", "Writes a string to a file."},
				{"parseInt", "OS.parseInt(s) -> int | nil", "Parses a string to an integer."},
				{"parseFloat", "OS.parseFloat(s) -> float | nil", "Parses a string to a float."},
			},
			"Task": {
				{"spawn", "Task.spawn(fn())", "Runs a closure in a new goroutine."},
				{"spawnWith", "Task.spawnWith(fn(arg), arg)", "Runs a closure with one argument in a goroutine."},
				{"wait", "Task.wait()", "Blocks until all tasks finish."},
				{"sleep", "Task.sleep(ms: int)", "Pauses the current task."},
				{"mutex", "Task.mutex() -> Mutex", "Creates a new mutex."},
			},
			"String": {
				{"split", "String.split(str, sep) -> array", "Splits by separator."},
				{"trim", "String.trim(str) -> string", "Removes leading/trailing whitespace."},
				{"trimLeft", "String.trimLeft(str) -> string", "Removes leading whitespace."},
				{"trimRight", "String.trimRight(str) -> string", "Removes trailing whitespace."},
				{"trimPrefix", "String.trimPrefix(str, prefix) -> string", "Removes a prefix if present."},
				{"trimSuffix", "String.trimSuffix(str, suffix) -> string", "Removes a suffix if present."},
				{"upper", "String.upper(str) -> string", "Converts to uppercase."},
				{"lower", "String.lower(str) -> string", "Converts to lowercase."},
				{"title", "String.title(str) -> string", "Converts to Title Case."},
				{"replace", "String.replace(str, old, new) -> string", "Replaces all occurrences."},
				{"contains", "String.contains(str, substr) -> bool", "Returns true if str contains substr."},
				{"startsWith", "String.startsWith(str, prefix) -> bool", "Returns true if str starts with prefix."},
				{"endsWith", "String.endsWith(str, suffix) -> bool", "Returns true if str ends with suffix."},
				{"indexOf", "String.indexOf(str, substr) -> int", "Returns index of first match, or -1."},
				{"lastIndexOf", "String.lastIndexOf(str, substr) -> int", "Returns index of last match, or -1."},
				{"count", "String.count(str, substr) -> int", "Counts occurrences of substr."},
				{"repeat", "String.repeat(str, n) -> string", "Repeats str n times."},
				{"padLeft", "String.padLeft(str, n, char) -> string", "Left-pads to length n."},
				{"padRight", "String.padRight(str, n, char) -> string", "Right-pads to length n."},
				{"chars", "String.chars(str) -> array", "Returns an array of characters."},
				{"len", "String.len(str) -> int", "Returns character count."},
				{"join", "String.join(array, sep) -> string", "Joins array with separator."},
				{"slice", "String.slice(str, start, end) -> string", "Returns substring."},
				{"reverse", "String.reverse(str) -> string", "Reverses the string."},
				{"lines", "String.lines(str) -> array", "Splits by newline."},
				{"toBytes", "String.toBytes(str) -> array", "Returns byte values as int array."},
				{"fromBytes", "String.fromBytes(arr) -> string", "Builds a string from byte values."},
				{"format", "String.format(template, ...args) -> string", "Substitutes %s %d %f in template."},
				{"isDigit", "String.isDigit(str) -> bool", "Returns true if all chars are digits."},
				{"isAlpha", "String.isAlpha(str) -> bool", "Returns true if all chars are letters."},
				{"isAlphaNum", "String.isAlphaNum(str) -> bool", "Returns true if all chars are letters or digits."},
			},
			"Array": {
				{"filter", "Array.filter(arr, fn(elem) -> bool) -> array", "Returns matching elements."},
				{"map", "Array.map(arr, fn(elem) -> any) -> array", "Transforms each element."},
				{"reduce", "Array.reduce(arr, fn(acc, elem) -> any, initial) -> any", "Reduces to a single value."},
				{"find", "Array.find(arr, fn(elem) -> bool) -> elem | nil", "Returns first matching element."},
				{"findIndex", "Array.findIndex(arr, fn(elem) -> bool) -> int", "Returns index of first match."},
				{"forEach", "Array.forEach(arr, fn(elem))", "Calls fn for each element."},
				{"contains", "Array.contains(arr, value) -> bool", "Returns true if value is in array."},
				{"reverse", "Array.reverse(arr) -> array", "Returns reversed copy."},
				{"sort", "Array.sort(arr) -> array", "Returns sorted copy."},
				{"sortBy", "Array.sortBy(arr, fn(elem) -> comparable) -> array", "Sorts by key function."},
				{"flat", "Array.flat(arr) -> array", "Flattens one level of nesting."},
				{"join", "Array.join(arr, sep) -> string", "Joins elements into a string."},
				{"slice", "Array.slice(arr, start, end) -> array", "Returns a sub-array."},
				{"unique", "Array.unique(arr) -> array", "Removes duplicates."},
				{"len", "Array.len(arr) -> int", "Returns length."},
				{"any", "Array.any(arr, fn(elem) -> bool) -> bool", "Returns true if any element matches."},
				{"all", "Array.all(arr, fn(elem) -> bool) -> bool", "Returns true if all elements match."},
			},
			"Time": {
				{"now", "Time.now() -> int", "Returns current Unix milliseconds."},
				{"sleep", "Time.sleep(ms: int)", "Pauses for ms milliseconds."},
				{"format", "Time.format(ms: int, layout: string) -> string", "Formats a timestamp."},
				{"parse", "Time.parse(str: string, layout: string) -> int", "Parses a time string."},
				{"since", "Time.since(ms: int) -> int", "Milliseconds elapsed."},
				{"until", "Time.until(ms: int) -> int", "Milliseconds until timestamp."},
				{"add", "Time.add(ms: int, duration: int) -> int", "Adds duration to timestamp."},
				{"diff", "Time.diff(a: int, b: int) -> int", "Returns a - b in milliseconds."},
				{"year", "Time.year(ms: int) -> int", "Extracts year."},
				{"month", "Time.month(ms: int) -> int", "Extracts month (1-12)."},
				{"day", "Time.day(ms: int) -> int", "Extracts day (1-31)."},
				{"hour", "Time.hour(ms: int) -> int", "Extracts hour (0-23)."},
				{"minute", "Time.minute(ms: int) -> int", "Extracts minute (0-59)."},
				{"second", "Time.second(ms: int) -> int", "Extracts second (0-59)."},
				{"weekday", "Time.weekday(ms: int) -> string", "Returns weekday name."},
				{"unix", "Time.unix(ms: int) -> int", "Converts ms to seconds."},
				{"fromUnix", "Time.fromUnix(sec: int) -> int", "Converts seconds to ms."},
				{"isBefore", "Time.isBefore(a: int, b: int) -> bool", "Returns true if a < b."},
				{"isAfter", "Time.isAfter(a: int, b: int) -> bool", "Returns true if a > b."},
				{"startOfDay", "Time.startOfDay(ms: int) -> int", "Returns midnight timestamp."},
				{"endOfDay", "Time.endOfDay(ms: int) -> int", "Returns 23:59:59.999 timestamp."},
				{"addDays", "Time.addDays(ms: int, days: int) -> int", "Adds calendar days."},
				{"addMonths", "Time.addMonths(ms: int, months: int) -> int", "Adds calendar months."},
				{"addYears", "Time.addYears(ms: int, years: int) -> int", "Adds calendar years."},
				{"duration", "Time.duration(ms: int) -> string", "Returns human-readable duration."},
				{"seconds", "Time.seconds(n: int) -> int", "n seconds in ms."},
				{"minutes", "Time.minutes(n: int) -> int", "n minutes in ms."},
				{"hours", "Time.hours(n: int) -> int", "n hours in ms."},
				{"days", "Time.days(n: int) -> int", "n days in ms."},
				{"utc", "Time.utc(ms: int) -> int", "Converts to UTC."},
				{"timezone", "Time.timezone() -> string", "Returns local timezone name."},
			},
			"Json": {
				{"stringify", "Json.stringify(value) -> string", "Converts to JSON string."},
				{"prettyPrint", "Json.prettyPrint(value) -> string", "Converts to indented JSON."},
				{"parse", "Json.parse(str: string) -> value", "Parses JSON string."},
				{"valid", "Json.valid(str: string) -> bool", "Returns true if valid JSON."},
				{"keys", "Json.keys(str: string) -> array", "Returns top-level keys."},
				{"get", "Json.get(str: string, key: string) -> value", "Gets a key from JSON object."},
				{"set", "Json.set(str: string, key: string, value) -> string", "Sets a key, returns new JSON."},
				{"merge", "Json.merge(a: string, b: string) -> string", "Merges two JSON objects."},
			},
			"File": {
				{"exists", "File.exists(path) -> bool", "Returns true if path exists."},
				{"read", "File.read(path) -> string | nil", "Reads file contents as string."},
				{"write", "File.write(path, content) -> bool", "Writes string to file."},
				{"append", "File.append(path, content) -> bool", "Appends string to file."},
				{"delete", "File.delete(path) -> bool", "Deletes a file."},
				{"rename", "File.rename(old, new) -> bool", "Renames a file."},
				{"copy", "File.copy(src, dst) -> bool", "Copies a file."},
				{"mkdir", "File.mkdir(path) -> bool", "Creates directory (and parents)."},
				{"listDir", "File.listDir(path) -> array", "Lists directory entries."},
				{"isDir", "File.isDir(path) -> bool", "Returns true if path is a directory."},
				{"isFile", "File.isFile(path) -> bool", "Returns true if path is a file."},
				{"size", "File.size(path) -> int", "Returns file size in bytes."},
				{"extension", "File.extension(path) -> string", "Returns file extension e.g. \".go\"."},
				{"basename", "File.basename(path) -> string", "Returns filename without directory."},
				{"dirname", "File.dirname(path) -> string", "Returns directory of path."},
				{"stem", "File.stem(path) -> string", "Returns filename without extension."},
				{"join", "File.join(...parts) -> string", "Joins path components."},
				{"abs", "File.abs(path) -> string", "Returns absolute path."},
				{"cwd", "File.cwd() -> string", "Returns current working directory."},
				{"glob", "File.glob(pattern) -> array", "Returns paths matching a glob pattern."},
			},
			"Regex": {
				{"compile", "Regex.compile(pattern) -> Regex", "Compiles a regex pattern for reuse."},
				{"test", "Regex.test(pattern, str) -> bool", "Returns true if pattern matches str."},
				{"find", "Regex.find(pattern, str) -> string | nil", "Returns first match."},
				{"findAll", "Regex.findAll(pattern, str) -> array", "Returns all matches."},
				{"replace", "Regex.replace(pattern, str, replacement) -> string", "Replaces all matches."},
				{"replaceFirst", "Regex.replaceFirst(pattern, str, replacement) -> string", "Replaces first match only."},
				{"split", "Regex.split(pattern, str) -> array", "Splits string by pattern."},
				{"groups", "Regex.groups(pattern, str) -> array", "Returns captured groups from first match."},
				{"groupsAll", "Regex.groupsAll(pattern, str) -> array", "Returns captured groups from all matches."},
				{"escape", "Regex.escape(str) -> string", "Escapes special regex characters."},
				{"count", "Regex.count(pattern, str) -> int", "Returns number of matches."},
			},
			"HttpClient": {
				{"get", "HttpClient.get(url, headers?) -> Response", "HTTP GET request."},
				{"post", "HttpClient.post(url, body, headers?) -> Response", "HTTP POST request."},
				{"put", "HttpClient.put(url, body, headers?) -> Response", "HTTP PUT request."},
				{"patch", "HttpClient.patch(url, body, headers?) -> Response", "HTTP PATCH request."},
				{"delete", "HttpClient.delete(url, headers?) -> Response", "HTTP DELETE request."},
				{"request", "HttpClient.request(method, url, body?, headers?) -> Response", "Generic HTTP request."},
				{"json", "HttpClient.json(response) -> value", "Parse response body as JSON."},
				{"setTimeout", "HttpClient.setTimeout(ms: int)", "Set request timeout in milliseconds."},
			},
		},
	}
}

// extractClasses parses class bodies from source and returns a map of
// className -> classInfo (fields set via self.x = ..., and methods).
func extractClasses(source string) map[string]classInfo {
	classes := map[string]classInfo{}
	reClassBlock := regexp.MustCompile(`(?s)\bclass\s+([A-Z][a-zA-Z0-9_]*)[^{]*\{(.*?)\n\}`)
	for _, m := range reClassBlock.FindAllStringSubmatch(source, -1) {
		name, body := m[1], m[2]
		seen := map[string]bool{}
		info := classInfo{}
		for _, fm := range reSelfField.FindAllStringSubmatch(body, -1) {
			if !seen[fm[1]] {
				seen[fm[1]] = true
				info.fields = append(info.fields, fm[1])
			}
		}
		for _, mm := range reMethod.FindAllStringSubmatch(body, -1) {
			if mm[1] != "init" {
				info.methods = append(info.methods, mm[1])
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

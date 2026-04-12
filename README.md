# 🪷 Lotus

A compiled, expression-oriented programming language with a stack-based virtual machine, written in Go.

```
let greeting = "Hello from Lotus!"
print(greeting)
```

---

## Architecture

Lotus follows a classic compiler pipeline — source code is lexed into tokens, parsed into an AST via a Pratt parser, compiled to bytecode, and executed on a stack-based VM with call frames and closure support.

```
Source → Lexer → Parser → Compiler → Bytecode → VM → Result
                   │                     │
              Pratt Parser          30+ Opcodes
              10 Precedence         Stack-based
              Levels                Call Frames
                                    Closures
```

| Component    | Package      | Description                                           |
| ------------ | ------------ | ----------------------------------------------------- |
| Token        | `token/`     | 40+ token types — keywords, operators, delimiters     |
| Lexer        | `lexer/`     | UTF-8 scanner with escape sequences and `//` comments |
| AST          | `ast/`       | 20+ node types for statements and expressions         |
| Parser       | `parser/`    | Pratt parser with 10 precedence levels                |
| Symbol Table | `compiler/`  | Scoped resolution with free variable capture          |
| Compiler     | `compiler/`  | Single-pass bytecode emitter                          |
| Opcodes      | `code/`      | Bytecode encoding/decoding with disassembler          |
| VM           | `vm/`        | Stack-based VM with call frames and closures          |
| Object       | `object/`    | Runtime value types with method dispatch              |
| Builtins     | `compiler/`  | Built-in functions shared by compiler and evaluator   |
| Evaluator    | `evaluator/` | Tree-walking interpreter (alternative execution mode) |
| REPL         | `repl/`      | Interactive read-eval-print loop                      |
| LSP          | `lotus-lsp/` | Language server (completions, hover, diagnostics)     |
| VS Code      | `vscode/`    | VS Code extension with syntax highlighting and LSP    |
| Version      | `version/`   | Build metadata (version, commit, build time)          |

## Quick Start

```bash
# Build
go build -o lotus .

# Run a file (VM engine, default)
./lotus examples/example.lotus

# Run a file with the tree-walking evaluator
./lotus --engine eval examples/example.lotus

# Start the interactive REPL (VM engine, default)
./lotus --console

# Start the web playground
./lotus --playground

# Start the REPL with the evaluator
./lotus --console --engine eval

# Disassemble a file (plain)
./lotus --dis examples/example.lotus

# Disassemble with inline opcode comments
./lotus --dis --annotated examples/example.lotus

# Print version info
./lotus --version

# Show help
./lotus --help

# Run tests
go test -v ./...
```

## VS Code Extension

The `vscode/` directory contains a VS Code extension that provides:

- **Syntax highlighting** — keywords, types, functions, operators, strings, comments
- **Autocomplete** — keywords, builtins, package members (`Math.`, `Console.`, `OS.`), user-defined symbols and class members
- **Hover documentation** — inline docs for built-in functions and packages
- **Bracket matching and auto-closing** — for `{}`, `[]`, `()`, `""`
- **Comment toggling** — `//` line comments

### Installing the Extension

The extension requires [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) on Windows, as the LSP server is compiled as a Linux binary.

**1. Build the LSP binary:**

```bash
cd lotus-lsp
GOOS=linux GOARCH=amd64 go build -o ../vscode/bin/lotus-lsp-linux .
```

**2. Build and package the extension:**

```bash
cd vscode
npm install
npm run package        # bundles + runs vsce package
```

**3. Install in VS Code:**

```bash
code --install-extension lotus-lang-0.1.0.vsix
```

Or: **Ctrl+Shift+P** → "Extensions: Install from VSIX" → select the `.vsix` file.

### LSP Features

The language server (`lotus-lsp/`) is built in Go and communicates with VS Code over stdio using the [JSON-RPC 2.0](https://www.jsonrpc.org/specification) protocol.

| Feature            | Details                                                        |
| ------------------ | -------------------------------------------------------------- |
| Completions        | Keywords, builtins, packages, user symbols, class members      |
| Dot completions    | `Math.`, `Console.`, `OS.`, `v.` (instance fields and methods) |
| Hover docs         | Signatures and descriptions for all builtins and packages      |
| Document sync      | Full document sync on open and change                          |
| Trigger characters | `.` triggers member completions                                |

### Extension Structure

```
vscode/
├── bin/                    ← compiled LSP binary (git-ignored)
│   └── lotus-lsp-linux
├── out/                    ← compiled extension JS (git-ignored)
│   └── extension.js
├── src/
│   └── extension.ts        ← extension entry point
├── syntaxes/
│   └── lotus-tmLanguage.json
├── language-configuration.json
├── package.json
└── tsconfig.json
```

---

## Language Reference

### Variables

```rust
let name = "Lotus"       // immutable binding
mut counter = 0          // mutable binding
counter = counter + 1    // reassignment (mut only)
```

Attempting to reassign a `let` binding is a runtime error.

### Types

| Type    | Examples                        |
| ------- | ------------------------------- |
| Integer | `42`, `-7`, `0`                 |
| Float   | `3.14`, `0.5`                   |
| String  | `"hello"`, `"line\nnewline"`    |
| Boolean | `true`, `false`                 |
| Nil     | `nil`                           |
| Array   | `[1, 2, 3]`, `["a", true, nil]` |
| Map     | `{"key": "value", "n": 42}`     |

### Operators

```rust
// Arithmetic
+  -  *  /  %

// Comparison
==  !=  <  >  <=  >=

// Logical (short-circuit)
&&  ||  !

// String concatenation
"hello" + " " + "world"
```

### Control Flow

```rust
// If expression (returns a value)
let max = if a > b { a } else { b }

// While loop
while condition {
    // ...
}

// For-in loop
for item in [1, 2, 3] {
    print(str(item))
}

// With range
for i in range(0, 10) {
    // ...
}

// Break and continue
while true {
    if done { break }
    if skip { continue }
}
```

### Functions

```rust
// Named function (supports recursion)
fn fibonacci(n) {
    if n <= 1 { return n }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

// Anonymous function
let double = fn(x) { x * 2 }

// Closures capture their environment
fn make_counter() {
    mut count = 0
    return fn() {
        count = count + 1
        return count
    }
}

// Higher-order functions
fn map(arr, f) {
    mut result = []
    for item in arr {
        result = push(result, f(item))
    }
    return result
}

let doubled = map([1, 2, 3], fn(x) { x * 2 })
// => [2, 4, 6]
```

### Classes

```rust
class Animal {
    fn init(self, name) {
        self.name = name
    }

    fn speak(self) {
        return self.name + " makes a sound."
    }
}

class Dog extends Animal {
    fn speak(self) {
        return self.name + " barks."
    }
}

let d = Dog("Rex")
print(d.speak())   // Rex barks.
```

### Modules

```rust
// math_utils.lotus
export fn add(a, b) { return a + b }
export let PI = 3.14159

// main.lotus
import { add, PI } from "math_utils.lotus"
print(str(add(1, 2)))   // 3
print(str(PI))          // 3.14159
```

### Indexing

```rust
let arr = [10, 20, 30]
arr[0]     // 10
arr[-1]    // 30 (negative indexing)

let m = {"name": "Alice"}
m["name"]  // "Alice"

"hello"[1] // "e"

// Index assignment (mut arrays and maps only)
mut nums = [1, 2, 3]
nums[0] = 99
```

### Built-in Functions

| Function     | Description                                                                        |
| ------------ | ---------------------------------------------------------------------------------- |
| `print(...)` | Print values separated by spaces                                                   |
| `len(x)`     | Length of string, array, or map                                                    |
| `push(a, v)` | Return new array with value appended                                               |
| `pop(a)`     | Return last element of array                                                       |
| `head(a)`    | Return first element of array                                                      |
| `tail(a)`    | Return array without first element                                                 |
| `type(x)`    | Return type name as string                                                         |
| `str(x)`     | Convert value to string                                                            |
| `int(x)`     | Convert value to integer                                                           |
| `range(...)` | Generate integer array: `range(n)`, `range(start, end)`, `range(start, end, step)` |

### Built-in Packages

| Package   | Members                                                                  |
| --------- | ------------------------------------------------------------------------ |
| `Console` | `readLine`, `prompt`, `print`, `printErr`, `clear`                       |
| `Math`    | `sqrt`, `abs`, `floor`, `pow`, `max`, `min`, `pi`                        |
| `OS`      | `exit`, `args`, `env`, `readFile`, `writeFile`, `parseInt`, `parseFloat` |

```rust
let name = Console.prompt("Enter your name: ")
print("Hello, " + name + "!")

print(str(Math.sqrt(16.0)))   // 4.0
print(str(Math.pi()))         // 3.141592653589793

let content = OS.readFile("hello.txt")
OS.writeFile("out.txt", "Hello from Lotus!")
```

### Comments

```rust
// Single-line comments
let x = 42 // inline comment
```

---

## Example: Quicksort

```rust
fn quicksort(arr) {
    if len(arr) <= 1 { return arr }

    let pivot = arr[0]
    mut less = []
    mut greater = []

    for i in range(1, len(arr)) {
        if arr[i] <= pivot {
            less = push(less, arr[i])
        } else {
            greater = push(greater, arr[i])
        }
    }

    let sorted_less = quicksort(less)
    let sorted_greater = quicksort(greater)

    mut result = sorted_less
    result = push(result, pivot)
    for item in sorted_greater {
        result = push(result, item)
    }
    return result
}

let unsorted = [38, 27, 43, 3, 9, 82, 10]
print("Sorted:", str(quicksort(unsorted)))
// => Sorted: [3, 9, 10, 27, 38, 43, 82]
```

---

## Execution Engines

Lotus supports two execution modes, selectable via `--engine`:

| Engine | Flag            | Description                                                       |
| ------ | --------------- | ----------------------------------------------------------------- |
| VM     | `--engine vm`   | Compiles to bytecode and executes on the stack-based VM (default) |
| Eval   | `--engine eval` | Tree-walking interpreter — simpler, no compilation step           |

Both engines share the same lexer, parser, and built-in functions. The VM engine is faster and is the default for file execution and the REPL.

## Bytecode & Disassembly

Lotus compiles to a custom bytecode with 30+ opcodes. You can inspect the generated bytecode for any source file using the `--dis` flag.

```bash
./lotus --dis examples/example.lotus
./lotus --dis --annotated examples/example.lotus
```

**Plain disassembly** (`--dis`):

```
=== main ===
0000 OpConstant 0
0003 OpSetGlobal 0
0006 OpGetBuiltin 0
0008 OpGetGlobal 0
0011 OpCall 1
0013 OpPop
```

**Annotated disassembly** (`--dis --annotated`):

```
=== main ===
0000 OpConstant 0               // push constant from pool
0003 OpSetGlobal 0              // store global variable
0006 OpGetBuiltin 0             // load built-in function by index
0008 OpGetGlobal 0              // load global variable
0011 OpCall 1                   // call function with N arguments
0013 OpPop                      // discard top of stack
```

**VM design:**

- Stack size: 2048 slots
- Global slots: 65536
- Max call frames: 1024
- Closure capture: free variables resolved at compile time, copied at closure creation
- Value types: all values are boxed `object.Object` interface values
- Method dispatch: objects support `.InvokeMethod()` for built-in methods
- Iteration: arrays implement the `Iterable` interface (`Reset`, `Next`)
- Hashing: integers, strings, and booleans implement the `Hashable` interface for map keys

## License

MIT

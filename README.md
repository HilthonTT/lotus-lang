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
| Version      | `version/`   | Build metadata (version, commit, build time)          |

## Quick Start

```bash
# Build
go build -o lotus .

# Run a file (VM engine, default)
./lotus examples/showcase.lotus

# Run a file with the tree-walking evaluator
./lotus --engine eval examples/showcase.lotus

# Start the interactive REPL (VM engine, default)
./lotus --console

# Start the REPL with the evaluator
./lotus --console --engine eval

# Disassemble a file (plain)
./lotus --dis examples/showcase.lotus

# Disassemble with inline opcode comments
./lotus --dis --annotated examples/showcase.lotus

# Print version info
./lotus --version

# Show help
./lotus --help

# Run tests
go test -v ./...
```

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

### Method Dispatch

Objects in Lotus support built-in methods via `.methods()`:

```rust
let arr = [1, 2, 3]
arr.len()       // 3
arr.methods()   // ["len", "methods"]

let m = {"a": 1, "b": 2}
m.len()         // 2
m.keys()        // ["a", "b"]
m.values()      // [1, 2]
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

### Comments

```rust
// Single-line comments
let x = 42 // inline comment
```

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

## Execution Engines

Lotus supports two execution modes, selectable via `--engine`:

| Engine | Flag            | Description                                                       |
| ------ | --------------- | ----------------------------------------------------------------- |
| VM     | `--engine vm`   | Compiles to bytecode and executes on the stack-based VM (default) |
| Eval   | `--engine eval` | Tree-walking interpreter — simpler, no compilation step           |

Both engines share the same lexer, parser, and built-in functions. The VM engine is faster and is the default for file execution and the REPL.

## Bytecode & Disassembly

Lotus compiles to a custom bytecode with 30+ opcodes. You can inspect the generated bytecode for any source file using the `--dis` flag. Each compiled function is printed under its own header. Anonymous functions appear as `<fn:N>` where `N` is their index in the constant pool.

```bash
./lotus --dis examples/showcase.lotus
./lotus --dis --annotated examples/showcase.lotus
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

=== quicksort ===
0000 OpConstant 103
0003 OpGetBuiltin 1
0005 OpGetLocal 0
0007 OpCall 1
0009 OpGreaterEq
0010 OpJumpFalse 19
0013 OpGetLocal 0
0015 OpReturn
0016 OpJump 20
0019 OpNil
0020 OpPop
...
0213 OpGetLocal 9
0215 OpReturn

=== <fn:79> ===
0000 OpGetLocal 0
0002 OpConstant 78
0005 OpMul
0006 OpReturn
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

=== quicksort ===
0000 OpConstant 103             // push constant from pool
0003 OpGetBuiltin 1             // load built-in function by index
0005 OpGetLocal 0               // load local variable
0007 OpCall 1                   // call function with N arguments
0009 OpGreaterEq                // greater-than-or-equal comparison
0010 OpJumpFalse 19             // jump if falsy (pops)
0013 OpGetLocal 0               // load local variable
0015 OpReturn                   // return value from function
...
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

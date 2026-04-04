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

| Component    | Package      | Description                                          |
|--------------|--------------|------------------------------------------------------|
| Token        | `token/`     | 40+ token types — keywords, operators, delimiters    |
| Lexer        | `lexer/`     | UTF-8 scanner with escape sequences and `//` comments|
| AST          | `ast/`       | 20+ node types for statements and expressions        |
| Parser       | `parser/`    | Pratt parser with 10 precedence levels               |
| Symbol Table | `compiler/`  | Scoped resolution with free variable capture         |
| Compiler     | `compiler/`  | Single-pass bytecode emitter                         |
| Opcodes      | `code/`      | Bytecode encoding/decoding with disassembler         |
| VM           | `vm/`        | Stack-based VM with call frames and closures         |
| Object       | `object/`    | Runtime value types with method dispatch             |
| Builtins     | `compiler/`  | Built-in functions                                   |

## Quick Start

```bash
# Build
go build -o lotus .

# Run a file
./lotus run examples/showcase.lotus

# Start the REPL
./lotus

# Disassemble (show bytecode)
./lotus dis examples/showcase.lotus

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

Attempting to reassign a `let` binding is a compile-time error.

### Types

| Type    | Examples                         |
|---------|----------------------------------|
| Integer | `42`, `-7`, `0`                  |
| Float   | `3.14`, `0.5`                    |
| String  | `"hello"`, `"line\nnewline"`     |
| Boolean | `true`, `false`                  |
| Nil     | `nil`                            |
| Array   | `[1, 2, 3]`, `["a", true, nil]`  |
| Map     | `{"key": "value", "n": 42}`      |

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

// Index assignment (mut arrays and maps)
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

| Function      | Description                                       |
|---------------|---------------------------------------------------|
| `print(...)`  | Print values separated by spaces                  |
| `len(x)`      | Length of string, array, or map                   |
| `push(a, v)`  | Return new array with value appended              |
| `pop(a)`      | Return last element of array                      |
| `head(a)`     | Return first element of array                     |
| `tail(a)`     | Return array without first element                |
| `type(x)`     | Return type name as string                        |
| `str(x)`      | Convert value to string                           |
| `int(x)`      | Convert value to integer                          |
| `range(...)`  | Generate integer array — `range(n)`, `range(start, end)`, `range(start, end, step)` |

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

## Bytecode & VM

Lotus compiles to a custom bytecode with 30+ opcodes executed by a stack-based virtual machine.

**Disassembly example:**

```
$ echo 'let x = 1 + 2; print(str(x))' | ./lotus dis /dev/stdin

0000 OpConstant 0        // push 1
0003 OpConstant 1        // push 2
0006 OpAdd                // 1 + 2 = 3
0007 OpSetGlobal 0       // x = 3
0010 OpGetBuiltin 0      // push print
0012 OpGetBuiltin 7      // push str
0014 OpGetGlobal 0       // push x
0017 OpCall 1             // str(x)
0019 OpCall 1             // print("3")
0021 OpPop
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

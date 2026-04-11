package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/evaluator"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/parser"
	"github.com/hilthontt/lotus/repl"
	"github.com/hilthontt/lotus/version"
	"github.com/hilthontt/lotus/vm"
)

func main() {
	engine := flag.String("engine", "vm", `Execution engine: "vm" (bytecode) or "eval" (tree-walking)`)
	console := flag.Bool("console", false, "Start an interactive REPL session")
	dis := flag.Bool("dis", false, "Disassemble a Lotus file instead of running it")
	annotated := flag.Bool("annotated", false, "Use annotated disassembly output (requires --dis)")
	help := flag.Bool("help", false, "Show this help message")
	ver := flag.Bool("version", false, "Print version information")
	playground := flag.Bool("playground", false, "Start the web playground")
	playgroundAddr := flag.String("playground-addr", ":3000", "Playground server address")

	flag.Usage = printHelp
	flag.Parse()

	switch {
	case *help:
		printHelp()
	case *ver:
		fmt.Println(version.GetVersionString())
	case *console:
		if err := validateEngine(*engine); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repl.Start(os.Stdin, os.Stdout, engine)
	case *dis:
		if len(flag.Args()) != 1 {
			fmt.Fprintln(os.Stderr, "error: expected a file path\nUsage: lotus --dis [--annotated] <file>")
			os.Exit(1)
		}
		disassembleFile(flag.Args()[0], *annotated)
	case *playground:
		if err := validateEngine(*engine); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		startPlayground(*playgroundAddr, *engine)
	default:
		if err := validateEngine(*engine); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(flag.Args()) != 1 {
			fmt.Fprintln(os.Stderr, "error: expected a file path\nUsage: lotus [options] <file>")
			os.Exit(1)
		}
		runFile(flag.Args()[0], *engine)
	}
}

func validateEngine(engine string) error {
	if engine != "vm" && engine != "eval" {
		return fmt.Errorf("error: unknown engine %q — must be \"vm\" or \"eval\"", engine)
	}
	return nil
}

func runFile(filePath, engine string) {
	program := mustParse(filePath)

	absPath, _ := filepath.Abs(filePath)

	var result object.Object
	if engine == "vm" {
		result = compileBytecodeAndRun(program, absPath)
	} else {
		result = evaluateAst(program)
	}

	if result != nil {
		fmt.Println(result.Inspect())
	}
}

func disassembleFile(filePath string, annotated bool) {
	program := mustParse(filePath)

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintf(os.Stderr, "compiler error: %s\n", err)
		os.Exit(1)
	}

	bytecode := comp.Bytecode()

	// Print top-level instructions.
	fmt.Println("=== main ===")
	if annotated {
		fmt.Print(code.DisassembleAnnotated(bytecode.Instructions))
	} else {
		fmt.Print(code.Disassemble(bytecode.Instructions))
	}

	// Print each compiled function's instructions.
	for i, obj := range bytecode.Constants {
		fn, ok := obj.(*object.CompiledFunction)
		if !ok {
			continue
		}
		name := fn.Name
		if name == "" {
			name = fmt.Sprintf("<fn:%d>", i)
		}
		fmt.Printf("\n=== %s ===\n", name)
		if annotated {
			fmt.Print(code.DisassembleAnnotated(fn.Instructions))
		} else {
			fmt.Print(code.Disassemble(fn.Instructions))
		}
	}
}

func mustParse(filePath string) *ast.Program {
	fmt.Println(filePath)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", filePath, err)
		os.Exit(1)
	}
	l := lexer.New(string(contents))
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		fmt.Fprintln(os.Stderr, "parse errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "\t%s\n", e)
		}
		os.Exit(1)
	}
	return program
}

func evaluateAst(program *ast.Program) object.Object {
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

func compileBytecodeAndRun(program *ast.Program, filePath string) object.Object {
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintf(os.Stderr, "compiler error: %s\n", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(filePath)
	loader := makeModuleLoaderFrom(absPath)
	machine := vm.NewWithLoader(comp.Bytecode(), loader)
	if err := machine.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "vm error: %s\n", err)
		os.Exit(1)
	}

	return machine.LastPoppedStackElement()
}

func makeModuleLoaderFrom(entryPath string) vm.ModuleLoader {
	cache := map[string]*object.Module{}

	var load func(importerPath, path string) (*object.Module, error)
	load = func(importerPath, path string) (*object.Module, error) {
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(importerPath), path)
		}
		path = filepath.Clean(path)

		if mod, ok := cache[path]; ok {
			return mod, nil
		}

		program := mustParse(path)

		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			return nil, fmt.Errorf("compile error in %q: %w", path, err)
		}

		bytecode := comp.Bytecode()
		machine := vm.NewWithLoader(bytecode, func(childPath string) (*object.Module, error) {
			return load(path, childPath)
		})
		if err := machine.Run(); err != nil {
			return nil, fmt.Errorf("runtime error in %q: %w", path, err)
		}

		mod := &object.Module{
			Path:    path,
			Exports: make(map[string]object.Object),
		}
		for name, globalIdx := range bytecode.ExportedSymbols {
			mod.Exports[name] = machine.GetGlobal(globalIdx)
		}

		cache[path] = mod
		return mod, nil
	}

	return func(path string) (*object.Module, error) {
		return load(entryPath, path)
	}
}

func printHelp() {
	fmt.Print(repl.Logo)
	fmt.Printf("  Version: %s\n\n", version.Version)
	fmt.Println("Usage:")
	fmt.Println("  lotus [options] <file>          Run a Lotus source file")
	fmt.Println("  lotus --dis <file>              Disassemble a Lotus source file")
	fmt.Println("  lotus --dis --annotated <file>  Disassemble with inline comments")
	fmt.Println("  lotus --console                 Start the interactive REPL")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lotus program.lotus")
	fmt.Println("  lotus --engine eval program.lotus")
	fmt.Println("  lotus --dis program.lotus")
	fmt.Println("  lotus --dis --annotated program.lotus")
	fmt.Println("  lotus --console")
	fmt.Println("  lotus --console --engine eval")
	fmt.Println("  lotus --version")
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/evaluator"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/parser"
	"github.com/hilthontt/lotus/vm"
)

// runFile runs a .lotus or a pre-compiled .lotusbc file.
func runFile(filePath, engine string) {
	absPath, _ := filepath.Abs(filePath)

	var result object.Object
	switch filepath.Ext(filePath) {
	case ".lotusbc":
		bc, err := compiler.ReadBytecode(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		result = runBytecode(bc, absPath)
	case ".lotus":
		program := mustParse(filePath)
		if engine == "vm" {
			result = compileBytecodeAndRun(program, absPath)
		} else {
			result = evaluateAst(program)
		}
	default:
		fatal(fmt.Sprintf("%q is not a .lotus file", filePath))
	}

	if result != nil && result.Type() != object.NIL_OBJ {
		fmt.Println(result.Inspect())
	}
}

// runBytecode runs a pre-loaded Bytecode through the VM.
// absPath must already be resolved by the caller.
func runBytecode(bc *compiler.Bytecode, absPath string) object.Object {
	machine := vm.NewWithLoader(bc, makeModuleLoaderFrom(absPath))
	machine.SetFilePath(absPath)
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vm error:", err)
		os.Exit(1)
	}
	return machine.LastPoppedStackElement()
}

// compileBytecodeAndRun compiles a parsed program and runs it through the VM.
// absPath must already be resolved by the caller.
func compileBytecodeAndRun(program *ast.Program, absPath string) object.Object {
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintln(os.Stderr, "compiler error:", err)
		os.Exit(1)
	}
	machine := vm.NewWithLoader(comp.Bytecode(), makeModuleLoaderFrom(absPath))
	machine.SetFilePath(absPath)
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vm error:", err)
		os.Exit(1)
	}
	return machine.LastPoppedStackElement()
}

func evaluateAst(program *ast.Program) object.Object {
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

// mustParse lexes and parses a .lotus file, exiting on any error.
func mustParse(filePath string) *ast.Program {
	if filepath.Ext(filePath) != ".lotus" {
		fmt.Fprintf(os.Stderr, "error: %q is not a .lotus file\n", filePath)
		os.Exit(1)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read %q: %s\n", filePath, err)
		os.Exit(1)
	}
	l := lexer.New(string(contents))
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "parse errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "\t%s\n", e)
		}
		os.Exit(1)
	}
	return program
}

// makeModuleLoaderFrom returns a ModuleLoader rooted at entryPath.
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

		var bc *compiler.Bytecode
		if filepath.Ext(path) == ".lotusbc" {
			var err error
			bc, err = compiler.ReadBytecode(path)
			if err != nil {
				return nil, fmt.Errorf("load %q: %w", path, err)
			}
		} else {
			program := mustParse(path)
			comp := compiler.New()
			if err := comp.Compile(program); err != nil {
				return nil, fmt.Errorf("compile %q: %w", path, err)
			}
			bc = comp.Bytecode()
		}

		machine := vm.NewWithLoader(bc, func(childPath string) (*object.Module, error) {
			return load(path, childPath)
		})
		machine.SetFilePath(path)
		if err := machine.Run(); err != nil {
			return nil, fmt.Errorf("runtime error in %q: %w", path, err)
		}

		mod := &object.Module{Path: path, Exports: make(map[string]object.Object)}
		for name, idx := range bc.ExportedSymbols {
			mod.Exports[name] = machine.GetGlobal(idx)
		}
		cache[path] = mod
		return mod, nil
	}

	return func(path string) (*object.Module, error) {
		return load(entryPath, path)
	}
}

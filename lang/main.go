package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/evaluator"
	"github.com/hilthontt/lotus/executable"
	"github.com/hilthontt/lotus/formatter"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/parser"
	"github.com/hilthontt/lotus/repl"
	"github.com/hilthontt/lotus/version"
	"github.com/hilthontt/lotus/vm"
)

func main() {
	engine := flag.String("engine", "vm", `Execution engine: "vm" or "eval"`)
	console := flag.Bool("console", false, "Start an interactive REPL session")
	dis := flag.Bool("dis", false, "Disassemble a .lotus file")
	annotated := flag.Bool("annotated", false, "Annotated disassembly (requires --dis)")
	help := flag.Bool("help", false, "Show help")
	ver := flag.Bool("version", false, "Print version information")
	playground := flag.Bool("playground", false, "Start the web playground")
	playgroundAddr := flag.String("playground-addr", ":3000", "Playground server address")
	buildOut := flag.String("build", "", "Output path for compiled executable")
	compile := flag.Bool("compile", false, "Compile .lotus to .lotusbc without running")
	fmtFlag := flag.Bool("fmt", false, "Format .lotus files in place")
	fmtCheck := flag.Bool("fmt-check", false, "Check formatting, exit 1 if unformatted")

	flag.Usage = printHelp
	flag.Parse()

	switch {
	case *help:
		printHelp()

	case *ver:
		fmt.Println(version.GetVersionString())

	case *console:
		if err := validateEngine(*engine); err != nil {
			fatal(err.Error())
		}
		repl.Start(os.Stdin, os.Stdout, engine)

	case *compile:
		// lotus --compile file.lotus  →  writes file.lotusbc
		if len(flag.Args()) != 1 {
			fatal("usage: lotus --compile <file.lotus>")
		}
		compileToBytecode(flag.Args()[0])

	case *dis:
		if len(flag.Args()) != 1 {
			fatal("usage: lotus --dis [--annotated] <file>")
		}
		disassembleFile(flag.Args()[0], *annotated)

	case *playground:
		if err := validateEngine(*engine); err != nil {
			fatal(err.Error())
		}
		startPlayground(*playgroundAddr, *engine)

	case *buildOut != "":
		src, output := *buildOut, *buildOut
		if strings.HasSuffix(output, ".lotus") {
			output = ""
		} else {
			if len(flag.Args()) != 1 {
				fatal("usage: lotus --build [output.exe] <file.lotus>")
			}
			src = flag.Args()[0]
		}
		if err := validateEngine(*engine); err != nil {
			fatal(err.Error())
		}
		if err := executable.BuildExecutable(src, output, *engine); err != nil {
			fmt.Fprintln(os.Stderr, "build error:", err)
			os.Exit(1)
		}

	case *fmtFlag || *fmtCheck:
		args := flag.Args()
		if len(args) == 0 {
			args = findLotusFiles(".")
		}
		runFmt(args, *fmtCheck)

	default:
		if err := validateEngine(*engine); err != nil {
			fatal(err.Error())
		}
		if len(flag.Args()) != 1 {
			fatal("usage: lotus [options] <file>")
		}
		runFile(flag.Args()[0], *engine)
	}
}

func validateEngine(engine string) error {
	if engine != "vm" && engine != "eval" {
		return fmt.Errorf("unknown engine %q — must be \"vm\" or \"eval\"", engine)
	}
	return nil
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}

// runFile runs a .lotus or a pre-compiled .lotusbc file.
func runFile(filePath, engine string) {
	ext := filepath.Ext(filePath)

	absPath, _ := filepath.Abs(filePath)

	var result object.Object
	if ext == ".lotusbc" {
		// Run pre-compiled bytecode directly.
		bc, err := compiler.ReadBytecode(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		result = runBytecode(bc, absPath)
	} else {
		// Parse + compile + run.
		if ext != ".lotus" {
			fatal(fmt.Sprintf("%q is not a .lotus file", filePath))
		}
		program := mustParse(filePath)
		if engine == "vm" {
			result = compileBytecodeAndRun(program, absPath)
		} else {
			result = evaluateAst(program)
		}
	}

	if result != nil && result.Type() != object.NIL_OBJ {
		fmt.Println(result.Inspect())
	}
}

// compileToBytecode compiles a .lotus file and writes a .lotusbc file.
func compileToBytecode(filePath string) {
	if filepath.Ext(filePath) != ".lotus" {
		fatal(fmt.Sprintf("%q is not a .lotus file", filePath))
	}
	program := mustParse(filePath)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintln(os.Stderr, "compiler error:", err)
		os.Exit(1)
	}
	outPath := strings.TrimSuffix(filePath, ".lotus") + ".lotusbc"
	if err := compiler.WriteBytecode(comp.Bytecode(), outPath); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}
	fmt.Printf("compiled: %s → %s\n", filePath, outPath)
}

func disassembleFile(filePath string, annotated bool) {
	var bc *compiler.Bytecode

	if filepath.Ext(filePath) == ".lotusbc" {
		var err error
		bc, err = compiler.ReadBytecode(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	} else {
		program := mustParse(filePath)
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			fmt.Fprintln(os.Stderr, "compiler error:", err)
			os.Exit(1)
		}
		bc = comp.Bytecode()
	}

	fmt.Println("=== main ===")
	if annotated {
		fmt.Print(code.DisassembleAnnotated(bc.Instructions))
	} else {
		fmt.Print(code.Disassemble(bc.Instructions))
	}

	for i, obj := range bc.Constants {
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
	if filepath.Ext(filePath) != ".lotus" {
		fmt.Fprintf(os.Stderr, "error: %q is not a .lotus file\n", filePath)
		os.Exit(1)
	}
	fmt.Println(filePath)
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

func evaluateAst(program *ast.Program) object.Object {
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

// runBytecode runs a pre-loaded Bytecode (from .lotusbc) through the VM.
func runBytecode(bc *compiler.Bytecode, filePath string) object.Object {
	absPath, _ := filepath.Abs(filePath)
	loader := makeModuleLoaderFrom(absPath)
	machine := vm.NewWithLoader(bc, loader)
	machine.SetFilePath(absPath)
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vm error:", err)
		os.Exit(1)
	}
	return machine.LastPoppedStackElement()
}

func compileBytecodeAndRun(program *ast.Program, filePath string) object.Object {
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintln(os.Stderr, "compiler error:", err)
		os.Exit(1)
	}
	absPath, _ := filepath.Abs(filePath)
	loader := makeModuleLoaderFrom(absPath)
	machine := vm.NewWithLoader(comp.Bytecode(), loader)
	machine.SetFilePath(absPath) // ← enables stack traces
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vm error:", err)
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

		// Support importing pre-compiled .lotusbc modules too.
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

func printHelp() {
	fmt.Print(repl.Logo)
	fmt.Printf("  Version: %s\n\n", version.Version)
	fmt.Println("Usage:")
	fmt.Println("  lotus <file.lotus>               Run a source file")
	fmt.Println("  lotus <file.lotusbc>             Run a pre-compiled file")
	fmt.Println("  lotus --compile <file.lotus>     Compile to .lotusbc")
	fmt.Println("  lotus --dis <file>               Disassemble")
	fmt.Println("  lotus --dis --annotated <file>   Disassemble with comments")
	fmt.Println("  lotus --console                  Interactive REPL")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lotus program.lotus")
	fmt.Println("  lotus --compile program.lotus && lotus program.lotusbc")
	fmt.Println("  lotus --dis --annotated program.lotus")
	fmt.Println("  lotus --console")
	fmt.Println("  lotus --version")
}

// runFmt formats each file. If checkOnly is true it reports unformatted
// files and exits 1 without writing, suitable for CI.
func runFmt(paths []string, checkOnly bool) {
	unformatted := 0
	for _, path := range paths {
		changed, err := fmtFile(path, checkOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fmt: %s: %s\n", path, err)
			continue
		}
		if changed {
			unformatted++
			if checkOnly {
				fmt.Fprintf(os.Stderr, "unformatted: %s\n", path)
			} else {
				fmt.Printf("formatted:   %s\n", path)
			}
		}
	}
	if checkOnly && unformatted > 0 {
		os.Exit(1)
	}
}

// fmtFile formats a single .lotus file.
// Returns (true, nil) if the file was changed (or would be changed in check mode).
func fmtFile(path string, checkOnly bool) (bool, error) {
	if filepath.Ext(path) != ".lotus" {
		return false, fmt.Errorf("not a .lotus file")
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return false, fmt.Errorf("parse error: %s", errs[0])
	}

	formatted := formatter.Format(program, l.Comments)

	// If nothing changed, nothing to do.
	if bytes.Equal(src, []byte(formatted)) {
		return false, nil
	}

	if checkOnly {
		return true, nil
	}

	// Write formatted output back to the file.
	if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
		return false, err
	}
	return true, nil
}

// findLotusFiles returns all .lotus files under root recursively.
func findLotusFiles(root string) []string {
	var files []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && filepath.Ext(path) == ".lotus" {
			files = append(files, path)
		}
		return nil
	})
	return files
}

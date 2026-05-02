package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	watch := flag.Bool("watch", false, "Re-run file on every save")

	flag.Usage = printHelp
	flag.Parse()

	switch {
	case *help:
		printHelp()

	case *ver:
		fmt.Println(version.GetVersionString())

	case *console:
		mustValidateEngine(*engine)
		repl.Start(os.Stdin, os.Stdout, engine)

	case *compile:
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
		mustValidateEngine(*engine)
		startPlayground(*playgroundAddr, *engine)

	case *buildOut != "":
		mustValidateEngine(*engine)
		src, out := resolveBuildPaths(*buildOut, flag.Args())
		if err := executable.BuildExecutable(src, out, *engine); err != nil {
			fmt.Fprintln(os.Stderr, "build error:", err)
			os.Exit(1)
		}

	case *fmtFlag || *fmtCheck:
		args := flag.Args()
		if len(args) == 0 {
			args = findLotusFiles(".")
		}
		runFmt(args, *fmtCheck)

	case *watch:
		mustValidateEngine(*engine)
		if len(flag.Args()) != 1 {
			fatal("usage: lotus --watch <file.lotus>")
		}
		watchFile(flag.Args()[0], *engine)

	default:
		mustValidateEngine(*engine)
		if len(flag.Args()) != 1 {
			fatal("usage: lotus [options] <file>")
		}
		runFile(flag.Args()[0], *engine)
	}
}

// resolveBuildPaths returns (srcFile, outputFile) from the --build flag value
// and any remaining positional args.
//
//   - lotus --build out.exe file.lotus  →  src=file.lotus, out=out.exe
//   - lotus --build file.lotus          →  src=file.lotus, out="" (executable derives name)
func resolveBuildPaths(buildFlag string, args []string) (src, out string) {
	if strings.HasSuffix(buildFlag, ".lotus") {
		// --build was given the source file directly; no separate output path.
		return buildFlag, ""
	}
	if len(args) != 1 {
		fatal("usage: lotus --build [output.exe] <file.lotus>")
	}
	return args[0], buildFlag
}

func mustValidateEngine(engine string) {
	if engine != "vm" && engine != "eval" {
		fatal(fmt.Sprintf("unknown engine %q — must be \"vm\" or \"eval\"", engine))
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}

// runFile runs a .lotus or a pre-compiled .lotusbc file.
func runFile(filePath, engine string) {
	absPath, _ := filepath.Abs(filePath)
	ext := filepath.Ext(filePath)

	var result object.Object
	switch ext {
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

func evaluateAst(program *ast.Program) object.Object {
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
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

func printHelp() {
	fmt.Print(repl.Logo)
	fmt.Printf("  Version: %s\n\n", version.Version)
	fmt.Println("Usage:")
	fmt.Println("  lotus <file.lotus>               Run a source file")
	fmt.Println("  lotus <file.lotusbc>             Run a pre-compiled file")
	fmt.Println("  lotus --compile <file.lotus>     Compile to .lotusbc")
	fmt.Println("  lotus --dis <file>               Disassemble")
	fmt.Println("  lotus --dis --annotated <file>   Disassemble with comments")
	fmt.Println("  lotus --fmt [files...]           Format in place")
	fmt.Println("  lotus --fmt-check [files...]     Check formatting (CI)")
	fmt.Println("  lotus --watch <file.lotus>       Re-run on every save")
	fmt.Println("  lotus --console                  Interactive REPL")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lotus program.lotus")
	fmt.Println("  lotus --watch program.lotus")
	fmt.Println("  lotus --fmt")
	fmt.Println("  lotus --compile program.lotus && lotus program.lotusbc")
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
	if bytes.Equal(src, []byte(formatted)) {
		return false, nil
	}
	if checkOnly {
		return true, nil
	}
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

// watchFile runs filePath every time it changes on disk.
// It polls for changes (no external dependencies needed).
// Ctrl+C exits cleanly.
func watchFile(filePath, engine string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fatal("watch: " + err.Error())
	}

	fmt.Printf("\033[2m  watching %s — press Ctrl+C to stop\033[0m\n\n", filepath.Base(filePath))

	var lastMod time.Time
	for {
		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "watch: %s\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if !info.ModTime().Equal(lastMod) {
			if !lastMod.IsZero() {
				fmt.Printf("\n\033[2m─── %s ───\033[0m\n\n", time.Now().Format("15:04:05"))
			}
			lastMod = info.ModTime()
			runFile(absPath, engine)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

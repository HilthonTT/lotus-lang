package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hilthontt/lotus/ast"
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
	engine := flag.String("engine", "vm", "Engine options are \"vm\" or \"eval\"")
	console := flag.Bool("console", false, "Provide console flag to enter interactive repl")
	help := flag.Bool("help", false, "Show this help message")
	ver := flag.Bool("version", false, "Print version information")

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
	contents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", filePath, err)
		os.Exit(1)
	}

	program := parse(string(contents))

	var result object.Object
	if engine == "vm" {
		result = compileBytecodeAndRun(program)
	} else {
		result = evaluateAst(program)
	}

	if result != nil {
		fmt.Println(result.Inspect())
	}
}

func parse(src string) *ast.Program {
	l := lexer.New(src)
	p := parser.New(l)
	return p.ParseProgram()
}

// Evaluate the AST with evaluator
func evaluateAst(program *ast.Program) object.Object {
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

// Compile program to bytecode, pass to VM, and run. Returns the last popped stack element (result)
func compileBytecodeAndRun(program *ast.Program) object.Object {
	comp := compiler.New()

	if err := comp.Compile(program); err != nil {
		fmt.Printf("compiler error: %s", err)
		os.Exit(1)
	}

	vm := vm.New(comp.Bytecode())
	if err := vm.Run(); err != nil {
		fmt.Printf("vm error: %s", err)
		os.Exit(1)
	}

	return vm.LastPoppedStackElement()
}

func printHelp() {
	fmt.Print(repl.Logo)
	fmt.Printf("  Version: %s\n\n", version.Version)
	fmt.Println("Usage:")
	fmt.Println("  lotus [options] <file>   Run a Lotus source file")
	fmt.Println("  lotus --console          Start the interactive REPL")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lotus program.lotus")
	fmt.Println("  lotus --engine eval program.lotus")
	fmt.Println("  lotus --console")
	fmt.Println("  lotus --console --engine eval")
	fmt.Println("  lotus --version")
}

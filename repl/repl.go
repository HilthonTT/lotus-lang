package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/evaluator"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/parser"
)

const Logo = `
██╗      ██████╗ ████████╗██╗   ██╗███████╗
██║     ██╔═══██╗╚══██╔══╝██║   ██║██╔════╝
██║     ██║   ██║   ██║   ██║   ██║███████╗
██║     ██║   ██║   ██║   ██║   ██║╚════██║
███████╗╚██████╔╝   ██║   ╚██████╔╝███████║
╚══════╝ ╚═════╝    ╚═╝    ╚═════╝ ╚══════╝

  A compiled language with a stack-based VM
  Type 'help' for commands, Ctrl+C to exit
`

const Oops = `
 ██████╗  ██████╗ ██████╗ ███████╗
██╔═══██╗██╔═══██╗██╔══██╗██╔════╝
██║   ██║██║   ██║██████╔╝███████╗
██║   ██║██║   ██║██╔═══╝ ╚════██║
╚██████╔╝╚██████╔╝██║     ███████║
 ╚═════╝  ╚═════╝ ╚═╝     ╚══════╝`

const prompt = ">> "

func Start(in io.Reader, out io.Writer, engine *string) {
	scanner := bufio.NewScanner(in)

	evalEnv := object.NewEnvironment()
	vmSt := newVMState()

	for {
		fmt.Fprint(out, prompt)

		if !scanner.Scan() {
			fmt.Fprintln(out) // clean newline on Ctrl+C / EOF
			return
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		program, ok := parseInput(line, out)
		if !ok {
			continue
		}

		var result object.Object
		var err error

		if engine != nil && *engine == "eval" {
			result = evaluator.Eval(program, evalEnv)
		} else {
			result, err = vmSt.run(program)
			if err != nil {
				fmt.Fprintf(out, "error: %s\n", err)
				continue
			}
		}

		if result != nil {
			fmt.Fprintln(out, result.Inspect())
		}
	}
}

func parseInput(line string, out io.Writer) (*ast.Program, bool) {
	l := lexer.New(line)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) != 0 {
		printParserErrors(out, errs)
		return nil, false
	}
	return program, true
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprint(out, Oops)
	fmt.Fprintln(out, "\nparser errors:")
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}

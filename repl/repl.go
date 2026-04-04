package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/parser"
)

const Oops = `
 ██████╗  ██████╗ ██████╗ ███████╗
██╔═══██╗██╔═══██╗██╔══██╗██╔════╝
██║   ██║██║   ██║██████╔╝███████╗
██║   ██║██║   ██║██╔═══╝ ╚════██║
╚██████╔╝╚██████╔╝██║     ███████║
 ╚═════╝  ╚═════╝ ╚═╝     ╚══════╝`

// TODO: Implement this later
func Start(in io.Reader, out io.Writer, engine *string) {
	scanner := bufio.NewScanner(in)
	// env := object.NewEnvironment()
	// constants := []object.Object{}
	// globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range compiler.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Printf(">> ")

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		_ = parser.New(l)

		// program := p.ParseProgram()
		// if len(p.Errors()) != 0 {
		// 	printParserErrors(out, p.Errors())
		// 	continue
		// }

		// if engine != nil && *engine == "eval" {
		// 	evaluate(program, env, out)
		// } else if engine != nil && *engine == "vm" {
		// 	if err := compileAndExecute(symbolTable, constants, program, globals, out); err != nil {
		// 		fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
		// 		continue
		// 	}
		// }
	}
}

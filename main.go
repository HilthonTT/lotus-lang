package main

import (
	"fmt"
	"os"

	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/version"
)

const logo = `
██╗      ██████╗ ████████╗██╗   ██╗███████╗
██║     ██╔═══██╗╚══██╔══╝██║   ██║██╔════╝
██║     ██║   ██║   ██║   ██║   ██║███████╗
██║     ██║   ██║   ██║   ██║   ██║╚════██║
███████╗╚██████╔╝   ██║   ╚██████╔╝███████║
╚══════╝ ╚═════╝    ╚═╝    ╚═════╝ ╚══════╝

  A compiled language with a stack-based VM
  Type 'help' for commands, Ctrl+C to exit
`

func main() {
	repl()

	data, err := os.ReadFile("./examples/file.lotus")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	input := string(data)
	tokens := lexer.Tokenize(input)

	for _, t := range tokens {
		fmt.Println(t.String())
	}
}

func repl() {
	v := version.GetVersionInfo()
	fmt.Print(logo, v.Version)
	fmt.Println()
}

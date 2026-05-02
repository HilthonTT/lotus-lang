package main

import (
	"flag"
	"fmt"

	"github.com/hilthontt/lotus/repl"
	"github.com/hilthontt/lotus/version"
)

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

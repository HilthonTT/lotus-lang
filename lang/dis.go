package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/object"
)

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

	printDisassembly("main", bc.Instructions, annotated)

	for i, obj := range bc.Constants {
		fn, ok := obj.(*object.CompiledFunction)
		if !ok {
			continue
		}
		name := fn.Name
		if name == "" {
			name = fmt.Sprintf("<fn:%d>", i)
		}
		fmt.Println()
		printDisassembly(name, fn.Instructions, annotated)
	}
}

func printDisassembly(name string, instructions code.Instructions, annotated bool) {
	fmt.Printf("=== %s ===\n", name)
	if annotated {
		fmt.Print(code.DisassembleAnnotated(instructions))
	} else {
		fmt.Print(code.Disassemble(instructions))
	}
}

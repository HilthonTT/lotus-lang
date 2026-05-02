package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/executable"
)

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

// runBuild resolves paths from the --build flag and builds a native executable.
func runBuild(buildFlag string, args []string, engine string) {
	src, out := resolveBuildPaths(buildFlag, args)
	if err := executable.BuildExecutable(src, out, engine); err != nil {
		fmt.Fprintln(os.Stderr, "build error:", err)
		os.Exit(1)
	}
}

// resolveBuildPaths returns (srcFile, outputFile) from the --build flag value
// and any remaining positional args.
//
//   - lotus --build out.exe file.lotus  →  src=file.lotus, out=out.exe
//   - lotus --build file.lotus          →  src=file.lotus, out="" (executable derives name)
func resolveBuildPaths(buildFlag string, args []string) (src, out string) {
	if strings.HasSuffix(buildFlag, ".lotus") {
		return buildFlag, ""
	}
	if len(args) != 1 {
		fatal("usage: lotus --build [output.exe] <file.lotus>")
	}
	return args[0], buildFlag
}

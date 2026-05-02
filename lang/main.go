package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hilthontt/lotus/repl"
	"github.com/hilthontt/lotus/version"
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
		runBuild(*buildOut, flag.Args(), *engine)

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

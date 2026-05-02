package main

import (
	"fmt"
	"os"
)

func mustValidateEngine(engine string) {
	if engine != "vm" && engine != "eval" {
		fatal(fmt.Sprintf("unknown engine %q — must be \"vm\" or \"eval\"", engine))
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}

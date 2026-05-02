package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hilthontt/lotus/formatter"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/parser"
)

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

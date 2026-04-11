package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/evaluator"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/object"
	"github.com/hilthontt/lotus/parser"
	"github.com/hilthontt/lotus/vm"
)

type runResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error,omitempty"`
}

func startPlayground(addr string, engine string) {
	fs := http.FileServer(http.Dir("."))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "playground.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(r.Body); err != nil {
			json.NewEncoder(w).Encode(runResponse{Error: "failed to read body"})
			return
		}
		source := buf.String()

		stdout, stderr, runErr := runSource(source, engine)
		resp := runResponse{Stdout: stdout, Stderr: stderr}
		if runErr != nil {
			resp.Error = runErr.Error()
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Printf("🪷  Lotus Playground running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, "playground error:", err)
		os.Exit(1)
	}
}

// runSource compiles and runs Lotus source, capturing stdout/stderr.
func runSource(source string, engine string) (stdout string, stderr string, err error) {
	// Capture os.Stdout by redirecting via a pipe
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	defer func() {
		// Restore
		wOut.Close()
		wErr.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		var obuf bytes.Buffer
		obuf.ReadFrom(rOut)
		stdout = obuf.String()

		var ebuf bytes.Buffer
		ebuf.ReadFrom(rErr)
		stderr = ebuf.String()
	}()

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) != 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		return
	}

	if engine == "eval" {
		env := object.NewEnvironment()
		result := evaluator.Eval(program, env)
		if result != nil && result.Type() == "ERROR" {
			err = fmt.Errorf("%s", result.Inspect())
		}
		return
	}

	// VM engine
	comp := compiler.New()
	if compErr := comp.Compile(program); compErr != nil {
		err = compErr
		return
	}

	machine := vm.New(comp.Bytecode())
	if vmErr := machine.Run(); vmErr != nil {
		err = vmErr
	}
	return
}

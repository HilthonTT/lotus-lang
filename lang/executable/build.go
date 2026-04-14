package executable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// BuildExecutable compiles a .lotus file into a native executable.
// It does this by:
//  1. Generating a Go main.go that embeds the Lotus source
//  2. Copying the go.mod so the module is resolvable
//  3. Running `go build` in that temp directory
func BuildExecutable(sourcePath, outputPath, engine string) error {
	// Read the Lotus source
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("could not read: %q: %w", sourcePath, err)
	}
	source := string(sourceBytes)

	// Resolve output path
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		outputPath = base + ".exe"
	}
	absOut, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("could not resolve output path: %w", err)
	}

	// Find the module root (where go.mod lives)
	moduleRoot, moduleName, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("could not find go.mod: %w", err)
	}

	// Create a temp directory for the generated program
	tmpDir, err := os.MkdirTemp("", "lotus-build-*")
	if err != nil {
		return fmt.Errorf("could not create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate main.go
	mainGo, err := generateMain(source, moduleName, engine)
	if err != nil {
		return fmt.Errorf("could not generate main.go: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("could not write main.go: %w", err)
	}

	// Write go.mod that replaces the lotus module with the local one
	goMod := fmt.Sprintf(`module lotus-app
 
go 1.26.1
 
require %s v0.0.0
 
replace %s => %s
`, moduleName, moduleName, moduleRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		return fmt.Errorf("could not write go.mod: %w", err)
	}

	// Run `go mod tidy` to pull in transitive deps
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmpDir
	tidy.Stdout = os.Stdout
	tidy.Stderr = os.Stderr
	if err := tidy.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Run `go build -o <output>`
	build := exec.Command("go", "build", "-o", absOut, ".")
	build.Dir = tmpDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Printf("Built: %s\n", absOut)
	return nil
}

// findModuleRoot walks up from the current directory to find go.mod.
func findModuleRoot() (root, moduleName string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	for {
		candidate := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(candidate)
		if err == nil {
			// Parse module name from first line: "module github.com/..."
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					name := strings.TrimPrefix(line, "module ")
					name = strings.TrimSpace(name)
					absDir, _ := filepath.Abs(dir)
					return absDir, name, nil
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", "", fmt.Errorf("no go.mod found")
}

var mainTemplate = template.Must(template.New("main").Parse(`package main

import (
	"fmt"
	"os"
	{{if eq .Engine "eval"}}
	"{{.Module}}/evaluator"
	"{{.Module}}/lexer"
	"{{.Module}}/object"
	"{{.Module}}/parser"
	{{else}}
	"path/filepath"
	"{{.Module}}/compiler"
	"{{.Module}}/lexer"
	"{{.Module}}/parser"
	"{{.Module}}/vm"
	{{end}}
)

const source = {{.SourceLiteral}}

func main() {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) != 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	{{if eq .Engine "eval"}}
	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)
	if result != nil && result.Type() == "ERROR" {
		fmt.Fprintln(os.Stderr, result.Inspect())
		os.Exit(1)
	}
	{{else}}
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintln(os.Stderr, "compiler error:", err)
		os.Exit(1)
	}

	exe, _ := os.Executable()
	_, _ = filepath.Abs(filepath.Dir(exe))

	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vm error:", err)
		os.Exit(1)
	}
	{{end}}
}
`))

type mainTemplateData struct {
	Module        string
	SourceLiteral string
	Engine        string
}

func generateMain(source, moduleName, engine string) (string, error) {
	// Escape the source as a raw Go string literal
	// Use backtick quoting, escaping any backticks in the source itself
	var literal string
	if !strings.Contains(source, "`") {
		literal = "`" + source + "`"
	} else {
		// Fall back to double-quoted string with proper escaping
		literal = fmt.Sprintf("%q", source)
	}

	var sb strings.Builder
	err := mainTemplate.Execute(&sb, mainTemplateData{
		Module:        moduleName,
		SourceLiteral: literal,
		Engine:        engine,
	})
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

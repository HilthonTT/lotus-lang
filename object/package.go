package object

import (
	"fmt"
	"strings"
)

// PackageFunction is a function exposed by a builtin package.
type PackageFunction func(args ...Object) Object

// Package is a named collection of builtin functions.
type Package struct {
	Name      string
	Functions map[string]PackageFunction
}

func (p *Package) Type() ObjectType {
	return PACKAGE_OBJ
}

func (p *Package) Inspect() string {
	keys := make([]string, 0, len(p.Functions))
	for k := range p.Functions {
		keys = append(keys, k)
	}
	return fmt.Sprintf("<package %s [%s]>", p.Name, strings.Join(keys, ", "))
}

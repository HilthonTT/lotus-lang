package ast

import (
	"bytes"
	"strings"

	"github.com/hilthontt/lotus/token"
)

// ExportStatement wraps any statement and marks it as exported.
// e.g. export let x = 5   /   export fn foo() {}   /   export class Bar {}
type ExportStatement struct {
	Token     token.Token // the 'export' token
	Statement Statement
}

func (es *ExportStatement) statementNode() {}

func (es *ExportStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExportStatement) String() string {
	return "export " + es.Statement.String()
}

// ImportStatement: import { x, y } from "path/to/file"
type ImportStatement struct {
	Token token.Token   // the 'import' token
	Names []*Identifier // names to import
	Path  string        // file path (unquoted)
}

func (is *ImportStatement) statementNode() {}

func (is *ImportStatement) TokenLiteral() string {
	return is.Token.Literal
}

func (is *ImportStatement) String() string {
	var out bytes.Buffer
	names := make([]string, len(is.Names))
	for i, n := range is.Names {
		names[i] = n.Value
	}
	out.WriteString("import { ")
	out.WriteString(strings.Join(names, ", "))
	out.WriteString(` } from "`)
	out.WriteString(is.Path)
	out.WriteString(`"`)
	return out.String()
}

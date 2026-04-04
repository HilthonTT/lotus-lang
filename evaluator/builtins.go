package evaluator

import "github.com/hilthontt/lotus/compiler"

var builtinFunctions = map[string]*compiler.BuiltinDef{
	"print": compiler.GetBuiltinByName("print"),
	"len":   compiler.GetBuiltinByName("len"),
	"push":  compiler.GetBuiltinByName("push"),
	"pop":   compiler.GetBuiltinByName("pop"),
	"head":  compiler.GetBuiltinByName("head"),
	"tail":  compiler.GetBuiltinByName("tail"),
	"type":  compiler.GetBuiltinByName("type"),
	"str":   compiler.GetBuiltinByName("str"),
	"int":   compiler.GetBuiltinByName("int"),
	"range": compiler.GetBuiltinByName("range"),
}

package evaluator

import (
	"github.com/hilthontt/lotus/object"
)

// No need to create new true/false or null objects every time we encounter one, they will
// be the same. Let's reference them instead
var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Nil{}
)

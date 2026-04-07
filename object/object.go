package object

type ObjectType string

const (
	INTEGER_OBJ  ObjectType = "INTEGER"
	FLOAT_OBJ    ObjectType = "FLOAT"
	BOOLEAN_OBJ  ObjectType = "BOOLEAN"
	STRING_OBJ   ObjectType = "STRING"
	NIL_OBJ      ObjectType = "NIL"
	ARRAY_OBJ    ObjectType = "ARRAY"
	MAP_OBJ      ObjectType = "MAP"
	CLOSURE_OBJ  ObjectType = "CLOSURE"
	BUILTIN_OBJ  ObjectType = "BUILTIN"
	HASH_OBJ     ObjectType = "HASH"
	CLASS_OBJ    ObjectType = "CLASS"
	INSTANCE_OBJ ObjectType = "INSTANCE"
	SUPER_OBJ    ObjectType = "SUPER"
)

// Object is the interface that all of our various object-types must implemented.
type Object interface {
	// Type returns the type of this object.
	Type() ObjectType

	// Inspect returns a string-representation of the given object.
	Inspect() string
}

// Hashable type can be hashed
type Hashable interface {
	HashKey() HashKey
}

func IsTruthy(obj Object) bool {
	switch o := obj.(type) {
	case *Boolean:
		return o.Value
	case *Nil:
		return false
	case *Integer:
		return o.Value != 0
	case *String:
		return o.Value != ""
	case *Array:
		return len(o.Elements) > 0
	default:
		return true
	}
}

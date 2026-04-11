package vm

import "github.com/hilthontt/lotus/object"

func nativeBoolToBooleanObj(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch o := obj.(type) {
	case *object.Boolean:
		return o.Value
	case *object.Nil:
		return false
	case *object.Integer:
		return o.Value != 0
	case *object.Float:
		return o.Value != 0
	case *object.String:
		return o.Value != ""
	case *object.Array:
		return len(o.Elements) > 0
	case *object.Map:
		return len(o.Pairs) > 0
	default:
		return true
	}
}

func isNumeric(obj object.Object) bool {
	t := obj.Type()
	return t == object.INTEGER_OBJ || t == object.FLOAT_OBJ
}

func toFloat(obj object.Object) float64 {
	switch o := obj.(type) {
	case *object.Integer:
		return float64(o.Value)
	case *object.Float:
		return o.Value
	default:
		return 0
	}
}

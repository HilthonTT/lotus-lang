package object

import "fmt"

// ok(value) and err("message") return a Result object.
// Results have .ok (bool), .value (any), .error (string) fields.
const RESULT_OBJ ObjectType = "RESULT"

type Result struct {
	Ok     bool
	Value  Object // non-nil when Ok == true
	ErrMsg string // non-empty when Ok == false
}

func (r *Result) Type() ObjectType { return RESULT_OBJ }
func (r *Result) Inspect() string {
	if r.Ok {
		if r.Value != nil {
			return fmt.Sprintf("ok(%s)", r.Value.Inspect())
		}
		return "ok()"
	}
	return fmt.Sprintf("err(%q)", r.ErrMsg)
}

// OkResult creates a successful Result.
func OkResult(val Object) *Result {
	if val == nil {
		val = &Nil{}
	}
	return &Result{Ok: true, Value: val}
}

// ErrResult creates a failed Result.
func ErrResult(msg string) *Result {
	return &Result{Ok: false, ErrMsg: msg, Value: &Nil{}}
}

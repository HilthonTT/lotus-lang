package evaluator

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/object"
)

// No need to create new true/false or null objects every time we encounter one, they will
// be the same. Let's reference them instead
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NIL   = &object.Nil{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// --- Program ---
	case *ast.Program:
		return evalProgram(node, env)

	// --- Statements ---
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if node.Mutable {
			env.Set(node.Name.Value, val)
		} else {
			env.SetConst(node.Name.Value, val)
		}

	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if !env.IsMutable(node.Name.Value) {
			return newError("cannot assign to immutable variable: %s", node.Name.Value)
		}
		env.Set(node.Name.Value, val)

	case *ast.IndexAssignStatement:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		return evalIndexAssign(left, index, value)

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			val := Eval(node.ReturnValue, env)
			if isError(val) {
				return val
			}
			return &ReturnValue{Value: val}
		}
		return &ReturnValue{Value: NIL}

	case *ast.WhileStatement:
		return evalWhileStatement(node, env)

	case *ast.ForStatement:
		return evalForStatement(node, env)

	case *ast.BreakStatement:
		return &BreakSignal{}

	case *ast.ContinueStatement:
		return &ContinueSignal{}

	// --- Expressions ---
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObj(node.Value)

	case *ast.NilLiteral:
		return NIL

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		// Short-circuit && and ||
		if node.Operator == "&&" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if !object.IsTruthy(left) {
				return FALSE
			}
			right := Eval(node.Right, env)
			if isError(right) {
				return right
			}
			return nativeBoolToBooleanObj(object.IsTruthy(right))
		}
		if node.Operator == "||" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if object.IsTruthy(left) {
				return TRUE
			}
			right := Eval(node.Right, env)
			if isError(right) {
				return right
			}
			return nativeBoolToBooleanObj(object.IsTruthy(right))
		}

		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		fn := &Function{
			Name:       node.Name,
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
		if node.Name != "" {
			env.SetConst(node.Name, fn)
		}
		return fn

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.MapLiteral:
		return evalMapLiteral(node, env)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *ReturnValue:
			return result.Value
		case *Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
		if result != nil {
			rt := result.Type()
			if rt == "RETURN_VALUE" || rt == "ERROR" || rt == "BREAK" || rt == "CONTINUE" {
				return result
			}
		}
	}
	return result
}

func evalWhileStatement(node *ast.WhileStatement, env *object.Environment) object.Object {
	for {
		condition := Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}

		if !object.IsTruthy(condition) {
			break
		}

		result := Eval(node.Body, env)
		if isError(result) {
			return result
		}

		if _, ok := result.(*BreakSignal); ok {
			break
		}

		if _, ok := result.(*ContinueSignal); ok {
			continue
		}
		if _, ok := result.(*ReturnValue); ok {
			return result
		}
	}
	return NIL
}

func evalForStatement(node *ast.ForStatement, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	arr, ok := iterable.(*object.Array)
	if !ok {
		return newError("for-in requires an array, got %s", iterable.Type())
	}

	for _, elem := range arr.Elements {
		env.Set(node.Variable.Value, elem)

		result := Eval(node.Body, env)
		if isError(result) {
			return result
		}
		if _, ok := result.(*BreakSignal); ok {
			break
		}
		if _, ok := result.(*ContinueSignal); ok {
			continue
		}
		if _, ok := result.(*ReturnValue); ok {
			return result
		}
	}
	return NIL
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return nativeBoolToBooleanObj(!object.IsTruthy(right))
	case "-":
		return evalMinusPrefix(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefix(right object.Object) object.Object {
	switch o := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -o.Value}
	case *object.Float:
		return &object.Float{Value: -o.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

// --- Infix ---

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfix(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfix(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfix(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObj(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObj(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfix(operator string, left, right object.Object) object.Object {
	l := left.(*object.Integer).Value
	r := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: l + r}
	case "-":
		return &object.Integer{Value: l - r}
	case "*":
		return &object.Integer{Value: l * r}
	case "/":
		if r == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: l / r}
	case "%":
		if r == 0 {
			return newError("modulo by zero")
		}
		return &object.Integer{Value: l % r}
	case "<":
		return nativeBoolToBooleanObj(l < r)
	case ">":
		return nativeBoolToBooleanObj(l > r)
	case "<=":
		return nativeBoolToBooleanObj(l <= r)
	case ">=":
		return nativeBoolToBooleanObj(l >= r)
	case "==":
		return nativeBoolToBooleanObj(l == r)
	case "!=":
		return nativeBoolToBooleanObj(l != r)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfix(operator string, left, right object.Object) object.Object {
	l := toFloat(left)
	r := toFloat(right)

	switch operator {
	case "+":
		return &object.Float{Value: l + r}
	case "-":
		return &object.Float{Value: l - r}
	case "*":
		return &object.Float{Value: l * r}
	case "/":
		if r == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: l / r}
	case "<":
		return nativeBoolToBooleanObj(l < r)
	case ">":
		return nativeBoolToBooleanObj(l > r)
	case "<=":
		return nativeBoolToBooleanObj(l <= r)
	case ">=":
		return nativeBoolToBooleanObj(l >= r)
	case "==":
		return nativeBoolToBooleanObj(l == r)
	case "!=":
		return nativeBoolToBooleanObj(l != r)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfix(operator string, left, right object.Object) object.Object {
	l := left.(*object.String).Value
	r := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: l + r}
	case "==":
		return nativeBoolToBooleanObj(l == r)
	case "!=":
		return nativeBoolToBooleanObj(l != r)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	if object.IsTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	}
	return NIL
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// Check builtins
	if builtin, ok := builtinFunctions[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
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

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == "ERROR"
}

func newError(format string, args ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}

func evalExpressions(exprs []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exprs {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func nativeBoolToBooleanObj(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIndexAssign(obj, index, value object.Object) object.Object {
	switch o := obj.(type) {
	case *object.Array:
		i := index.(*object.Integer).Value
		if i < 0 || i >= int64(len(o.Elements)) {
			return newError("array index out of bounds: %d", i)
		}
		o.Elements[i] = value
		return value
	case *object.Hash:
		hashable, ok := index.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", index.Type())
		}
		o.Pairs[hashable.HashKey()] = object.HashPair{Key: index, Value: value}
		return value
	default:
		return newError("index assignment not supported for %s", obj.Type())
	}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndex(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndex(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndex(array, index object.Object) object.Object {
	arr := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arr.Elements))

	if i < 0 {
		i = max + i
	}
	if i < 0 || i >= max {
		return NIL
	}
	return arr.Elements[i]
}

func evalHashIndex(hash, index object.Object) object.Object {
	h := hash.(*object.Hash)

	hashable, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := h.Pairs[hashable.HashKey()]
	if !ok {
		return NIL
	}
	return pair.Value
}

func evalStringIndex(str, index object.Object) object.Object {
	s := str.(*object.String).Value
	i := index.(*object.Integer).Value

	if i < 0 || i >= int64(len(s)) {
		return NIL
	}
	return &object.String{Value: string(s[i])}
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *Function:
		if len(args) != len(fn.Parameters) {
			return newError("%s: expected %d arguments, got %d", fn.Name, len(fn.Parameters), len(args))
		}
		extendedEnv := extendFunctionEnv(fn, args)
		result := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(result)
	case *object.Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NIL
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*ReturnValue); ok {
		return rv.Value
	}
	return obj
}

func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for _, keyNode := range node.Keys {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashable, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(node.Pairs[keyNode], env)
		if isError(value) {
			return value
		}

		pairs[hashable.HashKey()] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

package vm

import (
	"fmt"
	"math"

	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/compiler"
	"github.com/hilthontt/lotus/object"
)

// StackSize is an integer defining the size of our stack.
const StackSize = 2048

// MaxFrames defines the maximum frames allowed in the VM.
const MaxFrames = 1024

// GlobalsSize is the upper limit on the number of global bindings our VM supports.
const GlobalsSize = 65536

// True - pointer to a Lotus object.Boolean with value true.
var True = &object.Boolean{Value: true}

// False - pointer to a Lotus object.Boolean with value false.
var False = &object.Boolean{Value: false}

// Nil - pointer to a Lotus object.Nil.
var Nil = &object.Nil{}

// VM defines our Virtual Machine.
type VM struct {
	constants     []object.Object
	stack         []object.Object
	sp            int // Stack pointer: always points to the next free slot in the stack.
	globals       []object.Object
	frames        []*Frame
	framesIndex   int
	maxFramesUsed int
}

// New initializes and returns a pointer to a VM.
func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// NewWithGlobalsState creates a new VM with a compiler's bytecode and pre-existing globals (used in REPL).
func NewWithGlobalsState(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
	vm.maxFramesUsed++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// LastPoppedStackElement returns the last popped element on the top of the stack.
func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

// Run runs our VM and starts the fetch-decode-execute cycle.
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}

		case code.OpNil:
			if err := vm.push(Nil); err != nil {
				return err
			}

		// --- Arithmetic ---

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}

		case code.OpNegate:
			if err := vm.executeNegateOperator(); err != nil {
				return err
			}

		case code.OpNot:
			if err := vm.executeNotOperator(); err != nil {
				return err
			}

		case code.OpPlusPlus, code.OpMinusMinus:
			if err := vm.executePostfixOperator(op); err != nil {
				return err
			}

		// --- Comparison ---

		case code.OpEqual, code.OpNotEqual, code.OpGreater, code.OpGreaterEq:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		// --- Logical (short-circuit values handled by compiler jumps) ---

		// --- Jumps ---

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpJumpFalse:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case code.OpLoop:
			offset := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip -= offset

		// --- Globals ---

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			if err := vm.push(vm.globals[globalIndex]); err != nil {
				return err
			}

		// --- Locals ---

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			vm.stack[vm.currentFrame().basePointer+int(localIndex)] = vm.pop()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			if err := vm.push(vm.stack[vm.currentFrame().basePointer+int(localIndex)]); err != nil {
				return err
			}

		// --- Free variables ---

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			currentClosure := vm.currentFrame().closure
			if err := vm.push(currentClosure.Free[freeIndex]); err != nil {
				return err
			}

		// --- Builtins ---

		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			definition := compiler.Builtins[builtinIndex]
			builtin := &object.Builtin{Name: definition.Name, Fn: definition.Fn}
			if err := vm.push(builtin); err != nil {
				return err
			}

		// --- Data Structures ---

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp -= numElements

			if err := vm.push(array); err != nil {
				return err
			}

		case code.OpMap:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			m, err := vm.buildMap(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp -= numElements

			if err := vm.push(m); err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			if err := vm.executeIndexExpr(left, index); err != nil {
				return err
			}

		case code.OpIndexSet:
			value := vm.pop()
			index := vm.pop()
			obj := vm.pop()

			if err := vm.executeIndexSet(obj, index, value); err != nil {
				return err
			}

		// --- Functions & Closures ---

		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			if err := vm.pushClosure(int(constIndex), int(numFree)); err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			if err := vm.executeCall(int(numArgs)); err != nil {
				return err
			}

		// OpReturn and OpReturnNil both respect initInstance so that
		// Foo(...) returns the newly-created instance rather than init's nil.
		case code.OpReturn:
			returnValue := vm.pop()
			frame := vm.popFrame()
			if frame.isMethod {
				vm.sp = frame.basePointer
			} else {
				vm.sp = frame.basePointer - 1
			}
			if frame.initInstance != nil {
				if err := vm.push(frame.initInstance); err != nil {
					return err
				}
			} else {
				if err := vm.push(returnValue); err != nil {
					return err
				}
			}

		case code.OpReturnNil:
			frame := vm.popFrame()
			if frame.isMethod {
				vm.sp = frame.basePointer
			} else {
				vm.sp = frame.basePointer - 1
			}
			if frame.initInstance != nil {
				if err := vm.push(frame.initInstance); err != nil {
					return err
				}
			} else {
				if err := vm.push(Nil); err != nil {
					return err
				}
			}

		// --- OOP opcodes ---

		// OpNewClass creates an empty Class and pushes it.
		case code.OpNewClass:
			nameIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			name := vm.constants[nameIdx].(*object.String).Value
			class := &object.Class{
				Name:    name,
				Methods: make(map[string]*object.Closure),
			}
			if err := vm.push(class); err != nil {
				return err
			}

		// OpSetSuper pops super then class, links them, pushes class back.
		case code.OpSetSuper:
			super := vm.pop()
			cls := vm.pop()
			class, ok := cls.(*object.Class)
			if !ok {
				return fmt.Errorf("OpSetSuper: expected class, got %s", cls.Type())
			}
			superClass, ok := super.(*object.Class)
			if !ok {
				return fmt.Errorf("superclass must be a class, got %s", super.Type())
			}
			class.SuperClass = superClass
			if err := vm.push(class); err != nil {
				return err
			}

		// OpDefineMethod pops closure then class, registers the method (and sets
		// DefiningClass on the closure for super resolution), pushes class back.
		case code.OpDefineMethod:
			nameIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			methodName := vm.constants[nameIdx].(*object.String).Value
			closure := vm.pop().(*object.Closure)
			cls := vm.pop()
			class, ok := cls.(*object.Class)
			if !ok {
				return fmt.Errorf("OpDefineMethod: expected class, got %s", cls.Type())
			}
			closure.DefiningClass = class // enables 'super' inside this method
			class.Methods[methodName] = closure
			if err := vm.push(class); err != nil {
				return err
			}

		// OpGetField pops an instance and pushes the value of the named field.
		// Falls back to nil if neither a field nor a method with that name exists.
		case code.OpGetField:
			nameIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			fieldName := vm.constants[nameIdx].(*object.String).Value
			obj := vm.pop()
			if err := vm.executeGetField(obj, fieldName); err != nil {
				return err
			}

		// OpSetField pops (value, instance) and stores the field.
		case code.OpSetField:
			nameIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			fieldName := vm.constants[nameIdx].(*object.String).Value
			value := vm.pop()
			obj := vm.pop()
			if err := vm.executeSetField(obj, fieldName, value); err != nil {
				return err
			}

		// OpInvokeMethod: stack is [..., receiver, arg1..argN].
		// For a plain Instance, receiver == self.
		// For a SuperAccessor, we replace it with the real self before calling.
		case code.OpInvokeMethod:
			nameIdx := code.ReadUint16(ins[ip+1:])
			numArgs := int(code.ReadUint8(ins[ip+3:]))
			vm.currentFrame().ip += 3
			methodName := vm.constants[nameIdx].(*object.String).Value
			if err := vm.executeInvokeMethod(methodName, numArgs); err != nil {
				return err
			}

		// OpGetSuper pushes a SuperAccessor for the current method's defining class.
		case code.OpGetSuper:
			frame := vm.currentFrame()
			selfObj := vm.stack[frame.basePointer]
			instance, ok := selfObj.(*object.Instance)
			if !ok {
				return fmt.Errorf("'super' used in a non-method context")
			}
			defClass := frame.closure.DefiningClass
			if defClass == nil {
				return fmt.Errorf("'super' used in a closure that is not a method")
			}
			if defClass.SuperClass == nil {
				return fmt.Errorf("'super' used in '%s' which has no superclass", defClass.Name)
			}
			if err := vm.push(&object.SuperAccessor{Self: instance, SuperClass: defClass.SuperClass}); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}

	return nil
}

func (vm *VM) executeGetField(obj object.Object, name string) error {
	switch o := obj.(type) {
	case *object.Instance:
		if val, ok := o.Fields[name]; ok {
			return vm.push(val)
		}
		// Fallback: expose method as an unbound closure value
		if method, ok := o.Class.LookupMethod(name); ok {
			return vm.push(method)
		}
		return vm.push(Nil)
	default:
		return fmt.Errorf("field access on non-instance (%s)", obj.Type())
	}
}

func (vm *VM) executeSetField(obj object.Object, name string, value object.Object) error {
	instance, ok := obj.(*object.Instance)
	if !ok {
		return fmt.Errorf("field assignment on non-instance (%s)", obj.Type())
	}
	instance.Fields[name] = value
	return nil
}

// executeInvokeMethod dispatches obj.method(args) or super.method(args).
// The receiver sits at stack[sp-numArgs-1]; args are above it.
func (vm *VM) executeInvokeMethod(methodName string, numArgs int) error {
	receiver := vm.stack[vm.sp-numArgs-1]

	switch r := receiver.(type) {
	case *object.Instance:
		method, ok := r.Class.LookupMethod(methodName)
		if !ok {
			return fmt.Errorf("undefined method '%s' on class '%s'", methodName, r.Class.Name)
		}
		return vm.callMethod(method, numArgs)

	case *object.SuperAccessor:
		method, ok := r.SuperClass.LookupMethod(methodName)
		if !ok {
			return fmt.Errorf("undefined method '%s' in superclass of '%s'",
				methodName, r.SuperClass.Name)
		}
		// Replace SuperAccessor with the real self
		vm.stack[vm.sp-numArgs-1] = r.Self
		return vm.callMethod(method, numArgs)

	default:
		return fmt.Errorf("method call on non-instance (%s)", receiver.Type())
	}
}

// callClass handles Foo(args): creates an Instance, calls init if present,
// and ensures the instance (not nil) is left on the stack.
func (vm *VM) callClass(class *object.Class, numArgs int) error {
	instance := &object.Instance{
		Class:  class,
		Fields: make(map[string]object.Object),
	}

	initMethod, hasInit := class.LookupMethod("init")
	if !hasInit {
		if numArgs != 0 {
			return fmt.Errorf("'%s' takes no arguments (no 'init' defined)", class.Name)
		}
		vm.sp-- // remove the class object
		return vm.push(instance)
	}

	// Replace class slot with instance — it becomes self (local 0)
	vm.stack[vm.sp-numArgs-1] = instance
	if err := vm.callMethod(initMethod, numArgs); err != nil {
		return err
	}
	vm.currentFrame().initInstance = instance
	return nil
}

func (vm *VM) callMethod(cl *object.Closure, numArgs int) error {
	// numArgs does NOT include self — self is already on the stack below the args.
	// Total params = numArgs + 1 (self). basePointer points at self.
	totalArgs := numArgs + 1
	if totalArgs != cl.Fn.NumParams {
		return fmt.Errorf("%s: expected %d arguments, got %d",
			cl.Fn.Name, cl.Fn.NumParams-1, numArgs)
	}
	basePointer := vm.sp - totalArgs
	if vm.framesIndex < vm.maxFramesUsed {
		f := vm.frames[vm.framesIndex]
		f.closure = cl
		f.ip = -1
		f.basePointer = basePointer
		f.initInstance = nil
		f.isMethod = true
		vm.framesIndex++
	} else {
		f := NewFrame(cl, basePointer)
		f.isMethod = true
		vm.pushFrame(f)
	}
	vm.sp = basePointer + cl.Fn.NumLocals
	return nil
}

// --- Stack operations ---

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

// --- Binary operations ---

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.FLOAT_OBJ || rightType == object.FLOAT_OBJ:
		return vm.executeBinaryFloatOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftValue / rightValue
	case code.OpMod:
		if rightValue == 0 {
			return fmt.Errorf("modulo by zero")
		}
		result = leftValue % rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryFloatOperation(op code.Opcode, left, right object.Object) error {
	leftValue := toFloat(left)
	rightValue := toFloat(right)

	var result float64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftValue / rightValue
	case code.OpMod:
		result = math.Mod(leftValue, rightValue)
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&object.Float{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

// --- Comparison ---

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	// Numeric comparison (int or float)
	if isNumeric(left) && isNumeric(right) {
		return vm.executeNumericComparison(op, left, right)
	}

	// String comparison
	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeNumericComparison(op code.Opcode, left, right object.Object) error {
	leftVal := toFloat(left)
	rightVal := toFloat(right)

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal != rightVal))
	case code.OpGreater:
		return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
	case code.OpGreaterEq:
		return vm.push(nativeBoolToBooleanObj(leftVal >= rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeStringComparison(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal != rightVal))
	case code.OpGreater:
		return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
	case code.OpGreaterEq:
		return vm.push(nativeBoolToBooleanObj(leftVal >= rightVal))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

// --- Unary operators ---

func (vm *VM) executeNegateOperator() error {
	operand := vm.pop()

	switch o := operand.(type) {
	case *object.Integer:
		return vm.push(&object.Integer{Value: -o.Value})
	case *object.Float:
		return vm.push(&object.Float{Value: -o.Value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) executeNotOperator() error {
	operand := vm.pop()
	return vm.push(nativeBoolToBooleanObj(!isTruthy(operand)))
}

func (vm *VM) executePostfixOperator(op code.Opcode) error {
	operand := vm.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("postfix operator requires integer, got %s", operand.Type())
	}

	val := operand.(*object.Integer).Value
	if op == code.OpPlusPlus {
		val++
	} else {
		val--
	}

	return vm.push(&object.Integer{Value: val})
}

// --- Index operations ---

func (vm *VM) executeIndexExpr(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.MAP_OBJ:
		return vm.executeMapIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeMapIndex(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeStringIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements))

	// Negative indexing
	if i < 0 {
		i = max + i
	}
	if i < 0 || i >= max {
		return vm.push(Nil)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeMapIndex(m, index object.Object) error {
	mapObject := m.(*object.Hash)

	hashable, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as a hash key: %s", index.Type())
	}

	pair, ok := mapObject.Pairs[hashable.HashKey()]
	if !ok {
		return vm.push(Nil)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeStringIndex(str, index object.Object) error {
	s := str.(*object.String).Value
	i := index.(*object.Integer).Value

	if i < 0 || i >= int64(len(s)) {
		return vm.push(Nil)
	}

	return vm.push(&object.String{Value: string(s[i])})
}

func (vm *VM) executeIndexSet(obj, index, value object.Object) error {
	switch o := obj.(type) {
	case *object.Array:
		i := index.(*object.Integer).Value
		if i < 0 || i >= int64(len(o.Elements)) {
			return fmt.Errorf("array index out of bounds: %d", i)
		}
		o.Elements[i] = value
		return nil

	case *object.Hash:
		hashable, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("unusable as a hash key: %s", index.Type())
		}
		hk := hashable.HashKey()
		o.Pairs[hk] = object.HashPair{Key: index, Value: value}
		return nil

	default:
		return fmt.Errorf("index assignment not supported for %s", obj.Type())
	}
}

// --- Data structure builders ---

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildMap(startIndex, endIndex int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		hashable, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as a hash key: %s", key.Type())
		}

		pairs[hashable.HashKey()] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}, nil
}

// --- Function calls ---

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	case *object.Class:
		return vm.callClass(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function: %s", callee.Type())
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParams {
		return fmt.Errorf("%s: expected %d arguments, got %d",
			cl.Fn.Name, cl.Fn.NumParams, numArgs)
	}

	// Reuse existing frame slot if possible to reduce allocations
	if vm.framesIndex < vm.maxFramesUsed {
		vm.frames[vm.framesIndex].basePointer = vm.sp - numArgs
		vm.frames[vm.framesIndex].ip = -1
		vm.frames[vm.framesIndex].closure = cl
		vm.frames[vm.framesIndex].initInstance = nil // reset
		vm.framesIndex++
		vm.sp = vm.sp - numArgs + cl.Fn.NumLocals
	} else {
		frame := NewFrame(cl, vm.sp-numArgs)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + cl.Fn.NumLocals
	}
	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		return vm.push(result)
	}
	return vm.push(Nil)
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]

	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp -= numFree

	return vm.push(&object.Closure{Fn: function, Free: free})
}

// --- Helpers ---

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

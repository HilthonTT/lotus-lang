package vm

import (
	"fmt"
	"math"
	"os"
	"sync"

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

// ModuleLoader compiles and runs a .lotus file, returning its exported values.
// Provided by main.go to keep parsing/lexing out of the VM.
type ModuleLoader func(path string) (*object.Module, error)

// VM defines our Virtual Machine.
type VM struct {
	constants     []object.Object
	stack         []object.Object
	sp            int // Stack pointer: always points to the next free slot in the stack.
	globals       []object.Object
	globalsMu     sync.RWMutex
	frames        []*Frame
	framesIndex   int
	maxFramesUsed int
	loader        ModuleLoader
	moduleCache   map[string]*object.Module
	catchStack    []catchEntry
}

// New initializes and returns a pointer to a VM.
func New(bytecode *compiler.Bytecode) *VM {
	return NewWithLoader(bytecode, nil)
}

func NewWithLoader(bytecode *compiler.Bytecode, loader ModuleLoader) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	globals := make([]object.Object, GlobalsSize)

	seedCompiler := compiler.New()
	for _, name := range compiler.BuiltinPackageOrder {
		pkg := compiler.BuiltinPackages[name]
		sym, ok := seedCompiler.PublicResolve(name)
		if ok {
			globals[sym.Index] = pkg
		}
	}

	machine := &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     globals,
		frames:      frames,
		framesIndex: 1,
		loader:      loader,
		moduleCache: make(map[string]*object.Module),
	}

	for _, name := range []string{"Task", "Array", "String"} {
		sym, ok := seedCompiler.PublicResolve(name)
		if ok {
			if pkg, ok := globals[sym.Index].(*object.Package); ok {
				p := pkg
				p.CallVM = func(closure *object.Closure, args []object.Object) object.Object {
					return machine.callClosureSync(closure, args)
				}
			}
		}
	}

	return machine
}

// callClosureSync calls a Lotus closure synchronously and returns its result.
// Used by native packages (e.g. Task) to invoke Lotus handler functions.
func (vm *VM) callClosureSync(cl *object.Closure, args []object.Object) object.Object {
	fresh := &VM{
		constants:   vm.constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     vm.globals,
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 2, // 0=dummy, 1=worker
		loader:      vm.loader,
		moduleCache: vm.moduleCache,
	}

	// Frame 0: dummy with empty instructions.
	// When the worker returns, framesIndex drops to 1, currentFrame() = frames[0].
	// Its ip=-1, len(instructions)-1=-1, so -1 < -1 is false → Run() exits cleanly.
	dummyFn := &object.CompiledFunction{Instructions: code.Instructions{}}
	fresh.frames[0] = NewFrame(&object.Closure{Fn: dummyFn}, 0)

	// Frame 1: the actual worker closure.
	// Stack: [dummy_slot, arg0, arg1, ...]
	// basePointer=1 so OpReturn sets sp = 1-1 = 0, not -1.
	fresh.stack[0] = Nil
	for i, arg := range args {
		fresh.stack[1+i] = arg
	}
	fresh.frames[1] = &Frame{
		closure:     cl,
		ip:          -1,
		basePointer: 1,
		constants:   cl.Constants,
	}
	fresh.sp = 1 + cl.Fn.NumLocals

	if err := fresh.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "task error: %s\n", err)
		return Nil
	}

	// Return value was pushed at stack[sp-1] by OpReturn/OpReturnNil
	if fresh.sp > 0 {
		return fresh.stack[fresh.sp-1]
	}
	return Nil
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

func (vm *VM) GetGlobal(index int) object.Object {
	return vm.globals[index]
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
			if err := vm.push(vm.currentConstants()[constIndex]); err != nil {
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
				if !vm.redirectToCatch(err) {
					return err
				}
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
			vm.globalsMu.Lock()
			vm.globals[globalIndex] = vm.pop()
			vm.globalsMu.Unlock()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globalsMu.RLock()
			val := vm.globals[globalIndex]
			vm.globalsMu.RUnlock()
			if err := vm.push(val); err != nil {
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
			frame := vm.currentFrame() // get frame BEFORE popping
			vm.runDeferred(frame)      // run deferred closures
			vm.popFrame()              // now pop
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
			frame := vm.currentFrame()
			vm.runDeferred(frame) // run deferred closures
			vm.popFrame()
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
			name := vm.currentConstants()[nameIdx].(*object.String).Value
			class := &object.Class{Name: name, Methods: make(map[string]*object.Closure)}
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
			methodName := vm.currentConstants()[nameIdx].(*object.String).Value
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
			fieldName := vm.currentConstants()[nameIdx].(*object.String).Value
			obj := vm.pop()
			if err := vm.executeGetField(obj, fieldName); err != nil {
				return err
			}

			// OpSetField pops (value, instance) and stores the field.
		case code.OpSetField:
			nameIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			fieldName := vm.currentConstants()[nameIdx].(*object.String).Value
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
			methodName := vm.currentConstants()[nameIdx].(*object.String).Value
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

		case code.OpRunModule:
			pathIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			path := vm.currentConstants()[pathIdx].(*object.String).Value

			if vm.loader == nil {
				return fmt.Errorf("import of %q failed: no module loader configured", path)
			}

			// Check cache first
			if mod, ok := vm.moduleCache[path]; ok {
				if err := vm.push(mod); err != nil {
					return err
				}
				break
			}

			mod, err := vm.loader(path)
			if err != nil {
				return fmt.Errorf("import %q: %w", path, err)
			}
			vm.moduleCache[path] = mod
			if err := vm.push(mod); err != nil {
				return err
			}

		case code.OpDup:
			top := vm.stack[vm.sp-1]
			if err := vm.push(top); err != nil {
				return err
			}

		case code.OpBitAnd, code.OpBitOr, code.OpBitXor, code.OpLShift, code.OpRShift:
			if err := vm.executeBitwiseOperation(op); err != nil {
				return err
			}

		case code.OpBitNot:
			operand := vm.pop()
			i, ok := operand.(*object.Integer)
			if !ok {
				return fmt.Errorf("~ requires integer, got %s", operand.Type())
			}
			if err := vm.push(&object.Integer{Value: ^i.Value}); err != nil {
				return err
			}

		case code.OpDefer:
			// Pop the closure and add it to the current frame's deferred list.
			cl, ok := vm.pop().(*object.Closure)
			if !ok {
				return fmt.Errorf("defer: expected closure on stack")
			}
			vm.currentFrame().deferred = append(vm.currentFrame().deferred, cl)

		case code.OpThrow:
			val := vm.pop()
			var msg string
			switch v := val.(type) {
			case *object.String:
				msg = v.Value
			case *object.LotusError:
				msg = v.Message
			default:
				msg = val.Inspect()
			}
			throwErr := fmt.Errorf("%s", msg)
			if !vm.redirectToCatch(throwErr) {
				return throwErr
			}

		case code.OpTryBegin:
			catchOffset := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			vm.catchStack = append(vm.catchStack, catchEntry{
				frameIndex: vm.framesIndex,
				ip:         catchOffset,
				sp:         vm.sp,
			})

		case code.OpTryEnd:
			if len(vm.catchStack) > 0 {
				vm.catchStack = vm.catchStack[:len(vm.catchStack)-1]
			}

		case code.OpDefineInterface:
			// The interface object is already compiled as a constant and stored via
			// OpConstant + OpSetGlobal by the compiler. This opcode is reserved for
			// future use (e.g. runtime interface checks).
			vm.currentFrame().ip += 2

		case code.OpArraySliceFrom:
			start := vm.pop()
			arr := vm.pop()
			array, ok1 := arr.(*object.Array)
			idx, ok2 := start.(*object.Integer)
			if !ok1 || !ok2 {
				return fmt.Errorf("OpArraySliceFrom: expected array and int, got %s and %s", arr.Type(), start.Type())
			}
			from := min(max(int(idx.Value), 0), len(array.Elements))
			sliced := make([]object.Object, len(array.Elements)-from)
			copy(sliced, array.Elements[from:])
			if err := vm.push(&object.Array{Elements: sliced}); err != nil {
				return err
			}

		case code.OpSpread:
			val := vm.pop()
			arr, ok := val.(*object.Array)
			if !ok {
				// Non-array: wrap as single-element spread
				arr = &object.Array{Elements: []object.Object{val}}
			}
			if err := vm.push(&object.SpreadValue{Elements: arr.Elements}); err != nil {
				return err
			}

		case code.OpSpreadCall:
			numArgs := int(code.ReadUint8(ins[ip+1:]))
			vm.currentFrame().ip++

			// Collect raw args from stack (may contain SpreadValues)
			rawArgs := make([]object.Object, numArgs)
			for i := numArgs - 1; i >= 0; i-- {
				rawArgs[i] = vm.pop()
			}

			// Flatten any SpreadValues
			var flatArgs []object.Object
			for _, arg := range rawArgs {
				if spread, ok := arg.(*object.SpreadValue); ok {
					flatArgs = append(flatArgs, spread.Elements...)
				} else {
					flatArgs = append(flatArgs, arg)
				}
			}

			// Get callee (now at top of stack)
			callee := vm.pop()

			// Push callee and flat args back
			if err := vm.push(callee); err != nil {
				return err
			}
			for _, arg := range flatArgs {
				if err := vm.push(arg); err != nil {
					return err
				}
			}

			if err := vm.executeCall(len(flatArgs)); err != nil {
				if !vm.redirectToCatch(err) {
					return err
				}
			}

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}

	return nil
}

func (vm *VM) executeBitwiseOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	l, ok1 := left.(*object.Integer)
	r, ok2 := right.(*object.Integer)
	if !ok1 || !ok2 {
		return fmt.Errorf("bitwise operators require integers, got %s and %s", left.Type(), right.Type())
	}
	var result int64
	switch op {
	case code.OpBitAnd:
		result = l.Value & r.Value
	case code.OpBitOr:
		result = l.Value | r.Value
	case code.OpBitXor:
		result = l.Value ^ r.Value
	case code.OpLShift:
		result = l.Value << uint(r.Value)
	case code.OpRShift:
		result = l.Value >> uint(r.Value)
	}
	return vm.push(&object.Integer{Value: result})
}

// currentConstants returns the constants pool for the currently executing frame.
// Closures imported from other modules carry their own pool.
func (vm *VM) currentConstants() []object.Object {
	if c := vm.currentFrame().constants; c != nil {
		return c
	}
	return vm.constants
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

	case *object.Package:
		fn, ok := o.Functions[name]
		if !ok {
			return fmt.Errorf("package '%s' has no member '%s'", o.Name, name)
		}

		// Wrap as a Builtin so OpCall handles it normally
		return vm.push(&object.Builtin{
			Name: o.Name + "." + name,
			Fn:   object.BuiltinFunction(fn),
		})

	case *object.Module:
		val, ok := o.Exports[name]
		if !ok {
			return fmt.Errorf("module %q has no export '%s'", o.Path, name)
		}
		return vm.push(val)

	case *object.EnumDef:
		varDef, ok := o.Variants[name]
		if !ok {
			return fmt.Errorf("enum '%s' has no variant '%s'", o.Name, name)
		}
		if len(varDef.Fields) == 0 {
			return vm.push(&object.EnumVariant{EnumName: o.Name, VariantName: name})
		}
		// Variant with fields — return a constructor builtin
		enumName, variantName, fields := o.Name, name, varDef.Fields
		return vm.push(&object.Builtin{
			Name: o.Name + "." + name,
			Fn: func(args ...object.Object) object.Object {
				if len(args) != len(fields) {
					return &object.Nil{}
				}
				data := make(map[string]object.Object)
				for i, f := range fields {
					data[f] = args[i]
				}
				return &object.EnumVariant{EnumName: enumName, VariantName: variantName, Data: data}
			},
		})

	case *object.EnumVariant:
		// Access data field on a data-carrying variant: shape.radius
		return vm.push(o.GetField(name))

	case *object.Result:
		switch name {
		case "ok":
			return vm.push(&object.Boolean{Value: o.Ok})
		case "value":
			return vm.push(o.Value)
		case "error":
			return vm.push(&object.String{Value: o.ErrMsg})
		default:
			return fmt.Errorf("Result has no field '%s'", name)
		}

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

	case *object.Package:
		fn, ok := r.Functions[methodName]
		if !ok {
			return fmt.Errorf("package '%s' has no member '%s'", r.Name, methodName)
		}
		// Collect args (skip the package receiver slot), call, clean up
		args := make([]object.Object, numArgs)
		for i := range numArgs {
			args[i] = vm.stack[vm.sp-numArgs+i]
		}
		vm.sp -= numArgs + 1 // pop args + package receiver
		result := fn(args...)
		if result == nil {
			result = Nil
		}
		return vm.push(result)

	case *object.EnumDef:
		varDef, ok := r.Variants[methodName]
		if !ok {
			return fmt.Errorf("enum '%s' has no variant '%s'", r.Name, methodName)
		}
		args := make([]object.Object, numArgs)
		for i := range numArgs {
			args[i] = vm.stack[vm.sp-numArgs+i]
		}
		vm.sp -= numArgs + 1 // pop args + enum receiver

		if len(varDef.Fields) == 0 {
			return vm.push(&object.EnumVariant{EnumName: r.Name, VariantName: methodName})
		}
		if len(args) != len(varDef.Fields) {
			return fmt.Errorf("enum variant '%s.%s' expects %d args, got %d",
				r.Name, methodName, len(varDef.Fields), numArgs)
		}
		data := make(map[string]object.Object)
		for i, f := range varDef.Fields {
			data[f] = args[i]
		}
		return vm.push(&object.EnumVariant{
			EnumName:    r.Name,
			VariantName: methodName,
			Data:        data,
		})

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
		f.constants = cl.Constants // propagate
		f.deferred = nil
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

	// String + anything → auto-convert for interpolation
	if op == code.OpAdd && (leftType == object.STRING_OBJ || rightType == object.STRING_OBJ) {
		return vm.push(&object.String{Value: left.Inspect() + right.Inspect()})
	}

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

	if left.Type() == object.ENUM_VARIANT_OBJ && right.Type() == object.ENUM_VARIANT_OBJ {
		l := left.(*object.EnumVariant)
		r := right.(*object.EnumVariant)
		eq := l.EnumName == r.EnumName && l.VariantName == r.VariantName
		switch op {
		case code.OpEqual:
			return vm.push(nativeBoolToBooleanObj(eq))
		case code.OpNotEqual:
			return vm.push(nativeBoolToBooleanObj(!eq))
		}
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
	var elements []object.Object
	for i := startIndex; i < endIndex; i++ {
		if spread, ok := vm.stack[i].(*object.SpreadValue); ok {
			elements = append(elements, spread.Elements...)
		} else {
			elements = append(elements, vm.stack[i])
		}
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
	fn := cl.Fn

	if fn.IsVariadic {
		// Fixed params = fn.NumParams - 1 (last param is the rest array)
		fixedCount := fn.NumParams - 1
		if numArgs < fixedCount {
			return fmt.Errorf("%s: expected at least %d arguments, got %d",
				fn.Name, fixedCount, numArgs)
		}
		// Pack extra args into a rest array
		restCount := numArgs - fixedCount
		restElems := make([]object.Object, restCount)
		for i := range restCount {
			restElems[i] = vm.stack[vm.sp-restCount+i]
		}
		vm.sp -= restCount
		if err := vm.push(&object.Array{Elements: restElems}); err != nil {
			return err
		}
		numArgs = fn.NumParams // adjusted to match exactly
	} else {
		if numArgs != fn.NumParams {
			return fmt.Errorf("%s: expected %d arguments, got %d",
				fn.Name, fn.NumParams, numArgs)
		}
	}

	if vm.framesIndex < vm.maxFramesUsed {
		f := vm.frames[vm.framesIndex]
		f.basePointer = vm.sp - numArgs
		f.ip = -1
		f.closure = cl
		f.initInstance = nil
		f.isMethod = false
		f.constants = cl.Constants
		f.deferred = nil
		vm.framesIndex++
		vm.sp = vm.sp - numArgs + fn.NumLocals
	} else {
		frame := NewFrame(cl, vm.sp-numArgs)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + fn.NumLocals
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
	constant := vm.currentConstants()[constIndex]

	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp -= numFree

	return vm.push(&object.Closure{
		Fn:        function,
		Free:      free,
		Constants: vm.currentConstants(), // carry this pool into the closure
	})
}

// redirectToCatch attempts to redirect execution to a registered catch handler.
// Returns true if a handler was found and execution was redirected.
// Returns false if the error should propagate normally.
func (vm *VM) redirectToCatch(err error) bool {
	if len(vm.catchStack) == 0 {
		return false
	}
	entry := vm.catchStack[len(vm.catchStack)-1]
	vm.catchStack = vm.catchStack[:len(vm.catchStack)-1]

	// Unwind frames back to the try block's frame
	vm.framesIndex = entry.frameIndex
	vm.sp = entry.sp

	// Push the error message as a string (the catch variable receives it)
	vm.push(&object.LotusError{Message: err.Error()})

	// Jump to catch handler (ip will be incremented at top of loop → entry.ip)
	vm.currentFrame().ip = entry.ip - 1
	return true
}

// runDeferred calls all deferred closures in the given frame in LIFO order.
// Called just before OpReturn / OpReturnNil pops the frame.
func (vm *VM) runDeferred(frame *Frame) {
	for i := len(frame.deferred) - 1; i >= 0; i-- {
		cl := frame.deferred[i]
		// Use a fresh mini-VM so deferred calls don't corrupt the main stack
		result := vm.callClosureSync(cl, []object.Object{})
		_ = result
	}
}

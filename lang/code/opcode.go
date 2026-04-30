package code

type Opcode byte

const (
	OpConstant Opcode = iota // Push constant from pool
	OpPop                    // Pop top of stack
	OpTrue                   // Push true
	OpFalse                  // Push false
	OpNil                    // Push nil

	// Arithmetic
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNegate // Unary minus

	// Postfix/unary
	OpPlusPlus
	OpMinusMinus

	// Comparison
	OpEqual
	OpNotEqual
	OpGreater
	OpGreaterEq
	OpIn // pops [left, right] → pushes bool: left in right

	// Logic
	OpNot

	// Jumps
	OpJump      // Unconditional
	OpJumpFalse // Jump if top is falsy (pops)

	// Variables
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetFree    // Get free (captured) variable
	OpGetBuiltin // Get builtin by index

	// Data structures
	OpArray    // Build array from N elements on stack
	OpMap      // Build map from N*2 elements on stack
	OpIndex    // Index operation
	OpIndexSet // stack: [obj, index, value] -> sets obj[index] = value

	// Functions
	OpClosure   // Create closure from function constant + free vars
	OpCall      // Call function with N args
	OpReturn    // Return from function with value
	OpReturnNil // Return nil from function

	// Loops
	OpLoop // Jump backward (for loops)

	// String concat
	OpConcat

	// Mutable closures
	OpSetFree // Set free (captured) variable

	// OpNewClass creates an empty class. Operand: const_idx of the name string.
	OpNewClass
	// OpSetSuper pops the superclass then the class, links them, pushes the class back.
	OpSetSuper
	// OpDefineMethod pops the closure then the class, adds the method, pushes the class back.
	// Operand: const_idx of the method name string.
	OpDefineMethod
	// OpGetField pops an instance and pushes the named field (or nil).
	// Operand: const_idx of the field name string.
	OpGetField
	// OpSetField pops (value, instance) and stores value into the named field.
	// Operand: const_idx of the field name string.
	OpSetField
	// OpInvokeMethod performs a method call: receiver is below the N args on the stack.
	// Operands: const_idx of method name (2 bytes), num_args (1 byte).
	OpInvokeMethod
	// OpGetSuper pushes a SuperAccessor for the current method's self/superclass.
	OpGetSuper

	// OpRunModule compiles and runs a .lotus file, pushes a Module object.
	// Operand: const_idx of the path string (2 bytes).
	OpRunModule

	// OpDup duplicates the top of stack (used to access module fields multiple times).
	OpDup

	OpBitAnd
	OpBitOr
	OpBitXor
	OpBitNot
	OpLShift
	OpRShift

	// Error handling & defer
	OpDefer    // pop closure -> add to frame.deferred
	OpThrow    // pop value -> raise error (check catch stack first)
	OpTryBegin // operand: 2-byte catch offset
	OpTryEnd   // pop catch handler

	// Interfaces
	OpDefineInterface // operand: const_idx of name string (2 bytes)

	// Spread / variadic
	// OpSpread: pops an array, marks it for spreading.
	// When OpSpreadCall sees a spread marker, it unpacks the array.
	OpSpread

	// OpSpreadCall: like OpCall but resolves spread markers on the stack.
	// Operand: total number of arguments (including spread args).
	OpSpreadCall

	// OpArraySliceFrom: pops [start_int, array] → pushes array[start:]
	OpArraySliceFrom
)

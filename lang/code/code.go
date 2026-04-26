package code

import (
	"encoding/binary"
	"fmt"
)

// opcodeComments provides short inline descriptions for the disassembler.
var opcodeComments = map[Opcode]string{
	OpConstant:       "push constant from pool",
	OpPop:            "discard top of stack",
	OpTrue:           "push true",
	OpFalse:          "push false",
	OpNil:            "push nil",
	OpAdd:            "integer/float addition",
	OpSub:            "subtraction",
	OpMul:            "multiplication",
	OpDiv:            "division",
	OpMod:            "modulo",
	OpNegate:         "unary minus",
	OpEqual:          "equality check",
	OpNotEqual:       "inequality check",
	OpGreater:        "greater-than comparison",
	OpGreaterEq:      "greater-than-or-equal comparison",
	OpNot:            "logical NOT",
	OpJump:           "unconditional jump",
	OpJumpFalse:      "jump if falsy (pops)",
	OpGetGlobal:      "load global variable",
	OpSetGlobal:      "store global variable",
	OpGetLocal:       "load local variable",
	OpSetLocal:       "store local variable",
	OpGetFree:        "load captured (free) variable",
	OpGetBuiltin:     "load built-in function by index",
	OpArray:          "build array from N stack elements",
	OpMap:            "build map from N*2 stack elements",
	OpIndex:          "index into array/map/string",
	OpIndexSet:       "assign to array/map index",
	OpClosure:        "create closure with free variables",
	OpCall:           "call function with N arguments",
	OpReturn:         "return value from function",
	OpReturnNil:      "return nil from function",
	OpLoop:           "jump backward (loop)",
	OpConcat:         "string concatenation",
	OpSetFree:        "update captured (free) variable",
	OpNewClass:       "create a new class object",
	OpSetSuper:       "pop super+class, link superclass, push class",
	OpDefineMethod:   "pop closure+class, add method to class, push class",
	OpGetField:       "pop instance, push named field value",
	OpSetField:       "pop value+instance, set named field",
	OpInvokeMethod:   "invoke named method with N args (self is receiver)",
	OpGetSuper:       "push super-accessor for current self",
	OpRunModule:      "compile+run module file, push module object",
	OpDup:            "duplicate top of stack",
	OpSpread:         "mark array for spreading into call",
	OpSpreadCall:     "call with spread arguments",
	OpArraySliceFrom: "slice array from index to end",
}

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

type Definition struct {
	Name          string
	OperandWidths []int // byte-width of each operand
}

var definitions = map[Opcode]*Definition{
	OpConstant:        {"OpConstant", []int{2}},
	OpPop:             {"OpPop", []int{}},
	OpTrue:            {"OpTrue", []int{}},
	OpFalse:           {"OpFalse", []int{}},
	OpNil:             {"OpNil", []int{}},
	OpAdd:             {"OpAdd", []int{}},
	OpSub:             {"OpSub", []int{}},
	OpMul:             {"OpMul", []int{}},
	OpDiv:             {"OpDiv", []int{}},
	OpMod:             {"OpMod", []int{}},
	OpNegate:          {"OpNegate", []int{}},
	OpEqual:           {"OpEqual", []int{}},
	OpNotEqual:        {"OpNotEqual", []int{}},
	OpGreater:         {"OpGreater", []int{}},
	OpGreaterEq:       {"OpGreaterEq", []int{}},
	OpNot:             {"OpNot", []int{}},
	OpJump:            {"OpJump", []int{2}},
	OpJumpFalse:       {"OpJumpFalse", []int{2}},
	OpGetGlobal:       {"OpGetGlobal", []int{2}},
	OpSetGlobal:       {"OpSetGlobal", []int{2}},
	OpGetLocal:        {"OpGetLocal", []int{1}},
	OpSetLocal:        {"OpSetLocal", []int{1}},
	OpGetFree:         {"OpGetFree", []int{1}},
	OpGetBuiltin:      {"OpGetBuiltin", []int{1}},
	OpArray:           {"OpArray", []int{2}},
	OpMap:             {"OpMap", []int{2}},
	OpIndex:           {"OpIndex", []int{}},
	OpIndexSet:        {"OpIndexSet", []int{}},
	OpClosure:         {"OpClosure", []int{2, 1}}, // const_idx, num_free
	OpCall:            {"OpCall", []int{1}},
	OpReturn:          {"OpReturn", []int{}},
	OpReturnNil:       {"OpReturnNil", []int{}},
	OpLoop:            {"OpLoop", []int{2}},
	OpConcat:          {"OpConcat", []int{}},
	OpSetFree:         {"OpSetFree", []int{1}},
	OpMinusMinus:      {"OpMinusMinus", []int{}},
	OpPlusPlus:        {"OpPlusPlus", []int{}},
	OpNewClass:        {"OpNewClass", []int{2}},
	OpSetSuper:        {"OpSetSuper", []int{}},
	OpDefineMethod:    {"OpDefineMethod", []int{2}},
	OpGetField:        {"OpGetField", []int{2}},
	OpSetField:        {"OpSetField", []int{2}},
	OpInvokeMethod:    {"OpInvokeMethod", []int{2, 1}},
	OpGetSuper:        {"OpGetSuper", []int{}},
	OpRunModule:       {"OpRunModule", []int{2}}, // Operand: const_idx of the path string (2 bytes).
	OpDup:             {"OpDup", []int{}},
	OpBitAnd:          {"OpBitAnd", []int{}},
	OpBitOr:           {"OpBitOr", []int{}},
	OpBitXor:          {"OpBitXor", []int{}},
	OpBitNot:          {"OpBitNot", []int{}},
	OpLShift:          {"OpLShift", []int{}},
	OpRShift:          {"OpRShift", []int{}},
	OpDefer:           {"OpDefer", []int{}},
	OpThrow:           {"OpThrow", []int{}},
	OpTryBegin:        {"OpTryBegin", []int{2}},
	OpTryEnd:          {"OpTryEnd", []int{}},
	OpDefineInterface: {"OpDefineInterface", []int{2}},
	OpSpread:          {"OpSpread", []int{}},
	OpSpreadCall:      {"OpSpreadCall", []int{1}},
	OpArraySliceFrom:  {"OpArraySliceFrom", []int{}},
}

// Lookup finds the Definition for a given opcode byte.
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make encodes an opcode and its operands into a bytecode instruction.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		if i >= len(def.OperandWidths) {
			break // ignore extra operands beyond what the definition expects
		}
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands decodes operands from an instruction stream given its definition.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

// ReadUint16 decodes a big-endian uint16 from an instruction stream.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// ReadUint8 decodes a uint8 from an instruction stream.
func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

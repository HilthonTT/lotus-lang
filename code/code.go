package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Instructions is a slice of bytecode bytes.
type Instructions []byte

// String returns a human-readable disassembly, one instruction per line.
func (ins Instructions) String() string {
	return Disassemble(ins)
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s", def.Name)
}

// Disassemble returns a human-readable disassembly of the given instructions.
func Disassemble(ins Instructions) string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "%04d ERROR: %s\n", i, err)
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

// DisassembleAnnotated returns a disassembly with inline comments describing each opcode.
func DisassembleAnnotated(ins Instructions) string {
	var out strings.Builder

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "%04d ERROR: %s\n", i, err)
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		line := fmt.Sprintf("%04d %s", i, ins.fmtInstruction(def, operands))

		if comment, ok := opcodeComments[Opcode(ins[i])]; ok {
			fmt.Fprintf(&out, "%-30s // %s\n", line, comment)
		} else {
			fmt.Fprintf(&out, "%s\n", line)
		}

		i += 1 + read
	}

	return out.String()
}

// opcodeComments provides short inline descriptions for the disassembler.
var opcodeComments = map[Opcode]string{
	OpConstant:   "push constant from pool",
	OpPop:        "discard top of stack",
	OpTrue:       "push true",
	OpFalse:      "push false",
	OpNil:        "push nil",
	OpAdd:        "integer/float addition",
	OpSub:        "subtraction",
	OpMul:        "multiplication",
	OpDiv:        "division",
	OpMod:        "modulo",
	OpNegate:     "unary minus",
	OpEqual:      "equality check",
	OpNotEqual:   "inequality check",
	OpGreater:    "greater-than comparison",
	OpGreaterEq:  "greater-than-or-equal comparison",
	OpNot:        "logical NOT",
	OpJump:       "unconditional jump",
	OpJumpFalse:  "jump if falsy (pops)",
	OpGetGlobal:  "load global variable",
	OpSetGlobal:  "store global variable",
	OpGetLocal:   "load local variable",
	OpSetLocal:   "store local variable",
	OpGetFree:    "load captured (free) variable",
	OpGetBuiltin: "load built-in function by index",
	OpArray:      "build array from N stack elements",
	OpMap:        "build map from N*2 stack elements",
	OpIndex:      "index into array/map/string",
	OpIndexSet:   "assign to array/map index",
	OpClosure:    "create closure with free variables",
	OpCall:       "call function with N arguments",
	OpReturn:     "return value from function",
	OpReturnNil:  "return nil from function",
	OpLoop:       "jump backward (loop)",
	OpConcat:     "string concatenation",
	OpSetFree:    "update captured (free) variable",
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
)

type Definition struct {
	Name          string
	OperandWidths []int // byte-width of each operand
}

var definitions = map[Opcode]*Definition{
	OpConstant:   {"OpConstant", []int{2}},
	OpPop:        {"OpPop", []int{}},
	OpTrue:       {"OpTrue", []int{}},
	OpFalse:      {"OpFalse", []int{}},
	OpNil:        {"OpNil", []int{}},
	OpAdd:        {"OpAdd", []int{}},
	OpSub:        {"OpSub", []int{}},
	OpMul:        {"OpMul", []int{}},
	OpDiv:        {"OpDiv", []int{}},
	OpMod:        {"OpMod", []int{}},
	OpNegate:     {"OpNegate", []int{}},
	OpEqual:      {"OpEqual", []int{}},
	OpNotEqual:   {"OpNotEqual", []int{}},
	OpGreater:    {"OpGreater", []int{}},
	OpGreaterEq:  {"OpGreaterEq", []int{}},
	OpNot:        {"OpNot", []int{}},
	OpJump:       {"OpJump", []int{2}},
	OpJumpFalse:  {"OpJumpFalse", []int{2}},
	OpGetGlobal:  {"OpGetGlobal", []int{2}},
	OpSetGlobal:  {"OpSetGlobal", []int{2}},
	OpGetLocal:   {"OpGetLocal", []int{1}},
	OpSetLocal:   {"OpSetLocal", []int{1}},
	OpGetFree:    {"OpGetFree", []int{1}},
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},
	OpArray:      {"OpArray", []int{2}},
	OpMap:        {"OpMap", []int{2}},
	OpIndex:      {"OpIndex", []int{}},
	OpIndexSet:   {"OpIndexSet", []int{}},
	OpClosure:    {"OpClosure", []int{2, 1}}, // const_idx, num_free
	OpCall:       {"OpCall", []int{1}},
	OpReturn:     {"OpReturn", []int{}},
	OpReturnNil:  {"OpReturnNil", []int{}},
	OpLoop:       {"OpLoop", []int{2}},
	OpConcat:     {"OpConcat", []int{}},
	OpSetFree:    {"OpSetFree", []int{1}},
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

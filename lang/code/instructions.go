package code

import (
	"bytes"
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

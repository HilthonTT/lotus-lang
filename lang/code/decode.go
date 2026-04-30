package code

import "encoding/binary"

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

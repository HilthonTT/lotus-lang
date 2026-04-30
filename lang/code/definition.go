package code

import "fmt"

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
	OpIn:              {"OpIn", []int{}},
}

// Lookup finds the Definition for a given opcode byte.
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

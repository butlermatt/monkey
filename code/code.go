package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type OpCode byte

const (
	OpConstant OpCode = iota
	OpPop

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpTrue
	OpFalse
	OpNull

	OpEqual
	OpNotEqual
	OpGreater
	OpGreaterEqual

	OpMinus
	OpBang

	OpJumpNotTrue
	OpJump

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpGetFree

	OpArray
	OpHash
	OpIndex

	OpCall
	OpReturn
	OpReturnValue

	OpClosure
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			_, _ = fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		_, _ = fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	count := len(def.OperandWidths)

	if len(operands) != count {
		return fmt.Sprintf("ERROR: operand length %d does not match defined %d\n", len(operands), count)
	}

	switch count {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount (%d) for %s\n", count, def.Name)
}

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},

	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},

	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	OpNull:  {"OpNull", []int{}},

	OpEqual:        {"OpEqual", []int{}},
	OpNotEqual:     {"OpNotEqual", []int{}},
	OpGreater:      {"OpGreater", []int{}},
	OpGreaterEqual: {"OpGreaterEqual", []int{}},

	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},

	OpJumpNotTrue: {"OpJumpNotTrue", []int{2}},
	OpJump:        {"OpJump", []int{2}},

	OpGetGlobal:  {"OpGetGlobal", []int{2}},
	OpSetGlobal:  {"OpSetGlobal", []int{2}},
	OpGetLocal:   {"OpGetLocal", []int{1}},
	OpSetLocal:   {"OpSetLocal", []int{1}},
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},
	OpGetFree:    {"OpGetFree", []int{1}},

	OpArray: {"OpArray", []int{2}},
	OpHash:  {"OpHash", []int{2}},
	OpIndex: {"OpIndex", []int{}},

	OpCall:        {"OpCall", []int{1}},
	OpReturn:      {"OpReturn", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},

	OpClosure: {"OpClosure", []int{2, 1}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpCode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

func Make(op OpCode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instLen := 1
	for _, w := range def.OperandWidths {
		instLen += w
	}

	instruction := make([]byte, instLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 1:
			instruction[offset] = byte(o)
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}

		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	oper := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			oper[i] = int(ReadUint8(ins[offset:]))
		case 2:
			oper[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return oper, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

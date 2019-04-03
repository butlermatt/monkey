package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("Error: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("Error: unhandled operandCount for %s\n", def.Name)
}

type Definition struct {
	Name       string
	OperWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
}

// Lookup accepts a byte op and returns the definition for that Opcode
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d is undefined", op)
	}

	return def, nil
}

// Make generates the required bytecode for a specific opCode with the number of operands required.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instLen := 1
	for _, w := range def.OperWidths {
		instLen += w
	}

	inst := make([]byte, instLen)
	inst[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(inst[offset:], uint16(o))
		}
		offset += width
	}

	return inst
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperWidths))
	offset := 0

	for i, width := range def.OperWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

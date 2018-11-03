package vm

import (
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int // Instruction pointer points to compiled function instruction
	bp int // Base Pointer points to position on stack immediately before calling function
}

func NewFrame(fn *object.CompiledFunction, base int) *Frame {
	return &Frame{fn: fn, ip: -1, bp: base}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

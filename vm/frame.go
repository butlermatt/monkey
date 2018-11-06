package vm

import (
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/object"
)

type Frame struct {
	cl *object.Closure
	ip int // Instruction pointer points to compiled function instruction
	bp int // Base Pointer points to position on stack immediately before calling function
}

func NewFrame(cl *object.Closure, base int) *Frame {
	return &Frame{cl: cl, ip: -1, bp: base}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}

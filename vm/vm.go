package vm

import (
	"fmt"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to _next_ value, top is always stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constInd := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constInd])
			if err != nil {
				return err
			}
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			leftVal := left.(*object.Number).Value
			rightVal := right.(*object.Number).Value

			res := leftVal + rightVal
			// Ignore error because we've already popped 2 off the stack, it should not be able to overflow now.
			_ = vm.push(&object.Number{Value: res})
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

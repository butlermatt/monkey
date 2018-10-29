package vm

import (
	"fmt"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the _next_ value. Top of stack is stack[sp-1]

	globals []object.Object
}

func New(bytecode *compiler.ByteCode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),
	}
}

func NewWithGlobalStore(bytecode *compiler.ByteCode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.OpCode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			ci := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[ci])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpGreater, code.OpGreaterEqual, code.OpEqual, code.OpNotEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpPop:
			_ = vm.pop()
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpJumpNotTrue:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
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

func (vm *VM) executeBinaryOperation(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	lType := left.Type()
	rType := right.Type()

	if lType == object.NumberObj && rType == object.NumberObj {
		return vm.executeBinaryNumberOperation(op, left, right)
	} else if lType == object.StringObj && rType == object.StringObj {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", lType, rType)
}

func (vm *VM) executeBinaryNumberOperation(op code.OpCode, left, right object.Object) error {
	lVal := left.(*object.Number).Value
	rVal := right.(*object.Number).Value

	var result float64
	switch op {
	case code.OpAdd:
		result = lVal + rVal
	case code.OpSub:
		result = lVal - rVal
	case code.OpMul:
		result = lVal * rVal
	case code.OpDiv:
		result = lVal / rVal
	default:
		return fmt.Errorf("unknown number operator: %d", op)
	}

	return vm.push(&object.Number{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.OpCode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	lval := left.(*object.String).Value
	rval := right.(*object.String).Value

	return vm.push(&object.String{Value: lval + rval})
}

func (vm *VM) executeComparison(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.NumberObj && right.Type() == object.NumberObj {
		return vm.executeNumberComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeNumberComparison(op code.OpCode, left, right object.Object) error {
	lVal := left.(*object.Number).Value
	rVal := right.(*object.Number).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToObject(rVal == lVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToObject(rVal != lVal))
	case code.OpGreater:
		return vm.push(nativeBoolToObject(lVal > rVal))
	case code.OpGreaterEqual:
		return vm.push(nativeBoolToObject(lVal >= rVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	if operand == False || operand == Null {
		return vm.push(True)
	}

	return vm.push(False)
}

func (vm *VM) executeMinusOperator() error {
	oper := vm.pop()

	if oper.Type() != object.NumberObj {
		return fmt.Errorf("unsupported type for negation: %s", oper.Type())
	}

	value := oper.(*object.Number).Value
	return vm.push(&object.Number{Value: -value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func nativeBoolToObject(b bool) *object.Boolean {
	if b {
		return True
	}
	return False
}

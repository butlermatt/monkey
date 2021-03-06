package vm

import (
	"fmt"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/object"
)

const MaxFrames = 1024
const StackSize = 2048
const GlobalsSize = 65536

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the _next_ value. Top of stack is stack[sp-1]

	globals []object.Object

	frames   []*Frame
	frameInd int
}

func New(bytecode *compiler.ByteCode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),

		frames:   frames,
		frameInd: 1,
	}
}

func NewWithGlobalStore(bytecode *compiler.ByteCode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameInd-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameInd] = f
	vm.frameInd++
}

func (vm *VM) popFrame() *Frame {
	vm.frameInd--
	return vm.frames[vm.frameInd]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	var ip *int
	var ins code.Instructions
	var op code.OpCode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		ip = &(vm.currentFrame().ip)
		*ip++
		ins = vm.currentFrame().Instructions()
		op = code.OpCode(ins[*ip])

		switch op {
		case code.OpConstant:
			ci := code.ReadUint16(ins[*ip+1:])
			vm.currentFrame().ip += 2
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
			pos := int(code.ReadUint16(ins[*ip+1:]))
			*ip = pos - 1
		case code.OpJumpNotTrue:
			pos := int(code.ReadUint16(ins[*ip+1:]))
			*ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				*ip = pos - 1
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[*ip+1:])
			*ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[*ip+1:])
			*ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numEls := int(code.ReadUint16(ins[*ip+1:]))
			*ip += 2

			array := vm.buildArray(vm.sp-numEls, vm.sp)
			vm.sp = vm.sp - numEls

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numEls := int(code.ReadUint16(ins[*ip+1:]))
			*ip += 2

			hash, err := vm.buildHash(vm.sp-numEls, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numEls

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			ind := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, ind)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[*ip+1:])
			*ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			val := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.bp - 1

			err := vm.push(val)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.bp - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localInd := code.ReadUint8(ins[*ip+1:])
			*ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.bp+int(localInd)] = vm.pop()
		case code.OpGetLocal:
			localInd := code.ReadUint8(ins[*ip+1:])
			*ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.bp+int(localInd)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			bInd := code.ReadUint8(ins[*ip+1:])
			*ip += 1
			def := object.Builtins[bInd]
			err := vm.push(def.Builtin)
			if err != nil {
				return err
			}
		case code.OpGetFree:
			fInd := code.ReadUint8(ins[*ip+1:])
			*ip += 1

			closure := vm.currentFrame().cl
			err := vm.push(closure.Free[fInd])
			if err != nil {
				return err
			}
		case code.OpClosure:
			cInd := code.ReadUint16(ins[*ip+1:])
			numFree := code.ReadUint8(ins[*ip+3:])
			*ip += 3

			err := vm.pushClosure(int(cInd), int(numFree))
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

func (vm *VM) pushClosure(cInd int, numFree int) error {
	constant := vm.constants[cInd]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, Free: free}
	return vm.push(closure)
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

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	}

	return fmt.Errorf("calling non-function and non-built-in")
}

func (vm *VM) buildArray(start, end int) object.Object {
	els := make([]object.Object, end-start)

	for i := start; i < end; i++ {
		els[i-start] = vm.stack[i]
	}

	return &object.Array{Elements: els}
}

func (vm *VM) buildHash(start, end int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: val}
		hk, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		pairs[hk.HashKey()] = pair
	}

	return &object.Hash{Pairs: pairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ArrayObj && index.Type() == object.NumberObj:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HashObj:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arr := array.(*object.Array)
	i := int(index.(*object.Number).Value)

	max := len(arr.Elements) - 1
	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arr.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObj := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParams {
		return fmt.Errorf("wrong number of arguments: expected=%d, got=%d", cl.Fn.NumParams, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)

	vm.sp = frame.bp + cl.Fn.NumLocals

	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	return nil
}

func (vm *VM) callBuiltin(fn *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := fn.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	var err error
	if result != nil {
		err = vm.push(result)
	} else {
		err = vm.push(Null)
	}

	return err
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

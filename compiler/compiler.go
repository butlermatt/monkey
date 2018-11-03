package compiler

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/object"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.OpCode
	Position int
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type CompilationScope struct {
	instructions code.Instructions
	last         EmittedInstruction
	prev         EmittedInstruction
}

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable

	scopes   []CompilationScope
	scopeInd int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions: code.Instructions{},
		last:         EmittedInstruction{},
		prev:         EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeInd:    0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{Instructions: c.instructions(), Constants: c.constants}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.InfixExpression:
		if node.Operator == "<" || node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			if node.Operator == "<=" {
				c.emit(code.OpGreaterEqual)
			} else {
				c.emit(code.OpGreater)
			}
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreater)
		case ">=":
			c.emit(code.OpGreaterEqual)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Bogus value
		jumpNtPos := c.emit(code.OpJumpNotTrue, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit bogus jump location
		jumpPos := c.emit(code.OpJump, 9999)
		afterPos := len(c.instructions())
		c.changeOperand(jumpNtPos, afterPos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstIs(code.OpPop) {
				c.removeLastPop()
			}

		}
		afterAltPos := len(c.instructions())
		c.changeOperand(jumpPos, afterAltPos)
	case *ast.NumberLiteral:
		num := &object.Number{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(num))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		var keys []ast.Expression
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)
	case *ast.FunctionLiteral:
		c.enterScope()

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstIs(code.OpPop) {
			c.replacePopWithReturn()
		}

		if !c.lastInstIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symbolTable.numDef

		inst := c.leaveScope()
		fn := &object.CompiledFunction{Instructions: inst, NumLocals: numLocals}
		c.emit(code.OpConstant, c.addConstant(fn))
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		c.emit(code.OpCall)
	}

	return nil
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},
		last:         EmittedInstruction{},
		prev:         EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeInd++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	ins := c.instructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeInd--

	c.symbolTable = c.symbolTable.Outer

	return ins
}

func (c *Compiler) instructions() code.Instructions {
	return c.scopes[c.scopeInd].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	pos := len(c.instructions())
	inst := append(c.instructions(), ins...)

	c.scopes[c.scopeInd].instructions = inst

	return pos
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) setLastInstruction(op code.OpCode, pos int) {
	prev := c.scopes[c.scopeInd].last
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeInd].prev = prev
	c.scopes[c.scopeInd].last = last
}

func (c *Compiler) lastInstIs(op code.OpCode) bool {
	if len(c.instructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeInd].last.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeInd].last
	prev := c.scopes[c.scopeInd].prev

	old := c.instructions()
	n := old[:last.Position]

	c.scopes[c.scopeInd].instructions = n
	c.scopes[c.scopeInd].last = prev
}

func (c *Compiler) changeOperand(opPos int, oper int) {
	op := code.OpCode(c.instructions()[opPos])
	newInst := code.Make(op, oper)

	c.replaceInst(opPos, newInst)
}

func (c *Compiler) replaceInst(pos int, newInst []byte) {
	ins := c.instructions()

	for i := 0; i < len(newInst); i++ {
		ins[pos+i] = newInst[i]
	}
}

func (c *Compiler) replacePopWithReturn() {
	lastPos := c.scopes[c.scopeInd].last.Position
	c.replaceInst(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeInd].last.Opcode = code.OpReturnValue
}

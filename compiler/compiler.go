package compiler

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/object"
)

type EmittedInstruction struct {
	Opcode   code.OpCode
	Position int
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{Instructions: c.instructions, Constants: c.constants}
}

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	lastInst EmittedInstruction
	prevInst EmittedInstruction

	symbolTable *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		lastInst:     EmittedInstruction{},
		prevInst:     EmittedInstruction{},
		symbolTable:  NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
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
		if c.lastInstIsPop() {
			c.removeLastPop()
		}

		// Emit bogus jump location
		jumpPos := c.emit(code.OpJump, 9999)
		afterPos := len(c.instructions)
		c.changeOperand(jumpNtPos, afterPos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstIsPop() {
				c.removeLastPop()
			}

		}
		afterAltPos := len(c.instructions)
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
		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)

	}

	return nil
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return pos
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) setLastInstruction(op code.OpCode, pos int) {
	prev := c.lastInst
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.prevInst = prev
	c.lastInst = last
}

func (c *Compiler) lastInstIsPop() bool {
	return c.lastInst.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInst.Position]
	c.lastInst = c.prevInst
}

func (c *Compiler) changeOperand(opPos int, oper int) {
	op := code.OpCode(c.instructions[opPos])
	newInst := code.Make(op, oper)

	c.replaceInst(opPos, newInst)
}

func (c *Compiler) replaceInst(pos int, newInst []byte) {
	for i := 0; i < len(newInst); i++ {
		c.instructions[pos+i] = newInst[i]
	}
}

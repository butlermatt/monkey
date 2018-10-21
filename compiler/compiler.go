package compiler

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/object"
)

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
}

func New() *Compiler {
	return &Compiler{instructions: code.Instructions{}, constants: []object.Object{}}
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
	case *ast.InfixExpression:
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
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.NumberLiteral:
		num := &object.Number{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(num))
	}

	return nil
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
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

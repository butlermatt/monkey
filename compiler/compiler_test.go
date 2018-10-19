package compiler

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/object"
	"github.com/butlermatt/monkey/parser"
	"testing"
)

type compilerTestCase struct {
	name   string
	input  string
	consts []interface{}
	insts  []code.Instructions
}

func TestNumberArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "one plus two",
			input:  "1 + 2;",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)

			compiler := New()
			err := compiler.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			bc := compiler.ByteCode()
			err = testInstructions(tt.insts, bc.Instructions)
			if err != nil {
				t.Fatalf("testInstructions failed: %s", err)
			}

			err = testConstants(t, tt.consts, bc.Constants)
			if err != nil {
				t.Fatalf("testConstants failed: %s", err)
			}
		})
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instruction length.\nexpected=%q, got=%q", concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nexpected=%q, got=%q", i, ins, actual[i])
		}
	}

	return nil
}

func testConstants(t *testing.T, expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. expected=%d, got=%d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case float64:
			err := testNumberObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testNumberObject failed: %s", i, err)
			}
		}
	}

	return nil
}

func testNumberObject(expected float64, actual object.Object) error {
	res, ok := actual.(*object.Number)
	if !ok {
		return fmt.Errorf("object is wrong type. expected=*object.Number got=%T (%+[1]v)", actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%f, got=%f", expected, res.Value)
	}

	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	var out code.Instructions

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

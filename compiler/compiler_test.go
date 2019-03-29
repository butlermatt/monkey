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
	input         string
	expectedConst []interface{}
	expectedInst  []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:         "1 + 2",
			expectedConst: []interface{}{1.0, 2.0},
			expectedInst: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInst, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConst, bytecode.Constants)
		if err != nil {
			t.Fatalf("test constants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func concatInstructions(c []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, ins := range c {
		out = append(out, ins...)
	}

	return out
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instruction length.\nexpected=%q\ngot     =%q", concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nexpected=%q\ngot     =%q", i, concatted, actual)
		}
	}

	return nil
}

func testConstants(t *testing.T, expected []interface{}, actual []object.Object) error {
	t.Helper()

	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. expected=%d, got=%d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch c := constant.(type) {
		case int:
			err := testNumberObject(float64(c), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testNumberObject failed: %s", i, err)
			}
		case float64:
			err := testNumberObject(float64(c), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testNumberObject failed: %s", i, err)
			}
		}
	}

	return nil
}

func testNumberObject(expected float64, actual object.Object) error {
	result, ok := actual.(*object.Number)
	if !ok {
		return fmt.Errorf("object is not a number. got=%T (%+[1]v)", actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%f, got=%f", expected, result.Value)
	}

	return nil
}

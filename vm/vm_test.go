package vm

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/object"
	"github.com/butlermatt/monkey/parser"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestNumberArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", float64(1)},
		{"2", float64(2)},
		{"1 + 2", float64(3)},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parse(tt.input)

			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compile error: %s", err)
			}

			vm := New(comp.Bytecode())
			err = vm.Run()
			if err != nil {
				t.Fatalf("vm error: %s", err)
			}

			stackElem := vm.StackTop()
			testExpectedObject(t, tt.expected, stackElem)
		})
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case float64:
		err := testNumberObject(float64(expected), actual)
		if err != nil {
			t.Errorf("testNumberObject failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
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

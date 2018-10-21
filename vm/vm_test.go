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
	name     string
	input    string
	expected interface{}
}

func TestNumberArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"one", "1;", 1.0},
		{"two", "2;", 2.0},
		{"one plus two", "1 + 2;", 3.0},
		{"one minus two", "1 - 2;", -1.0},
		{"one times two", "1 * 2;", 2.0},
		{"four div two", "4 / 2;", 2.0},
		{"compound 1", "50 / 2 * 2 + 10 - 5", 55.0},
		{"compound 2", "5 * (2 + 10)", 60.0},
		{"compound 3", "5 + 5 + 5 + 5 - 10", 10.0},
		{"compound 4", "2 * 2 * 2 * 2 * 2", 32.0},
		{"compound 5", "5 * 2 + 10", 20.0},
		{"compound 6", "5 + 2 * 10", 25.0},
		{"compound 7", "5 * (2 + 10)", 60.0},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", "true;", true},
		{"false", "false;", false},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)

			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			vm := New(comp.ByteCode())
			err = vm.Run()
			if err != nil {
				t.Fatalf("vm error: %s", err)
			}

			lastStack := vm.LastPoppedStackElem()
			testExpectedObject(t, tt.expected, lastStack)
		})
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case float64:
		err := testNumberObject(expected, actual)
		if err != nil {
			t.Errorf("testNumberObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
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

func testBooleanObject(expected bool, actual object.Object) error {
	res, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object wrong type. expected=*object.Boolean, got=%T (%+[1]v)", actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%t, got=%t", expected, res.Value)
	}

	return nil
}

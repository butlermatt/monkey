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
		{"negative 5", "-5;", -5.0},
		{"negative 10.5", "-10.5;", -10.5},
		{"negative compound 1", "-50 + 100 + -50", 0.0},
		{"negative compound 2", "(5 + 10 * 2 + 15 / 3) * 2 + -10", 50.0},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", "true;", true},
		{"false", "false;", false},
		{"1 lt 2", "1 < 2", true},
		{"1 gt 2", "1 > 2", false},
		{"1 lteq 2", "1 <= 2", true},
		{"1 gteq 2", "1 >= 2", false},
		{"1 lt 1", "1 < 1", false},
		{"1 gt 1", "1 > 1", false},
		{"1 lteq 1", "1 <= 1", true},
		{"1 gteq 1", "1 >= 1", true},
		{"1 eq 1", "1 == 1", true},
		{"1 noteq 1", "1 != 1", false},
		{"true eq true", "true == true", true},
		{"false eq false", "false == false", true},
		{"true eq false", "true == false", false},
		{"true noteq false", "true != false", true},
		{"false noteq true", "false != true", true},
		{"1 lt 2 is true", "(1 < 2) == true", true},
		{"1 lt 2 is false", "(1 < 2) == false", false},
		{"1 lteq 2 is true", "(1 <= 2) == true", true},
		{"1 lteq 2 is false", "(1 <= 2) == false", false},
		{"1 gt 2 is true", "(1 > 2) == true", false},
		{"1 gt 2 is false", "(1 > 2) == false", true},
		{"1 gteq 2 is true", "(1 >= 2) == true", false},
		{"1 gteq 2 is false", "(1 >= 2) == false", true},
		{"not true", "!true", false},
		{"not false", "!false", true},
		{"not five", "!5", false},
		{"not not true", "!!true;", true},
		{"not not false", "!!false;", false},
		{"not not five", "!!5;", true},
		{"not if false", "!(if (false) { 5; })", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if true ten", "if (true) { 10 }", 10.0},
		{"if true ten else twenty", "if (true) { 10 } else { 20 }", 10.0},
		{"if false ten else twenty", "if (false) { 10 } else { 20 }", 20.0},
		{"if one ten", "if (1) { 10 }", 10.0},
		{"if 1 lt 2 ten", "if (1 < 2) { 10 }", 10.0},
		{"if 1 lteq 2 ten else twenty", "if (1 <= 2) { 10 } else { 20 }", 10.0},
		{"if 1 gteq 2 then else twenty", "if (1 >= 2) { 10 } else { 20 }", 20.0},
		{"if 1 gteq 2 ten", "if (1 >= 2) { 10; }", Null},
		{"if false ten", "if (false) { 10; }", Null},
		{"if null", "if ((if (false) { 10; })) { 10; } else { 20; }", 20.0},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one", "let one = 1; one", 1.0},
		{"let one and two", "let one = 1; let two = 2; one + two", 3.0},
		{"let one and one", "let one = 1; let two = one + one; one + two", 3.0},
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
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+[1]v)", actual)
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

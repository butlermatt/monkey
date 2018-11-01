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

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{name: "simple string", input: `"monkey"`, expected: "monkey"},
		{name: "simple concat", input: `"mon" + "key"`, expected: "monkey"},
		{name: "three concat", input: `"mon" + "key" + "banana"`, expected: "monkeybanana"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"empty array", "[]", []float64{}},
		{"simple array", "[1, 2, 3]", []float64{1, 2, 3}},
		{"simple expression array", "[1 + 2, 3 * 4, 5 + 6]", []float64{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"empty hash", "{}", map[object.HashKey]float64{}},
		{"simple hash", "{1: 2, 2: 3}",
			map[object.HashKey]float64{
				(&object.Number{Value: 1}).HashKey(): 2,
				(&object.Number{Value: 2}).HashKey(): 3,
			},
		},
		{"complex hash", "{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]float64{
				(&object.Number{Value: 2}).HashKey(): 4,
				(&object.Number{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"array index", "[1, 2, 3][1]", 2.0},
		{"array expression index", "[1, 2, 3][0 + 2]", 3.0},
		{"array of array", "[[1, 2, 3]][0][0]", 1.0},
		{"empty array first element", "[][0]", Null},
		{"array index out of bounds", "[1, 2, 3][99]", Null},
		{"array negative index", "[1][-1]", Null},
		{"hash index", "{1: 1, 2: 2}[1]", 1.0},
		{"hash index 2", "{1: 1, 2: 2}[2]", 2.0},
		{"hash absent index", "{1:1}[0]", Null},
		{"empty hash index", "{}[0]", Null},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			name: "assigned fn 5 plus 10 implicit return",
			input: `let fivePlusTen = fn() { 5 + 10; };
					fivePlusTen();`,
			expected: 15.0,
		},
		{
			name: "multiple fns in an expression",
			input: `let one = fn() { 1; };
					let two = fn() { 2; };
					one() + two();`,
			expected: 3.0,
		},
		{
			name: "nested calls",
			input: `let a = fn() { 1 };
					let b = fn() { a() + 1 };
					let c = fn() { b() + 1 };
					c();`,
			expected: 3.0,
		},
		{
			name: "explicit return early",
			input: `let exitEarly = fn() { return 99; 100; };
					exitEarly();`,
			expected: 99.0,
		},
		{
			name: "explicit return early return",
			input: `let exitEarly = fn() { return 99; return 100; };
					exitEarly();`,
			expected: 99.0,
		},
		{
			name: "no return",
			input: `let noReturn = fn() {};
					noReturn();`,
			expected: Null,
		},
		{
			name: "return a no return",
			input: `let noReturn = fn() {};
					let caller = fn() { noReturn(); };
					caller();`,
			expected: Null,
		},
		{
			name: "first class function",
			input: `let returnsOne = fn() { 1; };
					let returnsFunc = fn() { returnsOne; };
					returnsFunc()();`,
			expected: 1.0,
		},
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
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case []float64:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object incorrect type. expected=*object.Array, got=%T (%+[1]v)", actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong number of elements. expected=%d, got=%d", len(expected), len(array.Elements))
			return
		}

		for i, el := range expected {
			err := testNumberObject(el, array.Elements[i])
			if err != nil {
				t.Errorf("testNumberObject failed: %s", err)
			}
		}
	case map[object.HashKey]float64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is wrong type. expected=*object.Hash, got=%T (%+[1]v)", actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of values. expected=%d, got=%d", len(expected), len(hash.Pairs))
			return
		}
		for eKey, eVal := range expected {
			pair, ok := hash.Pairs[eKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}

			err := testNumberObject(eVal, pair.Value)
			if err != nil {
				t.Errorf("testNumberObject failed: %s", err)
			}
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

func testStringObject(expected string, actual object.Object) error {
	res, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object wrong type. expected=*object.String, got=%T (%+[1]v)", actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%q, got=%q", expected, res.Value)
	}

	return nil
}

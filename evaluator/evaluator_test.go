package evaluator

import (
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/object"
	"github.com/butlermatt/monkey/parser"
	"testing"
)

func TestEvalNumberExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5", 5.0},
		{"10.5", 10.5},
		{"-5", -5},
		{"-10.5", -10.5},
		{"5 + 5.5 + 5.0 + 5 - 10.5", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100.5 + -50.5", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"21 + 2 * -10.5", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testNumberObject(t, evaluated, tt.expected)
		})
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1.5 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 >= 1", true},
		{"1 <= 1", true},
		{"1.1 != 1.1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"false != false", false},
		{"(1 < 2) == true", true},
		{"(1 <= 2) == false", false},
		{"(1 >= 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"not true", "!true;", false},
		{"not false", "!false;", true},
		{"not five", "!5;", false},
		{"not not true", "!!true;", true},
		{"not not false", "!!false;", false},
		{"not not zero", "!!0;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval := testEval(tt.input)
			testBooleanObject(t, eval, tt.expected)
		})
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"if true", "if (true) { 10 }", 10.0},
		{"if false", "if (false) { 10 }", nil},
		{"if one", "if (1) { 10 }", 10.0},
		{"if 1 lte 2", "if (1 <= 2) { 10 }", 10.0},
		{"if 1 gt 2", "if (1 > 2) { 10 }", nil},
		{"if else 1 gt 2", "if (1 > 2) { 10 } else { 20 }", 20.0},
		{"if else 1 lte 2", "if (1 <= 2) { 10 } else { 20 }", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaled := testEval(tt.input)
			num, ok := tt.expected.(float64)
			if ok {
				testNumberObject(t, evaled, num)
			} else {
				testNullObject(t, evaled)
			}
		})
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"return ten", "return 10;", 10.0},
		{"return ten ignore", "return 10; 9;", 10.0},
		{"return expression", "return 2 * 5; 9;", 10.0},
		{"return expression ignore", "9; return 2 * 5; 9;", 10.0},
		{"nested return", "if (10 > 1) { if (10 > 1) { return 10; } return 1; }", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testNumberObject(t, evaluated, tt.expected)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"5 plus true", "5 + true;", "on line 1 - type mismatch: NUMBER + BOOLEAN"},
		{"5 plus true ignore", "5 + true; 5;", "on line 1 - type mismatch: NUMBER + BOOLEAN"},
		{"negative bool", "-true;", "on line 1 - unknown operator: -BOOLEAN"},
		{"true plus true", "true + true;", "on line 1 - unknown operator: BOOLEAN + BOOLEAN"},
		{"true plus true ignore", "5; true + false; 5;", "on line 1 - unknown operator: BOOLEAN + BOOLEAN"},
		{"if block true plus true", "if (10 > 1) { true + true; }", "on line 1 - unknown operator: BOOLEAN + BOOLEAN"},
		{
			"multi-line nested",
			`
if (10 > 1) {
	if (10 > 1) {
		return true + false;
	}
}`,
			"on line 4 - unknown operator: BOOLEAN + BOOLEAN",
		},
		{"unbound variable", "foobar;", "on line 1 - identifier not found: foobar"},
		{"minus string", `"Hello" - "World";`, "on line 1 - unknown operator: STRING - STRING"},
		{"invalid hashkey", `{"name": "Monkey"}[fn(x){x}];`, "on line 1 - unusable as hash key: FUNCTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Fatalf("unexpected return type. expected=*object.Error, got=%T (%+[1]v)", evaluated)
			}

			if errObj.Message != tt.expected {
				t.Fatalf("unexpected error message. expected=%q, got=%q", tt.expected, errObj.Message)
			}
		})
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"simple assignment", "let a = 5; a;", 5.0},
		{"expression assignment", "let a = 5 * 5; a;", 25.0},
		{"evaluated assignment", "let a = 5; let b = a; b;", 5.0},
		{"complex assignment", "let a = 5; let b = a; let c = a + b + 5; c;", 15.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testNumberObject(t, testEval(tt.input), tt.expected)
		})
	}
}

func TestFunctionObjects(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object wrong type. expected=*object.Function, got=%T (%+[1]v)", evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong number of parameters. expected=%d, got=%d", 1, len(fn.Parameters))
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("function parameter incorrect. expected=%q, got=%q", "x", fn.Parameters[0].String())
	}

	body := "(x + 2)"
	if fn.Body.String() != body {
		t.Fatalf("function body incorrect. expected=%q, got=%q", body, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"explicit return", "let ident = fn(x) { return x; }; ident(5);", 5.0},
		{"implicit return", "let ident = fn(x) { x; }; ident(5);", 5.0},
		{"double", "let double = fn(x) { x * 2; }; double(5);", 10.0},
		{"add", "let add = fn(x, y) { x + y; }; add(5, 5);", 10.0},
		{"recursive add", "let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20.0},
		{"anonymous", "fn(x) { x; }(5);", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testNumberObject(t, testEval(tt.input), tt.expected)
		})
	}
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x) {
  fn(y) { x + y };
};

let addTwo = newAdder(2);
addTwo(2);`

	testNumberObject(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello world!"`

	evaled := testEval(input)
	str, ok := evaled.(*object.String)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.String, got=%T (%+[1]v)", evaled)
	}

	if str.Value != "Hello world!" {
		t.Errorf("string value is incorrect. expected=%q, got=%q", "Hello World!", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!";`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.String, got=%T (%+[1]v", evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("string value incorrect. expected=%q, got=%q", "Hello World!", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"len-empty-string", `len("");`, 0.0},
		{"len-four", `len("four");`, 4.0},
		{"len-hello-world", `len("Hello world");`, 11.0},
		{"len-1", `len(1);`, "argument to `len` not supported. got=NUMBER"},
		{"len-one-two", `len("one", "two");`, "wrong number of arguments. expected=1, got=2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)

			switch expected := tt.expected.(type) {
			case float64:
				testNumberObject(t, evaluated, expected)
			case float32:
				testNumberObject(t, evaluated, float64(expected))
			case string:
				errObj, ok := evaluated.(*object.Error)
				if !ok {
					t.Fatalf("object incorrect type. expected=*object.Error got=%T (%+[1]v)", evaluated)
				}
				if errObj.Message != expected {
					t.Fatalf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
				}
			}
		})
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3];"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.Array, got=%T (%+[1]v)", evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong number of elements. expected=%d, got=%d", 3, len(result.Elements))
	}

	testNumberObject(t, result.Elements[0], 1)
	testNumberObject(t, result.Elements[1], 4)
	testNumberObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"zero index", "[1, 2, 3][0]", 1.0},
		{"one index", "[1, 2, 3][1]", 2.0},
		{"two index", "[1, 2, 3][2]", 3.0},
		{"identifier index", "let i = 0; [1, 2][i]", 1.0},
		{"expression index", "[1, 2, 3][1 + 1]", 3.0},
		{"identifier array", "let myArray = [1, 2, 3]; myArray[2];", 3.0},
		{"identifier expressions", "let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", 6.0},
		{"nesting identifiers", "let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i];", 2.0},
		{"out of bounds", "[1, 2, 3][3]", nil},
		{"negative", "[1, 2, 3][-1]", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			value, ok := tt.expected.(float64)
			if !ok {
				testNullObject(t, evaluated)
			} else {
				testNumberObject(t, evaluated, value)
			}
		})
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
{
	"one": 10 - 9,
	two: 1 + 1,
    "thr" + "ee": 6 / 2,
	4: 4,
	true: 5,
	false: 6
};`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)

	if !ok {
		t.Fatalf("eval return wrong object type. expected=*object.Hash, got=%T (%+[1]v)", evaluated)
	}

	expected := map[object.HashKey]float64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Number{Value: 4}).HashKey():       4,
		True.HashKey():                             5,
		False.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong number of pairs. expected=%d, got=%d", len(expected), len(result.Pairs))
	}

	for exKey, exVal := range expected {
		pair, ok := result.Pairs[exKey]
		if !ok {
			t.Errorf("hash has no pair for the key %d", exKey)
			continue
		}

		testNumberObject(t, pair.Value, exVal)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"foo-5", `{"foo": 5}["foo"]`, 5.0},
		{"foo-bar", `{"foo": 5}["bar"]`, nil},
		{"ident-foo", `let key = "foo"; {"foo": 5}[key]`, 5.0},
		{"empty-foo", `{}["foo"]`, nil},
		{"5-5", `{5: 5}[5]`, 5.0},
		{"true-5", `{true: 5}[true]`, 5.0},
		{"false-5", `{false: 5}[false]`, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			number, ok := tt.expected.(float64)
			if !ok {
				testNullObject(t, evaluated)
			} else {
				testNumberObject(t, evaluated, number)
			}
		})
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testNumberObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Number)
	if !ok {
		t.Errorf("object wrong type. expected=*object.Number, got=%T (%+[1]v)", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. expected=%f, got=%f", expected, result.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object wrong type. expected=*object.Boolean, got=%T (%+[1]v)", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. expected=%t, got=%t", expected, result.Value)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != Null {
		t.Errorf("object wrong type. expected=*Null, got=%T (%+[1]v)", obj)
		return false
	}

	return true
}

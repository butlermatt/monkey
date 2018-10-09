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

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	return Eval(program)
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

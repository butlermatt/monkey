package parser

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/lexer"
	"testing"
)

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error string
	}{
		{"let ident", "let 5 = 5;", `expected next token to be "IDENT", got "NUM" instead`},
		{"let equals", "let x 5;", `expected next token to be "=", got "NUM" instead`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			_ = p.ParseProgram()
			errs := p.Errors()
			if len(errs) != 1 {
				t.Fatalf("unexpected number of errors. expected=%d, got=%d\n", 1, len(errs))
			}

			if errs[0] != tt.error {
				t.Fatalf("unexpected error message. expected=%q, got=%q\n", tt.error, errs[0])
			}
		})
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ident string
		value interface{}
	}{
		{"five", "let x = 5;", "x", 5.0},
		{"true", "let y = true;", "y", true},
		{"ident", "let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParseErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("Program statements incorrect length. expected=%d, got=%d\n", 1, len(program.Statements))
			}

			stmt := program.Statements[0]
			testLetStatement(t, stmt, tt.ident)

			val := stmt.(*ast.LetStatement).Value
			testLiteralExpression(t, val, tt.value)
		})
	}

}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Fatalf("token literal did not match. expected=%q, got=%q\n", "let", s.TokenLiteral())
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement s is wrong type. expected=*ast.LetStatement, got=%T\n", s)
	}

	if letStmt.Name.Value != name {
		t.Fatalf("let Name.Value did not match. expected=%q, got=%q\n", name, letStmt.Name.Value)
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Fatalf("name.TokenLiteral did not match. expected=%q, got=%q\n", name, letStmt.Name.TokenLiteral())
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"five", "return 5;", 5.0},
		{"true", "return true;", true},
		{"ident", "return foobar;", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			program := p.ParseProgram()
			checkParseErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements wrong length. expected=%d, got=%d", 1, len(program.Statements))
			}

			returnStmt, ok := program.Statements[0].(*ast.ReturnStatement)
			if !ok {
				t.Fatalf("stmt wrong type. expected=*ast.ReturnStatement, got=%T", program.Statements[0])
			}

			if returnStmt.TokenLiteral() != "return" {
				t.Errorf("return statement TokenLiteral is wrong. expected=%q, got=%q", "return", returnStmt.TokenLiteral())
			}

			testLiteralExpression(t, returnStmt.ReturnValue, tt.expected)
		})
	}
}

func TestBoolaenExpressions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"true", "true;", true},
		{"false", "false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParseErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has incorrect number of statements. expected=%d, got=%d", 1, len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program.Statement[0] wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
			}

			testBoolean(t, stmt.Expression, tt.expect)
		})
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has incorrect number of statements. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statement[0] wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression wrong type. expected=*ast.Identifier, got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident value incorrect. expected=%q, got=%q", "foobar", ident.Value)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident token literal incorrect. expected=%q, got=%q", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has incorrect number of statements. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statement[0] wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expression wrong type. expected=*ast.NumberLiteral, got=%T", stmt.Expression)
	}

	if literal.Value != 5.0 {
		t.Errorf("literal value is incorrect. expected=%f, got=%f", 5.0, literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal tokenLiteral is incorrect. expected=%q, got=%q", "5", literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
		value    interface{}
	}{
		{"bang", "!5;", "!", 5.0},
		{"minus", "-15.3;", "-", 15.3},
		{"not true", "!true", "!", true},
		{"not false", "!false", "!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParseErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has incorrect number of statements. expected=%d, got=%d", 1, len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program.Statement[0] wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
			}

			exp, ok := stmt.Expression.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("expression wrong type. expected=*ast.PrefixExpression, got=%T", stmt.Expression)
			}

			if exp.Operator != tt.operator {
				t.Fatalf("operator is wrong value. expected=%q, got=%q", tt.operator, exp.Operator)
			}

			testLiteralExpression(t, exp.Right, tt.value)
		})
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"plus", "5 + 10;", 5.0, "+", 10.0},
		{"minus", "10.5 - 5.5;", 10.5, "-", 5.5},
		{"times", "5.0 * 10;", 5.0, "*", 10.0},
		{"divide", "10 / 5.0;", 10.0, "/", 5.0},
		{"greater than", "10 > 5;", 10.0, ">", 5.0},
		{"less than", "5 < 10;", 5.0, "<", 10.0},
		{"greater equal", "10 >= 5;", 10.0, ">=", 5.0},
		{"less equal", "5 <= 10;", 5.0, "<=", 10.0},
		{"equality", "5 == 5;", 5.0, "==", 5.0},
		{"not equal", "10 != 5;", 10.0, "!=", 5.0},
		{"true", "true == true;", true, "==", true},
		{"not true", "true != false", true, "!=", false},
		{"false", "false == false;", false, "==", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParseErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program statements incorrect length. expected=%d, got=%d", 1, len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
			}

			testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue)
		})
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"4 <= 5 != 4 >= 5", "((4 <= 5) != (4 >= 5))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 <= 5 == true", "((3 <= 5) == true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
		{"!!true", "(!(!true))"},
		{"a * [1, 2, 3, 4][b * c] * d", "((a * ([1, 2, 3, 4][(b * c)])) * d)"},
		{"add(a * b[2], b[1], 2 * [1, 2][1])", "add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("test %d: expected=%q, got=%q", i, tt.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements incorrect length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression wrong type. expected=*ast.IfExpression, got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("consequence is incorrect legth. expected=%d, got=%d", 1, len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence statement wrong type. expected=*ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("expression alternative wrong value. expected= <nil>, got=%+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements incorrect length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression wrong type. expected=*ast.IfExpression, got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("consequence is incorrect legth. expected=%d, got=%d", 1, len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence statement wrong type. expected=*ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Fatalf("alternative is incorrect legth. expected=%d, got=%d", 1, len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence statement wrong type. expected=*ast.ExpressionStatement. got=%T", exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements wrong length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("statement expression wrong type. expected=*ast.FunctionLiteral, got=%T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function parameters wrong length. expected=%d, got=%d", 2, len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function body statements wrong length. expected=%d, got=%d", 1, len(function.Body.Statements))
	}

	bodyStatement, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body statement wrong type. expected=*ast.ExpressionStatement, got=%T", function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStatement.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "empty", input: "fn() {};", expected: []string{}},
		{name: "single", input: "fn(x) {};", expected: []string{"x"}},
		{name: "multiple", input: "fn(x, y, z) {};", expected: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParseErrors(t, p)

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
			}

			function, ok := stmt.Expression.(*ast.FunctionLiteral)
			if !ok {
				t.Fatalf("statement expression wrong type. expected=*ast.FunctionLiteral, got=%T", stmt.Expression)
			}

			if len(function.Parameters) != len(tt.expected) {
				t.Fatalf("function parameters wrong length. expected=%d, got=%d", len(tt.expected), len(function.Parameters))
			}

			for i, ident := range tt.expected {
				testLiteralExpression(t, function.Parameters[i], ident)
			}
		})
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := `add(1, 2 * 3, 4 + 5);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements wrong length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("statement expression wrong type. expected=*ast.CallExpression, got=%T", stmt.Expression)
	}

	testIdentifier(t, function.Function, "add")

	if len(function.Arguments) != 3 {
		t.Fatalf("argument list wrong length. expected=%d, got=%d", 3, len(function.Arguments))
	}

	testLiteralExpression(t, function.Arguments[0], 1.0)
	testInfixExpression(t, function.Arguments[1], 2.0, "*", 3.0)
	testInfixExpression(t, function.Arguments[2], 4.0, "+", 5.0)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements unexpected length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("first statement incorrect type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression wrong type. expected=*ast.StringLiteral, got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal value incorrect. expected=%q, got=%q", "hello world", literal.Value)
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := `[1, 2 * 2, 3 + 3];`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("statement expression wrong type. expected=*ast.ArrayLiteral, got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("array contains wrong number of elements. expected=%d, got=%d", 3, len(array.Elements))
	}

	testNumberLiteral(t, array.Elements[0], 1.0)
	testInfixExpression(t, array.Elements[1], 2.0, "*", 2.0)
	testInfixExpression(t, array.Elements[2], 3.0, "+", 3.0)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 2];"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements wrong length. expected=%d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("first program statement wrong type. expected=*ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("statement expression wrong type. expected=*ast.IndexExpression, got=%T", stmt.Expression)
	}

	testIdentifier(t, indexExp.Left, "myArray")
	testInfixExpression(t, indexExp.Index, 1.0, "+", 2.0)
}

func checkParseErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors\n", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}

	t.FailNow()
}

func testNumberLiteral(t *testing.T, nl ast.Expression, value float64) bool {
	num, ok := nl.(*ast.NumberLiteral)
	if !ok {
		t.Errorf("expression wrong type. expected=*ast.NumberLiteral, got=%T", nl)
		return false
	}

	if num.Value != value {
		t.Errorf("number value is incorrect. expected=%f, got=%f", value, num.Value)
		return false
	}

	// TODO: test string literal?
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("expression wrong type. expected=*ast.Identifier, got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident value incorrect. expected=%q, got=%q", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident TokenLiteral incorrect. expected=%q, got=%q", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testBoolean(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("expression wrong type. expected=*ast.Boolean, got=%T", exp)
		return false
	}

	if b.Value != value {
		t.Errorf("boolean value incorrect. expected=%t, got=%t", value, b.Value)
		return false
	}

	if b.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("boolean tokenLiteral incorrect. expected=%t, got=%q", value, b.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case float64:
		return testNumberLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBoolean(t, exp, v)
	}

	t.Errorf("type of exp not handled by test. %T", expected)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, oper string, right interface{}) bool {
	ieExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expression incorrect type. expected=*ast.InfixExpression, got=%T", exp)
		return false
	}

	if !testLiteralExpression(t, ieExp.Left, left) {
		return false
	}

	if ieExp.Operator != oper {
		t.Errorf("expression operator incorrect. expected=%q, got=%q", oper, ieExp.Operator)
		return false
	}

	return testLiteralExpression(t, ieExp.Right, right)
}

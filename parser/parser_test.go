package parser

import (
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
	input := `
let x = 5;
let y = 10.0;
let foobar = 238232;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParseErrors(t, p)

	if program == nil {
		t.Fatal("ParseProgram returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("Program statements incorrect length. expected=%d, got=%d\n", 3, len(program.Statements))
	}

	tests := []struct {
		expectedIdent string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		t.Run(tt.expectedIdent, func(t *testing.T) {
			testLetStatement(t, stmt, tt.expectedIdent)
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
	input := `return 5;
return 10.0;
return 993322;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements wrong length. expected=3, got=%d", len(program.Statements))
	}

	// TODO: Make this better when I start evaluating expressions.
	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt wrong type. expected=*ast.ReturnStatement, got=%T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("return statement TokenLiteral is wrong. expected=%q, got=%q", "return", returnStmt.TokenLiteral())
		}
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
		value    float64
	}{
		{"bang", "!5;", "!", 5.0},
		{"minus", "-15.3;", "-", 15.3},
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
				t.Fatalf("expression wrong type. expected=*ast.NumberLiteral, got=%T", stmt.Expression)
			}

			if exp.Operator != tt.operator {
				t.Fatalf("operator is wrong value. expected=%q, got=%q", tt.operator, exp.Operator)
			}

			testNumberLiteral(t, exp.Right, tt.value)
		})
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		leftValue  float64
		operator   string
		rightValue float64
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

			exp, ok := stmt.Expression.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("statement expression wrong type. expected=*ast.InfixExpression, got=%T", stmt.Expression)
			}

			testNumberLiteral(t, exp.Left, tt.leftValue)

			if exp.Operator != tt.operator {
				t.Errorf("expression operator incorrect. expected=%q, got=%q", tt.operator, exp.Operator)
			}

			testNumberLiteral(t, exp.Right, tt.rightValue)
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

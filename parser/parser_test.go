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

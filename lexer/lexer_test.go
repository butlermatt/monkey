package lexer

import (
	"github.com/butlermatt/monkey/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `=+(){},;
let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
	}{
		{token.Assign, "=", 1},
		{token.Plus, "+", 1},
		{token.LParen, "(", 1},
		{token.RParen, ")", 1},
		{token.LBrace, "{", 1},
		{token.RBrace, "}", 1},
		{token.Comma, ",", 1},
		{token.Semicolon, ";", 1},

		{token.Let, "let", 2},
		{token.Ident, "five", 2},
		{token.Assign, "=", 2},
		{token.Int, "5", 2},
		{token.Semicolon, ";", 2},

		{token.Let, "let", 3},
		{token.Ident, "ten", 3},
		{token.Assign, "=", 3},
		{token.Int, "10", 3},
		{token.Semicolon, ";", 3},

		{token.Let, "let", 5},
		{token.Ident, "add", 5},
		{token.Assign, "=", 5},
		{token.Function, "fn", 5},
		{token.LParen, "(", 5},
		{token.Ident, "x", 5},
		{token.Comma, ",", 5},
		{token.Ident, "y", 5},
		{token.RParen, ")", 5},
		{token.LBrace, "{", 5},

		{token.Ident, "x", 6},
		{token.Plus, "+", 6},
		{token.Ident, "y", 6},
		{token.Semicolon, ";", 6},

		{token.RBrace, "}", 7},
		{token.Semicolon, ";", 7},

		{token.Let, "let", 9},
		{token.Ident, "result", 9},
		{token.Assign, "=", 9},
		{token.Ident, "add", 9},
		{token.LParen, "(", 9},
		{token.Ident, "five", 9},
		{token.Comma, ",", 9},
		{token.Ident, "ten", 9},
		{token.RParen, ")", 9},
		{token.Semicolon, ";", 9},

		{token.EOF, "", 10},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokenType wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

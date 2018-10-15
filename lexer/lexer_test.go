package lexer

import (
	"github.com/butlermatt/monkey/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `=+(){},;~
let five = 5;
let ten = 10.0;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10.0 > 5;

if (5 < 10.0) {
    return true;
} else {
    return false;
}

10.0 == 10.0;
10.0 != 9;
9 <= 10.0 >= 5;
"foobar";
"foo bar";
[1, 2];
{"foo": "bar"};
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
		{token.Illegal, "~", 1},

		{token.Let, "let", 2},
		{token.Ident, "five", 2},
		{token.Assign, "=", 2},
		{token.Num, "5", 2},
		{token.Semicolon, ";", 2},

		{token.Let, "let", 3},
		{token.Ident, "ten", 3},
		{token.Assign, "=", 3},
		{token.Num, "10.0", 3},
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

		{token.Bang, "!", 10},
		{token.Minus, "-", 10},
		{token.Slash, "/", 10},
		{token.Star, "*", 10},
		{token.Num, "5", 10},
		{token.Semicolon, ";", 10},

		{token.Num, "5", 11},
		{token.Lt, "<", 11},
		{token.Num, "10.0", 11},
		{token.Gt, ">", 11},
		{token.Num, "5", 11},
		{token.Semicolon, ";", 11},

		{token.If, "if", 13},
		{token.LParen, "(", 13},
		{token.Num, "5", 13},
		{token.Lt, "<", 13},
		{token.Num, "10.0", 13},
		{token.RParen, ")", 13},
		{token.LBrace, "{", 13},

		{token.Return, "return", 14},
		{token.True, "true", 14},
		{token.Semicolon, ";", 14},

		{token.RBrace, "}", 15},
		{token.Else, "else", 15},
		{token.LBrace, "{", 15},

		{token.Return, "return", 16},
		{token.False, "false", 16},
		{token.Semicolon, ";", 16},

		{token.RBrace, "}", 17},

		{token.Num, "10.0", 19},
		{token.Eq, "==", 19},
		{token.Num, "10.0", 19},
		{token.Semicolon, ";", 19},

		{token.Num, "10.0", 20},
		{token.NotEq, "!=", 20},
		{token.Num, "9", 20},
		{token.Semicolon, ";", 20},

		{token.Num, "9", 21},
		{token.LtEq, "<=", 21},
		{token.Num, "10.0", 21},
		{token.GtEq, ">=", 21},
		{token.Num, "5", 21},
		{token.Semicolon, ";", 21},

		{token.String, "foobar", 22},
		{token.Semicolon, ";", 22},

		{token.String, "foo bar", 23},
		{token.Semicolon, ";", 23},

		{token.LBracket, "[", 24},
		{token.Num, "1", 24},
		{token.Comma, ",", 24},
		{token.Num, "2", 24},
		{token.RBracket, "]", 24},
		{token.Semicolon, ";", 24},

		{token.LBrace, "{", 25},
		{token.String, "foo", 25},
		{token.Colon, ":", 25},
		{token.String, "bar", 25},
		{token.RBrace, "}", 25},
		{token.Semicolon, ";", 25},

		{token.EOF, "", 26},
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

package lexer

import "github.com/butlermatt/monkey/token"

type Lexer struct {
	input    string
	position int  // current position in input (points to current char)
	readPos  int  // current reading position in input (after current char)
	ch       byte // current character under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}

	l.position = l.readPos
	l.readPos += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	switch l.ch {
	case '=':
		tok = newToken(token.Assign, l.ch)
	case ';':
		tok = newToken(token.Semicolon, l.ch)
	case '(':
		tok = newToken(token.LParen, l.ch)
	case ')':
		tok = newToken(token.RParen, l.ch)
	case ',':
		tok = newToken(token.Comma, l.ch)
	case '+':
		tok = newToken(token.Plus, l.ch)
	case '{':
		tok = newToken(token.LBrace, l.ch)
	case '}':
		tok = newToken(token.RBrace, l.ch)
	case 0:
		tok = token.New(token.EOF, "")
	}

	l.readChar()
	return tok
}

func newToken(ty token.TokenType, ch byte) token.Token {
	return token.New(ty, string(ch))
}

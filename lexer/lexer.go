package lexer

import "github.com/butlermatt/monkey/token"

type Lexer struct {
	input    string
	position int  // current position in input (points to current char)
	readPos  int  // current reading position in input (after current char)
	line     int  // current line we're on in the file.
	ch       byte // current character under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
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
		tok = newToken(token.Assign, l.ch, l.line)
	case ';':
		tok = newToken(token.Semicolon, l.ch, l.line)
	case '(':
		tok = newToken(token.LParen, l.ch, l.line)
	case ')':
		tok = newToken(token.RParen, l.ch, l.line)
	case ',':
		tok = newToken(token.Comma, l.ch, l.line)
	case '+':
		tok = newToken(token.Plus, l.ch, l.line)
	case '{':
		tok = newToken(token.LBrace, l.ch, l.line)
	case '}':
		tok = newToken(token.RBrace, l.ch, l.line)
	case 0:
		tok = token.New(token.EOF, "", l.line)
	}

	l.readChar()
	return tok
}

func newToken(ty token.TokenType, ch byte, line int) token.Token {
	return token.New(ty, string(ch), line)
}

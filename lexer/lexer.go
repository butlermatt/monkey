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

func (l *Lexer) peek() byte {
	if l.readPos >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPos]
	}
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

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peek() == '=' {
			pos := l.position
			l.readChar()
			tok = token.New(token.Eq, l.input[pos:l.readPos], l.line)
		} else {
			tok = newToken(token.Assign, l.ch, l.line)
		}
	case '+':
		tok = newToken(token.Plus, l.ch, l.line)
	case '-':
		tok = newToken(token.Minus, l.ch, l.line)
	case '*':
		tok = newToken(token.Star, l.ch, l.line)
	case '/':
		tok = newToken(token.Slash, l.ch, l.line)
	case '!':
		if l.peek() == '=' {
			pos := l.position
			l.readChar()
			tok = token.New(token.NotEq, l.input[pos:l.readPos], l.line)
		} else {
			tok = newToken(token.Bang, l.ch, l.line)
		}
	case '<':
		if l.peek() == '=' {
			pos := l.position
			l.readChar()
			tok = token.New(token.LtEq, l.input[pos:l.readPos], l.line)
		} else {
			tok = newToken(token.Lt, l.ch, l.line)
		}
	case '>':
		if l.peek() == '=' {
			pos := l.position
			l.readChar()
			tok = token.New(token.GtEq, l.input[pos:l.readPos], l.line)
		} else {
			tok = newToken(token.Gt, l.ch, l.line)
		}
	case ';':
		tok = newToken(token.Semicolon, l.ch, l.line)
	case '(':
		tok = newToken(token.LParen, l.ch, l.line)
	case ')':
		tok = newToken(token.RParen, l.ch, l.line)
	case ',':
		tok = newToken(token.Comma, l.ch, l.line)
	case '{':
		tok = newToken(token.LBrace, l.ch, l.line)
	case '}':
		tok = newToken(token.RBrace, l.ch, l.line)
	case 0:
		tok = token.New(token.EOF, "", l.line)
	default:
		if isAlpha(l.ch) {
			tok.Line = l.line
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isNumber(l.ch) {
			tok.Line = l.line
			tok.Type = token.Num
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.Illegal, l.ch, l.line)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isAlphaNumeric(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line += 1
		}

		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isNumber(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isNumber(l.peek()) {
		l.readChar()
		for isNumber(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isNumber(ch)
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isNumber(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(ty token.TokenType, ch byte, line int) token.Token {
	return token.New(ty, string(ch), line)
}

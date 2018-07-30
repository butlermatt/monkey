package token

// TokenType is the constant type of a token.
type TokenType string

// Token is the individual token including type and the string literal which composes that type.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

func New(ty TokenType, lit string, line int) Token {
	return Token{Type: ty, Literal: lit, Line: line}
}

const (
	Illegal = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers & literals
	Ident = "IDENT"
	Int   = "INT"

	// Operators
	Assign = "="
	Plus   = "+"

	// Delimiters
	Comma     = ","
	Semicolon = ";"

	LParen = "("
	RParen = ")"
	LBrace = "{"
	RBrace = "}"

	// Keywords
	Function = "FUNCTION"
	Let      = "LET"
)

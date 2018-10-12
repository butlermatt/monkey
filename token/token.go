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
	Ident  = "IDENT"
	Num    = "NUM"
	String = "STRING"

	// Operators
	Assign = "="
	Plus   = "+"
	Minus  = "-"
	Bang   = "!"
	Star   = "*"
	Slash  = "/"

	// Comparison
	Eq    = "=="
	NotEq = "!="
	Lt    = "<"
	Gt    = ">"
	LtEq  = "<="
	GtEq  = ">="

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
	True     = "TRUE"
	False    = "FALSE"
	If       = "IF"
	Else     = "ELSE"
	Return   = "RETURN"
)

var keywords = map[string]TokenType{
	"fn":     Function,
	"let":    Let,
	"if":     If,
	"else":   Else,
	"true":   True,
	"false":  False,
	"return": Return,
}

// LookupIdent returns the appropriate TokenType based on the ident string provided.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return Ident
}

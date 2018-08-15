package ast

import (
	"github.com/butlermatt/monkey/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.Let, Literal: "let"},
				Name:  &Identifier{Token: token.Token{Type: token.Ident, Literal: "myVar"}, Value: "myVar"},
			},
		},
	}
}

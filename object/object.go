package object

import (
	"bytes"
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"strings"
)

type ObjectType string

const (
	NumberObj   ObjectType = "NUMBER"
	BooleanObj  ObjectType = "BOOLEAN"
	NullObj     ObjectType = "NULL"
	FunctionObj ObjectType = "FUNCTION"
	ReturnObj   ObjectType = "RETURN_VALUE"
	ErrorObj    ObjectType = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Number struct {
	Value float64
}

func (n *Number) Inspect() string  { return fmt.Sprintf("%f", n.Value) }
func (n *Number) Type() ObjectType { return NumberObj }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BooleanObj }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NullObj }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return ReturnObj }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
	Line    int
}

func (e *Error) Type() ObjectType { return ErrorObj }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR - Line %d: %s", e.Line, e.Message) }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteByte('(')
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

package object

import (
	"bytes"
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"strings"
)

type BuiltinFunction func(line int, args ...Object) Object

type ObjectType string

const (
	NumberObj   ObjectType = "NUMBER"
	BooleanObj  ObjectType = "BOOLEAN"
	NullObj     ObjectType = "NULL"
	StringObj   ObjectType = "STRING"
	ArrayObj    ObjectType = "ARRAY"
	FunctionObj ObjectType = "FUNCTION"
	ReturnObj   ObjectType = "RETURN_VALUE"
	ErrorObj    ObjectType = "ERROR"
	BuiltinObj  ObjectType = "BUILTIN"
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

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return StringObj }
func (s *String) Inspect() string  { return s.Value }

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ArrayObj }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	var els []string
	for _, e := range ao.Elements {
		els = append(els, e.Inspect())
	}

	out.WriteByte('[')
	out.WriteString(strings.Join(els, ", "))
	out.WriteByte(']')

	return out.String()
}

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

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinObj }
func (b *Builtin) Inspect() string  { return "builtin function" }

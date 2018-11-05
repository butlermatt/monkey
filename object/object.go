package object

import (
	"bytes"
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"hash/fnv"
	"strings"
)

type BuiltinFunction func(line int, args ...Object) Object

type ObjectType string

const (
	NumberObj           ObjectType = "NUMBER"
	BooleanObj          ObjectType = "BOOLEAN"
	NullObj             ObjectType = "NULL"
	StringObj           ObjectType = "STRING"
	ArrayObj            ObjectType = "ARRAY"
	HashObj             ObjectType = "HASH"
	FunctionObj         ObjectType = "FUNCTION"
	ReturnObj           ObjectType = "RETURN_VALUE"
	ErrorObj            ObjectType = "ERROR"
	BuiltinObj          ObjectType = "BUILTIN"
	CompiledFunctionObj ObjectType = "COMPILED_FUNCTION"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Number struct {
	Value float64
}

func (n *Number) Inspect() string  { return fmt.Sprintf("%f", n.Value) }
func (n *Number) Type() ObjectType { return NumberObj }
func (n *Number) HashKey() HashKey { return HashKey{Type: n.Type(), Value: uint64(n.Value)} }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BooleanObj }
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NullObj }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return StringObj }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

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

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HashObj }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteByte('{')
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteByte('}')

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

type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
}

func (cf *CompiledFunction) Type() ObjectType { return CompiledFunctionObj }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

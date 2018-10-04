package object

import "fmt"

type ObjectType string

const (
	NumberObj  ObjectType = "NUMBER"
	BooleanObj ObjectType = "BOOLEAN"
	NullObj    ObjectType = "NULL"
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

package object

import (
	"fmt"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{Fn: builtin_len}},
	{"puts", &Builtin{Fn: builtin_puts}},
	{"first", &Builtin{Fn: builtin_first}},
	{"last", &Builtin{Fn: builtin_last}},
	{"rest", &Builtin{Fn: builtin_rest}},
	{"push", &Builtin{Fn: builtin_push}},
}

func newError(line int, format string, a ...interface{}) *Error {
	msg := "on line %d - " + format
	args := []interface{}{line}
	args = append(args, a...)
	return &Error{Message: fmt.Sprintf(msg, args...)}
}

func builtin_len(line int, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=1, got=%d", len(args))
	}

	switch arg := args[0].(type) {
	case *Array:
		return &Number{Value: float64(len(arg.Elements))}
	case *String:
		return &Number{Value: float64(len(arg.Value))}
	}

	return newError(line, "argument to `len` not supported. got=%s", args[0].Type())
}

func builtin_puts(_ int, args ...Object) Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}

	return nil
}

func builtin_first(line int, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != ArrayObj {
		return newError(line, "argument to `first` must be an ARRAY. got=%s", args[0].Type())
	}

	arr := args[0].(*Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return nil
}

func builtin_last(line int, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != ArrayObj {
		return newError(line, "argument to `last` must be an ARRAY. got=%s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)
	if length > 0 {
		return arr.Elements[length-1]
	}

	return nil
}

func builtin_rest(line int, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != ArrayObj {
		return newError(line, "argument to `last` must be an ARRAY. got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)
	if length > 0 {
		newEls := make([]Object, length-1, length-1)
		copy(newEls, arr.Elements[1:length])
		return &Array{Elements: newEls}
	}

	return nil
}

func builtin_push(line int, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 2, len(args))
	}

	if args[0].Type() != ArrayObj {
		return newError(line, "argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	length := len(arr.Elements)

	newEls := make([]Object, length+1, length+1)
	copy(newEls, arr.Elements)
	newEls[length] = args[1]

	return &Array{Elements: newEls}
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}

	return nil
}

package evaluator

import "github.com/butlermatt/monkey/object"

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{Fn: builtin_len},
}

func builtin_len(line int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("on line %d - wrong number of arguments. expected=%d, got=%d", line, 1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Number{Value: float64(len(arg.Value))}
	}

	return newError("on line %d - argument to `len` not supported, got %s", line, args[0].Type())
}

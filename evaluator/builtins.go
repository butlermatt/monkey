package evaluator

import "github.com/butlermatt/monkey/object"

var builtins = map[string]*object.Builtin{
	"len":   &object.Builtin{Fn: builtin_len},
	"first": &object.Builtin{Fn: builtin_first},
	"last":  &object.Builtin{Fn: builtin_last},
	"rest":  &object.Builtin{Fn: builtin_rest},
	"push":  &object.Builtin{Fn: builtin_push},
}

func builtin_len(line int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Number{Value: float64(len(arg.Elements))}
	case *object.String:
		return &object.Number{Value: float64(len(arg.Value))}
	}

	return newError(line, "argument to `len` not supported, got %s", args[0].Type())
}

func builtin_first(line int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != object.ArrayObj {
		return newError(line, "argument to `first` must be an ARRAY. got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return Null
}

func builtin_last(line int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != object.ArrayObj {
		return newError(line, "argument to `last` must be an ARRAY. got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	if length > 0 {
		return arr.Elements[length-1]
	}

	return Null
}

func builtin_rest(line int, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 1, len(args))
	}

	if args[0].Type() != object.ArrayObj {
		return newError(line, "argument to `last` must be an ARRAY. got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	if length > 0 {
		newEls := make([]object.Object, length-1, length-1)
		copy(newEls, arr.Elements[1:length])
		return &object.Array{Elements: newEls}
	}

	return Null
}

func builtin_push(line int, args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError(line, "wrong number of arguments. expected=%d, got=%d", 2, len(args))
	}

	if args[0].Type() != object.ArrayObj {
		return newError(line, "argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)

	newEls := make([]object.Object, length+1, length+1)
	copy(newEls, arr.Elements)
	newEls[length] = args[1]

	return &object.Array{Elements: newEls}
}

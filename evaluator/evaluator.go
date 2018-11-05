package evaluator

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/object"
)

var (
	Null  = &object.Null{}
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.NumberLiteral:
		return &object.Number{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBoolean(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Token.Line, node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Token.Line, node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(node.Token.Line, function, args)
	case *ast.ArrayLiteral:
		els := evalExpressions(node.Elements, env)
		if len(els) == 1 && isError(els[0]) {
			return els[0]
		}
		return &object.Array{Elements: els}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(node.Token.Line, left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return Null
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result.Type() {
		case object.ReturnObj:
			return result.(*object.ReturnValue).Value
		case object.ErrorObj:
			return result
		}
	}

	return result
}

func evalPrefixExpression(line int, operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(line, right)
	}
	return newError(line, "unknown operator: %s%s", operator, right.Type())
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case True:
		return False
	case False:
		fallthrough
	case Null:
		return True
	}

	return False
}

func evalMinusPrefixOperatorExpression(line int, right object.Object) object.Object {
	if right.Type() != object.NumberObj {
		return newError(line, "unknown operator: -%s", right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalInfixExpression(line int, operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError(line, "type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.NumberObj:
		return evalNumberInfixExpression(line, operator, left, right)
	case left.Type() == object.StringObj:
		return evalStringInfixExpression(line, operator, left, right)
	case operator == "==":
		return nativeBoolToBoolean(left == right)
	case operator == "!=":
		return nativeBoolToBoolean(left != right)
	}

	return newError(line, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalNumberInfixExpression(line int, operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value

	switch operator {
	case "+":
		return &object.Number{Value: leftVal + rightVal}
	case "-":
		return &object.Number{Value: leftVal - rightVal}
	case "*":
		return &object.Number{Value: leftVal * rightVal}
	case "/":
		return &object.Number{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBoolean(leftVal < rightVal)
	case ">":
		return nativeBoolToBoolean(leftVal > rightVal)
	case "==":
		return nativeBoolToBoolean(leftVal == rightVal)
	case "!=":
		return nativeBoolToBoolean(leftVal != rightVal)
	case "<=":
		return nativeBoolToBoolean(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBoolean(leftVal >= rightVal)
	}

	return newError(line, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalStringInfixExpression(line int, operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError(line, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return Null
	}
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var res object.Object

	for _, statement := range block.Statements {
		res = Eval(statement, env)

		if res != nil && res.Type() == object.ReturnObj || res.Type() == object.ErrorObj {
			return res
		}
	}

	return res
}

func evalIdentifier(ident *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(ident.Value); ok {
		return val
	}

	if builtin, ok := builtins[ident.Value]; ok {
		return builtin
	}

	return newError(ident.Token.Line, "identifier not found: %s", ident.Value)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaled := Eval(e, env)
		if isError(evaled) {
			return []object.Object{evaled}
		}
		result = append(result, evaled)
	}

	return result
}

func evalIndexExpression(line int, left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ArrayObj && index.Type() == object.NumberObj:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HashObj:
		return evalHashIndexExpression(line, left, index)
	}

	return newError(line, "index operator not support: %s[%s]", left.Type(), index.Type())
}

func evalArrayIndexExpression(left, index object.Object) object.Object {
	arr := left.(*object.Array)
	ind := int(index.(*object.Number).Value)

	if ind < 0 || ind >= len(arr.Elements) {
		return Null
	}

	return arr.Elements[ind]
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range node.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}

		hk, ok := key.(object.Hashable)
		if !ok {
			return newError(node.Token.Line, "unusable as hash key: %s", key.Type())
		}

		value := Eval(v, env)
		if isError(value) {
			return value
		}

		hash := hk.HashKey()
		pairs[hash] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(line int, hash, index object.Object) object.Object {
	hashObj := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError(line, "unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return Null
	}
	return pair.Value
}

func applyFunction(line int, fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if result := fn.Fn(line, args...); result != nil {
			return result
		}

		return Null
	}

	return newError(line, "not a function: %s", fn.Type())
}

func nativeBoolToBoolean(input bool) *object.Boolean {
	if input {
		return True
	}

	return False
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case Null:
		return false
	case False:
		return false
	default:
		return true // True or any value.
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for pId, p := range fn.Parameters {
		env.Set(p.Value, args[pId])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if obj.Type() == object.ReturnObj {
		return obj.(*object.ReturnValue).Value
	}

	return obj
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ErrorObj
	}

	return false
}

func newError(line int, format string, a ...interface{}) *object.Error {
	msg := "on line %d - " + format
	args := []interface{}{line}
	args = append(args, a...)
	return &object.Error{Message: fmt.Sprintf(msg, args...)}
}

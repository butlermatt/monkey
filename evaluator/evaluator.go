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
	return newError("on line %d - unknown operator: %s%s", line, operator, right.Type())
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
		return newError("on line %d - unknown operator: -%s", line, right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalInfixExpression(line int, operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError("on line %d - type mismatch: %s %s %s", line, left.Type(), operator, right.Type())
	case left.Type() == object.NumberObj:
		return evalNumberInfixExpression(line, operator, left, right)
	case left.Type() == object.StringObj:
		return evalStringInfixExpression(line, operator, left, right)
	case operator == "==":
		return nativeBoolToBoolean(left == right)
	case operator == "!=":
		return nativeBoolToBoolean(left != right)
	}

	return newError("on line %d - unknown operator: %s %s %s", line, left.Type(), operator, right.Type())
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

	return newError("on line %d - unknown operator: %s %s %s", line, left.Type(), operator, right.Type())
}

func evalStringInfixExpression(line int, operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("on line %d - unknown operator: %s %s %s", line, left.Type(), operator, right.Type())
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
	val, ok := env.Get(ident.Value)
	if !ok {
		return newError("on line %d - identifier not found: %s", ident.Token.Line, ident.Value)
	}

	return val
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

func applyFunction(line int, fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("on line %d - not a function: %s", line, fn.Type())
	}

	extEnv := extendFunctionEnv(function, args)
	evaluated := Eval(function.Body, extEnv)
	return unwrapReturnValue(evaluated)
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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

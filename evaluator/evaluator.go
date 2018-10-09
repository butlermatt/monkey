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

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.NumberLiteral:
		return &object.Number{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBoolean(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Token.Line, node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		if isError(left) {
			return left
		}
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Token.Line, node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	}

	return nil
}

func evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement)

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
	case left.Type() == object.NumberObj && right.Type() == object.NumberObj:
		return evalNumberInfixExpression(line, operator, left, right)
	case left.Type() != right.Type():
		return newError("on line %d - type mismatch: %s %s %s", line, left.Type(), operator, right.Type())
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

func evalIfExpression(ie *ast.IfExpression) object.Object {
	condition := Eval(ie.Condition)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return Null
	}
}

func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var res object.Object

	for _, statement := range block.Statements {
		res = Eval(statement)

		if res != nil && res.Type() == object.ReturnObj || res.Type() == object.ErrorObj {
			return res
		}
	}

	return res
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

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ErrorObj
	}

	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

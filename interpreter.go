package main

import (
	"fmt"
	"reflect"
)

type Interpreter struct {
}

func (intr Interpreter) interpret(expr Expr) (result string, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			result = ""
			err = err1.(RuntimeError)
		}
	}()
	return fmt.Sprint(intr.evaluate(expr)), nil
}

func (intr Interpreter) evaluate(expr Expr) interface{} {
	return expr.accept(intr)
}

func (intr Interpreter) visitLiteralExpr(expr Literal) interface{} {
	return expr.Value
}

func (intr Interpreter) visitGroupingExpr(expr Grouping) interface{} {
	return intr.evaluate(expr.Expr)
}

func (intr Interpreter) visitUnaryExpr(expr Unary) interface{} {
	operand := intr.evaluate(expr.Right)
	switch expr.Operator.Type {
	case MINUS:
		checkNumberOperand(expr.Operator, operand)
		return -operand.(float64)
	case BANG:
		return !isTruthy(operand)
	}
	// we should never reach this as we handled all unary operators
	// TODO: isn't it safer to panic here ?
	return nil
}

func (intr Interpreter) visitBinaryExpr(expr Binary) interface{} {
	left := intr.evaluate(expr.Left)
	right := intr.evaluate(expr.Right)
	switch expr.Operator.Type {
	case MINUS:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case SLASH:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)
	case STAR:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)
	case PLUS:
		switch left := left.(type) {
		case float64:
			if right, ok := right.(float64); ok {
				return left + right
			}
		case string:
			if right, ok := right.(string); ok {
				return left + right
			}
		}
		panic(RuntimeError{message: fmt.Sprintf("Operands must be two numbers or two strings: %v", expr.Operator)})
	case GREATER:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case GREATER_EQUAL:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case LESS:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case LESS_EQUAL:
		checkNumberOperands(expr.Operator, left, right)
		return left.(float64) <= right.(float64)
	case EQUAL_EQUAL:
		return isEqual(left, right)
	case BANG_EQUAL:
		return !isEqual(left, right)
	case OR:
		return isTruthy(left) || isTruthy(right)
	case AND:
		return isTruthy(left) && isTruthy(right)
	}
	// we should never reach this as we handled all binary operators
	// TODO: isn't it safer to panic here ?
	return nil
}

func checkNumberOperands(token Token, left, right interface{}) {
	_, okLeft := left.(float64)
	_, okRight := right.(float64)
	if !(okLeft && okRight) {
		panic(RuntimeError{message: fmt.Sprintf("Operands must be numbers: %v", token)})
	}
}

func checkNumberOperand(token Token, operand interface{}) {
	if _, ok := operand.(float64); !ok {
		panic(RuntimeError{fmt.Sprintf("Operand must be number: %v", token)})
	}
}

func isTruthy(obj interface{}) bool {
	if obj == nil {
		return false
	}
	if v, ok := obj.(bool); ok {
		return v
	}
	return true
}

func isEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}

type RuntimeError struct {
	message string
}

func (re RuntimeError) Error() string {
	return re.message
}

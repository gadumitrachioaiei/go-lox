package main

import (
	"fmt"
	"strings"
)

/*
Our grammar for expressions:
expression -> literal | unary | binary | grouping
literal -> NUMBER | STRING | "true" | "false" | "nil"
grouping -> "(" expression ")"
unary -> ("-" | "!") expression
binary -> expression operator expression
operator -> "==" | "!=" | "<" | ">" | "<=" | ">=" | "+" | "-" | "*" | "/" | or | and

We need to transform it according to operator precedence and associativity, so it can be coded by a recursive descent parser
expression     → equality
equality       → comparison (("==" | "!=") comparison)*
comparison     → term (("<" | ">" | "<=" | ">=") term) *
term           → factor (("+" | "-" | "or") factor)*
factor         → unary (("*" | "/" | "and") unary)*
unary          → ("-" | "!") unary | primary
primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")"
*/

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (expr Expr, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			expr = nil
			err = err1.(ParseError)
		}
	}()
	return p.expression(), nil
}

func (p *Parser) expression() Expr {
	return p.equality()
}

// equality       → comparison (("==" | "!=") comparison)*
func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(EQUAL_EQUAL, BANG_EQUAL) {
		expr = Binary{
			Operator: p.previous(),
			Left:     expr,
			Right:    p.comparison(),
		}
	}
	return expr
}

// comparison     → term ((">" | ">=" | "<" | "<=") term) *
func (p *Parser) comparison() Expr {
	expr := p.term()
	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		expr = Binary{
			Operator: p.previous(),
			Left:     expr,
			Right:    p.term(),
		}
	}
	return expr
}

//term           → factor (("+" | "-") factor)*
func (p *Parser) term() Expr {
	expr := p.factor()
	for p.match(PLUS, MINUS, OR) {
		expr = Binary{
			Operator: p.previous(),
			Left:     expr,
			Right:    p.factor(),
		}
	}
	return expr
}

//factor         → unary (("*" | "/") unary)*
func (p *Parser) factor() Expr {
	expr := p.unary()
	for p.match(STAR, SLASH, AND) {
		expr = Binary{
			Operator: p.previous(),
			Left:     expr,
			Right:    p.unary(),
		}
	}
	return expr
}

//unary          → ("-" | "!") unary | primary
func (p *Parser) unary() Expr {
	if !p.match(MINUS, BANG) {
		return p.primary()
	}
	return Unary{
		Operator: p.previous(),
		Right:    p.unary(),
	}
}

// primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")"
func (p *Parser) primary() Expr {
	if p.match(NUMBER, STRING) {
		return Literal{Value: p.previous().Literal}
	}
	if p.match(TRUE) {
		return Literal{Value: true}
	}
	if p.match(FALSE) {
		return Literal{Value: false}
	}
	if p.match(NIL) {
		return Literal{Value: "null"}
	}
	if p.match(LEFT_PAREN) {
		expr := p.expression()
		p.consume(RIGHT_PAREN, "Expect ')' after expression.")
		return Grouping{Expr: expr}
	}
	panic(p.error(p.peek(), "Expect expression"))
}

func (p *Parser) consume(tokenType TokenType, message string) Token {
	if p.checkTokenType(tokenType) {
		return p.advance()
	}
	panic(p.error(p.peek(), message))
}

func (p *Parser) isAtEnd() bool {
	return p.tokens[p.current].Type == EOF
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) checkTokenType(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.tokens[p.current].Type == tokenType
}

func (p *Parser) match(tokenTypes ...TokenType) bool {
	for _, typ := range tokenTypes {
		if p.checkTokenType(typ) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) error(token Token, message string) ParseError {
	return ParseError{message: fmt.Sprintf("%s %s %d at '%s'", token.Lexeme, token.Type, token.Line, message)}
}

type Expr interface {
	accept(visitor Visitor) interface{}
}

type Binary struct {
	Operator    Token
	Left, Right Expr
}

func (bexpr Binary) accept(visitor Visitor) interface{} {
	return visitor.visitBinaryExpr(bexpr)
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (uexpr Unary) accept(visitor Visitor) interface{} {
	return visitor.visitUnaryExpr(uexpr)
}

type Literal struct {
	Value interface{}
}

func (lexpr Literal) accept(visitor Visitor) interface{} {
	return visitor.visitLiteralExpr(lexpr)
}

type Grouping struct {
	Expr Expr
}

func (gexpr Grouping) accept(visitor Visitor) interface{} {
	return visitor.visitGroupingExpr(gexpr)
}

type Visitor interface {
	visitBinaryExpr(expr Binary) interface{}
	visitGroupingExpr(expr Grouping) interface{}
	visitLiteralExpr(expr Literal) interface{}
	visitUnaryExpr(expr Unary) interface{}
}

type AstPrinter struct {
}

func (astp AstPrinter) Print(expr Expr) string {
	return expr.accept(astp).(string)
}

func (astp AstPrinter) visitBinaryExpr(expr Binary) interface{} {
	return astp.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (astp AstPrinter) visitGroupingExpr(expr Grouping) interface{} {
	return astp.parenthesize("group", expr.Expr)
}

func (astp AstPrinter) visitLiteralExpr(expr Literal) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", expr.Value)
}

func (astp AstPrinter) visitUnaryExpr(expr Unary) interface{} {
	return astp.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (astp AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var builder strings.Builder
	builder.WriteString("(")
	builder.WriteString(name)
	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.accept(astp).(string))
	}
	builder.WriteString(")")
	return builder.String()
}

type ParseError struct {
	message string
}

func (pe ParseError) Error() string {
	return pe.message
}

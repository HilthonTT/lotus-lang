package parser

import (
	"fmt"
	"strconv"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/token"
)

type (
	prefixParseFn    func() ast.Expression
	infixParseFn     func(ast.Expression) ast.Expression
	postfixParseFunc func() ast.Expression
)

// Operator precedences
const (
	_ int = iota
	LOWEST
	TERNARY     // ?:
	OR_PREC     // ||
	AND_PREC    // &&
	EQUALS      // == !=
	LESSGREATER // < > <= >=
	SUM         // + -
	PRODUCT     // * / %
	PREFIX      // -x !x
	CALL        // fn(x)
	INDEX       // array[i]
)

// each token precedence
var precedences = map[token.TokenType]int{
	token.OR:       OR_PREC,
	token.AND:      AND_PREC,
	token.EQ:       EQUALS,
	token.NOTEQ:    EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTEQ:     LESSGREATER,
	token.GTEQ:     LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.MODULO:   PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      INDEX, // field access binds as tightly as indexing
	token.QUESTION: TERNARY,
}

type Parser struct {
	l *lexer.Lexer

	// prevToken holds the previous token from our lexer.
	// (used for "++" + "--")
	prevToken token.Token

	// curTOken holds the current token from our lexer.
	curToken token.Token

	// peekToken holds the next token which will come from the lexer.
	peekToken token.Token

	// errors holds parsing-errors
	errors []string

	// prefixParseFns holds a map of parsing methods for
	// prefix-based syntax.
	prefixParseFns map[token.TokenType]prefixParseFn

	// infixParseFns holds a map of parsing methods for
	// infix-based syntax.
	infixParseFns map[token.TokenType]infixParseFn

	// postfixParseFns holds a map of parsing methods for
	// postfix-based syntax.
	postfixParseFns map[token.TokenType]postfixParseFunc
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENT:    p.parseIdentifier,
		token.INT:      p.parseIntegerLiteral,
		token.FLOAT:    p.parseFloatLiteral,
		token.STRING:   p.parseStringLiteral,
		token.TRUE:     p.parseBooleanLiteral,
		token.FALSE:    p.parseBooleanLiteral,
		token.NIL:      p.parseNilLiteral,
		token.BANG:     p.parsePrefixExpression,
		token.MINUS:    p.parsePrefixExpression,
		token.LPAREN:   p.parseGroupedExpression,
		token.LBRACKET: p.parseArrayLiteral,
		token.LBRACE:   p.parseMapLiteral,
		token.IF:       p.parseIfExpression,
		token.FN:       p.parseFunctionLiteral,
		token.SELF:     p.parseSelfExpression,
		token.SUPER:    p.parseSuperExpression,
	}

	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:     p.parseInfixExpression,
		token.MINUS:    p.parseInfixExpression,
		token.ASTERISK: p.parseInfixExpression,
		token.SLASH:    p.parseInfixExpression,
		token.MODULO:   p.parseInfixExpression,
		token.EQ:       p.parseInfixExpression,
		token.NOTEQ:    p.parseInfixExpression,
		token.LT:       p.parseInfixExpression,
		token.GT:       p.parseInfixExpression,
		token.LTEQ:     p.parseInfixExpression,
		token.GTEQ:     p.parseInfixExpression,
		token.AND:      p.parseInfixExpression,
		token.OR:       p.parseInfixExpression,
		token.LPAREN:   p.parseCallExpression,
		token.LBRACKET: p.parseIndexExpression,
		token.DOT:      p.parseDotExpression,
		token.QUESTION: p.parseTernaryExpression,
	}

	p.postfixParseFns = map[token.TokenType]postfixParseFunc{
		token.PLUSPLUS:   p.parsePostfixExpression,
		token.MINUSMINUS: p.parsePostfixExpression,
	}

	// Read two tokens so curToken and peekToken are set
	p.nextToken()
	p.nextToken()

	return p
}

// Errors return stored errors
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken moves to our next token from the lexer.
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekError(t token.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("line %d: expected %s, got %s",
		p.peekToken.Line, t, p.peekToken.Type))
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// ParseProgram used to parse the whole progra
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET, token.MUT:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case token.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	case token.CLASS:
		return p.parseClassStatement()
	case token.EXPORT:
		return p.parseExportStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	default:
		return p.parseExpressionOrAssignStatement()
	}
}

// parseExportStatement: export let/fn/class ...
func (p *Parser) parseExportStatement() *ast.ExportStatement {
	stmt := &ast.ExportStatement{Token: p.curToken}

	p.nextToken() // move past 'export'

	inner := p.parseStatement()
	if inner == nil {
		return nil
	}
	stmt.Statement = inner
	return stmt
}

// parseImportStatement: import { x, y } from "path"
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse comma-separated identifiers inside { }
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
	} else {
		p.nextToken()
		stmt.Names = append(stmt.Names, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			p.nextToken()
			stmt.Names = append(stmt.Names, &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})
		}
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
	}

	if !p.expectPeek(token.FROM) {
		return nil
	}
	if !p.expectPeek(token.STRING) {
		return nil
	}
	stmt.Path = p.curToken.Literal

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseClassStatement handles: class Foo [extends Bar] { fn method(self, ...) { } ... }
func (p *Parser) parseClassStatement() *ast.ClassStatement {
	stmt := &ast.ClassStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.EXTENDS) {
		p.nextToken() // consume 'extends'
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.SuperClass = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // move past {

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if !p.curTokenIs(token.FN) {
			p.errors = append(p.errors, fmt.Sprintf(
				"line %d: expected 'fn' in class body, got %s", p.curToken.Line, p.curToken.Type))
			p.nextToken()
			continue
		}
		expr := p.parseFunctionLiteral()
		if expr != nil {
			if fn, ok := expr.(*ast.FunctionLiteral); ok {
				if fn.Name == "" {
					p.errors = append(p.errors, fmt.Sprintf(
						"line %d: class methods must have a name", p.curToken.Line))
				} else {
					stmt.Methods = append(stmt.Methods, fn)
				}
			}
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	stmt.Mutable = p.curToken.Type == token.MUT

	// Check if the next token is an identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Optional type annotation: let x: int = ...
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type name
		stmt.TypeAnnot = &ast.TypeAnnotation{Name: p.curToken.Literal}
	}

	// Check if the next token is an assignement: '='
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// Optional semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	if p.curTokenIs(token.SEMICOLON) || p.curTokenIs(token.RBRACE) {
		return stmt
	}
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: no prefix parse function for %s",
			p.curToken.Line, p.curToken.Type))
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	postfix := p.postfixParseFns[p.peekToken.Type]
	if postfix != nil {
		p.nextToken()
		return postfix()
	}

	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseSelfExpression() ast.Expression {
	return &ast.SelfExpression{Token: p.curToken}
}

func (p *Parser) parseSuperExpression() ast.Expression {
	return &ast.SuperExpression{Token: p.curToken}
}

// parseDotExpression handles the '.' infix: obj.field or obj.method(args).
// The call-expression wrapping (if any) is handled by the LPAREN infix as normal.
func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	tok := p.curToken // the '.' token

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	field := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return &ast.FieldAccessExpression{
		Token: tok,
		Left:  left,
		Field: field,
	}
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	exp := &ast.TernaryExpression{
		Token:     p.curToken, // the '?' token
		Condition: condition,
	}

	p.nextToken() // move past '?'
	exp.Consequence = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken() // move past ':'
	exp.Alternative = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// parseExpressionOrAssignStatement handles plain expressions and also detects
// assignment targets: identifier, index expressions, and now field access.
func (p *Parser) parseExpressionOrAssignStatement() ast.Statement {
	expr := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.ASSIGN) {
		switch target := expr.(type) {
		case *ast.Identifier:
			p.nextToken() // consume =
			p.nextToken()
			val := p.parseExpression(LOWEST)
			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			return &ast.AssignStatement{Token: p.curToken, Name: target, Value: val}

		case *ast.IndexExpression:
			p.nextToken()
			p.nextToken()
			val := p.parseExpression(LOWEST)
			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			return &ast.IndexAssignStatement{
				Token: p.curToken,
				Left:  target.Left,
				Index: target.Index,
				Value: val,
			}

		case *ast.FieldAccessExpression:
			// self.x = value  (or any obj.field = value)
			p.nextToken() // consume =
			p.nextToken()
			val := p.parseExpression(LOWEST)
			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			return &ast.FieldAssignStatement{
				Token: p.curToken,
				Left:  target.Left,
				Field: target.Field,
				Value: val,
			}
		}
	}

	stmt := &ast.ExpressionStatement{Token: p.curToken, Expression: expr}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	return &ast.IntegerLiteral{
		Token: p.curToken,
		Value: val,
	}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	val, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("line %d: could not parse %q as float",
			p.curToken.Line, p.curToken.Literal))
		return nil
	}
	return &ast.FloatLiteral{Token: p.curToken, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parsePostfixExpression() ast.Expression {
	return &ast.PostfixExpression{
		Token:    p.prevToken,
		Operator: p.curToken.Literal,
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	prec := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(prec)
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}
	p.nextToken()
	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		exp.Alternative = p.parseBlockStatement()
	}
	return exp
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// Optional name: fn name(...) or fn(...)
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters, lit.ParamTypes = p.parseFunctionParameters()

	// Optional return type: -> int
	if p.peekTokenIs(token.ARROW) {
		p.nextToken() // consume '->'
		p.nextToken() // move to type name
		lit.ReturnType = &ast.TypeAnnotation{Name: p.curToken.Literal}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

// parseFunctionParameters allows 'self' and 'super' keyword tokens as parameter names
// so that method definitions can include them naturally.
// parseFunctionParameters — each param can have ': Type'
func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, []*ast.TypeAnnotation) {
	var params []*ast.Identifier
	var types []*ast.TypeAnnotation

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params, types
	}
	p.nextToken()

	if !p.isParamToken() {
		p.errors = append(p.errors, fmt.Sprintf(
			"line %d: expected parameter name, got %s", p.curToken.Line, p.curToken.Type))
		return nil, nil
	}
	params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	types = append(types, p.parseOptionalTypeAnnot())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		if !p.isParamToken() {
			p.errors = append(p.errors, fmt.Sprintf(
				"line %d: expected parameter name, got %s", p.curToken.Line, p.curToken.Type))
			return nil, nil
		}
		params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		types = append(types, p.parseOptionalTypeAnnot())
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, nil
	}
	return params, types
}

// parseOptionalTypeAnnot reads ': TypeName' if present, else returns nil.
func (p *Parser) parseOptionalTypeAnnot() *ast.TypeAnnotation {
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type name
		return &ast.TypeAnnotation{Name: p.curToken.Literal}
	}
	return nil
}

// isParamToken returns true if the current token is valid as a parameter name.
func (p *Parser) isParamToken() bool {
	t := p.curToken.Type
	return t == token.IDENT || t == token.SELF
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	return &ast.CallExpression{
		Token:     p.curToken,
		Function:  function,
		Arguments: p.parseExpressionList(token.RPAREN),
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	return &ast.ArrayLiteral{
		Token:    p.curToken,
		Elements: p.parseExpressionList(token.RBRACKET),
	}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	m := &ast.MapLiteral{
		Token: p.curToken,
		Pairs: make(map[ast.Expression]ast.Expression),
	}
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		val := p.parseExpression(LOWEST)
		m.Pairs[key] = val
		m.Keys = append(m.Keys, key)
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	p.expectPeek(token.RBRACE)
	return m
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var list []ast.Expression
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

package parser

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/token"
)

type (
	prefixParseFn    func() ast.Expression
	infixParseFn     func(ast.Expression) ast.Expression
	postfixParseFunc func() ast.Expression
)

// parseMatchExpression: match x { 1 -> "one"  _ -> "other" }
func (p *Parser) parseMatchExpression() ast.Expression {
	expr := &ast.MatchExpression{Token: p.curToken}
	p.nextToken()
	expr.Subject = p.parseExpression(LOWEST)
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // move past {
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		arm := &ast.MatchArm{}
		// Wildcard
		if p.curTokenIs(token.IDENT) && p.curToken.Literal == "_" {
			arm.IsWild = true
		} else {
			arm.Pattern = p.parseExpression(LOWEST)
		}
		if !p.expectPeek(token.ARROW) {
			return nil
		}
		p.nextToken()
		arm.Body = p.parseExpression(LOWEST)
		expr.Arms = append(expr.Arms, arm)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return expr
}

// parseOptionalDotExpression: obj?.field
func (p *Parser) parseOptionalDotExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	return &ast.OptionalFieldAccess{
		Token: tok,
		Left:  left,
		Field: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf(
			"line %d, col %d: unexpected token %q",
			p.curToken.Line, p.curToken.Col, p.curToken.Literal,
		))
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

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	return &ast.CallExpression{
		Token:     p.curToken,
		Function:  function,
		Arguments: p.parseExpressionList(token.RPAREN),
	}
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

func (p *Parser) parseSpreadExpression() ast.Expression {
	tok := p.curToken
	p.nextToken()
	val := p.parseExpression(LOWEST)
	return &ast.SpreadExpression{
		Token: tok,
		Value: val,
	}
}

// parsePipeExpression: left |> fn(args)
// Desugars: value |> fn(a, b) → fn(value, a, b)
func (p *Parser) parsePipeExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	right := p.parseExpression(PIPE_PREC)
	return &ast.PipeExpression{Token: tok, Left: left, Right: right}
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

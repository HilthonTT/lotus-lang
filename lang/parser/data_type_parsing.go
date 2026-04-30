package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/token"
)

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
	tok := p.curToken
	raw := tok.Literal
	if !strings.Contains(raw, "${") {
		return &ast.StringLiteral{Token: tok, Value: raw}
	}
	return p.buildInterpolation(tok, raw)
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// Optional name: fn name(...) or fn(...)
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if p.peekTokenIs(token.LT) {
		lit.TypeParams = p.parseTypeParams()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters, lit.ParamTypes = p.parseFunctionParameters()

	// Check if last param is variadic (its type annotation is "...array")
	if len(lit.ParamTypes) > 0 {
		last := lit.ParamTypes[len(lit.ParamTypes)-1]
		if last != nil && last.Name == "...array" {
			lit.IsVariadic = true
		}
	}

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

// parseInterfaceMethodSig parses one method signature inside an interface body.
func (p *Parser) parseInterfaceMethodSig() *ast.InterfaceMethodSig {
	// fn methodName(self, param: type, ...) -> returnType
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	sig := &ast.InterfaceMethodSig{Name: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Parse params (we care about count and types, not names for interfaces)
	paramCount := 0
	for !p.peekTokenIs(token.RPAREN) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.SELF) {
			// skip self
		} else if p.curTokenIs(token.IDENT) {
			paramCount++
			// optional type annotation
			if p.peekTokenIs(token.COLON) {
				p.nextToken() // consume :
				p.nextToken() // type name
				sig.ParamTypes = append(sig.ParamTypes, &ast.TypeAnnotation{Name: p.curToken.Literal})
			} else {
				sig.ParamTypes = append(sig.ParamTypes, nil)
			}
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	sig.ParamCount = paramCount
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Optional return type
	if p.peekTokenIs(token.ARROW) {
		p.nextToken()
		p.nextToken()
		sig.ReturnType = &ast.TypeAnnotation{Name: p.curToken.Literal}
	}

	return sig
}

// parseTypeParams parses <T>, <T, U>, <T: Constraint> after a fn or class name.
// The parser consumes the < ... > but the compiler ignores type params.
func (p *Parser) parseTypeParams() []ast.TypeParam {
	// curToken is '<'
	p.nextToken() // move past <
	var params []ast.TypeParam
	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.IDENT) {
			tp := ast.TypeParam{Name: p.curToken.Literal}
			// optional constraint: T: Comparable
			if p.peekTokenIs(token.COLON) {
				p.nextToken()
				p.nextToken()
				tp.Constraint = p.curToken.Literal
			}
			params = append(params, tp)
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	// curToken should now be '>'
	return params
}

package parser

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/token"
)

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

	if p.curTokenIs(token.ELLIPSIS) {
		p.nextToken() // move to param name
		restIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		params = append(params, restIdent)
		types = append(types, &ast.TypeAnnotation{Name: "...array"})
		if !p.expectPeek(token.RPAREN) {
			return nil, nil
		}
		return params, types
	}

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

		// Handle ...rest after fixed params: fn f(a, b, ...rest)
		if p.curTokenIs(token.ELLIPSIS) {
			p.nextToken()
			restIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			params = append(params, restIdent)
			types = append(types, &ast.TypeAnnotation{Name: "...array"})
			if !p.expectPeek(token.RPAREN) {
				return nil, nil
			}
			return params, types
		}

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
	if !p.peekTokenIs(token.COLON) {
		return nil
	}
	p.nextToken() // consume ':'
	p.nextToken() // move to type name

	// Handle fn(...) -> T function type annotations
	if p.curTokenIs(token.FN) {
		name := "fn"
		// Consume (param, types, ...)
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // consume (
			depth := 1
			for depth > 0 && !p.curTokenIs(token.EOF) {
				p.nextToken()
				if p.curTokenIs(token.LPAREN) {
					depth++
				}
				if p.curTokenIs(token.RPAREN) {
					depth--
				}
			}
			name = "fn(...)"
		}
		// Consume optional -> ReturnType
		if p.peekTokenIs(token.ARROW) {
			p.nextToken() // ->
			p.nextToken() // return type name
			name = "fn(...) -> " + p.curToken.Literal
		}
		return &ast.TypeAnnotation{Name: name}
	}

	return &ast.TypeAnnotation{Name: p.curToken.Literal}
}

package parser

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/token"
)

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET, token.MUT:
		return p.parseLetStatementWithDestructure()
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
	case token.ENUM:
		return p.parseEnumStatement()
	case token.TRY:
		return p.parseTryCatchStatement()
	case token.DEFER:
		return p.parseDeferStatement()
	case token.THROW:
		return p.parseThrowStatement()
	case token.INTERFACE:
		return p.parseInterfaceStatement()
	default:
		return p.parseExpressionOrAssignStatement()
	}
}

// parseEnumStatement: enum Color { Red, Green, Blue(value) }
func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if !p.curTokenIs(token.IDENT) {
			p.nextToken()
			continue
		}
		variant := &ast.EnumVariantDef{Name: p.curToken.Literal}
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // consume (
			p.nextToken() // first field
			for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
				if p.curTokenIs(token.IDENT) {
					variant.Fields = append(variant.Fields, p.curToken.Literal)
				}
				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
				}
				p.nextToken()
			}
		}
		stmt.Variants = append(stmt.Variants, variant)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return stmt
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

	if p.peekTokenIs(token.LT) {
		stmt.TypeParams = p.parseTypeParams()
	}

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

func (p *Parser) parseLetStatement() ast.Statement {
	tok := p.curToken
	mutable := tok.Type == token.MUT
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	firstName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Multi-assignment: let a, b = ...
	if p.peekTokenIs(token.COMMA) {
		stmt := &ast.MultiLetStatement{Token: tok, Mutable: mutable}
		stmt.Names = append(stmt.Names, firstName)
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			stmt.Names = append(stmt.Names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		}
		if !p.expectPeek(token.ASSIGN) {
			return nil
		}
		p.nextToken()
		for {
			stmt.Values = append(stmt.Values, p.parseExpression(LOWEST))
			if !p.peekTokenIs(token.COMMA) {
				break
			}
			p.nextToken()
			p.nextToken()
		}
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return stmt
	}

	// ... rest of existing parseLetStatement (single assignment)
	single := &ast.LetStatement{Token: tok, Mutable: mutable, Name: firstName}
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
		single.TypeAnnot = &ast.TypeAnnotation{Name: p.curToken.Literal}
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	single.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return single
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

func (p *Parser) parseForStatement() ast.Statement {
	tok := p.curToken

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	firstName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Detect "for i, item in arr"
	if p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume ','
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		secondName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(token.IN) {
			return nil
		}
		p.nextToken()
		iterable := p.parseExpression(LOWEST)
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		body := p.parseBlockStatement()
		return &ast.ForIndexStatement{
			Token:    tok,
			Index:    firstName,
			Variable: secondName,
			Iterable: iterable,
			Body:     body,
		}
	}

	// Original "for item in arr"
	stmt := &ast.ForStatement{Token: tok}
	stmt.Variable = firstName
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

	// Multi-assign: a, b = x, y  (no let/mut keyword)
	if p.peekTokenIs(token.COMMA) {
		if _, ok := expr.(*ast.Identifier); ok {
			stmt := &ast.MultiAssignStatement{Token: p.curToken}
			stmt.Names = append(stmt.Names, expr)
			for p.peekTokenIs(token.COMMA) {
				p.nextToken() // consume ','
				p.nextToken() // next name
				stmt.Names = append(stmt.Names, p.parseExpression(LOWEST))
			}
			if !p.expectPeek(token.ASSIGN) {
				return nil
			}
			p.nextToken()
			for {
				stmt.Values = append(stmt.Values, p.parseExpression(LOWEST))
				if !p.peekTokenIs(token.COMMA) {
					break
				}
				p.nextToken()
				p.nextToken()
			}
			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			return stmt
		}
	}

	if op, ok := compoundOps[p.peekToken.Type]; ok {
		p.nextToken() // consume operator
		p.nextToken() // move to RHS
		val := p.parseExpression(LOWEST)
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return &ast.CompoundAssignStatement{
			Token:    p.curToken,
			Name:     expr,
			Operator: op,
			Value:    val,
		}
	}

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

// parseDeferStatement: defer someCall(args...)
// The call is kept as-is; the compiler wraps it in a closure.
func (p *Parser) parseDeferStatement() *ast.DeferStatement {
	stmt := &ast.DeferStatement{Token: p.curToken}
	p.nextToken()
	stmt.Call = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseTryCatchStatement:
//
//	try { ... } catch err { ... }
//	try { ... } catch { ... }       <- anonymous catch (no bindin
func (p *Parser) parseTryCatchStatement() *ast.TryCatchStatement {
	stmt := &ast.TryCatchStatement{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Try = p.parseBlockStatement()

	if !p.expectPeek(token.CATCH) {
		return nil
	}

	// Optional binding: catch err { ... }
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.CatchVar = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Catch = p.parseBlockStatement()

	return stmt
}

// parseThrowStatement: throw "message"  or  throw ErrorValue
func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
	stmt := &ast.ThrowStatement{Token: p.curToken}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseInterfaceStatement:
//
//	interface Drawable {
//	    fn draw(self) -> string
//	    fn area(self) -> float
//	}
func (p *Parser) parseInterfaceStatement() *ast.InterfaceStatement {
	stmt := &ast.InterfaceStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // move past {

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.FN) {
			sig := p.parseInterfaceMethodSig()
			if sig != nil {
				stmt.Methods = append(stmt.Methods, sig)
			}
		}
		p.nextToken()
	}
	return stmt
}

// parseLetStatement — updated to detect destructuring
// Detects: let [a, b, ...rest] = ...  and  let { name, age } = ...
func (p *Parser) parseLetStatementWithDestructure() ast.Statement {
	tok := p.curToken
	mutable := tok.Type == token.MUT

	// Array destructure: let [a, b, ...rest] = expr
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // consume '['
		stmt := &ast.ArrayDestructureStatement{Token: tok, Mutable: mutable}
		p.nextToken() // first element
		for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.ELLIPSIS) {
				p.nextToken()
				rest := &ast.SpreadExpression{
					Token: p.curToken,
					Value: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
				}
				stmt.Names = append(stmt.Names, rest)
			} else if p.curTokenIs(token.IDENT) {
				stmt.Names = append(stmt.Names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
			}
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
		if !p.expectPeek(token.ASSIGN) {
			return nil
		}
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return stmt
	}

	// Map destructure: let { name, age } = expr
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken() // consume '{'
		stmt := &ast.MapDestructureStatement{Token: tok, Mutable: mutable}
		p.nextToken()
		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.IDENT) {
				stmt.Keys = append(stmt.Keys, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
			}
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
		if !p.expectPeek(token.ASSIGN) {
			return nil
		}
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return stmt
	}

	// Fall through to original parseLetStatement logic
	return p.parseLetStatement()
}

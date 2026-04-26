package parser

import (
	"fmt"
	"strconv"
	"strings"

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
	PIPE_PREC    // |>
	NULLCOALESCE // ??
	TERNARY      // ?:
	OR_PREC      // ||
	AND_PREC     // &&
	BITWISE_OR   // |
	BITWISE_XOR  // ^
	BITWISE_AND  // &
	EQUALS       // == !=
	LESSGREATER  // < > <= >=
	SHIFT        // << >>
	SUM          // + -
	PRODUCT      // * / %
	PREFIX       // -x !x ~x
	CALL         // fn(x)
	INDEX        // arr[i] obj.field
)

// each token precedence
var precedences = map[token.TokenType]int{
	token.NULLCOALESCE: NULLCOALESCE,
	token.OR:           OR_PREC,
	token.AND:          AND_PREC,
	token.BITOR:        BITWISE_OR,
	token.BITXOR:       BITWISE_XOR,
	token.BITAND:       BITWISE_AND,
	token.EQ:           EQUALS,
	token.NOTEQ:        EQUALS,
	token.LT:           LESSGREATER,
	token.GT:           LESSGREATER,
	token.LTEQ:         LESSGREATER,
	token.GTEQ:         LESSGREATER,
	token.LSHIFT:       SHIFT,
	token.RSHIFT:       SHIFT,
	token.PLUS:         SUM,
	token.MINUS:        SUM,
	token.ASTERISK:     PRODUCT,
	token.SLASH:        PRODUCT,
	token.MODULO:       PRODUCT,
	token.LPAREN:       CALL,
	token.LBRACKET:     INDEX,
	token.DOT:          INDEX, // field access binds as tightly as indexing
	token.OPTDOT:       INDEX,
	token.QUESTION:     TERNARY,
	token.PIPE:         PIPE_PREC,
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
		token.MATCH:    p.parseMatchExpression,
		token.TILDE:    p.parsePrefixExpression,
		token.ELLIPSIS: p.parseSpreadExpression,
	}

	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:         p.parseInfixExpression,
		token.MINUS:        p.parseInfixExpression,
		token.ASTERISK:     p.parseInfixExpression,
		token.SLASH:        p.parseInfixExpression,
		token.MODULO:       p.parseInfixExpression,
		token.EQ:           p.parseInfixExpression,
		token.NOTEQ:        p.parseInfixExpression,
		token.LT:           p.parseInfixExpression,
		token.GT:           p.parseInfixExpression,
		token.LTEQ:         p.parseInfixExpression,
		token.GTEQ:         p.parseInfixExpression,
		token.AND:          p.parseInfixExpression,
		token.OR:           p.parseInfixExpression,
		token.LPAREN:       p.parseCallExpression,
		token.LBRACKET:     p.parseIndexExpression,
		token.DOT:          p.parseDotExpression,
		token.QUESTION:     p.parseTernaryExpression,
		token.NULLCOALESCE: p.parseInfixExpression,
		token.BITOR:        p.parseInfixExpression,
		token.BITXOR:       p.parseInfixExpression,
		token.BITAND:       p.parseInfixExpression,
		token.LSHIFT:       p.parseInfixExpression,
		token.RSHIFT:       p.parseInfixExpression,
		token.OPTDOT:       p.parseOptionalDotExpression,
		token.PIPE:         p.parsePipeExpression,
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

func (p *Parser) buildInterpolation(tok token.Token, raw string) ast.Expression {
	type part struct {
		text   string
		isExpr bool
	}
	var parts []part
	s := raw
	for len(s) > 0 {
		idx := strings.Index(s, "${")
		if idx == -1 {
			if len(s) > 0 {
				parts = append(parts, part{s, false})
			}
			break
		}
		if idx > 0 {
			parts = append(parts, part{s[:idx], false})
		}
		s = s[idx+2:]
		depth, end := 1, 0
		for end < len(s) && depth > 0 {
			if s[end] == '{' {
				depth++
			}
			if s[end] == '}' {
				depth--
			}
			end++
		}
		parts = append(parts, part{s[:end-1], true})
		s = s[end:]
	}
	var result ast.Expression
	for _, pt := range parts {
		var expr ast.Expression
		if pt.isExpr {
			l := lexer.New(pt.text)
			sub := New(l)
			expr = sub.parseExpression(LOWEST)
		} else {
			expr = &ast.StringLiteral{Token: tok, Value: pt.text}
		}
		if result == nil {
			result = expr
		} else {
			result = &ast.InfixExpression{Token: tok, Left: result, Operator: "+", Right: expr}
		}
	}
	if result == nil {
		return &ast.StringLiteral{Token: tok, Value: ""}
	}
	return result
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

// parseExpressionOrAssignStatement — add compound assignment detection.
// Add this BEFORE the existing p.peekTokenIs(token.ASSIGN) check:
//
// Compound assignment operators:
var compoundOps = map[token.TokenType]string{
	token.PLUS_ASSIGN:   "+=",
	token.MINUS_ASSIGN:  "-=",
	token.MUL_ASSIGN:    "*=",
	token.DIV_ASSIGN:    "/=",
	token.MOD_ASSIGN:    "%=",
	token.AND_ASSIGN:    "&=",
	token.OR_ASSIGN:     "|=",
	token.XOR_ASSIGN:    "^=",
	token.LSHIFT_ASSIGN: "<<=",
	token.RSHIFT_ASSIGN: ">>=",
}

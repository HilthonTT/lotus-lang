package parser

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/token"
)

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
		token.IN:           p.parseInfixExpression,
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
	p.errors = append(p.errors, fmt.Sprintf(
		"line %d, col %d: expected %q but got %q",
		p.peekToken.Line,
		p.peekToken.Col,
		string(t),
		p.peekToken.Literal,
	))
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

// ParseProgram used to parse the whole program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		prevErrCount := len(p.errors)

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// If new errors appeared during this statement, synchronize
		// to the next safe statement boundary instead of stopping.
		if len(p.errors) > prevErrCount {
			p.synchronize()
		} else {
			p.nextToken()
		}
	}
	return program
}

// isParamToken returns true if the current token is valid as a parameter name.
func (p *Parser) isParamToken() bool {
	t := p.curToken.Type
	return t == token.IDENT || t == token.SELF
}

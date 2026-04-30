package parser

import (
	"strings"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/lexer"
	"github.com/hilthontt/lotus/token"
)

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

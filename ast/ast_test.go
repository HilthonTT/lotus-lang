package ast

import (
	"testing"

	"github.com/hilthontt/lotus/token"
)

func TestProgramNode(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	emptyProgram := &Program{}

	if program.TokenLiteral() != "let" {
		t.Errorf("program.TokenLiteral() wrong. Expected: 'let'. Got: %q", program.TokenLiteral())
	}

	if program.String() != "let myVar = anotherVar;" {
		t.Errorf("program.String() wrong. Got: %q", program.String())
	}

	if emptyProgram.TokenLiteral() != "" {
		t.Errorf("emptyProgram.String() wrong. Expected \"\" Got: %q", program.String())
	}
}

func TestArrayLiteral(t *testing.T) {
	arrLit := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "8"},
				Value: 8,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "13"},
				Value: 13,
			},
		},
	}

	if arrLit.TokenLiteral() != "[" {
		t.Errorf("Wrong TokenLiteral for ArrayLiteral. Expected: '['. Got: %s", arrLit.TokenLiteral())
	}

	if arrLit.String() != "[8, 13]" {
		t.Errorf("Wrong String representation for ArrayLiteral. Expected: [8, 13]. Got: %s", arrLit.String())
	}
}

func TestBlockStatement(t *testing.T) {
	bs := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.STRING, Literal: "number"},
					Value: "number",
				},
				Value: &IntegerLiteral{
					Token: token.Token{Type: token.INT, Literal: "66"},
					Value: 66,
				},
			},
		},
	}

	if bs.TokenLiteral() != "{" {
		t.Errorf("Wrong TokenLiteral for BlockStatement. Expected: '{'. Got: %s", bs.TokenLiteral())
	}

	if bs.String() != "let number = 66;" {
		t.Errorf("Wrong String representation for BlockStatement. Expected: 'let number = 66;'. Got: %s", bs.String())
	}
}

func TestBoolean(t *testing.T) {
	b := &BooleanLiteral{
		Token: token.Token{Type: token.TRUE, Literal: "true"},
		Value: true,
	}

	if b.TokenLiteral() != "true" {
		t.Errorf("Wrong TokenLiteral for Boolean. Expected: 'true'. Got: %s", b.TokenLiteral())
	}

	if b.String() != "true" {
		t.Errorf("Wrong String representation for Boolean. Expected: 'true'. Got: %s", b.String())
	}
}

func TestCallExpression(t *testing.T) {
	ce := &CallExpression{
		Token: token.Token{Type: token.LPAREN, Literal: "("},
		Function: &FunctionLiteral{
			Token:      token.Token{Type: token.FN, Literal: "fn"},
			Parameters: []*Identifier{},
			Body:       &BlockStatement{},
			Name:       "add",
		},
		Arguments: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "1"},
				Value: 1,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "2"},
				Value: 2,
			},
		},
	}

	if ce.TokenLiteral() != "(" {
		t.Errorf("Wrong TokenLiteral for CallExpression. Expected: '('. Got: %s", ce.TokenLiteral())
	}

	if ce.String() != "fn add() {  }(1, 2)" {
		t.Errorf("Wrong String representation for CallExpression. Expected: 'fn add() {  }(1, 2)'. Got: %s", ce.String())
	}
}

func TestLetStatement(t *testing.T) {
	cs := &LetStatement{
		Token: token.Token{Type: token.LET, Literal: "let"},
		Name: &Identifier{
			Token: token.Token{Type: token.STRING, Literal: "lotus"},
			Value: "lotus",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.STRING, Literal: "lang"},
			Value: "lang",
		},
	}

	if cs.TokenLiteral() != "let" {
		t.Errorf("Wrong TokenLiteral for LetStatement. Expected: 'let'. Got: %s", cs.TokenLiteral())
	}

	if cs.String() != "let lotus = lang;" {
		t.Errorf("Wrong String representation for LetStatement. Expected: 'let let = lang;'. Got: %s", cs.String())
	}
}

func TestExpressionStatement(t *testing.T) {
	es := &ExpressionStatement{
		Token: token.Token{Type: token.INT, Literal: "1000"},
		Expression: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "1000"},
		},
	}
	blankExpr := &ExpressionStatement{}

	if es.TokenLiteral() != "1000" {
		t.Errorf("Wrong TokenLiteral for ExpressionStatement. Expected: '1000'. Got: %s", es.TokenLiteral())
	}

	if es.String() != "1000" {
		t.Errorf("Wrong String representation for ExpressionStatement. Expected: '1000'. Got: %s", es.String())
	}

	if blankExpr.String() != "" {
		t.Errorf("Wrong String representation for empty ExpressionStatement. Expected: empty string \" \". Got: %s", es.String())
	}
}

func TestFunctionLiteral(t *testing.T) {
	fl := &FunctionLiteral{
		Token: token.Token{Type: token.FN, Literal: "fn"},
		Parameters: []*Identifier{
			{
				Token: token.Token{Type: token.STRING, Literal: "arg1"},
				Value: "arg1",
			},
			{
				Token: token.Token{Type: token.STRING, Literal: "arg2"},
				Value: "arg2",
			},
		},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&LetStatement{
					Token: token.Token{Type: token.LET, Literal: "let"},
					Name: &Identifier{
						Token: token.Token{Type: token.STRING, Literal: "number"},
						Value: "number",
					},
					Value: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "66"},
						Value: 66,
					},
				},
			},
		},
		Name: "add",
	}

	if fl.TokenLiteral() != "fn" {
		t.Errorf("Wrong TokenLiteral for FunctionLiteral. Expected: 'func'. Got: %s", fl.TokenLiteral())
	}

	if fl.String() != "fn add(arg1, arg2) { let number = 66; }" {
		t.Errorf("Wrong String representation for FunctionLiteral. Expected: 'fn add(arg1, arg2) { let number = 66; }'. Got: %s", fl.String())
	}
}

func TestIdentifier(t *testing.T) {
	ident := &Identifier{
		Token: token.Token{Type: token.STRING, Literal: "lotus"},
		Value: "lotus",
	}

	if ident.TokenLiteral() != "lotus" {
		t.Errorf("Wrong TokenLiteral for Identifier. Expected: 'lotus'. Got: %s", ident.TokenLiteral())
	}

	if ident.String() != "lotus" {
		t.Errorf("Wrong String representation for Identifier. Expected: 'lotus'. Got: %s", ident.String())
	}
}

func TestIfExpression(t *testing.T) {
	ie := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Condition: &BooleanLiteral{
			Token: token.Token{Type: token.TRUE, Literal: "true"},
			Value: true,
		},
		Consequence: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&LetStatement{
					Token: token.Token{Type: token.LET, Literal: "let"},
					Name: &Identifier{
						Token: token.Token{Type: token.STRING, Literal: "whenTrue"},
						Value: "whenTrue",
					},
					Value: &BooleanLiteral{
						Token: token.Token{Type: token.INT, Literal: "true"},
						Value: true,
					},
				},
			},
		},
		Alternative: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{
				&LetStatement{
					Token: token.Token{Type: token.LET, Literal: "let"},
					Name: &Identifier{
						Token: token.Token{Type: token.STRING, Literal: "whenFalse"},
						Value: "whenFalse",
					},
					Value: &BooleanLiteral{
						Token: token.Token{Type: token.INT, Literal: "false"},
						Value: false,
					},
				},
			},
		},
	}

	if ie.TokenLiteral() != "if" {
		t.Errorf("Wrong TokenLiteral for IfExpression. Expected: 'if'. Got: %s", ie.TokenLiteral())
	}

	if ie.String() != "if true { let whenTrue = true; } else { let whenFalse = false; }" {
		t.Errorf("Wrong String representation for IfExpression. Expected: 'if true { let whenTrue = true; } else { let whenFalse = false; }'. Got: %s", ie.String())
	}
}

func TestIndexExpression(t *testing.T) {
	arrLit := &ArrayLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "8"},
				Value: 8,
			},
			&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "13"},
				Value: 13,
			},
		},
	}

	ie := &IndexExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Left:  arrLit,
		Index: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "0"},
			Value: 0,
		},
	}

	if ie.TokenLiteral() != "[" {
		t.Errorf("Wrong TokenLiteral for IndexExpression. Expected: '['. Got: %s", ie.TokenLiteral())
	}

	if ie.String() != "([8, 13][0])" {
		t.Errorf("Wrong String representation for IndexExpression. Expected: '([8, 13][0])'. Got: %s", ie.String())
	}
}

func TestInfixExpression(t *testing.T) {
	ie := &InfixExpression{
		Token: token.Token{Type: token.PLUS, Literal: "+"},
		Left: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
		Operator: "+",
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "10"},
			Value: 10,
		},
	}

	if ie.TokenLiteral() != "+" {
		t.Errorf("Wrong TokenLiteral for InfixExpression. Expected: '+'. Got: %s", ie.TokenLiteral())
	}

	if ie.String() != "(5 + 10)" {
		t.Errorf("Wrong String representation for InfixExpression. Expected: '(5 + 10)'. Got: %s", ie.String())
	}
}

func TestIntegerLiteral(t *testing.T) {
	il := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "10"},
		Value: 10,
	}

	if il.TokenLiteral() != "10" {
		t.Errorf("Wrong TokenLiteral for IntegerLiteral. Expected: '10'. Got: %s", il.TokenLiteral())
	}

	if il.String() != "10" {
		t.Errorf("Wrong String representation for IntegerLiteral. Expected: '10'. Got: %s", il.String())
	}
}

func TestPrefixExpression(t *testing.T) {
	pe := &PrefixExpression{
		Token: token.Token{Type: token.MINUS, Literal: "-"},
		Right: &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "5"},
			Value: 5,
		},
		Operator: "-",
	}

	if pe.TokenLiteral() != "-" {
		t.Errorf("Wrong TokenLiteral for PrefixExpression. Expected: '-'. Got: %s", pe.TokenLiteral())
	}

	if pe.String() != "(-5)" {
		t.Errorf("Wrong String representation for PrefixExpression. Expected: '(-5)'. Got: %s", pe.String())
	}
}

func TestReturnStatement(t *testing.T) {
	rs := &ReturnStatement{
		Token: token.Token{Type: token.RETURN, Literal: "return"},
		ReturnValue: &StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: "items"},
			Value: "items",
		},
	}

	if rs.TokenLiteral() != "return" {
		t.Errorf("Wrong TokenLiteral for ReturnStatement. Expected: 'return'. Got: %s", rs.TokenLiteral())
	}

	if rs.String() != "return items;" {
		t.Errorf("Wrong String representation for ReturnStatement. Expected: 'return items;'. Got: %s", rs.String())
	}
}

func TestStringLiteral(t *testing.T) {
	sl := &StringLiteral{
		Token: token.Token{Type: token.STRING, Literal: "this string is so literal"},
		Value: "this string is so literal",
	}

	if sl.TokenLiteral() != "this string is so literal" {
		t.Errorf("Wrong TokenLiteral for StringLiteral. Expected: 'this string is so literal'. Got: %s", sl.TokenLiteral())
	}

	if sl.String() != "this string is so literal" {
		t.Errorf("Wrong String representation for StringLiteral. Expected: 'this string is so literal'. Got: %s", sl.String())
	}
}

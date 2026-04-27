package compiler

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/object"
)

type Bytecode struct {
	Instructions    code.Instructions
	Constants       []object.Object
	ExportedSymbols map[string]int
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable
	scopes      []CompilationScope
	scopeIndex  int

	loopStarts []int
	breakPos   [][]int

	inClass         bool
	exportedSymbols map[string]int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	symbolTable := NewSymbolTable()
	for i, v := range Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}
	for _, name := range BuiltinPackageOrder {
		symbolTable.Define(name, false)
	}
	return &Compiler{
		constants:       []object.Object{},
		symbolTable:     symbolTable,
		scopes:          []CompilationScope{mainScope},
		scopeIndex:      0,
		exportedSymbols: make(map[string]int),
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions:    c.currentInstructions(),
		Constants:       c.constants,
		ExportedSymbols: c.exportedSymbols,
	}
}

func (c *Compiler) PublicResolve(name string) (Symbol, bool) {
	return c.symbolTable.Resolve(name)
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpClosure)
	}
}

func (c *Compiler) setSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpSetLocal, s.Index)
	case FreeScope:
		c.emit(code.OpSetFree, s.Index)
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
	return instructions
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.LetStatement:
		sym := c.symbolTable.Define(node.Name.Value, node.Mutable)
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		if sym.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}

	case *ast.AssignStatement:
		sym, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Name.Value)
		}
		if !sym.Mutable {
			return fmt.Errorf("cannot assign to immutable variable: %s", node.Name.Value)
		}
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		switch sym.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobal, sym.Index)
		case LocalScope:
			c.emit(code.OpSetLocal, sym.Index)
		}

	case *ast.IndexAssignStatement:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.emit(code.OpIndexSet)

	case *ast.ClassStatement:
		if err := c.compileClass(node); err != nil {
			return err
		}

	case *ast.FieldAssignStatement:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		nameIdx := c.addConstant(&object.String{Value: node.Field.Value})
		c.emit(code.OpSetField, nameIdx)

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			if err := c.Compile(node.ReturnValue); err != nil {
				return err
			}
			c.emit(code.OpReturn)
		} else {
			c.emit(code.OpReturnNil)
		}

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *ast.WhileStatement:
		loopStart := len(c.currentInstructions())
		c.loopStarts = append(c.loopStarts, loopStart)
		c.breakPos = append(c.breakPos, []int{})
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		exitPos := c.emit(code.OpJumpFalse, 9999)
		if err := c.Compile(node.Body); err != nil {
			return err
		}
		c.emitLoop(loopStart)
		afterLoop := len(c.currentInstructions())
		c.replaceOperand(exitPos, afterLoop)
		for _, bp := range c.breakPos[len(c.breakPos)-1] {
			c.replaceOperand(bp, afterLoop)
		}
		c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
		c.breakPos = c.breakPos[:len(c.breakPos)-1]

	case *ast.ForStatement:
		iterSym := c.symbolTable.Define("__iter__", true)
		counterSym := c.symbolTable.Define("__counter__", true)
		elemSym := c.symbolTable.Define(node.Variable.Value, true)
		if err := c.Compile(node.Iterable); err != nil {
			return err
		}
		c.setSymbol(iterSym)
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 0}))
		c.setSymbol(counterSym)
		loopStart := len(c.currentInstructions())
		c.loopStarts = append(c.loopStarts, loopStart)
		c.breakPos = append(c.breakPos, []int{})
		c.emit(code.OpGetBuiltin, builtinIndex("len"))
		c.loadSymbol(iterSym)
		c.emit(code.OpCall, 1)
		c.loadSymbol(counterSym)
		c.emit(code.OpGreater)
		exitPos := c.emit(code.OpJumpFalse, 9999)
		c.loadSymbol(iterSym)
		c.loadSymbol(counterSym)
		c.emit(code.OpIndex)
		c.setSymbol(elemSym)
		if err := c.Compile(node.Body); err != nil {
			return err
		}
		c.loadSymbol(counterSym)
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 1}))
		c.emit(code.OpAdd)
		c.setSymbol(counterSym)
		c.emitLoop(loopStart)
		afterLoop := len(c.currentInstructions())
		c.replaceOperand(exitPos, afterLoop)
		for _, bp := range c.breakPos[len(c.breakPos)-1] {
			c.replaceOperand(bp, afterLoop)
		}
		c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
		c.breakPos = c.breakPos[:len(c.breakPos)-1]

	case *ast.ForIndexStatement:
		return c.compileForIndex(node)

	case *ast.BreakStatement:
		pos := c.emit(code.OpJump, 9999)
		if len(c.breakPos) > 0 {
			c.breakPos[len(c.breakPos)-1] = append(c.breakPos[len(c.breakPos)-1], pos)
		}

	case *ast.ContinueStatement:
		if len(c.loopStarts) > 0 {
			c.emitLoop(c.loopStarts[len(c.loopStarts)-1])
		}

	case *ast.IntegerLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: node.Value}))

	case *ast.FloatLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.Float{Value: node.Value}))

	case *ast.StringLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.String{Value: node.Value}))

	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.NilLiteral:
		c.emit(code.OpNil)

	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}
		c.loadSymbol(sym)

	case *ast.SelfExpression:
		sym, ok := c.symbolTable.Resolve("self")
		if !ok {
			return fmt.Errorf("'self' used outside of a method")
		}
		c.loadSymbol(sym)

	case *ast.SuperExpression:
		if !c.inClass {
			return fmt.Errorf("'super' used outside of a class method")
		}
		c.emit(code.OpGetSuper)

	case *ast.FieldAccessExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		nameIdx := c.addConstant(&object.String{Value: node.Field.Value})
		c.emit(code.OpGetField, nameIdx)

	case *ast.PrefixExpression:
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpNegate)
		case "!":
			c.emit(code.OpNot)
		case "~":
			c.emit(code.OpBitNot)
		default:
			return fmt.Errorf("unknown prefix operator: %s", node.Operator)
		}

	case *ast.PostfixExpression:
		symbol, ok := c.symbolTable.Resolve(node.TokenLiteral())
		if !ok {
			return fmt.Errorf("undefined variable %s", node.TokenLiteral())
		}
		if !symbol.Mutable {
			return fmt.Errorf("cannot assign to immutable variable: %s", node.TokenLiteral())
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
			switch node.Operator {
			case "++":
				c.emit(code.OpPlusPlus)
			case "--":
				c.emit(code.OpMinusMinus)
			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}
			c.emit(code.OpSetGlobal, symbol.Index)
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
			switch node.Operator {
			case "++":
				c.emit(code.OpPlusPlus)
			case "--":
				c.emit(code.OpMinusMinus)
			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}
			c.emit(code.OpSetLocal, symbol.Index)
			c.emit(code.OpGetLocal, symbol.Index)
		}

	case *ast.InfixExpression:
		if node.Operator == "<" {
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			if err := c.Compile(node.Left); err != nil {
				return err
			}
			c.emit(code.OpGreater)
			return nil
		}
		if node.Operator == "<=" {
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			if err := c.Compile(node.Left); err != nil {
				return err
			}
			c.emit(code.OpGreaterEq)
			return nil
		}
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if node.Operator == "&&" {
			falseJump := c.emit(code.OpJumpFalse, 9999)
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			endJump := c.emit(code.OpJump, 9999)
			c.replaceOperand(falseJump, len(c.currentInstructions()))
			c.emit(code.OpFalse)
			c.replaceOperand(endJump, len(c.currentInstructions()))
			return nil
		}
		if node.Operator == "||" {
			falseJump := c.emit(code.OpJumpFalse, 9999)
			c.emit(code.OpTrue)
			endJump := c.emit(code.OpJump, 9999)
			c.replaceOperand(falseJump, len(c.currentInstructions()))
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			c.replaceOperand(endJump, len(c.currentInstructions()))
			return nil
		}
		if node.Operator == "??" {
			c.emit(code.OpDup)
			c.emit(code.OpNil)
			c.emit(code.OpEqual)
			hasVal := c.emit(code.OpJumpFalse, 9999)
			c.emit(code.OpPop)
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			end := c.emit(code.OpJump, 9999)
			c.replaceOperand(hasVal, len(c.currentInstructions()))
			c.replaceOperand(end, len(c.currentInstructions()))
			return nil
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "%":
			c.emit(code.OpMod)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">":
			c.emit(code.OpGreater)
		case ">=":
			c.emit(code.OpGreaterEq)
		case "&":
			c.emit(code.OpBitAnd)
		case "|":
			c.emit(code.OpBitOr)
		case "^":
			c.emit(code.OpBitXor)
		case "<<":
			c.emit(code.OpLShift)
		case ">>":
			c.emit(code.OpRShift)
		case "in":
			c.emit(code.OpIn)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}

	case *ast.IfExpression:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		falseJump := c.emit(code.OpJumpFalse, 9999)
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		} else if !c.lastInstructionIs(code.OpReturn) {
			c.emit(code.OpNil)
		}
		endJump := c.emit(code.OpJump, 9999)
		c.replaceOperand(falseJump, len(c.currentInstructions()))
		if node.Alternative != nil {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			} else if !c.lastInstructionIs(code.OpReturn) {
				c.emit(code.OpNil)
			}
		} else {
			c.emit(code.OpNil)
		}
		c.replaceOperand(endJump, len(c.currentInstructions()))

	case *ast.TernaryExpression:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		falseJump := c.emit(code.OpJumpFalse, 9999)
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
		endJump := c.emit(code.OpJump, 9999)
		c.replaceOperand(falseJump, len(c.currentInstructions()))
		if err := c.Compile(node.Alternative); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
		c.replaceOperand(endJump, len(c.currentInstructions()))

	case *ast.FunctionLiteral:
		var outerSym *Symbol
		if node.Name != "" {
			s := c.symbolTable.Define(node.Name, false)
			outerSym = &s
		}
		c.enterScope()
		isVariadic := false
		for i, p := range node.Parameters {
			isLast := i == len(node.Parameters)-1
			if isLast && node.IsVariadic {
				// Mark as variadic — VM will pack extra args into array
				c.symbolTable.Define(p.Value, false)
				isVariadic = true
			} else {
				c.symbolTable.Define(p.Value, false)
			}
		}
		_ = isVariadic // VM reads NumParams and IsVariadic from CompiledFunction

		if err := c.Compile(node.Body); err != nil {
			return err
		}
		if !c.lastInstructionIs(code.OpReturn) {
			if c.lastInstructionIs(code.OpPop) {
				c.replaceLastPopWithReturn()
			} else {
				c.emit(code.OpReturnNil)
			}
		}
		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.NumDefinitions
		instructions := c.leaveScope()
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}
		fn := &object.CompiledFunction{
			Instructions: instructions,
			NumLocals:    numLocals,
			NumParams:    len(node.Parameters),
			Name:         node.Name,
			IsVariadic:   node.IsVariadic,
		}
		fnIdx := c.addConstant(fn)
		c.emit(code.OpClosure, fnIdx, len(freeSymbols))
		if outerSym != nil {
			if outerSym.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, outerSym.Index)
				c.emit(code.OpGetGlobal, outerSym.Index)
			} else {
				c.emit(code.OpSetLocal, outerSym.Index)
				c.emit(code.OpGetLocal, outerSym.Index)
			}
		}

	case *ast.SpreadExpression:
		// Standalone spread — compile the value (spread resolution is at call site)
		return c.Compile(node.Value)

	case *ast.CallExpression:
		// Check for spread args
		hasSpread := false
		for _, a := range node.Arguments {
			if _, ok := a.(*ast.SpreadExpression); ok {
				hasSpread = true
				break
			}
		}

		if hasSpread {
			if err := c.Compile(node.Function); err != nil {
				return err
			}
			for _, a := range node.Arguments {
				if spread, ok := a.(*ast.SpreadExpression); ok {
					if err := c.Compile(spread.Value); err != nil {
						return err
					}
					c.emit(code.OpSpread)
				} else {
					if err := c.Compile(a); err != nil {
						return err
					}
				}
			}
			c.emit(code.OpSpreadCall, len(node.Arguments))
			return nil
		}

		// Method call: obj.method(args) or super.method(args)
		if fieldAccess, ok := node.Function.(*ast.FieldAccessExpression); ok {
			if _, isSuper := fieldAccess.Left.(*ast.SuperExpression); isSuper {
				if !c.inClass {
					return fmt.Errorf("'super' used outside of a class method")
				}
				c.emit(code.OpGetSuper)
			} else {
				if err := c.Compile(fieldAccess.Left); err != nil {
					return err
				}
			}
			for _, a := range node.Arguments {
				if err := c.Compile(a); err != nil {
					return err
				}
			}
			nameIdx := c.addConstant(&object.String{Value: fieldAccess.Field.Value})
			c.emit(code.OpInvokeMethod, nameIdx, len(node.Arguments))
			return nil
		}

		// Regular call
		if err := c.Compile(node.Function); err != nil {
			return err
		}
		for _, a := range node.Arguments {
			if err := c.Compile(a); err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))

	case *ast.ArrayLiteral:
		// Check for spread elements in array literals: [...a, ...b]
		hasSpread := false
		for _, el := range node.Elements {
			if _, ok := el.(*ast.SpreadExpression); ok {
				hasSpread = true
				break
			}
		}
		if hasSpread {
			return c.compileSpreadArray(node.Elements)
		}
		for _, el := range node.Elements {
			if err := c.Compile(el); err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))

	case *ast.MapLiteral:
		for _, k := range node.Keys {
			if err := c.Compile(k); err != nil {
				return err
			}
			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}
		c.emit(code.OpMap, len(node.Keys)*2)

	case *ast.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(code.OpIndex)

	case *ast.ExportStatement:
		if err := c.compileExport(node); err != nil {
			return err
		}

	case *ast.ImportStatement:
		if err := c.compileImport(node); err != nil {
			return err
		}

	case *ast.OptionalFieldAccess:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		c.emit(code.OpDup)
		c.emit(code.OpNil)
		c.emit(code.OpEqual)
		isNilJump := c.emit(code.OpJumpFalse, 9999)
		c.emit(code.OpPop)
		c.emit(code.OpNil)
		endJump := c.emit(code.OpJump, 9999)
		c.replaceOperand(isNilJump, len(c.currentInstructions()))
		nameIdx := c.addConstant(&object.String{Value: node.Field.Value})
		c.emit(code.OpGetField, nameIdx)
		c.replaceOperand(endJump, len(c.currentInstructions()))

	case *ast.MatchExpression:
		if err := c.Compile(node.Subject); err != nil {
			return err
		}
		var endJumps []int
		for _, arm := range node.Arms {
			if arm.IsWild {
				c.emit(code.OpPop)
				if err := c.Compile(arm.Body); err != nil {
					return err
				}
				if c.lastInstructionIs(code.OpPop) {
					c.removeLastPop()
				}
			} else {
				c.emit(code.OpDup)
				if err := c.Compile(arm.Pattern); err != nil {
					return err
				}
				c.emit(code.OpEqual)
				nextArm := c.emit(code.OpJumpFalse, 9999)
				c.emit(code.OpPop)
				if err := c.Compile(arm.Body); err != nil {
					return err
				}
				if c.lastInstructionIs(code.OpPop) {
					c.removeLastPop()
				}
				endJumps = append(endJumps, c.emit(code.OpJump, 9999))
				c.replaceOperand(nextArm, len(c.currentInstructions()))
			}
		}
		for _, j := range endJumps {
			c.replaceOperand(j, len(c.currentInstructions()))
		}

	case *ast.EnumStatement:
		sym := c.symbolTable.Define(node.Name.Value, false)
		variants := make(map[string]*object.EnumVariantDef)
		for _, v := range node.Variants {
			variants[v.Name] = &object.EnumVariantDef{Fields: v.Fields}
		}
		c.emit(code.OpConstant, c.addConstant(&object.EnumDef{Name: node.Name.Value, Variants: variants}))
		c.setSymbol(sym)

	case *ast.MultiLetStatement:
		for _, val := range node.Values {
			if err := c.Compile(val); err != nil {
				return err
			}
		}
		for i := len(node.Names) - 1; i >= 0; i-- {
			sym := c.symbolTable.Define(node.Names[i].Value, node.Mutable)
			c.setSymbol(sym)
		}

	case *ast.MultiAssignStatement:
		for _, val := range node.Values {
			if err := c.Compile(val); err != nil {
				return err
			}
		}
		for i := len(node.Names) - 1; i >= 0; i-- {
			name, ok := node.Names[i].(*ast.Identifier)
			if !ok {
				return fmt.Errorf("invalid assignment target")
			}
			sym, ok := c.symbolTable.Resolve(name.Value)
			if !ok {
				return fmt.Errorf("undefined variable: %s", name.Value)
			}
			c.setSymbol(sym)
		}

	case *ast.DeferStatement:
		body := &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: node.Call},
			},
		}
		anonFn := &ast.FunctionLiteral{Parameters: []*ast.Identifier{}, Body: body}
		if err := c.Compile(anonFn); err != nil {
			return err
		}
		c.emit(code.OpDefer)

	case *ast.ThrowStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.emit(code.OpThrow)

	case *ast.TryCatchStatement:
		tryBeginPos := c.emit(code.OpTryBegin, 9999)
		if err := c.Compile(node.Try); err != nil {
			return err
		}
		c.emit(code.OpTryEnd)
		skipCatchJump := c.emit(code.OpJump, 9999)
		catchStart := len(c.currentInstructions())
		c.replaceOperand(tryBeginPos, catchStart)
		if node.CatchVar != nil {
			sym := c.symbolTable.Define(node.CatchVar.Value, false)
			c.setSymbol(sym)
		} else {
			c.emit(code.OpPop)
		}
		if err := c.Compile(node.Catch); err != nil {
			return err
		}
		c.replaceOperand(skipCatchJump, len(c.currentInstructions()))

	case *ast.InterfaceStatement:
		sym := c.symbolTable.Define(node.Name.Value, false)
		methods := make([]object.InterfaceMethodSpec, len(node.Methods))
		for i, m := range node.Methods {
			methods[i] = object.InterfaceMethodSpec{Name: m.Name, ParamCount: m.ParamCount}
			if m.ReturnType != nil {
				methods[i].ReturnType = m.ReturnType.Name
			}
		}
		iface := &object.Interface{Name: node.Name.Value, Methods: methods}
		c.emit(code.OpConstant, c.addConstant(iface))
		c.setSymbol(sym)

	case *ast.CompoundAssignStatement:
		switch target := node.Name.(type) {
		case *ast.Identifier:
			sym, ok := c.symbolTable.Resolve(target.Value)
			if !ok {
				return fmt.Errorf("undefined variable: %s", target.Value)
			}
			if !sym.Mutable {
				return fmt.Errorf("cannot assign to immutable: %s", target.Value)
			}
			c.loadSymbol(sym)
			if err := c.Compile(node.Value); err != nil {
				return err
			}
			c.emitCompoundOp(node.Operator)
			c.setSymbol(sym)
		case *ast.IndexExpression:
			return c.compileCompoundIndex(node, target)
		case *ast.FieldAccessExpression:
			return c.compileCompoundField(node, target)
		default:
			return fmt.Errorf("invalid compound assignment target")
		}

	case *ast.ArrayDestructureStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		for i, name := range node.Names {
			switch n := name.(type) {
			case *ast.Identifier:
				c.emit(code.OpDup)
				c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: int64(i)}))
				c.emit(code.OpIndex)
				sym := c.symbolTable.Define(n.Value, node.Mutable)
				c.setSymbol(sym)
			case *ast.SpreadExpression:
				ident, ok := n.Value.(*ast.Identifier)
				if !ok {
					return fmt.Errorf("rest element must be an identifier")
				}
				c.emit(code.OpDup)
				c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: int64(i)}))
				c.emit(code.OpArraySliceFrom)
				sym := c.symbolTable.Define(ident.Value, node.Mutable)
				c.setSymbol(sym)
			}
		}
		c.emit(code.OpPop)

	case *ast.MapDestructureStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		for i, key := range node.Keys {
			isLast := i == len(node.Keys)-1
			if !isLast {
				c.emit(code.OpDup)
			}
			c.emit(code.OpConstant, c.addConstant(&object.String{Value: key.Value}))
			c.emit(code.OpIndex)
			sym := c.symbolTable.Define(key.Value, node.Mutable)
			c.setSymbol(sym)
		}

	case *ast.PipeExpression:
		switch call := node.Right.(type) {
		case *ast.CallExpression:
			if err := c.Compile(call.Function); err != nil {
				return err
			}
			if err := c.Compile(node.Left); err != nil {
				return err
			}
			for _, arg := range call.Arguments {
				if err := c.Compile(arg); err != nil {
					return err
				}
			}
			c.emit(code.OpCall, len(call.Arguments)+1)
		default:
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			if err := c.Compile(node.Left); err != nil {
				return err
			}
			c.emit(code.OpCall, 1)
		}
	}

	return nil
}

// compileSpreadArray handles [...a, b, ...c] array literals.
// Builds the array by concatenating segments using Array.flat.
func (c *Compiler) compileSpreadArray(elements []ast.Expression) error {
	for _, el := range elements {
		if spread, ok := el.(*ast.SpreadExpression); ok {
			if err := c.Compile(spread.Value); err != nil {
				return err
			}
			c.emit(code.OpSpread) // marks as SpreadValue
		} else {
			if err := c.Compile(el); err != nil {
				return err
			}
		}
	}
	// OpArray will see SpreadValues and flatten them in buildArray
	c.emit(code.OpArray, len(elements))
	return nil
}

func (c *Compiler) compileExport(node *ast.ExportStatement) error {
	if err := c.Compile(node.Statement); err != nil {
		return err
	}
	switch inner := node.Statement.(type) {
	case *ast.LetStatement:
		sym, ok := c.symbolTable.Resolve(inner.Name.Value)
		if ok && sym.Scope == GlobalScope {
			c.exportedSymbols[inner.Name.Value] = sym.Index
		}
	case *ast.ExpressionStatement:
		if fn, ok := inner.Expression.(*ast.FunctionLiteral); ok && fn.Name != "" {
			sym, ok := c.symbolTable.Resolve(fn.Name)
			if ok && sym.Scope == GlobalScope {
				c.exportedSymbols[fn.Name] = sym.Index
			}
		}
	case *ast.ClassStatement:
		sym, ok := c.symbolTable.Resolve(inner.Name.Value)
		if ok && sym.Scope == GlobalScope {
			c.exportedSymbols[inner.Name.Value] = sym.Index
		}
	}
	return nil
}

func (c *Compiler) compileImport(node *ast.ImportStatement) error {
	if len(node.Names) == 0 {
		return fmt.Errorf("import statement has no names")
	}
	pathIdx := c.addConstant(&object.String{Value: node.Path})
	c.emit(code.OpRunModule, pathIdx)
	for i, name := range node.Names {
		isLast := i == len(node.Names)-1
		if !isLast {
			c.emit(code.OpDup)
		}
		nameIdx := c.addConstant(&object.String{Value: name.Value})
		c.emit(code.OpGetField, nameIdx)
		sym := c.symbolTable.Define(name.Value, false)
		if sym.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}
	}
	return nil
}

func (c *Compiler) compileClass(node *ast.ClassStatement) error {
	sym := c.symbolTable.Define(node.Name.Value, false)
	nameIdx := c.addConstant(&object.String{Value: node.Name.Value})
	c.emit(code.OpNewClass, nameIdx)
	if node.SuperClass != nil {
		superSym, ok := c.symbolTable.Resolve(node.SuperClass.Value)
		if !ok {
			return fmt.Errorf("undefined class '%s'", node.SuperClass.Value)
		}
		c.loadSymbol(superSym)
		c.emit(code.OpSetSuper)
	}
	prevInClass := c.inClass
	c.inClass = true
	for _, method := range node.Methods {
		if err := c.compileMethod(method); err != nil {
			return err
		}
		methodNameIdx := c.addConstant(&object.String{Value: method.Name})
		c.emit(code.OpDefineMethod, methodNameIdx)
	}
	c.inClass = prevInClass
	c.setSymbol(sym)
	return nil
}

func (c *Compiler) compileMethod(method *ast.FunctionLiteral) error {
	c.enterScope()
	for _, p := range method.Parameters {
		c.symbolTable.Define(p.Value, false)
	}
	if err := c.Compile(method.Body); err != nil {
		return err
	}
	if !c.lastInstructionIs(code.OpReturn) {
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		} else {
			c.emit(code.OpReturnNil)
		}
	}
	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.NumDefinitions
	instructions := c.leaveScope()
	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}
	fn := &object.CompiledFunction{
		Instructions: instructions,
		NumLocals:    numLocals,
		NumParams:    len(method.Parameters),
		Name:         method.Name,
	}
	fnIdx := c.addConstant(fn)
	c.emit(code.OpClosure, fnIdx, len(freeSymbols))
	return nil
}

func (c *Compiler) addInstruction(ins []byte) int {
	pos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) replaceOperand(pos int, operand int) {
	ins := c.currentInstructions()
	ins[pos+1] = byte(operand >> 8)
	ins[pos+2] = byte(operand)
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emitLoop(loopStart int) {
	pos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), code.Make(code.OpLoop, 0)...)
	offset := pos - loopStart + 1
	c.scopes[c.scopeIndex].instructions[pos+1] = byte(offset >> 8)
	c.scopes[c.scopeIndex].instructions[pos+2] = byte(offset)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	c.scopes[c.scopeIndex].instructions = c.currentInstructions()[:len(c.currentInstructions())-1]
}

func (c *Compiler) replaceLastPopWithReturn() {
	ins := c.currentInstructions()
	ins[len(ins)-1] = byte(code.OpReturn)
}

func (c *Compiler) emitCompoundOp(op string) {
	switch op {
	case "+=":
		c.emit(code.OpAdd)
	case "-=":
		c.emit(code.OpSub)
	case "*=":
		c.emit(code.OpMul)
	case "/=":
		c.emit(code.OpDiv)
	case "%=":
		c.emit(code.OpMod)
	case "&=":
		c.emit(code.OpBitAnd)
	case "|=":
		c.emit(code.OpBitOr)
	case "^=":
		c.emit(code.OpBitXor)
	case "<<=":
		c.emit(code.OpLShift)
	case ">>=":
		c.emit(code.OpRShift)
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) compileCompoundIndex(node *ast.CompoundAssignStatement, target *ast.IndexExpression) error {
	// Step 1: load current value arr[i]
	if err := c.Compile(target.Left); err != nil {
		return err
	}
	if err := c.Compile(target.Index); err != nil {
		return err
	}
	c.emit(code.OpIndex)

	// Step 2: compile RHS and apply operator
	if err := c.Compile(node.Value); err != nil {
		return err
	}
	c.emitCompoundOp(node.Operator)

	// Step 3: save result to a temp local
	resultSym := c.symbolTable.Define("__ci_result__", true)
	c.setSymbol(resultSym)

	// Step 4: store result back to arr[i]
	if err := c.Compile(target.Left); err != nil {
		return err
	}
	if err := c.Compile(target.Index); err != nil {
		return err
	}
	c.loadSymbol(resultSym)
	c.emit(code.OpIndexSet)

	return nil
}

func (c *Compiler) compileCompoundField(node *ast.CompoundAssignStatement, target *ast.FieldAccessExpression) error {
	// Step 1: load current value obj.field
	if err := c.Compile(target.Left); err != nil {
		return err
	}
	nameIdx := c.addConstant(&object.String{Value: target.Field.Value})
	c.emit(code.OpGetField, nameIdx)

	// Step 2: compile RHS and apply operator
	if err := c.Compile(node.Value); err != nil {
		return err
	}
	c.emitCompoundOp(node.Operator)

	// Step 3: save result
	resultSym := c.symbolTable.Define("__cf_result__", true)
	c.setSymbol(resultSym)

	// Step 4: store result back to obj.field
	if err := c.Compile(target.Left); err != nil {
		return err
	}
	c.loadSymbol(resultSym)
	nameIdx2 := c.addConstant(&object.String{Value: target.Field.Value})
	c.emit(code.OpSetField, nameIdx2)

	return nil
}

func (c *Compiler) compileForIndex(node *ast.ForIndexStatement) error {
	iterSym := c.symbolTable.Define("__fi_iter__", true)
	counterSym := c.symbolTable.Define("__fi_counter__", true)
	elemSym := c.symbolTable.Define(node.Variable.Value, true)
	indexSym := c.symbolTable.Define(node.Index.Value, true)

	// __fi_iter__ = iterable
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	c.setSymbol(iterSym)

	// __fi_counter__ = 0
	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 0}))
	c.setSymbol(counterSym)

	loopStart := len(c.currentInstructions())
	c.loopStarts = append(c.loopStarts, loopStart)
	c.breakPos = append(c.breakPos, []int{})

	// if len(__fi_iter__) <= __fi_counter__ { break }
	c.emit(code.OpGetBuiltin, builtinIndex("len"))
	c.loadSymbol(iterSym)
	c.emit(code.OpCall, 1)
	c.loadSymbol(counterSym)
	c.emit(code.OpGreater)
	exitPos := c.emit(code.OpJumpFalse, 9999)

	// indexVar = __fi_counter__
	c.loadSymbol(counterSym)
	c.setSymbol(indexSym)

	// elemVar = __fi_iter__[__fi_counter__]
	c.loadSymbol(iterSym)
	c.loadSymbol(counterSym)
	c.emit(code.OpIndex)
	c.setSymbol(elemSym)

	// body
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// __fi_counter__++
	c.loadSymbol(counterSym)
	c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 1}))
	c.emit(code.OpAdd)
	c.setSymbol(counterSym)

	c.emitLoop(loopStart)
	afterLoop := len(c.currentInstructions())
	c.replaceOperand(exitPos, afterLoop)

	breaks := c.breakPos[len(c.breakPos)-1]
	for _, bp := range breaks {
		c.replaceOperand(bp, afterLoop)
	}
	c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
	c.breakPos = c.breakPos[:len(c.breakPos)-1]

	return nil
}

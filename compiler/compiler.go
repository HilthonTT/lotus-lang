package compiler

import (
	"fmt"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/code"
	"github.com/hilthontt/lotus/object"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

// EmittedInstruction represents an instruction through an opcode and it's position
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// CompilationScope - Before we start compiling a function's body (enter a new scope) we push
// a new CompilationScope on to the scopes stack. While compiling inside this scope, the emit
// method of the compiler will modify only the fields of the current CompilationScope. Once we're
// done compiling the function, we leave the scope by popping it off the scopes stack and putting
// the instructions in a new *object.CompiledFunction
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

	// For break/continue
	loopStarts []int   // stack of loop start positions
	breakPos   [][]int // stack of break placeholder positions

	inClass bool // true while compiling class methods (enables 'super')
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

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

// Bytecode returns a pointer to a Bytecode intialized with our compilers instructions & constants
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

// currentInstructions returns the instructions for the current scopes index
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

// Compile walks the AST recursively and compiles nodes
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
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

		// OOP statements

	case *ast.ClassStatement:
		if err := c.compileClass(node); err != nil {
			return err
		}

	case *ast.FieldAssignStatement:
		// obj.field = value  ->  push obj, push value, OpSetField
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

		// Patch breaks
		breaks := c.breakPos[len(c.breakPos)-1]
		for _, bp := range breaks {
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
		c.setSymbol(elemSym) // was loadSymbol → STORE, not load

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		c.loadSymbol(counterSym)
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: 1}))
		c.emit(code.OpAdd)
		c.setSymbol(counterSym) // was loadSymbol → STORE, not load

		c.emitLoop(loopStart)
		afterLoop := len(c.currentInstructions())
		c.replaceOperand(exitPos, afterLoop)

		breaks := c.breakPos[len(c.breakPos)-1]
		for _, bp := range breaks {
			c.replaceOperand(bp, afterLoop)
		}
		c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
		c.breakPos = c.breakPos[:len(c.breakPos)-1]

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

	// OOP expressions

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

		// Load -> mutate -> store -> reload
		// The reload is required because ExpressionStatement always emits OpPop.
		// It also makes postfix usable as a sub-expression: let y = x++
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
			c.emit(code.OpGetGlobal, symbol.Index) // keep a value on the stack for OpPop
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
			c.emit(code.OpGetLocal, symbol.Index) // same for locals
		}

	case *ast.InfixExpression:
		// Handle < and <= by swapping operands
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

		// Left first for everything else
		if err := c.Compile(node.Left); err != nil {
			return err
		}

		// Short-circuit &&
		if node.Operator == "&&" {
			falseJump := c.emit(code.OpJumpFalse, 9999)
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			endJump := c.emit(code.OpJump, 9999)
			falsePos := len(c.currentInstructions())
			c.replaceOperand(falseJump, falsePos)
			c.emit(code.OpFalse)
			endPos := len(c.currentInstructions())
			c.replaceOperand(endJump, endPos)
			return nil
		}

		// Short-circuit ||
		if node.Operator == "||" {
			falseJump := c.emit(code.OpJumpFalse, 9999)
			c.emit(code.OpTrue)
			endJump := c.emit(code.OpJump, 9999)
			falsePos := len(c.currentInstructions())
			c.replaceOperand(falseJump, falsePos)
			if err := c.Compile(node.Right); err != nil {
				return err
			}
			endPos := len(c.currentInstructions())
			c.replaceOperand(endJump, endPos)
			return nil
		}

		// Right, then emit the operator
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
			c.removeLastPop() // expression value stays on stack
		} else if !c.lastInstructionIs(code.OpReturn) {
			// Block ended with a non-expression (e.g. assignment) — push nil
			// so the if-expression always leaves a value
			c.emit(code.OpNil)
		}

		endJump := c.emit(code.OpJump, 9999)
		afterConsequence := len(c.currentInstructions())
		c.replaceOperand(falseJump, afterConsequence)

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
		afterAlternative := len(c.currentInstructions())
		c.replaceOperand(endJump, afterAlternative)

	case *ast.FunctionLiteral:
		// For named functions, define the symbol BEFORE entering scope
		// so recursive references resolve correctly
		var outerSym *Symbol
		if node.Name != "" {
			s := c.symbolTable.Define(node.Name, false)
			outerSym = &s
		}

		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false)
		}
		if err := c.Compile(node.Body); err != nil {
			return err
		}
		// Implicit return nil if last instruction isn't a return
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
		}
		fnIdx := c.addConstant(fn)
		c.emit(code.OpClosure, fnIdx, len(freeSymbols))

		// If named function, bind to the pre-defined symbol
		if outerSym != nil {
			if outerSym.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, outerSym.Index)
				c.emit(code.OpGetGlobal, outerSym.Index)
			} else {
				c.emit(code.OpSetLocal, outerSym.Index)
				c.emit(code.OpGetLocal, outerSym.Index)
			}
		}

	case *ast.CallExpression:
		// Method call: obj.method(args)  or  super.method(args)
		if fieldAccess, ok := node.Function.(*ast.FieldAccessExpression); ok {
			if _, isSuper := fieldAccess.Left.(*ast.SuperExpression); isSuper {
				// super.method(args): push SuperAccessor, then args, then OpInvokeMethod
				if !c.inClass {
					return fmt.Errorf("'super' used outside of a class method")
				}
				c.emit(code.OpGetSuper)
			} else {
				// regular obj.method(args): push obj, then args
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

		// Regular function call
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
	}

	return nil
}

// compileClass emits bytecode for a class definition.
//
//	OpNewClass name_idx
//	[OpLoadSuperClass; OpSetSuper]
//	for each method: OpClosure ...; OpDefineMethod name_idx
//	OpSetGlobal/Local
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

// compileMethod compiles a single class method into a closure on the stack.
// It is the caller's responsibility to emit OpDefineMethod afterwards.
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
	newInstructionPos := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return newInstructionPos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
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
	ins := code.Make(code.OpLoop, 0) // placeholder
	pos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	// After OpLoop executes at ip=pos, VM does: frame.ip -= offset
	// Then next iteration does ip++, executing at pos - offset + 1
	// We want pos - offset + 1 = loopStart, so offset = pos - loopStart + 1
	offset := pos - loopStart + 1
	c.scopes[c.scopeIndex].instructions[pos+1] = byte(offset >> 8)
	c.scopes[c.scopeIndex].instructions[pos+2] = byte(offset)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	ins := c.currentInstructions()
	if len(ins) == 0 {
		return false
	}

	return ins[len(ins)-1] == byte(op)
}

func (c *Compiler) removeLastPop() {
	c.scopes[c.scopeIndex].instructions = c.currentInstructions()[:len(c.currentInstructions())-1]
}

func (c *Compiler) replaceLastPopWithReturn() {
	ins := c.currentInstructions()
	ins[len(ins)-1] = byte(code.OpReturn)
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants

	return compiler
}

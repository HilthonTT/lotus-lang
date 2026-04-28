package formatter

import (
	"fmt"
	"strings"

	"github.com/hilthontt/lotus/ast"
	"github.com/hilthontt/lotus/lexer"
)

const tab = "    "

// Formatter pretty-prints a Lotus AST back to canonical source,
// preserving blank lines and comments from the original.
//
// Blank line rule: if two consecutive things (comment or statement)
// have a source-line gap > 1, one blank line is emitted between them.
//
// Block invariant: block() writes "{\n...\n}" with NO trailing newline.
// Every caller adds \n (or " else "/ " catch ") itself.
// This keeps "} else {" and "} catch {" on the same line as the closing brace.
type Formatter struct {
	sb       strings.Builder
	depth    int
	comments []lexer.CommentToken // sorted ascending by Line
	nextCmt  int                  // next unwritten comment index
	lastLine int                  // source line of the last thing we wrote (0 = nothing yet)
}

// Format returns the canonical source for program.
// Pass comments from lexer.Comments(); pass nil for no comment preservation.
func Format(program *ast.Program, comments []lexer.CommentToken) string {
	f := &Formatter{comments: comments}
	f.program(program)
	return f.sb.String()
}

// ── primitives ────────────────────────────────────────────────────────────────

func (f *Formatter) w(s string)  { f.sb.WriteString(s) }
func (f *Formatter) nl()         { f.sb.WriteByte('\n') }
func (f *Formatter) ind()        { f.sb.WriteString(strings.Repeat(tab, f.depth)) }
func (f *Formatter) in()         { f.depth++ }
func (f *Formatter) out()        { f.depth-- }
func (f *Formatter) cur() string { return strings.Repeat(tab, f.depth) }

// ── blank lines + comment injection ──────────────────────────────────────────

// advanceTo writes any pending comments whose source line < targetLine,
// and inserts a blank line whenever the gap between consecutive items is > 1.
// Call this before writing any statement or the trailing comments.
func (f *Formatter) advanceTo(targetLine int) {
	// flush comments that precede targetLine
	for f.nextCmt < len(f.comments) && f.comments[f.nextCmt].Line < targetLine {
		cLine := f.comments[f.nextCmt].Line
		if f.lastLine > 0 && cLine > f.lastLine+1 {
			f.nl() // blank line before comment
		}
		f.ind()
		f.w(f.comments[f.nextCmt].Text)
		f.nl()
		f.lastLine = cLine
		f.nextCmt++
	}
	// blank line before the statement itself
	if f.lastLine > 0 && targetLine > f.lastLine+1 {
		f.nl()
	}
}

// setLine records that we just wrote something on sourceLine.
func (f *Formatter) setLine(sourceLine int) {
	if sourceLine > 0 {
		f.lastLine = sourceLine
	}
}

// stmtLine returns the first source line of a statement.
func stmtLine(s ast.Statement) int {
	switch n := s.(type) {
	case *ast.LetStatement:
		return n.Token.Line
	case *ast.AssignStatement:
		return n.Token.Line
	case *ast.CompoundAssignStatement:
		return n.Token.Line
	case *ast.IndexAssignStatement:
		return n.Token.Line
	case *ast.FieldAssignStatement:
		return n.Token.Line
	case *ast.MultiLetStatement:
		return n.Token.Line
	case *ast.MultiAssignStatement:
		return n.Token.Line
	case *ast.ArrayDestructureStatement:
		return n.Token.Line
	case *ast.MapDestructureStatement:
		return n.Token.Line
	case *ast.ReturnStatement:
		return n.Token.Line
	case *ast.ExpressionStatement:
		return n.Token.Line
	case *ast.WhileStatement:
		return n.Token.Line
	case *ast.ForStatement:
		return n.Token.Line
	case *ast.ForIndexStatement:
		return n.Token.Line
	case *ast.BreakStatement:
		return n.Token.Line
	case *ast.ContinueStatement:
		return n.Token.Line
	case *ast.ClassStatement:
		return n.Token.Line
	case *ast.InterfaceStatement:
		return n.Token.Line
	case *ast.EnumStatement:
		return n.Token.Line
	case *ast.ImportStatement:
		return n.Token.Line
	case *ast.ExportStatement:
		return n.Token.Line
	case *ast.DeferStatement:
		return n.Token.Line
	case *ast.ThrowStatement:
		return n.Token.Line
	case *ast.TryCatchStatement:
		return n.Token.Line
	case *ast.BlockStatement:
		return n.Token.Line
	}
	return 0
}

// ── program ───────────────────────────────────────────────────────────────────

func (f *Formatter) program(p *ast.Program) {
	for _, s := range p.Statements {
		f.stmt(s)
	}
	// flush any trailing comments after the last statement
	for f.nextCmt < len(f.comments) {
		cLine := f.comments[f.nextCmt].Line
		if f.lastLine > 0 && cLine > f.lastLine+1 {
			f.nl()
		}
		f.ind()
		f.w(f.comments[f.nextCmt].Text)
		f.nl()
		f.lastLine = cLine
		f.nextCmt++
	}
}

// ── statements ────────────────────────────────────────────────────────────────

func (f *Formatter) stmt(s ast.Statement) {
	line := stmtLine(s)
	f.advanceTo(line)
	f.setLine(line)

	switch n := s.(type) {
	case *ast.LetStatement:
		f.ind()
		f.letStmt(n)
	case *ast.AssignStatement:
		f.ind()
		f.w(n.Name.Value + " = ")
		f.expr(n.Value)
		f.nl()
	case *ast.CompoundAssignStatement:
		f.ind()
		f.expr(n.Name)
		f.w(" " + n.Operator + " ")
		f.expr(n.Value)
		f.nl()
	case *ast.IndexAssignStatement:
		f.ind()
		f.expr(n.Left)
		f.w("[")
		f.expr(n.Index)
		f.w("] = ")
		f.expr(n.Value)
		f.nl()
	case *ast.FieldAssignStatement:
		f.ind()
		f.expr(n.Left)
		f.w("." + n.Field.Value + " = ")
		f.expr(n.Value)
		f.nl()
	case *ast.MultiLetStatement:
		f.ind()
		f.multiLet(n)
	case *ast.MultiAssignStatement:
		f.ind()
		f.multiAssign(n)
	case *ast.ArrayDestructureStatement:
		f.ind()
		f.arrDestr(n)
	case *ast.MapDestructureStatement:
		f.ind()
		f.mapDestr(n)
	case *ast.ReturnStatement:
		f.ind()
		if n.ReturnValue == nil {
			f.w("return")
		} else {
			f.w("return ")
			f.expr(n.ReturnValue)
		}
		f.nl()
	case *ast.ExpressionStatement:
		f.ind()
		f.expr(n.Expression)
		f.nl()
	case *ast.WhileStatement:
		f.ind()
		f.w("while ")
		f.expr(n.Condition)
		f.w(" ")
		f.block(n.Body)
		f.nl()
	case *ast.ForStatement:
		f.ind()
		f.w("for " + n.Variable.Value + " in ")
		f.expr(n.Iterable)
		f.w(" ")
		f.block(n.Body)
		f.nl()
	case *ast.ForIndexStatement:
		f.ind()
		f.w("for " + n.Index.Value + ", " + n.Variable.Value + " in ")
		f.expr(n.Iterable)
		f.w(" ")
		f.block(n.Body)
		f.nl()
	case *ast.BreakStatement:
		f.ind()
		f.w("break")
		f.nl()
	case *ast.ContinueStatement:
		f.ind()
		f.w("continue")
		f.nl()
	case *ast.DeferStatement:
		f.ind()
		f.w("defer ")
		f.expr(n.Call)
		f.nl()
	case *ast.ThrowStatement:
		f.ind()
		f.w("throw ")
		f.expr(n.Value)
		f.nl()
	case *ast.TryCatchStatement:
		f.ind()
		f.w("try ")
		f.block(n.Try)
		f.w(" catch")
		if n.CatchVar != nil {
			f.w(" " + n.CatchVar.Value)
		}
		f.w(" ")
		f.block(n.Catch)
		f.nl()
	case *ast.ClassStatement:
		f.ind()
		f.class(n)
	case *ast.InterfaceStatement:
		f.ind()
		f.iface(n)
	case *ast.EnumStatement:
		f.ind()
		f.enum(n)
	case *ast.ImportStatement:
		f.ind()
		names := make([]string, len(n.Names))
		for i, nm := range n.Names {
			names[i] = nm.Value
		}
		f.w(`import { ` + strings.Join(names, ", ") + ` } from "` + n.Path + `"`)
		f.nl()
	case *ast.ExportStatement:
		f.ind()
		f.w("export ")
		f.stmtInline(n.Statement)
	case *ast.BlockStatement:
		f.ind()
		f.block(n)
		f.nl()
	default:
		f.ind()
		f.w(s.String())
		f.nl()
	}
}

// stmtInline writes without leading indent (used after `export`).
func (f *Formatter) stmtInline(s ast.Statement) {
	switch n := s.(type) {
	case *ast.LetStatement:
		f.letStmt(n)
	case *ast.ExpressionStatement:
		f.expr(n.Expression)
		f.nl()
	case *ast.ClassStatement:
		f.class(n)
	default:
		f.w(s.String())
		f.nl()
	}
}

func (f *Formatter) letStmt(n *ast.LetStatement) {
	if n.Mutable {
		f.w("mut ")
	} else {
		f.w("let ")
	}
	f.w(n.Name.Value)
	if n.TypeAnnot != nil {
		f.w(": " + n.TypeAnnot.Name)
	}
	f.w(" = ")
	f.expr(n.Value)
	f.nl()
}

func (f *Formatter) multiLet(n *ast.MultiLetStatement) {
	if n.Mutable {
		f.w("mut ")
	} else {
		f.w("let ")
	}
	ns := make([]string, len(n.Names))
	for i, x := range n.Names {
		ns[i] = x.Value
	}
	vs := make([]string, len(n.Values))
	for i, x := range n.Values {
		vs[i] = f.es(x)
	}
	f.w(strings.Join(ns, ", ") + " = " + strings.Join(vs, ", "))
	f.nl()
}

func (f *Formatter) multiAssign(n *ast.MultiAssignStatement) {
	ns := make([]string, len(n.Names))
	for i, x := range n.Names {
		ns[i] = f.es(x)
	}
	vs := make([]string, len(n.Values))
	for i, x := range n.Values {
		vs[i] = f.es(x)
	}
	f.w(strings.Join(ns, ", ") + " = " + strings.Join(vs, ", "))
	f.nl()
}

func (f *Formatter) arrDestr(n *ast.ArrayDestructureStatement) {
	kw := "let "
	if n.Mutable {
		kw = "mut "
	}
	ps := make([]string, len(n.Names))
	for i, x := range n.Names {
		ps[i] = f.es(x)
	}
	f.w(kw + "[" + strings.Join(ps, ", ") + "] = ")
	f.expr(n.Value)
	f.nl()
}

func (f *Formatter) mapDestr(n *ast.MapDestructureStatement) {
	kw := "let "
	if n.Mutable {
		kw = "mut "
	}
	ks := make([]string, len(n.Keys))
	for i, x := range n.Keys {
		ks[i] = x.Value
	}
	f.w(kw + "{ " + strings.Join(ks, ", ") + " } = ")
	f.expr(n.Value)
	f.nl()
}

// ── block ─────────────────────────────────────────────────────────────────────

// block writes "{ \n ... \n}" with NO trailing newline.
// After writing, it bumps f.lastLine by 1 to account for the "}" line,
// so blank-line detection works correctly for the statement that follows.
func (f *Formatter) block(b *ast.BlockStatement) {
	f.w("{\n")
	f.in()
	for _, s := range b.Statements {
		f.stmt(s)
	}
	f.out()
	f.ind()
	f.w("}")
	f.lastLine++ // approximate the "}" line so gaps after blocks are detected correctly
}

// ── class / interface / enum ──────────────────────────────────────────────────

func (f *Formatter) class(n *ast.ClassStatement) {
	f.w("class " + n.Name.Value)
	if len(n.TypeParams) > 0 {
		ps := make([]string, len(n.TypeParams))
		for i, tp := range n.TypeParams {
			if tp.Constraint != "" {
				ps[i] = tp.Name + ": " + tp.Constraint
			} else {
				ps[i] = tp.Name
			}
		}
		f.w("<" + strings.Join(ps, ", ") + ">")
	}
	if n.SuperClass != nil {
		f.w(" extends " + n.SuperClass.Value)
	}
	f.w(" {\n")
	f.in()
	for i, m := range n.Methods {
		f.ind()
		f.fnWrite(m)
		if i < len(n.Methods)-1 {
			f.nl()
		}
	}
	f.out()
	f.ind()
	f.w("}\n")
}

func (f *Formatter) iface(n *ast.InterfaceStatement) {
	f.w("interface " + n.Name.Value + " {\n")
	f.in()
	for _, m := range n.Methods {
		f.ind()
		f.w("fn " + m.Name + "(self")
		for i, pt := range m.ParamTypes {
			f.w(", " + fmt.Sprintf("p%d", i))
			if pt != nil {
				f.w(": " + pt.Name)
			}
		}
		f.w(")")
		if m.ReturnType != nil {
			f.w(" -> " + m.ReturnType.Name)
		}
		f.nl()
	}
	f.out()
	f.ind()
	f.w("}\n")
}

func (f *Formatter) enum(n *ast.EnumStatement) {
	f.w("enum " + n.Name.Value + " {\n")
	f.in()
	for i, v := range n.Variants {
		f.ind()
		f.w(v.Name)
		if len(v.Fields) > 0 {
			f.w("(" + strings.Join(v.Fields, ", ") + ")")
		}
		if i < len(n.Variants)-1 {
			f.w(",")
		}
		f.nl()
	}
	f.out()
	f.ind()
	f.w("}\n")
}

// ── expressions ───────────────────────────────────────────────────────────────

// expr writes e to the buffer. Block-containing expressions write inline.
func (f *Formatter) expr(e ast.Expression) {
	switch n := e.(type) {
	case *ast.IfExpression:
		f.ifWrite(n)
	case *ast.FunctionLiteral:
		f.fnWrite(n)
	case *ast.MatchExpression:
		f.matchWrite(n)
	default:
		f.w(f.es(e))
	}
}

// es converts a simple expression to a string.
// For block-containing expressions it captures expr() output and trims
// the trailing newline so they embed cleanly inside larger expressions.
func (f *Formatter) es(e ast.Expression) string {
	if e == nil {
		return ""
	}
	switch n := e.(type) {
	case *ast.Identifier:
		return n.Value
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", n.Value)
	case *ast.FloatLiteral:
		s := fmt.Sprintf("%g", n.Value)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case *ast.StringLiteral:
		return `"` + escStr(n.Value) + `"`
	case *ast.BooleanLiteral:
		if n.Value {
			return "true"
		}
		return "false"
	case *ast.NilLiteral:
		return "nil"
	case *ast.SelfExpression:
		return "self"
	case *ast.SuperExpression:
		return "super"
	case *ast.PrefixExpression:
		return n.Operator + f.es(n.Right)
	case *ast.PostfixExpression:
		return n.Token.Literal + n.Operator
	case *ast.InfixExpression:
		return f.es(n.Left) + " " + n.Operator + " " + f.es(n.Right)
	case *ast.TernaryExpression:
		return f.es(n.Condition) + " ? " + f.es(n.Consequence) + " : " + f.es(n.Alternative)
	case *ast.IndexExpression:
		return f.es(n.Left) + "[" + f.es(n.Index) + "]"
	case *ast.FieldAccessExpression:
		return f.es(n.Left) + "." + n.Field.Value
	case *ast.OptionalFieldAccess:
		return f.es(n.Left) + "?." + n.Field.Value
	case *ast.SpreadExpression:
		return "..." + f.es(n.Value)
	case *ast.PipeExpression:
		return f.es(n.Left) + " |> " + f.es(n.Right)
	case *ast.CallExpression:
		args := make([]string, len(n.Arguments))
		for i, a := range n.Arguments {
			args[i] = f.es(a)
		}
		return f.es(n.Function) + "(" + strings.Join(args, ", ") + ")"
	case *ast.ArrayLiteral:
		return f.arrLit(n)
	case *ast.MapLiteral:
		return f.mapLit(n)
	// Block-containing: capture and strip trailing \n for inline embedding.
	case *ast.IfExpression, *ast.FunctionLiteral, *ast.MatchExpression:
		saved := f.sb
		f.sb = strings.Builder{}
		f.expr(e)
		result := strings.TrimRight(f.sb.String(), "\n")
		f.sb = saved
		return result
	default:
		return e.String()
	}
}

func (f *Formatter) arrLit(n *ast.ArrayLiteral) string {
	if len(n.Elements) == 0 {
		return "[]"
	}
	ps := make([]string, len(n.Elements))
	for i, el := range n.Elements {
		ps[i] = f.es(el)
	}
	if s := "[" + strings.Join(ps, ", ") + "]"; len(s) <= 60 {
		return s
	}
	var sb strings.Builder
	sb.WriteString("[\n")
	f.in()
	for _, p := range ps {
		sb.WriteString(f.cur() + p + ",\n")
	}
	f.out()
	sb.WriteString(f.cur() + "]")
	return sb.String()
}

func (f *Formatter) mapLit(n *ast.MapLiteral) string {
	if len(n.Keys) == 0 {
		return "{}"
	}
	ps := make([]string, len(n.Keys))
	for i, k := range n.Keys {
		ps[i] = f.es(k) + ": " + f.es(n.Pairs[k])
	}
	if s := "{ " + strings.Join(ps, ", ") + " }"; len(s) <= 60 {
		return s
	}
	var sb strings.Builder
	sb.WriteString("{\n")
	f.in()
	for _, p := range ps {
		sb.WriteString(f.cur() + p + ",\n")
	}
	f.out()
	sb.WriteString(f.cur() + "}")
	return sb.String()
}

// ── block-containing expression writers ──────────────────────────────────────

func (f *Formatter) ifWrite(n *ast.IfExpression) {
	f.w("if " + f.es(n.Condition) + " ")
	f.block(n.Consequence)
	if n.Alternative != nil {
		f.w(" else ")
		f.block(n.Alternative)
	}
	// no trailing \n — stmt() adds it
}

func (f *Formatter) fnWrite(n *ast.FunctionLiteral) {
	f.w("fn")
	if n.Name != "" {
		f.w(" " + n.Name)
	}
	if len(n.TypeParams) > 0 {
		ps := make([]string, len(n.TypeParams))
		for i, tp := range n.TypeParams {
			ps[i] = tp.Name
		}
		f.w("<" + strings.Join(ps, ", ") + ">")
	}
	params := make([]string, len(n.Parameters))
	for i, p := range n.Parameters {
		s := p.Value
		if i == len(n.Parameters)-1 && n.IsVariadic {
			s = "..." + s
		} else if i < len(n.ParamTypes) && n.ParamTypes[i] != nil {
			s += ": " + n.ParamTypes[i].Name
		}
		params[i] = s
	}
	f.w("(" + strings.Join(params, ", ") + ")")
	if n.ReturnType != nil {
		f.w(" -> " + n.ReturnType.Name)
	}
	f.w(" ")
	f.block(n.Body)
	// no trailing \n
}

func (f *Formatter) matchWrite(n *ast.MatchExpression) {
	f.w("match " + f.es(n.Subject) + " {\n")
	f.in()
	for _, arm := range n.Arms {
		f.ind()
		if arm.IsWild {
			f.w("_")
		} else {
			f.w(f.es(arm.Pattern))
		}
		f.w(" -> " + f.es(arm.Body) + ",\n")
	}
	f.out()
	f.ind()
	f.w("}")
}

// ── string escaping ───────────────────────────────────────────────────────────

func escStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

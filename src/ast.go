package src

import "fmt"

type Visitor interface {
	Visit(node Node)
}

type AST struct {
	Root Node
}

type Program struct {
	Statements []Node
}

func (p *Program) Print() {
	for _, stmt := range p.Statements {
		stmt.Print()
	}
}

func (p *Program) Accept(visitor Visitor) {
	for _, stmt := range p.Statements {
		stmt.Accept(visitor)
	}
}

func (a *AST) Accept(visitor Visitor) {
	visitor.Visit(a.Root)
}

type VisitorFunc func(node Node)

// Node is the base interface for all AST nodes
type Node interface {
	Accept(visitor Visitor)
	Print()
}

type Expr interface {
	Node
}

type BinaryExpr struct {
	Left     Expr
	Operator tokenKind // "+", "-", "*", "/", ">", ">=", etc.
	Right    Expr
}

func (b *BinaryExpr) Print() {
	fmt.Printf("BinaryExpr: %v %v %v\n", b.Left, b.Operator.ToString(), b.Right)
}

func (b *BinaryExpr) Accept(visitor Visitor) {
	visitor.Visit(b)
}

type UnaryExpr struct {
	Operator tokenKind
	Operand  Expr
}

func (u *UnaryExpr) Accept(visitor Visitor) {
	visitor.Visit(u)
}

func (u *UnaryExpr) Print() {
	fmt.Printf("UnaryExpr: %v %v\n", u.Operator.ToString(), u.Operand)
}

type StringLiteral struct {
	string
}

func (s *StringLiteral) Print() {
	fmt.Printf("StringLiteral: %s\n", s.string)
}
func (s *StringLiteral) Accept(visitor Visitor) {
	visitor.Visit(s)
}

type NumLiteral struct {
	Value int
}

type BoolLiteral struct {
	bool
}

func (b *BoolLiteral) Accept(visitor Visitor) {
	visitor.Visit(b)
}

func (b *BoolLiteral) Print() {
	fmt.Printf("BoolLiteral: %v\n", b.bool)
}

func IsConstExpr(expr Expr) bool {
	switch e := expr.(type) {
	case *NumLiteral:
		return true
	case *Ident:
		return false
	case *BinaryExpr:
		return IsConstExpr(e.Left) && IsConstExpr(e.Right)
	case *UnaryExpr:
		return IsConstExpr(e.Operand)
	default:
		return false
	}
}

func (n *NumLiteral) Accept(visitor Visitor) {
	visitor.Visit(n)
}
func (n *NumLiteral) Print() {
	fmt.Printf("NumLiteral: %v\n", n.Value)
}

type Ident struct {
	Name string
}

type Array struct {
	Items []Expr
}

func (a *Array) Accept(visitor Visitor) {
	visitor.Visit(a)
}

func (a *Array) Print() {
	fmt.Printf("Array: %v\n", a.Items)
}

type ForLoop struct {
	Var       Expr
	Start     Expr
	Condition Expr
	Step      Expr
	Body      Node
}

func (f *ForLoop) Accept(visitor Visitor) {
	visitor.Visit(f)
}

func (f *ForLoop) Print() {
	fmt.Printf("ForLoop: %v %v %v %v\n", f.Var, f.Start, f.Condition, f.Step)
}

func (i *Ident) Accept(visitor Visitor) {
	visitor.Visit(i)
}

func (i *Ident) Print() {
	fmt.Printf("Ident: %s\n", i.Name)
}

type Stmt interface {
	Node
}

type LetExpr struct {
	Variable Ident
	Value    Expr
}

type ReAssignExpr struct {
	Variable Ident
	NewValue Expr
}

func (r *ReAssignExpr) Accept(visitor Visitor) {
	visitor.Visit(r)
}

func (r *ReAssignExpr) Print() {
	fmt.Printf("ReAssignExpr: %s = %v\n", r.Variable, r.NewValue)
}

func (a *LetExpr) Accept(visitor Visitor) {
	visitor.Visit(a)
}

func (a *LetExpr) Print() {
	fmt.Printf("LetExpr: %v = %s\n", a.Variable, a.Value)
}

type PrintCall struct {
	Value Expr
}

func (p *PrintCall) Accept(visitor Visitor) {
	visitor.Visit(p)
}

func (p *PrintCall) Print() {
	fmt.Printf("PrintCall: %v\n", p.Value)
}

type InputIntCall struct {
	Input Expr
}

type InputStrCall struct {
	Input Expr
}

func (i *InputStrCall) Accept(visitor Visitor) {
	visitor.Visit(i)
}
func (i *InputStrCall) Print() {
	fmt.Printf("InputStrCall %s", i.Input)
}

type CallExpr struct {
	Function    Ident
	Args        FuncArgs
	IsRecursive bool
	IsTail      bool
}

type FuncArg struct {
	Value Expr
}

func (f *FuncArg) Accept(visitor Visitor) {
	visitor.Visit(f)
}

func (f *FuncArg) Print() {
	fmt.Printf("FuncArg: %v\n", f.Value)
}

func (c *CallExpr) Accept(visitor Visitor) {
	visitor.Visit(c)
}
func (c *CallExpr) Print() {
	fmt.Printf("CallExpr: %v %v\n", c.Function, c.Args)
}

type ReturnExpr struct {
	Value Expr
}

func (r *ReturnExpr) Accept(visitor Visitor) {
	visitor.Visit(r)
}

func (r *ReturnExpr) Print() {
	fmt.Printf("ReturnExpr: %v\n", r.Value)
}

type FuncDef struct {
	Name    Ident
	Params  []FnParam
	Body    *Block
	RetType tokenKind
}

type FnParam struct {
	Name string
	Type tokenKind
}

type FuncArgs struct {
	Args []FuncArg
}

func (f *FuncArgs) Accept(visitor Visitor) {
	visitor.Visit(f)
}

func (f *FuncArgs) Print() {
	fmt.Printf("FuncArgs: %v\n", f.Args)
}

type Block struct {
	Statements []Node
}

func (b *Block) Print() {
	for _, stmt := range b.Statements {
		stmt.Print()
	}
}
func (b *Block) Accept(visitor Visitor) {
	visitor.Visit(b)
}

func (f *FuncDef) Accept(visitor Visitor) {
	visitor.Visit(f)
}

func (f *FuncDef) Print() {
	fmt.Printf("Func: body: %v, params: %v, retType: %s", f.Body, f.Params, f.RetType.ToString())
}

func (i *InputIntCall) Accept(visitor Visitor) {
	visitor.Visit(i)
}

func (i *InputIntCall) Print() {
	fmt.Printf("InputCall %s", i.Input)
}

// IfStmt represents a conditional expression (e.g., `if ... then ... else ...`)
type IfStmt struct {
	Condition Expr
	IfBlock   Node
	ElseBlock Node
}

func (i *IfStmt) Accept(visitor Visitor) {
	visitor.Visit(i)
}
func (i *IfStmt) Print() {
	fmt.Printf("IfExpr: if %v then %v\n else %v\n", i.Condition, i.IfBlock, i.ElseBlock)
}

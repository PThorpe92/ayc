package src

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
)

func isUnaryOperator(tk tokenKind) bool {
	switch tk {
	case Minus, Bang, BitNot:
		return true
	default:
		return false
	}
}

type Stack[T any] struct {
	t []T
}

func (s *Stack[T]) Get(idx int) T {
	return s.t[idx]
}

func (s *Stack[T]) Push(v T) {
	s.t = append(s.t, v)
}

func (s *Stack[T]) Pop() T {
	v := s.t[len(s.t)-1]
	s.t = s.t[:len(s.t)-1]
	return v
}

func (s *Stack[T]) Peek() T {
	return s.t[len(s.t)-1]
}

func (s *Stack[T]) Len() int {
	return len(s.t)
}

func isBinaryOperator(tk tokenKind) bool {
	switch tk {
	case Plus, Minus, Mul, Div, Mod, Gte, Gt, Lt, Lte, And, Or, Neq, LShift, RShift:
		return true
	default:
		return false
	}
}

type Parser struct {
	isRepl   bool
	input    string
	tokens   []Token
	pos      int
	currFunc *FuncDef
	Ast      *AST
}

func (parser *Parser) current() *Token {
	if parser.pos < len(parser.tokens) {
		return &parser.tokens[parser.pos]
	} else {
		return &Token{kind: EOF, val: ""}
	}
}

func (parser *Parser) peek() *Token {
	return &parser.tokens[parser.pos+1]
}

func (parser *Parser) next() *Token {
	parser.pos++
	return parser.current()
}

func (par *Parser) PrintError(tk *Token, expected tokenKind) string {
	line := tk.span.line
	errMsg := ""
	if expected == EOF {
		errMsg = fmt.Sprintf("Unexpected token %v on line %d", tk.kind.ToString(), line)
	} else {
		errMsg = fmt.Sprintf("Expected %v, got %v on line %d", expected.ToString(), tk.kind.ToString(), line)
	}
	relevantCode := strings.Split(par.input, "\n")[line-1]
	return fmt.Sprintf(`%s
		%s
		%s%s^^^^^^^^^^^^%s`, errMsg, relevantCode, strings.Repeat(" ", tk.span.len), Red, Reset)
}

func (par *Parser) assertToken(tk *Token, expected tokenKind, err ...string) error {
	if tk.kind != expected {
		fmt.Println(par.PrintError(tk, expected))
		fmt.Println(err)
		if !par.isRepl {
			os.Exit(1)
		}
		return fmt.Errorf("Expected %v, got %v", expected, tk.kind)
	}
	return nil
}

func NewParser(lxr *Lexer) *Parser {
	parser := &Parser{
		input:    lxr.src,
		tokens:   lxr.tokens,
		pos:      0,
		Ast:      nil,
		currFunc: nil,
	}
	return parser
}

func (par *Parser) Parse() *AST {
	var statements []Node
	for par.current().kind != EOF {
		stmt := par.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		} else {
			break
		}
	}
	par.Ast = &AST{Root: &Program{Statements: statements}}
	return par.Ast
}

func (par *Parser) parseStatement() Node {
	switch par.current().kind {
	case Let:
		return par.parseDeclaration()
	case Print:
		return par.parsePrintStatement()
	case Identifier:
		return par.parseExpression(0)
	case Defn:
		return par.parseFunctionDef()
	case Return:
		return par.parseReturnStatement()
	case If:
		return par.parseIfStatement()
	case LBrace:
		return par.parseBlock()
	case EOF:
		return nil
	default:
		if !par.isRepl {
			panic(fmt.Sprintf("Unexpected token: %v", par.current().val))
		} else {
			par.PrintError(par.current(), EOF)
			return nil
		}
	}
}

func (par *Parser) parsePrintStatement() Node {
	par.next()
	expr := par.parseExpression(0)
	return &PrintCall{Value: expr}
}

func (par *Parser) parseReturnStatement() Node {
	par.next()
	expr := par.parseExpression(0)
	return &ReturnExpr{Value: expr}
}

func (par *Parser) parseDeclaration() Node {
	slog.Debug("Parsing declaration. Current token: ", slog.String("token", par.current().kind.ToString()))
	par.next()
	ident := par.current()
	if err := par.assertToken(ident, Identifier, ""); err != nil {
		return nil
	}
	par.next()
	if err := par.assertToken(par.current(), Eq, ""); err != nil {
		return nil
	}
	par.next()
	return &LetExpr{
		Variable: Ident{Name: ident.val},
		Value:    par.parseExpression(0),
	}
}

func (par *Parser) parseFunctionCall(fName string) *CallExpr {
	par.next()
	var args []FuncArg
	for par.current().kind != RParen && par.current().kind != EOF {
		expr := par.parseExpression(0)
		args = append(args, FuncArg{expr})
		if par.current().kind == Comma {
			par.next()
		} else {
			break
		}
	}
	if err := par.assertToken(par.current(), RParen); err != nil {
		return nil
	}

	par.next()
	isRecursive := false
	if par.currFunc != nil && par.currFunc.Name.Name == fName {
		isRecursive = true
	}
	return &CallExpr{
		Function:    Ident{Name: fName},
		Args:        FuncArgs{args},
		IsRecursive: isRecursive,
	}
}

func isType(tk tokenKind) bool {
	return slices.Contains([]tokenKind{Int, String, Bool, Void}, tk)
}

func (par *Parser) parseFunctionDef() Node {
	par.next()
	if err := par.assertToken(par.current(), Identifier); err != nil {
		return nil
	}
	fName := par.current().val
	par.currFunc = &FuncDef{
		Name: Ident{fName},
	}
	par.next()
	if err := par.assertToken(par.current(), LParen); err != nil {
		return nil
	}

	par.next()
	params := []FnParam{}
	for par.current().kind != RParen {
		argName := ""
		if par.current().kind == Identifier {
			argName = par.current().val
			par.next()
		}
		typ := Void
		if par.current().kind == Colon {
			par.next()
			if !isType(par.current().kind) {
				par.assertToken(par.current(), EOF, "Expected type, got %v", par.current().val)
				return nil
			}
			typ = par.current().kind
			par.next()
		}
		if par.current().kind == Comma {
			par.next()
		}
		params = append(params, FnParam{Name: argName, Type: typ})
	}
	par.next() // )
	if err := par.assertToken(par.current(), Arrow); err != nil {
		return nil
	}

	par.next() // ->
	if !isType(par.current().kind) {
		panic(fmt.Sprintf("Expected type, got %v", par.current().kind))
	}
	retType := par.current().val
	par.next()
	if err := par.assertToken(par.current(), LBrace); err != nil {
		return nil
	}

	body := par.parseBlock().(*Block)
	def := &FuncDef{
		Name:    Ident{Name: fName},
		Params:  params,
		Body:    body,
		RetType: retType,
	}
	par.currFunc = def
	return def
}

func (par *Parser) parseAssignment(left Expr) Expr {
	op := par.current().kind
	par.next()
	right := par.parseExpression(5 - 1)
	return &BinaryExpr{
		Left:     left,
		Operator: op,
		Right:    right,
	}
}

func (par *Parser) parseIfStatement() Node {
	par.next()
	cond := par.parseExpression(5)
	slog.Debug("Parsing conditional expression", slog.Any("current", par.current()))
	if err := par.assertToken(par.current(), LBrace); err != nil {
		return nil
	}
	ifBlock := par.parseBlock()
	if par.current().kind == Else {
		par.next()
		if err := par.assertToken(par.current(), LBrace); err != nil {
			return nil
		}
		elseBlk := par.parseBlock()
		return &IfStmt{
			Condition: cond,
			IfBlock:   ifBlock,
			ElseBlock: elseBlk,
		}
	}
	return &IfStmt{
		Condition: cond,
		IfBlock:   ifBlock,
	}
}

func (par *Parser) parseBlock() Node {
	par.next()
	slog.Debug("Parsing block.", slog.String("curentToken:", par.current().kind.ToString()))
	var statements []Node
	for par.current().kind != RBrace {
		stmt := par.parseStatement()
		switch st := stmt.(type) {
		case *ReturnExpr:
			if call, ok := st.Value.(*CallExpr); ok {
				call.IsTail = true
				if call.Function.Name == par.currFunc.Name.Name {
					call.IsRecursive = true
				}
			}
		default:
			if stmt != nil {
				statements = append(statements, stmt)
			} else {
				break
			}
		}
	}
	par.next()
	return &Block{Statements: statements}
}

func (par *Parser) parseLiteral() Expr {
	val, _ := strconv.Atoi(par.current().val)
	par.next()
	return &NumLiteral{Value: val}
}

func (par *Parser) parseIdentifier() Expr {
	ident := par.current().val
	par.next()
	// If the next token is '(', it's a function call.
	if par.current().kind == LParen {
		return par.parseFunctionCall(ident)
	}
	if par.current().kind == Eq {
		// If the next token is '=', it's an assignment.
		return par.parseAssignment(&Ident{Name: ident})
	}
	return &Ident{Name: ident}
}

func (par *Parser) parseGrouping() Expr {
	par.next()
	expr := par.parseExpression(0)
	if err := par.assertToken(par.current(), RParen, "You likely forgot a closing parenthesis"); err != nil {
		return nil
	}
	par.next()
	return expr
}

func (par *Parser) parseBinary(left Expr, op tokenKind) Expr {
	precedence := precedence(op)
	par.next()
	right := par.parseExpression(precedence)
	return &BinaryExpr{
		Left:     left,
		Operator: op,
		Right:    right,
	}
}

func (par *Parser) parseInputCall() Expr {
	kind := par.current().kind
	par.next() // consume 'input'
	if err := par.assertToken(par.current(), LParen); err != nil {
		return nil
	}
	par.next()
	val := par.parseExpression(0)
	slog.Debug("Parsed inputcall, curr token: ", slog.String("token", par.current().kind.ToString()))
	par.next() // consume input
	switch kind {
	case InputStr:
		return &InputStrCall{
			Input: val,
		}
	default:
		return &InputIntCall{
			Input: val,
		}
	}
}

func (par *Parser) parseExpression(prec int) Expr {
	token := par.current()
	var left Expr
	switch token.kind {
	case Literal:
		left = par.parseLiteral()
	case Identifier:
		left = par.parseIdentifier()
	case LParen:
		left = par.parseGrouping()
	case Minus:
		operand := par.parseExpression(20)
		left = &UnaryExpr{Operator: Minus, Operand: operand}
	case String:
		left = &StringLiteral{token.val}
		par.next()
	case InputInt, InputStr:
		left = par.parseInputCall()
	default:
		panic(fmt.Sprintf("Unexpected token: %v", token.val))
	}

	for prec < precedence(par.current().kind) {
		switch par.current().kind {
		case Plus, Minus, Mul, Div, EqEq, Neq, Gt, Lt, Gte, Lte:
			op := par.current().kind
			par.next()
			right := par.parseExpression(precedence(op))
			left = &BinaryExpr{Left: left, Operator: op, Right: right}
		default:
			return left
		}
	}

	return left
}

func precedence(tk tokenKind) int {
	switch tk {
	case Plus, Minus:
		return 10
	case Mul, Div:
		return 20
	case Neq, Gt, EqEq, Lt, Gte, Lte:
		return 5
	case And, Or:
		return 3
	case If:
		return 1
	default:
		return 0
	}
}

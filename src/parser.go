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
	isRepl bool
	input  string
	tokens []Token
	pos    int
	Ast    *AST
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
		input:  lxr.src,
		tokens: lxr.tokens,
		pos:    0,
		Ast:    nil,
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
		return par.parseIdentifier()
	case Defn:
		return par.parseFunctionDef()
	case Return:
		return par.parseReturnStatement()
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
	return &CallExpr{
		Function: Ident{Name: fName},
		Args:     FuncArgs{args},
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
	par.next()
	if err := par.assertToken(par.current(), LParen); err != nil {
		return nil
	}

	par.next()
	params := []Ident{}
	for par.current().kind != RParen {
		if par.current().kind == Comma {
			par.next()
		}
		if err := par.assertToken(par.current(), Identifier); err != nil {
			return nil
		}
		params = append(params, Ident{Name: par.current().val})
		par.next()
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
	return &FuncDef{
		Name:    Ident{Name: fName},
		Params:  params,
		Body:    body,
		RetType: retType,
	}
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

func (par *Parser) parseBlock() Node {
	par.next()
	slog.Debug("Parsing block.", slog.String("curentToken:", par.current().kind.ToString()))
	var statements []Node
	for par.current().kind != RBrace {
		stmt := par.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		} else {
			break
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

func (par *Parser) parseConditionalExpr() Expr {
	slog.Debug("Parsing conditional expression", slog.Any("current", par.current()))
	cond := par.parseExpression(5)
	if err := par.assertToken(par.current(), Then); err != nil {
		return nil
	}
	par.next() // then
	thenExpr := par.parseExpression(0)
	if err := par.assertToken(par.current(), Else); err != nil {
		return nil
	}
	par.next() // else
	elseExpr := par.parseExpression(0)
	return &IfExpr{
		Condition:  cond,
		ThenBranch: thenExpr,
		ElseBranch: elseExpr,
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
	case If:
		par.next()
		left = par.parseConditionalExpr()
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
		case If:
			left = par.parseConditionalExprWithPrecedence(left)
		default:
			return left
		}
	}

	return left
}

func (par *Parser) parseConditionalExprWithPrecedence(cond Expr) Expr {
	if err := par.assertToken(par.current(), Then); err != nil {
		return nil
	}
	par.next() // then
	thenExpr := par.parseExpression(0)

	if err := par.assertToken(par.current(), Else); err != nil {
		return nil
	}
	par.next() // else
	elseExpr := par.parseExpression(0)
	return &IfExpr{
		Condition:  cond,
		ThenBranch: thenExpr,
		ElseBranch: elseExpr,
	}
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

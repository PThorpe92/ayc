package src

import "fmt"

type Analyzer struct {
	prog      *AST
	vars      map[string]Expr
	evaluated *Program
}

// lets attempt to evaluate constexpr's and store the variables in the analyzer

func NewAnalyzer(ast *AST) *Analyzer {
	return &Analyzer{
		prog:      ast,
		vars:      make(map[string]Expr),
		evaluated: &Program{},
	}
}

func (an *Analyzer) AnalyzeAndEval() *AST {
	stmts := an.prog.Root.(*Program).Statements
	for _, stmt := range stmts {
		stmt.Accept(an)
	}
	return &AST{Root: an.evaluated}
}

func (an *Analyzer) PrintOptimizedTree() {
	for _, stmt := range an.evaluated.Statements {
		stmt.Print()
	}
}

func (an *Analyzer) pushNode(node Node) {
	an.evaluated.Statements = append(an.evaluated.Statements, node)
}

func (an *Analyzer) Visit(node Node) {
	switch n := node.(type) {
	case *LetExpr:
		if IsConstExpr(n.Value) {
			cnst := an.attemptConstEval(n.Value)
			an.vars[n.Variable.Name] = cnst
			an.pushNode(&LetExpr{Variable: n.Variable, Value: cnst})
		} else {
			an.vars[n.Variable.Name] = n.Value
			an.pushNode(n)
		}
	case *CallExpr:
		if _, ok := an.vars[n.Function.Name]; !ok {
			panic(fmt.Sprintf("function %s used before declaration", n.Function.Name))
		}
		args := n.Args
		tmp := []FuncArg{}
		for _, arg := range args.Args {
			if IsConstExpr(&arg) {
				cnst := an.attemptConstEval(&arg)
				tmp = append(tmp, FuncArg{Value: cnst})
			} else {
				tmp = append(tmp, arg)
			}
		}
		an.pushNode(&CallExpr{Function: n.Function, Args: FuncArgs{Args: tmp}})
	case *FuncDef:
		params := n.Params
		for _, param := range params {
			an.vars[param.Name] = nil
		}
		for _, stmt := range n.Body.Statements {
			an.Visit(stmt)
		}
	case *Block:
		for _, stmt := range n.Statements {
			an.Visit(stmt)
		}
	case *PrintCall:
		if IsConstExpr(n.Value) {
			n.Value = an.attemptConstEval(n.Value)
		}
		an.pushNode(&PrintCall{Value: n.Value})
	case *IfExpr:
		an.pushNode(an.foldConditional(n))
	case *ReAssignExpr:
		if _, ok := an.vars[n.Variable.Name]; !ok {
			panic(fmt.Sprintf("variable %s used before declaration", n.Variable))
		}
		if IsConstExpr(n.NewValue) {
			n.NewValue = an.attemptConstEval(n.NewValue)
			an.vars[n.Variable.Name] = n.NewValue
		}
		an.pushNode(&ReAssignExpr{Variable: n.Variable, NewValue: n.NewValue})
	default:
		an.pushNode(node)
	}
}

func (an *Analyzer) attemptConstEval(expr Expr) Expr {
	switch e := expr.(type) {
	case *BinaryExpr:
		l := an.attemptConstEval(e.Left)
		r := an.attemptConstEval(e.Right)
		if l != nil && r != nil {
			return an.evalBinaryExpr(e.Operator, l, r)
		}
		return e
	case *UnaryExpr:
		operand := an.attemptConstEval(e.Operand)
		if operand != nil {
			return an.evalUnaryExpr(e.Operator, operand)
		}
		return e
	case *NumLiteral:
		return e
	case *Ident:
		if v, ok := an.vars[e.Name]; ok {
			switch ex := v.(type) {
			case *NumLiteral:
				return ex
			case *BoolLiteral:
				return ex
			case *StringLiteral:
				return ex
			}
			return e
		} else {
			panic(fmt.Sprintf("variable %s used before declaration", e.Name))
		}
	default:
		return nil
	}
}

func (an *Analyzer) evalConditionalExpr(cond *IfExpr) Expr {
	cond.Condition = an.attemptConstEval(cond.Condition)
	if cond.Condition == nil {
		return cond
	}
	if cond.Condition.(*BoolLiteral).bool {
		return cond.ThenBranch
	}
	return cond.ElseBranch
}

func (an *Analyzer) foldConditional(cond *IfExpr) Expr {
	cond.Condition = an.attemptConstEval(cond.Condition)
	if cond.Condition == nil {
		return cond
	}
	if cond.Condition.(*BoolLiteral).bool {
		return cond.ThenBranch
	}
	return cond.ElseBranch
}

func (an *Analyzer) evalBinaryExpr(op tokenKind, l, r Expr) Expr {
	lhs := an.attemptConstEval(l)
	rhs := an.attemptConstEval(r)
	if lhs == nil || rhs == nil {
		return nil
	}
	switch left := lhs.(type) {
	case *NumLiteral:
		rhs, ok := rhs.(*NumLiteral)
		if ok {
			return handleConstMath(left, rhs, op)
		}
	case *BoolLiteral:
		rhs, ok := rhs.(*BoolLiteral)
		if ok {
			return handleConstBoolLogic(left, rhs, op)
		}
	case *Ident:
		if v, ok := an.vars[left.Name]; ok {
			switch ex := v.(type) {
			case *NumLiteral:
				if rhs, ok := rhs.(*NumLiteral); ok {
					return handleConstMath(ex, rhs, op)
				}
			case *BoolLiteral:
				if rhs, ok := rhs.(*BoolLiteral); ok {
					return handleConstBoolLogic(ex, rhs, op)
				}
			case *StringLiteral:
				if rhs, ok := rhs.(*StringLiteral); ok {
					if op == Plus {
						return &StringLiteral{string: ex.string + rhs.string}
					} else {
						panic("Unknown operator for type string")
					}
				}
			case *Ident:
				return an.evalBinaryExpr(op, ex, rhs)
			default:
				return nil
			}
		} else {
			panic(fmt.Sprintf("variable %s used before declaration", left.Name))
		}
	}
	return nil
}

func handleConstBoolLogic(lhs, rhs *BoolLiteral, op tokenKind) Expr {
	switch op {
	case EqEq:
		return &BoolLiteral{lhs.bool == rhs.bool}
	case Neq:
		return &BoolLiteral{lhs.bool != rhs.bool}
	case And:
		return &BoolLiteral{lhs.bool && rhs.bool}
	case Not:
		return &BoolLiteral{lhs.bool || rhs.bool}
	}
	return nil
}

func handleConstMath(lhs, rhs *NumLiteral, op tokenKind) Expr {
	switch op {
	case Plus:
		return &NumLiteral{Value: lhs.Value + rhs.Value}
	case Minus:
		return &NumLiteral{Value: lhs.Value - rhs.Value}
	case Mul:
		return &NumLiteral{Value: lhs.Value * rhs.Value}
	case Div:
		return &NumLiteral{Value: lhs.Value / rhs.Value}
	case EqEq:
		return &BoolLiteral{lhs.Value == rhs.Value}
	case Neq:
		return &BoolLiteral{lhs.Value != rhs.Value}
	case Gt:
		return &BoolLiteral{lhs.Value > rhs.Value}
	case Gte:
		return &BoolLiteral{lhs.Value >= rhs.Value}
	case Lt:
		return &BoolLiteral{lhs.Value < rhs.Value}
	case Lte:
		return &BoolLiteral{lhs.Value <= rhs.Value}
	case LShift:
		return &NumLiteral{Value: lhs.Value << rhs.Value}
	case RShift:
		return &NumLiteral{Value: lhs.Value >> rhs.Value}
	case BitAnd:
		return &NumLiteral{Value: lhs.Value & rhs.Value}
	case BitOr:
		return &NumLiteral{Value: lhs.Value | rhs.Value}
	case BitXor:
		return &NumLiteral{Value: lhs.Value ^ rhs.Value}
	default:
		panic("Unknown operator")
	}
}

func (an *Analyzer) evalUnaryExpr(op tokenKind, operand Expr) Expr {
	operand = an.attemptConstEval(operand)
	switch operand.(type) {
	case *NumLiteral:
		val := operand.(*NumLiteral).Value
		switch op {
		case Minus:
			return &NumLiteral{Value: -val}
		case BitNot:
			return &NumLiteral{Value: ^val}
		default:
			panic("Unknown operator")
		}
	case *Ident:
		val := operand.(*Ident)
		if v, ok := an.vars[val.Name]; ok {
			switch ex := v.(type) {
			case *NumLiteral:
				switch op {
				case BitNot:
					return &NumLiteral{Value: ^ex.Value}
				case Minus:
					return &NumLiteral{Value: -ex.Value}
				default:
					// not a constexpr
					return operand
				}
			case *BoolLiteral:
				switch op {
				case Not:
					return &BoolLiteral{!ex.bool}
				default:
					// not a constexpr
					return operand
				}
			case *Ident:
				// try to resolve the identifier
				return an.evalUnaryExpr(op, ex)
			default:
				// not a constexpr
				return operand
			}
		} else {
			panic(fmt.Sprintf("variable %s used before declaration", val.Name))
		}
	}
	return operand
}

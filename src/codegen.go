package src

import (
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type BytecodeEmitter struct {
	Instructions   []Instruction
	register       int
	labelCounter   int
	varRegisterMap map[string]int
	funcMap        map[string]string // ident name to label
}

func NewBytecodeEmitter() *BytecodeEmitter {
	return &BytecodeEmitter{
		Instructions:   []Instruction{},
		register:       1,
		labelCounter:   0,
		varRegisterMap: make(map[string]int),
		funcMap:        make(map[string]string),
	}
}

func (be *BytecodeEmitter) Walk(ast *AST) {
	mainLabel := "__begin%"

	for _, stmt := range ast.Root.(*Program).Statements {
		if fn, ok := stmt.(*FuncDef); ok {
			funcLabel := "__func%_" + fn.Name.Name
			be.funcMap[fn.Name.Name] = funcLabel
		}
	}
	be.EmitLabel(mainLabel)
	for _, stmt := range ast.Root.(*Program).Statements {
		if _, ok := stmt.(*FuncDef); !ok {
			be.Visit(stmt)
		}
	}
	be.Emit(HALT, 0)
	for _, stmt := range ast.Root.(*Program).Statements {
		if fn, ok := stmt.(*FuncDef); ok {
			be.Visit(fn)
		}
	}
}

func (be *BytecodeEmitter) OutputToFile(file string) error {
	if !strings.HasSuffix(file, ".aycb") {
		file += ".aycb"
	}
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Error creating file: ", err)
	}
	defer f.Close()
	encoder := gob.NewEncoder(f)
	return encoder.Encode(be.Instructions)
}

func (be *BytecodeEmitter) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&be.Instructions)
	if err != nil {
		log.Fatal("Error decoding file: ", err)
	}
	return nil
}

func (be *BytecodeEmitter) Compile(ast *AST) []Instruction {
	be.Walk(ast)
	return be.Instructions
}

func (be *BytecodeEmitter) PrintBytecode() {
	for i, instr := range be.Instructions {
		fmt.Printf("%d: %s %v\n", i, instr.Opcode.String(), instr.Args)
	}
}

var maxRegisters = 100

type Instruction struct {
	Opcode Opcode
	Args   []interface{}
}

type Opcode int

const (
	ADD Opcode = iota
	SUB
	MUL
	DIV
	MOD
	JGT
	JLT
	JGE
	JLE
	NOT
	BAND
	BOR
	BXOR
	BNOT
	LSHIFT
	RSHIFT
	MOV
	SYSCALL
	CMP
	LOAD
	PUSH
	POP
	STORE
	FNCALL
	JNE
	RET
	JMP
	JMP_IF
	JNT
	LABEL
	MOV_IF
	NOP
	PRINT
	INPUT
	INPUTSTR
	HALT
)

func (oc Opcode) String() string {
	return opMap[oc]
}
func init() {
	gob.Register(Instruction{})
	gob.Register(Opcode(0))
	gob.Register(Register(0))
	gob.Register(LitValue{})
}

var opMap = map[Opcode]string{
	ADD:      "ADD",
	SUB:      "SUB",
	MUL:      "MUL",
	DIV:      "DIV",
	MOV_IF:   "MOV_IF",
	MOD:      "MOD",
	JMP:      "JMP",
	JMP_IF:   "JMP_IF",
	JNT:      "JNT",
	JLE:      "JLE",
	NOT:      "NOT",
	BAND:     "BAND",
	BOR:      "BOR",
	BXOR:     "BXOR",
	BNOT:     "BNOT",
	LOAD:     "LOAD",
	STORE:    "STORE",
	LSHIFT:   "LSHIFT",
	RSHIFT:   "RSHIFT",
	PUSH:     "PUSH",
	POP:      "POP",
	JNE:      "JNE",
	MOV:      "MOV",
	JGE:      "JGE",
	JGT:      "JGT",
	FNCALL:   "FCALL",
	JLT:      "JLT",
	RET:      "RET",
	SYSCALL:  "SYSCALL",
	PRINT:    "PRINT",
	INPUT:    "INPUT",
	INPUTSTR: "INPUTSTR",
	HALT:     "HALT",
	LABEL:    "LABEL",
}

var opcodeMap = map[tokenKind]Opcode{
	Plus:   ADD,
	Minus:  SUB,
	Mul:    MUL,
	Div:    DIV,
	Mod:    MOD,
	Not:    NOT,
	BitAnd: BAND,
	BitOr:  BOR,
	BitXor: BXOR,
	BitNot: BNOT,
	LShift: LSHIFT,
	RShift: RSHIFT,
	Return: RET,
	Eq:     MOV,
	Gt:     JGT,
	Lt:     JLT,
	Gte:    JGE,
	Lte:    JLE,
	EqEq:   JMP_IF,
	Neq:    JNE,
}

func (be *BytecodeEmitter) Emit(opcode Opcode, args ...interface{}) {
	be.Instructions = append(be.Instructions, Instruction{Opcode: opcode, Args: args})
}

const RAX = 0

func (be *BytecodeEmitter) AllocateRegister(name string) int {
	if reg, exists := be.varRegisterMap[name]; exists {
		return reg
	}
	if be.register >= maxRegisters {
		panic("Out of registers!")
	}
	reg := be.register
	be.varRegisterMap[name] = reg
	be.register++
	return reg
}

func (be *BytecodeEmitter) allocTemp(reg int) int {
	return be.AllocateRegister(fmt.Sprintf("temp%d", reg))
}

func (be *BytecodeEmitter) NewLabel() string {
	be.labelCounter++
	return fmt.Sprintf("0x%x", be.labelCounter)
}

func (be *BytecodeEmitter) EmitLabel(label string) {
	be.Emit(LABEL, label)
}

func (be *BytecodeEmitter) Visit(node Node) {
	switch n := node.(type) {
	case nil:
		return
	case *FuncDef:
		n.Print()
		funcLabel := "__func%_" + n.Name.Name
		be.EmitLabel(funcLabel)
		be.funcMap[n.Name.Name] = funcLabel
		slices.Reverse(n.Params)
		for _, param := range n.Params {
			reg := be.AllocateRegister(param.Name)
			be.Emit(POP, reg)
			be.varRegisterMap[param.Name] = reg
		}
		hasRet := false
		for _, stmt := range n.Body.Statements {
			switch st := stmt.(type) {
			case *ReturnExpr:
				hasRet = true
				valReg := be.CompileExpr(st.Value, false)
				be.Emit(MOV, Register(valReg), RAX)
				be.Emit(RET)
				return
			default:
				be.Visit(st)
			}
		}
		if !hasRet {
			be.Emit(LOAD, &LitValue{0}, RAX)
			be.Emit(RET)
		}
	case *CallExpr:
		_ = be.CompileExpr(n, false)
	case *IfStmt:
		elseLabel := be.NewLabel()
		endLabel := be.NewLabel()
		tmp := be.CompileExpr(n.Condition, true)
		be.Emit(JNT, tmp, elseLabel)
		for _, stmt := range n.IfBlock.(*Block).Statements {
			slog.Debug("stmt: ", slog.Any("stmt", stmt))
			be.Visit(stmt)
		}
		be.Emit(JMP, endLabel)
		be.EmitLabel(elseLabel)
		if n.ElseBlock != nil {
			for _, stmt := range n.ElseBlock.(*Block).Statements {
				be.Visit(stmt)
			}
		}
		be.EmitLabel(endLabel)
	case *ReturnExpr:
		valReg := be.CompileExpr(n.Value, false)
		be.Emit(MOV, Register(valReg), RAX)
		be.Emit(RET)
	case *LetExpr:
		valueReg := be.CompileExpr(n.Value, false)
		be.varRegisterMap[n.Variable.Name] = valueReg
	case *ReAssignExpr:
		valueReg := be.CompileExpr(n.NewValue, false)
		be.varRegisterMap[n.Variable.Name] = valueReg
		be.Emit(STORE, valueReg, n.Variable.Name)
	case *PrintCall:
		reg := be.CompileExpr(n.Value, false)
		be.Emit(SYSCALL, PRINT, reg)
	}
}

func isConditionalOp(op tokenKind) bool {
	switch op {
	case EqEq, Neq, Gt, Gte, Lt, Lte:
		return true
	default:
		return false
	}
}

func (be *BytecodeEmitter) CompileExpr(expr Expr, isConditional bool) int {
	switch e := expr.(type) {
	case nil:
		return 0
	case *NumLiteral:
		reg := be.register
		be.register++
		be.Emit(MOV, &LitValue{e.Value}, reg)
		return reg
	case *FuncArg:
		return be.CompileExpr(e.Value, false)
	case *Ident:
		if reg, found := be.varRegisterMap[e.Name]; found {
			return reg
		}
		slog.Debug("var map: ", slog.Any("map", be.varRegisterMap))
		panic(fmt.Sprintf("Variable %s not found", e.Name))
	case *StringLiteral:
		reg := be.allocTemp(be.register)
		be.Emit(MOV, &LitValue{e.string}, reg)
		return reg
	case *CallExpr:
		fnLabel, exists := be.funcMap[e.Function.Name]
		if !exists {
			panic(fmt.Sprintf("Undefined function: %s", e.Function.Name))
		}
		for _, arg := range e.Args.Args {
			argReg := be.CompileExpr(arg.Value, false)
			be.Emit(PUSH, Register(argReg))
		}
		be.Emit(FNCALL, fnLabel)
		return RAX
	case *InputIntCall:
		reg := be.CompileExpr(e.Input, false)
		tmp := be.allocTemp(be.register)
		be.Emit(SYSCALL, INPUT, reg, tmp)
		return tmp
	case *ReturnExpr:
		reg := be.CompileExpr(e.Value, false)
		be.Emit(MOV, Register(reg), RAX)
		be.Emit(RET)
		return RAX
	case *InputStrCall:
		reg := be.CompileExpr(e.Input, false)
		tmp := be.allocTemp(be.register)
		be.Emit(SYSCALL, INPUTSTR, reg, tmp)
		return tmp
	case *UnaryExpr:
		operandReg := be.CompileExpr(e.Operand, false)
		resultReg := be.allocTemp(be.register)
		be.Emit(opcodeMap[e.Operator], operandReg, resultReg)
		return resultReg
	case *BinaryExpr:
		if e.Operator == Eq {
			ident, ok := e.Left.(*Ident)
			if !ok {
				panic("Left-hand side of assignment must be an identifier")
			}
			valueReg := be.CompileExpr(e.Right, false)
			reg := be.AllocateRegister(ident.Name)
			be.Emit(MOV, Register(valueReg), reg)
			return reg
		}
		leftReg := be.CompileExpr(e.Left, isConditionalOp(e.Operator))
		rightReg := be.CompileExpr(e.Right, isConditionalOp(e.Operator))
		if !isConditionalOp(e.Operator) { // Handle arithmetic or bitwise operators
			resultReg := be.AllocateRegister(fmt.Sprintf("__temp%d", be.register))
			be.Emit(opcodeMap[e.Operator], leftReg, rightReg, resultReg)
			return resultReg
		}
		trueLabel := be.NewLabel()
		endLabel := be.NewLabel()
		switch e.Operator {
		case EqEq, Neq, Gt, Gte, Lt, Lte:
			be.Emit(opcodeMap[e.Operator], leftReg, rightReg, trueLabel)
		default:
			panic(fmt.Sprintf("Unhandled operator %v in BinaryExpr", e.Operator))
		}
		resultReg := be.allocTemp(be.register)
		be.Emit(MOV, &LitValue{0}, resultReg)
		be.Emit(JMP, endLabel)
		be.EmitLabel(trueLabel)
		be.Emit(MOV, &LitValue{Value: 1}, resultReg)
		be.EmitLabel(endLabel)
		return resultReg

	case *BoolLiteral:
		reg := be.allocTemp(be.register)
		val := 0
		if e.bool {
			val = 1
		}
		be.Emit(MOV, val, reg)

	}
	return 0
}

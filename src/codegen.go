package src

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
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
		register:       1, // 0 is reserved for the return value
		labelCounter:   0,
		varRegisterMap: make(map[string]int),
		funcMap:        make(map[string]string),
	}
}

func (be *BytecodeEmitter) Walk(ast *AST) {
	mainLabel := "__begin%"
	be.Emit(JMP, mainLabel)
	stmts := ast.Root.(*Program).Statements
	for _, stmt := range stmts {
		if fn, ok := stmt.(*FuncDef); ok {
			be.Visit(fn) // emit the labels early
		}
	}
	be.EmitLabel(mainLabel)
	for _, stmt := range stmts {
		if _, ok := stmt.(*FuncDef); !ok {
			be.Visit(stmt)
		}
	}
	be.Emit(HALT, 0)
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
	LOAD
	PUSH
	POP
	STORE
	FNCALL
	RET
	JMP
	JMP_IF
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

type TempReg struct {
	Reg int
}

func (t *TempReg) Accept(visitor Visitor) {
	visitor.Visit(t)
}
func (t *TempReg) Print() {
	fmt.Printf("TempReg: %d\n", t.Reg)
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
	MOV:      "MOV",
	JGE:      "JGE",
	JGT:      "JGT",
	FNCALL:   "FCALL",
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
	Eq:     MOV,
	Gt:     JGT,
	Lt:     JLT,
	Gte:    JGE,
	Lte:    JLE,
	Return: RET,
}

func (be *BytecodeEmitter) Emit(opcode Opcode, args ...interface{}) {
	be.Instructions = append(be.Instructions, Instruction{Opcode: opcode, Args: args})
}

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
		for i := len(n.Params) - 1; i >= 0; i-- {
			param := n.Params[i]
			reg := be.AllocateRegister(param.Name)
			be.Emit(POP, reg) // Pop into register
			be.varRegisterMap[param.Name] = reg
		}
		hasRet := false
		for _, stmt := range n.Body.Statements {
			switch st := stmt.(type) {
			case *ReturnExpr:
				hasRet = true
				valReg := be.CompileExpr(st.Value, false)
				be.Emit(PUSH, Register(valReg)) // push return value
				be.Emit(RET)
			default:
				be.Visit(st)
			}
		}
		if !hasRet {
			be.Emit(PUSH, 0)
			be.Emit(RET)
		}
	case *CallExpr:
		_ = be.CompileExpr(n, false)
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
		be.Emit(LOAD, e.Value, reg)
		return reg
	case *FuncArg:
		return be.CompileExpr(e.Value, false)
	case *Ident:
		if reg, found := be.varRegisterMap[e.Name]; found {
			return reg
		}
		panic(fmt.Sprintf("Variable %s not found", e.Name))
	case *StringLiteral:
		reg := be.allocTemp(be.register)
		be.Emit(LOAD, e.string, reg)
		return reg
	case *CallExpr:
		fnLabel := be.funcMap[e.Function.Name]
		argRegs := []int{}
		for _, arg := range e.Args.Args {
			argReg := be.CompileExpr(arg.Value, false)
			argRegs = append(argRegs, argReg)
		}
		for i := len(argRegs) - 1; i >= 0; i-- {
			be.Emit(PUSH, Register(argRegs[i]))
		}
		be.Emit(FNCALL, fnLabel)
		resultReg := be.AllocateRegister(fmt.Sprintf("__res%d", be.register))
		be.Emit(POP, resultReg)
		return resultReg
	case *InputIntCall:
		reg := be.CompileExpr(e.Input, false)
		inputReg := be.allocTemp(reg)
		be.Emit(SYSCALL, INPUT, reg, inputReg)
		return inputReg
	case *InputStrCall:
		reg := be.CompileExpr(e.Input, false)
		inputReg := be.allocTemp(reg)
		be.Emit(SYSCALL, INPUTSTR, reg, inputReg)
		return inputReg
	case *BinaryExpr:
		if e.Operator == Eq {
			ident, ok := e.Left.(*Ident)
			if !ok {
				panic("Left-hand side of assignment must be an identifier")
			}
			valueReg := be.CompileExpr(e.Right, false)
			var reg int
			if existingReg, found := be.varRegisterMap[ident.Name]; found {
				reg = existingReg
			} else {
				reg = be.AllocateRegister(ident.Name)
			}
			be.varRegisterMap[ident.Name] = reg
			be.Emit(STORE, valueReg, ident.Name)
			return reg
		}
		leftReg := be.CompileExpr(e.Left, isConditionalOp(e.Operator))
		rightReg := be.CompileExpr(e.Right, isConditionalOp(e.Operator))
		resultReg := be.AllocateRegister(fmt.Sprintf("__temp%d", be.register))
		if !isConditional {
			be.Emit(opcodeMap[e.Operator], leftReg, rightReg, resultReg) // Perform operation
			return resultReg
		}
		trueLabel := be.NewLabel()
		endLabel := be.NewLabel()
		be.Emit(opcodeMap[e.Operator], leftReg, rightReg, trueLabel)
		be.Emit(LOAD, 0, resultReg)
		be.Emit(JMP, endLabel)
		be.EmitLabel(trueLabel)
		be.Emit(LOAD, 1, resultReg)
		be.EmitLabel(endLabel)
		return resultReg

	case *BoolLiteral:
		reg := be.allocTemp(be.register)
		val := 0
		if e.bool {
			val = 1
		}
		be.Emit(MOV, val, reg)

	case *IfExpr:
		condReg := be.CompileExpr(e.Condition, true)
		thenLabel := be.NewLabel()
		endLabel := be.NewLabel()
		be.Emit(JMP_IF, condReg, thenLabel)
		elseReg := be.CompileExpr(e.ElseBranch, false)
		resultReg := be.AllocateRegister(fmt.Sprintf("__cond%d", be.register))
		be.Emit(MOV, elseReg, resultReg)
		be.Emit(JMP, endLabel)
		be.EmitLabel(thenLabel)
		thenReg := be.CompileExpr(e.ThenBranch, false)
		be.Emit(MOV, thenReg, resultReg)
		be.EmitLabel(endLabel)
		return resultReg
	}
	return 0
}

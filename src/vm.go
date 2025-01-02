package src

import (
	"fmt"
	"log/slog"
)

type GoVM struct {
	program   []Instruction
	pc        int
	Regs      map[string]int
	registers [100]interface{}
	callStack Stack[int]
	stack     Stack[interface{}]
	labels    map[string]int
	symbols   map[string]interface{}
}

func NewVM(insns []Instruction) *GoVM {
	vm := &GoVM{
		program:   insns,
		pc:        0,
		registers: [100]interface{}{},
		stack:     Stack[interface{}]{},
		callStack: Stack[int]{},
		labels:    make(map[string]int),
		symbols:   make(map[string]interface{}),
		Regs:      make(map[string]int),
	}
	vm.setLabels()
	return vm
}

type Register int

type LitValue struct {
	Value interface{}
}

func (vm *GoVM) fetchNext() Instruction {
	return vm.program[vm.pc]
}

func (vm *GoVM) pop() interface{} {
	val := vm.stack.Peek()
	slog.Debug("Popping data stack: ", slog.Any("val", val))
	return vm.stack.Pop()
}

func (vm *GoVM) push(val interface{}) {
	vm.stack.Push(val)
}

func (vm *GoVM) Exec() {
	for vm.pc < len(vm.program) {
		op := vm.fetchNext()
		switch op.Opcode {
		case MOV:
			src := op.Args[0]
			dest := op.Args[1].(int)
			switch srcVal := src.(type) {
			case int:
				// register value by default
				vm.registers[dest] = vm.registers[srcVal]
			case Register:
				vm.registers[dest] = vm.registers[int(srcVal)]
			case *LitValue:
				vm.registers[dest] = srcVal.Value
			case LitValue:
				vm.registers[dest] = srcVal.Value
			default:
				vm.registers[dest] = srcVal.(int)
			}
		case ADD:
			// ADD reg1, reg2, dest
			arg1, arg2, dest := vm.getThreeArgs(op.Args)
			slog.Debug("Adding: ", slog.Int("arg1", arg1), slog.Int("arg2", arg2), slog.Int("dest", dest))
			switch arg1 := vm.registers[arg1].(type) {
			case int:
				vm.registers[dest] = arg1 + vm.registers[arg2].(int)
			case string:
				vm.registers[dest] = arg1 + vm.registers[arg2].(string)
			case *LitValue:
				switch v := arg1.Value.(type) {
				case int:
					vm.registers[dest] = v + vm.registers[arg2].(int)
				case string:
					vm.registers[dest] = v + vm.registers[arg2].(string)
				}
			default:
				panic("Unknown type")
			}
		case SUB:
			// SUB reg1, reg2, dest
			arg1, arg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[arg1].(int) - vm.registers[arg2].(int)
		case MUL:
			// MUL reg1, reg2, dest
			arg1, arg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[arg1].(int) * vm.registers[arg2].(int)
		case DIV:
			// DIV reg1, reg2, dest
			arg1, arg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[arg1].(int) / vm.registers[arg2].(int)
		case MOD:
			// MOD reg1, reg2, dest
			arg1, arg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[arg1].(int) % vm.registers[arg2].(int)
		case FNCALL:
			// always push return value onto the stack
			label := op.Args[0].(string)
			slog.Debug("PC before fncall:", slog.Int("pc", vm.pc))
			vm.callStack.Push(vm.pc + 1)
			vm.pc = vm.labels[label]
			slog.Debug("PC after fncall: ", slog.String("label", label), slog.Int("pc", vm.pc))
			continue
		case RET:
			vm.pc = vm.callStack.Pop()
			// the return value of the function should be in RAX
			slog.Debug("Returning from function: ", slog.Int("pc", vm.pc), slog.Any("rax", vm.registers[RAX]))
			continue
		case SYSCALL:
			slog.Debug("SYSTEM CALL ")
			// CALL func args...
			fn := op.Args[0].(Opcode)
			switch fn {
			case PRINT:
				val := vm.registers[op.Args[1].(int)]
				slog.Debug("Printing value: ", slog.Any("val", val))
				switch val := val.(type) {
				case int:
					fmt.Printf("PRINT: %d\n", val)
				case string:
					fmt.Printf("PRINT: %s\n", val)
				default:
					fmt.Printf("%v", val)
				}
			case INPUT:
				prompt := vm.registers[op.Args[1].(int)].(string)
				dest := op.Args[2].(int)
				fmt.Printf("%s: ", prompt)
				var input int
				fmt.Scanln(&input)
				vm.registers[dest] = input
			case INPUTSTR:
				prompt := vm.registers[op.Args[1].(int)].(string)
				dest := op.Args[2].(int)
				fmt.Printf("%s: ", prompt)
				var input string
				fmt.Scanln(&input)
				vm.registers[dest] = input
			default:
				panic("Unknown function")
			}
		case JMP:
			// JMP label
			label := op.Args[0].(string)
			vm.pc = vm.findLabel(label)
			continue
		case JMP_IF: // JMP_IF reg1, reg2, label
			reg := op.Args[0].(int)
			reg2 := op.Args[1].(int)
			label := op.Args[2].(string)
			if vm.registers[reg].(int) == vm.registers[reg2].(int) {
				vm.pc = vm.findLabel(label)
				continue
			}
		case JNT:
			reg := op.Args[0].(int)
			label := op.Args[1].(string)
			if vm.registers[reg].(int) == 0 {
				vm.pc = vm.findLabel(label)
				continue
			}
		case JNE:
			reg1, reg2 := getTwoArgs(op.Args)
			if vm.registers[reg1].(int) != vm.registers[reg2].(int) {
				label := op.Args[2].(string)
				vm.pc = vm.findLabel(label)
				continue
			}
		case JGT:
			reg1, reg2 := getTwoArgs(op.Args)
			if vm.registers[reg1].(int) > vm.registers[reg2].(int) {
				label := op.Args[2].(string)
				vm.pc = vm.findLabel(label)
				continue
			}
		case JLT:
			arg1, arg2 := getTwoArgs(op.Args)
			reg, ok := vm.registers[arg1].(int)
			if !ok {
				reg = arg1
			}
			reg2, ok := vm.registers[arg1].(int)
			if !ok {
				reg2 = arg2
			}
			if reg < reg2 {
				label := op.Args[2].(string)
				vm.pc = vm.findLabel(label)
				continue
			}
		case JGE:
			reg1, reg2 := getTwoArgs(op.Args)
			if vm.registers[reg1].(int) >= vm.registers[reg2].(int) {
				label := op.Args[2].(string)
				vm.pc = vm.findLabel(label)
				continue
			}
		case JLE:
			reg1, reg2 := getTwoArgs(op.Args)
			if vm.registers[reg1].(int) <= vm.registers[reg2].(int) {
				label := op.Args[2].(string)
				vm.pc = vm.findLabel(label)
				continue
			}
		case LABEL:
			label := op.Args[0].(string)
			vm.labels[label] = vm.pc
		case HALT:
			return
		case NOP:
			vm.pc++
			continue
		case LSHIFT:
			reg1, reg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[reg1].(int) << vm.registers[reg2].(int)
		case RSHIFT:
			reg1, reg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[reg1].(int) >> vm.registers[reg2].(int)
		case BAND:
			reg1, reg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[reg1].(int) & vm.registers[reg2].(int)
		case BOR:
			reg1, reg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[reg1].(int) | vm.registers[reg2].(int)
		case BXOR:
			reg1, reg2, dest := vm.getThreeArgs(op.Args)
			vm.registers[dest] = vm.registers[reg1].(int) ^ vm.registers[reg2].(int)
		case BNOT:
			reg, dest := getTwoArgs(op.Args)
			vm.registers[dest] = ^vm.registers[reg].(int)
		case NOT:
			reg, dest := getTwoArgs(op.Args)
			if vm.registers[reg] == 0 {
				vm.registers[dest] = 1
			} else {
				vm.registers[dest] = 0
			}
		case MOV_IF:
			condReg, thenReg, elseReg, dest := getFourArgs(op.Args)
			if vm.registers[condReg] != 0 {
				vm.registers[dest] = vm.registers[thenReg] // Move 'then' value
			} else {
				vm.registers[dest] = vm.registers[elseReg] // Move 'else' value
			}
		case LOAD:
			// LOAD val, dest
			switch val := op.Args[0].(type) {
			case Register:
				vm.registers[op.Args[1].(int)] = vm.registers[val].(int)
			case int:
				vm.registers[op.Args[1].(int)] = vm.registers[val]
			case *LitValue:
				slog.Debug("Loading literal value: ", slog.Any("var", val.Value))
				vm.registers[op.Args[1].(int)] = val.Value
			case string:
				slog.Debug("Loading variable: ", slog.String("var", val))
				// load the string value into the register
				vm.registers[op.Args[1].(int)] = vm.symbols[val]
			}
		case STORE:
			// STORE reg, addr. store register into variable
			src := op.Args[0].(int)
			addr := op.Args[1].(string)
			vm.symbols[addr] = vm.registers[src]
			if v, ok := vm.registers[src].(string); ok {
				fmt.Printf("Stored %s in %s\n", v, addr)
			}
		case PUSH:
			switch val := op.Args[0].(type) {
			case Register:
				slog.Debug("Pushing register val onto the stack: ", slog.Any("val", vm.registers[val]))
				vm.push(vm.registers[val])
			case *LitValue:
				slog.Debug("Pushing literal val onto the stack: ", slog.Any("val", val.Value))
				vm.push(val.Value)
			case *NumLiteral:
				slog.Debug("Pushing literal val onto the stack: ", slog.Any("val", val.Value))
				vm.push(vm.registers[val.Value])
			case int:
				slog.Debug("Pushing register value onto the stack: ", slog.Any("val", vm.registers[val]))
				vm.push(vm.registers[val])
			}
		case POP:
			reg := op.Args[0].(int)
			slog.Debug("reg: ", slog.Int("reg", reg))
			vm.registers[reg] = vm.pop()
		}
		vm.pc++
	}
}

func (vm *GoVM) getThreeArgs(args []interface{}) (int, int, int) {
	arg1 := args[0].(int)
	arg2 := args[1].(int)
	switch dest := args[2].(type) {
	case int:
		return arg1, arg2, dest
	case string:
		res := vm.symbols[dest].(int)
		return arg1, arg2, res
	case Register:
		return arg1, arg2, vm.registers[dest].(int)
	default:
		return arg1, arg2, dest.(int)
	}
}

func getTwoArgs(args []interface{}) (int, int) {
	arg := args[0].(int)
	dest := args[1].(int)
	return arg, dest
}

func getFourArgs(args []interface{}) (int, int, int, int) {
	arg1 := args[0].(int)
	arg2 := args[1].(int)
	arg3 := args[2].(int)
	dest := args[3].(int)
	return arg1, arg2, arg3, dest
}

func (vm *GoVM) setLabels() {
	for i, insn := range vm.program {
		if insn.Opcode == LABEL {
			vm.labels[insn.Args[0].(string)] = i
		}
	}
}

func (vm *GoVM) findLabel(label string) int {
	return vm.labels[label]
}

# Basic Interpreter and Bytecode VM for arithmetic operations


## REGISTERS:

**%RBX:**   Special purpose Register
**$RAX:**   Special purpose Register

**%R1-%R10:** General Purpose Registers

**$SP:**    Stack Pointer
**$PC:**    Program Counter


### STACKS:
1. Program Stack:

    This stack contains the actual program instructions in the form of bytecode.
    The Program Counter ($PC) tracks the current instruction being executed by pointing to its index in this stack.

2. Data Stack:

    This stack is used to store temporary data during program execution, such as intermediate results or function call states.
    The Stack Pointer ($SP) points to the top of the data stack and is adjusted whenever data is pushed to or popped from the stack.

MEMORY MODEL:

    The virtual machine operates with a memory array to simulate a simple RAM structure.
    The general-purpose registers (%R1-%R10) are used for direct computation and referencing specific addresses or data in memory.

PROGRAM EXECUTION FLOW:

    Instruction Fetch:
        The VM fetches the instruction at the address specified by the Program Counter ($PC).

    Instruction Decode:
        The fetched instruction is decoded to determine the operation and its operands (e.g., registers, constants, or memory locations).

    Instruction Execution:
        Based on the decoded instruction, the VM executes the operation:
            Arithmetic (e.g., ADD, SUB, MULT, DIV).
            Memory access (e.g., LOAD, STORE).
            Control flow (e.g., JMP, CALL, RET).
        The result of the operation may update registers, memory, or stack contents.

    Update Program Counter:
        The $PC is incremented to point to the next instruction unless a control flow operation (e.g., JMP) modifies it.


## OPERATORS:

#### arithmetic operators

**ADD** $arg1, $arg2, $dest  = Add the value of $arg1 and $arg2 and store the result in $dest

**MULT** $arg1, $arg2, $dest = Multiply the value of $arg1 and $arg2 and store the result in $dest

**SUB** $arg1, $arg2, $dest  = Subtract the value of $arg1 and $arg2 and store the result in $dest

**DIV** $arg1, $arg2, $dest  = Divide the value of $arg1 by $arg2 and store the result in $dest

**MOD** $arg1, $arg2, $dest  = Compute the modulus of $arg1 and $arg2 and store the result in $dest


#### logical operators
**AND** $arg1, $arg2, $dest  = Compute the logical AND of $arg1 and $arg2 and store the result in $dest

**OR** $arg1, $arg2, $dest   = Compute the logical OR of $arg1 and $arg2 and store the result in $dest


#### bitwise operators

**BITAND** $arg1, $arg2, $dest = Compute the bitwise AND of $arg1 and $arg2 and store the result in $dest

**BITOR** $arg1, $arg2, $dest = Compute the bitwise OR of $arg1 and $arg2 and store the result in $dest

**BITNOT** $arg1, $dest       = Compute the bitwise NOT of $arg1 and store the result in $dest



#### comparison operators

**JMP** $arg1                 = Jump to the instruction at memory address $arg1

**JGT** $arg1, $arg2, $dest   = Jump to the instruction at memory address $arg1 if $arg1 is greater than $arg2

**JLT** $arg1, $arg2, $dest   = Jump to the instruction at memory address $arg1 if $arg1 is less than $arg2

JEQ $arg1, $arg2, $dest   = Jump to the instruction at memory address $arg1 if $arg1 is equal to $arg2

JNE $arg1, $arg2, $dest   = Jump to the instruction at memory address $arg1 if $arg1 is not equal to $arg2

JZ $arg1, $dest           = Jump to the instruction at memory address $arg1 if $dest is zero





MOV $arg1, $dest         = Move the value of $arg1 to $dest

LOAD $arg1, $dest        = Load the value at memory address $arg1 to $dest

STORE $arg1, $dest       = Store the value of $arg1 to memory address $dest

PUSH $arg1               = Push the value of $arg1 onto the stack

POP $dest                = Pop the top value from the stack and store it in $dest

NOP 					 = No operation

SYSCALL: (%RAX, %RBX) = (value of operation, calling argument)

INSTRUCTION SET (KEY OPERATIONS):

    Arithmetic Instructions:
        ADD %R1, %R2, %R3: Adds the values in %R1 and %R2 and stores the result in %R3.
        SUB %R1, %R2, %R3: Subtracts the value in %R2 from %R1 and stores the result in %R3.
        MULT %R1, %R2, %R3: Multiplies the values in %R1 and %R2 and stores the result in %R3.
        DIV %R1, %R2, %R3: Divides the value in %R1 by %R2 and stores the result in %R3.

    Memory Instructions:
        LOAD %R1, $address: Loads the value from the memory address $address into %R1.
        STORE %R1, $address: Stores the value in %R1 to the memory address $address.

    Stack Instructions:
        PUSH %R1: Pushes the value in %R1 onto the data stack.
        POP %R1: Pops the top value from the data stack into %R1.

    Control Flow Instructions:
        JMP $address: Sets $PC to $address, jumping to a different part of the program.
        JGT $addr: Sets $PC to $addr conditionally jumping, if $arg1 is GreaterThan $arg2


#### FUNCTIONS/SYSTEM CALLS:

  - PRINT: 1
  - INPUT: 2

... in the future: `open`, `write`, `seek` ? filesystem primitives

Place the functions value into %RAX, and the value to
perform the operation on into %RBX, and call "SYSCALL"

>Imagine the following code:


```

    let x = ((44 + 23) * (23 - 12)) / 4;

    let y = if (x > 100) then 1 else (0);

    print(if y > x then y else x);

    let z = 2 * (5 + x)

    print(z);

```
R1-10

#### OUTPUTS BYTECODE:

// x = (44 + 23) * 2

MOV 44, %R1           // $R1 = 44
MOV 23, %R2           // $R2 = 23
ADD %R1, %R2, $RAX    // $RA = 44 + 23
MOV 2, %R1            // $R1 = 2
MOV $RAX, %R2         // $R2 = $RAX
MULT %R1, %R2, $RAX   // $RAX = (44 + 23) * 2
PUSH $RAX             // push the value of %RAX onto the stack
XOR $RAX $RAX
// y = if (x > 100) then 1 else 0

MOV 100, %R1          // $R1 = 100
POP %R2
JGT $R2, %R1, 2       // If x > 100, jump/skip 2 lines to ***
MOV 0, $R3            // Else case: y = 0
JMP 1                 // Skip the "then" part
MOV 1, $R3            // y = 1 ***

// print(x, y)
MOV 1 %RAX           // move 'print' command into special register
POP %RBX              // pop X off the stack and place into other special register
SYSCALL               // X is printed to the screen

MOV $R3 %RBX          // RAX already contains '1' to print
SYSCALL                // Print y

// z = 2 * (5 + x)
MOV 5, %R1            // $R1 = 5
ADD %R1, %R2, %RAX    // %RAX = 5 + x
MOV 2, %R1            // $R1 = 2
MULT %R1, %RAX, %RAX  // %RAX = 2 * (5 + x)
PUSH %RAX

// print(z)
MOV 1 %RAX            // move 'PRINT' call to RAX
POP %RBX              // pop z off the stack and place it in RBX
SYSCALL               // Z is printed to the screen

HALT                  // End of program


So first step, we need to lexically analyze our input file,
and output an array of TOKEN's that are easier to parse.

1. Write a lexer, that takes in a file (you can use the input example)
and outputs an array of tokens. EXAMPLE::

TOKEN: KEYWORD: "let"  line: 1
TOKEN: IDENTIFIER: "x"  line: 1
TOKEN: SYMBOL: " = "  line: 1
TOKEN: SYMBOL: " ( "  line: 1
TOKEN: SYMBOL: " ( "  line: 1
TOKEN: LITERAL: 44    line: 1
TOKEN: SYMBOL: " + "  line: 1
TOKEN: LITERAL: " 23 "  line: 1
TOKEN: SYMBOL: " ) "  line: 1
TOKEN: SYMBOL: " * "  line: 1
TOKEN: LITERAL: " 2 "  line: 1

... you get the idea.

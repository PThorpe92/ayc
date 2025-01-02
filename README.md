# **Ayc**. A simple, buggy language to keep a promise.


>New Years' Eve 2023
I wrote down a list of things in a `.md` document that I wanted to accomplish by the end of the following year.

At that point, I had probably started maybe 3 or 4 attempts to write a compiler that each eventually ended up in the dreaded, but well populated `Side-Project Graveyard`. I remember thinking how upset I would be if I didn't make good on my #1 item, which was to actually finish a language.

So **for the last 4 days of 2024**, I decided to make good on my promise to myself. This will double as a teaching project at work, so here we have it... 3.5 days worth of writing a language, while also hacking on other things.


It's not the greatest language, and there are lots (**LOTS**) of bugs, with many, many unfinished edge-cases. But it has a `REPL`, it can output bytecode to
a binary file, which it can execute later. It makes an attempt to evaluate `constexpr`'s, and output optimized bytecode.


```
Usage of ./ayc:
  -O    Enable optimizations
  -d    Enable debug mode
  -e    Start REPL
  -i string
        Input file
  -o string
        Output bytecode file
  -r string
        Run bytecode file
```

### Syntax:

"python, with type hints, `let`, and braces"

![image](https://github.com/user-attachments/assets/03b40d4d-6198-438c-8e4e-9ea7c1be9b11)

Makes some attempt at error messages. Although they are most likely the parser's fault and not the users ;)

![image](https://github.com/user-attachments/assets/94f2be7f-83e0-4f5f-8ffb-81723406f67c)


Currently only supports `input` + `print` for stdin/out


### Recursive fizzbuzz:
(because functional reasons... not because I'm too lazy to implement `for` loops)

```python

def fizz(n: int, acc: int) -> int {
	if (acc == 0) {
		return 0
	}
  if ((n % 3) == 0) {
      print("fizz")
   }
   if ((n % 5) == 0) {
	  print("buzz")
  } else {
	print("fizzbuzz")
  }
	fizz(n + 1, acc - 1)
}

let in = input("enter a number to print fizzbuzz to: ")
fizz(in, in)
```

OUTPUTS:

```py
0: LABEL [__begin%_]
1: MOV [0xc00002d4e0 1]
2: SYSCALL [INPUT 1 2]
3: PUSH [2]
4: PUSH [2]
5: FCALL [__func%_fizz]
6: HALT [0]
7: LABEL [__func%_fizz]
8: POP [3]
9: POP [4]
10: MOV [0xc00002d5c0 5]
11: JMP_IF [3 5 0x3]
12: MOV [0xc00002d5e0 6]
13: JMP [0x4]
14: LABEL [0x3]
15: MOV [0xc00002d630 6]
16: LABEL [0x4]
17: JNT [6 0x1]
18: MOV [0xc00002d670 7]
19: MOV [7 0]
20: RET []
21: JMP [0x2]
22: LABEL [0x1]
23: LABEL [0x2]
24: MOV [0xc00002d6e0 8]
25: MOD [4 8 9]
26: MOV [0xc00002d6f0 10]
27: JMP_IF [9 10 0x7]
28: MOV [0xc00002d710 11]
29: JMP [0x8]
30: LABEL [0x7]
31: MOV [0xc00002d760 11]
32: LABEL [0x8]
33: JNT [11 0x5]
34: MOV [0xc00002d7a0 12]
35: SYSCALL [PRINT 12]
36: JMP [0x6]
37: LABEL [0x5]
38: LABEL [0x6]
39: MOV [0xc00002d820 13]
40: MOD [4 13 14]
41: MOV [0xc00002d830 15]
42: JMP_IF [14 15 0xb]
43: MOV [0xc00002d850 16]
44: JMP [0xc]
45: LABEL [0xb]
46: MOV [0xc00002d8a0 16]
47: LABEL [0xc]
48: JNT [16 0x9]
49: MOV [0xc00002d8e0 17]
50: SYSCALL [PRINT 17]
51: JMP [0xa]
52: LABEL [0x9]
53: MOV [0xc00002d940 18]
54: SYSCALL [PRINT 18]
55: LABEL [0xa]
56: MOV [0xc00002d980 19]
57: ADD [4 19 20]
58: PUSH [20]
59: MOV [0xc00002d9a0 21]
60: SUB [3 21 22]
61: PUSH [22]
62: FCALL [__func%_fizz]
63: LOAD [0xc00002d9e0 0]
64: RET []
```

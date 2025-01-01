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

"python, with type hints, and 'let'"


Currently only supports `input` + `print` for stdin/out

```python

def foo(n: int, acc: int) -> int {
	let x = 168
	if (n > 10) {
	    x = 100 
	} else {
	     x = 50
	}
	let y = acc + x
	print(x + y)
}
let var = input("enter a number to add to result of foo: ")
let next = input("enter another number: ")
let x = foo(var, next)
print(x)
```

![image](https://github.com/user-attachments/assets/03b40d4d-6198-438c-8e4e-9ea7c1be9b11)


#### OUTPUTS
```
0: JMP [__begin%]
1: LABEL [__func%_foo]
2: POP [1]
3: POP [2]
4: LOAD [168 3]
5: LOAD [10 4]
6: JGT [2 4 0x3]    ; bug :)
7: JGT [2 4 0x3]
8: LOAD [0 5]
9: JMP [0x4]
10: LABEL [0x3]
11: LOAD [1 5]
12: LABEL [0x4]
13: JMP_IF [5 0x1]
14: JMP [0x2]
15: LABEL [0x1]
16: LABEL [0x2]
17: ADD [1 3 6]
18: ADD [3 6 7]
19: SYSCALL [PRINT 7]
20: PUSH [0]
21: RET []
22: LABEL [__begin%]
23: LOAD [enter a number to add to result of foo:  8]
24: SYSCALL [INPUT 8 8]
25: LOAD [enter another number:  9]
26: SYSCALL [INPUT 9 9]
27: PUSH [9]
28: PUSH [8]
29: FCALL [__func%_foo]
30: POP [10]
31: SYSCALL [PRINT 10]
32: HALT [0]
enter a number to add to result of foo: 23
enter another number: 42

PRINT: 359
```

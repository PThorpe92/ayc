package main

import (
	"bufio"
	"flag"
	"fmt"
	"interpreter/src"
	"io"
	"log"
	"log/slog"
	"os"
)

const helpStr = `
	ayc Interpreter and Bytecode Compiler
	Usage:
	--repl  (-e)                : Start REPL
	--input (-i) [filepath.ayc] : Input file
	--debug (-d)                : Enable debug mode
	--optimize (-O)             : Enable optimizations
	--output (-o)               : Output bytecode file
	--run (-r) [filepath.aycb]  : Run bytecode file`

type Ayc struct {
	*Args
	inputBuffer string
	scanner     io.ByteScanner
}

func New() *Ayc {
	return &Ayc{parseArgs(), "", nil}
}

type Args struct {
	repl         bool
	inputFile    *string
	debug        bool
	optimize     bool
	bytecodeFile *string
	outputFile   *string
}

func main() {
	ayc := New()
	if ayc.debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	ayc.run()
}

func (a *Ayc) run() {
	if a.repl {
		a.execRepl()
	}
	if *a.bytecodeFile != "" {
		fmt.Println("Running bytecode file: ", *a.bytecodeFile)
		a.runBytecode(*a.bytecodeFile)
		return
	}
	if *a.inputFile != "" {
		if *a.outputFile != "" {
			a.compileToFile()
		} else {
			a.execInput(nil)
		}
	}
}

func (a *Ayc) runBytecode(bytecodeFile string) {
	be := src.NewBytecodeEmitter()
	be.LoadFromFile(bytecodeFile)
	if a.debug {
		be.PrintBytecode()
	}
	vm := src.NewVM(be.Instructions)
	vm.Exec()
}

func (a *Ayc) execInput(input *string) {
	var lexer *src.Lexer
	if input == nil {
		lexer = src.NewLexer(*a.inputFile)
	} else {
		lexer = src.NewInputLexer(*input)
	}
	parser := lexer.Tokenize()
	ast := parser.Parse()
	if a.optimize {
		analyzer := src.NewAnalyzer(ast)
		ast = analyzer.AnalyzeAndEval()
		if a.debug {
			analyzer.PrintOptimizedTree()
		}
	}
	be := src.NewBytecodeEmitter()
	be.Walk(ast)
	if a.debug {
		be.PrintBytecode()
	}
	vm := src.NewVM(be.Instructions)
	vm.Exec()
}

func (a *Ayc) compileToFile() {
	var lexer *src.Lexer
	lexer = src.NewLexer(*a.inputFile)
	parser := lexer.Tokenize()
	ast := parser.Parse()
	if a.optimize {
		analyzer := src.NewAnalyzer(ast)
		ast = analyzer.AnalyzeAndEval()
		if a.debug {
			analyzer.PrintOptimizedTree()
		}
	}
	be := src.NewBytecodeEmitter()
	be.Walk(ast)
	if a.debug {
		be.PrintBytecode()
	}
	err := be.OutputToFile(*a.outputFile)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
	}
}

func (a *Ayc) execRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	fmt.Println("ayc REPL. Type 'exit' to quit, type 'exec!' to run code")
	for {
		var input string
		for {
			fmt.Printf(">> ")
			scanner.Scan()
			input = scanner.Text()
			if input == "exit" {
				os.Exit(0)
			}
			if input == "help" {
				fmt.Println(helpStr)
			}
			if input == "exec!" {
				a.execInput(&a.inputBuffer)
			} else {
				a.inputBuffer += input + "\n"
			}
		}
	}
}

func parseArgs() *Args {
	inputFile := flag.String("i", "", "Input file")
	optimize := flag.Bool("O", false, "Enable optimizations")
	debug := flag.Bool("d", false, "Enable debug mode")
	bytecodeFile := flag.String("r", "", "Run bytecode file")
	outputFile := flag.String("o", "", "Output bytecode file")
	repl := flag.Bool("e", false, "Start REPL")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println(helpStr)
		log.Fatal(helpStr)
	}
	return &Args{
		inputFile:    inputFile,
		optimize:     *optimize,
		debug:        *debug,
		bytecodeFile: bytecodeFile,
		outputFile:   outputFile,
		repl:         *repl,
	}
}

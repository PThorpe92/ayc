package src

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"unicode"
)

type Lexer struct {
	src         string
	pos         int
	current     rune
	currentLine int
	tokens      []Token
}

func NewInputLexer(input string) *Lexer {
	return &Lexer{
		src:         input,
		pos:         0,
		currentLine: 1,
		current:     rune(input[0]),
		tokens:      []Token{},
	}
}

func NewLexer(filepath string) *Lexer {
	src, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}
	return &Lexer{
		src:         string(src),
		pos:         0,
		currentLine: 1,
		current:     rune(src[0]),
		tokens:      []Token{},
	}
}

func (lxr *Lexer) Tokenize() *Parser {
	for lxr.pos <= len(lxr.src) || lxr.current != '\x00' {
		token := lxr.readToken()
		if token.kind == EOF {
			break
		}
		lxr.tokens = append(lxr.tokens, token)
		lxr.next()
	}
	parser := NewParser(lxr)
	return parser
}

var keywords = map[string]tokenKind{
	"let":       Let,
	"false":     False,
	"true":      True,
	"if":        If,
	"else":      Else,
	"then":      Then,
	"print":     Print,
	"input":     InputInt,
	"input_str": InputStr,
	"and":       And,
	"or":        Or,
	"return":    Return,
	"not":       Not,
	"def":       Defn,
	"int":       Int,
	"str":       String,
	"void":      Void,
	"for":       For,
	"while":     While,
}

func (lxr *Lexer) skipComment() {
	for lxr.current != '\n' {
		lxr.next()
	}
	lxr.next()
}

func (lxr *Lexer) skipWhitespace() {
	if lxr.current == '\n' {
		lxr.currentLine++
	}
	for unicode.IsSpace(lxr.current) {
		lxr.next()
	}
}

func (lxr *Lexer) peek() rune {
	if lxr.pos+1 >= len(lxr.src) {
		return '\x00'
	} else {
		return rune(lxr.src[lxr.pos+1])
	}
}

func (lxr *Lexer) next() {
	lxr.pos++
	if lxr.pos >= len(lxr.src) {
		lxr.current = '\x00'
		return
	}
	lxr.current = rune(lxr.src[lxr.pos])
}

func (lxr *Lexer) readToken() Token {
	lxr.skipWhitespace()
	kind := fromChar(lxr.current)
	switch kind {
	case Literal:
		return lxr.readNumber()
	case Identifier, Underscore:
		return lxr.readIdent()
	case Quote:
		return lxr.readStringLit()
	default:
		// if we get an EOF token, but not at the end of the file, its an error
		if kind == EOF && lxr.pos < len(lxr.src) {
			log.Fatalf("illegal token on line: %v", lxr.currentLine)
		}
		if kind == Div && fromChar(lxr.peek()) == Div {
			slog.Debug("found comment")
			lxr.skipComment()
		}
		// if it's not any of the above, it's a single char token
		return lxr.readOp(kind)
	}
}

func (lxr *Lexer) readOp(cur tokenKind) Token {
	doubleOps := []tokenKind{Eq, Lt, Gt, BitAnd, BitOr, Bang, Minus}
	curChar := string(lxr.current)
	if slices.Contains(doubleOps, cur) && slices.Contains(doubleOps, fromChar(lxr.peek())) {
		lxr.next()
		kind := doubleOp(cur, fromChar(lxr.current))
		curChar += string(lxr.current)
		return newToken(kind, fmt.Sprintf("%s ", curChar), lxr.currentLine, lxr.pos-2)
	}
	return newToken(cur, fmt.Sprintf("%s ", curChar), lxr.currentLine, lxr.pos-1)
}

func (lxr *Lexer) readNumber() Token {
	numLit := "" // new empty string to store the number
	for unicode.IsDigit(lxr.current) {
		// while the current char is a number, store it in the string
		numLit += string(lxr.current)
		lxr.next()
	}
	lxr.tokens = append(lxr.tokens, newToken(Literal, numLit, lxr.currentLine, lxr.pos-len(numLit)))
	return lxr.readToken()
}

func (lxr *Lexer) readStringLit() Token {
	lxr.next()
	strLit := ""
	for lxr.current != '"' {
		strLit += string(lxr.current)
		lxr.next()
	}
	lxr.next()
	lxr.tokens = append(lxr.tokens, newToken(String, strLit, lxr.currentLine, lxr.pos-len(strLit)))
	return lxr.readToken()
}

func (lxr *Lexer) readIdent() Token {
	ident := "" // new empty string to store the identifier
	// at this point we don't know if it's a keyword or a variable
	for unicode.IsLetter(lxr.current) || unicode.IsDigit(lxr.current) || lxr.current == '_' {
		ident += string(lxr.current)
		lxr.next()
	}
	//check if the identifier is a keyword
	kind, ok := keywords[ident]
	if !ok {
		// if not, it's a variable so we return Identifier
		lxr.tokens = append(lxr.tokens, newToken(Identifier, ident, lxr.currentLine, lxr.pos-len(ident)))
		return lxr.readToken()
	}
	lxr.tokens = append(lxr.tokens, newToken(kind, ident, lxr.currentLine, lxr.pos-len(ident)))
	return lxr.readToken()
}

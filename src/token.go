package src

import (
	"fmt"
	"unicode"
)

type Token struct {
	kind tokenKind
	val  string
	span Span
}

type Span struct {
	line int
	pos  int
	len  int
}

func (tk *Token) Span() Span {
	return tk.span
}

const (
	Red   = "\033[31m"
	Reset = "\033[0m"
	Green = "\033[32m"
	Blue  = "\033[34m"
	Bold  = "\033[1m"
)

type tokenKind int

func newToken(kind tokenKind, val string, line int, pos int) Token {
	return Token{
		kind: kind,
		val:  val,
		span: Span{line: line, pos: pos, len: len(val)},
	}
}

func (tk *Token) Print() {
	fmt.Printf("TOKEN: %v  line: %d\n", tk.val, tk.span.line)
}

const (
	EOF tokenKind = iota

	Underscore
	Literal
	Identifier

	/* Delimiters */
	Quote
	LParen
	RParen
	LBrace
	RBrace
	Period
	Comma
	Arrow
	Dash
	/* Operators */
	Eq     // =
	EqEq   // ==
	Neq    // !=
	Minus  // - (subtraction, or unary negation)
	Plus   // + (addition)
	Mul    // * (multiplication)
	Div    // / (division)
	Mod    // % (modulus)
	Bang   // !bool (logical NOT)
	BitAnd // &   (bitwise and)
	BitOr  // |   (bitwise or)
	BitXor // ^     (bitwise xor)
	BitNot // ~   (bitwise not)
	Pipe
	Gt     // >   (greater than)
	Gte    // >   (greater than or equal to)
	Lte    // <    (less than or equal to)
	Lt     // <    (less than)
	LShift // <<  (bitshift left)
	RShift // >>  (bitshift right)

	/* Keywords */
	If
	Else
	Is
	Let
	Then
	Int
	Colon
	Semicolon
	False
	And
	For
	While
	Not
	Or
	Return
	True
	InputStr
	InputInt
	Print
	Void
	Bool
	String
	Defn
)

func (tk tokenKind) ToString() string {
	switch tk {
	case EOF:
		return "EOF"
	case Underscore:
		return "Underscore"
	case Literal:
		return "Literal"
	case Identifier:
		return "Identifier"
	case LParen:
		return "LParen"
	case RParen:
		return "RParen"
	case Period:
		return "Period"
	case Comma:
		return "Comma"
	case Eq:
		return "Eq"
	case EqEq:
		return "EqEq"
	case Neq:
		return "Neq"
	case Minus:
		return "Minus"
	case Plus:
		return "Plus"
	case Mul:
		return "Mul"
	case Div:
		return "Div"
	case Mod:
		return "Mod"
	case Bang:
		return "Bang"
	case BitAnd:
		return "BitAnd"
	case BitOr:
		return "BitOr"
	case BitXor:
		return "BitXor"
	case BitNot:
		return "BitNot"
	case Gt:
		return "Gt"
	case Gte:
		return "Gte"
	case Lte:
		return "Lte"
	case Lt:
		return "Lt"
	case LShift:
		return "LShift"
	case RShift:
		return "RShift"
	case If:
		return "If"
	case Arrow:
		return "Arrow"
	case Else:
		return "Else"
	case Let:
		return "Let"
	case Then:
		return "Then"
	case String:
		return "String"
	case Semicolon:
		return "Semicolon"
	case Colon:
		return "Colon"
	case False:
		return "False"
	case And:
		return "And"
	case Is:
		return "Is"
	case Not:
		return "Not"
	case LBrace:
		return "LBrace"
	case Return:
		return "Return"
	case RBrace:
		return "RBrace"
	case Or:
		return "Or"
	case True:
		return "True"
	case InputInt:
		return "Input"
	case Print:
		return "Print"
	case Defn:
		return "Defn"
	case Void:
		return "Void"
	case Int:
		return "Int"
	case For:
		return "For"
	case InputStr:
		return "InputStr"
	default:
		return "Unknown"
	}
}
func doubleOp(l, r tokenKind) tokenKind {
	switch l {
	case Eq:
		if r == Eq {
			return EqEq
		}
	case Gt:
		switch r {
		case Eq:
			return Gte
		case Gt:
			return RShift
		default:
			return EOF
		}
	case BitAnd:
		if r == BitAnd {
			return And
		}
	case BitOr:
		if r == BitOr {
			return Or
		}
	case Lt:
		switch r {
		case Eq:
			return Lte
		case Lt:
			return LShift
		default:
			return EOF
		}
	case Bang:
		if r == Eq {
			return Neq
		}
	case Minus:
		if r == Gt {
			return Arrow
		}
	}
	return EOF
}

func fromChar(char rune) tokenKind {
	if unicode.IsLetter(char) {
		return Identifier
	}
	if unicode.IsDigit(char) {
		return Literal
	}
	switch char {
	case '{':
		return LBrace
	case '}':
		return RBrace
	case '(':
		return LParen
	case ')':
		return RParen
	case '.':
		return Period
	case '-':
		return Minus
	case '+':
		return Plus
	case '*':
		return Mul
	case '/':
		return Div
	case '=':
		return Eq
	case ';':
		return Semicolon
	case '_':
		return Underscore
	case ',':
		return Comma
	case ':':
		return Colon
	case '!':
		return Bang
	case '%':
		return Mod
	case '<':
		return Lt
	case '>':
		return Gt
	case '^':
		return BitXor
	case '&':
		return BitAnd
	case '|':
		return BitOr
	case '~':
		return BitNot
	case '"':
		return Quote
	case '\x00':
		return EOF
	default:
		return EOF
	}
}

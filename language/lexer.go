package language

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type StateFn func(*Lexer) StateFn

type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF

	/* The “Byte Order Mark” is a special Unicode character which may appear at
	   the beginning of a file containing Unicode which programs may use to determine
	   the fact that the text stream is Unicode, what endianness the text stream is in,
	    and which of several Unicode encodings to interpret.*/
	UNICODE_BOM

	WHITE_SPACE
	LINE_TERMINATOR

	COMMENT
	COMMA

	// Punctuators
	BANG   // !
	DOLLAR // $
	LPAREN // (
	RPAREN // )
	SPREAD // ...
	LBRACK // [
	RBRACK // ]
	COLON  // :
	EQ     // =
	AT     // @
	LBRACE // {
	RBRACE // }
	PIPE   // |

	NAME   // /[_A-Za-z][_0-9A-Za-z]*/
	INT    // 12345
	FLOAT  // 123.45
	STRING // "abc"
	BOOL   // true, false
	NULL   // null

	// Keywords
	QUERY
	MUTATION
	FRAGMENT
	TYPE
	INTERFACE
	UNION
	SCALAR
	ENUM
	INPUT
	EXTEND
	IMPLEMENTS
	ON
)

func (tokenType TokenType) String() string {
	switch tokenType {
	case ILLEGAL:
		return "Illegal"
	case EOF:
		return "EndOfFile"
	case UNICODE_BOM:
		return "UnicodeBOM"
	case WHITE_SPACE:
		return "WhiteSpace"
	case LINE_TERMINATOR:
		return "LineTerminator"
	case COMMENT:
		return "Comment"
	case BANG:
		return "!"
	case DOLLAR:
		return "$"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case SPREAD:
		return "..."
	case LBRACK:
		return "["
	case RBRACK:
		return "]"
	case COLON:
		return ":"
	case EQ:
		return "="
	case AT:
		return "@"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case PIPE:
		return "|"
	case NAME:
		return "Name"
	case INT:
		return "Int"
	case FLOAT:
		return "Float"
	case STRING:
		return "String"
	case BOOL:
		return "Boolean"
	case NULL:
		return "Null"
	case QUERY:
		return "Query"
	case MUTATION:
		return "Mutation"
	case ON:
		return "On"
	case FRAGMENT:
		return "Fragment"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type  TokenType
	Val   string
	Start *Position
	End   *Position
}

func (token Token) String() string {
	switch token.Type {
	case EOF:
		return "EndOfFile"
	case ILLEGAL:
		return token.Val
	}
	if len(token.Val) > 10 {
		return fmt.Sprintf("%.10q...", token.Val)
	}
	return fmt.Sprintf("%q", token.Val)
}

type Lexer struct {
	Input        string
	Line         int
	Column       int
	Start        int
	Pos          int
	Width        int
	InitialState StateFn
	Tokens       chan Token
}

func (lexer *Lexer) Run() {
	lexer.Line = 1
	for state := lexer.InitialState; state != nil; {
		state = state(lexer)
	}
	close(lexer.Tokens)
}

func (lexer *Lexer) Emit(tokenType TokenType) {
	line := lexer.Line
	column := lexer.Column - (lexer.Pos - lexer.Start) + 1
	start := lexer.Start
	end := lexer.Pos
	startPos := &Position{
		Index:  start,
		Line:   line,
		Column: column,
	}
	endPos := &Position{
		Index:  end,
		Line:   line,
		Column: column + end - start - 1,
	}
	lexer.Tokens <- Token{tokenType, lexer.Input[lexer.Start:lexer.Pos], startPos, endPos}
	lexer.Start = lexer.Pos
	lexer.Width = 0
}

func (lexer *Lexer) Next() (rn rune) {
	if lexer.Pos >= len(lexer.Input) {
		lexer.Width = 0
		lexer.Column -= 1
		return -1
	}
	rn, lexer.Width = utf8.DecodeRuneInString(lexer.Input[lexer.Pos:])
	lexer.Column += lexer.Width
	lexer.Pos += lexer.Width
	return rn
}

func (lexer *Lexer) Ignore() {
	lexer.Start = lexer.Pos
}

func (lexer *Lexer) Backup() {
	lexer.Column -= lexer.Width
	lexer.Pos -= lexer.Width
}

func (lexer *Lexer) Peek() rune {
	rn := lexer.Next()
	lexer.Backup()
	return rn
}

func (lexer *Lexer) Accept(valid string) bool {
	if strings.IndexRune(valid, lexer.Next()) >= 0 {
		return true
	}
	lexer.Backup()
	return false
}

func (lexer *Lexer) AcceptRun(valid string) {
	for strings.IndexRune(valid, lexer.Next()) >= 0 {
	}
	lexer.Backup()
}

func (lexer *Lexer) AcceptString(valid string) bool {
	if strings.HasPrefix(lexer.Input[lexer.Pos:], valid) {
		lexer.Width = len(valid)
		lexer.Column += lexer.Width
		lexer.Pos += lexer.Width
		return true
	}
	return false
}

func (lexer *Lexer) Errorf(format string, args ...interface{}) StateFn {
	line := lexer.Line
	column := lexer.Column - (lexer.Width) + 1
	start := lexer.Start
	end := lexer.Pos
	startPos := &Position{
		Index:  start,
		Line:   line,
		Column: column,
	}
	endPos := &Position{
		Index:  end,
		Line:   line,
		Column: column + end - start - 1,
	}
	lexer.Tokens <- Token{
		ILLEGAL,
		fmt.Sprintf(format, args...),
		startPos,
		endPos,
	}
	return nil
}

func Lex(initialState StateFn, input string) chan Token {
	lexer := &Lexer{
		Input:        input,
		Tokens:       make(chan Token),
		InitialState: initialState,
	}
	go lexer.Run()
	return lexer.Tokens
}

func IsWhiteSpace(rn rune) bool {
	if rn == '\u0009' || rn == '\u0020' {
		return true
	}
	return false
}
func IsDigit(rn rune) bool {
	return (rn >= '0' && rn <= '9')
}
func IsAlphabet(rn rune) bool {
	return (rn >= 'a' && rn <= 'z') || (rn >= 'A' && rn <= 'Z')
}
func IsAllowedInNamePrefix(rn rune) bool {
	return (rn >= 'a' && rn <= 'z') || (rn >= 'A' && rn <= 'Z') || rn == '_'
}
func IsAllowedInName(rn rune) bool {
	return (rn >= 'a' && rn <= 'z') || (rn >= 'A' && rn <= 'Z') || rn == '_' || (rn >= '0' && rn <= '9')
}

func LexText(lexer *Lexer) StateFn {
	for {
		switch rn := lexer.Next(); {
		case rn == '\ufeff':
			lexer.Ignore()
		case IsWhiteSpace(rn):
			lexer.Ignore()
		case rn == '\u000A', rn == '\u000D':
			if rn == '\u000D' && lexer.Peek() == '\u000A' {
				lexer.Next()
			}
			lexer.Line += 1
			lexer.Column = 0
			lexer.Ignore()
		case rn == ',':
			lexer.Ignore()
		case rn == '!':
			lexer.Emit(BANG)
		case rn == '$':
			lexer.Emit(DOLLAR)
		case rn == '(':
			lexer.Emit(LPAREN)
		case rn == ')':
			lexer.Emit(RPAREN)
		case rn == ':':
			lexer.Emit(COLON)
		case rn == '=':
			lexer.Emit(EQ)
		case rn == '@':
			lexer.Emit(AT)
		case rn == '[':
			lexer.Emit(LBRACK)
		case rn == ']':
			lexer.Emit(RBRACK)
		case rn == '{':
			lexer.Emit(LBRACE)
		case rn == '}':
			lexer.Emit(RBRACE)
		case rn == '|':
			lexer.Emit(PIPE)
		case rn == '#':
			lexer.Backup()
			return LexComment
		case rn == '"':
			lexer.Backup()
			return LexQuote
		case rn == '.':
			lexer.Backup()
			if lexer.AcceptString("...") {
				lexer.Emit(SPREAD)
			}
		case IsDigit(rn):
			lexer.Backup()
			return LexNumber
		case IsAllowedInNamePrefix(rn):
			lexer.Backup()
			return LexName
		case rn == -1:
			lexer.Emit(EOF)
			return nil
		default:
			return lexer.Errorf("Unexpected character in input : '%v' ", rn)
		}
	}
	lexer.Emit(EOF)
	return nil
}

func LexNumber(lexer *Lexer) StateFn {
	lexer.Accept("+-")
	numberType := INT
	digits := "0123456789"
	if lexer.Accept("0") && lexer.Accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	lexer.AcceptRun(digits)
	if lexer.Accept(".") {
		lexer.AcceptRun(digits)
		numberType = FLOAT
	}
	if lexer.Accept("eE") {
		lexer.Accept("+-")
		lexer.AcceptRun("0123456789")
	}
	if IsAllowedInNamePrefix(lexer.Peek()) {
		lexer.Next()
		return lexer.Errorf("Bad number syntax: %q", lexer.Input[lexer.Start:lexer.Pos])
	}
	lexer.Emit(numberType)
	return LexText
}

func LexComment(lexer *Lexer) StateFn {
	for {
		switch rn := lexer.Next(); rn {
		case -1, '\u000A', '\u000D':
			if rn == '\u000D' && lexer.Peek() == '\u000A' {
				lexer.Next()
			}
			lexer.Line += 1
			lexer.Column = 0
			lexer.Ignore()
			return LexText
		}
	}
	return LexText
}

func LexQuote(lexer *Lexer) StateFn {
	quote := lexer.Next()
Loop:
	for {
		switch lexer.Next() {
		case '\\':
			if rn := lexer.Next(); rn != -1 && rn != '\u000A' {
				if rn == 'u' {
					len := 4
					for {
						lexer.Accept("0123456789abcdefABCDEF")
						len++
					}
					if len != 4 {
						return lexer.Errorf("Invalid unicode character")
					}
					break
				} else if rn == '"' || rn == '\\' || rn == '/' || rn == 'b' || rn == 'f' || rn == 'n' || rn == 'r' || rn == 't' {
					break
				}
				return lexer.Errorf("Invalid escape sequence")
			}
			fallthrough
		case -1, '\u000A':
			return lexer.Errorf("Unterminated quoted string")
		case quote:
			break Loop
		}
	}
	lexer.Emit(STRING)
	return LexText
}

func LexName(lexer *Lexer) StateFn {
	// Check for keywords and special symbols
	if lexer.AcceptString("true") || lexer.AcceptString("false") {
		lexer.Emit(BOOL)
		return LexText
	} else if lexer.AcceptString("null") {
		lexer.Emit(NULL)
		return LexText
	} else if lexer.AcceptString("query") {
		lexer.Emit(QUERY)
		return LexText
	} else if lexer.AcceptString("mutation") {
		lexer.Emit(MUTATION)
		return LexText
	} else if lexer.AcceptString("fragment") {
		lexer.Emit(FRAGMENT)
		return LexText
	} else if lexer.AcceptString("type") {
		lexer.Emit(TYPE)
		return LexText
	} else if lexer.AcceptString("interface") {
		lexer.Emit(INTERFACE)
		return LexText
	} else if lexer.AcceptString("union") {
		lexer.Emit(UNION)
		return LexText
	} else if lexer.AcceptString("scalar") {
		lexer.Emit(SCALAR)
		return LexText
	} else if lexer.AcceptString("enum") {
		lexer.Emit(ENUM)
		return LexText
	} else if lexer.AcceptString("input") {
		lexer.Emit(INPUT)
		return LexText
	} else if lexer.AcceptString("extend") {
		lexer.Emit(EXTEND)
		return LexText
	} else if lexer.AcceptString("implements") {
		lexer.Emit(IMPLEMENTS)
		return LexText
	}

	for IsAllowedInNamePrefix(lexer.Next()) {
	}
	lexer.Backup()
	for IsAllowedInName(lexer.Next()) {
	}
	lexer.Backup()
	lexer.Emit(NAME)
	return LexText
}

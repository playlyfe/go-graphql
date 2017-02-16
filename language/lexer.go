package language

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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
	DESCRIPTION
	DEPRECATION
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

)

func (tokenType TokenType) String() string {
	switch tokenType {
	case ILLEGAL:
		return "Illegal"
	case EOF:
		return "EOF"
	case UNICODE_BOM:
		return "UnicodeBOM"
	case WHITE_SPACE:
		return "WhiteSpace"
	case LINE_TERMINATOR:
		return "LineTerminator"
	case COMMENT:
		return "Comment"
	case DESCRIPTION:
		return "Description"
	case DEPRECATION:
		return "Deprecation"
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
		return "null"
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
		return "EOF"
	case ILLEGAL:
		return token.Val
	case NAME:
		return fmt.Sprintf("Name %q", token.Val)
	}
	if len(token.Val) > 10 {
		return fmt.Sprintf("%.10s...", token.Val)
	}
	return fmt.Sprintf("%s", token.Val)
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
	lexer.Column = 1
	for state := lexer.InitialState; state != nil; {
		state = state(lexer)
	}
	close(lexer.Tokens)
}

func (lexer *Lexer) runeToString(rn rune) string {
	var character string
	if unicode.IsControl(rn) {
		character = fmt.Sprintf("%U", rn)[2:]
		character = `\u` + character
	} else if rn == -1 {
		return "<EOF>"
	} else {
		character = string(rn)
	}
	return character
}

func (lexer *Lexer) Emit(tokenType TokenType) {
	line := lexer.Line
	column := lexer.Column - utf8.RuneCountInString(lexer.Input[lexer.Start:lexer.Pos])
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
		Column: column + end - start,
	}
	var value string
	var err error
	buffer := make([]byte, lexer.Pos-lexer.Start)
	copy(buffer, []byte(lexer.Input[lexer.Start:lexer.Pos]))
	value = string(buffer)
	if tokenType == STRING {
		value, err = strconv.Unquote(value)
		if err != nil {
			panic(err)
		}
	}
	lexer.Tokens <- Token{tokenType, value, startPos, endPos}
	lexer.Start = lexer.Pos
	lexer.Width = 0
}

func (lexer *Lexer) Next() (rn rune) {
	if lexer.Pos >= len(lexer.Input) {
		lexer.Width = 0
		return -1
	}
	rn, lexer.Width = utf8.DecodeRuneInString(lexer.Input[lexer.Pos:])
	lexer.Column += 1
	lexer.Pos += lexer.Width
	return rn
}

func (lexer *Lexer) Ignore() {
	lexer.Start = lexer.Pos
}

func (lexer *Lexer) Backup() {
	if lexer.Width > 0 {
		lexer.Column -= 1
	}
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

func (lexer *Lexer) AcceptRun(valid string) int {
	count := 0
	for strings.IndexRune(valid, lexer.Next()) >= 0 {
		count++
	}
	lexer.Backup()
	return count
}

func (lexer *Lexer) AcceptString(valid string) bool {
	if strings.HasPrefix(lexer.Input[lexer.Pos:], valid) {
		lexer.Width = len(valid)
		lexer.Column += utf8.RuneCountInString(valid)
		lexer.Pos += lexer.Width
		return true
	}
	return false
}

func (lexer *Lexer) Errorf(format string, args ...interface{}) StateFn {
	line := lexer.Line
	column := lexer.Column - utf8.RuneCountInString(lexer.Input[lexer.Start:lexer.Pos])
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
		Column: column + end - start,
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
			lexer.Ignore()
			lexer.Line += 1
			lexer.Column = 1
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
			if lexer.AcceptString("..") {
				lexer.Emit(SPREAD)
			} else {
				return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid character "%s" found in document`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
			}

		case IsDigit(rn), rn == '-':
			lexer.Backup()
			return LexNumber
		case IsAllowedInNamePrefix(rn):
			lexer.Backup()
			return LexName
		case rn == -1:
			lexer.Emit(EOF)
			return nil
		default:
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid character "%s" found in document`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
		}
	}
	lexer.Emit(EOF)
	return nil
}

func LexNumber(lexer *Lexer) StateFn {
	lexer.Accept("+-")
	numberType := INT
	digits := "0123456789"
	if lexer.Accept("0") {
		if lexer.Accept(digits) {
			lexer.Backup()
			rn := lexer.Next()
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid number, unexpected digit after 0: "%s"`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
		}
	}
	lexer.AcceptRun(digits)
	if lexer.Accept(".") {
		count := lexer.AcceptRun(digits)
		if count == 0 {
			rn := lexer.Next()
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid number, expected digit but got: "%s"`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
		}
		numberType = FLOAT
	}
	if lexer.Accept("eE") {
		lexer.Accept("+-")
		count := lexer.AcceptRun("0123456789")
		if count == 0 {
			rn := lexer.Next()
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid number, expected digit but got: "%s"`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
		}
		numberType = FLOAT
	}
	if IsAllowedInNamePrefix(lexer.Peek()) {
		rn := lexer.Next()
		return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid number, expected digit but got: "%s"`, lexer.Line, lexer.Column-1, lexer.runeToString(rn))
	}
	lexer.Emit(numberType)
	return LexText
}

func LexComment(lexer *Lexer) StateFn {
	isDescription := false
	if lexer.AcceptString("##") {
		isDescription = true
	}
	for {
		switch rn := lexer.Next(); rn {
		case -1, '\u000A', '\u000D':
			if rn == '\u000D' && lexer.Peek() == '\u000A' {
				lexer.Next()
			}
			if isDescription {
				lexer.Emit(DESCRIPTION)
			} else {
				lexer.Ignore()
			}
			lexer.Line += 1
			lexer.Column = 1
			return LexText
		}
	}
	return LexText
}

func LexQuote(lexer *Lexer) StateFn {
	quote := lexer.Next()
	index := 1
Loop:
	for {
		switch rn := lexer.Next(); rn {
		case '\\':
			if rn = lexer.Next(); rn != -1 && rn != '\u000A' {
				if rn == 'u' {
					len := 0
					for lexer.Accept("0123456789abcdefABCDEF") {
						len++
					}
					if len != 4 {
						lexer.Start = lexer.Pos - len - 2
						return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid unicode character "\u%s" found in string`, lexer.Line, lexer.Column-1-len, lexer.Input[lexer.Pos-len:lexer.Pos])
					}
					break
				} else if rn == '"' || rn == '\\' || rn == '/' || rn == 'b' || rn == 'f' || rn == 'n' || rn == 'r' || rn == 't' {
					break
				}
				lexer.Start = lexer.Pos - 2
				return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid escape sequence "%s" found in string`, lexer.Line, lexer.Column-1, lexer.Input[lexer.Pos-2:lexer.Pos])
			}
			fallthrough
		case -1:
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Closing quotation missing in string`, lexer.Line, lexer.Column)
		case '\u000A', '\u000D':
			return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Closing quotation missing in string`, lexer.Line, lexer.Column-1)
		case quote:
			break Loop
		default:
			if rn < '\u0020' && rn != '\u0009' {
				var character string
				if unicode.IsControl(rn) {
					character = fmt.Sprintf("%U", rn)[2:]
					character = `\u` + character
				} else {
					character = string(rn)
				}
				lexer.Start = lexer.Pos - 1
				return lexer.Errorf(`GraphQL Syntax Error (%d:%d) Invalid character "%s" found within string`, lexer.Line, lexer.Column-1, character)
			}
			index++
		}
	}
	lexer.Emit(STRING)
	return LexText
}

func LexName(lexer *Lexer) StateFn {
	for IsAllowedInNamePrefix(lexer.Next()) {
	}
	lexer.Backup()
	for IsAllowedInName(lexer.Next()) {
	}
	lexer.Backup()
	lexer.Emit(NAME)
	return LexText
}

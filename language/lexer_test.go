package language

import (
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

type expectedToken struct {
	Type TokenType
	Val  string
}

func verifyTokens(input string, result []Token, expectedResult []expectedToken) {
	for index, expectedToken := range expectedResult {
		actualToken := result[index]
		So(actualToken.Type, ShouldEqual, expectedToken.Type)
		So(actualToken.Val, ShouldEqual, expectedToken.Val)
		tokenText := input[actualToken.Start.Index:actualToken.End.Index]
		if actualToken.Type == STRING {
			value, err := strconv.Unquote(tokenText)
			So(err, ShouldEqual, nil)
			So(value, ShouldEqual, expectedToken.Val)
		} else {
			So(tokenText, ShouldEqual, expectedToken.Val)
		}
		lines := strings.Split(input, "\n")
		So(actualToken.Start.Line <= len(lines), ShouldEqual, true)
		So(actualToken.Start.Column <= len(lines[actualToken.Start.Line-1]), ShouldEqual, true)
		So(actualToken.End.Line <= len(lines), ShouldEqual, true)
		So(actualToken.End.Column <= len(lines[actualToken.End.Line-1])+1, ShouldEqual, true)

		// The column numbers do not correspond directly to string indexes because runes may have a byte width > 1
		// Because of this we need to count 1 rune per column to find the correct index in the input string
		actualLine := lines[actualToken.Start.Line-1]
		startCol := actualToken.Start.Column
		startIndex := 0
		for startCol > 1 {
			_, size := utf8.DecodeRuneInString(actualLine)
			startCol--
			startIndex += size
			actualLine = actualLine[size:]
		}
		actualLine = lines[actualToken.Start.Line-1]
		endCol := actualToken.End.Column
		endIndex := 0
		for endCol > 1 {
			_, size := utf8.DecodeRuneInString(actualLine)
			endCol--
			endIndex += size
			actualLine = actualLine[size:]
		}
		if actualToken.Type == STRING {
			value, err := strconv.Unquote(lines[actualToken.Start.Line-1][startIndex:endIndex])
			So(err, ShouldEqual, nil)
			So(value, ShouldEqual, expectedToken.Val)
		} else {
			So(lines[actualToken.Start.Line-1][startIndex:endIndex], ShouldEqual, expectedToken.Val)
		}
	}
}

func LexInput(initialState StateFn, input string) ([]Token, error) {
	tokens := Lex(initialState, input)
	result := []Token{}
	for {
		token := <-tokens
		if token.Type == ILLEGAL {
			return nil, &GraphQLError{
				Message: token.Val,
				Source:  input,
				Start:   token.Start,
				End:     token.End,
			}
		}
		result = append(result, token)
		if token.Type == EOF {
			break
		}
	}
	return result, nil
}

func TestLexer(t *testing.T) {

	Convey("Lexer", t, func() {
		Convey("disallows uncommon control characters", func() {
			input := "\u0007"
			result, err := LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:1) Invalid character \"\\u0007\" found in document\n\n1|\u0007\n  ^")
			So(result, ShouldEqual, nil)
		})

		Convey("accepts BOM header", func() {
			input := "\uFEFF foo"
			result, err := LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{NAME, "foo"},
			})
		})
		Convey("skips whitespace", func() {
			input := `
				foo
				`
			result, err := LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{NAME, "foo"},
			})
			input = `
				#comment
		        foo#comment`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{NAME, "foo"},
			})

			input = `,,,foo,,,`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{NAME, "foo"},
			})

		})
		// Warning !!! : If you comment this test case , go go format can mess up whitespace formatting
		Convey("errors respect whitespace", func() {
			input := `

    ?

`
			result, err := LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (3:5) Invalid character \"?\" found in document\n\n1|\n2|\n3|    ?\n      ^\n4|\n5|")
			So(result, ShouldEqual, nil)
		})

		Convey("lexes strings", func() {
			input := `"simple"`
			result, err := LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, `simple`},
			})

			input = `" white space "`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, ` white space `},
			})

			input = `"quote \""`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, `quote "`},
			})

			input = `"escaped \n\r\b\t\f"`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, "escaped \n\r\b\t\f"},
			})

			input = `"slashes \\\\ \\/"`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, `slashes \\ \/`},
			})

			input = `"unicode \u1234\u5678\u90AB\uCDEF"`
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{STRING, "unicode \u1234\u5678\u90AB\uCDEF"},
			})

		})

		Convey("reports useful string errors", func() {

			input := `"`
			result, err := LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:2) Closing quotation missing in string\n\n1|\"\n  ^")
			So(result, ShouldEqual, nil)

			input = "\"no end quote"
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:14) Closing quotation missing in string\n\n1|\"no end quote\n  ^^^^^^^^^^^^^")
			So(result, ShouldEqual, nil)

			input = "\"contains unescaped \u0007 control char\""
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:21) Invalid character \"\\u0007\" found within string\n\n1|\"contains unescaped \u0007 control char\"\n                      ^")
			So(result, ShouldEqual, nil)

			input = "\"null-byte is not \u0000 end of file\""
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:19) Invalid character \"\\u0000\" found within string\n\n1|\"null-byte is not \u0000 end of file\"\n                    ^")
			So(result, ShouldEqual, nil)

			input = "\"multi\nline\""
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Closing quotation missing in string\n\n1|\"multi\n  ^^^^^^^\n2|line\"")
			So(result, ShouldEqual, nil)

			// This test will fail on linux, osx and windows systems
			/*
				            input = "\"multi\rline\""
							result, err = LexInput(LexText, input)
							So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Closing quotation missing in string\n\n1|\"multi\r  ^^^^^^^\n2|line\"")
							So(result, ShouldEqual, nil)
			*/
			input = `"bad \z esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid escape sequence \"\\z\" found in string\n\n1|\"bad \\z esc\"\n       ^^")
			So(result, ShouldEqual, nil)

			input = `"bad \x esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid escape sequence \"\\x\" found in string\n\n1|\"bad \\x esc\"\n       ^^")
			So(result, ShouldEqual, nil)

			input = `"bad \u1 esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid unicode character \"\\u1\" found in string\n\n1|\"bad \\u1 esc\"\n       ^^^")
			So(result, ShouldEqual, nil)

			input = `"bad \u0XX1 esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid unicode character \"\\u0\" found in string\n\n1|\"bad \\u0XX1 esc\"\n       ^^^")
			So(result, ShouldEqual, nil)

			input = `"bad \uXXXX esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid unicode character \"\\u\" found in string\n\n1|\"bad \\uXXXX esc\"\n       ^^")
			So(result, ShouldEqual, nil)

			input = `"bad \uFXXX esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid unicode character \"\\uF\" found in string\n\n1|\"bad \\uFXXX esc\"\n       ^^^")
			So(result, ShouldEqual, nil)

			input = `"bad \uXXXF esc"`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:7) Invalid unicode character \"\\u\" found in string\n\n1|\"bad \\uXXXF esc\"\n       ^^")
			So(result, ShouldEqual, nil)
		})

		Convey("lexes numbers", func() {
			input := "4"
			result, err := LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{INT, "4"},
			})

			input = "4.123"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "4.123"},
			})

			input = "-4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{INT, "-4"},
			})

			input = "9"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{INT, "9"},
			})

			input = "0"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{INT, "0"},
			})

			input = "-4.123"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-4.123"},
			})

			input = "0.123"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "0.123"},
			})

			input = "123e4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "123e4"},
			})

			input = "123E4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "123E4"},
			})

			input = "123e-4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "123e-4"},
			})

			input = "123e+4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "123e+4"},
			})

			input = "-1.123e4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-1.123e4"},
			})

			input = "-1.123E4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-1.123E4"},
			})

			input = "-1.123e-4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-1.123e-4"},
			})

			input = "-1.123e+4"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-1.123e+4"},
			})

			input = "-1.123e4567"
			result, err = LexInput(LexText, input)
			So(err, ShouldEqual, nil)
			verifyTokens(input, result, []expectedToken{
				{FLOAT, "-1.123e4567"},
			})

		})

		Convey("lex reports useful number errors", func() {
			input := `00`
			result, err := LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:2) Invalid number, unexpected digit after 0: \"0\"\n\n1|00\n  ^^")
			So(result, ShouldEqual, nil)

			input = `+1`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:1) Invalid character \"+\" found in document\n\n1|+1\n  ^")
			So(result, ShouldEqual, nil)

			input = `1.`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:2) Invalid number, expected digit but got: \"<EOF>\"\n\n1|1.\n  ^^")
			So(result, ShouldEqual, nil)

			input = `.123`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:1) Invalid character \".\" found in document\n\n1|.123\n  ^")
			So(result, ShouldEqual, nil)

			input = `1.A`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:3) Invalid number, expected digit but got: \"A\"\n\n1|1.A\n  ^^^")
			So(result, ShouldEqual, nil)

			input = `-A`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:2) Invalid number, expected digit but got: \"A\"\n\n1|-A\n  ^^")
			So(result, ShouldEqual, nil)

			input = `1.0e`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:4) Invalid number, expected digit but got: \"<EOF>\"\n\n1|1.0e\n  ^^^^")
			So(result, ShouldEqual, nil)

			input = `1.0eA`
			result, err = LexInput(LexText, input)
			So(err.Error(), ShouldEqual, "GraphQL Syntax Error (1:5) Invalid number, expected digit but got: \"A\"\n\n1|1.0eA\n  ^^^^^")
			So(result, ShouldEqual, nil)
		})
	})

}

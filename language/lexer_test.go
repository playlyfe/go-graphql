package language

import (
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
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
		So(tokenText, ShouldEqual, expectedToken.Val)
		lines := strings.Split(input, "\n")
		So(actualToken.Start.Line <= len(lines), ShouldEqual, true)
		So(actualToken.Start.Column <= len(lines[actualToken.Start.Line-1]), ShouldEqual, true)
		So(actualToken.End.Line <= len(lines), ShouldEqual, true)
		So(actualToken.End.Column <= len(lines[actualToken.End.Line-1]), ShouldEqual, true)
		So(lines[actualToken.Start.Line-1][actualToken.Start.Column-1:actualToken.End.Column], ShouldEqual, expectedToken.Val)
	}
}

func LexInput(initialState StateFn, input string) []Token {
	tokens := Lex(initialState, input)
	result := []Token{}
	for {
		token := <-tokens
		result = append(result, token)
		if token.Type == EOF {
			break
		}
	}
	return result
}

func TestLexer(t *testing.T) {

	Convey("Lexer", t, func() {

		Convey("should lex simple expressions", func() {
			input := `
            # Fetch the user
            {
              user(id: 4) {
                name# Get the users name
                ...F1
              }
            }
            `
			result := LexInput(LexText, input)
			verifyTokens(input, result, []expectedToken{
				{LBRACE, "{"},
				{NAME, "user"},
				{LPAREN, "("},
				{NAME, "id"},
				{COLON, ":"},
				{INT, "4"},
				{RPAREN, ")"},
				{LBRACE, "{"},
				{NAME, "name"},
				{SPREAD, "..."},
				{NAME, "F1"},
				{RBRACE, "}"},
				{RBRACE, "}"},
				{EOF, ""},
			})
		})

		Convey("should be able to lex complex expresions", func() {
		})

	})

}

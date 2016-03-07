package language

import (
	"fmt"
	"strconv"
	"strings"
)

type GraphQLError struct {
	Message string
	Field   *Field
	Source  string
	Start   *Position
	End     *Position
}

func (err *GraphQLError) Error() string {
	if err.Source != "" {
		lines := []string{err.Message, ""}
		loc := strings.Split(err.Source, "\n")
		startLine := err.Start.Line - 2
		endLine := err.Start.Line + 2
		if startLine < 1 {
			startLine = 1
		}
		if endLine > len(loc) {
			endLine = len(loc)
		}
		startLineNumber := strconv.FormatInt(int64(startLine), 10)
		endLineNumber := strconv.FormatInt(int64(endLine), 10)
		numberColumnWidth := len(startLineNumber)
		if len(endLineNumber) > len(startLineNumber) {
			numberColumnWidth = len(endLineNumber)
		}
		index := startLine
		for index <= endLine {
			lines = append(lines, fmt.Sprintf("%"+strconv.FormatInt(int64(numberColumnWidth), 10)+"d|%s", index, loc[index-1]))
			index++
			if index-1 >= err.Start.Line && index-1 <= err.End.Line {
				highlight := ""
				for index := 1; index < err.Start.Column+numberColumnWidth+1; index++ {
					highlight += " "
				}
				// This is an edge case, some tokens have 0 width
				if err.Start.Column == err.End.Column {
					highlight += "^"
				}
				for index := err.Start.Column; index < err.End.Column; index++ {
					highlight += "^"
				}
				lines = append(lines, highlight)
			}
		}
		return strings.Join(lines, "\n")
	} else {
		return err.Message
	}
}

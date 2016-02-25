package language

import (
	//"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParser(t *testing.T) {

	Convey("Parser", t, func() {

		Convey("should parse simple expressions", func() {
			input := `
            # Fetch the user
            {
              user(id: 4) {
                name# Get the users name
              }
            }
            `
			parser := &Parser{}
			_, err := parser.Parse(input)
			So(err, ShouldEqual, nil)
			/*output, err := json.MarshalIndent(result, "\t", "  ")
			So(err, ShouldEqual, nil)
			println(string(output))*/
		})

		Convey("should be able to parse complex expresions", func() {
		})

		Convey("should parse simple type definitions", func() {
			input := `
            type user {
                name: String
                password: String!
            }
            `
			parser := &Parser{}
			_, err := parser.Parse(input)
			So(err, ShouldEqual, nil)
			/*output, err := json.MarshalIndent(result, "\t", "  ")
			So(err, ShouldEqual, nil)
			println(string(output))*/
		})

	})

}

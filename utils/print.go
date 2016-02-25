package utils

import (
	"encoding/json"
)

func PrintJSON(value interface{}) {
	output, err := json.MarshalIndent(value, "  ", "  ")
	if err != nil {
		panic(err)
	}
	println(string(output))
}

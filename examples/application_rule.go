package main

import (
	"encoding/json"
	"fmt"
	"github.com/oarkflow/filters"
	"os"
)

func main() {
	dosData, err := os.ReadFile("data.json")
	if err != nil {
		panic(err)
	}
	var data map[string]any
	err = json.Unmarshal(dosData, &data)
	if err != nil {
		panic(err)
	}
	content, err := os.ReadFile("sample.json")
	if err != nil {
		panic(err)
	}
	var applicationRule *filters.ApplicationRule
	err = json.Unmarshal(content, &applicationRule)
	if err != nil {
		panic(err)
	}
	applicationRule.BuildRuleFromRequest(nil)
	fmt.Println(applicationRule.Rule.Validate(data))
}

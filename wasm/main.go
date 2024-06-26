//go:build JS
// +build JS

package main

import (
	_ "crypto/sha512"
	"encoding/json"
	"syscall/js"

	"github.com/oarkflow/filters"
)

func main() {
	done := make(chan struct{}, 0)
	js.Global().Set("parseSQL", js.FuncOf(parseSQL))
	js.Global().Set("match", js.FuncOf(match))
	<-done
}

func parseSQL(this js.Value, args []js.Value) interface{} {
	condition, err := filters.ParseSQL(args[0].String())
	if err != nil {
		return err.Error()
	}
	bt, _ := json.Marshal(condition)
	return string(bt)
}

func match(this js.Value, args []js.Value) interface{} {
	data := args[1].String()
	condition, err := filters.ParseSQL(args[0].String())
	if err != nil {
		return err.Error()
	}
	return condition.Match()
	bt, _ := json.Marshal(condition)
	return string(bt)
}

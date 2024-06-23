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

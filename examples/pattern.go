package main

import (
	"fmt"

	"github.com/oarkflow/filters/pattern"
)

func main() {
	a := 3
	b := 15
	result, err := pattern.
		Match(a, b).
		Case(func(args ...any) (any, error) {
			fmt.Println(args)
			return 5, nil
		}, 3, 15).
		Default(func(args ...any) (any, error) {
			fmt.Println(args)
			return 2, nil
		}).
		Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(result, err)
}

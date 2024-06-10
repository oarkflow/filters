package main

import (
	"fmt"

	"github.com/oarkflow/filters/pattern"
)

func main() {
	b := 15
	result, err := pattern.
		Match[int](nil, b).
		Case(func(args ...any) (int, error) {
			fmt.Println(args)
			return 5, nil
		}, pattern.NONE, pattern.ANY).
		Default(func(args ...any) (int, error) {
			fmt.Println(args)
			return 2, nil
		}).
		Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(result, err)
}

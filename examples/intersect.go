package main

import (
	"fmt"
	convert "github.com/oarkflow/convert/v2"
	"github.com/oarkflow/expr"
	"github.com/oarkflow/filters"
	"slices"
)

// Intersection - Original Intersection function
func Intersection(p ...[]any) []any {
	var rs []any
	pLen := len(p)
	if pLen == 0 {
		return rs
	}
	if pLen == 1 {
		return p[0]
	}
	first := p[0]
	rest := p[1:]
	rLen := len(rest)
	for _, f := range first {
		j := 0
		for _, rs := range rest {
			if !slices.Contains(rs, f) {
				break
			}
			j++
		}
		if j == rLen {
			rs = append(rs, f)
		}
	}
	return rs
}

func main() {
	expr.AddFunction("intersect", func(params ...any) (any, error) {
		if len(params) < 2 {
			return nil, fmt.Errorf("invalid number of arguments")
		}
		var args [][]any
		for _, param := range params {
			p, ok := convert.ToSlice[any](param)
			if !ok {
				return nil, fmt.Errorf("unable to convert to slice")
			}
			args = append(args, p)
		}
		return Intersection(args...), nil
	})
	data := map[string]any{
		"cpts": []any{"1221", "21221"},
		"dxs":  []any{"rx120", "1221"},
	}
	f := filters.NewFilter("{{len(intersect(cpts, dxs))}}", filters.Equal, 1)
	fmt.Println(f.Match(data))
}

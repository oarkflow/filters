package main

import (
	"fmt"
	"github.com/oarkflow/date"
	"github.com/oarkflow/expr"
	"github.com/oarkflow/filters"
	"github.com/oarkflow/filters/utils"
)

var data1 = map[string]any{
	"patient": map[string]any{
		"first_name": "John",
		"gender":     "male",
		"salary":     25000,
		"dob":        "1989-04-19",
	},
	"cpt": map[string]any{
		"code": "code1",
	},
}

func builtinAge(params ...any) (any, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("age expects 1 argument")
	}
	left := params[0]
	t, err := utils.ParseTime(left)
	if err != nil {
		return nil, err
	}
	return date.CalculateToNow(t), err
}

func main() {
	expr.AddFunction("age", builtinAge)

	lookup := &filters.Lookup{
		Data: []map[string]any{
			{
				"title":   "Min Salary",
				"salary":  12000,
				"min_age": 18,
				"max_age": 20,
			},
			{
				"title":   "Avg Salary",
				"salary":  13000,
				"min_age": 21,
				"max_age": 30,
			},
			{
				"title":   "Max Salary",
				"salary":  25000,
				"min_age": 31,
				"max_age": 40,
			},
		},
		Condition: "map(filter(lookup, age(data.patient.dob) >= .min_age && age(data.patient.dob) <= .max_age), .salary)",
	}
	filter := filters.NewFilter("patient.salary", filters.GreaterThanEqualCount, 1)
	filter.SetLookup(lookup)
	fmt.Println(filters.Match(data1, filter))
}

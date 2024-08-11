package main

import (
	"fmt"
	"strings"

	"github.com/oarkflow/expr"
)

type WorkItemClientRef struct {
	ClientRef  *string
	WorkItemID int
}

type Data struct {
	WorkItemClientRefs []WorkItemClientRef
}

func main() {
	expr.AddFunction("isEmpty", func(params ...any) (any, error) {
		if len(params) != 1 {
			return false, fmt.Errorf("isEmpty expects 1 argument")
		}
		arg := params[0]
		switch arg := arg.(type) {
		case nil:
			return true, nil
		case string:
			return arg == strings.TrimSpace(""), nil
		}
		return false, nil
	})
	clientRef1 := "ref1"
	clientRef2 := "ref2"
	data := Data{
		WorkItemClientRefs: []WorkItemClientRef{
			{ClientRef: &clientRef1, WorkItemID: 1},
			{ClientRef: nil, WorkItemID: 2},
			{ClientRef: &clientRef2, WorkItemID: 3},
			{ClientRef: nil, WorkItemID: 4},
		},
	}
	workItemID := 4
	env := map[string]interface{}{
		"work_item_id": workItemID,
		"data":         data,
	}
	// result, err := expr.Eval("work_item_id in map(filter(data.WorkItemClientRefs, .ClientRef == null || .ClientRef == ''), .WorkItemID)", env)
	result, err := expr.Eval("work_item_id in map(filter(data.WorkItemClientRefs, .ClientRef == null || .ClientRef == ''), .WorkItemID)", env)
	if err != nil {
		fmt.Printf("Error running expression: %v\n", err)
		return
	}
	fmt.Printf("Is work_item_id %d in the list? %v\n", workItemID, result)
}

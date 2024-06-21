package main

import (
	"fmt"
	"reflect"

	"github.com/oarkflow/filters"
)

func main() {
	sqlWhere := "price > 100 AND (item = 'laptop' OR discount > 50) AND deleted_at IS NULL AND is_active IS NOT NULL AND is_active = true"

	filterGroup, err := filters.FromSQL(sqlWhere)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, f := range filterGroup.Filters {
		fmt.Println(reflect.TypeOf(f.Value))
		fmt.Println("Field", f.Field, "Operator", f.Operator, "Value", f.Value)
	}
	fmt.Printf("Filter Group: %+v\n", filterGroup)
}

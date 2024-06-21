package main

import (
	"fmt"

	"github.com/oarkflow/filters"
)

func main() {
	sqlWhere := "(price > 100 AND (item = 'laptop' OR discount > 50) AND deleted_at IS NULL AND is_active IS NOT NULL AND is_active = true) OR (Salary>10 AND LoggedInAt BETWEEN {{CreatedAt}} AND {{VerifiedAt}} AND (LastName LIKE John% || FirstName LIKE Doe%))"

	filterGroup, err := filters.FromSQL(sqlWhere)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(filterGroup.Operator)
}

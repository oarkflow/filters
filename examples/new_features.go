package main

import (
	"fmt"

	"github.com/oarkflow/filters"
)

func main() {
	data := []map[string]any{
		{"name": "John Doe", "age": 30, "city": "New York", "active": true, "score": 0},
		{"name": "Jane Smith", "age": 25, "city": "Los Angeles", "active": false, "score": 85},
		{"name": "Bob Johnson", "age": 35, "city": nil, "active": true, "score": 92},
		{"name": "john doe", "age": 28, "city": "Chicago", "active": true, "score": 78},
	}

	fmt.Println("=== Null/Zero Checks ===")
	// IsNull operator
	nullFilter := filters.NewFilter("city", filters.IsNull, nil)
	filtered := filters.ApplyGroup(data, &filters.FilterGroup{Operator: filters.AND, Filters: []filters.Condition{nullFilter}})
	fmt.Println("Records with null city:", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}

	// IsZero operator
	zeroFilter := filters.NewFilter("score", filters.IsZero, nil)
	filtered = filters.ApplyGroup(data, &filters.FilterGroup{Operator: filters.AND, Filters: []filters.Condition{zeroFilter}})
	fmt.Println("Records with zero score:", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}

	fmt.Println("\n=== NotIn Operator ===")
	notInFilter := filters.NewFilter("age", filters.NotIn, []any{25, 35})
	filtered = filters.ApplyGroup(data, &filters.FilterGroup{Operator: filters.AND, Filters: []filters.Condition{notInFilter}})
	fmt.Println("Records with age NOT in [25, 35]:", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}

	fmt.Println("\n=== Case-Sensitive String Operations ===")
	// Case-sensitive contains
	csFilter := filters.NewFilter("name", filters.ContainsCS, "John")
	filtered = filters.ApplyGroup(data, &filters.FilterGroup{Operator: filters.AND, Filters: []filters.Condition{csFilter}})
	fmt.Println("Records with name containing 'John' (case-sensitive):", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}

	// Case-insensitive contains (existing)
	ciFilter := filters.NewFilter("name", filters.Contains, "John")
	filtered = filters.ApplyGroup(data, &filters.FilterGroup{Operator: filters.AND, Filters: []filters.Condition{ciFilter}})
	fmt.Println("Records with name containing 'John' (case-insensitive):", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}

	fmt.Println("\n=== Query String Parsing with Null Operators ===")
	// Parse query with null check
	parsedFilters, err := filters.ParseQuery("?city=isnull&active=eq:true")
	if err != nil {
		fmt.Println("Error parsing query:", err)
		return
	}

	// Convert []*Filter to []Condition
	var conditions []filters.Condition
	for _, f := range parsedFilters {
		conditions = append(conditions, f)
	}

	group := &filters.FilterGroup{
		Operator: filters.AND,
		Filters:  conditions,
	}
	filtered = filters.ApplyGroup(data, group)
	fmt.Println("Records with null city AND active=true:", len(filtered))
	for _, item := range filtered {
		fmt.Printf("  %+v\n", item)
	}
}

package main

import (
	"fmt"
	"time"

	"github.com/oarkflow/filters"
)

func main() {
	filter, err := filters.ParseQuery(`?Name:starts_with:Jane&CreatedAt:between:2022-06-01,2023-01-01`)
	if err != nil {
		panic(err)
	}
	// Sample data (map type)
	mapData := []map[string]any{
		{"Age": 25, "City": "New York", "CreatedAt": "2023-01-01 12:00:00", "Name": "John Doe"},
		{"Age": 30, "City": "Los Angeles", "CreatedAt": "2022-06-15 15:30:00", "Name": "Jane Doe"},
		{"Age": 35, "City": "Chicago", "CreatedAt": "2021-12-25 08:45:00", "Name": "Alice Smith"},
		{"Age": 40, "City": "Houston", "CreatedAt": "2022-11-11 20:15:00", "Name": "Bob Johnson"},
	}
	group2 := filters.FilterGroup{
		Operator: filters.AND,
		Filters:  filter,
	}
	fmt.Println(filter)
	// Apply filters to map data
	filteredMapData, err := filters.ApplyGroup(mapData, []filters.FilterGroup{group2})
	if err != nil {
		fmt.Println("Error applying filters:", err)
	}

	// Print filtered map data
	fmt.Println("Filtered Map Data")
	for _, item := range filteredMapData {
		fmt.Println(item)
	}

}

func structData() {
	mapData := []map[string]any{
		{"Age": 25, "City": "New York", "CreatedAt": "2023-01-01 12:00:00", "Name": "John Doe"},
		{"Age": 30, "City": "Los Angeles", "CreatedAt": "2022-06-15 15:30:00", "Name": "Jane Doe"},
		{"Age": 35, "City": "Chicago", "CreatedAt": "2021-12-25 08:45:00", "Name": "Alice Smith"},
		{"Age": 40, "City": "Houston", "CreatedAt": "2022-11-11 20:15:00", "Name": "Bob Johnson"},
	}
	// Sample data (struct type)
	type Person struct {
		Age       int       `json:"age"`
		City      string    `json:"city"`
		CreatedAt time.Time `json:"created_at"`
		Name      string    `json:"name"`
	}

	structData := []Person{
		{Age: 25, City: "New York", CreatedAt: time.Date(2023, 01, 01, 12, 0, 0, 0, time.UTC), Name: "John Doe"},
		{Age: 30, City: "Los Angeles", CreatedAt: time.Date(2022, 06, 15, 15, 30, 0, 0, time.UTC), Name: "Jane Doe"},
		{Age: 35, City: "Chicago", CreatedAt: time.Date(2021, 12, 25, 8, 45, 0, 0, time.UTC), Name: "Alice Smith"},
		{Age: 40, City: "Houston", CreatedAt: time.Date(2022, 11, 11, 20, 15, 0, 0, time.UTC), Name: "Bob Johnson"},
	}
	group2 := filters.FilterGroup{
		Operator: filters.AND,
		Filters: []filters.Filter{
			{Field: "CreatedAt", Operator: filters.BETWEEN, Value: []any{"2022-06-01", "2023-01-01"}},
			{Field: "Name", Operator: filters.STARTS_WITH, Value: "Jane"},
		},
	}
	// Apply filters to struct data
	filteredStructData, err := filters.ApplyGroup(structData, []filters.FilterGroup{group2})
	if err != nil {
		fmt.Println("Error applying filters:", err)
		return
	}

	fmt.Println("Filtered Struct Data")
	// Print filtered struct data
	for _, item := range filteredStructData {
		fmt.Println(item)
	}

	group1 := filters.FilterGroup{
		Operator: filters.AND,
		Filters: []filters.Filter{
			{Field: "Age", Operator: filters.GREATER_THAN, Value: 27},
			{Field: "City", Operator: filters.CONTAINS, Value: "Hous"},
		},
	}
	// Create a binary expression
	binaryExpr := filters.BinaryExpr[map[string]any]{
		Left:     &group1,
		Operator: filters.AND,
		Right:    &group2,
	}

	// Apply filters to map data using binary expression
	filteredMapData, err := filters.ApplyBinaryFilter(mapData, binaryExpr)
	if err != nil {
		fmt.Println("Error applying filters:", err)
		return
	}

	// Print filtered map data
	for _, item := range filteredMapData {
		fmt.Println(item)
	}
}

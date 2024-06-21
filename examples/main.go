package main

import (
	"fmt"
	"time"

	"github.com/oarkflow/filters"
)

func main() {
	filter, err := filters.ParseQuery(`?Name:starts_with:Jane&LoggedInAt:between:{{CreatedAt}},{{VerifiedAt}}`)
	if err != nil {
		panic(err)
	}
	// Sample data (map type)
	mapData := []map[string]any{
		{"Age": 25, "City": "NewFilter York", "CreatedAt": "2023-01-01 12:00:00", "VerifiedAt": "2023-01-01 12:00:00", "LoggedInAt": "2023-01-01 12:00:00", "Name": "Sujit Doe"},
		{"Age": 30, "City": "Los Angeles", "CreatedAt": "2022-06-15 15:30:00", "VerifiedAt": "2022-06-15 15:45:00", "LoggedInAt": "2022-06-15 15:33:00", "Name": "Jane Doe"},
		{"Age": 35, "City": "Chicago", "CreatedAt": "2021-12-25 08:45:00", "VerifiedAt": "2021-12-25 08:45:00", "LoggedInAt": "2021-12-25 08:45:00", "Name": "Alice Smith"},
		{"Age": 40, "City": "Houston", "CreatedAt": "2022-11-11 20:15:00", "VerifiedAt": "2022-11-11 20:15:00", "LoggedInAt": "2022-11-11 20:15:00", "Name": "Bob Johnson"},
	}
	group2 := &filters.FilterGroup{
		Operator: filters.AND,
		Filters:  filter,
	}
	// Apply filters to map data
	filteredMapData := filters.ApplyGroup(mapData, group2)
	// Print filtered map data
	fmt.Println("Filtered Map Data")
	for _, item := range filteredMapData {
		fmt.Println(item)
	}
	structData()
}

func structData() {
	// Sample data (struct type)
	type Person struct {
		Age        int       `json:"age"`
		City       string    `json:"city"`
		CreatedAt  time.Time `json:"created_at"`
		VerifiedAt time.Time `json:"verified_at"`
		LoggedInAt time.Time `json:"logged_in_at"`
		Name       string    `json:"name"`
	}

	structData := []Person{
		{Age: 25, City: "NewFilter York", CreatedAt: time.Date(2023, 01, 01, 12, 0, 0, 0, time.UTC), Name: "Sujit Baniya"},
		{Age: 30, City: "Los Angeles", LoggedInAt: time.Date(2022, 06, 15, 15, 33, 0, 0, time.UTC), VerifiedAt: time.Date(2022, 06, 15, 15, 45, 0, 0, time.UTC), CreatedAt: time.Date(2022, 06, 15, 15, 30, 0, 0, time.UTC), Name: "Jane Doe"},
		{Age: 35, City: "Chicago", CreatedAt: time.Date(2021, 12, 25, 8, 45, 0, 0, time.UTC), Name: "Alice Smith"},
		{Age: 40, City: "Houston", CreatedAt: time.Date(2022, 11, 11, 20, 15, 0, 0, time.UTC), Name: "Bob Johnson"},
	}
	group2 := &filters.FilterGroup{
		Operator: filters.AND,
		Filters: []*filters.Filter{
			{Field: "LoggedInAt", Operator: filters.Between, Value: []any{"{{CreatedAt}}", "{{VerifiedAt}}"}},
			{Field: "Name", Operator: filters.StartsWith, Value: "Jane"},
		},
	}
	// Apply filters to struct data
	filteredStructData := filters.ApplyGroup(structData, group2)

	fmt.Println("Filtered Struct Data")
	// Print filtered struct data
	for _, item := range filteredStructData {
		fmt.Println(item)
	}
}

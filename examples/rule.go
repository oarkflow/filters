package main

import (
	"fmt"
	"time"

	"github.com/oarkflow/filters"
)

func main() {
	seq := filters.NewRule()
	filter1 := filters.NewFilter("LoggedInAt", filters.Between, []string{"{{CreatedAt}}", "{{VerifiedAt}}"})
	filter2 := filters.NewFilter("Name", filters.Expression, "Name=='Jane Doe'")
	filter3 := filters.NewFilter("Name", filters.Equal, "Bob Johnson")
	group1 := filters.NewFilterGroup(filters.AND, false, filter1, filter2)
	group2 := filters.NewFilterGroup(filters.AND, false, filter3)
	seq.AddCondition(filters.OR, false, group1, group2)

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
	seq.SetErrorResponse("this is test", "DENY")
	data, err := seq.Validate(Person{Age: 30, City: "Los Angeles", LoggedInAt: time.Date(2022, 06, 15, 15, 33, 0, 0, time.UTC), VerifiedAt: time.Date(2022, 06, 15, 15, 45, 0, 0, time.UTC), CreatedAt: time.Date(2022, 06, 15, 15, 30, 0, 0, time.UTC), Name: "Jane Doe"})
	if err != nil {
		panic(err)
	}
	fmt.Println(data)

	// Validate filters to struct data
	filteredStructData := filters.FilterCondition(structData, seq)

	fmt.Println("Filtered Struct Data")
	// Print filtered struct data
	for _, item := range filteredStructData {
		fmt.Println(item)
	}
}

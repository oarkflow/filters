package main

import (
	"fmt"
	"reflect"
)

// Define possible actions
const (
	DENY                      = "DENY"
	WARNING                   = "WARNING"
	WARNING_WITH_CONFIRMATION = "WARNING_WITH_CONFIRMATION"
	ALLOW                     = "ALLOW"
)

// Define binary operators
const (
	AND = "AND"
	OR  = "OR"
	NOT = "NOT"
)

// UserData interface
type UserData interface{}

// Matcher interface
type Matcher interface {
	Match(data UserData) bool
}

// Condition struct
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
}

// Match method for Condition
func (c Condition) Match(data UserData) bool {
	fieldValue := getFieldValue(data, c.Field)
	return compare(fieldValue, c.Operator, c.Value)
}

// Function to get the value of a field from UserData
func getFieldValue(data UserData, field string) interface{} {
	v := reflect.ValueOf(data)

	// Check if the data is a map
	if v.Kind() == reflect.Map {
		if value := v.MapIndex(reflect.ValueOf(field)); value.IsValid() {
			return value.Interface()
		}
		return nil
	}

	// Check if the data is a struct
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(field)
		if f.IsValid() {
			return f.Interface()
		}
		return nil
	}

	return nil
}

// Function to compare values based on operator
func compare(fieldValue interface{}, operator string, value interface{}) bool {
	switch operator {
	case "==":
		return fieldValue == value
	case "!=":
		return fieldValue != value
	case "<":
		return fieldValue.(int) < value.(int)
	case "<=":
		return fieldValue.(int) <= value.(int)
	case ">":
		return fieldValue.(int) > value.(int)
	case ">=":
		return fieldValue.(int) >= value.(int)
	case "&&":
		return fieldValue.(bool) && value.(bool)
	case "||":
		return fieldValue.(bool) || value.(bool)
	default:
		return false
	}
}

// Rule struct
type Rule struct {
	Conditions  []Matcher // List of conditions
	Conjunction string    // AND, OR, NOT
	Action      string    // Action to take if rule fails
	Message     string    // Message for the rule
}

// Validation result structure
type ValidationResult struct {
	Status  bool   // Status of the validation
	Message string // Message for the validation
	Action  string // Action to take
}

// Rule engine function to validate data
func validateData(data UserData, rules []Rule) []ValidationResult {
	results := []ValidationResult{}

	for _, rule := range rules {
		isValid := evaluateRule(data, rule)
		results = append(results, ValidationResult{
			Status:  !isValid, // Status is true if the rule is not valid
			Message: "",
			Action:  ALLOW, // Default action if valid
		})

		if !isValid {
			results[len(results)-1].Message = rule.Message
			results[len(results)-1].Action = rule.Action
		}
	}

	return results
}

// Evaluate a rule based on its conjunction
func evaluateRule(data UserData, rule Rule) bool {
	var result bool

	for i, condition := range rule.Conditions {
		if i == 0 {
			result = condition.Match(data) // First condition sets the initial result
		} else {
			switch rule.Conjunction {
			case AND:
				result = result && condition.Match(data)
			case OR:
				result = result || condition.Match(data)
			case NOT:
				result = !condition.Match(data) // NOT operator inverts the condition
			}
		}
	}

	return result
}

func main() {
	// Define rules
	rules := []Rule{
		{
			Conditions: []Matcher{
				Condition{Field: "Age", Operator: ">=", Value: 18},
			},
			Conjunction: AND,
			Action:      DENY,
			Message:     "User is under 18 years old.",
		},
		{
			Conditions: []Matcher{
				Condition{Field: "Salary", Operator: "<", Value: 30000},
			},
			Conjunction: AND,
			Action:      WARNING,
			Message:     "Salary is below the minimum threshold.",
		},
		{
			Conditions: []Matcher{
				Condition{Field: "Country", Operator: "!=", Value: "USA"},
			},
			Conjunction: AND,
			Action:      WARNING_WITH_CONFIRMATION,
			Message:     "User is not located in the USA.",
		},
		{
			Conditions: []Matcher{
				Condition{Field: "IsActive", Operator: "==", Value: true},
			},
			Conjunction: AND,
			Action:      ALLOW,
			Message:     "User is active.",
		},
	}

	// Example struct data
	userStruct := struct {
		Age      int
		Salary   int
		Country  string
		IsActive bool
	}{
		Age:      16,
		Salary:   25000,
		Country:  "Canada",
		IsActive: true,
	}

	// Example map data
	userMap := map[string]interface{}{
		"Age":      16,
		"Salary":   25000,
		"Country":  "Canada",
		"IsActive": true,
	}

	// Validate the struct data
	fmt.Println("Validating struct data:")
	validationResults := validateData(userStruct, rules)
	for _, result := range validationResults {
		fmt.Printf("Status: %t, Message: %s, Action: %s\n", result.Status, result.Message, result.Action)
	}

	// Validate the map data
	fmt.Println("\nValidating map data:")
	validationResults = validateData(userMap, rules)
	for _, result := range validationResults {
		fmt.Printf("Status: %t, Message: %s, Action: %s\n", result.Status, result.Message, result.Action)
	}
}

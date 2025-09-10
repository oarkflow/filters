# Filters

A powerful Go library for filtering data based on rules and conditions.

## Features

- Rule-based filtering for maps and structs
- Support for various operators: equal, not equal, greater than, less than, contains, starts with, ends with, in, between, pattern (regex), expression, and more
- Case-sensitive and case-insensitive string operations
- SQL-like query parsing
- Query string parsing for web APIs
- Lookup functionality for complex data relationships
- Count-based operations for arrays
- Null and zero value checks

## Installation

```bash
go get github.com/oarkflow/filters
```

## Quick Start

### Basic Filtering

```go
package main

import (
    "fmt"
    "github.com/oarkflow/filters"
)

func main() {
    data := []map[string]any{
        {"name": "John", "age": 30},
        {"name": "Jane", "age": 25},
    }

    filter := filters.NewFilter("age", filters.GreaterThan, 26)
    result := filters.ApplyGroup(data, &filters.FilterGroup{
        Filters: []filters.Condition{filter},
    })

    fmt.Println(result) // [{"name":"John","age":30}]
}
```

### Query String Parsing

```go
filters, err := filters.ParseQuery("?name:contains:John&age:gt:25")
if err != nil {
    panic(err)
}
// Use filters...
```

### SQL-like Parsing

```go
rule, err := filters.ParseSQL("SELECT * FROM users WHERE age > 25 AND name LIKE 'J%'")
if err != nil {
    panic(err)
}
// Use rule...
```

## Operators

### Comparison Operators
- `eq` / `ne` - Equal / Not Equal
- `gt` / `ge` / `lt` / `le` - Greater Than / Greater Equal / Less Than / Less Equal

### String Operators
- `contains` / `ncontains` - Contains / Not Contains (case-insensitive)
- `startswith` / `nstartswith` - Starts With / Not Starts With (case-insensitive)
- `endswith` / `nendswith` - Ends With / Not Ends With (case-insensitive)
- `contains_cs` / `ncontains_cs` - Case-sensitive versions
- `startswith_cs` / `nstartswith_cs` - Case-sensitive versions
- `endswith_cs` / `nendswith_cs` - Case-sensitive versions

### Other Operators
- `in` / `nin` - In / Not In
- `between` - Between two values
- `pattern` - Regex pattern matching
- `expr` - Expression evaluation
- `null` / `nnull` - Is Null / Is Not Null
- `izero` / `nzero` - Is Zero / Is Not Zero

### Case-Sensitive String Operators
- `contains_cs` / `ncontains_cs` - Contains / Not Contains (case-sensitive)
- `startswith_cs` / `nstartswith_cs` - Starts With / Not Starts With (case-sensitive)
- `endswith_cs` / `nendswith_cs` - Ends With / Not Ends With (case-sensitive)

### Count Operators (for arrays)
- `gtc` / `gec` / `ltc` / `lec` / `eqc` / `nec` - Greater/Less/Equal count

## Advanced Usage

### With Lookups

```go
filter := filters.NewFilter("items", filters.GreaterThanEqualCount, 1)
lookup := &filters.Lookup{
    Data: []map[string]any{{"id": 1, "active": true}},
    Condition: "lookup.active == true",
}
filter.SetLookup(lookup)
```

### Complex Rules

```go
rule := filters.NewRule()
rule.AddCondition(filters.AND, false,
    filters.NewFilter("age", filters.GreaterThan, 18),
    filters.NewFilter("status", filters.Equal, "active"),
)
```

## Examples

### Null and Zero Checks

```go
// Check for null values
nullFilter := filters.NewFilter("city", filters.IsNull, nil)
filtered := filters.ApplyGroup(data, &filters.FilterGroup{
    Operator: filters.AND,
    Filters:  []filters.Condition{nullFilter},
})

// Check for zero values
zeroFilter := filters.NewFilter("score", filters.IsZero, nil)
filtered = filters.ApplyGroup(data, &filters.FilterGroup{
    Operator: filters.AND,
    Filters:  []filters.Condition{zeroFilter},
})
```

### Case-Sensitive String Operations

```go
// Case-sensitive contains
csFilter := filters.NewFilter("name", filters.ContainsCS, "John")
filtered := filters.ApplyGroup(data, &filters.FilterGroup{
    Operator: filters.AND,
    Filters:  []filters.Condition{csFilter},
})

// Case-insensitive contains (default)
ciFilter := filters.NewFilter("name", filters.Contains, "John")
filtered = filters.ApplyGroup(data, &filters.FilterGroup{
    Operator: filters.AND,
    Filters:  []filters.Condition{ciFilter},
})
```

### Query String Parsing with Operators

```go
// Parse query with operators
filters, err := filters.ParseQuery("?city=isnull&active=eq:true&name=contains_cs:John")
if err != nil {
    panic(err)
}

group := &filters.FilterGroup{
    Operator: filters.AND,
    Filters:  filters, // []*Filter implements []Condition
}
filtered := filters.ApplyGroup(data, group)
```

See the `examples/` directory for more usage examples.

## License

MIT License

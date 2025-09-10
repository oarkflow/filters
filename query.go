package filters

import (
	"errors"
	"net/url"
	"slices"
	"strings"
)

// ParseQuery parses the query string and returns Filter or Query.
func ParseQuery(queryString string, exceptFields ...string) (filters []*Filter, err error) {
	queryParams, err1 := url.ParseQuery(strings.TrimPrefix(queryString, "?"))
	if err != nil {
		err = err1
		return
	}
	for key, values := range queryParams {
		if strings.Contains(key, ":") {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				field := parts[0]
				if len(exceptFields) > 0 && slices.Contains(exceptFields, field) {
					continue
				}
				operator := strings.ToLower(parts[1])
				if _, exists := validOperators[Operator(operator)]; exists {
					// Operators that don't require value
					if Operator(operator) == IsNull || Operator(operator) == NotNull || Operator(operator) == IsZero || Operator(operator) == NotZero {
						filters = append(filters, NewFilter(field, Operator(operator), nil))
					} else {
						filters = append(filters, NewFilter(field, Operator(operator), ""))
					}
				} else {
					filters = append(filters, NewFilter(field, Equal, parts[1]))
				}
			} else if len(parts) == 3 {
				if len(exceptFields) > 0 && slices.Contains(exceptFields, parts[0]) {
					continue
				}
				// Handle complex field:operator:value
				field := parts[0]
				operator := parts[1]
				opValue := parts[2]
				if _, exists := validOperators[Operator(strings.ToLower(operator))]; !exists {
					return nil, errors.New("invalid operator " + operator)
				}
				// For between operator, split values into two parts
				var val any
				if strings.Contains(opValue, ",") {
					betweenParts := strings.Split(opValue, ",")
					if Operator(operator) == Between && len(betweenParts) != 2 {
						return nil, errors.New("operator must have at least two values")
					}
					if Operator(operator) == In && len(betweenParts) < 1 {
						return nil, errors.New("operator must have at least two values")
					}
					for i, p := range betweenParts {
						p = strings.TrimSpace(p)
						betweenParts[i] = p
					}

					val = betweenParts
				} else {
					val = opValue
				}
				filters = append(filters, NewFilter(field, Operator(operator), val))
			}
		} else {
			if len(exceptFields) > 0 && slices.Contains(exceptFields, key) {
				continue
			}
			if len(values) == 1 {
				value := values[0]
				if _, exists := validOperators[Operator(strings.ToLower(value))]; exists {
					// Operators that don't require value
					if Operator(strings.ToLower(value)) == IsNull || Operator(strings.ToLower(value)) == NotNull || Operator(strings.ToLower(value)) == IsZero || Operator(strings.ToLower(value)) == NotZero {
						filters = append(filters, NewFilter(key, Operator(value), nil))
					} else {
						filters = append(filters, NewFilter(key, Operator(value), ""))
					}
				} else if strings.Contains(value, ":") {
					parts := strings.Split(value, ":")
					if len(parts) == 2 {
						operator := parts[0]
						opValue := parts[1]
						if _, exists := validOperators[Operator(strings.ToLower(operator))]; !exists {
							return nil, errors.New("invalid operator " + operator)
						}
						// For between operator, split values into two parts
						var val any
						if strings.Contains(opValue, ",") {
							betweenParts := strings.Split(opValue, ",")
							if Operator(operator) == Between && len(betweenParts) != 2 {
								return nil, errors.New("operator must have at least two values")
							}
							if Operator(operator) == In && len(betweenParts) < 1 {
								return nil, errors.New("operator must have at least two values")
							}
							for i, p := range betweenParts {
								p = strings.TrimSpace(p)
								betweenParts[i] = p
							}

							val = betweenParts
						} else {
							val = opValue
						}
						filters = append(filters, NewFilter(key, Operator(operator), val))
					} else {
						filters = append(filters, NewFilter(key, Equal, value))
					}
				} else {
					filters = append(filters, NewFilter(key, Equal, value))
				}
			} else if len(values) > 1 {
				filters = append(filters, NewFilter(key, In, values))
			}
		}
	}
	return
}

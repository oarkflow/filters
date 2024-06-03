package filters

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/oarkflow/filters/utils"
)

type Operator string

const (
	EQUAL              Operator = "eq"
	LESS_THAN          Operator = "lt"
	LESS_THAN_EQUAL    Operator = "le"
	GREATER_THAN       Operator = "gt"
	GREATER_THAN_EQUAL Operator = "ge"
	NOT_EQUAL          Operator = "ne"
	CONTAINS           Operator = "contains"
	NOT_CONTAINS       Operator = "not_contains"
	BETWEEN            Operator = "between"
	IN                 Operator = "in"
	STARTS_WITH        Operator = "starts_with"
	ENDS_WITH          Operator = "ends_with"
)

var (
	validOperators = map[Operator]struct{}{
		EQUAL:              {},
		LESS_THAN:          {},
		LESS_THAN_EQUAL:    {},
		GREATER_THAN:       {},
		GREATER_THAN_EQUAL: {},
		NOT_EQUAL:          {},
		CONTAINS:           {},
		NOT_CONTAINS:       {},
		BETWEEN:            {},
		IN:                 {},
		STARTS_WITH:        {},
		ENDS_WITH:          {},
	}
)

type BooleanOperator string

const (
	AND BooleanOperator = "AND"
	OR  BooleanOperator = "OR"
	NOT BooleanOperator = "NOT"
)

type Filter struct {
	Field    string
	Operator Operator
	Value    any
}

type FilterGroup struct {
	Operator BooleanOperator
	Filters  []Filter
}

type Query struct {
	Bool BoolQuery `json:"bool"`
}

type BoolQuery struct {
	Must    []any `json:"must"`
	Should  []any `json:"should"`
	MustNot []any `json:"must_not"`
}

type TermQuery struct {
	Term map[string]any `json:"term"`
}

type RangeQuery struct {
	Range map[string]map[string]any `json:"range"`
}

type WildcardQuery struct {
	Wildcard map[string]string `json:"wildcard"`
}

func ValidateFilters(filters ...Filter) error {
	for _, filter := range filters {
		if filter.Field == "" {
			return errors.New("filter field cannot be empty")
		}
		if _, exists := validOperators[filter.Operator]; !exists {
			return fmt.Errorf("invalid operator: %s", filter.Operator)
		}
		if filter.Operator == BETWEEN {
			if reflect.TypeOf(filter.Value).Kind() != reflect.Slice || reflect.ValueOf(filter.Value).Len() != 2 {
				return errors.New("between filter must have a slice of two elements as value")
			}
		}
		if filter.Operator == IN {
			if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
				return errors.New("in filter must have a slice as value")
			}
		}
	}

	return nil
}

type BinaryExpr[T any] struct {
	Left     *FilterGroup
	Operator BooleanOperator
	Right    *FilterGroup
}

func ApplyBinaryFilter[T any](data []T, expr BinaryExpr[T]) ([]T, error) {
	if expr.Left == nil || expr.Right == nil {
		return nil, errors.New("missing left or right filter group")
	}

	leftResult, err := ApplyGroup(data, []FilterGroup{*expr.Left})
	if err != nil {
		return nil, err
	}

	rightResult, err := ApplyGroup(data, []FilterGroup{*expr.Right})
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case AND:
		return utils.Intersection(leftResult, rightResult), nil
	case OR:
		return utils.Union(leftResult, rightResult), nil
	default:
		return nil, errors.New("unsupported boolean operator")
	}
}

func filterToQuery(filter Filter) any {
	switch filter.Operator {
	case EQUAL:
		return TermQuery{Term: map[string]any{filter.Field: filter.Value}}
	case LESS_THAN:
		return RangeQuery{Range: map[string]map[string]any{filter.Field: {"lt": filter.Value}}}
	case LESS_THAN_EQUAL:
		return RangeQuery{Range: map[string]map[string]any{filter.Field: {"le": filter.Value}}}
	case GREATER_THAN:
		return RangeQuery{Range: map[string]map[string]any{filter.Field: {"gt": filter.Value}}}
	case GREATER_THAN_EQUAL:
		return RangeQuery{Range: map[string]map[string]any{filter.Field: {"ge": filter.Value}}}
	case NOT_EQUAL:
		return BoolQuery{Must: []any{TermQuery{Term: map[string]any{filter.Field: map[string]any{"ne": filter.Value}}}}}
	case CONTAINS:
		return WildcardQuery{Wildcard: map[string]string{filter.Field: fmt.Sprintf("*%v*", filter.Value)}}
	case NOT_CONTAINS:
		return BoolQuery{Must: []any{WildcardQuery{Wildcard: map[string]string{filter.Field: fmt.Sprintf("!*%v*", filter.Value)}}}}
	case STARTS_WITH:
		return WildcardQuery{Wildcard: map[string]string{filter.Field: fmt.Sprintf("%v*", filter.Value)}}
	case ENDS_WITH:
		return WildcardQuery{Wildcard: map[string]string{filter.Field: fmt.Sprintf("*%v", filter.Value)}}
	case IN:
		return TermQuery{Term: map[string]any{filter.Field: map[string]any{"in": filter.Value}}}
	case BETWEEN:
		values, ok := filter.Value.([]any)
		if !ok || len(values) != 2 {
			return nil
		}
		return RangeQuery{Range: map[string]map[string]any{filter.Field: {"gte": values[0], "lte": values[1]}}}
	default:
		return nil
	}
}

func filtersToQuery(filters []Filter) (Query, error) {
	if err := ValidateFilters(filters...); err != nil {
		return Query{}, err
	}

	var mustQueries []any
	for _, filter := range filters {
		query := filterToQuery(filter)
		if query != nil {
			mustQueries = append(mustQueries, query)
		}
	}
	return Query{
		Bool: BoolQuery{
			Must: mustQueries,
		},
	}, nil
}

func ApplyGroup[T any](data []T, filterGroups []FilterGroup) ([]T, error) {
	var result []T
	for _, item := range data {
		matches := true
		for _, group := range filterGroups {
			if !MatchGroup(item, group) {
				matches = false
				break
			}
		}
		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

func MatchGroup[T any](item T, group FilterGroup) bool {
	switch group.Operator {
	case AND:
		for _, filter := range group.Filters {
			if !Match(item, filter) {
				return false
			}
		}
		return true
	case OR:
		for _, filter := range group.Filters {
			if Match(item, filter) {
				return true
			}
		}
		return false
	case NOT:
		for _, filter := range group.Filters {
			if Match(item, filter) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func Match[T any](item T, filter Filter) bool {
	fieldVal := reflect.ValueOf(item)
	var fieldValue reflect.Value
	var val any

	if fieldVal.Kind() == reflect.Map {
		mapKey := reflect.ValueOf(filter.Field)
		if !mapKey.IsValid() {
			return false
		}
		fieldValue = fieldVal.MapIndex(mapKey)
		if !fieldValue.IsValid() {
			return false
		}
		val = fieldValue.Interface()
	} else {
		fieldValue = utils.GetFieldName(fieldVal, filter.Field)
		if !fieldValue.IsValid() {
			return false
		}
		val = fieldValue.Interface()
	}

	if val == nil {
		return false
	}

	switch filter.Operator {
	case EQUAL:
		return reflect.DeepEqual(val, filter.Value)
	case LESS_THAN:
		return utils.Compare(val, filter.Value) < 0
	case LESS_THAN_EQUAL:
		return utils.Compare(val, filter.Value) <= 0
	case GREATER_THAN:
		return utils.Compare(val, filter.Value) > 0
	case GREATER_THAN_EQUAL:
		return utils.Compare(val, filter.Value) >= 0
	case NOT_EQUAL:
		return !reflect.DeepEqual(val, filter.Value)
	case CONTAINS:
		return strings.Contains(val.(string), filter.Value.(string))
	case NOT_CONTAINS:
		return !strings.Contains(val.(string), filter.Value.(string))
	case STARTS_WITH:
		return strings.HasPrefix(val.(string), filter.Value.(string))
	case ENDS_WITH:
		return strings.HasSuffix(val.(string), filter.Value.(string))
	case IN:
		switch vt := filter.Value.(type) {
		case []string:
			for _, v := range vt {
				if reflect.DeepEqual(val, v) {
					return true
				}
			}
		case []any:
			for _, v := range vt {
				if reflect.DeepEqual(val, v) {
					return true
				}
			}
		}
		return false
	case BETWEEN:
		switch values := filter.Value.(type) {
		case []string:
			return utils.Compare(val, values[0]) >= 0 && utils.Compare(val, values[1]) <= 0
		case []any:
			return utils.Compare(val, values[0]) >= 0 && utils.Compare(val, values[1]) <= 0
		}
		return false
	default:
		return false
	}
}

// Converts string to appropriate type (int, float, time, or string)
func convertValue(value string) (any, error) {
	if i, err := strconv.Atoi(value); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return value, nil
}

func New(field string, operator Operator, value any) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// ParseQuery parses the query string and returns Filter or Query.
func ParseQuery(url *url.URL) ([]Filter, error) {
	queryParams := url.Query()
	var filters []Filter

	for key, values := range queryParams {
		if strings.Contains(key, ":") {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				filters = append(filters, New(parts[0], EQUAL, parts[1]))
			} else if len(parts) == 3 {
				// Handle complex field:operator:value
				field := parts[0]
				operator := parts[1]
				opValue := parts[2]
				if _, exists := validOperators[Operator(strings.ToLower(operator))]; !exists {
					return nil, errors.New("invalid operator")
				}
				// For between operator, split values into two parts
				var val any
				if strings.Contains(opValue, ",") {
					betweenParts := strings.Split(opValue, ",")
					if Operator(operator) == BETWEEN && len(betweenParts) != 2 {
						return nil, errors.New("operator must have at least two values")
					}
					if Operator(operator) == IN && len(betweenParts) < 1 {
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
				filters = append(filters, New(field, Operator(operator), val))
			}
		} else {
			if len(values) == 1 {
				filters = append(filters, New(key, EQUAL, values[0]))
			} else if len(values) > 1 {
				filters = append(filters, New(key, IN, values))
			}
		}
	}
	return filters, nil
}

// Helper function to remove a filter from a slice of filters
func removeFilter(filters []Filter, filterToRemove Filter) []Filter {
	for i, filter := range filters {
		if filter.Field == filterToRemove.Field && filter.Operator == filterToRemove.Operator && reflect.DeepEqual(filter.Value, filterToRemove.Value) {
			return append(filters[:i], filters[i+1:]...)
		}
	}
	return filters
}

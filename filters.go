package filters

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oarkflow/filters/utils"
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

func (filter Filter) Validate() error {
	if filter.Field == "" {
		return errors.New("filter field cannot be empty")
	}
	if _, exists := validOperators[filter.Operator]; !exists {
		return fmt.Errorf("invalid operator: %s", filter.Operator)
	}
	if filter.Operator == Between {
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice || reflect.ValueOf(filter.Value).Len() != 2 {
			return errors.New("between filter must have a slice of two elements as value")
		}
	}
	if filter.Operator == In {
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
			return errors.New("in filter must have a slice as value")
		}
	}
	return nil
}

type Group[T any] struct {
	Left     *FilterGroup
	Operator BooleanOperator
	Right    *FilterGroup
}

func (expr Group[T]) Match(data []T) ([]T, error) {
	if expr.Left == nil || expr.Right == nil {
		return nil, errors.New("missing left or right filter group")
	}

	leftResult, err := ApplyGroup(data, *expr.Left)
	if err != nil {
		return nil, err
	}

	rightResult, err := ApplyGroup(data, *expr.Right)
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

func ApplyGroup[T any](data []T, filterGroups ...FilterGroup) ([]T, error) {
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

func New(field string, operator Operator, value any) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

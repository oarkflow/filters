package filters

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oarkflow/filters/utils"
)

type Filter struct {
	Field     string
	Operator  Operator
	Value     any
	Reverse   bool
	err       error
	validated bool
}

type FilterGroup struct {
	Operator Boolean
	Filters  []*Filter
	Reverse  bool
}

type Join struct {
	Left     *FilterGroup
	Right    *FilterGroup
	Operator Boolean
	Reverse  bool
}

func (filter *Filter) Validate() error {
	filter.validated = true
	if filter.Field == "" {
		filter.err = errors.New("filter field cannot be empty")
		return filter.err
	}
	if _, exists := validOperators[filter.Operator]; !exists {
		filter.err = fmt.Errorf("invalid operator: %s", filter.Operator)
		return filter.err
	}
	if filter.Operator == Between {
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice || reflect.ValueOf(filter.Value).Len() != 2 {
			filter.err = errors.New("between filter must have a slice of two elements as value")
			return filter.err
		}
	}
	if filter.Operator == In {
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
			filter.err = errors.New("in filter must have a slice as value")
			return filter.err
		}
	}
	return nil
}

func FilterJoin[T any](data []T, expr Join) ([]T, error) {
	if expr.Left == nil || expr.Right == nil {
		return nil, errors.New("missing left or right filter group")
	}

	leftResult := ApplyGroup(data, expr.Left)
	rightResult := ApplyGroup(data, expr.Right)
	switch expr.Operator {
	case AND:
		return utils.Intersection(leftResult, rightResult), nil
	case OR:
		return utils.Union(leftResult, rightResult), nil
	default:
		return nil, errors.New("unsupported boolean operator")
	}
}

func MatchJoin[T any](item T, expr Join) bool {
	if expr.Left == nil || expr.Right == nil {
		return false
	}
	leftResult := MatchGroup(item, expr.Left)
	rightResult := MatchGroup(item, expr.Right)
	switch expr.Operator {
	case AND:
		if expr.Reverse {
			return !(leftResult && rightResult)
		}
		return leftResult && rightResult
	case OR:
		if expr.Reverse {
			return !(leftResult || rightResult)
		}
		return leftResult || rightResult
	default:
		return false
	}
}

func ApplyGroup[T any](data []T, filterGroups ...*FilterGroup) (result []T) {
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
	return
}

func MatchGroup[T any](item T, group *FilterGroup) bool {
	switch group.Operator {
	case AND:
		matched := true
		for _, filter := range group.Filters {
			if !Match(item, filter) {
				matched = false
				break
			}
		}
		if group.Reverse {
			return !matched
		}
		return matched
	case OR:
		matched := false
		for _, filter := range group.Filters {
			if Match(item, filter) {
				matched = true
			}
		}
		if group.Reverse {
			return !matched
		}
		return matched
	default:
		return false
	}
}

func NewFilter(field string, operator Operator, value any) *Filter {
	return &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

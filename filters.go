package filters

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/oarkflow/xid"
)

type Lookup struct {
	Data             any `json:"data"`
	Handler          func(string) (any, error)
	Type             string `json:"type"`
	Source           string `json:"source"`
	Condition        string `json:"condition"`
	HandlerCondition string `json:"handler_condition"`
}

type Filter struct {
	Key       string   `json:"key"`
	Field     string   `json:"field"`
	Operator  Operator `json:"operator"`
	Value     any      `json:"value"`
	Reverse   bool     `json:"reverse"`
	Lookup    *Lookup  `json:"lookup"`
	err       error
	validated bool
}

func (filter *Filter) Match(data any) bool {
	return Match(data, filter)
}

func (filter *Filter) SetLookup(lookup *Lookup) {
	filter.Lookup = lookup
}

func Match[T any](item T, filter *Filter) bool {
	matched := match(item, filter)
	if filter.Reverse {
		return !matched
	}
	return matched
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
	if filter.Operator == In && filter.Lookup == nil {
		if reflect.TypeOf(filter.Value).Kind() != reflect.Slice {
			filter.err = errors.New("in filter must have a slice as value")
			return filter.err
		}
	}
	return nil
}

func NewFilter(field string, operator Operator, value any, keys ...string) *Filter {
	filter := &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
	if len(keys) > 0 {
		filter.Key = keys[0]
	} else {
		filter.Key = xid.New().String()
	}
	return filter
}

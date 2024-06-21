package filters

import (
	"reflect"
	"slices"
	"strings"

	"github.com/oarkflow/filters/convert"
	"github.com/oarkflow/filters/utils"
)

type Operator string

const (
	Equal            Operator = "eq"
	LessThan         Operator = "lt"
	LessThanEqual    Operator = "le"
	GreaterThan      Operator = "gt"
	GreaterThanEqual Operator = "ge"
	NotEqual         Operator = "ne"
	Contains         Operator = "contains"
	NotContains      Operator = "not_contains"
	Between          Operator = "between"
	In               Operator = "in"
	StartsWith       Operator = "starts_with"
	EndsWith         Operator = "ends_with"
	NotIn            Operator = "not_in"
	NotZero          Operator = "not_zero"
	IsZero           Operator = "is_zero"
	IsNull           Operator = "is_null"
	NotNull          Operator = "not_null"
)

var (
	validOperators = map[Operator]struct{}{
		Equal:            {},
		LessThan:         {},
		LessThanEqual:    {},
		GreaterThan:      {},
		GreaterThanEqual: {},
		NotEqual:         {},
		Contains:         {},
		NotContains:      {},
		Between:          {},
		In:               {},
		StartsWith:       {},
		EndsWith:         {},
	}
)

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
	case Equal:
		return checkEq(val, filter)
	case NotEqual:
		return checkNeq(val, filter)
	case GreaterThan:
		return checkGt(val, filter)
	case LessThan:
		return checkLt(val, filter)
	case GreaterThanEqual:
		return checkGte(val, filter)
	case LessThanEqual:
		return checkLte(val, filter)
	case Between:
		return checkBetween(val, filter)
	case In:
		return checkIn(val, filter)
	case NotIn:
		return checkNotIn(val, filter)
	case Contains:
		return checkContains(val, filter)
	case NotContains:
		return checkNotContains(val, filter)
	case StartsWith:
		return checkStartsWith(val, filter)
	case EndsWith:
		return checkEndsWith(val, filter)
	case IsZero:
		return reflect.ValueOf(val).IsZero()
	case NotZero:
		return !fieldValue.IsZero()
	case IsNull:
		return fieldValue.IsNil()
	case NotNull:
		return !fieldValue.IsNil()
	}
	return false
}

func checkEq[T comparable](val T, filter Filter) bool {
	data, ok := convert.To(val, filter.Value)
	if !ok {
		return ok
	}
	return val == data
}

func checkNeq[T comparable](val T, filter Filter) bool {
	data, ok := convert.To(val, filter.Value)
	if !ok {
		return ok
	}
	return val != data
}

func checkGt[T comparable](data T, filter Filter) bool {
	return convert.Compare(data, filter.Value) > 0
}

func checkLt[T any](data T, filter Filter) bool {
	return convert.Compare(data, filter.Value) < 0
}

func checkGte[T any](data T, filter Filter) bool {
	return convert.Compare(data, filter.Value) >= 0
}

func checkLte[T any](data T, filter Filter) bool {
	return convert.Compare(data, filter.Value) <= 0
}

func checkBetween[T any](data T, filter Filter) bool {
	switch values := filter.Value.(type) {
	case []string:
		return utils.Compare(data, values[0]) >= 0 && utils.Compare(data, values[1]) <= 0
	case []any:
		return utils.Compare(data, values[0]) >= 0 && utils.Compare(data, values[1]) <= 0
	}
	return false
}

func checkIn[T comparable](data T, filter Filter) bool {
	sl, ok := convert.ToSlice(data, filter.Value)
	if !ok {
		return false
	}
	return slices.Contains(sl, data)
}

func checkNotIn[T comparable](data T, filter Filter) bool {
	sl, ok := convert.ToSlice(data, filter.Value)
	if !ok {
		return false
	}
	return !slices.Contains(sl, data)
}

func checkContains[T comparable](data T, filter Filter) bool {
	switch val := any(data).(type) {
	case string:
		switch gtVal := filter.Value.(type) {
		case string:
			return strings.Contains(val, gtVal)
		}
		return false
	}

	return false
}

func checkNotContains[T any](data T, filter Filter) bool {
	switch val := any(data).(type) {
	case string:
		switch gtVal := filter.Value.(type) {
		case string:
			return !strings.Contains(val, gtVal)
		}
		return false
	}

	return false
}

func checkStartsWith[T any](data T, filter Filter) bool {
	switch val := any(data).(type) {
	case string:
		switch gtVal := filter.Value.(type) {
		case string:
			return strings.HasPrefix(val, gtVal)
		}
		return false
	}

	return false
}

func checkEndsWith[T any](data T, filter Filter) bool {
	switch val := any(data).(type) {
	case string:
		switch gtVal := filter.Value.(type) {
		case string:
			return strings.HasSuffix(val, gtVal)
		}
		return false
	}

	return false
}

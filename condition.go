package filters

import (
	"reflect"
	"strings"

	"github.com/oarkflow/expr"

	"github.com/oarkflow/filters/convert"
	"github.com/oarkflow/filters/utils"
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

func Match[T any](item T, filter *Filter) bool {
	matched := match(item, filter)
	if filter.Reverse {
		return !matched
	}
	return matched
}

func match[T any](item T, filter *Filter) bool {
	if filter.validated {
		if filter.err != nil {
			return false
		}
	} else {
		if filter.Validate() != nil {
			return false
		}
	}
	fieldVal := reflect.ValueOf(item)
	fieldValue := getFieldValue(fieldVal, filter.Field)
	if !fieldValue.IsValid() {
		return false
	}
	val, err := resolveFilterValue(item, filter.Value)
	if err != nil {
		return false
	}
	switch filter.Operator {
	case Equal:
		return checkEq(fieldValue.Interface(), val)
	case NotEqual:
		return checkNeq(fieldValue.Interface(), val)
	case GreaterThan:
		return checkGt(fieldValue.Interface(), val)
	case LessThan:
		return checkLt(fieldValue.Interface(), val)
	case GreaterThanEqual:
		return checkGte(fieldValue.Interface(), val)
	case LessThanEqual:
		return checkLte(fieldValue.Interface(), val)
	case Between:
		return checkBetween(fieldValue.Interface(), val)
	case In:
		return checkIn(fieldValue.Interface(), val)
	case NotIn:
		return checkNotIn(fieldValue.Interface(), val)
	case Contains:
		return checkContains(fieldValue.Interface(), val)
	case NotContains:
		return checkNotContains(fieldValue.Interface(), val)
	case StartsWith:
		return checkStartsWith(fieldValue.Interface(), val)
	case EndsWith:
		return checkEndsWith(fieldValue.Interface(), val)
	case IsZero:
		return fieldValue.IsZero()
	case NotZero:
		return !fieldValue.IsZero()
	case IsNull:
		return fieldValue.IsNil()
	case NotNull:
		return !fieldValue.IsNil()
	}
	return false
}

// Resolve filter value that may reference another field
func resolveFilterValue(fieldVal, value any) (any, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
			referenceField := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(v, "}}"), "{{"))
			return expr.Eval(referenceField, fieldVal)
		}
		return v, nil
	case []string:
		var resolvedValues []any
		for _, val := range v {
			resolvedValue, err := resolveFilterValue(fieldVal, val)
			if err != nil {
				return nil, err
			}
			resolvedValues = append(resolvedValues, resolvedValue)
		}
		return resolvedValues, nil
	case []any:
		var resolvedValues []any
		for _, t := range v {
			switch t := t.(type) {
			case string:
				resolvedValue, err := resolveFilterValue(fieldVal, t)
				if err != nil {
					return nil, err
				}
				resolvedValues = append(resolvedValues, resolvedValue)
			}
		}
		return resolvedValues, nil
	default:
		return value, nil
	}
}

func getFieldValue(fieldVal reflect.Value, fieldName string) reflect.Value {
	if fieldVal.Kind() == reflect.Map {
		mapKey := reflect.ValueOf(fieldName)
		if !mapKey.IsValid() {
			return mapKey
		}
		fieldValue := fieldVal.MapIndex(mapKey)
		if !fieldValue.IsValid() {
			return fieldValue
		}
		return fieldValue
	}
	fieldValue := utils.GetFieldName(fieldVal, fieldName)
	if !fieldValue.IsValid() {
		return fieldValue
	}
	return fieldValue
}

func checkEq(val, value any) bool {
	switch val := val.(type) {
	case string:
		data, ok := convert.ToString(value)
		if !ok {
			return ok
		}
		return strings.EqualFold(val, data)
	default:
		data, ok := convert.To(val, value)
		if !ok {
			return ok
		}
		return val == data
	}
}

func checkNeq(val, value any) bool {
	switch val := val.(type) {
	case string:
		data, ok := convert.ToString(value)
		if !ok {
			return ok
		}
		return !strings.EqualFold(val, data)
	default:
		data, ok := convert.To(val, value)
		if !ok {
			return ok
		}
		return val != data
	}
}

func checkGt(data, value any) bool {
	return convert.Compare(data, value) > 0
}

func checkLt(data, value any) bool {
	return convert.Compare(data, value) < 0
}

func checkGte(data, value any) bool {
	return convert.Compare(data, value) >= 0
}

func checkLte(data, value any) bool {
	return convert.Compare(data, value) <= 0
}

func checkBetween(data, value any) bool {
	switch values := value.(type) {
	case []string:
		return utils.Compare(data, values[0]) >= 0 && utils.Compare(data, values[1]) <= 0
	case []any:
		return utils.Compare(data, values[0]) >= 0 && utils.Compare(data, values[1]) <= 0
	}
	return false
}

func checkIn(data, value any) bool {
	sl, ok := convert.ToSlice(data, value)
	if !ok {
		return false
	}
	return utils.Contains(sl, data)
}

func checkNotIn(data, value any) bool {
	sl, ok := convert.ToSlice(data, value)
	if !ok {
		return false
	}
	return !utils.Contains(sl, data)
}

func checkContains(data, value any) bool {
	switch val := data.(type) {
	case string:
		switch gtVal := value.(type) {
		case string:
			return strings.Contains(strings.ToLower(val), strings.ToLower(gtVal))
		}
		return false
	}

	return false
}

func checkNotContains(data, value any) bool {
	switch val := data.(type) {
	case string:
		switch gtVal := value.(type) {
		case string:
			return !strings.Contains(strings.ToLower(val), strings.ToLower(gtVal))
		}
		return false
	}

	return false
}

func checkStartsWith(data, value any) bool {
	switch val := data.(type) {
	case string:
		switch gtVal := value.(type) {
		case string:
			return strings.HasPrefix(strings.ToLower(val), strings.ToLower(gtVal))
		}
		return false
	}

	return false
}

func checkEndsWith(data, value any) bool {
	switch val := data.(type) {
	case string:
		switch gtVal := value.(type) {
		case string:
			return strings.HasSuffix(strings.ToLower(val), strings.ToLower(gtVal))
		}
		return false
	}

	return false
}

package filters

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/oarkflow/dipper"
	"github.com/oarkflow/expr"

	"github.com/oarkflow/convert"

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
		Expression:       {},
		Pattern:          {},
		Between:          {},
		In:               {},
		StartsWith:       {},
		NotStartsWith:    {},
		EndsWith:         {},
		NotEndsWith:      {},
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
	if !filter.validated {
		if filter.Validate() != nil {
			return false
		}
	}
	if filter.err != nil {
		return false
	}
	fieldValue, err := dipper.Get(item, filter.Field)
	if err != nil {
		return false
	}
	val, err := resolveFilterValue(item, filter.Value)
	if err != nil {
		return false
	}
	switch filter.Operator {
	case Equal:
		return checkEq(fieldValue, val)
	case NotEqual:
		return checkNeq(fieldValue, val)
	case GreaterThan:
		return checkGt(fieldValue, val)
	case LessThan:
		return checkLt(fieldValue, val)
	case GreaterThanEqual:
		return checkGte(fieldValue, val)
	case LessThanEqual:
		return checkLte(fieldValue, val)
	case Between:
		return checkBetween(fieldValue, val)
	case Expression:
		v, ok := filter.Value.(string)
		if !ok {
			return false
		}
		r, err := expr.Eval(v, item)
		fmt.Println(r, err)
		if err != nil || r == nil {
			return false
		}
		return true
	case Pattern:
		v, ok := filter.Value.(string)
		if !ok {
			return false
		}
		re, err := regexp.Compile(v)
		if err != nil {
			return false
		}
		vt, ok := fieldValue.(string)
		if !ok {
			return false
		}
		return re.MatchString(vt)
	case In:
		return checkIn(fieldValue, val)
	case EqualCount:
		return checkEqCount(fieldValue, val)
	case NotIn:
		return checkNotIn(fieldValue, val)
	case Contains:
		return checkContains(fieldValue, val)
	case NotContains:
		return checkNotContains(fieldValue, val)
	case StartsWith:
		return checkStartsWith(fieldValue, val)
	case EndsWith:
		return checkEndsWith(fieldValue, val)
	case NotStartsWith:
		return !checkStartsWith(fieldValue, val)
	case NotEndsWith:
		return !checkEndsWith(fieldValue, val)
	case IsZero:
		return reflect.ValueOf(fieldValue).IsZero()
	case NotZero:
		return !reflect.ValueOf(fieldValue).IsZero()
	case IsNull:
		return reflect.ValueOf(fieldValue).IsNil()
	case NotNull:
		return !reflect.ValueOf(fieldValue).IsNil()
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

func checkEqCount(val, value any) bool {
	valKind := reflect.ValueOf(val)
	if valKind.Kind() != reflect.Slice {
		if val == nil {
			return false
		}
		var dArray []any
		dArray = append(dArray, val)
		valKind = reflect.ValueOf(dArray)
	}
	var gtVal int
	switch v := value.(type) {
	case []any:
		gtVal = len(v)
	default:
		g, err := strconv.Atoi(fmt.Sprintf("%v", value))
		if err != nil {
			return false
		}
		gtVal = g
	}
	return valKind.Len() == gtVal && valKind.Len() != 0
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

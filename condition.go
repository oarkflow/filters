package filters

import (
	"fmt"
	"github.com/oarkflow/dipper"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/oarkflow/expr"

	"github.com/oarkflow/convert"

	"github.com/oarkflow/filters/utils"
)

var (
	validOperators = map[Operator]struct{}{
		In:                    {},
		Equal:                 {},
		Pattern:               {},
		Between:               {},
		LessThan:              {},
		NotEqual:              {},
		Contains:              {},
		EndsWith:              {},
		Expression:            {},
		EqualCount:            {},
		StartsWith:            {},
		NotContains:           {},
		GreaterThan:           {},
		NotEndsWith:           {},
		LessThanEqual:         {},
		NotEqualCount:         {},
		NotStartsWith:         {},
		LesserThanCount:       {},
		GreaterThanEqual:      {},
		GreaterThanCount:      {},
		LesserThanEqualCount:  {},
		GreaterThanEqualCount: {},
	}
	countOperators = []Operator{
		GreaterThanEqualCount,
		GreaterThanCount,
		LesserThanEqualCount,
		LesserThanCount,
		NotEqualCount,
		EqualCount,
	}
)

func validatedCount(input string, lookupData any) bool {
	rs, err := expr.Eval(input, map[string]any{"data": lookupData})
	if err != nil {
		return false
	}
	converted, done := convert.ToBool(rs)
	if !done {
		return false
	}
	return converted
}

func processValidate(op string, val any, lookupData, fieldValue any) bool {
	fieldValue, err := utils.FilterSlice(lookupData, fieldValue)
	if err != nil {
		return false
	}
	return validatedCount(fmt.Sprintf("len(data) %s %v", op, val), fieldValue)
}

func validateCount(op string, vat any, lookupData, fieldValue any) bool {
	val := reflect.ValueOf(fieldValue)
	if val.Kind() == reflect.Slice {
		if val.Len() > 0 && val.Index(0).Elem().Kind() == reflect.Slice {
			for i := 0; i < val.Len(); i++ {
				innerSlice := val.Index(i)
				if !processValidate(op, vat, lookupData, innerSlice.Interface()) {
					return false
				}
			}
			return true
		} else {
			return processValidate(op, vat, lookupData, fieldValue)
		}
	}
	return false
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
	var fieldValue any
	var err error
	if strings.Contains(filter.Field, "{{") {
		fieldValue, err = resolveFilterField(item, filter.Field)
		if err != nil {
			return false
		}
	} else {
		fieldValue, err = dipper.Get(item, filter.Field)
		if err != nil {
			return false
		}
	}
	val, err := resolveFilterValue(item, filter.Value)
	if err != nil {
		return false
	}
	var lookupData any
	if filter.Lookup != nil {
		if filter.Lookup.Data != nil {
			lookupData = filter.Lookup.Data
		} else if filter.Lookup.Handler != nil {
			rs, err := filter.Lookup.Handler(item, filter.Lookup.HandlerCondition)
			if err != nil {
				return false
			}
			lookupData = rs
		}
		if filter.Lookup.Condition != "" {
			lookupData = map[string]any{"data": item, "lookup": lookupData}
			rs, err := expr.Eval(filter.Lookup.Condition, lookupData)
			if err != nil {
				return false
			}
			lookupData = rs
		}
	}
	if lookupData != nil && utils.IsSlice(lookupData) {
		lookupLength, err := utils.GetSliceLength(lookupData)
		fieldLength, _ := utils.GetSliceLength(fieldValue)
		if utils.IsSlice(fieldValue) && fieldLength == 0 {
			return false
		}
		if err != nil {
			return false
		}
		if lookupLength == 0 {
			return true
		}
	}
	if !slices.Contains(countOperators, filter.Operator) && lookupData != nil {
		val = lookupData
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
	case GreaterThanEqualCount:
		return validateCount(">=", val, lookupData, fieldValue)
	case GreaterThanCount:
		return validateCount(">", val, lookupData, fieldValue)
	case LesserThanEqualCount:
		return validateCount("<=", val, lookupData, fieldValue)
	case LesserThanCount:
		return validateCount("<", val, lookupData, fieldValue)
	case NotEqualCount:
		return validateCount("!=", val, lookupData, fieldValue)
	case EqualCount:
		return validateCount("==", val, lookupData, fieldValue)
	case Expression:
		v, ok := filter.Value.(string)
		if !ok {
			return false
		}
		r, err := expr.Eval(v, item)
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

func resolveString(item any, v string) (any, error) {
	if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
		referenceField := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(v, "}}"), "{{"))
		return expr.Eval(referenceField, item)
	}
	return v, nil
}

func resolveFilterField(item, field any) (any, error) {
	switch v := field.(type) {
	case string:
		return resolveString(item, v)
	}
	return field, nil
}

// Resolve filter value that may reference another field
func resolveFilterValue(fieldVal, value any) (any, error) {
	switch v := value.(type) {
	case string:
		return resolveString(fieldVal, v)
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

func checkComparison(val, value any, isEqual bool) bool {
	var comparisonResult bool
	switch val := val.(type) {
	case string:
		data, ok := convert.ToString(value)
		if !ok {
			return ok
		}
		comparisonResult = strings.EqualFold(val, data)
	default:
		data, ok := convert.To(val, value)
		if !ok {
			return ok
		}
		comparisonResult = val == data
	}
	if isEqual {
		return comparisonResult
	}
	return !comparisonResult
}

func checkEq(val, value any) bool {
	return checkComparison(val, value, true)
}

func checkNeq(val, value any) bool {
	return checkComparison(val, value, false)
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

// Utility function to handle string-based operations
func stringOperation(data, value any, op func(string, string) bool) bool {
	strData, ok1 := data.(string)
	strValue, ok2 := value.(string)
	if !ok1 || !ok2 {
		return false
	}
	return op(strings.ToLower(strData), strings.ToLower(strValue))
}

func checkIn(data, value any) bool {
	if reflect.TypeOf(data).Kind() == reflect.Slice {
		return utils.Contains(data, value)
	}
	sl, ok := convert.ToSlice(data, value)
	if !ok {
		return false
	}
	return utils.Contains(sl, data)
}

func checkNotIn(data, value any) bool {
	return !checkIn(data, value)
}

func checkContains(data, value any) bool {
	return stringOperation(data, value, strings.Contains)
}

func checkNotContains(data, value any) bool {
	return !checkContains(data, value)
}

func checkStartsWith(data, value any) bool {
	return stringOperation(data, value, strings.HasPrefix)
}

func checkEndsWith(data, value any) bool {
	return stringOperation(data, value, strings.HasSuffix)
}

package filters

import (
	"slices"
	"strings"

	"github.com/oarkflow/filters/convert"
	"github.com/oarkflow/filters/utils"
)

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

func checkNotNull[T any](data T, filter Filter) bool {
	return any(data) != nil
}

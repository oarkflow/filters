package utils

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/oarkflow/date"
)

func Compare(a, b any) int {
	switch a := a.(type) {
	case int:
		ai := a
		bi := b.(int)
		switch {
		case ai < bi:
			return -1
		case ai > bi:
			return 1
		default:
			return 0
		}
	case string:
		isADateTime := IsValidDateTime(a)
		if !isADateTime {
			switch b := b.(type) {
			case string:
				return strings.Compare(a, b)
			default:
				return strings.Compare(a, fmt.Sprint(b))
			}
		}
		at, err := ParseTime(a)
		if err != nil {
			return 0
		}
		switch b := b.(type) {
		case string:
			bt, err := ParseTime(b)
			if err != nil {
				return 0
			}
			switch {
			case at.Before(bt):
				return -1
			case at.After(bt):
				return 1
			default:
				return 0
			}
		case time.Time:
			switch {
			case at.Before(b):
				return -1
			case at.After(b):
				return 1
			default:
				return 0
			}
		}
		return 0
	case time.Time:
		at, err := ParseTime(a)
		if err != nil {
			return 0
		}
		bt, err := ParseTime(b)
		if err != nil {
			return 0
		}
		switch {
		case at.Before(bt):
			return -1
		case at.After(bt):
			return 1
		default:
			return 0
		}
	default:
		return 0
	}
}

func IsValidDateTime(str string) bool {
	str = strings.ToUpper(str)
	_, err := date.Parse(str)
	if err != nil {
		return false
	}
	return true
}

// ParseTime convert date string to time.Time
func ParseTime(s any, layouts ...string) (t time.Time, err error) {
	var layout string
	str := ""
	if len(layouts) > 0 { // custom layout
		layout = layouts[0]
	} else {
		switch s := s.(type) {
		case time.Time:
			return s, nil
		case string:
			return date.Parse(s)
		case int:
			return time.Unix(int64(s), 0), nil
		case int64:
			return time.Unix(s, 0), nil
		}
	}
	if layout == "" {
		err = errors.New("invalid params")
		return
	}

	// has 'T' eg: "2006-01-02T15:04:05"
	if strings.ContainsRune(str, 'T') {
		layout = strings.Replace(layout, " ", "T", -1)
	}

	// eg: "2006/01/02 15:04:05"
	if strings.ContainsRune(str, '/') {
		layout = strings.Replace(layout, "-", "/", -1)
	}

	t, err = time.Parse(layout, str)
	// t, err = time.ParseInLocation(layout, s, time.Local)
	return
}

func Intersection[T any](a, b []T) []T {
	set := make(map[string]struct{})
	var intersection []T

	for _, item := range a {
		set[Serialize(item)] = struct{}{}
	}

	for _, item := range b {
		if _, exists := set[Serialize(item)]; exists {
			intersection = append(intersection, item)
		}
	}

	return intersection
}

func Union[T any](a, b []T) []T {
	set := make(map[string]struct{})
	var union []T

	for _, item := range a {
		set[Serialize(item)] = struct{}{}
		union = append(union, item)
	}

	for _, item := range b {
		if _, exists := set[Serialize(item)]; !exists {
			union = append(union, item)
		}
	}

	return union
}

func Serialize[T any](item T) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Map {
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		var builder strings.Builder
		for _, k := range keys {
			builder.WriteString(fmt.Sprintf("%s:%v|", k, v.MapIndex(k)))
		}
		return builder.String()
	}

	var builder strings.Builder
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		builder.WriteString(fmt.Sprintf("%s:%v|", t.Field(i).Name, v.Field(i).Interface()))
	}

	return builder.String()
}

// SearchDeeplyNestedSlice searches for a target slice in a nested slice.
// It returns true if any of the target slice elements are found in the nested slice
// or in any of its nested slices. Otherwise, it returns false.
func SearchDeeplyNestedSlice(nestedSlice []any, targetSlice []any) bool {
	targetMap := make(map[any]struct{})
	for _, target := range targetSlice {
		targetMap[target] = struct{}{}
	}

	for _, element := range nestedSlice {
		switch v := element.(type) {
		case []any:
			if SearchDeeplyNestedSlice(v, targetSlice) {
				return true
			}
		default:
			if _, found := targetMap[v]; found {
				return true
			}
		}
	}
	return false
}

// FlattenSlice flattens a nested slice into a single slice.
func FlattenSlice(slice []any) []any {
	var result []any
	for _, element := range slice {
		switch element := element.(type) {
		case []any:
			result = append(result, FlattenSlice(element)...)
		default:
			result = append(result, element)
		}
	}
	return result
}

// SumIntSlice sums up all the elements in a slice and returns the result.
func SumIntSlice(slice []any) int {
	var sum int
	for _, element := range slice {
		switch element := element.(type) {
		case int:
			sum += element
		case float64:
			sum += int(element)
		}
	}
	return sum
}

func Contains(sl, data any) bool {
	val := reflect.ValueOf(sl)
	// Iterate over the slice to check if data is present
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		if reflect.DeepEqual(data, item) {
			return true
		}
	}

	return false
}
func Flatten(input any) []any {
	var result []any
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			result = append(result, Flatten(v.Index(i).Interface())...)
		}
	} else {
		result = append(result, input)
	}
	return result
}

// ItemExists - Generic function to check if any item in slice1 exists in slice2
func ItemExists(slice1, slice2 any) bool {
	v1 := reflect.ValueOf(slice1)
	v2 := reflect.ValueOf(slice2)
	if v1.Kind() != reflect.Slice || v2.Kind() != reflect.Slice {
		panic("both arguments must be slices")
	}
	lookup := make(map[interface{}]bool)
	for i := 0; i < v2.Len(); i++ {
		lookup[v2.Index(i).Interface()] = true
	}
	for i := 0; i < v1.Len(); i++ {
		if lookup[v1.Index(i).Interface()] {
			return true
		}
	}
	return false
}

func BuiltinAge(params ...any) (any, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("age expects 1 argument")
	}
	left := params[0]
	t, err := ParseTime(left)
	if err != nil {
		return nil, err
	}
	return date.CalculateToNow(t), err
}

func FilterSlice(lookupData, fieldValue any) (any, error) {
	// Get the reflect.Value of lookupData and fieldValue
	lookupVal := reflect.ValueOf(lookupData)
	fieldVal := reflect.ValueOf(fieldValue)

	// Ensure both are slices
	if lookupVal.Kind() != reflect.Slice || fieldVal.Kind() != reflect.Slice {
		return fieldValue, errors.New("both lookupData and fieldValue must be slices")
	}

	// Create a map to store the elements of lookupData for quick lookup
	lookupMap := make(map[any]struct{})
	for i := 0; i < lookupVal.Len(); i++ {
		val := lookupVal.Index(i).Interface()
		lookupMap[val] = struct{}{}
	}

	// Filter fieldValue to only include elements that are in lookupMap
	filtered := reflect.MakeSlice(fieldVal.Type(), 0, 0)
	for i := 0; i < fieldVal.Len(); i++ {
		val := fieldVal.Index(i).Interface()
		if _, exists := lookupMap[val]; exists {
			filtered = reflect.Append(filtered, fieldVal.Index(i))
		}
	}

	return filtered.Interface(), nil
}

// GetSliceLength returns the length of a slice or an error if the input is not a slice
func GetSliceLength(data any) (int, error) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return 0, errors.New("provided data is not a slice")
	}
	return val.Len(), nil
}

func IsSlice(data any) bool {
	return reflect.ValueOf(data).Kind() == reflect.Slice
}

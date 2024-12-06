package filters

type FilterGroup struct {
	Operator Boolean
	Filters  []Condition
	Reverse  bool
}

func NewFilterGroup(operator Boolean, reverse bool, conditions ...Condition) *FilterGroup {
	return &FilterGroup{Operator: operator, Filters: conditions, Reverse: reverse}
}

func (group *FilterGroup) Match(data any) bool {
	return MatchGroup(data, group)
}

func ApplyGroup[T any](collection []T, filterGroups ...*FilterGroup) []T {
	var position, size = 0, len(collection)
	for i := 0; i < size; i++ {
		matches := true
		for _, group := range filterGroups {
			if !MatchGroup(collection[i], group) {
				matches = false
				break
			}
		}
		if matches {
			collection[position] = collection[i]
			position++
		}
	}

	return collection[:position]
}

func MatchGroup[T any](item T, group *FilterGroup) bool {
	switch group.Operator {
	case AND:
		matched := true
		for _, filter := range group.Filters {
			switch filter := filter.(type) {
			case *FilterGroup:
				if !MatchGroup(item, filter) {
					matched = false
					break
				}
			case *Filter:
				if !Match(item, filter) {
					matched = false
					break
				}
			}
		}
		if group.Reverse {
			return !matched
		}
		return matched
	case OR:
		matched := false
		for _, filter := range group.Filters {
			switch filter := filter.(type) {
			case *FilterGroup:
				if MatchGroup(item, filter) {
					matched = true
					break
				}
			case *Filter:
				if Match(item, filter) {
					matched = true
					break
				}
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

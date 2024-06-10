package pattern

import (
	"strconv"
	"sync"

	"github.com/oarkflow/filters"
)

type (
	Handler[T any] func(args ...any) (T, error)
	Case[T any]    struct {
		handler     Handler[T]
		args        []any
		defaultCase bool
		matchFound  bool
		err         error
		result      T
	}
	Matcher[T any] struct {
		Error  error
		values map[string]any
		cases  []Case[T]
	}
)

var (
	filterPool = sync.Pool{
		New: func() any {
			return make([]filters.Filter, 0, 10)
		},
	}
)

func (p *Case[T]) match(values map[string]any) *Case[T] {
	if p == nil {
		return nil
	}
	if p.err != nil || p.matchFound {
		return p
	}
	valueLen := len(values)
	matchesLen := len(p.args)
	if matchesLen == 0 || valueLen == 0 {
		p.err = NoValueOrCaseError
		return p
	}
	if matchesLen != valueLen {
		p.err = InvalidArgumentsError
		return p
	}
	if p.handler == nil {
		p.err = InvalidHandler
		return p
	}

	// Get a slice from the pool
	rules := filterPool.Get().([]filters.Filter)
	// Reset the slice length without reallocating
	rules = rules[:0]

	for i, match := range p.args {
		field := "f_" + strconv.Itoa(i+1)
		switch match := match.(type) {
		case string:
			switch match {
			case EXISTS:
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.NotZero,
					Value:    "",
				})
			case NOTEXISTS:
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.IsZero,
					Value:    "",
				})
			case NONE:
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.IsNull,
				})
			case ANY:
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.Equal,
					Value:    match,
				})
			default:
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.Equal,
					Value:    match,
				})
			}
		case filters.Filter:
			rules = append(rules, match)
		default:
			rules = append(rules, filters.Filter{
				Field:    field,
				Operator: filters.Equal,
				Value:    match,
			})
		}
	}
	group2 := filters.FilterGroup{
		Operator: filters.AND,
		Filters:  rules,
	}
	response := filters.MatchGroup(values, group2)

	// Put the slice back into the pool
	filterPool.Put(rules)

	if response {
		p.matchFound = true
		result, err := p.handler(p.args...)
		if err != nil {
			p.err = err
			return p
		}
		p.result = result
	}
	return p
}

func (p *Case[T]) matcherDefault() *Case[T] {
	if p == nil {
		return nil
	}
	if p.err != nil || p.matchFound {
		return p
	}
	p.matchFound = true
	result, err := p.handler(nil)
	if err != nil {
		p.err = err
		return p
	}
	p.result = result
	return p
}

const (
	ANY       = "ANY-VAL"
	NONE      = "NONE-VAL"
	EXISTS    = "EXISTS-VAL"
	NOTEXISTS = "NOT-EXISTS-VAL"
)

func Match[T any](values ...any) *Matcher[T] {
	if len(values) == 0 {
		return &Matcher[T]{
			Error: NoValueError,
		}
	}
	mp := make(map[string]any, len(values))
	for i, v := range values {
		mp["f_"+strconv.Itoa(i+1)] = v
	}

	return &Matcher[T]{values: mp}
}

func (p *Matcher[T]) Case(handler Handler[T], matches ...any) *Matcher[T] {
	p.addCase(handler, false, matches...)
	return p
}

func (p *Matcher[T]) Default(handler Handler[T]) *Matcher[T] {
	p.addCase(handler, true)
	return p
}

func (p *Matcher[T]) addCase(handler Handler[T], defaultCase bool, args ...any) *Matcher[T] {
	if p == nil {
		return nil
	}
	if p.Error != nil {
		return p
	}
	p.cases = append(p.cases, Case[T]{
		handler:     handler,
		defaultCase: defaultCase,
		args:        args,
	})
	return p
}

func (p *Matcher[T]) Result() (T, error) {
	var t T
	if p == nil {
		return t, NoMatcherError
	}
	for _, currentCase := range p.cases {
		var matchedCase *Case[T]
		if currentCase.defaultCase {
			matchedCase = currentCase.matcherDefault()
		} else {
			matchedCase = currentCase.match(p.values)
		}

		if matchedCase.err != nil {
			return t, matchedCase.err
		} else if matchedCase.matchFound {
			return matchedCase.result, nil
		}
	}
	return t, nil
}

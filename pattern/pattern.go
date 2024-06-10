package pattern

import (
	"strconv"

	"github.com/oarkflow/xid"

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
		cases  map[xid.ID]Case[T]
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

	for i, match := range p.args {
		field := "f_" + strconv.Itoa(i+1)
		switch match := match.(type) {
		case string:
			switch match {
			case EXISTS:
				if !filters.Match(values, filters.Filter{
					Field:    field,
					Operator: filters.NotZero,
					Value:    "",
				}) {
					return p
				}
			case NOTEXISTS:
				if !filters.Match(values, filters.Filter{
					Field:    field,
					Operator: filters.IsZero,
					Value:    "",
				}) {
					return p
				}
			case NONE:
				filter := filters.Filter{
					Field:    field,
					Operator: filters.IsNull,
				}
				if !filters.Match(values, filter) {
					return p
				}
			case ANY:

			default:
				filter := filters.Filter{
					Field:    field,
					Operator: filters.Equal,
					Value:    match,
				}
				if !filters.Match(values, filter) {
					return p
				}
			}
		default:
			filter := filters.Filter{
				Field:    field,
				Operator: filters.Equal,
				Value:    match,
			}
			if !filters.Match(values, filter) {
				return p
			}
		}
	}
	p.matchFound = true
	result, err := p.handler(p.args...)
	if err != nil {
		p.err = err
		return p
	}
	p.result = result
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
			cases: make(map[xid.ID]Case[T], 2),
		}
	}
	mp := make(map[string]any, len(values))
	for i, v := range values {
		mp["f_"+strconv.Itoa(i+1)] = v
	}

	return &Matcher[T]{values: mp, cases: make(map[xid.ID]Case[T], 2)}
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
	p.cases[xid.New()] = Case[T]{
		handler:     handler,
		defaultCase: defaultCase,
		args:        args,
	}
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

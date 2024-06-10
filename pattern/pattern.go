package pattern

import (
	"fmt"

	"github.com/oarkflow/xid"

	"github.com/oarkflow/filters"
)

var randomField = fmt.Sprintf("field_%s_", xid.New().String())

type (
	Handler func(args ...any) (any, error)
	Case    struct {
		handler     Handler
		args        []any
		defaultCase bool
		matchFound  bool
		err         error
		result      any
	}
)

func (p *Case) match(values map[string]any) *Case {
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
	var rules []filters.Filter

	for i, match := range p.args {
		field := fmt.Sprintf("%s%d", randomField, i+1)
		switch match := match.(type) {
		case string:
			if match == EXISTS {
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.NotZero,
					Value:    "",
				})
			} else if match == NOTEXISTS {
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.IsZero,
					Value:    "",
				})
			} else if match != ANY {
				rules = append(rules, filters.Filter{
					Field:    field,
					Operator: filters.Equal,
					Value:    match,
				})
			}
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

func (p *Case) matcherDefault() *Case {
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

type Matcher struct {
	Error  error
	values map[string]any
	cases  []Case
}

const (
	ANY       = "ANY-VAL"
	NONE      = "NONE-VAL"
	EXISTS    = "EXISTS-VAL"
	NOTEXISTS = "NOT-EXISTS-VAL"
)

func Match(values ...any) *Matcher {
	if len(values) == 0 {
		return &Matcher{
			Error: NoValueError,
		}
	}
	mp := make(map[string]any)
	for i, v := range values {
		mp[fmt.Sprintf("%s%d", randomField, i+1)] = v
	}

	return &Matcher{values: mp}
}

func (p *Matcher) Case(handler Handler, matches ...any) *Matcher {
	p.addCase(handler, false, matches...)
	return p
}

func (p *Matcher) Default(handler Handler) *Matcher {
	p.addCase(handler, true)
	return p
}

func (p *Matcher) addCase(handler Handler, defaultCase bool, args ...any) *Matcher {
	if p == nil {
		return nil
	}
	if p.Error != nil {
		return p
	}
	p.cases = append(p.cases, Case{
		handler:     handler,
		defaultCase: defaultCase,
		args:        args,
	})
	return p
}

func (p *Matcher) Result() (any, error) {
	if p == nil {
		return nil, NoMatcherError
	}
	for _, currentCase := range p.cases {
		var matchedCase *Case
		if currentCase.defaultCase {
			matchedCase = currentCase.matcherDefault()
		} else {
			matchedCase = currentCase.match(p.values)
		}

		if matchedCase.err != nil {
			return nil, matchedCase.err
		} else if matchedCase.matchFound {
			return matchedCase.result, nil
		}
	}
	return nil, nil
}

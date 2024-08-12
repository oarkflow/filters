package filters

import (
	"sort"
	"sync"
)

type Priority int

const (
	HighestPriority Priority = 1
	LowestPriority  Priority = 0
)

type PriorityRule struct {
	Rule     *Rule
	Priority int
}

type Config struct {
	Rules    []*PriorityRule
	Priority Priority
}

type GroupRule struct {
	mu     *sync.RWMutex
	Key    string          `json:"key,omitempty"`
	Rules  []*PriorityRule `json:"rules,omitempty"`
	config Config
}

func NewRuleGroup(config ...Config) *GroupRule {
	cfg := Config{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &GroupRule{
		Rules:  cfg.Rules,
		config: cfg,
		mu:     &sync.RWMutex{},
	}
}

func (r *GroupRule) AddRule(rule *Rule, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rules = append(r.Rules, &PriorityRule{
		Rule:     rule,
		Priority: priority,
	})
}

func (r *GroupRule) ApplyHighestPriority(data any, fn ...CallbackFn) (any, error) {
	return r.apply(r.sortByPriority("DESC"), data, fn...)
}

func (r *GroupRule) ApplyLowestPriority(data any, fn ...CallbackFn) (any, error) {
	return r.apply(r.sortByPriority(), data, fn...)
}

func (r *GroupRule) Apply(data any, fn ...CallbackFn) (any, error) {
	if r.config.Priority == HighestPriority {
		return r.ApplyHighestPriority(data, fn...)
	}
	return r.ApplyLowestPriority(data, fn...)
}

func (r *GroupRule) apply(sortedRules []*Rule, data any, fn ...CallbackFn) (any, error) {
	for _, rule := range sortedRules {
		response, err := rule.Validate(data, fn...)
		if err != nil {
			return nil, err
		}
		if response != nil {
			return response, nil
		}
	}
	return nil, nil
}

func (r *GroupRule) SortByPriority(direction ...string) []*Rule {
	return r.sortByPriority(direction...)
}

func (r *GroupRule) sortByPriority(direction ...string) []*Rule {
	dir := "ASC"
	if len(direction) > 0 {
		dir = direction[0]
	}
	if dir == "DESC" {
		sort.Sort(sort.Reverse(byPriority(r.Rules)))
	} else {
		sort.Sort(byPriority(r.Rules))
	}
	res := make([]*Rule, 0, len(r.Rules))
	for _, q := range r.Rules {
		res = append(res, q.Rule)
	}
	return res
}

type byPriority []*PriorityRule

func (x byPriority) Len() int           { return len(x) }
func (x byPriority) Less(i, j int) bool { return x[i].Priority < x[j].Priority }
func (x byPriority) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

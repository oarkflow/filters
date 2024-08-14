package filters

import (
	"errors"
)

type GroupRequest struct {
	Operator Boolean   `json:"operator"`
	Filters  []*Filter `json:"filters"`
	Reverse  bool      `json:"reverse"`
}

type Group struct {
	Left     *GroupRequest `json:"left"`
	Operator Boolean       `json:"operator"`
	Right    *GroupRequest `json:"right"`
	Reverse  bool          `json:"reverse"`
}

type RuleRequest struct {
	successHandler CallbackFn
	ID             string          `json:"id,omitempty"`
	ErrorMsg       string          `json:"error_msg"`
	ErrorAction    string          `json:"error_action"`
	Conditions     []*GroupRequest `json:"conditions"`
	Groups         []*Group        `json:"groups"`
	rule           *Rule
}

type ApplicationRule struct {
	Rule      *RuleRequest `json:"rule" gorm:"type:jsonb"`
	Key       string       `json:"key"`
	Level     string       `json:"level"`
	ParentKey string       `json:"parent_key"`
}

func (r *RuleRequest) SetRule(rule *Rule) {
	r.rule = rule
}

func (r *RuleRequest) Validate(data any, callbackFn ...CallbackFn) (any, error) {
	if r.rule == nil {
		return nil, errors.New("rule not provided")
	}
	return r.rule.Validate(data, callbackFn...)
}

func (applicationRule *ApplicationRule) BuildRuleFromRequest(getCondition func(string) *Filter) {
	rule := NewRule()
	if len(applicationRule.Rule.Conditions) > 0 {
		for _, cons := range applicationRule.Rule.Conditions {
			var conditions []Condition
			for _, con := range cons.Filters {
				if con.FilterKey != "" && getCondition != nil {
					fil := getCondition(con.FilterKey)
					if fil != nil {
						conditions = append(conditions, fil)
					}
				} else {
					conditions = append(conditions, con)
				}
			}
			if len(conditions) > 0 {
				rule.AddCondition(cons.Operator, cons.Reverse, conditions...)
			}
		}
	}
	if len(applicationRule.Rule.Groups) > 0 {
		var groups []Condition
		for _, cons := range applicationRule.Rule.Groups {
			var left []Condition
			if cons.Left != nil {
				for _, con := range cons.Left.Filters {
					if con.FilterKey != "" && getCondition != nil {
						fil := getCondition(con.FilterKey)
						if fil != nil {
							left = append(left, fil)
						}
					} else {
						left = append(left, con)
					}
				}
				if len(left) > 0 {
					group := NewFilterGroup(cons.Left.Operator, cons.Left.Reverse, left...)
					groups = append(groups, group)
				}
			}
			var right []Condition
			if cons.Right != nil {
				for _, con := range cons.Right.Filters {
					if con.FilterKey != "" {
						fil := getCondition(con.FilterKey)
						if fil != nil {
							right = append(right, fil)
						}
					} else {
						right = append(right, con)
					}
				}
				if len(right) > 0 {
					group := NewFilterGroup(cons.Right.Operator, cons.Right.Reverse, right...)
					groups = append(groups, group)
				}
			}
			if len(groups) > 0 {
				rule.AddCondition(cons.Operator, cons.Reverse, groups...)
			}
		}
	}
	rule.SetErrorResponse(applicationRule.Rule.ErrorMsg, applicationRule.Rule.ErrorAction)
	applicationRule.Rule.SetRule(rule)
}

package filters

import (
	"database/sql/driver"
	"encoding/json"
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

// Scan implements the Scanner interface.
// This is used to convert the JSONB type in the database into a rule.Rule struct.
func (r *RuleRequest) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}
	return json.Unmarshal(b, &r)
}

// Value implements the driver Valuer interface.
// This is used to convert the rule.Rule struct into a JSONB type in the database.
func (r RuleRequest) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func handleConditions(request *GroupRequest, getCondition func(string) *Filter) (conditions []Condition) {
	for _, con := range request.Filters {
		if con.FilterKey != "" && getCondition != nil {
			fil := getCondition(con.FilterKey)
			if fil != nil {
				conditions = append(conditions, fil)
			}
		} else {
			conditions = append(conditions, con)
		}
	}
	return
}

func handleGroupRequest(request *GroupRequest, getCondition func(string) *Filter) *FilterGroup {
	if request == nil {
		return nil
	}
	conditions := handleConditions(request, getCondition)
	if len(conditions) > 0 {
		return NewFilterGroup(request.Operator, request.Reverse, conditions...)
	}
	return nil
}

func (applicationRule *ApplicationRule) BuildRuleFromRequest(getCondition func(string) *Filter) {
	rule := NewRule()
	if len(applicationRule.Rule.Conditions) > 0 {
		for _, cons := range applicationRule.Rule.Conditions {
			conditions := handleConditions(cons, getCondition)
			if len(conditions) > 0 {
				rule.AddCondition(cons.Operator, cons.Reverse, conditions...)
			}
		}
	}
	if len(applicationRule.Rule.Groups) > 0 {
		var groups []Condition
		for _, cons := range applicationRule.Rule.Groups {
			left := handleGroupRequest(cons.Left, getCondition)
			if left != nil {
				groups = append(groups, left)
			}
			right := handleGroupRequest(cons.Right, getCondition)
			if right != nil {
				groups = append(groups, right)
			}
			if len(groups) > 0 {
				rule.AddCondition(cons.Operator, cons.Reverse, groups...)
			}
		}
	}
	rule.SetErrorResponse(applicationRule.Rule.ErrorMsg, applicationRule.Rule.ErrorAction)
	applicationRule.Rule.SetRule(rule)
}

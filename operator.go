package filters

type (
	Boolean  string
	Operator string
)

const (
	AND Boolean = "AND"
	OR  Boolean = "OR"

	Equal            Operator = "eq"
	LessThan         Operator = "lt"
	LessThanEqual    Operator = "le"
	GreaterThan      Operator = "gt"
	GreaterThanEqual Operator = "ge"
	NotEqual         Operator = "ne"
	Contains         Operator = "contains"
	NotContains      Operator = "not_contains"
	Between          Operator = "between"
	In               Operator = "in"
	StartsWith       Operator = "starts_with"
	EndsWith         Operator = "ends_with"
	NotIn            Operator = "not_in"
	NotZero          Operator = "not_zero"
	IsZero           Operator = "is_zero"
	IsNull           Operator = "is_null"
	NotNull          Operator = "not_null"
)

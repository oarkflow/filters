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
	NotContains      Operator = "ncontains"
	Between          Operator = "between"
	In               Operator = "in"
	StartsWith       Operator = "startswith"
	NotStartsWith    Operator = "nstartswith"
	EndsWith         Operator = "endswith"
	NotEndsWith      Operator = "nendswith"
	NotIn            Operator = "nin"
	NotZero          Operator = "nzero"
	IsZero           Operator = "izero"
	IsNull           Operator = "null"
	NotNull          Operator = "nnull"
)

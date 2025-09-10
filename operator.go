package filters

type (
	Boolean  string
	Operator string
)

const (
	AND Boolean = "AND"
	OR  Boolean = "OR"

	Equal                 Operator = "eq"
	LessThan              Operator = "lt"
	LessThanEqual         Operator = "le"
	GreaterThan           Operator = "gt"
	GreaterThanEqual      Operator = "ge"
	NotEqual              Operator = "ne"
	EqualCount            Operator = "eqc"
	NotEqualCount         Operator = "nec"
	GreaterThanCount      Operator = "gtc"
	LesserThanCount       Operator = "ltc"
	GreaterThanEqualCount Operator = "gec"
	LesserThanEqualCount  Operator = "lec"
	Contains              Operator = "contains"
	NotContains           Operator = "ncontains"
	Between               Operator = "between"
	Expression            Operator = "expr"
	Pattern               Operator = "pattern"
	In                    Operator = "in"
	StartsWith            Operator = "startswith"
	NotStartsWith         Operator = "nstartswith"
	EndsWith              Operator = "endswith"
	NotEndsWith           Operator = "nendswith"
	NotIn                 Operator = "nin"
	NotZero               Operator = "nzero"
	IsZero                Operator = "izero"
	IsNull                Operator = "null"
	NotNull               Operator = "nnull"
	ContainsCS            Operator = "contains_cs"
	NotContainsCS         Operator = "ncontains_cs"
	StartsWithCS          Operator = "startswith_cs"
	NotStartsWithCS       Operator = "nstartswith_cs"
	EndsWithCS            Operator = "endswith_cs"
	NotEndsWithCS         Operator = "nendswith_cs"
)

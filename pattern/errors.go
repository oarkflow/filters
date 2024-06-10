package pattern

import "errors"

var (
	NoMatcherError        = errors.New("no matcher provided")
	NoValueError          = errors.New("no values to match")
	NoValueOrCaseError    = errors.New("no values or cases to match")
	InvalidArgumentsError = errors.New("cases arguments is invalid for values length")
	InvalidHandler        = errors.New("case handler not provided")
)

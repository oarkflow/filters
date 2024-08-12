package filters

import (
	"errors"

	"github.com/oarkflow/filters/utils"
)

type Join struct {
	Left     *FilterGroup
	Right    *FilterGroup
	Operator Boolean
	Reverse  bool
}

func FilterJoin[T any](data []T, expr Join) ([]T, error) {
	if expr.Left == nil || expr.Right == nil {
		return nil, errors.New("missing left or right filter group")
	}

	leftResult := ApplyGroup(data, expr.Left)
	rightResult := ApplyGroup(data, expr.Right)
	switch expr.Operator {
	case AND:
		return utils.Intersection(leftResult, rightResult), nil
	case OR:
		return utils.Union(leftResult, rightResult), nil
	default:
		return nil, errors.New("unsupported boolean operator")
	}
}

func MatchJoin[T any](item T, expr Join) bool {
	if expr.Left == nil || expr.Right == nil {
		return false
	}
	leftResult := MatchGroup(item, expr.Left)
	rightResult := MatchGroup(item, expr.Right)
	switch expr.Operator {
	case AND:
		if expr.Reverse {
			return !(leftResult && rightResult)
		}
		return leftResult && rightResult
	case OR:
		if expr.Reverse {
			return !(leftResult || rightResult)
		}
		return leftResult || rightResult
	default:
		return false
	}
}

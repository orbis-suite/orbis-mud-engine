package conditions

import (
	"fmt"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/expressions"
)

type ExpressionTrue struct {
	Expression expressions.Expression
}

var _ entities.Condition = &ExpressionTrue{}

func (et *ExpressionTrue) Id() entities.ConditionType {
	return entities.ConditionExpressionTrue
}

func (et *ExpressionTrue) Check(ev *entities.Event) (bool, error) {
	exprResult, err := et.Expression.Eval(ev)
	if err != nil {
		return false, fmt.Errorf("could not evaluate expression in condition: %w", err)
	}

	if exprResult.K != models.KindBool {
		return false, fmt.Errorf("expression in condition did not evaluate to boolean")
	}

	return exprResult.B, nil
}

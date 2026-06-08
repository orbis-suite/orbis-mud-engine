package conditions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Or struct {
	Left  entities.Condition
	Right entities.Condition
}

var _ entities.Condition = &Or{}

func (o *Or) Id() entities.ConditionType {
	return entities.ConditionOr
}

func (o *Or) Check(ev *entities.Event) (bool, error) {
	lCheck, err := o.Left.Check(ev)
	if err != nil {
		return false, fmt.Errorf("or left condition: %w", err)
	}

	rCheck, err := o.Right.Check(ev)
	if err != nil {
		return false, fmt.Errorf("or right condition: %w", err)
	}

	return (lCheck || rCheck), nil
}

package conditions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Not struct {
	Cond entities.Condition
}

var _ entities.Condition = &Not{}

func (n *Not) Id() entities.ConditionType {
	return entities.ConditionNot
}

func (n *Not) Check(ev *entities.Event) (bool, error) {
	check, err := n.Cond.Check(ev)

	if err != nil {
		return false, fmt.Errorf("not condition: %w", err)
	}

	return !check, nil
}

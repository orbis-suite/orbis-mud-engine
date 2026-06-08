package conditions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type IsPresent struct {
	EventRole entities.EventRole
}

var _ entities.Condition = &IsPresent{}

func (h *IsPresent) Id() entities.ConditionType {
	return entities.ConditionIsPresent
}

func (ip *IsPresent) Check(ev *entities.Event) (bool, error) {
	var e *entities.Entity
	switch ip.EventRole {
	case entities.EventRoleSource:
		e = ev.Source
	case entities.EventRoleInstrument:
		e = ev.Instrument
	case entities.EventRoleTarget:
		e = ev.Target
	default:
		return false, fmt.Errorf("invalid role '%s' for is present condition", ip.EventRole.String())
	}

	return (e != nil), nil
}

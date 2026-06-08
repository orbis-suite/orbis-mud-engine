package conditions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type EventRolesEqual struct {
	EventRole1 entities.EventRole
	EventRole2 entities.EventRole
}

var _ entities.Condition = &EventRolesEqual{}

func (ier *EventRolesEqual) Id() entities.ConditionType {
	return entities.ConditionEventRolesEqual
}

func (ere *EventRolesEqual) Check(ev *entities.Event) (bool, error) {
	var e1 *entities.Entity
	switch ere.EventRole1 {
	case entities.EventRoleSource:
		e1 = ev.Source
	case entities.EventRoleInstrument:
		e1 = ev.Instrument
	case entities.EventRoleTarget:
		e1 = ev.Target
	default:
		return false, fmt.Errorf("invalid role '%s' for event roles equal condition", ere.EventRole1.String())
	}

	var e2 *entities.Entity
	switch ere.EventRole2 {
	case entities.EventRoleSource:
		e2 = ev.Source
	case entities.EventRoleInstrument:
		e2 = ev.Instrument
	case entities.EventRoleTarget:
		e2 = ev.Target
	default:
		return false, fmt.Errorf("invalid role '%s' for event roles equal condition", ere.EventRole2.String())
	}

	if e1 == nil && e2 == nil {
		return true, nil
	}

	return e1 == e2, nil
}

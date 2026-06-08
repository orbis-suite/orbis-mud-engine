package conditions

import (
	"fmt"
	"slices"

	"example.com/mud/world/entities"
)

type HasTag struct {
	EventRole entities.EventRole
	Tag       string
}

var _ entities.Condition = &HasTag{}

func (h *HasTag) Id() entities.ConditionType {
	return entities.ConditionHasTag
}

func (h *HasTag) Check(ev *entities.Event) (bool, error) {
	var e *entities.Entity
	switch h.EventRole {
	case entities.EventRoleSource:
		e = ev.Source
	case entities.EventRoleInstrument:
		e = ev.Instrument
	case entities.EventRoleTarget:
		e = ev.Target
	default:
		return false, fmt.Errorf("invalid role '%s' for has tag condition", h.EventRole.String())
	}

	if e == nil {
		return false, nil
	}

	return slices.Contains(e.Tags, h.Tag), nil
}

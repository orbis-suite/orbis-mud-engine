package conditions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type HasChild struct {
	ParentRole    entities.EventRole
	ComponentType entities.ComponentType
	ChildRole     entities.EventRole
}

var _ entities.Condition = &HasChild{}

func (h *HasChild) Id() entities.ConditionType {
	return entities.ConditionHasChild
}

func (h *HasChild) Check(ev *entities.Event) (bool, error) {
	var parent *entities.Entity
	switch h.ParentRole {
	case entities.EventRoleSource:
		parent = ev.Source
	case entities.EventRoleInstrument:
		parent = ev.Instrument
	case entities.EventRoleTarget:
		parent = ev.Target
	case entities.EventRoleRoom:
		parent = ev.Room
	default:
		return false, fmt.Errorf("invalid parent role '%s' for has child condition", h.ParentRole.String())
	}

	component, err := parent.RequireComponentWithChildren(h.ComponentType)
	if err != nil {
		return false, fmt.Errorf("error executing has child condition: %w", err)
	}

	var child *entities.Entity
	switch h.ChildRole {
	case entities.EventRoleSource:
		child = ev.Source
	case entities.EventRoleInstrument:
		child = ev.Instrument
	case entities.EventRoleTarget:
		child = ev.Target
	case entities.EventRoleRoom:
		child = ev.Room
	default:
		return false, fmt.Errorf("invalid child role '%s' for has child condition", h.ChildRole.String())
	}

	return component.GetChildren().HasChild(child), nil
}

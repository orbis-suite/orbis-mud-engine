package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Move struct {
	RoleObject      entities.EventRole
	RoleDestination entities.EventRole
	ComponentType   entities.ComponentType
}

var _ entities.Action = &Move{}

func (m *Move) Execute(ev *entities.Event) error {
	origin, err := ev.GetRole(m.RoleObject)
	if err != nil {
		return fmt.Errorf("move execute object to move: %w", err)
	}

	destination, err := ev.GetRole(m.RoleDestination)
	if err != nil {
		return fmt.Errorf("move execute destination: %w", err)
	}

	component, err := destination.RequireComponentWithChildren(m.ComponentType)
	if err != nil {
		return fmt.Errorf("error executing copy action: %w", err)
	}

	// remove entity from old parent
	oldParent := origin.Parent
	oldParent.RemoveChild(origin)

	// add entity to new parent
	component.AddChild(origin)

	return nil
}

package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Copy struct {
	EntityId      string
	EventRole     entities.EventRole
	ComponentType entities.ComponentType
}

var _ entities.Action = &Copy{}

func (c *Copy) Execute(ev *entities.Event) error {
	if ev.EntitiesById == nil {
		return fmt.Errorf("entities by id map in event may not be nil for copy action")
	}

	recipient, err := ev.GetRole(c.EventRole)
	if err != nil {
		return fmt.Errorf("Copy execute: %w", err)
	}

	component, err := recipient.RequireComponentWithChildren(c.ComponentType)
	if err != nil {
		return fmt.Errorf("Copy execute: %w", err)
	}

	// check id of entity to copy exists
	entityToCopy, ok := ev.EntitiesById[c.EntityId]
	if !ok {
		return fmt.Errorf("Copy execute: entity '%s' doesn't exist", c.EntityId)
	}

	component.AddChild(
		entityToCopy.Copy(component),
	)

	return nil
}

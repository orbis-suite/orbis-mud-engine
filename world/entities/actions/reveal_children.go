package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type RevealChildren struct {
	Role          entities.EventRole
	ComponentType entities.ComponentType
	Reveal        bool
}

var _ entities.Action = &RevealChildren{}

func (r *RevealChildren) Execute(ev *entities.Event) error {
	var role *entities.Entity
	switch r.Role {
	case entities.EventRoleSource:
		role = ev.Source
	case entities.EventRoleInstrument:
		role = ev.Instrument
	case entities.EventRoleTarget:
		role = ev.Target
	default:
		return fmt.Errorf("invalid origin role '%s' for reveal children action", r.Role.String())
	}

	if role == nil {
		return fmt.Errorf("role '%s' is empty for reveal children event", r.Role)
	}

	component, err := role.RequireComponentWithChildren(r.ComponentType)
	if err != nil {
		return fmt.Errorf("error executing reveal children action: %w", err)
	}

	component.GetChildren().SetRevealed(r.Reveal)

	return nil
}

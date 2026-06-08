package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Destroy struct {
	Role entities.EventRole
}

var _ entities.Action = &Destroy{}

func (d *Destroy) Execute(ev *entities.Event) error {

	var role *entities.Entity
	switch d.Role {
	case entities.EventRoleSource:
		role = ev.Source
	case entities.EventRoleInstrument:
		role = ev.Instrument
	case entities.EventRoleTarget:
		role = ev.Target
	case entities.EventRoleRoom:
		role = ev.Room
	default:
		return fmt.Errorf("invalid role '%s' for destroy action", d.Role.String())
	}

	if role == nil {
		return fmt.Errorf("role '%s' is empty for destroy event", d.Role)
	}

	// remove role from parent (is this enough for garbage collection to kick in?)
	role.Parent.RemoveChild(role)

	return nil
}

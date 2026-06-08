package actions

import (
	"fmt"

	"example.com/mud/world/entities"
	"example.com/mud/world/entities/expressions"
)

type SetField struct {
	Role       entities.EventRole
	Field      string
	Expression expressions.Expression
}

var _ entities.Action = &SetField{}

func (sf *SetField) Execute(ev *entities.Event) error {
	var e *entities.Entity
	switch sf.Role {
	case entities.EventRoleSource:
		e = ev.Source
	case entities.EventRoleInstrument:
		e = ev.Instrument
	case entities.EventRoleTarget:
		e = ev.Target
	default:
		return fmt.Errorf("invalid role '%s' for SetField action", sf.Role)
	}

	if e == nil {
		return fmt.Errorf("role '%s' is empty for SetField event", sf.Role)
	}

	exprResult, err := sf.Expression.Eval(ev)
	if err != nil {
		return fmt.Errorf("could not evaluate expression in SetField: %w", err)
	}

	return e.SetField(sf.Field, exprResult)
}

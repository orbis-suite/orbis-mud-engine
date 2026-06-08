package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Print struct {
	Text      string
	EventRole entities.EventRole
}

var _ entities.Action = &Print{}

func (p *Print) Execute(ev *entities.Event) error {
	if ev.Publisher == nil {
		return fmt.Errorf("publisher in event may not be nil for print action")
	}

	recipient, err := ev.GetRole(p.EventRole)
	if err != nil {
		return fmt.Errorf("Copy execute: %w", err)
	}

	message, err := entities.FormatEventMessage(p.Text, ev)
	if err != nil {
		return err
	}

	ev.Publisher.PublishTo(ev.Room, recipient, message)

	return nil
}

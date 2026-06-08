package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Publish struct {
	Text string
}

var _ entities.Action = &Publish{}

func (p *Publish) Execute(ev *entities.Event) error {
	if ev.Publisher == nil {
		return fmt.Errorf("publisher in event may not be nil for publish action")
	}

	message, err := entities.FormatEventMessage(p.Text, ev)
	if err != nil {
		return err
	}

	ev.Publisher.Publish(ev.Room, message, []*entities.Entity{ev.Source, ev.Instrument, ev.Target})

	return nil
}

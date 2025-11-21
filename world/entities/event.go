package entities

import (
	"fmt"

	"example.com/mud/utils"
	"example.com/mud/world/scheduler"
)

// allows us to use the bus without tightly coupling a
// specific publisher to our world model
type Publisher interface {
	Publish(room *Entity, text string, exclude []*Entity)
	PublishTo(recipient *Entity, text string)
	Move(toRoom *Entity, player *Entity)
}

type Scheduler interface {
	Add(job *scheduler.Job)
}

type Event struct {
	Type              string
	Publisher         Publisher
	Scheduler         Scheduler
	EntitiesById      map[string]*Entity
	CommandParameters map[string]string
	Room              *Entity
	Source            *Entity
	Instrument        *Entity
	Target            *Entity
}

func (e *Event) GetRole(role EventRole) (*Entity, error) {
	var roleEntity *Entity

	switch role {
	case EventRoleSource:
		roleEntity = e.Source
	case EventRoleInstrument:
		roleEntity = e.Instrument
	case EventRoleTarget:
		roleEntity = e.Target
	case EventRoleRoom:
		roleEntity = e.Room
	default:
		return nil, fmt.Errorf("invalid role '%s'", role.String())
	}

	return roleEntity, nil
}

func (e *Event) RequireRole(role EventRole) (*Entity, error) {
	entity, err := e.GetRole(role)
	if err != nil {
		return nil, fmt.Errorf("require role: %w", err)
	}

	if entity == nil {
		return nil, fmt.Errorf("role %s for event is nil", role.String())
	}

	return entity, nil
}

type Rule struct {
	When []Condition
	Then []Action
}

func FormatEventMessage(message string, ev *Event) (string, error) {
	eventMap := make(map[string]string, 4)

	if ev.Source != nil {
		role := EventRoleSource.String()
		eventMap[role] = ev.Source.Name
	}

	if ev.Instrument != nil {
		role := EventRoleInstrument.String()
		eventMap[role] = ev.Instrument.Name
	}

	if ev.Target != nil {
		role := EventRoleTarget.String()
		eventMap[role] = ev.Target.Name
	}

	message, err := utils.FormatText(message, eventMap)
	if err != nil {
		return "", fmt.Errorf("format event message: %w", err)
	}

	return message, nil
}

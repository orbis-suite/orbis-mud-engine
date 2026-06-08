package entities

import (
	"fmt"
	"strconv"

	"example.com/mud/models"
	"example.com/mud/utils"
	"example.com/mud/world/scheduler"
)

// allows us to use the bus without tightly coupling a
// specific publisher to our world model
type Publisher interface {
	Publish(room *Entity, text string, exclude []*Entity)
	PublishTo(room *Entity, recipient *Entity, text string)
}

type Scheduler interface {
	Add(job *scheduler.Job)
}

type Event struct {
	Type         string
	Publisher    Publisher
	Scheduler    Scheduler
	EntitiesById map[string]*Entity
	Room         *Entity
	Source       *Entity
	Instrument   *Entity
	Target       *Entity
	Message      string
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

	if roleEntity == nil {
		return nil, fmt.Errorf("role %s for event is nil", role.String())
	}

	return roleEntity, nil
}

type Rule struct {
	When []Condition
	Then []Action
}

func FormatEventMessage(message string, ev *Event) (string, error) {
	eventMap := make(map[string]string, 4)

	eventMap[EventRoleMessage.String()] = ev.Message

	if ev.Source != nil {
		role := EventRoleSource.String()
		eventMap[role] = ev.Source.Name
		eventMap[fmt.Sprintf("%s.description", role)] = ev.Source.Description

		for f, v := range ev.Source.Fields {
			switch v.K {
			case models.KindBool:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatBool(v.B)
			case models.KindInt:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatInt(int64(v.I), 10)
			case models.KindString:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = v.S
			}
		}
	}

	if ev.Instrument != nil {
		role := EventRoleInstrument.String()
		eventMap[role] = ev.Instrument.Name
		eventMap[fmt.Sprintf("%s.description", role)] = ev.Instrument.Description

		for f, v := range ev.Instrument.Fields {
			switch v.K {
			case models.KindBool:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatBool(v.B)
			case models.KindInt:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatInt(int64(v.I), 10)
			case models.KindString:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = v.S
			}
		}
	}

	if ev.Target != nil {
		role := EventRoleTarget.String()
		eventMap[role] = ev.Target.Name
		eventMap[fmt.Sprintf("%s.description", role)] = ev.Target.Description

		for f, v := range ev.Target.Fields {
			switch v.K {
			case models.KindBool:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatBool(v.B)
			case models.KindInt:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = strconv.FormatInt(int64(v.I), 10)
			case models.KindString:
				eventMap[fmt.Sprintf("%s.%s", role, f)] = v.S
			}
		}
	}

	if ev.Message != "" {
		eventMap[EventRoleMessageString] = ev.Message
	}

	message, err := utils.FormatText(message, eventMap)
	if err != nil {
		return "", fmt.Errorf("format event message: %w", err)
	}

	return message, nil
}

package components

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Eventful struct {
	Rules map[string][]*entities.Rule
}

var _ entities.Component = &Eventful{}

func (e *Eventful) Id() entities.ComponentType {
	return entities.ComponentEventful
}

func (e *Eventful) Copy() entities.Component {
	return &Eventful{
		Rules: e.Rules,
	}
}

func (c *Eventful) OnEvent(ev *entities.Event) (bool, error) {
	for _, r := range c.Rules[ev.Type] {
		match, err := matchWhen(r.When, ev)
		if err != nil {
			return false, err
		}

		if match {
			for _, a := range r.Then {
				if err := a.Execute(ev); err != nil {
					return false, fmt.Errorf("error executing action: %w", err)
				}
			}
			// only match on first match, return after
			return true, nil
		}
	}
	return false, nil
}

func (c *Eventful) AddRule(eventType string, rule *entities.Rule) {
	c.Rules[eventType] = append(c.Rules[eventType], rule)
}

func matchWhen(conditions []entities.Condition, ev *entities.Event) (bool, error) {
	ret := true

	for _, c := range conditions {
		check, err := c.Check(ev)
		if err != nil {
			return false, fmt.Errorf("error checking conditions: %w", err)
		}

		ret = ret && check

		if !ret {
			break
		}
	}

	return ret, nil
}

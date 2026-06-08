package actions

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Conditional struct {
	RuleChain []*entities.Rule
}

var _ entities.Action = &Conditional{}

func (c *Conditional) Execute(ev *entities.Event) error {
	// evaluate each rule in chain, in order
	for _, rule := range c.RuleChain {
		runActions := true

		for _, condition := range rule.When {
			checkValue, err := condition.Check(ev)
			if err != nil {
				return fmt.Errorf("conditional checking conditions: %w", err)
			}

			runActions = runActions && checkValue

			// short circuit after first false
			if !runActions {
				break
			}
		}

		if runActions {
			for _, action := range rule.Then {
				err := action.Execute(ev)
				if err != nil {
					return fmt.Errorf("conditional running action: %w", err)
				}
			}
			return nil
		}
	}

	return nil
}

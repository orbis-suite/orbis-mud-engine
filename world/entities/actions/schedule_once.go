package actions

import (
	"time"

	"example.com/mud/world/entities"
	"example.com/mud/world/scheduler"
)

type ScheduleOnce struct {
	Nanoseconds time.Duration
	Actions     []entities.Action
}

var _ entities.Action = &ScheduleOnce{}

func (c *ScheduleOnce) Execute(ev *entities.Event) error {
	ev.Scheduler.Add(&scheduler.Job{
		NextRun: time.Now().Add(c.Nanoseconds),
		RunFunc: func() {
			for _, a := range c.Actions {
				err := a.Execute(ev)
				if err != nil {
					break
				}
			}
		},
	})

	return nil
}

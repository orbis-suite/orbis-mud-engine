package actions

import (
	"time"

	"example.com/mud/world/entities"
	"example.com/mud/world/scheduler"
)

type ScheduleRepeating struct {
	Nanoseconds time.Duration
	Rule        *entities.Rule
}

var _ entities.Action = &ScheduleRepeating{}

func (sr *ScheduleRepeating) Execute(ev *entities.Event) error {
	// defining a schedule function before instantiating it allows recursion inside the schedule function
	var schedule func(next time.Time)

	schedule = func(next time.Time) {
		ev.Scheduler.Add(&scheduler.Job{
			NextRun: next,
			RunFunc: func() {
				for _, condition := range sr.Rule.When {
					ok, err := condition.Check(ev)
					if err != nil || !ok {
						// stop rescheduling if any condition fails or errors
						return
					}
				}

				for _, action := range sr.Rule.Then {
					if err := action.Execute(ev); err != nil {
						// stop on first action error
						return
					}
				}

				// if we reach this far, reschedule again.
				schedule(next.Add(sr.Nanoseconds))
			},
		})
	}

	// kick off the first run
	schedule(time.Now().Add(sr.Nanoseconds))
	return nil
}

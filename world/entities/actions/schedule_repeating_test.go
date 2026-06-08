// actions/schedule_repeating_test.go
package actions

import (
	"errors"
	"testing"
	"time"

	"example.com/mud/mocks"
	"example.com/mud/world/entities"
	"example.com/mud/world/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduleRepeating_Execute(t *testing.T) {
	t.Parallel()

	type fields struct {
		delay time.Duration
		rule  *entities.Rule
	}
	type tc struct {
		name  string
		build func(t *testing.T) (fields, *entities.Event, func(t *testing.T))
	}

	const tolerance = 5 * time.Second

	newScheduler := func(t *testing.T, captured *[]*scheduler.Job) *mocks.MockScheduler {
		t.Helper()
		ms := new(mocks.MockScheduler)
		ms.
			On("Add", mock.MatchedBy(func(j *scheduler.Job) bool {
				*captured = append(*captured, j)
				return j != nil
			})).
			Return().
			Maybe()
		return ms
	}
	newCond := func(t *testing.T) *mocks.MockCondition {
		t.Helper()
		return new(mocks.MockCondition)
	}
	newAction := func(t *testing.T) *mocks.MockAction {
		t.Helper()
		return new(mocks.MockAction)
	}

	cases := []tc{
		{
			name: "repeats when all conditions true and actions succeed",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				var jobs []*scheduler.Job
				mSched := newScheduler(t, &jobs)

				delay := 25 * time.Millisecond
				start := time.Now()

				c1 := newCond(t)
				a1 := newAction(t)
				a2 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				c1.On("Check", ev).Return(true, nil).Once()
				a1.On("Execute", ev).Return(nil).Once()
				a2.On("Execute", ev).Return(nil).Once()

				c1.On("Check", ev).Return(true, nil).Once()
				a1.On("Execute", ev).Return(nil).Once()
				a2.On("Execute", ev).Return(nil).Once()

				rule := &entities.Rule{When: []entities.Condition{c1}, Then: []entities.Action{a1, a2}}

				verify := func(t *testing.T) {
					require.GreaterOrEqual(t, len(jobs), 1)
					first := jobs[0]
					assert.WithinDuration(t, start.Add(delay), first.NextRun, tolerance)

					first.RunFunc()

					require.GreaterOrEqual(t, len(jobs), 2)
					second := jobs[1]
					assert.WithinDuration(t, first.NextRun.Add(delay), second.NextRun, tolerance)

					second.RunFunc()
					require.GreaterOrEqual(t, len(jobs), 3)

					c1.AssertExpectations(t)
					a1.AssertExpectations(t)
					a2.AssertExpectations(t)
					mSched.AssertExpectations(t)
				}

				return fields{delay: delay, rule: rule}, ev, verify
			},
		},
		{
			name: "does not reschedule when any condition is false",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				var jobs []*scheduler.Job
				mSched := newScheduler(t, &jobs)

				delay := 10 * time.Millisecond
				start := time.Now()

				c1 := newCond(t)
				c2 := newCond(t)
				a1 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				c1.On("Check", ev).Return(true, nil).Once()
				c2.On("Check", ev).Return(false, nil).Once()

				rule := &entities.Rule{When: []entities.Condition{c1, c2}, Then: []entities.Action{a1}}

				verify := func(t *testing.T) {
					require.Equal(t, 1, len(jobs))
					first := jobs[0]
					assert.WithinDuration(t, start.Add(delay), first.NextRun, tolerance)

					first.RunFunc()
					assert.Equal(t, 1, len(jobs))

					c1.AssertExpectations(t)
					c2.AssertExpectations(t)
					a1.AssertNotCalled(t, "Execute", ev)
					mSched.AssertExpectations(t)
				}

				return fields{delay: delay, rule: rule}, ev, verify
			},
		},
		{
			name: "does not reschedule when a condition errors",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				var jobs []*scheduler.Job
				mSched := newScheduler(t, &jobs)

				delay := 15 * time.Millisecond

				c1 := newCond(t)
				a1 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				c1.On("Check", ev).Return(false, errors.New("oops")).Once()

				rule := &entities.Rule{When: []entities.Condition{c1}, Then: []entities.Action{a1}}

				verify := func(t *testing.T) {
					require.Equal(t, 1, len(jobs))
					jobs[0].RunFunc()
					assert.Equal(t, 1, len(jobs))
					a1.AssertNotCalled(t, "Execute", ev)
					c1.AssertExpectations(t)
					mSched.AssertExpectations(t)
				}

				return fields{delay: delay, rule: rule}, ev, verify
			},
		},
		{
			name: "stops on first action error and does not reschedule",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				var jobs []*scheduler.Job
				mSched := newScheduler(t, &jobs)

				delay := 20 * time.Millisecond

				c1 := newCond(t)
				a1 := newAction(t)
				a2 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				c1.On("Check", ev).Return(true, nil).Once()
				a1.On("Execute", ev).Return(errors.New("kaboom")).Once()

				rule := &entities.Rule{When: []entities.Condition{c1}, Then: []entities.Action{a1, a2}}

				verify := func(t *testing.T) {
					require.Equal(t, 1, len(jobs))
					jobs[0].RunFunc()

					assert.Equal(t, 1, len(jobs))
					c1.AssertExpectations(t)
					a1.AssertExpectations(t)
					a2.AssertNotCalled(t, "Execute", ev)
					mSched.AssertExpectations(t)
				}

				return fields{delay: delay, rule: rule}, ev, verify
			},
		},
		{
			name: "checks all conditions until a failing one, then stops (no actions, no reschedule)",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				var jobs []*scheduler.Job
				mSched := newScheduler(t, &jobs)

				delay := 30 * time.Millisecond

				c1 := newCond(t)
				c2 := newCond(t)
				c3 := newCond(t)
				a1 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				c1.On("Check", ev).Return(true, nil).Once()
				c2.On("Check", ev).Return(false, nil).Once()

				rule := &entities.Rule{When: []entities.Condition{c1, c2, c3}, Then: []entities.Action{a1}}

				verify := func(t *testing.T) {
					require.Equal(t, 1, len(jobs))
					jobs[0].RunFunc()

					assert.Equal(t, 1, len(jobs))
					c1.AssertExpectations(t)
					c2.AssertExpectations(t)
					c3.AssertNotCalled(t, "Check", ev)
					a1.AssertNotCalled(t, "Execute", ev)
					mSched.AssertExpectations(t)
				}

				return fields{delay: delay, rule: rule}, ev, verify
			},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			t.Parallel()

			f, ev, verify := cse.build(t)

			act := &ScheduleRepeating{
				Nanoseconds: f.delay,
				Rule:        f.rule,
			}

			err := act.Execute(ev)
			require.NoError(t, err, "Execute should not return an error")

			verify(t)
		})
	}
}

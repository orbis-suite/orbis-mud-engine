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

func TestScheduleOnce_Execute(t *testing.T) {
	t.Parallel()

	type fields struct {
		delay   time.Duration
		actions []entities.Action
	}
	type want struct {
		// approxNextRun is compared to now+delay within tolerance
		approxNextRun bool
		stopOnErrorAt int // -1 means no error expected; otherwise index where error should stop subsequent actions
	}
	type tc struct {
		name  string
		build func(t *testing.T) (fields, *entities.Event, func(t *testing.T))
		want  want
	}

	newAction := func(t *testing.T) *mocks.MockAction {
		t.Helper()
		return new(mocks.MockAction)
	}
	newScheduler := func(t *testing.T) *mocks.MockScheduler {
		t.Helper()
		return new(mocks.MockScheduler)
	}

	const tolerance = 5 * time.Second

	cases := []tc{
		{
			name: "runs all actions when none error",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				mSched := newScheduler(t)
				a1 := newAction(t)
				a2 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				// Capture job so we can run it later
				var captured *scheduler.Job
				now := time.Now()
				delay := 50 * time.Millisecond

				mSched.
					On("Add", mock.MatchedBy(func(j *scheduler.Job) bool {
						captured = j
						// sanity check here; precise assertion happens later
						return j != nil && j.NextRun.After(now)
					})).
					Once()

				// Both actions succeed
				a1.On("Execute", ev).Return(nil).Once()
				a2.On("Execute", ev).Return(nil).Once()

				verify := func(t *testing.T) {
					require.NotNil(t, captured, "Scheduler.Add should be called with a job")
					assert.WithinDuration(t, now.Add(delay), captured.NextRun, tolerance)
					// Run scheduled task
					captured.RunFunc()
					a1.AssertExpectations(t)
					a2.AssertExpectations(t)
					mSched.AssertExpectations(t)
				}

				return fields{
					delay:   delay,
					actions: []entities.Action{a1, a2},
				}, ev, verify
			},
			want: want{approxNextRun: true, stopOnErrorAt: -1},
		},
		{
			name: "stops on first action error (first action errors)",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				mSched := newScheduler(t)
				a1 := newAction(t)
				a2 := newAction(t)
				a3 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				var captured *scheduler.Job
				now := time.Now()
				delay := 1 * time.Millisecond

				mSched.
					On("Add", mock.MatchedBy(func(j *scheduler.Job) bool {
						captured = j
						return j != nil
					})).
					Once()

				a1.On("Execute", ev).Return(errors.New("boom")).Once()
				// a2 and a3 should NOT be called at all

				verify := func(t *testing.T) {
					require.NotNil(t, captured)
					assert.WithinDuration(t, now.Add(delay), captured.NextRun, tolerance)
					captured.RunFunc()

					a1.AssertExpectations(t)
					a2.AssertNotCalled(t, "Execute", ev)
					a3.AssertNotCalled(t, "Execute", ev)
					mSched.AssertExpectations(t)
				}

				return fields{
					delay:   delay,
					actions: []entities.Action{a1, a2, a3},
				}, ev, verify
			},
			want: want{approxNextRun: true, stopOnErrorAt: 0},
		},
		{
			name: "stops on second action error (first ok, second errors, third skipped)",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				mSched := newScheduler(t)
				a1 := newAction(t)
				a2 := newAction(t)
				a3 := newAction(t)

				ev := &entities.Event{Scheduler: mSched}

				var captured *scheduler.Job
				now := time.Now()
				delay := 2 * time.Millisecond

				mSched.
					On("Add", mock.MatchedBy(func(j *scheduler.Job) bool {
						captured = j
						return j != nil
					})).
					Once()

				a1.On("Execute", ev).Return(nil).Once()
				a2.On("Execute", ev).Return(errors.New("kaboom")).Once()
				// a3 should not run
				// (ScheduleOnce breaks on first error)

				verify := func(t *testing.T) {
					require.NotNil(t, captured)
					assert.WithinDuration(t, now.Add(delay), captured.NextRun, tolerance)
					captured.RunFunc()

					a1.AssertExpectations(t)
					a2.AssertExpectations(t)
					a3.AssertNotCalled(t, "Execute", ev)
					mSched.AssertExpectations(t)
				}

				return fields{
					delay:   delay,
					actions: []entities.Action{a1, a2, a3},
				}, ev, verify
			},
			want: want{approxNextRun: true, stopOnErrorAt: 1},
		},
		{
			name: "no actions: schedules but RunFunc is a no-op",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				mSched := newScheduler(t)
				ev := &entities.Event{Scheduler: mSched}

				var captured *scheduler.Job
				now := time.Now()
				delay := 5 * time.Millisecond

				mSched.
					On("Add", mock.MatchedBy(func(j *scheduler.Job) bool {
						captured = j
						return j != nil
					})).
					Once()

				verify := func(t *testing.T) {
					require.NotNil(t, captured)
					assert.WithinDuration(t, now.Add(delay), captured.NextRun, tolerance)
					// Should not panic or call anything
					captured.RunFunc()
					mSched.AssertExpectations(t)
				}

				return fields{
					delay:   delay,
					actions: nil,
				}, ev, verify
			},
			want: want{approxNextRun: true, stopOnErrorAt: -1},
		},
	}

	for _, cse := range cases {
		cse := cse
		t.Run(cse.name, func(t *testing.T) {
			t.Parallel()

			f, ev, verify := cse.build(t)

			act := &ScheduleOnce{
				Nanoseconds: f.delay,
				Actions:     f.actions,
			}

			err := act.Execute(ev)
			require.NoError(t, err, "Execute should not return an error")

			verify(t)
		})
	}
}

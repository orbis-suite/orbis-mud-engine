// actions/conditional_test.go
package actions

import (
	"errors"
	"testing"

	"example.com/mud/mocks"
	"example.com/mud/world/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditional_Execute(t *testing.T) {
	t.Parallel()

	type fields struct {
		rules []*entities.Rule
	}

	type want struct {
		errContains string
	}

	type tc struct {
		name  string
		build func(t *testing.T) (fields, *entities.Event, // fields + event
			// also return a cleanup/assert func to verify call behavior
			func(t *testing.T))
		want want
	}

	makeAction := func(t *testing.T) *mocks.MockAction {
		t.Helper()
		return new(mocks.MockAction)
	}
	makeCond := func(t *testing.T) *mocks.MockCondition {
		t.Helper()
		return new(mocks.MockCondition)
	}

	cases := []tc{
		{
			name: "happy path: first rule matches; executes all actions in that rule; stops processing later rules",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}

				// Rule 1: two conditions -> true,true ; two actions -> both execute
				c1a := makeCond(t)
				c1b := makeCond(t)
				a1a := makeAction(t)
				a1b := makeAction(t)

				c1a.On("Check", ev).Return(true, nil).Once()
				// Should be called because previous true
				c1b.On("Check", ev).Return(true, nil).Once()

				// Both actions should run
				a1a.On("Execute", ev).Return(nil).Once()
				a1b.On("Execute", ev).Return(nil).Once()

				r1 := &entities.Rule{
					When: []entities.Condition{c1a, c1b},
					Then: []entities.Action{a1a, a1b},
				}

				// Rule 2 should never be evaluated
				c2 := makeCond(t)
				a2 := makeAction(t)
				r2 := &entities.Rule{
					When: []entities.Condition{c2},
					Then: []entities.Action{a2},
				}

				return fields{rules: []*entities.Rule{r1, r2}}, ev, func(t *testing.T) {
					c1a.AssertExpectations(t)
					c1b.AssertExpectations(t)
					a1a.AssertExpectations(t)
					a1b.AssertExpectations(t)

					// Not called at all
					c2.AssertNotCalled(t, "Check", ev)
					a2.AssertNotCalled(t, "Execute", ev)
				}
			},
			want: want{errContains: ""},
		},
		{
			name: "happy path: first rule short-circuits on false; second rule matches and runs its actions",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}

				// Rule 1: c1a -> false, c1b should NOT be called (short-circuit)
				c1a := makeCond(t)
				c1b := makeCond(t)
				a1 := makeAction(t)

				c1a.On("Check", ev).Return(false, nil).Once()
				// c1b not called; a1 not called

				r1 := &entities.Rule{
					When: []entities.Condition{c1a, c1b},
					Then: []entities.Action{a1},
				}

				// Rule 2: c2 -> true; actions run
				c2 := makeCond(t)
				a2a := makeAction(t)
				a2b := makeAction(t)

				c2.On("Check", ev).Return(true, nil).Once()
				a2a.On("Execute", ev).Return(nil).Once()
				a2b.On("Execute", ev).Return(nil).Once()

				r2 := &entities.Rule{
					When: []entities.Condition{c2},
					Then: []entities.Action{a2a, a2b},
				}

				return fields{rules: []*entities.Rule{r1, r2}}, ev, func(t *testing.T) {
					c1a.AssertExpectations(t)
					c1b.AssertNotCalled(t, "Check", ev)
					a1.AssertNotCalled(t, "Execute", ev)

					c2.AssertExpectations(t)
					a2a.AssertExpectations(t)
					a2b.AssertExpectations(t)
				}
			},
			want: want{errContains: ""},
		},
		{
			name: "error path: condition check returns error (wraps with 'conditional checking conditions:')",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}

				cErr := makeCond(t)
				cErr.On("Check", ev).Return(false, errors.New("boom")).Once()

				// Ensure no actions are run
				a := makeAction(t)

				r := &entities.Rule{
					When: []entities.Condition{cErr},
					Then: []entities.Action{a},
				}

				return fields{rules: []*entities.Rule{r}}, ev, func(t *testing.T) {
					cErr.AssertExpectations(t)
					a.AssertNotCalled(t, "Execute", ev)
				}
			},
			want: want{errContains: "conditional checking conditions: boom"},
		},
		{
			name: "error path: action execute returns error (wraps with 'conditional running action:')",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}

				c := makeCond(t)
				c.On("Check", ev).Return(true, nil).Once()

				a1 := makeAction(t)
				a2 := makeAction(t)

				a1.On("Execute", ev).Return(errors.New("kaboom")).Once()
				// a2 should not run because a1 failed
				// (the code returns immediately on error)

				r := &entities.Rule{
					When: []entities.Condition{c},
					Then: []entities.Action{a1, a2},
				}

				return fields{rules: []*entities.Rule{r}}, ev, func(t *testing.T) {
					c.AssertExpectations(t)
					a1.AssertExpectations(t)
					a2.AssertNotCalled(t, "Execute", ev)
				}
			},
			want: want{errContains: "conditional running action: kaboom"},
		},
		{
			name: "no rules: returns nil",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}
				return fields{rules: nil}, ev, func(t *testing.T) {}
			},
			want: want{errContains: ""},
		},
		{
			name: "no rule matches: returns nil and executes no actions",
			build: func(t *testing.T) (fields, *entities.Event, func(t *testing.T)) {
				ev := &entities.Event{}

				c1 := makeCond(t)
				c2 := makeCond(t)
				a1 := makeAction(t)
				a2 := makeAction(t)

				c1.On("Check", ev).Return(false, nil).Once()
				c2.On("Check", ev).Return(false, nil).Once()

				r1 := &entities.Rule{When: []entities.Condition{c1}, Then: []entities.Action{a1}}
				r2 := &entities.Rule{When: []entities.Condition{c2}, Then: []entities.Action{a2}}

				return fields{rules: []*entities.Rule{r1, r2}}, ev, func(t *testing.T) {
					c1.AssertExpectations(t)
					c2.AssertExpectations(t)
					a1.AssertNotCalled(t, "Execute", ev)
					a2.AssertNotCalled(t, "Execute", ev)
				}
			},
			want: want{errContains: ""},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f, ev, verify := c.build(t)

			act := &Conditional{RuleChain: f.rules}
			err := act.Execute(ev)

			if c.want.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), c.want.errContains)
			}

			// run per-case verifications (call counts, not-called assertions, etc.)
			verify(t)
		})
	}
}

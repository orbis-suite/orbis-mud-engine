package actions

import (
	"testing"

	"example.com/mud/mocks"
	"example.com/mud/world/entities"
	"github.com/stretchr/testify/require"
)

func TestPrint_Execute(t *testing.T) {
	t.Parallel()

	newEntity := func(name string) *entities.Entity {
		return entities.NewEntity(
			name,
			"desc",
			[]string{name},
			nil,
			nil,
			nil,
		)
	}

	type tc struct {
		name      string
		print     Print
		setup     func(t *testing.T) (ev entities.Event, expectRecipient *entities.Entity)
		wantErr   bool
		errString string
	}

	cases := []tc{
		{
			name:  "error when Publisher is nil",
			print: Print{Text: "hello", EventRole: entities.EventRoleSource},
			setup: func(t *testing.T) (entities.Event, *entities.Entity) {
				room := newEntity("room")
				src := newEntity("source")
				return entities.Event{
					Publisher: nil,
					Room:      room,
					Source:    src,
				}, nil
			},
			wantErr:   true,
			errString: "publisher in event may not be nil for print action",
		},
		{
			name:  "error on invalid role",
			print: Print{Text: "hello", EventRole: entities.EventRoleUnknown},
			setup: func(t *testing.T) (entities.Event, *entities.Entity) {
				room := newEntity("room")
				src := newEntity("source")
				mp := new(mocks.MockPublisher) // won’t be used
				return entities.Event{
					Publisher: mp,
					Room:      room,
					Source:    src,
				}, nil
			},
			wantErr:   true,
			errString: "invalid role",
		},
		{
			name:  "success - prints to Source",
			print: Print{Text: "hello", EventRole: entities.EventRoleSource},
			setup: func(t *testing.T) (entities.Event, *entities.Entity) {
				room := newEntity("room")
				src := newEntity("source")
				mp := new(mocks.MockPublisher)

				mp.
					On("PublishTo", room, src, "hello").
					Return(nil).
					Once()

				return entities.Event{
					Publisher: mp,
					Room:      room,
					Source:    src,
				}, src
			},
		},
		{
			name:  "success - prints to Instrument",
			print: Print{Text: "hello", EventRole: entities.EventRoleInstrument},
			setup: func(t *testing.T) (entities.Event, *entities.Entity) {
				room := newEntity("room")
				src := newEntity("source")
				inst := newEntity("instrument")
				mp := new(mocks.MockPublisher)

				mp.
					On("PublishTo", room, inst, "hello").
					Return(nil).
					Once()

				return entities.Event{
					Publisher:  mp,
					Room:       room,
					Source:     src,
					Instrument: inst,
				}, inst
			},
		},
		{
			name:  "success - prints to Target",
			print: Print{Text: "hello", EventRole: entities.EventRoleTarget},
			setup: func(t *testing.T) (entities.Event, *entities.Entity) {
				room := newEntity("room")
				tgt := newEntity("target")
				mp := new(mocks.MockPublisher)

				mp.
					On("PublishTo", room, tgt, "hello").
					Return(nil).
					Once()

				return entities.Event{
					Publisher: mp,
					Room:      room,
					Target:    tgt,
				}, tgt
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ev, _ := c.setup(t)

			err := c.print.Execute(&ev)
			if c.wantErr {
				require.Error(t, err)
				if c.errString != "" {
					require.Contains(t, err.Error(), c.errString)
				}
				return
			}

			require.NoError(t, err)

			// Verify mock expectations if a mock was used
			if mp, ok := ev.Publisher.(*mocks.MockPublisher); ok {
				mp.AssertExpectations(t)
			}
		})
	}
}

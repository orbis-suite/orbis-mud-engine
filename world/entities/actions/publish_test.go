package actions

import (
	"testing"

	"example.com/mud/mocks"
	"example.com/mud/world/entities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPublish_Execute(t *testing.T) {
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
		publish   Publish
		setup     func(t *testing.T) entities.Event
		wantErr   bool
		errString string
	}

	cases := []tc{
		{
			name:    "error when Publisher is nil",
			publish: Publish{Text: "hello"},
			setup: func(t *testing.T) entities.Event {
				return entities.Event{
					Publisher:  nil,
					Room:       newEntity("room"),
					Source:     newEntity("source"),
					Instrument: newEntity("instrument"),
					Target:     newEntity("target"),
				}
			},
			wantErr:   true,
			errString: "publisher in event may not be nil for publish action",
		},
		{
			name:    "success publishes formatted message to room with exclusions",
			publish: Publish{Text: "hello"},
			setup: func(t *testing.T) entities.Event {
				room := newEntity("room")
				source := newEntity("source")
				instrument := newEntity("instrument")
				target := newEntity("target")

				mp := new(mocks.MockPublisher)

				mp.
					On(
						"Publish",
						room,
						"hello",
						mock.MatchedBy(func(ex []*entities.Entity) bool {
							if len(ex) != 3 {
								return false
							}
							return ex[0] == source && ex[1] == instrument && ex[2] == target
						}),
					).
					Return(nil).
					Once()

				return entities.Event{
					Publisher:  mp,
					Room:       room,
					Source:     source,
					Instrument: instrument,
					Target:     target,
				}
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ev := c.setup(t)

			err := c.publish.Execute(&ev)
			if c.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errString)
				return
			}

			require.NoError(t, err)

			if mp, ok := ev.Publisher.(*mocks.MockPublisher); ok {
				mp.AssertExpectations(t)
			}
		})
	}
}

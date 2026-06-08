package actions

import (
	"testing"

	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	"github.com/stretchr/testify/require"
)

func TestCopy_Execute(t *testing.T) {
	t.Parallel()

	const (
		toCopyName    = "Orb"
		recipientName = "Sailor"
	)

	newSrc := func() *entities.Entity {
		return entities.NewEntity(
			toCopyName,
			"A mysterious orb",
			[]string{"orb"},
			nil,
			nil,
			nil,
		)
	}

	type tc struct {
		name      string
		copy      Copy
		setup     func(t *testing.T) (ev entities.Event, recipientContainer *components.Container, src *entities.Entity)
		wantErr   bool
		errString string
	}

	cases := []tc{
		{
			name: "error when EntitiesById is nil",
			copy: Copy{
				EntityId:      "anything",
				EventRole:     entities.EventRoleSource,
				ComponentType: entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *entities.Entity) {
				return entities.Event{EntitiesById: nil}, nil, nil
			},
			wantErr:   true,
			errString: "entities by id map in event may not be nil",
		},
		{
			name: "error on invalid role",
			copy: Copy{
				EntityId:      toCopyName,
				EventRole:     entities.EventRoleUnknown,
				ComponentType: entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *entities.Entity) {
				return entities.Event{
					EntitiesById: map[string]*entities.Entity{},
				}, nil, nil
			},
			wantErr:   true,
			errString: "invalid role 'unknown'",
		},
		{
			name: "error when component type is invalid",
			copy: Copy{
				EntityId:      toCopyName,
				EventRole:     entities.EventRoleTarget,
				ComponentType: entities.ComponentUnknown,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *entities.Entity) {
				recipient, container := makeContainerRecipient(recipientName)
				require.NotNil(t, container, "recipient should have a container")

				src := newSrc()
				ev := entities.Event{
					EntitiesById: map[string]*entities.Entity{
						toCopyName: src,
					},
					Target: recipient,
				}
				return ev, container, src
			},
			wantErr:   true,
			errString: "entity does not have component with children",
		},
		{
			name: "error when entity id does not exist in event map",
			copy: Copy{
				EntityId:      "invalid-id",
				EventRole:     entities.EventRoleTarget,
				ComponentType: entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *entities.Entity) {
				recipient, container := makeContainerRecipient(recipientName)
				require.NotNil(t, container, "recipient should have a container")

				ev := entities.Event{
					EntitiesById: map[string]*entities.Entity{},
					Target:       recipient,
				}
				return ev, container, nil
			},
			wantErr:   true,
			errString: "entity 'invalid-id' doesn't exist",
		},
		{
			name: "success",
			copy: Copy{
				EntityId:      toCopyName,
				EventRole:     entities.EventRoleTarget,
				ComponentType: entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *entities.Entity) {
				recipient, container := makeContainerRecipient(recipientName)
				require.NotNil(t, container, "recipient should have a container")

				src := newSrc()
				ev := entities.Event{
					EntitiesById: map[string]*entities.Entity{
						toCopyName: src,
					},
					Target: recipient,
				}
				return ev, container, src
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ev, recipientContainer, src := c.setup(t)

			err := c.copy.Execute(&ev)
			if c.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errString)
				return
			}

			require.NoError(t, err)

			require.NotNil(t, recipientContainer, "recipient container expected for success cases")
			children := recipientContainer.GetChildren().GetChildren()

			require.Equal(t, toCopyName, children[0].Name)
			require.NotSame(t, src, children[0], "expected a copy, not the original pointer")
		})
	}
}

// makeContainerRecipient creates an entity with a Container component and returns both.
func makeContainerRecipient(name string) (*entities.Entity, *components.Container) {
	tags := []string{"sailor"}

	e := entities.NewEntity(
		name,
		"A lonely sailor",
		tags,
		nil,
		nil,
		nil,
	)

	c := components.NewContainer()
	e.Add(c)

	return e, c
}

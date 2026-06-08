package actions

import (
	"testing"

	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	"github.com/stretchr/testify/require"
)

func TestMove_Execute(t *testing.T) {
	t.Parallel()

	newObject := func() *entities.Entity {
		return entities.NewEntity(
			"Orb",
			"A mysterious orb",
			[]string{"orb"},
			nil,
			nil,
			nil,
		)
	}

	type tc struct {
		name      string
		move      Move
		setup     func(t *testing.T) (ev entities.Event, originContainer, destinationContainer *components.Container, target *entities.Entity)
		wantErr   bool
		errString string
	}

	cases := []tc{
		{
			name: "error when object role is invalid",
			move: Move{
				RoleObject:      entities.EventRoleUnknown,
				RoleDestination: entities.EventRoleTarget,
				ComponentType:   entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *components.Container, *entities.Entity) {
				object := newObject()

				_, originContainer := makeContainerRecipient("1")
				originContainer.AddChild(object)

				destination, destinationContainer := makeContainerRecipient("2")

				return entities.Event{
					Source: object,
					Target: destination,
				}, originContainer, destinationContainer, object
			},
			wantErr:   true,
			errString: "move execute object to move: invalid role",
		},
		{
			name: "error when destination role is invalid",
			move: Move{
				RoleObject:      entities.EventRoleSource,
				RoleDestination: entities.EventRoleUnknown,
				ComponentType:   entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *components.Container, *entities.Entity) {
				object := newObject()

				_, originContainer := makeContainerRecipient("1")
				originContainer.AddChild(object)

				destination, destinationContainer := makeContainerRecipient("2")

				return entities.Event{
					Source: object,
					Target: destination,
				}, originContainer, destinationContainer, object
			},
			wantErr:   true,
			errString: "move execute destination: invalid role",
		},
		{
			name: "error when component type is invalid",
			move: Move{
				RoleObject:      entities.EventRoleSource,
				RoleDestination: entities.EventRoleTarget,
				ComponentType:   entities.ComponentUnknown,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *components.Container, *entities.Entity) {
				object := newObject()

				_, originContainer := makeContainerRecipient("1")
				originContainer.AddChild(object)

				destination, destinationContainer := makeContainerRecipient("2")

				return entities.Event{
					Source: object,
					Target: destination,
				}, originContainer, destinationContainer, object
			},
			wantErr:   true,
			errString: "entity does not have component with children",
		},
		{
			name: "success",
			move: Move{
				RoleObject:      entities.EventRoleSource,
				RoleDestination: entities.EventRoleTarget,
				ComponentType:   entities.ComponentContainer,
			},
			setup: func(t *testing.T) (entities.Event, *components.Container, *components.Container, *entities.Entity) {
				object := newObject()

				_, originContainer := makeContainerRecipient("1")
				originContainer.AddChild(object)

				destination, destinationContainer := makeContainerRecipient("2")

				return entities.Event{
					Source: object,
					Target: destination,
				}, originContainer, destinationContainer, object
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ev, originContainer, destinationContainer, object := c.setup(t)

			err := c.move.Execute(&ev)
			if c.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errString)
				return
			}

			require.NoError(t, err)

			// origin container should have no children
			require.NotNil(t, originContainer, "origin container expected for success cases")
			originChildren := originContainer.GetChildren().GetChildren()
			require.Len(t, originChildren, 0)

			// destination container should now have one child
			require.NotNil(t, destinationContainer, "destination container expected for success cases")
			destinationChildren := destinationContainer.GetChildren().GetChildren()
			require.Len(t, destinationChildren, 1)

			require.Same(t, destinationChildren[0], object, "Expected the same object to be moved to the destination, not a copy")
		})
	}
}

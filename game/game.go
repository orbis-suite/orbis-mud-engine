package main

import (
	"example.com/mud/game/commands"
	"example.com/mud/game/entities"
	"example.com/mud/game/rooms"
	"example.com/mud/sdk"
)

// Game implements sdk.Game, providing the world definition and event handling.
type Game struct{}

func (g *Game) GetManifest() *sdk.Manifest {
	return &sdk.Manifest{
		Rooms:    rooms.All(),
		Entities: entities.All(),
		Commands: commands.All(),
	}
}

func (g *Game) HandleEvent(e *sdk.Event) []sdk.Action {
	return entities.Dispatch(e)
}

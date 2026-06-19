package main

import (
	"fmt"

	"example.com/mud/game/commands"
	"example.com/mud/game/entities"
	"example.com/mud/game/rooms"
	"example.com/mud/sdk"
)

// Game implements sdk.Game, providing the world definition and event handling.
type Game struct{}

func (g *Game) GetManifest() *sdk.Manifest {
	worldRooms := rooms.GenerateWorld()
	startingRoom := worldRooms[0].ID
	fmt.Println(startingRoom)

	return &sdk.Manifest{
		StartingRoom: startingRoom,
		Rooms:        worldRooms,
		Entities:     entities.All(),
		Commands:     commands.All(),
	}
}

func (g *Game) HandleEvent(e *sdk.Event) []sdk.Action {
	return entities.Dispatch(e)
}

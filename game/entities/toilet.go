package entities

import (
	"example.com/mud/sdk"

	"example.com/mud/game/commands"
)

var toiletDef = &sdk.EntityDef{
	ID:          "Toilet",
	Name:        "Toilet",
	Description: "A {'toilet' | bold | yellow}, it's shiny and porcelain.",
	Aliases:     []string{"toilet"},
	Tags:        []string{"furniture"},
	Reactions: map[sdk.Command][]sdk.Action{
		commands.KissCmd.Name: sdk.Actions(
			sdk.Print("source", "Maybe... no. You reconsider. {'Do not kiss the toilet.' | bold | underline }"),
			sdk.After(200, sdk.Print("source", "Nevermind you do it.")),
		),
	},
}

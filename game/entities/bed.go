package entities

import "example.com/mud/sdk"

var bedDef = &sdk.EntityDef{
	ID:          "Bed",
	Name:        "Bed",
	Description: "A {'bed' | bold | yellow} is well-made and looks inviting.",
	Aliases:     []string{"bed"},
	Tags:        []string{"furniture"},
}

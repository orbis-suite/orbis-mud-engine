package entities

import "example.com/mud/sdk"

var nickelDef = &sdk.EntityDef{
	ID:          "Nickel",
	Name:        "Nickel",
	Description: "A shining {'nickel' | bold | yellow} lies here, Thomas Jefferson's handsome side profile glinting faintly as though pleased with its escape.",
	Aliases:     []string{"nickel"},
	Tags:        []string{"item"},
}

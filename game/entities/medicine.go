package entities

import "example.com/mud/sdk"

var medicineDef = &sdk.EntityDef{
	ID:          "Medicine",
	Name:        "Medicine",
	Description: "Bottles of pills line the shelves -- the eponymous medicine for which the cabinet is named.",
	Aliases:     []string{"medicine", "pills", "bottles"},
	Tags:        []string{"consumable"},
}

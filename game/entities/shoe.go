package entities

import "example.com/mud/sdk"

var shoeDef = &sdk.EntityDef{
	ID:                 "Shoe",
	Name:               "Shoe",
	Description:        "A battered left shoe, the sole hangs unattached at the toe.",
	Aliases:            []string{"shoe"},
	Tags:               []string{"item"},
	ContainerID:        "Box",
	ContainerComponent: "Container",
}

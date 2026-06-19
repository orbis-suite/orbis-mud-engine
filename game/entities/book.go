package entities

import "example.com/mud/sdk"

var bookDef = &sdk.EntityDef{
	ID:                 "Book",
	Name:               "Book",
	Description:        "A {'book' | bold | yellow} with a leather cover, a bold adaptation of VeggieTales with human characters.",
	Aliases:            []string{"book"},
	Tags:               []string{"item"},
	ContainerID:        "Box",
	ContainerComponent: "Container",
}

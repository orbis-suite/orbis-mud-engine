package entities

import "example.com/mud/sdk"

var boxDef = &sdk.EntityDef{
	ID:                "Box",
	Name:              "Box",
	Description:       "A cardboard {'box' | bold | yellow} is here, too.",
	Aliases:           []string{"box"},
	Tags:              []string{"furniture"},
	HasContainer:      true,
	ContainerPrefix:   "Inside the box:",
	ContainerRevealed: true,
}

func boxReact(e *sdk.Event) []sdk.Action {
	switch e.Command {
	case "open":
		return sdk.Actions(
			sdk.Reveal("target", "Container"),
			sdk.Print("source", "You open the box."),
			sdk.Publish("{source} opens the box", "source"),
		)
	case "close":
		return sdk.Actions(
			sdk.Hide("target", "Container"),
			sdk.Print("source", "You close the box."),
			sdk.Publish("{source} closes the box.", "source"),
		)
	}
	return nil
}

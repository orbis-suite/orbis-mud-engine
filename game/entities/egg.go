package entities

import "example.com/mud/sdk"

var eggDef = &sdk.EntityDef{
	ID:                 "Egg",
	Name:               "Egg",
	Description:        "A bulbous, green-speckled egg.",
	Aliases:            []string{"egg"},
	Tags:               []string{"egg"},
	Fields:             map[string]string{"angry": "false"},
	ContainerID:        "Player",
	ContainerComponent: "Inventory",
}

func eggReact(e *sdk.Event) []sdk.Action {
	switch e.Command {
	case "attack":
		if e.Target != nil && !e.Target.FieldBool("angry") {
			return sdk.Actions(
				sdk.SetField("target", "angry", true),
				sdk.Print("source", "The egg is now angry that you hit it."),
			)
		}
		return sdk.Actions(
			sdk.SetField("target", "angry", false),
			sdk.Print("source", "The egg is calmed after you strike it again"),
		)
	case "take", "drop":
		return itemReact(e)
	}
	return nil
}

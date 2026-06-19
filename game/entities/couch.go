package entities

import "example.com/mud/sdk"

var couchDef = &sdk.EntityDef{
	ID:          "Couch",
	Name:        "Couch",
	Description: "A soft, inviting {'couch' | bold | yellow} rests here, its cushions sagged just enough to suggest long use. It seems comfortable, with plenty of room for something to be hidden within.",
	Aliases:     []string{"couch"},
	Tags:        []string{"furniture"},
}

func couchReact(e *sdk.Event) []sdk.Action {
	switch e.Command {
	case "attack":
		if e.Instrument != nil && e.Instrument.HasTag("egg") {
			return sdk.Actions(
				sdk.Print("source", "You hit the egg upon the couch, gently, as to not disturb the egg."),
				sdk.Publish("{source} hits their egg upon the couch, smiling vacantly to themselves.", "source"),
			)
		}
		return sdk.Actions(
			sdk.Print("source", "As you beat upon the couch, a {'nickel' | bold | yellow} falls out."),
			sdk.Publish("{source} beats upon the couch, and a shining {'nickel'| bold | yellow} falls out from under a cushion.", "source"),
			sdk.Spawn("Nickel", "room", "Room"),
		)
	case "kiss":
		return kissableReact(e)
	}
	return nil
}

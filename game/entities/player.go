package entities

import (
	"fmt"

	"example.com/mud/sdk"
)

var playerDef = &sdk.EntityDef{
	ID:           "Player",
	Name:         "Player",
	Description:  "Player Template",
	Aliases:      []string{"player"},
	Tags:         []string{"player"},
	HasInventory: true,
}

func playerReact(e *sdk.Event) []sdk.Action {
	if e.Command != "attack" {
		return nil
	}
	if e.Instrument != nil && e.Instrument.HasTag("player") {
		return sdk.Actions(sdk.Print("source", fmt.Sprintf("You attempt to beat %s with %s, but they are too heavy to lift.", e.Target.Name, e.Instrument.Name)))
	}
	if e.Source != nil && e.Source.TemplateID == e.TargetID {
		return sdk.Actions(
			sdk.Print("source", "You hit yourself upon the head, hard enough to hurt."),
			sdk.Publish("{source} hits themselves upon the head, their rage directed inward.", "source"),
		)
	}
	if e.Instrument != nil && e.Instrument.TemplateID == e.TargetID {
		return sdk.Actions(sdk.Print("source", fmt.Sprintf("You wonder if you might be able to beat %s with themselves, but disregard the idea.", e.Target.Name)))
	}
	return sdk.Actions(
		sdk.Print("source", fmt.Sprintf("You beat a great big indent into %s's head", e.Target.Name)),
		sdk.Print("target", "{source} caves your head in."),
		sdk.Publish(fmt.Sprintf("{source} violently whacks %s upon their head.", e.Target.Name), "source", "target"),
	)
}

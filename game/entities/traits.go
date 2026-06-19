package entities

import (
	"fmt"

	"example.com/mud/sdk"
)

// itemReact handles take/drop for standard items.
func itemReact(e *sdk.Event) []sdk.Action {
	switch e.Command {
	case "take":
		if e.Source != nil && !e.Source.HasChildInComponent(e.TargetID, "Inventory") {
			return sdk.Actions(
				sdk.Print("source", fmt.Sprintf("You pocket %s", e.Target.Name)),
				sdk.Publish(fmt.Sprintf("{source} pockets %s", e.Target.Name), "source"),
				sdk.Move("target", "source", "Inventory"),
			)
		}
		return sdk.Actions(sdk.Print("source", fmt.Sprintf("You're already carrying %s", e.Target.Name)))

	case "drop":
		if e.Source != nil && e.Source.HasChildInComponent(e.TargetID, "Inventory") {
			return sdk.Actions(
				sdk.Print("source", fmt.Sprintf("You drop %s onto the ground.", e.Target.Name)),
				sdk.Publish(fmt.Sprintf("{source} drops %s onto the ground.", e.Target.Name), "source"),
				sdk.Move("target", "room", "Room"),
			)
		}
		return sdk.Actions(sdk.Print("source", fmt.Sprintf("You aren't carrying %s", e.Target.Name)))
	}
	return nil
}

// kissableReact handles kiss for entities that can be kissed.
func kissableReact(e *sdk.Event) []sdk.Action {
	if e.Command == "kiss" {
		return sdk.Actions(
			sdk.Print("source", fmt.Sprintf("You kiss the %s", e.Target.Name)),
			sdk.Publish(fmt.Sprintf("{source} kisses the %s.", e.Target.Name), "source"),
		)
	}
	return nil
}

// hittableReact handles attack for entities that can be hit.
func hittableReact(e *sdk.Event) []sdk.Action {
	if e.Command != "attack" {
		return nil
	}
	if e.Instrument != nil && e.Instrument.TemplateID == e.TargetID {
		return sdk.Actions(sdk.Print("source", "You can't hit something with itself."))
	}
	if e.Instrument != nil {
		return sdk.Actions(
			sdk.Print("source", fmt.Sprintf("You hit the %s with %s", e.Target.Name, e.Instrument.Name)),
			sdk.Publish(fmt.Sprintf("{source} hits the %s with %s.", e.Target.Name, e.Instrument.Name), "source"),
		)
	}
	return sdk.Actions(
		sdk.Print("source", fmt.Sprintf("You hit the %s", e.Target.Name)),
		sdk.Publish(fmt.Sprintf("{source} hits the %s.", e.Target.Name), "source"),
	)
}

// standardReact handles both kiss and attack (the Standard trait from the DSL).
func standardReact(e *sdk.Event) []sdk.Action {
	if acts := kissableReact(e); acts != nil {
		return acts
	}
	return hittableReact(e)
}

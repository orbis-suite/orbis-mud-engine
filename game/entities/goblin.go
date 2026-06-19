package entities

import (
	"fmt"

	"example.com/mud/sdk"
)

var goblinDef = &sdk.EntityDef{
	ID:           "Goblin",
	Name:         "Goblin",
	Description:  "A funny {'goblin' | bold | yellow} man no bigger than your fist smiles warmly.",
	Aliases:      []string{"goblin", "man"},
	Tags:         []string{"npc"},
	HasInventory: true,
}

func goblinReact(e *sdk.Event) []sdk.Action {
	switch e.Command {
	case "attack":
		return sdk.Actions(
			sdk.Print("source", "As you throw a {'punch' | yellow} at the goblin, he jumps around you, {'kissing' | red} your forehead."),
			sdk.Publish("{source} tries and fails to attack the goblin, yet they're rewarded with a gentle {'kiss' | red } from the creature.", "source"),
		)

	case "kiss":
		if e.Source != nil && !e.Source.HasChildInComponent("Goblin", "Inventory") {
			return sdk.Actions(
				sdk.Print("source", "You give the goblin a kiss upon his {'sweaty' | blue} brow, and he {'hops' | italic} into your pocket."),
				sdk.Publish("{source} gives the goblin a {'kiss' | bold | red}, before the goblin {'jumps' | italic} into {source}'s pocket.", "source"),
				sdk.Move("target", "source", "Inventory"),
			)
		}
		return sdk.Actions(
			sdk.Print("source", "You look into your pocket and plant another kiss upon the goblin's cheek."),
			sdk.Publish("{source} gives the goblin in their pocket a big wet {'kiss' | bold | red}.", "source"),
		)

	case "give":
		if e.Instrument != nil && e.Instrument.HasTag("item") {
			return sdk.Actions(
				sdk.Print("source", fmt.Sprintf("You give the goblin your %s, and he accepts it happily. 'You win!' He says, 'You win the game for giving the goblin an item!'", e.Instrument.Name)),
				sdk.Publish(fmt.Sprintf("{source} gives the goblin %s. The goblin is overjoyed.", e.Instrument.Name), "source"),
				sdk.Move("instrument", "target", "Inventory"),
			)
		}
		return sdk.Actions(
			sdk.Print("source", "The goblin gives you a smile, shaking his head softly. 'I don't want that stupid smelly thing,' he says."),
			sdk.Publish(fmt.Sprintf("{source} tries to give the goblin %s, but he refuses to take it.", e.Instrument.Name), "source"),
		)
	}
	return nil
}

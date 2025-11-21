package commands

import (
	"fmt"

	"example.com/mud/models"
)

var Commands = map[string]struct{}{}

var DirectionAliases = map[string]string{
	"n":     models.DirectionNorth,
	"north": models.DirectionNorth,

	"s":     models.DirectionSouth,
	"south": models.DirectionSouth,

	"e":    models.DirectionEast,
	"east": models.DirectionEast,

	"w":    models.DirectionWest,
	"west": models.DirectionWest,

	"u":  models.DirectionUp,
	"up": models.DirectionUp,

	"d":    models.DirectionDown,
	"down": models.DirectionDown,
}

var VerbAliases = map[string]string{}

var Patterns = []models.Pattern{}

func RegisterBuiltInCommands() error {
	return RegisterCommands([]*models.CommandDefinition{
		&helpCommand,
		&inventoryCommand,
		&lookCommand,
		// &moveCommand,
		&mapCommand,
		&trackCommand,
	})
}

func RegisterCommands(defs []*models.CommandDefinition) error {
	for _, cd := range defs {
		if len(cd.Aliases) == 0 {
			return fmt.Errorf("command '%s' has no aliases", cd.Name)
		}

		canonical := cd.Aliases[0]
		Commands[canonical] = struct{}{}

		for _, alias := range cd.Aliases {
			VerbAliases[alias] = canonical
		}

		for _, pat := range cd.Patterns {
			Patterns = append(Patterns, models.Pattern{
				Kind:           cd.Name,
				Tokens:         pat.Tokens,
				HelpMessage:    pat.HelpMessage,
				NoMatchMessage: pat.NoMatchMessage,
			})
		}
	}
	return nil
}

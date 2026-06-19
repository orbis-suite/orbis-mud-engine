package commands

import "example.com/mud/sdk"

var KissCmd = &sdk.CommandDef{
	Name:    "kiss",
	Aliases: []string{"kiss", "smooch"},
	Patterns: []sdk.CommandPattern{
		{Syntax: "kiss {target}", NoMatch: "you don't want to kiss that."},
	},
}

func All() []*sdk.CommandDef {
	return []*sdk.CommandDef{
		{
			Name:    "attack",
			Aliases: []string{"attack", "hit", "beat"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "attack {target}", NoMatch: "You don't want to attack that."},
				{Syntax: "attack {target} with {instrument}", NoMatch: "You don't want to attack that with that."},
			},
		},
		KissCmd,
		{
			Name:    "open",
			Aliases: []string{"open"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "open {target}", NoMatch: "You can't open that."},
			},
		},
		{
			Name:    "close",
			Aliases: []string{"close"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "close {target}", NoMatch: "You can't close that."},
			},
		},
		{
			Name:    "take",
			Aliases: []string{"take", "grab", "pickup"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "take {target}", NoMatch: "you can't pick that up."},
			},
		},
		{
			Name:    "drop",
			Aliases: []string{"drop"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "drop {target}", NoMatch: "you can't drop that."},
			},
		},
		{
			Name:    "give",
			Aliases: []string{"give", "hand"},
			Patterns: []sdk.CommandPattern{
				{Syntax: "give {instrument} to {target}", NoMatch: "You can't give that to that."},
				{Syntax: "give {target} {instrument}", NoMatch: "You can't give that to that."},
			},
		},
	}
}

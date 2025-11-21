local Command = require("src.types.command")

local move: Command = {
	aliases = { "move", "go" },
	patterns = {
		{
			syntax = "move {direction}",
			help = "move to another room.",
			noMatch = "You can't get there.",
		},
		{
			syntax = "{direction}",
			help = "move to another room.",
			noMatch = "You can't get there.",
		}
	},
}

local map: Command = {
	aliases = { "map", "m" },
	patterns = {
		{
			syntax = "map",
			help = "look at your map",
			noMatch = "You can't look at your map right now.",
		},
	},
}

return {
    move = move,
	map = map,
} as { string: Command }
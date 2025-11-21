local Command = require("src.types.command")

local attack: Command = {
	aliases = { "attack", "hit", "beat" },
	patterns = {
		{
			syntax = "attack {target}",
			help = "attack something or somebody.",
			noMatch = "You reconsider attacking {target}",
		},
	},
}

local talk: Command = {
	aliases = { "talk", "chat" },
	patterns = {
		{
			syntax = "talk to {target}",
			help = "attack something or somebody.",
			noMatch = "You reconsider attacking {target}",
		},
	},
}

return {
	attack = attack,
	talk = talk,
} as { string: Command }

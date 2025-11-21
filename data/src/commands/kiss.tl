local Command = require("src.types.command")

local kiss: Command = {
	aliases = { "kiss", "smooch" },
	patterns = {
		{
			syntax = "kiss {target}",
			help = "kiss something or somebody.",
			noMatch = "You reconsider kissing {target}",
		},
	},
}

return {
	kiss = kiss,
} as { string: Command }
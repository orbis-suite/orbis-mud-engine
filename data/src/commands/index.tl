local CoreMix = require("src.std.core.mix")
local Command = require("src.types.command")
local attackCommands = require("src.commands.attack")
local kissCommands = require("src.commands.kiss")
local moveCommands = require("src.commands.move")

local commands = CoreMix.mix(attackCommands, kissCommands, moveCommands)

return commands as { string: Command }
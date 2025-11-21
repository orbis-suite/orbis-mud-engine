local record CommandPattern
    syntax: string
    help: string
    noMatch: string
end

local record Command
    aliases: {string}
    patterns: {CommandPattern}
end

return Command
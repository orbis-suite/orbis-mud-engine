local record EntitySnapshot
    id: string
    name: string
    description: string
    aliases: {string}
    tags: {string}
    fields: {string: any}
    children: {string: {EntitySnapshot}}
end

return EntitySnapshot
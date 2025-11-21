local type ReactionMap = {string: {string: function(): boolean}}

local record Entity
    name: string
    description: string
    aliases: {string}
    tags: {string} | nil
    fields: {string: any} | nil
    init: function() | nil
    reactions: ReactionMap | nil
    children: {string: {Entity}} | nil
end

local record Trait
    aliases: {string} | nil
    tags: {string}| nil
    reactions: ReactionMap | nil
    init: function() | nil
    fields: {string: any} | nil
    children: {string: {Entity}} | nil
end

return {
    ReactionMap = ReactionMap,
    Entity = Entity,
    Trait = Trait,
}
local entity_types = require("src.types.entity")
local ReactionMap = entity_types.ReactionMap
local Entity = entity_types.Entity
local Trait = entity_types.Trait

local CoreFunctional = require("src.std.core.functional")
local CoreReact = require("src.std.core.react")

local record CoreTraits
    apply_traits: function(Entity, {Trait}): Entity
end

local M: CoreTraits = {}

local function merge_reactions(
    child: ReactionMap | nil,
    parent: ReactionMap | nil
): ReactionMap | nil
    if child == nil and parent == nil then
        return nil
    end

    local out: ReactionMap = {}

    local function copy_into(src: ReactionMap)
        for reaction, variants in pairs(src) do
            local dest_variants = out[reaction]
            if dest_variants == nil then
                dest_variants = {}
                out[reaction] = dest_variants
            end

            for variant, fn in pairs(variants) do
                dest_variants[variant] = CoreReact.compose(fn, dest_variants[variant])
            end
        end
    end

    -- parent first: provides defaults
    if parent ~= nil then
        copy_into(parent as ReactionMap)
    end

    -- child second: overrides / extends
    if child ~= nil then
        copy_into(child as ReactionMap)
    end

    return out
end

local function apply_trait(e: Entity, t: Trait): Entity
    e.reactions = merge_reactions(e.reactions, t.reactions)
    return e
end

function M.apply_traits(entity: Entity, traits: {Trait}): Entity
    return CoreFunctional.reduce(traits, entity, function(acc: Entity, t: Trait):Entity
        return apply_trait(acc, t)
    end)
end

return M
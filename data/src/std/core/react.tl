local record ReactionRule
    cond: function(): boolean
    fn: function()
end

local record CoreReact
    when: function(cond: function():boolean, fn: function()): ReactionRule
    default: function(fn: function()): ReactionRule
    match: function(rules: { ReactionRule }): function():boolean
    compose: function(fChild: function():boolean|nil, fParent: function():boolean|nil): function():boolean
end

local R: CoreReact = {}

function R.when(cond: function():boolean, fn: function()): ReactionRule
    return { cond = cond, fn = fn }
end

function R.default(fn: function()): ReactionRule
    return {
        cond = function(): boolean
            return true
        end,
        fn = fn
    }
end

function R.match(rules: { ReactionRule }): function():boolean
    return function():boolean
        for _, r in ipairs(rules) do
            if r.cond() then
                r.fn()
                return true
            end 
        end
        return false
    end
end

function R.compose(fChild: function():boolean|nil, fParent: function():boolean|nil): function():boolean
    if not fChild then
        return fParent
    end

    if not fParent then
        return fChild
    end

    return
        function():boolean
            return fChild() or fParent()
        end
end

return R

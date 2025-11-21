local record Functional
    map:    function<T, U>(arr: {T}, fn: function(T): U): {U}
    filter: function<T>(arr: {T}, fn: function(T): boolean): {T}
    reduce: function<T, U>(arr: {T}, init: U, fn: function(U, T): U): U
    find:   function<T>(arr: {T}, fn: function(T): boolean): T | nil
    any:    function<T>(arr: {T}, fn: function(T): boolean): boolean
    all:    function<T>(arr: {T}, fn: function(T): boolean): boolean
end

local F: Functional = {} as Functional

function F.map<T, U>(arr: {T}, fn: function(T): U): {U}
    local out: {U} = {}
    for i, v in ipairs(arr) do
        out[i] = fn(v)
    end
    return out
end

function F.filter<T>(arr: {T}, fn: function(T): boolean): {T}
    local out: {T} = {}
    for _, v in ipairs(arr) do
        if fn(v) then
            table.insert(out, v)
        end
    end
    return out
end

function F.reduce<T, U>(arr: {T}, init: U, fn: function(U, T): U): U
    local acc = init
    for _, v in ipairs(arr) do
        acc = fn(acc, v)
    end
    return acc
end

function F.find<T>(arr: {T}, fn: function(T): boolean): T | nil
    for _, v in ipairs(arr) do
        if fn(v) then
            return v
        end
    end

    return nil
end

function F.any<T>(arr: {T}, fn: function(T): boolean): boolean
    for _, v in ipairs(arr) do
        if fn(v) then
            return true
        end
    end

    return false
end

function F.all<T>(arr: {T}, fn: function(T): boolean): boolean
    for _, v in ipairs(arr) do
        if not fn(v) then
            return false
        end
    end

    return true
end

return F

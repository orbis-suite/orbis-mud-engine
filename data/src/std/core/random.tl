-- Lehmer / Park–Miller RNG
-- state = (state * 48271) % 2147483647

local record Rng
    state: integer
end

local record CoreRandom
    new: function(seed: integer): Rng
    next: function(self: Rng): integer
    nextf: function(self: Rng): number
    get_range: function(self: Rng, min: integer, max: integer): integer
end

local M: CoreRandom = {}

local A = 48271
local MMOD = 2147483647 -- 2^31 - 1

function M.new(seed: integer): Rng
    -- Park–Miller requires 1 <= state <= MMOD-1
    local s = seed % MMOD
    if s <= 0 then
        s = s + (MMOD - 1)
    end

    local rng: Rng = {
        state = s,
    }

    return rng
end

function M.next(self: Rng): integer
    self.state = (self.state * A) % MMOD
    return self.state
end

function M.nextf(self: Rng): number
    return M.next(self) / (MMOD - 1)
end

-- inclusive range
function M.get_range(self: Rng, min: integer, max: integer): integer
    if max < min then
        error("get_range: max < min")
    end

    local span = max - min + 1
    local n = M.next(self) 
    local r = n % span
    return min + r
end

return {
    Rng = Rng,
    CoreRandom = M
}

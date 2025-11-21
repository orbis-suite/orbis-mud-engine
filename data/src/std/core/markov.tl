local type Chain = { string: {string} }

local record CoreMarkov
    train: function(tokens: {string}): Chain
    generate: function(chain: Chain, length: integer): {string}
    generate_with_start: function(chain: Chain, start: string, length: integer): {string}
end

local M: CoreMarkov = {}

function M.train(tokens: {string}): Chain
  local chain: Chain = {}

  for i = 1, #tokens - 1 do
    local cur = tokens[i]
    local nxt = tokens[i + 1]

    if chain[cur] == nil then
      chain[cur] = {}
    end

    table.insert(chain[cur], nxt)
  end

  return chain
end

function M.generate(chain: Chain, length: integer): {string}
  -- randomly select a starting word from the chain
  local keys: {string} = {}

  for k, _ in pairs(chain) do
    table.insert(keys, k)
  end

  local start = keys[math.random(1, #keys)]

  return M.generate_with_start(chain, start, length)
end

function M.generate_with_start(chain: Chain, start: string, length: integer): {string}
  local out: {string} = {}
  local cur = start

  for _ = 1, length do
    table.insert(out, cur)

    local nexts = chain[cur]
    if not nexts or #nexts == 0 then
      break
    end

    -- TODO: use CoreRandom
    local idx = math.random(1, #nexts)
    cur = nexts[idx]
  end

  return out
end

return M
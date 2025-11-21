local record CoreMix
  mix: function(...: {any:any}): {any:any}
end

local M: CoreMix = {}

-- Shallow merge of tables (rightmost wins)
function M.mix(...: {any:any}): {any:any}
  local out: {any:any} = {}
  local n = select("#", ...)
  for i = 1, n do
    local t = select(i, ...) as {any:any}|nil
    if t ~= nil then
      local tt = t as {any:any}
      for k, v in pairs(tt) do
        out[k] = v
      end
    end
  end
  return out
end

return M

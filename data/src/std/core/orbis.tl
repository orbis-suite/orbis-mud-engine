require("src.std.api")

local EntitySnapshot = require("src.types.entity_snapshot")

local record CoreOrbis
  color: function(color: string, msg: string): string
  print: function(role: string, msg: string)
  publish: function(msg: string)
  publish_to_room: function(room_id: string, msg: string)

  get_entity: function(role: string): EntitySnapshot
  set_entity: function(role: string, path: string, value: any)
  get_room: function(room_id: string): EntitySnapshot
  get_param: function(param: string): string

  move_to: function(role_to_move: string, destination_role: string, children_group: string)
  move_to_room: function(role_to_move: string, destination_id: string, children_group: string)

  after: function(delay: integer, fn: function())
  repeat_while: function(delay: integer, cond: function(): boolean, fn: function())
end

local O: CoreOrbis = {}

function O.color(color: string, msg: string): string
  return API.color(color, msg)
end

function O.print(role: string, msg: string)
  API.print(role, msg)
end

function O.publish(msg: string)
  API.publish(msg)
end

function O.publish_to_room(room_id: string, msg: string)
  API.publish_to_room(room_id, msg)
end

function O.get_entity(role: string): EntitySnapshot
  return API.get_entity(role)
end

function O.get_room(room_id: string): EntitySnapshot
  return API.get_room(room_id)
end

function O.get_param(param: string): string
  return API.get_param(param)
end

function O.set_entity(role: string, path: string, value: any)
  return API.set_entity(role, path, value)
end

function O.move_to_room(role_to_move: string, destination_id: string, children_group: string)
  return API.move_to_room(role_to_move, destination_id, children_group)
end

function O.after(delay: integer, fn: function())
  API.after(delay, fn)
end

function O.repeat_while(delay: integer, cond: function(): boolean, fn: function())
  local function tick()
    if not cond() then
      return
    end

    fn()
    O.after(delay, tick)
  end

  O.after(delay, tick)
end

return O

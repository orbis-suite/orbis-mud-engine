local EntitySnapshot = require("src.types.entity_snapshot")

local record ApiDefinition
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
end

global API: ApiDefinition
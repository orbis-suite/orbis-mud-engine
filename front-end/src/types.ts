export interface RoomContent {
  name: string
  description: string
  exits: string[]
  children: { name: string; description: string }[]
}

export interface TextContent { text: string }
export interface EntityContent { name: string; description: string }
export interface InventoryContent { items: string[] }
export interface MapCell { color: string; icon: string }
export interface MapData { grid: MapCell[][]; playerX: number; playerY: number }

export type Direction = 'north' | 'south' | 'east' | 'west' | 'up' | 'down' | 'in' | 'out'

export type ClientMessage =
  | { type: 'text'; text: string }
  | { type: 'move'; direction: Direction }

export interface WSMessage {
  panel: 'main' | 'map' | 'inventory' | 'room'
  content: unknown
}

export type Phase = 'modal' | 'connecting' | 'playing'

export interface State {
  phase: Phase
  nameError: string
  lines: string[]
  room: RoomContent | null
  map: MapData | null
  inventoryOpen: boolean
  inventory: string[]
}

export type Action =
  | { type: 'connecting' }
  | { type: 'disconnected' }
  | { type: 'name_error'; error: string }
  | { type: 'message'; msg: WSMessage }
  | { type: 'close_inventory' }
  | { type: 'clear_inventory' }

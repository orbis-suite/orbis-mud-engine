package response

import "encoding/json"

const (
	PanelMain      = "main"
	PanelRoom      = "room"
	PanelMap       = "map"
	PanelInventory = "inventory"
)

// Response is the result type of world.Parse.
type Response interface {
	Panel() string
	isResponse()
}

// WSMessage is the JSON envelope sent to WebSocket clients.
// Content is json.RawMessage so it holds the marshaled concrete Response type directly.
type WSMessage struct {
	Panel   string          `json:"panel"`
	Content json.RawMessage `json:"content"`
}

// ChildSummary is a compact representation of an entity inside a room or container.
type ChildSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoomDescription is returned by look (no target), move, and the opening message.
type RoomDescription struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Exits       []string       `json:"exits"`
	Children    []ChildSummary `json:"children"`
}

func (RoomDescription) Panel() string { return PanelRoom }
func (RoomDescription) isResponse()   {}

// EntityDescription is returned by look <target>.
type EntityDescription struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Children    []ChildSummary `json:"children,omitempty"`
}

func (EntityDescription) Panel() string { return PanelMain }
func (EntityDescription) isResponse()   {}

// InventoryList is returned by the inventory command.
type InventoryList struct {
	Items []string `json:"items"`
}

func (InventoryList) Panel() string { return PanelInventory }
func (InventoryList) isResponse()   {}

// MapCell is one cell in the map grid.
type MapCell struct {
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// MapView is returned by the map command.
// Grid is row-major: Grid[y][x]. PlayerX and PlayerY are the grid coordinates
// of the player's current room.
type MapView struct {
	Grid    [][]MapCell `json:"grid"`
	PlayerX int         `json:"playerX"`
	PlayerY int         `json:"playerY"`
}

func (MapView) Panel() string { return PanelMap }
func (MapView) isResponse()   {}

// Text is the fallback for all plain-text responses: help, track, game actions, errors.
type Text struct {
	Value string `json:"text"`
}

func (Text) Panel() string { return PanelMain }
func (Text) isResponse()   {}

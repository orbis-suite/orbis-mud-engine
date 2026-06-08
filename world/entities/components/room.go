package components

import (
	"fmt"
	"strings"

	"example.com/mud/world/entities"
)

type Room struct {
	MapIcon  string
	MapColor string
	Exits    map[string]string

	children entities.IChildren
}

var _ entities.Component = &Room{}
var _ entities.ComponentWithChildren = &Room{}

func NewRoom() *Room {
	return &Room{
		MapIcon:  "O",
		children: NewChildren(),
	}
}

func (r *Room) Id() entities.ComponentType {
	return entities.ComponentRoom
}

func (r *Room) Copy() entities.Component {
	return &Room{
		MapIcon:  r.MapIcon,
		MapColor: r.MapColor,
		Exits:    r.Exits,
		children: r.children.Copy(),
	}
}

func (r *Room) AddChild(child *entities.Entity) error {
	err := r.GetChildren().AddChild(child)
	if err != nil {
		return fmt.Errorf("Inventory add child: %w", err)
	}

	child.Parent = r

	return nil
}

func (r *Room) RemoveChild(child *entities.Entity) {
	child.Parent = nil
	r.GetChildren().RemoveChild(child)
}

func (r *Room) GetChildren() entities.IChildren {
	return r.children
}

func (r *Room) GetNeighboringRoomId(direction string) (string, bool) {
	roomId, ok := r.Exits[direction]
	return roomId, ok
}

func (r *Room) GetExitText() string {
	var b strings.Builder
	b.WriteString("Exits: ")

	for exit := range r.Exits {
		b.WriteString(exit)
		b.WriteString(", ")
	}

	result := strings.TrimSuffix(b.String(), ", ")
	return result
}

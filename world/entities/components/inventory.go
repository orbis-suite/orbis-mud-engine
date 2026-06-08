package components

import (
	"fmt"
	"strings"

	"example.com/mud/world/entities"
)

type Inventory struct {
	children entities.IChildren
}

func NewInventory() *Inventory {
	return &Inventory{
		children: NewChildren(),
	}
}

var _ entities.Component = &Inventory{}
var _ entities.ComponentWithChildren = &Inventory{}

func (i *Inventory) Id() entities.ComponentType {
	return entities.ComponentInventory
}

func (i *Inventory) Copy() entities.Component {
	iCopy := &Inventory{
		children: i.children.Copy(),
	}

	for _, child := range i.children.GetChildren() {
		iCopy.AddChild(child.Copy(iCopy))
	}

	return iCopy
}

func (i *Inventory) AddChild(child *entities.Entity) error {
	err := i.GetChildren().AddChild(child)
	if err != nil {
		return fmt.Errorf("Inventory add child: %w", err)
	}

	child.Parent = i

	return nil
}

func (i *Inventory) RemoveChild(child *entities.Entity) {
	child.Parent = nil
	i.GetChildren().RemoveChild(child)
}

func (i *Inventory) GetChildren() entities.IChildren {
	return i.children
}

func (i *Inventory) Print() (string, error) {
	var b strings.Builder

	b.WriteString("You are carrying: [")

	for _, child := range i.GetChildren().GetChildren() {
		if n := child.Name; n != "" {
			b.WriteString(n)
			b.WriteString(", ")
		}
	}

	return strings.TrimSuffix(b.String(), ", ") + "]", nil
}

package components

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Container struct {
	children entities.IChildren
}

var _ entities.Component = &Container{}
var _ entities.ComponentWithChildren = &Container{}

func NewContainer() *Container {
	return &Container{
		children: NewChildren(),
	}
}

func (i *Container) Id() entities.ComponentType {
	return entities.ComponentContainer
}

func (c *Container) Copy() entities.Component {
	cCopy := &Container{
		children: c.children.Copy(),
	}

	for _, child := range c.children.GetChildren() {
		cCopy.AddChild(child)
	}

	return cCopy
}

func (c *Container) AddChild(child *entities.Entity) error {
	err := c.GetChildren().AddChild(child)
	if err != nil {
		return fmt.Errorf("Container add child: %w", err)
	}

	child.Parent = c

	return nil
}

func (c *Container) RemoveChild(child *entities.Entity) {
	child.Parent = nil
	c.GetChildren().RemoveChild(child)
}

func (c *Container) GetChildren() entities.IChildren {
	return c.children
}

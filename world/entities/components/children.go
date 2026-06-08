package components

import (
	"fmt"

	"example.com/mud/world/entities"
)

type Children struct {
	revealed bool
	prefix   string

	// TODO rename child -> children
	childByAlias   map[string][]*entities.Entity
	aliasesByChild map[*entities.Entity][]string
}

var _ entities.IChildren = &Children{}

func NewChildren() *Children {
	return &Children{
		childByAlias:   make(map[string][]*entities.Entity),
		aliasesByChild: make(map[*entities.Entity][]string),
	}
}

// copy handles fields, but not children
func (c *Children) Copy() entities.IChildren {
	copiedChildren := NewChildren()

	copiedChildren.revealed = c.revealed
	copiedChildren.prefix = c.prefix

	return copiedChildren
}

func (c *Children) GetPrefix() string {
	return c.prefix
}

func (c *Children) GetRevealed() bool {
	return c.revealed
}

func (c *Children) SetPrefix(p string) {
	c.prefix = p
}

func (c *Children) SetRevealed(r bool) {
	c.revealed = r
}

func (c *Children) AddChild(child *entities.Entity) error {
	aliases := child.Aliases

	if len(aliases) == 0 {
		return nil
	}
	for _, alias := range aliases {
		c.aliasesByChild[child] = append(c.aliasesByChild[child], alias)
		c.childByAlias[alias] = append(c.childByAlias[alias], child)
	}

	return nil
}

func (c *Children) RemoveChild(child *entities.Entity) {
	aliases, ok := c.aliasesByChild[child]
	if !ok {
		return
	}

	for _, alias := range aliases {
		// NOTE: This is potentially slow, consider making childrenByAlias a map[string]map[*Entity]struct{}
		oldEntities := c.childByAlias[alias]
		newEntities := make([]*entities.Entity, 0, len(oldEntities))
		for _, oe := range oldEntities {
			if oe != child {
				newEntities = append(newEntities, oe)
			}
		}
		c.childByAlias[alias] = newEntities
	}
	delete(c.aliasesByChild, child) // delete entry from aliasesByItem
}

func (c *Children) GetChildren() []*entities.Entity {
	children := make([]*entities.Entity, 0)
	for child := range c.aliasesByChild {
		children = append(children, child)
	}
	return children
}

func (c *Children) GetChildrenByAlias(alias string) []entities.AmbiguityOption {
	eMatches := make([]entities.AmbiguityOption, 0, 10)

	children := c.childByAlias[alias]
	for _, child := range children {
		eMatches = append(eMatches, entities.AmbiguityOption{
			Text:   fmt.Sprintf("%s: %s", c.GetPrefix(), child.Name),
			Entity: child,
		})
	}

	for _, children := range c.GetChildren() {
		for _, cwc := range children.GetComponentsWithChildren() {
			if !cwc.GetChildren().GetRevealed() {
				continue
			}

			grandchildren := cwc.GetChildren().GetChildrenByAlias(alias)
			if len(grandchildren) > 0 {
				eMatches = append(eMatches, grandchildren...)
			}
		}
	}

	return eMatches
}

func (c *Children) HasChild(e *entities.Entity) bool {
	_, ok := c.aliasesByChild[e]

	return ok
}

func (c *Children) ReindexAliasesForEntity(e *entities.Entity) error {
	c.RemoveChild(e)
	err := c.AddChild(e)
	if err != nil {
		return fmt.Errorf("error reindexing aliases for entity '%s': %w", e.Name, err)
	}

	return nil
}

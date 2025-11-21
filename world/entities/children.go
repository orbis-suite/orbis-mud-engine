package entities

import (
	"fmt"
)

type IChildren interface {
	Copy() IChildren

	SetRevealed(r bool)
	GetRevealed() bool
	SetPrefix(p string)
	GetPrefix() string

	AddChild(child *Entity) error
	RemoveChild(child *Entity)
	GetChildren() []*Entity
	GetChildrenByAlias(alias string) []AmbiguityOption
	HasChild(e *Entity) bool
	ReindexAliasesForEntity(e *Entity) error
}

type Children struct {
	revealed bool
	prefix   string

	// TODO rename child -> children
	childByAlias   map[string][]*Entity
	aliasesByChild map[*Entity][]string
}

var _ IChildren = &Children{}

func NewChildren() *Children {
	return &Children{
		childByAlias:   make(map[string][]*Entity),
		aliasesByChild: make(map[*Entity][]string),
	}
}

// copy handles fields, but not children
func (c *Children) Copy() IChildren {
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

func (c *Children) AddChild(child *Entity) error {
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

func (c *Children) RemoveChild(child *Entity) {
	aliases, ok := c.aliasesByChild[child]
	if !ok {
		return
	}

	for _, alias := range aliases {
		// NOTE: This is potentially slow, consider making childrenByAlias a map[string]map[*Entity]struct{}
		oldEntities := c.childByAlias[alias]
		newEntities := make([]*Entity, 0, len(oldEntities))
		for _, oe := range oldEntities {
			if oe != child {
				newEntities = append(newEntities, oe)
			}
		}
		c.childByAlias[alias] = newEntities
	}
	delete(c.aliasesByChild, child) // delete entry from aliasesByItem
}

func (c *Children) GetChildren() []*Entity {
	children := make([]*Entity, 0)
	for child := range c.aliasesByChild {
		children = append(children, child)
	}
	return children
}

func (c *Children) GetChildrenByAlias(alias string) []AmbiguityOption {
	eMatches := make([]AmbiguityOption, 0, 10)

	// get children
	children := c.childByAlias[alias]
	for _, child := range children {
		eMatches = append(eMatches, AmbiguityOption{
			Text:   fmt.Sprintf("%s: %s", c.GetPrefix(), child.Name),
			Entity: child,
		})
	}

	// recursively get children of children
	for _, children := range c.GetChildren() {
		eMatches = append(
			eMatches,
			children.GetChildrenByAlias(alias)...,
		)
	}

	return eMatches
}

func (c *Children) HasChild(e *Entity) bool {
	_, ok := c.aliasesByChild[e]

	return ok
}

func (c *Children) ReindexAliasesForEntity(e *Entity) error {
	c.RemoveChild(e)
	err := c.AddChild(e)
	if err != nil {
		return fmt.Errorf("error reindexing aliases for entity '%s': %w", e.Name, err)
	}

	return nil
}

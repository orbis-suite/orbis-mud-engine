package entities

import (
	"fmt"
)

type ComponentType int

const (
	ComponentUnknown ComponentType = iota
	ComponentRoom
	ComponentEventful
	ComponentInventory
	ComponentContainer
)

const (
	ComponentUnknownString   = "Unknown"
	ComponentRoomString      = "Room"
	ComponentEventfulString  = "Eventful"
	ComponentInventoryString = "Inventory"
	ComponentContainerString = "Container"
)

func ParseComponentType(s string) (ComponentType, error) {
	switch s {
	case ComponentRoomString:
		return ComponentRoom, nil
	case ComponentEventfulString:
		return ComponentEventful, nil
	case ComponentInventoryString:
		return ComponentInventory, nil
	case ComponentContainerString:
		return ComponentContainer, nil
	default:
		return ComponentUnknown, fmt.Errorf("unknown component type '%s'", s)
	}
}

func (ct ComponentType) String() string {
	switch ct {
	case ComponentRoom:
		return ComponentRoomString
	case ComponentEventful:
		return ComponentEventfulString
	case ComponentInventory:
		return ComponentInventoryString
	case ComponentContainer:
		return ComponentContainerString
	default:
		return ComponentUnknownString
	}
}

type Component interface {
	Id() ComponentType
	Copy() Component
}

type ComponentWithChildren interface {
	AddChild(child *Entity) error
	RemoveChild(child *Entity)

	GetChildren() IChildren
}

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

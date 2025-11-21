// ir/entity.go
package ir

import (
	"fmt"
	"maps"

	"example.com/mud/world/entities"
)

type EntityIR struct {
	Name        string
	Description string
	Aliases     []string
	Tags        []string
	Fields      map[string]any
	InitFunc    entities.ReactionFunc
	Reactions   map[string]map[entities.EventRole]entities.ReactionFunc
	Children    map[string][]*EntityIR
}

func (e *EntityIR) Build() (*entities.Entity, error) {
	ent := entities.NewEntity(
		e.Name,
		e.Description,
		e.Aliases,
		e.Tags,
		map[string]any{},
		e.InitFunc,
		nil,
	)

	ent.InitFunc = e.InitFunc

	// attach reactions
	for kind, roleReactions := range e.Reactions {
		for role, fn := range roleReactions {
			ent.AddReaction(kind, role, fn)
		}
	}

	// attach fields
	maps.Copy(ent.Fields, e.Fields)

	for group, children := range e.Children {
		for _, child := range children {
			childEntity, err := child.Build()
			if err != nil {
				return nil, fmt.Errorf("could not build child for entity '%s': %w", e.Name, err)
			}

			ent.AddChild(group, childEntity)
		}
	}

	return ent, nil
}

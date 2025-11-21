package entities

import (
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type ReactionFunc func(*Event)
type Entity struct {
	mu        sync.RWMutex
	reactions map[string]map[EventRole]ReactionFunc
	children  map[string]IChildren

	Id          string
	Name        string
	Description string
	Aliases     []string
	Tags        []string
	Fields      map[string]any
	InitFunc    ReactionFunc
	Parent      *Entity
}

func NewEntity(name, description string, aliases []string, tags []string, fields map[string]any, initFunc ReactionFunc, parent *Entity) *Entity {
	return &Entity{
		reactions: map[string]map[EventRole]ReactionFunc{},
		children:  map[string]IChildren{},

		Id:          uuid.NewString(),
		Name:        name,
		Description: description,
		Aliases:     aliases,
		Tags:        tags,
		Fields:      fields,
		InitFunc:    initFunc,
		Parent:      parent,
	}
}

func (e *Entity) Copy(parent *Entity) *Entity {
	fieldsCopy := make(map[string]any, len(e.Fields))
	maps.Copy(fieldsCopy, e.Fields)

	aliasesCopy := append([]string(nil), e.Aliases...)
	tagsCopy := append([]string(nil), e.Tags...)

	newEntity := NewEntity(
		e.Name,
		e.Description,
		aliasesCopy,
		tagsCopy,
		fieldsCopy,
		e.InitFunc,
		parent,
	)

	maps.Copy(newEntity.reactions, e.reactions)

	return newEntity
}

func (e *Entity) SetField(path string, value any) error {
	if e == nil || path == "" {
		return fmt.Errorf("entity set field: path cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}

	parts := strings.Split(path, ".")
	m := e.Fields

	// walk through nested maps
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]

		next, ok := m[key]
		if !ok {
			nm := make(map[string]any)
			m[key] = nm
			m = nm
			continue
		}

		nm, ok := next.(map[string]any)
		if !ok {
			// stomp over existing non-map field
			nm = make(map[string]any)
			m[key] = nm
		}

		m = nm
	}

	last := parts[len(parts)-1]
	m[last] = value

	return nil
}

func (e *Entity) GetField(fieldName string) any {
	return e.Fields[fieldName]
}

func (e *Entity) setAliases(aliases []string) error {
	// TODO
	// e.Aliases = aliases

	// // entities are indexed by aliases for performance reasons, so we need to reindex
	// err := e.Parent.GetChildren().ReindexAliasesForEntity(e)
	// if err != nil {
	// 	return fmt.Errorf("error setting aliases for entity '%s': %w", e.Name, err)
	// }

	// return nil
	return nil
}

func (e *Entity) AddReaction(kind string, role EventRole, r ReactionFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.reactions[kind] == nil {
		e.reactions[kind] = map[EventRole]ReactionFunc{}
	}
	e.reactions[kind][role] = r
}

func (e *Entity) GetReaction(kind string, role EventRole) (ReactionFunc, bool) {
	r, ok := e.reactions[kind][role]
	return r, ok
}

func (e *Entity) AddChild(group string, child *Entity) error {
	if e.children[group] == nil {
		e.children[group] = NewChildren()
	}

	err := e.children[group].AddChild(child)
	if err != nil {
		return fmt.Errorf("could not add '%s' group child: %w", group, err)
	}

	child.Parent = e

	return nil
}

func (e *Entity) GetChildren() map[string][]*Entity {
	out := map[string][]*Entity{}

	for group, groupChildren := range e.children {
		groupEntities := []*Entity{}
		groupEntities = append(groupEntities, groupChildren.GetChildren()...)

		out[group] = groupEntities
	}

	return out
}

func (e *Entity) GetChildrenByAlias(alias string) []AmbiguityOption {
	eMatches := make([]AmbiguityOption, 0, 10)

	for _, children := range e.children {
		eMatches = append(
			eMatches,
			children.GetChildrenByAlias(alias)...,
		)
	}

	return eMatches
}

func (e *Entity) GetDescription() (string, error) {
	// var b strings.Builder

	// formatted, err := utils.FormatText(e.Description, map[string]string{})
	// if err != nil {
	// 	return "", fmt.Errorf("could not format description for entity '%s': %w", e.Name, err)
	// }

	// b.WriteString(fmt.Sprintf("- %s", formatted))

	// for _, cwc := range e.GetComponentsWithChildren() {
	// 	if !cwc.GetChildren().GetRevealed() {
	// 		continue
	// 	}

	// 	children := cwc.GetChildren().GetChildren()
	// 	if len(children) == 0 {
	// 		continue
	// 	}

	// 	var childB strings.Builder
	// 	childB.WriteString("\n")

	// 	childB.WriteString(fmt.Sprintf("%s%s:", models.Tab, cwc.GetChildren().GetPrefix()))
	// 	childB.WriteString(" (\n")

	// 	for _, child := range children {
	// 		cDescription, err := child.GetDescription()
	// 		if err != nil {
	// 			return "", fmt.Errorf("could not format description for entity '%s': %w", child.Name, err)
	// 		}

	// 		childB.WriteString(fmt.Sprintf("%s%s%s", models.Tab, models.Tab, cDescription))
	// 		childB.WriteString("\n")
	// 	}

	// 	b.WriteString(childB.String())

	// 	b.WriteString(fmt.Sprintf("%s)", models.Tab))
	// }

	// return b.String(), nil
	return "", nil
}

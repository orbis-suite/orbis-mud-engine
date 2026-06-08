package dsl

import (
	"fmt"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
)

type ComponentDef struct {
	Name   string      `parser:"@Ident"`
	Fields []*FieldDef `parser:"'{' { @@ } '}'"`
}

type componentBuilder func(def *ComponentDef) (entities.Component, error)

var componentBuilders = map[string]componentBuilder{}

func registerComponentBuilder(name string, b componentBuilder) {
	componentBuilders[name] = b
}

func init() {
	registerComponentBuilder("Room", buildRoom)
	registerComponentBuilder("Inventory", buildInventory)
	registerComponentBuilder("Container", buildContainer)
}

func (def *ComponentDef) Build() (entities.Component, error) {
	if b, ok := componentBuilders[def.Name]; ok {
		return b(def)
	}
	return nil, fmt.Errorf("could not match component name %s", def.Name)
}

func buildRoom(def *ComponentDef) (entities.Component, error) {
	rm := components.NewRoom()
	rm.GetChildren().SetPrefix("In the room")

	for _, f := range def.Fields {
		// exits are the only place where maps are supported for now
		// TODO support maps elsewhere
		if f.Key == "exits" {
			m := f.Value.AsMap()
			if m == nil {
				m = map[string]string{}
			}
			rm.Exits = m
			continue
		}

		value, err := immediateEvalExpression(f.Value)
		if err != nil {
			return nil, fmt.Errorf("could not get value '%s' for Room: %w", f.Key, err)
		}
		switch f.Key {
		case "prefix":
			if value.K != models.KindString {
				return nil, fmt.Errorf("room: prefix must be string")
			}
			rm.GetChildren().SetPrefix(value.S)
		case "icon":
			if value.K != models.KindString {
				return nil, fmt.Errorf("room: icon must be string")
			}

			if len(value.S) != 1 {
				return nil, fmt.Errorf("invalid map icon '%s': must be 1 character", value.S)
			}
			rm.MapIcon = value.S
		case "color":
			if value.K != models.KindString {
				return nil, fmt.Errorf("room: color must be string")
			}

			rm.MapColor = value.S
		case "children":
			continue
		default:
			return nil, fmt.Errorf("room: unknown field %s", f.Key)
		}
	}
	return rm, nil
}

func buildInventory(def *ComponentDef) (entities.Component, error) {
	inventory := components.NewInventory()
	for _, f := range def.Fields {
		value, err := immediateEvalExpression(f.Value)
		if err != nil {
			return nil, fmt.Errorf("could not get value '%s' for Room: %w", f.Key, err)
		}

		switch f.Key {
		case "prefix":
			if value.K != models.KindString {
				return nil, fmt.Errorf("inventory: prefix must be string")
			}
			inventory.GetChildren().SetPrefix(value.S)
		case "revealed":
			if value.K != models.KindBool {
				return nil, fmt.Errorf("inventory: revealed must be a boolean")
			}
			inventory.GetChildren().SetRevealed(value.B)
		case "children":
			continue
		default:
			return nil, fmt.Errorf("inventory: unknown field %s", f.Key)
		}
	}
	return inventory, nil
}

func buildContainer(def *ComponentDef) (entities.Component, error) {
	container := components.NewContainer()
	for _, f := range def.Fields {
		value, err := immediateEvalExpression(f.Value)
		if err != nil {
			return nil, fmt.Errorf("could not get value '%s' for Room: %w", f.Key, err)
		}

		switch f.Key {
		case "prefix":
			if value.K != models.KindString {
				return nil, fmt.Errorf("container: prefix must be string")
			}
			container.GetChildren().SetPrefix(value.S)
		case "revealed":
			if value.K != models.KindBool {
				return nil, fmt.Errorf("container: revealed must be a boolean")
			}
			container.GetChildren().SetRevealed(value.B)
		case "children":
			continue
		default:
			return nil, fmt.Errorf("room: unknown field %s", f.Key)
		}
	}
	return container, nil
}

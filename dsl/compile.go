package dsl

import (
	"fmt"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
)

type collectedDefs struct {
	entitiesById map[string]EntityDef
	traitsById   map[string]TraitDef
	commandsById map[string]CommandDef
}

type ChildrenPlan map[string]map[entities.ComponentType][]string

type LoweredEntity struct {
	name           string
	description    string
	tags           []string
	aliases        []string
	components     []entities.Component
	fields         map[string]models.Value
	rulesByCommand map[string][]*entities.Rule
}

type entityPrototype struct {
	id  string
	ent *entities.Entity
	def EntityDef
}

type entityPrototypes struct {
	prototypesById map[string]*entityPrototype
	traitsById     map[string]TraitDef
	childrenPlan   ChildrenPlan
	visiting       map[string]struct{}
}

func Compile(ast *DSL) (map[string]*entities.Entity, []*models.CommandDefinition, error) {
	if ast == nil {
		return nil, nil, fmt.Errorf("nil DSL")
	}

	collectedDefs, err := collectDefs(ast.Declarations)
	if err != nil {
		return nil, nil, fmt.Errorf("could not collect top level declarations: %w", err)
	}

	prototypes, err := collectedDefs.collectPrototypes()
	if err != nil {
		return nil, nil, fmt.Errorf("could not collect prototype entities: %w", err)
	}

	entitiesById, err := prototypes.instantiatePrototypes()
	if err != nil {
		return nil, nil, fmt.Errorf("could not instantiate prototype entities: %w", err)
	}

	commands := make([]*models.CommandDefinition, 0, len(collectedDefs.commandsById))
	for _, c := range collectedDefs.commandsById {
		cd, err := c.Build()
		if err != nil {
			return nil, nil, fmt.Errorf("could not instantiate command '%s': %w", c.Name, err)
		}

		commands = append(commands, cd)
	}

	return entitiesById, commands, nil
}

// collect entity, command and trait definitions
func collectDefs(decls []*TopLevel) (*collectedDefs, error) {
	entitiesById := make(map[string]EntityDef, len(decls))
	commandsById := make(map[string]CommandDef, len(decls))
	traitsById := make(map[string]TraitDef, len(decls))

	for _, declaration := range decls {
		if declaration == nil {
			return nil, fmt.Errorf("declaration at top level is nil")
		}

		if ed := declaration.Entity; ed != nil {
			if _, exists := entitiesById[ed.Name]; exists {
				return nil, fmt.Errorf("duplicate entity %s", ed.Name)
			}

			entitiesById[ed.Name] = *ed
		} else if td := declaration.Trait; td != nil {
			if _, exists := entitiesById[td.Name]; exists {
				return nil, fmt.Errorf("duplicate trait %s", td.Name)
			}

			traitsById[declaration.Trait.Name] = *declaration.Trait
		} else if ec := declaration.Command; ec != nil {
			if _, exists := commandsById[ec.Name]; exists {
				return nil, fmt.Errorf("duplicate command %s", ec.Name)
			}

			commandsById[declaration.Command.Name] = *declaration.Command
		} else {
			return nil, fmt.Errorf("declaration at top level is empty")
		}
	}

	return &collectedDefs{
		entitiesById: entitiesById,
		traitsById:   traitsById,
		commandsById: commandsById,
	}, nil
}

// expand traits in each entity definition
func (c *collectedDefs) collectPrototypes() (*entityPrototypes, error) {
	ep := &entityPrototypes{
		prototypesById: map[string]*entityPrototype{},
		traitsById:     c.traitsById,
		childrenPlan:   map[string]map[entities.ComponentType][]string{},
		visiting:       map[string]struct{}{},
	}

	// build prototypes of each entity and put them in name->builtEntity map
	for name, ed := range c.entitiesById {
		// build prototype and populate pending children
		prototypeEntity, err := ep.buildPrototype(name, ed.Blocks)
		if err != nil {
			return nil, fmt.Errorf("build %s: %w", name, err)
		}
		ep.prototypesById[name] = &entityPrototype{
			id:  name,
			ent: prototypeEntity,
			def: ed,
		}
	}

	return ep, nil
}

// create prototype entity with components. collect child prototype names into the sidecar for later.
func (ep *entityPrototypes) buildPrototype(id string, blocks []*EntityBlock) (*entities.Entity, error) {

	loweredEntity, err := ep.lowerEntity(id, blocks)
	if err != nil {
		return nil, fmt.Errorf("could not build prototype: %w", err)
	}

	e := entities.NewEntity(
		loweredEntity.name,
		loweredEntity.description,
		loweredEntity.aliases,
		loweredEntity.tags,
		loweredEntity.fields,
		nil,
	)

	for _, c := range loweredEntity.components {
		e.Add(c)
	}

	if len(loweredEntity.rulesByCommand) > 0 {
		// create eventful if it doesn't already exist
		eventful, ok := entities.GetComponent[*components.Eventful](e)
		if !ok {
			eventful = &components.Eventful{
				Rules: map[string][]*entities.Rule{},
			}
			e.Add(eventful)
		}

		for command, rulesByCommand := range loweredEntity.rulesByCommand {
			for _, r := range rulesByCommand {
				eventful.AddRule(command, r)
			}
		}
	}

	for _, block := range blocks {
		if block.Component == nil {
			continue
		}

		for _, f := range block.Component.Fields {
			if f.Key == "children" {
				if ep.childrenPlan[id] == nil {
					ep.childrenPlan[id] = make(map[entities.ComponentType][]string)
				}

				// populate pending children map
				componentType, err := entities.ParseComponentType(block.Component.Name)
				if err != nil {
					return nil, fmt.Errorf("could not build prototype '%s': %w", id, err)
				}

				// get list of strings from expression
				childrenStrings, err := immediateEvalExpressionAs(f.Value, models.KindStringList)
				if err != nil {
					return nil, fmt.Errorf("could not get children list for prototype '%s': %w", id, err)
				}

				ep.childrenPlan[id][componentType] =
					append(ep.childrenPlan[id][componentType], childrenStrings.SL...)
			}
		}
	}

	return e, nil
}

// recursively expand traits in entities
func (ep *entityPrototypes) lowerEntity(id string, blocks []*EntityBlock) (*LoweredEntity, error) {
	if _, ok := ep.visiting[id]; ok {
		return nil, fmt.Errorf("cycle detected at %q", id)
	}
	ep.visiting[id] = struct{}{}
	defer func() { delete(ep.visiting, id) }()

	var name string
	var description string
	var aliases []string
	var tags []string
	fields := make(map[string]models.Value)

	components := make([]entities.Component, 0, len(blocks))
	rulesByCommand := make(map[string][]*entities.Rule, len(blocks))

	for _, block := range blocks {
		if block.Reaction != nil {
			// process reaction
			rules, err := block.Reaction.Build()
			if err != nil {
				return nil, err
			}
			// rules at the entity level come first
			for _, command := range block.Reaction.Commands {
				rulesByCommand[command] = append(rules, rulesByCommand[command]...)
			}
		} else if block.Component != nil {
			// process component into prototype without children
			comp, err := block.Component.Build()
			if err != nil {
				return nil, fmt.Errorf("could not process component %s: %w", block.Component.Name, err)
			}
			components = append(components, comp)
		} else if block.Trait != nil {
			// TODO this dereferences a nil pointer if the trait doesn't exist
			loweredTrait, err := ep.lowerEntity(block.Trait.Name, ep.traitsById[block.Trait.Name].Blocks)
			if err != nil {
				return nil, fmt.Errorf("could not process trait '%s': %w", block.Trait.Name, err)
			}

			// first write over fields that were passed into trait
			for _, f := range block.Trait.Fields {
				value, err := immediateEvalExpression(f.Value)
				if err != nil {
					return nil, fmt.Errorf("could not get process trait '%s' field '%s': %w", block.Trait.Name, f.Key, err)
				}

				// only include fields passed into trait that aren't already defined
				if _, ok := fields[f.Key]; !ok {
					fields[f.Key] = value
				}
			}

			// first add fields that were inherited from trait
			for k, tf := range loweredTrait.fields {
				// only include fields from trait that aren't already defined
				if _, ok := fields[k]; !ok {
					fields[k] = tf
				}
			}

			components = append(components, loweredTrait.components...)
			for command, traitRules := range loweredTrait.rulesByCommand {
				// rules at the trait level come second
				rulesByCommand[command] = append(rulesByCommand[command], traitRules...)
			}

		} else if block.Field != nil {
			f := block.Field
			value, err := immediateEvalExpression(block.Field.Value)
			if err != nil {
				return nil, fmt.Errorf("could not get process field '%s' for entity '%s': %w", block.Field.Key, id, err)
			}

			switch f.Key {
			case "name":
				if value.K != models.KindString {
					return nil, fmt.Errorf("name must be a string")
				}
				name = value.S
			case "description":
				if value.K != models.KindString {
					return nil, fmt.Errorf("description must be a string")
				}
				description = value.S
			case "aliases":
				if value.K != models.KindStringList {
					return nil, fmt.Errorf("aliases must be a string list")
				}
				aliases = value.SL
			case "tags":
				if value.K != models.KindStringList {
					return nil, fmt.Errorf("tags must be a string list")
				}
				tags = value.SL
			default:
				fields[f.Key] = value
			}
		} else {
			return nil, fmt.Errorf("could not expand empty entity block")
		}
	}

	// only do verification if at a top level entity
	if len(ep.visiting) == 1 {
		// verify name, description, and aliases are set. Empty tags is ok
		if name == "" {
			return nil, fmt.Errorf("entity '%s' has no name", id)
		}
		if description == "" {
			return nil, fmt.Errorf("entity '%s' has no description", id)
		}
		if len(aliases) == 0 {
			return nil, fmt.Errorf("entity '%s' has no aliases", id)
		}
	}

	return &LoweredEntity{
		name:           name,
		description:    description,
		tags:           tags,
		aliases:        aliases,
		components:     components,
		fields:         fields,
		rulesByCommand: rulesByCommand,
	}, nil
}

// loop through prototypes and instantiate them into a map of entities by name
func (ep *entityPrototypes) instantiatePrototypes() (map[string]*entities.Entity, error) {
	out := make(map[string]*entities.Entity, len(ep.prototypesById))
	for name := range ep.prototypesById {
		entity, err := ep.instantiate(name, nil)
		if err != nil {
			return nil, fmt.Errorf("could not instantiate '%s': %w", name, err)
		}
		out[name] = entity
	}
	return out, nil
}

// recursively instantiate a named prototype and wire up children for all child-holding components.
func (ep *entityPrototypes) instantiate(id string, parent entities.ComponentWithChildren) (*entities.Entity, error) {
	be, ok := ep.prototypesById[id]
	if !ok {
		return nil, fmt.Errorf("unknown prototype %q", id)
	}
	if _, ok := ep.visiting[id]; ok {
		return nil, fmt.Errorf("cycle detected at %q", id)
	}
	ep.visiting[id] = struct{}{}
	defer func() { delete(ep.visiting, id) }()

	inst := be.ent.Copy(parent)

	// for each child-capable component on the entity, look up its pending child names from the prototype’s sidecar and attach recursively.
	if rm, ok := entities.GetComponent[*components.Room](inst); ok {
		slot := ep.childrenPlan[id][entities.ComponentRoom]
		if len(slot) > 0 && len(rm.GetChildren().GetChildren()) == 0 {
			for _, childName := range slot {
				childInst, err := ep.instantiate(childName, rm)
				if err != nil {
					return nil, err
				}
				rm.AddChild(childInst)
			}
		}
	}

	if inventory, ok := entities.GetComponent[*components.Inventory](inst); ok {
		slot := ep.childrenPlan[id][entities.ComponentInventory]
		if len(slot) > 0 && len(inventory.GetChildren().GetChildren()) == 0 {
			for _, childName := range slot {
				childInst, err := ep.instantiate(childName, inventory)
				if err != nil {
					return nil, err
				}
				inventory.AddChild(childInst)
			}
		}
	}

	if container, ok := entities.GetComponent[*components.Container](inst); ok {
		slot := ep.childrenPlan[id][entities.ComponentContainer]
		if len(slot) > 0 && len(container.GetChildren().GetChildren()) == 0 {
			for _, childName := range slot {
				childInst, err := ep.instantiate(childName, container)
				if err != nil {
					return nil, err
				}
				container.AddChild(childInst)
			}
		}
	}

	return inst, nil
}

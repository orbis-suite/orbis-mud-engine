package lua_runtime

import (
	"errors"
	"fmt"
	"strings"

	"example.com/mud/lua_runtime/ir"
	"example.com/mud/models"
	"example.com/mud/world/entities"
	lua "github.com/yuin/gopher-lua"
)

func (lr *LuaRuntime) LoadFile(path string) (map[string]*entities.Entity, []*models.CommandDefinition, error) {
	if err := lr.L.DoFile(path); err != nil {
		return nil, nil, fmt.Errorf("load lua file: %w", err)
	}

	val := lr.L.Get(-1)
	lr.L.Pop(1)

	root, ok := val.(*lua.LTable)
	if !ok {
		return nil, nil, fmt.Errorf("file %s did not return a table", path)
	}

	// build entities
	entitiesTable := tryGetTable(root, "entities")
	if entitiesTable == nil {
		return nil, nil, fmt.Errorf("expected a table field 'entities' at top level")
	}

	entityMap, err := lr.buildEntityMap(entitiesTable)
	if err != nil {
		return nil, nil, fmt.Errorf("could not build entities: %w", err)
	}

	// build commands
	commandsTable := tryGetTable(root, "commands")
	if commandsTable == nil {
		return nil, nil, fmt.Errorf("expected a table field 'commands' at top level")
	}

	commands, err := lr.buildCommands(commandsTable)
	if err != nil {
		return nil, nil, fmt.Errorf("could not build commands: %w", err)
	}

	return entityMap, commands, nil
}

func tryGetTable(t *lua.LTable, key string) *lua.LTable {
	if v := t.RawGetString(key); v != lua.LNil {
		if tbl, ok := v.(*lua.LTable); ok {
			return tbl
		}
	}
	return nil
}

func (lr *LuaRuntime) buildEntityMap(entitiesTable *lua.LTable) (map[string]*entities.Entity, error) {
	entityMap := make(map[string]*entities.Entity)
	var errs []error

	// Build each entity under entities/Entities
	entitiesTable.ForEach(func(k, v lua.LValue) {
		t, ok := v.(*lua.LTable)
		if !ok {
			errs = append(errs, fmt.Errorf("entity '%s' is not a table", lua.LVAsString(k)))
			return
		}

		id := lua.LVAsString(k)

		entityIR, err := lr.buildEntityIR(t)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not build entity IR for '%s': %w", id, err))
			return
		}

		runtimeEntity, err := entityIR.Build()
		if err != nil {
			errs = append(errs, fmt.Errorf("could not build runtime entity for '%s': %w", id, err))
			return
		}

		entityMap[id] = runtimeEntity
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return entityMap, nil
}

func (lr *LuaRuntime) buildEntityIR(val *lua.LTable) (*ir.EntityIR, error) {
	if val == nil {
		return nil, fmt.Errorf("could not build nil entity IR")
	}

	e := &ir.EntityIR{
		Name:        getString(val, "name"),
		Description: getString(val, "description"),
		Aliases:     getStringArray(val, "aliases"),
		Tags:        getStringArray(val, "tags"),
	}

	// build init func
	if v, ok := val.RawGetString("init").(*lua.LFunction); ok {
		e.InitFunc = lr.wrapLuaReaction(v)
	}

	// build reactions
	if reactionTable := tryGetTable(val, "reactions"); reactionTable != nil {
		reactions, err := lr.buildReactions(reactionTable)
		if err != nil {
			return nil, fmt.Errorf("could not build reactions for entity: %w", err)
		}

		e.Reactions = reactions
	}

	// build fields
	if fieldsTable, ok := val.RawGetString("fields").(*lua.LTable); ok {
		fields, err := luaTableToMap(fieldsTable)
		if err != nil {
			return nil, fmt.Errorf("could not build fields for entity: %w", err)
		}
		e.Fields = fields
	}

	// build children
	if childrenTable, ok := val.RawGetString("children").(*lua.LTable); ok {
		children, err := lr.buildChildren(childrenTable)
		if err != nil {
			return nil, fmt.Errorf("could not build children for entity: %w", err)
		}
		e.Children = children
	}

	return e, nil
}

func (lr *LuaRuntime) buildReactions(reactionsTable *lua.LTable) (map[string]map[entities.EventRole]entities.ReactionFunc, error) {
	reacts := make(map[string]map[entities.EventRole]entities.ReactionFunc)
	var errs []error

	reactionsTable.ForEach(func(k, v lua.LValue) {
		commandName := lua.LVAsString(k)

		roleTable, ok := v.(*lua.LTable)
		if !ok {
			errs = append(errs, fmt.Errorf("reaction '%s' is not a table of roles", commandName))
			return
		}

		roleMap := make(map[entities.EventRole]entities.ReactionFunc)

		roleTable.ForEach(func(roleKey, roleVal lua.LValue) {
			roleStr := strings.ToLower(lua.LVAsString(roleKey))

			fn, ok := roleVal.(*lua.LFunction)
			if !ok {
				errs = append(errs, fmt.Errorf("reaction '%s' for role '%s' is not a function", commandName, roleStr))
				return
			}

			eventRole, err := entities.ParseEventRole(roleStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("reaction '%s': %w", commandName, err))
				return
			}

			roleMap[eventRole] = lr.wrapLuaReaction(fn)
		})

		if len(roleMap) > 0 {
			reacts[commandName] = roleMap
		}
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return reacts, nil
}

func (lr *LuaRuntime) buildChildren(childrenTable *lua.LTable) (map[string][]*ir.EntityIR, error) {
	result := make(map[string][]*ir.EntityIR)
	var errs []error

	childrenTable.ForEach(func(groupKey, groupVal lua.LValue) {
		groupName := lua.LVAsString(groupKey)
		if groupName == "" {
			errs = append(errs, fmt.Errorf("children: empty group name"))
			return
		}

		arr, ok := groupVal.(*lua.LTable)
		if !ok {
			errs = append(errs, fmt.Errorf("children group '%s' is not an array table", groupName))
			return
		}

		inlineChildren, err := lr.parseInlineChildren(arr)
		if err != nil {
			errs = append(errs, fmt.Errorf("children group '%s': %w", groupName, err))
			return
		}

		if len(inlineChildren) > 0 {
			result[groupName] = inlineChildren
		}
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return result, nil
}

func (lr *LuaRuntime) parseInlineChildren(arr *lua.LTable) ([]*ir.EntityIR, error) {
	children := make([]*ir.EntityIR, 0, arr.Len())
	arr.ForEach(func(_, child lua.LValue) {
		ct, ok := child.(*lua.LTable)
		if !ok {
			// TODO
			return
		}
		childIR, err := lr.buildEntityIR(ct)
		if err != nil {
			// TODO
			return
		}

		children = append(children, childIR)
	})

	return children, nil
}

func (lr *LuaRuntime) buildCommands(commandsTable *lua.LTable) ([]*models.CommandDefinition, error) {
	commandDefinitions := []*models.CommandDefinition{}
	var errs []error

	commandsTable.ForEach(func(k, v lua.LValue) {
		t, ok := v.(*lua.LTable)
		if !ok {
			errs = append(errs, fmt.Errorf("command '%s' is not a table", lua.LVAsString(k)))
			return
		}

		id := lua.LVAsString(k)

		patternsTable := tryGetTable(t, "patterns")
		if patternsTable == nil {
			errs = append(errs, fmt.Errorf("expected a table field 'patterns' for command '%s'", id))
			return
		}

		patterns, err := lr.buildCommandPatterns(patternsTable)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not build patterns for command '%s': %w", id, err))
			return
		}

		commandDefinitions = append(commandDefinitions, &models.CommandDefinition{
			Name:     id,
			Aliases:  getStringArray(t, "aliases"),
			Patterns: patterns,
		})
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return commandDefinitions, nil
}

func (lr *LuaRuntime) buildCommandPatterns(patternTable *lua.LTable) ([]models.CommandPattern, error) {
	commandPatterns := []models.CommandPattern{}
	var errs []error

	patternTable.ForEach(func(k, v lua.LValue) {
		t, ok := v.(*lua.LTable)
		if !ok {
			errs = append(errs, fmt.Errorf("command '%s' is not a table", lua.LVAsString(k)))
			return
		}

		tokens := tokenizeCommandSyntax(getString(t, "syntax"))

		commandPatterns = append(commandPatterns, models.CommandPattern{
			Tokens:         tokens,
			HelpMessage:    getString(t, "help"),
			NoMatchMessage: getString(t, "noMatch"),
		})
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return commandPatterns, nil
}

func tokenizeCommandSyntax(s string) []models.PatToken {
	var tokens []models.PatToken
	parts := strings.Fields(s)

	for _, part := range parts[:len(parts)-1] {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			slot := strings.Trim(part, "{}")
			tokens = append(tokens, models.Slot(slot))
		} else {
			tokens = append(tokens, models.Lit(part))
		}
	}

	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "{") && strings.HasSuffix(lastPart, "}") {
		slot := strings.Trim(lastPart, "{}")
		if strings.Contains(slot, "...") {
			tokens = append(tokens, models.SlotRest(strings.TrimSuffix(slot, "...")))
		} else {
			tokens = append(tokens, models.Slot(slot))
		}
	} else {
		tokens = append(tokens, models.Lit(lastPart))
	}

	return tokens
}

func luaTableToMap(t *lua.LTable) (map[string]any, error) {
	m := make(map[string]any)
	var errs []error

	t.ForEach(func(k, luaValue lua.LValue) {
		key := lua.LVAsString(k)
		if key == "" {
			errs = append(errs, fmt.Errorf("lua table to map: empty key"))
			return
		}

		value, err := luaValueToAny(luaValue)
		if err != nil {
			errs = append(errs, fmt.Errorf("lua table to map: invalid value %v", luaValue))
			return
		}
		m[key] = value
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return m, nil
}

func luaValueToAny(v lua.LValue) (any, error) {
	switch v := v.(type) {
	case *lua.LTable:
		return luaTableToMap(v)
	case lua.LNumber:
		return float64(v), nil
	case lua.LString:
		return string(v), nil
	case lua.LBool:
		return bool(v), nil
	case *lua.LNilType:
		return nil, nil
	default:
		return nil, fmt.Errorf("lua value to any: invalid type")
	}
}

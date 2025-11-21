package lua_runtime

import (
	"fmt"

	"example.com/mud/world/entities"
	lua "github.com/yuin/gopher-lua"
)

// convert a runtime entity into a snapshot for lua
func entityToTable(L *lua.LState, e *entities.Entity) (*lua.LTable, error) {
	if e == nil {
		return L.NewTable(), nil
	}

	t := L.NewTable()

	t.RawSetString("id", lua.LString(e.Id))
	t.RawSetString("name", lua.LString(e.Name))
	t.RawSetString("description", lua.LString(e.Description))

	// aliases
	aliases := L.NewTable()
	for _, a := range e.Aliases {
		aliases.Append(lua.LString(a))
	}
	t.RawSetString("aliases", aliases)

	// tags
	tags := L.NewTable()
	for _, tag := range e.Tags {
		tags.Append(lua.LString(tag))
	}
	t.RawSetString("tags", tags)

	// fields
	fields, err := mapToTable(L, e.Fields)
	if err != nil {
		return nil, fmt.Errorf("entity to table: %w", err)
	}
	t.RawSetString("fields", fields)

	// children
	children := L.NewTable()
	for group, groupChildren := range e.GetChildren() {
		childrenGroup := L.NewTable()

		for _, groupChild := range groupChildren {
			childTable, err := entityToTable(L, groupChild)
			if err != nil {
				return nil, fmt.Errorf("child entity to table: %w", err)
			}
			childrenGroup.Append(childTable)
		}

		children.RawSetString(group, childrenGroup)
	}
	t.RawSetString("children", children)

	return t, nil
}

// recursively convert a map[string]any into a lua table
func mapToTable(L *lua.LState, m map[string]any) (*lua.LTable, error) {
	t := L.NewTable()

	for k, v := range m {
		luaValue, err := valueToLuaValue(L, v)
		if err != nil {
			return nil, fmt.Errorf("map to lua table: %w", err)
		}

		t.RawSetString(k, luaValue)
	}

	return t, nil
}

func valueToLuaValue(L *lua.LState, v any) (lua.LValue, error) {
	switch x := v.(type) {
	case nil:
		return lua.LNil, nil
	case string:
		return lua.LString(x), nil
	case bool:
		return lua.LBool(x), nil
	case float64:
		return lua.LNumber(x), nil
	case float32:
		return lua.LNumber(float64(x)), nil
	case int:
		return lua.LNumber(float64(x)), nil
	case int8:
		return lua.LNumber(float64(x)), nil
	case int16:
		return lua.LNumber(float64(x)), nil
	case int32:
		return lua.LNumber(float64(x)), nil
	case int64:
		return lua.LNumber(float64(x)), nil
	case uint:
		return lua.LNumber(float64(x)), nil
	case uint8:
		return lua.LNumber(float64(x)), nil
	case uint16:
		return lua.LNumber(float64(x)), nil
	case uint32:
		return lua.LNumber(float64(x)), nil
	case uint64:
		return lua.LNumber(float64(x)), nil
	case map[string]any:
		table, err := mapToTable(L, x)
		if err != nil {
			return nil, fmt.Errorf("nested table in value to lua value: %w", err)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("value to lua value invalid type: %v", v)
	}
}

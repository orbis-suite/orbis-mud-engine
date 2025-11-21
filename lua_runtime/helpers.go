package lua_runtime

import (
	lua "github.com/yuin/gopher-lua"
)

func getString(t *lua.LTable, key string) string {
	if s, ok := t.RawGetString(key).(lua.LString); ok {
		return string(s)
	}
	return ""
}
func getStringArray(t *lua.LTable, key string) []string {
	if v, ok := t.RawGetString(key).(*lua.LTable); ok {
		var out []string
		v.ForEach(func(_, x lua.LValue) {
			if s, ok := x.(lua.LString); ok {
				out = append(out, string(s))
			}
		})
		return out
	}
	return nil
}

func getStringMap(t *lua.LTable, key string) map[string]string {
	if v, ok := t.RawGetString(key).(*lua.LTable); ok {
		out := make(map[string]string)
		v.ForEach(func(k, x lua.LValue) { out[lua.LVAsString(k)] = lua.LVAsString(x) })
		return out
	}
	return nil
}

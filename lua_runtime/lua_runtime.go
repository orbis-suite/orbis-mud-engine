package lua_runtime

import (
	lua "github.com/yuin/gopher-lua"
)

type LuaRuntime struct{ L *lua.LState }

func NewLuaRuntime() *LuaRuntime { return &LuaRuntime{L: lua.NewState()} }

func (lr *LuaRuntime) Close() { lr.L.Close() }

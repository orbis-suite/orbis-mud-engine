package lua_runtime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"example.com/mud/world/entities"
	"example.com/mud/world/scheduler"
	lua "github.com/yuin/gopher-lua"
)

func (lr *LuaRuntime) wrapLuaReaction(fn *lua.LFunction) entities.ReactionFunc {
	return func(ev *entities.Event) {
		// schedule a job to run the reaction function, in the single-threaded scheduler
		ev.Scheduler.Add(
			&scheduler.Job{
				NextRun: time.Now(),
				RunFunc: func() error {
					co, _ := lr.L.NewThread()
					injectOrbisApi(&LuaRuntime{L: co}, ev)

					if err := co.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}); err != nil {
						return err
					}

					return nil
				},
			},
		)
	}
}

func injectOrbisApi(lr *LuaRuntime, ev *entities.Event) {
	orbis := lr.L.NewTable()

	lr.L.SetField(orbis, "color", lr.L.NewClosure(getApiColor()))
	lr.L.SetField(orbis, "print", lr.L.NewClosure(getApiPrint(ev)))
	lr.L.SetField(orbis, "publish", lr.L.NewClosure(getApiPublish(ev)))
	lr.L.SetField(orbis, "publish_to_room", lr.L.NewClosure(getApiPublishToRoom(ev)))

	lr.L.SetField(orbis, "move_to_room", lr.L.NewClosure(getApiMoveToEntity(ev)))

	lr.L.SetField(orbis, "get_entity", lr.L.NewClosure(getApiGetEntity(ev)))
	lr.L.SetField(orbis, "set_entity", lr.L.NewClosure(getApiSetEntity(ev)))
	lr.L.SetField(orbis, "get_room", lr.L.NewClosure(getApiGetRoom(ev)))
	lr.L.SetField(orbis, "get_param", lr.L.NewClosure(getApiGetParam(ev)))

	lr.L.SetField(orbis, "after", lr.L.NewClosure(getApiAfter(lr, ev)))

	lr.L.SetGlobal("API", orbis)
}

func luaErr(L *lua.LState, format string, args ...interface{}) int {
	L.RaiseError(format, args...)
	return 0
}

func getApiColor() func(L *lua.LState) int {
	return func(L *lua.LState) int {
		hex := strings.TrimSpace(L.CheckString(1))
		msg := L.CheckString(2)

		if len(hex) != 6 {
			L.ArgError(1, "Orbis.Color: color must be 6 hex digits, e.g. 'ff0000'")
			return 0
		}

		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			L.ArgError(1, "Orbis.Color: invalid red component in color")
			return 0
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			L.ArgError(1, "Orbis.Color: invalid green component in color")
			return 0
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			L.ArgError(1, "Orbis.Color: invalid blue component in color")
			return 0
		}

		// 24-bit (truecolor) ANSI: 38 = foreground, 2 = RGB
		// e.g. "\x1b[38;2;255;0;0mHello\x1b[0m"
		colored := fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, msg)

		L.Push(lua.LString(colored))
		return 1
	}
}

func getApiPrint(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roleString := L.CheckString(1)
		eventRole, err := entities.ParseEventRole(roleString)
		if err != nil {
			return luaErr(L, "Orbis.Print: invalid role '%s': %v", roleString, err)
		}

		recipient, err := ev.RequireRole(eventRole)
		if err != nil {
			return luaErr(L, "Orbis.Print: get role '%s': %v", roleString, err)

		}

		messageRaw := L.CheckString(2)

		message, err := entities.FormatEventMessage(messageRaw, ev)
		if err != nil {
			return luaErr(L, "Orbis.Print: format message: %v", err)
		}

		if ev.Publisher != nil {
			ev.Publisher.PublishTo(recipient, message)
		}
		return 0
	}

}

func getApiPublish(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		messageRaw := L.CheckString(1)
		message, err := entities.FormatEventMessage(messageRaw, ev)
		if err != nil {
			return luaErr(L, "Orbis.Publish: format message: %v", err)
		}

		if ev.Publisher != nil {
			ev.Publisher.Publish(ev.Room, message, []*entities.Entity{ev.Source, ev.Instrument, ev.Target})
		}
		return 0
	}
}

func getApiPublishToRoom(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roomId := L.CheckString(1)

		room, ok := ev.EntitiesById[roomId]
		if !ok {
			return luaErr(L, "Orbis.PublishToRoom: room with id '%s' does not exist", roomId)
		}

		messageRaw := L.CheckString(2)
		message, err := entities.FormatEventMessage(messageRaw, ev)
		if err != nil {
			return luaErr(L, "Orbis.PublishToRoom: format message: %v", err)
		}

		if ev.Publisher != nil {
			ev.Publisher.Publish(room, message, []*entities.Entity{ev.Source, ev.Instrument, ev.Target})
		}
		return 0
	}
}

func getApiGetEntity(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roleString := L.CheckString(1)
		eventRole, err := entities.ParseEventRole(roleString)
		if err != nil {
			return luaErr(L, "Orbis.GetEntity: invalid role '%s': %v", roleString, err)
		}

		entity, err := ev.GetRole(eventRole)
		if err != nil {
			return luaErr(L, "Orbis.GetEntity: get role '%s': %v", roleString, err)
		} else if entity == nil {
			L.Push(lua.LNil)
			return 1
		}

		snapshot, err := entityToTable(L, entity)
		if err != nil {
			return luaErr(L, "Orbis.GetEntity: could not build entity snapshot: %v", err)
		}

		L.Push(snapshot)
		return 1
	}
}

func getApiGetRoom(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roomId := L.CheckString(1)

		room, ok := ev.EntitiesById[roomId]
		if !ok {
			return luaErr(L, "Orbis.GetRoom: room with id '%s' does not exist", roomId)
		}

		snapshot, err := entityToTable(L, room)
		if err != nil {
			return luaErr(L, "Orbis.GetRoom: could not build entity snapshot: %v", err)
		}

		L.Push(snapshot)
		return 1
	}
}

func getApiGetParam(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		paramName := L.CheckString(1)

		param, ok := ev.CommandParameters[paramName]
		if !ok {
			return luaErr(L, "Orbis.GetParam: no command parameter by the name '%s'", paramName)
		}

		L.Push(lua.LString(param))
		return 1
	}
}

func getApiAfter(lr *LuaRuntime, ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		msDelay := L.CheckInt(1)

		f := L.CheckFunction(2)
		wrappedF := lr.wrapLuaReaction(f)

		ev.Scheduler.Add(&scheduler.Job{
			NextRun: time.Now().Add(time.Duration(msDelay) * time.Millisecond),
			RunFunc: func() error {
				wrappedF(ev)
				return nil
			},
		})
		return 0
	}
}

func getApiSetEntity(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roleString := L.CheckString(1)
		eventRole, err := entities.ParseEventRole(roleString)
		if err != nil {
			return luaErr(L, "Orbis.SetEntity: invalid role '%s': %v", roleString, err)
		}

		entity, err := ev.RequireRole(eventRole)
		if err != nil {
			return luaErr(L, "Orbis.SetEntity: get role '%s': %v", roleString, err)
		}

		path := L.CheckString(2)

		luaValue := L.CheckAny(3)
		value, err := luaValueToAny(luaValue)
		if err != nil {
			return luaErr(L, "Orbis.SetEntity: invalid value: %s", err)
		}

		err = entity.SetField(path, value)
		if err != nil {
			return luaErr(L, "Orbis.SetEntity: set field '%s': %v", path, err)
		}

		return 0
	}
}

func getApiMoveToEntity(ev *entities.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		roleToMoveString := L.CheckString(1)
		roleToMove, err := entities.ParseEventRole(roleToMoveString)
		if err != nil {
			return luaErr(L, "Orbis.MoveToEntity: invalid role to move '%s': %v", roleToMove, err)
		}

		entityToMove, err := ev.RequireRole(roleToMove)
		if err != nil {
			return luaErr(L, "Orbis.MoveToEntity: get role '%s': %v", roleToMoveString, err)
		}

		destinationId := L.CheckString(2)
		destination, ok := ev.EntitiesById[destinationId]
		if !ok {
			return luaErr(L, "Orbis.MoveToEntity: invalid destination id '%s'", destinationId)
		}

		childrenGroup := L.CheckString(3)

		destination.AddChild(childrenGroup, entityToMove)

		// subscribe player to new room
		ev.Publisher.Move(destination, entityToMove)

		return 0
	}
}

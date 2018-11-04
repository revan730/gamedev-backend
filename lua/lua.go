package lua

import (
	"fmt"

	t "github.com/revan730/gamedev-backend/types"
	"github.com/yuin/gopher-lua"
)

type JumperInterpreter struct {
	userData *t.User
}

func NewInterpreter(userData *t.User) JumperInterpreter {
	return JumperInterpreter{
		userData: userData,
	}
}

func (i JumperInterpreter) DoString(luaStr string) bool {
	knowledge := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Knowledge))
		return 1
	}
	performance := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Performance))
		return 1
	}
	sober := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Sober))
		return 1
	}
	prestige := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Prestige))
		return 1
	}
	connections := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Connections))
		return 1
	}
	jump := func(L *lua.LState) int {
		page := L.ToInt(1)
		i.userData.CurrentPage = int64(page)
		return 0
	}
	flagCheck := func(L *lua.LState) int {
		flag := L.ToString(1)
		checked := i.userData.IsFlagSet(flag)
		L.Push(lua.LBool(checked))
		return 1
	}
	L := lua.NewState()
	defer L.Close()
	L.SetGlobal("knowledge", L.NewFunction(knowledge))
	L.SetGlobal("performance", L.NewFunction(performance))
	L.SetGlobal("sober", L.NewFunction(sober))
	L.SetGlobal("prestige", L.NewFunction(prestige))
	L.SetGlobal("connections", L.NewFunction(connections))
	L.SetGlobal("jump", L.NewFunction(jump))
	L.SetGlobal("flagCheck", L.NewFunction(flagCheck))
	if err := L.DoString(luaStr); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

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

// DoString interpretes provided Lua script, returning
// false in case of any errors
func (i JumperInterpreter) DoString(luaStr string) bool {
	knowledge := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Knowledge))
		return 1
	}
	addKnowledge := func(L *lua.LState) int {
		diff := L.ToInt(1)
		i.userData.Knowledge += diff
		return 0
	}
	performance := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Performance))
		return 1
	}
	addPerformance := func(L *lua.LState) int {
		diff := L.ToInt(1)
		i.userData.Performance += diff
		return 0
	}
	sober := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Sober))
		return 1
	}
	addSober := func(L *lua.LState) int {
		diff := L.ToInt(1)
		i.userData.Sober += diff
		return 0
	}
	prestige := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Prestige))
		return 1
	}
	addPrestige := func(L *lua.LState) int {
		diff := L.ToInt(1)
		i.userData.Prestige += diff
		return 0
	}
	connections := func(L *lua.LState) int {
		L.Push(lua.LNumber(i.userData.Connections))
		return 1
	}
	addConnections := func(L *lua.LState) int {
		diff := L.ToInt(1)
		i.userData.Connections += diff
		return 0
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
	setFlag := func(L *lua.LState) int {
		flag := L.ToString(1)
		i.userData.SetFlag(flag)
		return 0
	}
	L := lua.NewState()
	defer L.Close()
	L.SetGlobal("knowledge", L.NewFunction(knowledge))
	L.SetGlobal("performance", L.NewFunction(performance))
	L.SetGlobal("sober", L.NewFunction(sober))
	L.SetGlobal("prestige", L.NewFunction(prestige))
	L.SetGlobal("connections", L.NewFunction(connections))
	L.SetGlobal("addKnowledge", L.NewFunction(addKnowledge))
	L.SetGlobal("addPerformance", L.NewFunction(addPerformance))
	L.SetGlobal("addSober", L.NewFunction(addSober))
	L.SetGlobal("addPrestige", L.NewFunction(addPrestige))
	L.SetGlobal("addConnections", L.NewFunction(addConnections))
	L.SetGlobal("jump", L.NewFunction(jump))
	L.SetGlobal("flagCheck", L.NewFunction(flagCheck))
	L.SetGlobal("setFlag", L.NewFunction(setFlag))
	if err := L.DoString(luaStr); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

package project

import (
	"os"

	lua "github.com/yuin/gopher-lua"
)

const queryTypeName = "query"

type LuaQuery struct {
	reader *LogReader
}

func registerQueryType(L *lua.LState, logFileName string) {
	mt := L.NewTypeMetatable("log_reader_query")
	// todo rename to _query
	L.SetGlobal("query", mt)

	L.SetField(mt, "new", L.NewFunction(func(L *lua.LState) int {
		f, err := os.Open(logFileName)
		if err != nil {
			L.RaiseError("can't open log file %s: %s", logFileName, err)
			return 0
		}
		reader := NewLogReader(f)

		ud := L.NewUserData()
		ud.Value = &LuaQuery{reader: reader}
		L.SetMetatable(ud, L.GetTypeMetatable("log_reader_mt"))
		L.Push(ud)
		return 1
	}))

	mtInst := L.NewTypeMetatable("log_reader_mt")
	L.SetField(mtInst, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"rows": queryRows,
	}))
}

func queryRows(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lq, ok := ud.Value.(*LuaQuery)
	if !ok {
		L.ArgError(1, "expected query userdata")
		return 0
	}

	iterFn := L.NewFunction(func(L *lua.LState) int {
		if !lq.reader.Next() {
			if err := lq.reader.Err(); err != nil {
				L.RaiseError("lua query error: %s", err)
			}
			return 0
		}
		L.Push(goToLuaValue(L, lq.reader.Row()))
		return 1
	})

	L.Push(iterFn)
	return 1
}

package project

import (
	"database/sql"

	lua "github.com/yuin/gopher-lua"
)

const queryTypeName = "query"

type LuaQuery struct {
	rows *sql.Rows
}

func registerLuaQuery(db *sql.DB, L *lua.LState) {
	mt := L.NewTypeMetatable(queryTypeName)
	L.SetGlobal("query", mt)
	// static attributes
	L.SetField(mt, "new", L.NewFunction(newLuaQuery(db, L)))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"rows": luaQueryRows,
	}))
}

func newLuaQuery(db *sql.DB, L *lua.LState) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		sqlQuery := L.CheckString(1)
		rows, err := db.Query(sqlQuery)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		q := &LuaQuery{
			rows: rows,
		}
		ud := L.NewUserData()
		ud.Value = q
		L.SetMetatable(ud, L.GetTypeMetatable(queryTypeName))
		L.Push(ud)
		return 1
	}
}

func checkQuery(L *lua.LState) *LuaQuery {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*LuaQuery); ok {
		return v
	}
	L.ArgError(1, "Query expected")
	return nil
}

func luaNext(L *lua.LState) int {
	q := checkQuery(L)
	if q == nil {
		L.Push(lua.LNil)
		return 1
	}
	if !q.rows.Next() {
		L.Push(lua.LNil)
		return 1
	}
	v := make(map[string]any)
	if err := scanToMap(q.rows, v); err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(goToLuaValue(L, v))
	return 1
}

func luaQueryRows(L *lua.LState) int {
	q := checkQuery(L)
	if q == nil {
		L.Push(lua.LNil)
		return 1
	}
	// return iterator function
	iter := L.NewFunction(func(L *lua.LState) int {
		if !q.rows.Next() {
			L.Push(lua.LNil) // конец итерации
			return 1
		}
		v := make(map[string]any)
		if err := scanToMap(q.rows, v); err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(goToLuaValue(L, v))
		return 1
	})

	L.Push(iter)
	return 1
}

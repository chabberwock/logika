package jsqlite

import (
	"database/sql"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

type QueryReader interface {
	Next() bool
	Row() (map[string]string, error)
	Err() error
}

type SQLReader struct {
	*sql.Rows
}

func NewSQLiteQuery(rows *sql.Rows) *SQLReader {
	return &SQLReader{
		Rows: rows,
	}
}

func (s *SQLReader) Row() (result map[string]string, err error) {
	if err = s.Rows.Scan(&result); err != nil {
		return nil, err
	}
	return result, nil
}

const luaQueryName = "query"

func registerLuaQueryReader(L *lua.LState) {
	mt := L.NewTypeMetatable(luaQueryName)
	L.SetGlobal(luaQueryName, mt)
	L.SetField(mt, "new", L.NewFunction(newLuaQuery))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"next": func(L *lua.LState) int {
			fmt.Println("next called")
			return 0
		},
	}))

}

func newLuaQuery(L *lua.LState) int {
	q := &LuaQuery{}
	ud := L.NewUserData()
	ud.Value = q
	L.SetMetatable(ud, L.GetTypeMetatable(luaQueryName))
	L.Push(ud)
	return 1
}

type LuaQuery struct {
}

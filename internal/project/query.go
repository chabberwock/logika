package project

import (
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/query"
	lua "github.com/yuin/gopher-lua"
)

const queryTypeName = "query"

type LuaQuery struct {
	db    *clover.DB
	query *query.Query
}

func registerQueryType(L *lua.LState, db *clover.DB) {
	mt := L.NewTypeMetatable("clover_query")
	// todo rename to _query
	L.SetGlobal("query", mt)

	L.SetField(mt, "new", L.NewFunction(func(L *lua.LState) int {
		params := L.CheckTable(1)

		q := query.NewQuery(DataCollectionName).Sort(query.SortOption{Field: "line"})

		// filter
		if f := L.GetField(params, "filter"); f.Type() == lua.LTTable {
			t := f.(*lua.LTable)
			t.ForEach(func(k, v lua.LValue) {
				q = q.Where(query.Field(k.String()).Eq(luaValueToGo(v)))
			})
		}

		// filter
		if f := L.GetField(params, "sort"); f.Type() == lua.LTTable {
			t := f.(*lua.LTable)
			var sortOptions []query.SortOption
			t.ForEach(func(k, v lua.LValue) {
				dir := 1
				if v.String() == "-1" {
					dir = -1
				}
				sortOptions = append(sortOptions, query.SortOption{Field: k.String(), Direction: dir})
			})
			q = q.Sort(sortOptions...)
		}

		// limit
		if lv := L.GetField(params, "limit"); lv.Type() == lua.LTNumber {
			q = q.Limit(int(lua.LVAsNumber(lv)))
		}

		// skip
		if sv := L.GetField(params, "skip"); sv.Type() == lua.LTNumber {
			q = q.Skip(int(lua.LVAsNumber(sv)))
		}

		ud := L.NewUserData()
		ud.Value = &LuaQuery{db: db, query: q}
		L.SetMetatable(ud, L.GetTypeMetatable("clover_query_mt"))
		L.Push(ud)
		return 1
	}))

	// Метатаблица для экземпляров query
	mtInst := L.NewTypeMetatable("clover_query_mt")
	L.SetField(mtInst, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"rows":  queryRows,
		"first": first,
	}))
}

func first(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lq, ok := ud.Value.(*LuaQuery)
	if !ok {
		L.ArgError(1, "expected query userdata")
		return 0
	}

	doc, err := lq.db.FindFirst(lq.query)
	if err != nil {
		L.RaiseError("FindAll failed: %v", err)
		return 0
	}
	row := L.NewTable()
	for k, v := range doc.AsMap() {
		L.SetField(row, k, goToLuaValue(L, v))
	}
	L.Push(row)
	return 1
}

func queryRows(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lq, ok := ud.Value.(*LuaQuery)
	if !ok {
		L.ArgError(1, "expected query userdata")
		return 0
	}

	docs, err := lq.db.FindAll(lq.query)
	if err != nil {
		L.RaiseError("FindAll failed: %v", err)
		return 0
	}

	index := 0
	iterFn := L.NewFunction(func(L *lua.LState) int {
		if index >= len(docs) {
			return 0 // end of iteration
		}
		doc := docs[index]
		index++
		row := L.NewTable()
		for k, v := range doc.AsMap() {
			L.SetField(row, k, goToLuaValue(L, v))
		}
		L.Push(row)
		return 1
	})

	L.Push(iterFn)
	return 1
}

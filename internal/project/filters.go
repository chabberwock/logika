package project

import (
	lua "github.com/yuin/gopher-lua"
)

type Filter struct {
	Title       string     `json:"title"`
	FilterTable lua.LValue `json:"-"`
	Fields      any        `json:"fields"`
	Storage     string     `json:"storage"`
}

func luaValueToGo(value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LTable:
		isArray := true
		maxIndex := 0

		v.ForEach(func(key, val lua.LValue) {
			if key.Type() != lua.LTNumber {
				isArray = false
				return
			}
			i := int(lua.LVAsNumber(key))
			if i > maxIndex {
				maxIndex = i
			}
		})

		if isArray {
			arr := make([]interface{}, 0, maxIndex)
			for i := 1; i <= maxIndex; i++ {
				val := v.RawGetInt(i)
				arr = append(arr, luaValueToGo(val))
			}
			return arr
		}

		obj := make(map[string]interface{})
		v.ForEach(func(key, val lua.LValue) {
			obj[key.String()] = luaValueToGo(val)
		})
		return obj

	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	default:
		return nil
	}
}

func fromLua(L *lua.LState, v lua.LValue) Filter {
	data := v.(*lua.LTable)
	return Filter{
		Title:       data.RawGetString("title").String(),
		FilterTable: data,
		Fields:      luaValueToGo(L.GetField(data, "fields")),
	}
}

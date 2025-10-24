package project

import (
	"bytes"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"logika/internal/fs"
	"logika/internal/jsqlite"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

const (
	logFileName = "src.log"
	dbFileName  = "log.sqlite"
	luaFileName = "workspace.lua"
)

//go:embed workspace.default.lua
var defaultWorkspace []byte

type Controller struct {
	db              *sql.DB
	dir             string
	previewTemplate string
	luaOutput       bytes.Buffer
}

func NewController() *Controller {
	return new(Controller)
}

func (c *Controller) Query(filters map[string]any, offset, limit int) ([]map[string]any, error) {
	var resp []map[string]any
	L, err := c.lua(&bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer L.Close()
	sqlQuery := "select * from data"
	var qp []any
	if offset > 0 {
		sqlQuery += " offset ?"
		qp = append(qp, offset)
	}
	if limit > 0 {
		sqlQuery += " limit ?"
		qp = append(qp, limit)
	}
	rows, err := c.db.Query(sqlQuery, qp...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := make(map[string]any)
		if err := scanToMap(rows, v); err != nil {
			return nil, err
		}
		ok, err := c.applyPostFilters(L, v, filters)
		if err != nil {
			return nil, err
		}
		if ok {
			resp = append(resp, v)
		}
	}
	return resp, err
}

func (c *Controller) lua(buf *bytes.Buffer) (*lua.LState, error) {
	L := lua.NewState()
	L.SetGlobal("_store", L.NewFunction(c.luaStoreData))
	L.SetGlobal("_load", L.NewFunction(c.luaLoadData))
	registerLuaQuery(c.db, L)

	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			buf.WriteString(L.Get(i).String() + "\t")
		}
		buf.WriteString("\n")
		return 0
	}))

	if err := L.DoFile(filepath.Join(c.dir, luaFileName)); err != nil {
		L.Close()
		return nil, err
	}
	return L, nil
}

func (c *Controller) luaStoreData(L *lua.LState) int {
	key := L.ToString(1)
	value := L.Get(2)
	if err := c.StoreData(key, luaValueToGo(value)); err != nil {

		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func (c *Controller) luaLoadData(L *lua.LState) int {
	key := L.ToString(1)
	v, err := c.LoadData(key)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(goToLuaValue(L, v))
	return 1
}

func (c *Controller) LoadData(key string) (v any, err error) {
	row := c.db.QueryRow("select `value` from storage where `key`=?", key)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var jsonData []byte
	err = row.Scan(&jsonData)
	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *Controller) StoreData(key string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = c.db.Exec("replace into storage (`key`, `value`) values (?, ?)", key, string(valueJSON))
	return err
}

func (c *Controller) createStorageIfNotExists() error {
	query := "create table if not exists storage (`key` text primary key, `value` text)"
	_, err := c.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) applyPostFilters(L *lua.LState, data map[string]any, filterOptions map[string]any) (bool, error) {
	f, err := c.filters(L)
	if err != nil {
		return false, err
	}
	for filterName, filterValue := range filterOptions {
		if filterValue == nil {
			continue
		}
		filter, ok := f[filterName]
		if !ok {
			continue
		}
		if err := L.CallByParam(lua.P{
			Fn:      filter.FilterFunc,
			NRet:    1,
			Protect: true,
		}, goToLuaValue(L, filterValue), goToLuaValue(L, data)); err != nil {
			return false, err
		}
		filterResult, ok := L.Get(-1).(lua.LBool)
		L.Pop(1) // Очистка стека после получения результата
		if !ok || filterResult == lua.LFalse {
			return false, nil
		}
	}
	return true, nil
}

func (c *Controller) Close() error {
	return c.db.Close()
}

func goToLuaValue(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, v2 := range val {
			tbl.RawSetString(k, goToLuaValue(L, v2))
		}
		return tbl
	case []interface{}:
		tbl := L.NewTable()
		for i, v2 := range val {
			tbl.RawSetInt(i+1, goToLuaValue(L, v2)) // Lua — 1-индексация
		}
		return tbl
	case string:
		return lua.LString(val)
	case float64:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case nil:
		return lua.LNil
	default:
		return lua.LString("<unsupported:" + fmt.Sprintf("%T", v) + ">")
	}
}

func (c *Controller) Import(src, destDir string) error {
	if err := os.MkdirAll(destDir, 0777); err != nil {
		return err
	}
	if err := c.checkDirEmpty(destDir); err != nil {
		return err
	}
	logfilePath := filepath.Join(destDir, logFileName)
	if err := fs.CopyFile(src, logfilePath); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, luaFileName), defaultWorkspace, 0755); err != nil {
		return err
	}
	destDBPath := filepath.Join(destDir, dbFileName)
	conv := jsqlite.NewConverter()
	err := conv.Run(logfilePath, destDBPath)
	if err != nil {
		return err
	}
	return c.Open(destDir)
}

func (c *Controller) checkDirEmpty(projectDir string) error {
	f, err := os.Open(projectDir)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return nil
	}
	return errors.New("directory is not empty")
}

func (c *Controller) Open(projectDir string) error {
	c.dir = projectDir
	db, err := sql.Open("sqlite3", filepath.Join(c.dir, dbFileName))
	if err != nil {
		return err
	}
	c.db = db
	if err := c.createStorageIfNotExists(); err != nil {
		return err
	}
	L, err := c.lua(&c.luaOutput)
	if err != nil {
		return err
	}
	defer L.Close()
	if err := c.luaInit(L); err != nil {
		fmt.Printf("init error: %v\n", err)
		return err
	}
	if err := c.luaGetPreviewTemplate(L); err != nil {
		fmt.Printf("preview error: %v\n", err)
		return fmt.Errorf("getting preview template: %w", err)
	}
	return nil
}

func (c *Controller) Dir() string {
	return c.dir
}

func (c *Controller) FiltersJSON() (string, error) {
	L, err := c.lua(&bytes.Buffer{})
	if err != nil {
		return "", err
	}
	defer L.Close()
	f, err := c.filters(L)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Controller) filters(L *lua.LState) (map[string]Filter, error) {
	resp := make(map[string]Filter)
	s, err := c.luaSettingsTable(L)
	if err != nil {
		return resp, err
	}
	filterList := s.RawGet(lua.LString("filters")).(*lua.LTable)
	filterList.ForEach(func(k, v lua.LValue) {
		resp[k.String()] = fromLua(L, v)
	})
	return resp, nil
}

func (c *Controller) luaInit(L *lua.LState) error {
	s, err := c.luaSettingsTable(L)
	if err != nil {
		return err
	}
	if err := L.CallByParam(lua.P{
		Fn:      s.RawGet(lua.LString("init")),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) luaSettingsTable(L *lua.LState) (*lua.LTable, error) {
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("_settings"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil, fmt.Errorf("getting settings: %w", err)
	}
	val := L.Get(-1)
	L.Pop(1)
	return val.(*lua.LTable), nil
}

func (c *Controller) PreviewTemplate() string {
	return c.previewTemplate
}

func (c *Controller) luaGetPreviewTemplate(L *lua.LState) error {
	s, err := c.luaSettingsTable(L)
	if err != nil {
		return err
	}
	c.previewTemplate = s.RawGet(lua.LString("previewTemplate")).String()
	return nil
}

func (c *Controller) Settings() (string, error) {
	b, err := os.ReadFile(filepath.Join(c.dir, luaFileName))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Controller) SaveSettings(settings string) error {
	return os.WriteFile(filepath.Join(c.dir, luaFileName), []byte(settings), 0755)
}

func (c *Controller) LuaOutput() string {
	return c.luaOutput.String()
}

func (c *Controller) RunScript(script string) (string, error) {
	var buf bytes.Buffer
	L, err := c.lua(&buf)
	if err != nil {
		return "", err
	}
	defer L.Close()
	if err := L.DoString(script); err != nil {
		return "", err
	}
	return buf.String(), nil
}

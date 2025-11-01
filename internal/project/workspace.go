package project

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"logika/internal/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	lua "github.com/yuin/gopher-lua"
)

const (
	logFileName = "src.log"
	dbFileName  = "log.sqlite"
	luaFileName = "workspace.lua"
)

//go:embed workspace.default.lua
var defaultWorkspaceConfig []byte

type Workspace struct {
	db              *clover.DB
	dir             string
	previewTemplate string
	luaOutput       bytes.Buffer
	bookmarksMu     sync.Mutex
}

func NewController() *Workspace {
	return new(Workspace)
}

func (c *Workspace) ID() string {
	hasher := md5.New()
	hasher.Write([]byte(c.dir))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (c *Workspace) Query(filters map[string]any, offset, limit int) ([]map[string]any, error) {
	var resp []map[string]any
	L, err := c.lua(&bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer L.Close()
	appFilters, err := c.filters(L)
	if err != nil {
		return nil, err
	}
	var matcherErr error
	matcher := func(doc *document.Document) bool {
		res, err := c.applyPostFilters(L, appFilters, doc.AsMap(), filters)
		if err != nil {
			matcherErr = err
			return false
		}
		return res
	}
	q := query.NewQuery(DataCollectionName).MatchFunc(matcher).Sort(query.SortOption{Field: "line"})
	if offset > 0 {
		q = q.Skip(offset)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	err = c.db.ForEach(q, func(doc *document.Document) bool {
		resp = append(resp, doc.AsMap())
		return true
	})
	if err != nil {
		return nil, err
	}
	if matcherErr != nil {
		return nil, matcherErr
	}
	return resp, nil
}

func (c *Workspace) lua(buf *bytes.Buffer) (*lua.LState, error) {
	L := lua.NewState()
	L.SetGlobal("_store", L.NewFunction(c.luaStoreData))
	L.SetGlobal("_load", L.NewFunction(c.luaLoadData))
	L.SetGlobal("_ranges", L.NewFunction(c.luaRanges))
	L.SetGlobal("_bookmark", L.NewFunction(c.luaBookmark))
	registerQueryType(L, c.db)

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

func (c *Workspace) luaStoreData(L *lua.LState) int {
	key := L.ToString(1)
	value := L.Get(2)
	if err := c.StoreData(key, luaValueToGo(value)); err != nil {
		L.RaiseError(err.Error())
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func (c *Workspace) luaLoadData(L *lua.LState) int {
	key := L.ToString(1)
	v, err := c.LoadData(key)
	if err != nil {
		L.RaiseError(err.Error())
		L.Push(lua.LNil)
		return 1
	}
	L.Push(goToLuaValue(L, v))
	return 1
}

func (c *Workspace) luaRanges(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	titleField := L.CheckString(3)

	firstLine, err := c.db.FindFirst(query.
		NewQuery(DataCollectionName).
		Sort(query.SortOption{Field: "line", Direction: 1}).
		Limit(1))
	if err != nil {
		L.RaiseError(err.Error())
		return 1
	}

	prev := firstLine
	tbl := L.NewTable()
	idx := 1
	q := query.NewQuery(DataCollectionName).Where(query.Field(key).Eq(value))
	c.db.ForEach(q, func(doc *document.Document) bool {
		if doc.ObjectId() == prev.ObjectId() {
			return true
		}
		m, err := doRange(prev, doc, titleField, false)
		if err != nil {
			L.RaiseError(err.Error())
			return false
		}
		tbl.Insert(idx, goToLuaValue(L, m))
		prev = doc
		idx++
		return true
	})
	lastLine, err := c.db.FindFirst(query.
		NewQuery(DataCollectionName).
		Sort(query.SortOption{Field: "line", Direction: -1}).
		Limit(1))
	if err != nil {
		L.RaiseError(err.Error())
		return 1
	}
	m, err := doRange(prev, lastLine, titleField, true)
	if err != nil {
		L.RaiseError(err.Error())
		return 1
	}
	tbl.Insert(idx, goToLuaValue(L, m))
	L.Push(tbl)
	return 1
}

func (c *Workspace) luaBookmark(L *lua.LState) int {
	line := L.CheckString(1)
	label := L.CheckString(2)
	if err := c.AddBookmark(line, label); err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func doRange(left, right *document.Document, titleField string, isLast bool) (map[string]string, error) {
	m := map[string]string{
		"title":     left.Get(titleField).(string),
		"startLine": fmt.Sprintf("%d", left.Get("line").(int64)),
	}
	if isLast {
		m["endLine"] = fmt.Sprintf("%d", right.Get("line").(int64))
	} else {
		m["endLine"] = fmt.Sprintf("%d", right.Get("line").(int64)-1)
	}
	return m, nil
}

func (c *Workspace) LoadData(key string) (v any, err error) {
	q := query.NewQuery(SettingsCollectionName).Where(query.Field("key").Eq(key))
	doc, err := c.db.FindFirst(q)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	return doc.Get("value"), nil
}

func (c *Workspace) StoreData(key string, value any) error {
	q := query.NewQuery(SettingsCollectionName).Where(query.Field("key").Eq(key))
	doc, err := c.db.FindFirst(q)
	if err != nil {
		return err
	}
	if doc == nil {
		doc = document.NewDocument()
		doc.Set("key", key)
	}
	doc.Set("value", value)
	return c.db.Save(SettingsCollectionName, doc)
}

func (c *Workspace) createStorageIfNotExists() error {
	return c.db.CreateCollection("storage")
}

func (c *Workspace) applyPostFilters(L *lua.LState, f map[string]Filter, data map[string]any, filterOptions map[string]any) (bool, error) {
	for filterName, filterValue := range filterOptions {
		if filterValue == nil {
			continue
		}
		filter, ok := f[filterName]
		if !ok {
			continue
		}
		tbl := filter.FilterTable.(*lua.LTable)
		if err := L.CallByParam(lua.P{
			Fn:      tbl.RawGetString("filterFunc"),
			NRet:    1,
			Protect: true,
		}, tbl, goToLuaValue(L, filterValue), goToLuaValue(L, data)); err != nil {
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

func (c *Workspace) Close() error {
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
	case map[string]string:
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

func (c *Workspace) Import(src, destDir string) error {
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
	if err := os.WriteFile(filepath.Join(destDir, luaFileName), defaultWorkspaceConfig, 0755); err != nil {
		return err
	}
	if err := importLogFile(logfilePath, destDir); err != nil {
		return err
	}
	return nil
}

func (c *Workspace) checkDirEmpty(projectDir string) error {
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

func (c *Workspace) Open(projectDir string) error {
	c.dir = projectDir
	if c.db != nil {
		c.Close()
	}
	db, err := clover.Open(c.dir)
	if err != nil {
		return err
	}
	c.db = db
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
		return fmt.Errorf("getting preview template: %w", err)
	}
	return nil
}

func (c *Workspace) Dir() string {
	return c.dir
}

func (c *Workspace) FiltersJSON() (string, error) {
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

func (c *Workspace) filters(L *lua.LState) (map[string]Filter, error) {
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

func (c *Workspace) luaInit(L *lua.LState) error {
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

func (c *Workspace) luaSettingsTable(L *lua.LState) (*lua.LTable, error) {
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

func (c *Workspace) PreviewTemplate() string {
	return c.previewTemplate
}

func (c *Workspace) luaGetPreviewTemplate(L *lua.LState) error {
	s, err := c.luaSettingsTable(L)
	if err != nil {
		return err
	}
	c.previewTemplate = s.RawGet(lua.LString("previewTemplate")).String()
	return nil
}

func (c *Workspace) Settings() (string, error) {
	b, err := os.ReadFile(filepath.Join(c.dir, luaFileName))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Workspace) SaveSettings(settings string) error {
	return os.WriteFile(filepath.Join(c.dir, luaFileName), []byte(settings), 0755)
}

func (c *Workspace) LuaOutput() string {
	return c.luaOutput.String()
}

func (c *Workspace) RunScript(script string) (string, error) {
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

func (c *Workspace) AddBookmark(line string, label string) error {
	c.bookmarksMu.Lock()
	defer c.bookmarksMu.Unlock()
	v, err := c.LoadData("bookmarks")
	if err != nil {
		return err
	}
	if v == nil {
		v = make(map[string]any)
	}
	b := v.(map[string]any)
	b[line] = label
	return c.StoreData("bookmarks", b)
}

func (c *Workspace) DeleteBookmark(line string) error {
	c.bookmarksMu.Lock()
	defer c.bookmarksMu.Unlock()
	v, err := c.LoadData("bookmarks")
	if err != nil {
		return err
	}
	if v == nil {
		v = make(map[string]any)
	}
	b := v.(map[string]any)
	delete(b, line)
	return c.StoreData("bookmarks", b)
}

func (c *Workspace) Bookmarks() (any, error) {
	c.bookmarksMu.Lock()
	defer c.bookmarksMu.Unlock()
	v, err := c.LoadData("bookmarks")
	if err != nil {
		return nil, err
	}
	if v == nil {
		return map[string]string{}, nil
	}
	return v, nil
}

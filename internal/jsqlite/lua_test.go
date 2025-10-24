package jsqlite

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	registerLuaQueryReader(L)
	require.NoError(t, L.DoFile(filepath.Join("testdata", "create-query.lua")))
}

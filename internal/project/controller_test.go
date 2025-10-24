package project

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestController_Import(t *testing.T) {
	c := NewController()
	projectDir, err := os.MkdirTemp("", "TestController_Import")
	require.NoError(t, err)
	require.NoError(t, c.Import("./testdata/simple.log", projectDir))
	t.Log(projectDir)
}

func TestController_LoadData(t *testing.T) {
	c := NewController()
	projectDir, err := os.MkdirTemp("", "TestController_Import")
	require.NoError(t, err)
	require.NoError(t, c.Import("./testdata/simple.log", projectDir))
	data, err := c.LoadData("test")
	require.NoError(t, err)
	t.Logf("data: %+v", data)
}

func TestController_StoreData(t *testing.T) {
	c := NewController()
	projectDir, err := os.MkdirTemp("", "TestController_Import")
	require.NoError(t, err)
	require.NoError(t, c.Import("./testdata/simple.log", projectDir))
	testData := map[string]string{
		"foo": "bar",
	}
	err = c.StoreData("test", testData)
	require.NoError(t, err)
	data, err := c.LoadData("test")
	require.NoError(t, err)
	t.Logf("data: %+v", data)
}

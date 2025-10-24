package jsqlite

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConverter_Run(t *testing.T) {
	const testDB = "./out.db"
	c := NewConverter()
	require.NoError(t, os.Remove(testDB))
	err := c.Run("testdata/simple.log", testDB)
	require.NoError(t, err)
}

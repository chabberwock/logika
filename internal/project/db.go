package project

import (
	"github.com/ostafen/clover/v2"
)

const (
	DataCollectionName     = "data"
	SettingsCollectionName = "settings"
)

type Row struct {
	ID   int            `json:"id"`
	Data map[string]any `json:"data"`
}

func initSettings(destDir string) error {
	db, err := clover.Open(destDir)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.CreateCollection(SettingsCollectionName)
}

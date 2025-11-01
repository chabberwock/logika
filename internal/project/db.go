package project

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
)

const (
	DataCollectionName     = "data"
	SettingsCollectionName = "settings"
)

type Row struct {
	ID   int            `json:"id"`
	Data map[string]any `json:"data"`
}

func importLogFile(src, destDir string) error {
	db, err := clover.Open(destDir)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.CreateCollection(DataCollectionName); err != nil {
		return err
	}
	if err := db.CreateCollection(SettingsCollectionName); err != nil {
		return err
	}
	lineNumber := 0
	batch, err := NewBatchReader(src, 10000)
	if err != nil {
		return err
	}
	m := make(map[string]any)
	for batch.Next() {
		docs := make([]*document.Document, 0)
		for _, line := range batch.Lines() {
			lineNumber++
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				return fmt.Errorf("unmarshalling line %d: %w", lineNumber, err)
			}
			doc := document.NewDocument()
			doc.Set("line", lineNumber)
			doc.Set("data", m)
			if doc == nil {
				return fmt.Errorf("failed to import line: %s", line)
			}
			docs = append(docs, doc)
		}
		if err := db.Insert(DataCollectionName, docs...); err != nil {
			return fmt.Errorf("inserting line %d: %w", lineNumber, err)
		}
	}
	return batch.Err()
}

// deprecated
func scanToMap(rows *sql.Rows, dest map[string]any) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}
	err = rows.Scan(valPtrs...)
	if err != nil {
		return err
	}
	for i, c := range cols {
		dest[c] = vals[i]
	}
	return nil
}

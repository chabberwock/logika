package jsqlite

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	primaryKeyField  = "__id"
	rawTextField     = "__text"
	primaryTableName = "data"
)

type Field struct {
	Name      string
	FieldType string
}

type FieldList struct {
	Fields []Field
}

type Converter struct {
}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Run(src string, dest string) error {
	db, err := sql.Open("sqlite3", dest)
	if err != nil {
		return fmt.Errorf("opening sqlite db: %w", err)
	}
	defer db.Close()
	if err := c.createTableSchema(src, primaryTableName, db); err != nil {
		return err
	}
	if err := c.importFile(src, primaryTableName, db); err != nil {
		return err
	}
	return nil
}

func (c *Converter) createTableSQL(tableName string, fl FieldList) string {
	var res []string

	res = append(res,
		fmt.Sprintf(`"%s" integer not null PRIMARY KEY`, primaryKeyField),
		fmt.Sprintf(`"%s" string`, rawTextField),
	)
	for _, f := range fl.Fields {
		res = append(res, fmt.Sprintf("`%s` %s", f.Name, f.FieldType))
	}
	return "create table `" + tableName + "` (" + strings.Join(res, ",") + ")"
}

func (c *Converter) insertSQL(tableName string, lineNumber int, data map[string]any, rawText string) (string, []any) {
	query := "INSERT INTO `" + tableName + "` "
	fieldNames := []string{
		primaryKeyField,
		rawTextField,
	}
	placeholders := []string{
		"?",
		"?",
	}
	fieldValues := []any{
		lineNumber,
		rawText,
	}

	for k := range data {
		fieldNames = append(fieldNames, "`"+k+"`")
		placeholders = append(placeholders, "?")
		fieldValues = append(fieldValues, data[k])
	}
	query = query + fmt.Sprintf("(%s) values ", strings.Join(fieldNames, ","))
	query = query + "(" + strings.Join(placeholders, ",") + ")"
	return query, fieldValues
}

func (c *Converter) fieldList(filename string) (FieldList, error) {
	f, err := os.Open(filename)
	if err != nil {
		return FieldList{}, fmt.Errorf("opening source file: %w", err)
	}
	scanner := bufio.NewScanner(f)
	m := make(map[string]any)
	lineNumber := 0
	foundFields := make(map[string]Field)
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return FieldList{}, fmt.Errorf("unmarshalling line %d: %w", lineNumber, err)
		}
		for k := range m {
			if _, ok := foundFields[k]; ok {
				continue
			}
			foundFields[k] = Field{Name: k, FieldType: "string"}
		}
	}
	var result FieldList
	for k := range foundFields {
		result.Fields = append(result.Fields, foundFields[k])
	}
	return result, scanner.Err()
}

func (c *Converter) createTableSchema(src, tableName string, db *sql.DB) error {
	fl, err := c.fieldList(src)
	if err != nil {
		return err
	}
	_, err = db.Exec(c.createTableSQL(tableName, fl))
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	return nil
}

func (c *Converter) importFile(src string, tableName string, db *sql.DB) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	scanner := bufio.NewScanner(f)
	m := make(map[string]any)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return fmt.Errorf("unmarshalling line %d: %w", lineNumber, err)
		}
		query, args := c.insertSQL(tableName, lineNumber, m, line)
		if _, err := db.Exec(query, args...); err != nil {
			return fmt.Errorf("inserting line %d: %w", lineNumber, err)
		}
	}
	return scanner.Err()

}

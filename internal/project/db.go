package project

import "database/sql"

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

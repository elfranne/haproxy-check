package main

import (
	"bytes"
	"database/sql"
)

func createDB(data *statsData) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	cols, err := data.ColumnNames()
	if err != nil {
		return nil, err
	}
	if err := metricsDDLTmpl.Execute(&buf, cols); err != nil {
		return nil, err
	}

	if _, err := db.Exec(buf.String()); err != nil {
		return nil, err
	}

	buf.Reset()

	if err := insertMetricTmpl.Execute(&buf, cols); err != nil {
		return nil, err
	}

	rows, err := data.Rows()
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if _, err := db.Exec(buf.String(), row...); err != nil {
			return nil, err
		}
	}

	return db, nil
}

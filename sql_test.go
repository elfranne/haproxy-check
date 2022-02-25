package main

import (
	"bytes"
	"database/sql"
	"testing"
)

func TestMetricsDDL(t *testing.T) {
	columnNames := []string{"one", "two", "three", "four"}
	var buf bytes.Buffer
	if err := metricsDDLTmpl.Execute(&buf, columnNames); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(buf.String()); err != nil {
		t.Fatal(err)
	}
}

func TestInsertMetrics(t *testing.T) {
	columnNames := []string{"one", "two", "three", "four"}
	var buf bytes.Buffer
	if err := metricsDDLTmpl.Execute(&buf, columnNames); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(buf.String()); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := insertMetricTmpl.Execute(&buf, columnNames); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(buf.String(), "1", "2", "3", "4"); err != nil {
		t.Fatal(err)
	}
}

package main

import "testing"

func TestCreateDB(t *testing.T) {
	data := &statsData{
		data: testDataCSV,
	}
	db, err := createDB(data)
	if err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("SELECT count(*) FROM metrics;")
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got, want := count, 9; got != want {
		t.Errorf("bad count: got %d, want %d", got, want)
	}
	row = db.QueryRow("SELECT count(*) FROM metrics WHERE smax = 10;")
	if err := row.Scan(&count); err != nil {
		t.Fatal(err)
	}
	if got, want := count, 1; got != want {
		t.Errorf("bad count: got %d, want %d", got, want)
	}
}

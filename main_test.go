package main

import (
	_ "embed"
	"testing"
)

//go:embed testdata.csv
var testDataCSV []byte

func TestMain(t *testing.T) {
}

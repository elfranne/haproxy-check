package main

import (
	_ "embed"
)

//go:embed testdata.csv
var testDataCSV []byte

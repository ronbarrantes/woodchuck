package main

import "fmt"

// #### CSV FILE ####
type CsvFile struct {
	filename string
	path     string
}

func Csv() *CsvFile {
	return &CsvFile{}
}

func (f *CsvFile) DeleteCSV() {
	fmt.Println("Deleting...")
}

func (f *CsvFile) InitCSV() {
	// Initalize the CSV
	fmt.Println("Initializing...")
}

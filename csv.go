package main

import "fmt"

// #### CSV FILE ####
type CsvFile struct {
	filename string
	path     string
}

type Writer struct {
	Comma   rune // Field delimiter (set to ',' by NewWriter)
	UseCRLF bool // True to use \r\n as the line terminator
	// contains filtered or unexported fields
}

func Csv(p, f string) *CsvFile {
	return &CsvFile{
		path:     p,
		filename: f,
	}
}

func (f *CsvFile) DeleteCSV() {
	fmt.Println("Deleting...")
}

type CSVLogEntry struct {
	Timestamp string `csv:"timestamp"`
	LogID     int    `csv:"log_id"`
	LogLevel  string `csv:"level"` // "error", "warning", "log"
	UserID    string `csv:"user_id"`
	Message   string `csv:"message"`
}

func (f *CsvFile) InitCSV() {
	// Initalize the CSV
	fmt.Println("Initializing...")
	// check if the file exist and if it doesn't
	// create a new file with file
}

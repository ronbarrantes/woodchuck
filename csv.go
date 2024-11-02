package main

import (
	"fmt"
	"time"

	"github.com/ronbarrantes/woodchuck/utils"
)

// #### CSV FILE ####
type CsvFile struct {
	filename string
	path     string
	fullpath string
}

type Writer struct {
	Comma   rune // Field delimiter (set to ',' by NewWriter)
	UseCRLF bool // True to use \r\n as the line terminator
}

func Csv(p, f string) *CsvFile {
	return &CsvFile{
		path:     p,
		filename: f,
	}
}

type CSVLogEntry struct {
	Timestamp string `csv:"timestamp"`
	LogID     int    `csv:"log_id"`
	LogLevel  string `csv:"level"` // "error", "warning", "log"
	UserID    string `csv:"user_id"`
	Message   string `csv:"message"`
}

func (f *CsvFile) InitCSV() error {
	// Initalize the CSV
	fmt.Println("Initializing...")
	// Get the name
	currDate := time.Now().Format("2006-01-02")
	currentFileName := currDate + "-" + f.filename

	fullpath, err := utils.EnsureDirectoryAndFile(f.path, currentFileName)
	if err != nil {
		return err
	}

	f.fullpath = fullpath
	return nil
}

func (f *CsvFile) DeleteCSV() error {
	fmt.Println("Deleting...")
	err := utils.RemoveFile(f.fullpath)
	if err != nil {
		return err
	}

	f.fullpath = ""

	return nil
}

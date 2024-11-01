package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
	// Get the name
	currDate := time.Now().Format("2006-01-02")
	currentFileName := currDate + "-" + f.filename

	fullPath := FullPath(f.path, currentFileName)

	fmt.Println("--->>>", fullPath)

	// create a directory if it doesnt exist

	// check if the file exist and if it doesn't
	// create a new file with file

	if err := EnsureDirectoryAndFile(f.path, currentFileName); err != nil {
		panic("File cannot be created")
	}

}
func FullPath(path, filename string) string {
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	return path + filename
}

// EnsureDirectoryAndFile checks if a directory and file exist and creates them if they don't
func EnsureDirectoryAndFile(path, filename string) error {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Directory does not exist, creating it...")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Combine directory and filename to get the full file path
	filePath := filepath.Join(path, filename)

	// Check if the file exists in the directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("File does not exist, creating it...")
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
	}

	return nil
}

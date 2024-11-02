package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

	err := EnsureDirectoryAndFile(f.path, currentFileName)
	if err != nil {
		return err
	}

	f.fullpath = filepath.Join(f.path, f.filename)
	return nil
}

func (f *CsvFile) DeleteCSV() error {
	fmt.Println("Deleting...")
	err := RemoveFile(f.fullpath)
	if err != nil {
		return err
	}

	f.fullpath = ""
	return nil
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

		// Write headers to the file
		writer := csv.NewWriter(file)
		headers := []string{"timestamp", "log_id", "level", "user_id", "message"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
		writer.Flush()
	}

	return nil
}

// Remove File
func RemoveFile(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Println("Nothing to delete")
		return fmt.Errorf("File %s doesn't exist", err)
	}

	if err := os.Remove(filepath); err != nil {
		return err
	}

	return nil
}

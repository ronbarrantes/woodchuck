package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	LogLevel  string `csv:"level"` // "error", "warning", "log"
	LogID     int    `csv:"log_id"`
	UserID    string `csv:"user_id"`
	Message   string `csv:"message"`
}

func (f *CsvFile) InitCSV() (*csv.Writer, error) {
	// Initalize the CSV
	fmt.Println("Initializing...")
	// Get the name
	currDate := time.Now().Format("2006-01-02")
	currentFileName := currDate + "-" + f.filename

	csvWriter, err := EnsureDirectoryAndFile(f.path, currentFileName)
	if err != nil {
		return nil, err
	}

	f.fullpath = filepath.Join(f.path, currentFileName)
	return csvWriter, nil
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

// EnsureDirectoryAndFile checks if a directory and file exist and creates them if they don't
func EnsureDirectoryAndFile(path, filename string) (*csv.Writer, error) {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Directory does not exist, creating it...")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Combine directory and filename to get the full file path
	filePath := filepath.Join(path, filename)

	// Check if the file exists in the directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("File does not exist, creating it...")
		file, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		// Write headers to the file
		writer := csv.NewWriter(file)
		headers := []string{"timestamp", "level", "log_id", "user_id", "message"}
		if err := writer.Write(headers); err != nil {
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
		writer.Flush()
		return writer, nil
	}

	// Open the existing file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	writer := csv.NewWriter(file)
	return writer, nil
}

func (f *CsvFile) WriteToCSV(entry *CSVLogEntry) error {
	// Open the file in append mode, create if it doesn't exist
	file, err := os.OpenFile(f.fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)

	// Convert the entry to a slice of strings
	record := []string{
		entry.Timestamp,
		strings.ToUpper(fmt.Sprintf("%s", entry.LogLevel)),
		fmt.Sprintf("%d", entry.LogID),
		entry.UserID,
		entry.Message,
	}

	// Write the entry to the CSV file
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write entry to CSV: %w", err)
	}

	// Flush the writer
	writer.Flush()

	// Check for any error during the flush
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

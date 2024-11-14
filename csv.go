package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func NewCsvFile(p, f string) *CsvFile {
	currDate := time.Now().Format("2006-01-02")
	currentFileName := currDate + "-" + f
	fullpath := filepath.Join(p, currentFileName)

	return &CsvFile{
		path:     p,
		filename: f,
		fullpath: fullpath,
	}
}

type CSVLogEntry struct {
	Timestamp string `csv:"timestamp"`
	LogLevel  string `csv:"level"` // "error", "warning", "log"
	LogID     int    `csv:"log_id"`
	UserID    string `csv:"user_id"`
	Message   string `csv:"message"`
}

func CreateFullPath(path, name string) string {
	currDate := time.Now().Format("2006-01-02")
	currentFileName := currDate + "-" + name

	return filepath.Join(path, currentFileName)
}

func (f *CsvFile) InitCSV() (*csv.Writer, error) {
	fmt.Println("Initializing...")
	csvWriter, err := f.EnsureDirectoryAndFile()
	if err != nil {
		return nil, err
	}

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
func (f *CsvFile) EnsureDirectoryAndFile() (*csv.Writer, error) {

	// Check if directory exists
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		fmt.Println("Directory does not exist, creating it...")
		if err := os.MkdirAll(f.path, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Check if the file exists in the directory
	if _, err := os.Stat(f.fullpath); os.IsNotExist(err) {
		fmt.Println("File does not exist, creating it...")
		file, err := os.Create(f.fullpath)
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
	file, err := os.OpenFile(f.fullpath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
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
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Create a buffered writer
	bufWriter := bufio.NewWriter(file)
	writer := csv.NewWriter(bufWriter)

	// Convert the entry to a slice of strings
	record := []string{
		entry.Timestamp,
		strings.ToUpper(entry.LogLevel),
		strconv.Itoa(entry.LogID),
		entry.UserID,
		entry.Message,
	}

	// Write the entry to the CSV file
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write entry to CSV: %w", err)
	}

	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	// Flush the buffered writer
	if err := bufWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

func (f *CsvFile) ReadLastItemCSV() (*CSVLogEntry, error) {
	file, err := os.Open(f.fullpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("no records found: %w", err)
	}
	lastRecord := records[len(records)-1]

	logID, err := strconv.Atoi(lastRecord[2])
	if err != nil {
		return nil, fmt.Errorf("invalid log ID: %w", err)
	}

	return &CSVLogEntry{
		Timestamp: lastRecord[0],
		LogLevel:  lastRecord[1],
		LogID:     logID,
		UserID:    lastRecord[3],
		Message:   lastRecord[4],
	}, nil
}

func (f *CsvFile) ReadLastLogID() (int, error) {
	lastEntry, err := f.ReadLastItemCSV()

	if err != nil {
		return 0, fmt.Errorf("Cannot get last entry: %w", err)
	}

	return lastEntry.LogID, nil
}

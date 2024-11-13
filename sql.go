package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBFile struct {
	file     string
	path     string
	fullpath string
	db       *gorm.DB
}

func NewDBFile(p, f string) *DBFile {
	fullpath := filepath.Join(p, f)
	return &DBFile{
		path:     p,
		file:     f,
		fullpath: fullpath,
	}
}

type DBLogModel struct {
	gorm.Model
	Timestamp time.Time
	LogLevel  string
	LogID     int
	UserID    string
	Message   string
}

func (f *DBFile) InitDB() error {
	// Ensure the database directory exists
	if err := EnsureDir(f.path); err != nil {
		panic("Fail to create path")
	}

	fmt.Printf("Initializing database at path: %s\n", f.fullpath)

	// Open SQLite connection
	db, err := gorm.Open(sqlite.Open(f.fullpath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&DBLogModel{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	f.db = db
	return nil
}

func (f *DBFile) WriteLog(log DBLogModel) error {
	if f.db == nil {
		return fmt.Errorf("database not initialized")
	}

	result := f.db.Create(&log)
	return result.Error
}

func EnsureDir(path string) error {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Directory does not exist, creating it...")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

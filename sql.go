package main

import (
	"fmt"
	"path/filepath"

	"github.com/ronbarrantes/woodchuck/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBFile struct {
	file     string
	path     string
	fullpath string
	db       *gorm.DB
}

// Creates a new db file
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
	LogLevel string
	UserID   string
	Message  string
}

func (f *DBFile) InitDB() error {
	// Ensure the database directory exists
	if err := utils.EnsureDir(f.path); err != nil {
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

// Write to the logs
func (f *DBFile) WriteLog(log DBLogModel) error {
	if f.db == nil {
		return fmt.Errorf("database not initialized")
	}

	result := f.db.Create(&log)
	return result.Error
}

// Read the logs
func (f *DBFile) ReadLogs() ([]DBLogModel, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var logs []DBLogModel
	result := f.db.Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}

	return logs, nil
}

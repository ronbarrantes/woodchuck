package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ronbarrantes/woodchuck/utils"
)

type Config struct {
	Port        string
	LogDir      string
	LogFilename string
	DBDir       string
	DBFilename  string
}

func main() {
	fmt.Println("WORKING UP HERE")

	config := loadConfig()
	db := NewDBFile(config.DBDir, config.DBFilename)

	if err := db.InitDB(); err != nil {
		panic(err)
	}

	//	Example of writing to the database
	logEntry := DBLogModel{
		Timestamp: time.Now(),
		LogLevel:  "info",
		LogID:     1,
		UserID:    "user123",
		Message:   "This is a log message",
	}

	if err := db.WriteLog(logEntry); err != nil {
		panic(err)
	}

	server := Server(config.Port)
	server.Run()
}

func loadConfig() Config {
	utils.LoadEnvs()
	return Config{
		Port:        os.Getenv("PORT"),
		LogDir:      os.Getenv("LOG_DIR"),
		LogFilename: os.Getenv("LOG_FILE"),
		DBDir:       os.Getenv("DB_DIR"),
		DBFilename:  os.Getenv("DB_FILE"),
	}
}

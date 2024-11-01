package main

import (
	"fmt"
	"os"

	"github.com/ronbarrantes/woodchuck/utils"
)

func main() {
	utils.LoadEnvs()
	port := os.Getenv("PORT")
	logDir := os.Getenv("LOG_DIR")
	logFile := os.Getenv("LOG_FILE")

	csv := Csv(logDir, logFile)
	if err := csv.InitCSV(); err != nil {
		panic(err)
	}

	if err := csv.DeleteCSV(); err != nil {
		fmt.Errorf("There was an error deleting the file %w", err)
	}

	fmt.Println("fullpath:", csv.fullpath)

	server := Server(port)
	server.Run()
}

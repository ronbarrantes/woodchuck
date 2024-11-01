package main

import (
	"os"

	"github.com/ronbarrantes/woodchuck/utils"
)

func main() {
	utils.LoadEnvs()
	port := os.Getenv("PORT")
	logDir := os.Getenv("LOG_DIR")
	logFile := os.Getenv("LOG_FILE")

	csv := Csv(logDir, logFile)
	csv.InitCSV()

	server := Server(port)
	server.Run()
}

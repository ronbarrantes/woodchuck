package main

import (
	"os"

	"github.com/ronbarrantes/woodchuck/utils"
)

func main() {
	utils.LoadEnvs()
	PORT := os.Getenv("PORT")
	server := Server(PORT)
	csv := Csv()
	csv.InitCSV()
	server.Run()
}

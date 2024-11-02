package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// GenerateRandomID generates a random string of a fixed length.
// The generated string consists of digits and lowercase letters.
// It returns the generated string and an error if the random bytes
// could not be generated.
func GenerateRandomID() (string, error) {
	length := 10
	keys := "1234567890abcdefghijklmnopqrstuvwxyz"
	keyLen := byte(len(keys))

	var sb strings.Builder
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random bytes: %v", err)
	}

	for _, b := range randomBytes {
		sb.WriteByte(keys[b%keyLen])
	}

	return sb.String(), nil
}

// MakeMapToArray converts a map with any key and value types to a slice of values.
// It returns the slice of values and an error if any occurs.
func MakeMapToArray[Key comparable, Value any](m map[Key]Value) ([]Value, error) {
	if m == nil {
		return []Value{}, nil
	}

	result := make([]Value, 0, len(m))
	for _, value := range m {
		result = append(result, value)
	}
	return result, nil
}

// LoadEnvs sets up the ability to load env variables
func LoadEnvs() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

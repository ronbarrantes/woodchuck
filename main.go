package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ronbarrantes/woodchuck/utils"
)

type Config struct {
	Port        string
	LogDir      string
	LogFilename string
}

func main() {
	config := loadConfig()

	csv := Csv(config.LogDir, config.LogFilename)

	_, err := csv.InitCSV()
	if err != nil {
		panic(err)
	}

	server := Server(config.Port, csv)
	server.Run()
}

func loadConfig() Config {
	utils.LoadEnvs()
	return Config{
		Port:        os.Getenv("PORT"),
		LogDir:      os.Getenv("LOG_DIR"),
		LogFilename: os.Getenv("LOG_FILE"),
	}
}

// ### CONSTS AND VARS ###
const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// ### TYPES ###
type LogIDGenerator struct {
	ID int
	mu sync.Mutex
}

type ApiServer struct {
	listenAddress string
	logCounter    *LogIDGenerator
	csvFile       *CsvFile
}

type LogLevel string

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	LogID     int       `json:"log_id"`
	LogLevel  LogLevel  `json:"level"` // "error", "warning", "log"
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
}

// ### FUNCTIONS ###
func Server(address string, csv *CsvFile) *ApiServer {
	lastId, err := csv.ReadLastLogID()
	if err != nil {
		fmt.Println("First log id will be initalize to 0")
		lastId = 0
	}

	return &ApiServer{
		listenAddress: address,
		logCounter:    NewLogCounter(lastId),
		csvFile:       csv,
	}
}

// ### METHODS ###
func NewLogCounter(id int) *LogIDGenerator {
	return &LogIDGenerator{
		ID: id,
	}
}

func (id *LogIDGenerator) inc() int {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.ID++
	return id.ID
}

func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	default:
		return false
	}
}

func (l *LogLevel) UnmarshalJSON(data []byte) error {
	var level string
	if err := json.Unmarshal(data, &level); err != nil {
		return err
	}
	*l = LogLevel(level)
	if !l.IsValid() {
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

func (s *ApiServer) Run() {
	fmt.Printf("Listening to %s\n", s.listenAddress)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	router := mux.NewRouter()

	router.HandleFunc("/", s.handleMainPage)
	router.HandleFunc("/api/v1/log", s.handlePath)
	// .Methods("POST")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+s.listenAddress, corsHandler(router)))
}

func (s *ApiServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Woodchuck")
}

func (s *ApiServer) handlePath(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		s.handlePostLog(w, r)
	default:
		http.Error(w, fmt.Sprintf("Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

func (s *ApiServer) handlePostLog(w http.ResponseWriter, r *http.Request) {
	var log LogEntry

	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if log.Message == "" {
		http.Error(w, "The message is required", http.StatusBadRequest)
		return
	}

	registeredLog, err := s.CreateLog(r.RemoteAddr, log.LogLevel, log.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logEntry := CSVLogEntry{
		Timestamp: registeredLog.Timestamp.Format("2006-01-02T15:04:05Z"),
		LogLevel:  string(registeredLog.LogLevel),
		LogID:     registeredLog.LogID,
		UserID:    registeredLog.UserID,
		Message:   registeredLog.Message,
	}

	err = s.csvFile.WriteToCSV(&logEntry)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusOK, registeredLog)
}

func (s *ApiServer) CreateLog(uid string, lvl LogLevel, msg string) (LogEntry, error) {
	if lvl == "" || uid == "" || msg == "" {
		err := errors.New("cannot create log entry")
		return LogEntry{}, err
	}

	return LogEntry{
		Timestamp: time.Now(),
		LogLevel:  lvl,
		UserID:    uid,
		Message:   msg,
		LogID:     s.logCounter.inc(),
	}, nil
}

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
	// Initalize the CSV
	fmt.Println("Initializing...")
	// Get the name

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
		//log.Printf("There was an error %v", err)
		return 0, fmt.Errorf("Cannot get last entry: %w", err)
	}

	return lastEntry.LogID, nil
}

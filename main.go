package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ronbarrantes/woodchuck/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Port        string
	LogDir      string
	LogFilename string
	DBDir       string
	DBFilename  string
}

func main() {
	config := loadConfig()
	db := NewDBFile(config.DBDir, config.DBFilename)

	if err := db.InitDB(); err != nil {
		panic(err)
	}

	server := Server(config.Port, db)
	server.Run()
}

func loadConfig() Config {
	utils.LoadEnvs()
	return Config{
		Port:       os.Getenv("PORT"),
		DBDir:      os.Getenv("DB_DIR"),
		DBFilename: os.Getenv("DB_FILE"),
	}
}

// API // in api.go
// ### CONSTS AND VARS ###
const (
	LogLevelInfo  JSONLogLevel = "info"
	LogLevelWarn  JSONLogLevel = "warn"
	LogLevelError JSONLogLevel = "error"
)

// ### TYPES ###

type APIServer struct {
	listenAddress string
	db            *DBFile
	logChannel    chan Log
}

type JSONLogLevel string

type LogEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	LogID     int          `json:"log_id"`
	LogLevel  JSONLogLevel `json:"level"` // "error", "warning", "log"
	UserID    string       `json:"user_id"`
	Message   string       `json:"message"`
}

// ### FUNCTIONS ###
func Server(address string, db *DBFile) *APIServer {
	return &APIServer{
		listenAddress: address,
		db:            db,
		logChannel:    make(chan Log, 100), // Adjust the buffer size as needed
	}
}

// ### METHODS ###
func (l JSONLogLevel) IsValid() bool {
	switch l {
	case LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	default:
		return false
	}
}

func (l *JSONLogLevel) UnmarshalJSON(data []byte) error {
	var level string
	if err := json.Unmarshal(data, &level); err != nil {
		return err
	}
	*l = JSONLogLevel(level)
	if !l.IsValid() {
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

func (s *APIServer) Run() {
	fmt.Printf("Listening to %s\n", s.listenAddress)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	router := mux.NewRouter()

	// Serve static files from the "static" directory
	router.HandleFunc("/", s.handleMainPage)
	router.HandleFunc("/api/v1/logs", s.handlePath) // .Methods("POST")
	router.HandleFunc("/api/v1/events", s.handleSSEEvent)

	staticFileDirectory := http.Dir("./static/")
	staticFileHandler := http.StripPrefix("/", http.FileServer(staticFileDirectory))
	router.PathPrefix("/").Handler(staticFileHandler)

	log.Fatal(http.ListenAndServe("0.0.0.0:"+s.listenAddress, corsHandler(router)))
}

func (s *APIServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func (s *APIServer) handlePath(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGetLog(w, r)
	case "POST":
		s.handlePostLog(w, r)
	default:
		http.Error(w, fmt.Sprintf("Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

func (s *APIServer) handleSSEEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	fmt.Println("ABOUT TO SEND THE EVENT MESSAGE")

	for {
		select {
		case log := <-s.logChannel:
			// Assuming log has fields ID and Message similar to GetLogs
			logData := fmt.Sprintf("ID: %d, Message: %s", log.ID, log.Message)
			fmt.Fprintf(w, "data: %s\n\n", logData)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

// Gets all the logs
func (s *APIServer) handleGetLog(w http.ResponseWriter, _ *http.Request) {
	results, err := s.db.ReadLogs()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var logs []LogEntry

	for _, log := range results {

		logs = append(logs, LogEntry{
			Timestamp: log.CreatedAt.UTC(),
			UserID:    log.UserID,
			LogLevel:  JSONLogLevel(log.LogLevel),
			LogID:     int(log.ID),
			Message:   log.Message,
		})
	}

	utils.WriteJSON(w, http.StatusOK, logs)
}

func (s *APIServer) handlePostLog(w http.ResponseWriter, r *http.Request) {
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

	newLog, err := s.db.WriteLog(registeredLog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("ALMOST DONE WITH THE LOG")

	s.logChannel <- newLog

	fmt.Println("DONE WITH THE NEW LOG")

	utils.WriteJSON(w, http.StatusOK, registeredLog)
}

func (s *APIServer) CreateLog(uid string, lvl JSONLogLevel, msg string) (Log, error) {
	if lvl == "" || uid == "" || msg == "" {
		err := errors.New("cannot create log entry")
		return Log{}, err
	}

	return Log{
		LogLevel: string(lvl),
		UserID:   uid,
		Message:  msg,
	}, nil
}

// / SQL in sql.go
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

type Log struct {
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
	if err := db.AutoMigrate(&Log{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	f.db = db
	return nil
}

// Write to the logs
func (f *DBFile) WriteLog(log Log) (Log, error) {
	if f.db == nil {
		return Log{}, fmt.Errorf("database not initialized")
	}

	result := f.db.Create(&log)
	if result.Error != nil {
		return Log{}, result.Error
	}

	fmt.Println("The log --->>", log)

	return log, nil
}

// Read the logs
func (f *DBFile) ReadLogs() ([]Log, error) {
	if f.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var logs []Log
	result := f.db.Find(&logs)
	if result.Error != nil {
		return nil, result.Error
	}

	return logs, nil
}

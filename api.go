package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ronbarrantes/woodchuck/utils"
)

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
func Server(address string) *ApiServer {
	// lastId, err := csv.ReadLastLogID()
	// if err != nil {
	// 	fmt.Println("First log id will be initalize to 0")
	// 	lastId = 0
	// }

	lastId := 0
	return &ApiServer{
		listenAddress: address,
		logCounter:    NewLogCounter(lastId),
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

	// logEntry := CSVLogEntry{
	// 	Timestamp: registeredLog.Timestamp.Format("2006-01-02T15:04:05Z"),
	// 	LogLevel:  string(registeredLog.LogLevel),
	// 	LogID:     registeredLog.LogID,
	// 	UserID:    registeredLog.UserID,
	// 	Message:   registeredLog.Message,
	// }

	// err = s.csvFile.WriteToCSV(&logEntry)

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

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

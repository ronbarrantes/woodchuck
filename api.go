package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

type ApiServer struct {
	listenAddress string
	db            *DBFile
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
func Server(address string, db *DBFile) *ApiServer {
	return &ApiServer{
		listenAddress: address,
		db:            db,
	}
}

// ### METHODS ###
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

	// Serve static files from the "static" directory
	router.HandleFunc("/", s.handleMainPage)
	router.HandleFunc("/api/v1/logs", s.handlePath) // .Methods("POST")

	staticFileDirectory := http.Dir("./static/")
	staticFileHandler := http.StripPrefix("/", http.FileServer(staticFileDirectory))
	router.PathPrefix("/").Handler(staticFileHandler)

	log.Fatal(http.ListenAndServe("0.0.0.0:"+s.listenAddress, corsHandler(router)))
}

func (s *ApiServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func (s *ApiServer) handlePath(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGetLog(w, r)
	case "POST":
		s.handlePostLog(w, r)
	default:
		http.Error(w, fmt.Sprintf("Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

// Gets all the logs
func (s *ApiServer) handleGetLog(w http.ResponseWriter, _ *http.Request) {
	results, err := s.db.ReadLogs()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var logs []LogEntry

	for _, log := range results {

		logs = append(logs, LogEntry{
			Timestamp: log.CreatedAt,
			UserID:    log.UserID,
			LogLevel:  LogLevel(log.LogLevel),
			LogID:     int(log.ID),
			Message:   log.Message,
		})
	}

	utils.WriteJSON(w, http.StatusOK, logs)
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

	if err := s.db.WriteLog(registeredLog); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusOK, registeredLog)
}

func (s *ApiServer) CreateLog(uid string, lvl LogLevel, msg string) (DBLogModel, error) {
	if lvl == "" || uid == "" || msg == "" {
		err := errors.New("cannot create log entry")
		return DBLogModel{}, err
	}

	return DBLogModel{
		LogLevel: string(lvl),
		UserID:   uid,
		Message:  msg,
	}, nil
}

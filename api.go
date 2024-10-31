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

// 	currentTime := time.Now()
// 	formattedTime := currentTime.Format("2006-01-02 15:04:05")

type LogState int

type LogNumber struct {
	ID int
	mu sync.Mutex
}

type ApiServer struct {
	listenAddress string
	logCounter    *LogNumber
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	LogID     int       `json:"log_id"`
	LogLevel  string    `json:"level"` // LOG ERROR WARNING
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	// Source    string // maybe
	// Trace     string // maybe
}

const (
	LOG LogState = iota
	WARN
	ERROR
)

var stateName = map[LogState]string{
	LOG:   "LOG",
	WARN:  "WARN",
	ERROR: "ERROR",
}

func (ls LogState) String() string {
	return stateName[ls]
}

func (s *ApiServer) Run() {
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	router := mux.NewRouter()

	// Have a place to see the logs (HTML) AND (Server)
	router.HandleFunc("/", s.handleMainPage)

	router.HandleFunc("/v1/api/log", s.handlePath)

	fmt.Printf("Listening to %s\n", s.listenAddress)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+s.listenAddress, corsHandler(router)))
}

func Server(address string) *ApiServer {

	return &ApiServer{
		listenAddress: address,
		logCounter:    NewLogID(),
	}
}

// may put some kind of UI on here
func (s *ApiServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Woodchuck")
}

// In Charge of setting the POST route
func (s *ApiServer) handlePath(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		s.handlePostLog(w, r)
	default:
		http.Error(w, fmt.Sprintf("Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

func (id *LogNumber) inc() int {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.ID++
	return id.ID
}

func NewLogID() *LogNumber {
	return &LogNumber{
		ID: 0,
	}
}

func CreateLog(uid string, lvl string, msg string, lid *LogNumber) (LogEntry, error) {

	if lvl == "" || uid == "" || msg == "" {
		fmt.Println(lvl, uid, msg, lid.ID)
		err := errors.New("Cannot create log entry")
		fmt.Printf("%s\n", err)
		return LogEntry{}, err
	}

	return LogEntry{
		Timestamp: time.Now(),
		LogLevel:  lvl,
		UserID:    uid,
		Message:   msg,
		LogID:     lid.inc(),
	}, nil
}

func (s *ApiServer) handlePostLog(w http.ResponseWriter, r *http.Request) {
	var log LogEntry

	// Decode the JSON body into the struct
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if log.LogLevel == "" || log.Message == "" {
		http.Error(w, "The level and message are required", http.StatusBadRequest)
		return
	}

	registeredLog, err := CreateLog(r.RemoteAddr, log.LogLevel, log.Message, s.logCounter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusOK, registeredLog)
}

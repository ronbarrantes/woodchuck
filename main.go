package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

// ### API SERVER ###
type ApiServer struct {
	listenAddress string
}

type LogState int

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

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	LogID     int       `json:"logId"`
	LogLevel  string    `json:"level"` // LOG ERROR WARNING
	UserID    string    `json:"userId"`
	Message   string    `json:"message"`
	// Source    string // maybe
	// Trace     string // maybe
}

func (s *ApiServer) Run() {
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	router := mux.NewRouter()

	router.HandleFunc("/v1/api/log", s.handlePath)
	// Have a place to see the logs (HTML) AND (Server)

	router.HandleFunc("/", s.handleMainPage)

	fmt.Printf("Listening to %s\n", s.listenAddress)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+s.listenAddress, corsHandler(router)))
}

func Server(address string) *ApiServer {
	return &ApiServer{
		listenAddress: address,
	}
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

type LogIDIncrementor struct {
	ID int
	mu sync.Mutex
}

func (id *LogIDIncrementor) inc() int {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.ID++
	return id.ID
}

func NewLogID() *LogIDIncrementor {
	return &LogIDIncrementor{
		ID: 0,
	}

}

func CreateLog(lvl string, uid string, msg string) (LogEntry, error) {
	var log LogEntry

	logId := NewLogID()
	// 	currentTime := time.Now()
	// 	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	fmt.Println("level", lvl)
	if lvl == "" {
		fmt.Println("soemthing wrong with level")
		err := errors.New("cannot create log entry")
		return LogEntry{}, err
	}

	// if uid != "" {
	// 	fmt.Println("soemthing wrong with uid")
	// 	err := errors.New("cannot create uid entry")
	// 	return LogEntry{}, err
	// }

	// if msg != "" {
	// 	fmt.Println("soemthing wrong with msg")
	// 	err := errors.New("cannot create msg entry")
	// 	return LogEntry{}, err
	// }

	if lvl == "" || uid == "" || msg == "" {
		fmt.Println(lvl, uid, msg, logId)
		err := errors.New("cannot create log entry")
		fmt.Printf("%s\n", err)
		return LogEntry{}, err
	}

	log = LogEntry{
		Timestamp: time.Now(),
		LogLevel:  lvl,
		UserID:    uid,
		Message:   msg,
		LogID:     logId.inc(),
	}

	return log, nil
}

func (s *ApiServer) handlePostLog(w http.ResponseWriter, r *http.Request) {
	var log LogEntry

	// Decode the JSON body into the struct
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if log.LogLevel == "" || log.UserID == "" || log.Message == "" {
		fmt.Println(log.LogLevel, log.UserID, log.Message)
		http.Error(w, "level, userId, and message are required", http.StatusBadRequest)
		return
	}

	registeredLog, err := CreateLog(log.LogLevel, log.UserID, log.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJSON(w, http.StatusOK, registeredLog)
}

// #### CSV FILE ####
type CsvFile struct {
	filename string
	path     string
}

func Csv() *CsvFile {
	return &CsvFile{}
}

func (f *CsvFile) DeleteCSV() {
	fmt.Println("Deleting...")
}

func (f *CsvFile) InitCSV() {
	// Initalize the CSV
	fmt.Println("Initializing...")
}

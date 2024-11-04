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

	fmt.Println(csv.fullpath)
	server := Server(port)
	server.Run()
}

// package main

// import (
//     "github.com/gorilla/mux"
//     "github.com/gorilla/websocket"
//     "net/http"
// )

// var upgrader = websocket.Upgrader{
//     CheckOrigin: func(r *http.Request) bool {
//         return true
//     },
// }

// var connections = make(map[*websocket.Conn]bool)

// func main() {
//     router := mux.NewRouter()
//     router.HandleFunc("/ws", handleConnections)
//     http.ListenAndServe(":8080", router)
// }

// func handleConnections(w http.ResponseWriter, r *http.Request) {
//     ws, err := upgrader.Upgrade(w, r, nil)
//     if err != nil {
//         return
//     }
//     defer ws.Close()
//     connections[ws] = true

//     for {
//         _, msg, err := ws.ReadMessage()
//         if err != nil {
//             delete(connections, ws)
//             break
//         }
//         for conn := range connections {
//             if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
//                 delete(connections, conn)
//             }
//         }
//     }
// }

// <!DOCTYPE html>
// <html>
// <head>
//     <title>WebSocket Example</title>
// </head>
// <body>
//     <h1>WebSocket Example</h1>
//     <script>
//         const ws = new WebSocket('ws://localhost:8080/ws');
//         ws.onmessage = function(event) {
//             console.log('Message from server:', event.data);
//         };
//         ws.onopen = function() {
//             ws.send('Hello Server!');
//         };
//     </script>
// </body>
// </html>

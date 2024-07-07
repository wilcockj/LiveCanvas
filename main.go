package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	ClientID string  `json:"clientID"`
	X0       float64 `json:"x0"`
	Y0       float64 `json:"y0"`
	X1       float64 `json:"x1"`
	Y1       float64 `json:"y1"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients     = make(map[*websocket.Conn]string)
	broadcast   = make(chan Message)
	canvasState = []Message{}
	mu          sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clientID := ws.RemoteAddr().String()
	mu.Lock()
	clients[ws] = clientID
	mu.Unlock()

	// Send the current canvas state to the new client
	mu.Lock()
	for _, msg := range canvasState {
		if err := ws.WriteJSON(msg); err != nil {
			log.Printf("error: %v", err)
			ws.Close()
			delete(clients, ws)
			break
		}
	}
	mu.Unlock()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}

		// Store the message in the canvas state
		mu.Lock()
		canvasState = append(canvasState, msg)
		mu.Unlock()

		msg.ClientID = clientID
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client, id := range clients {
			if id != msg.ClientID {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
		mu.Unlock()
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	port := ":8080"
	log.Printf("Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

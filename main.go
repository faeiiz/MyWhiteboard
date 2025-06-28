package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients  = make(map[*websocket.Conn]bool)
	mutex    = sync.Mutex{}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	log.Println("New client connected")

	for {

		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		mutex.Lock()
		for client := range clients {
			if client == ws {
				continue
			}
			err := client.WriteMessage(messageType, msg)
			if err != nil {
				log.Printf("Write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}

	mutex.Lock()
	delete(clients, ws)
	mutex.Unlock()
	log.Println("Client disconnected")
}

func main() {
	// Serve static files from "./static"
	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	log.Println("Server started on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

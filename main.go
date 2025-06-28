package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	ID       string // UUID
	Username string
}

var (
	// clients holds the active WebSocket connections
	clients = make(map[*websocket.Conn]User)

	// mutex is used to synchronize access to the clients map
	mutex = sync.Mutex{}

	messageHistory []string

	// upgrader is used to upgrader http to ws connections
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

	_, joinMsg, err := ws.ReadMessage()
	if err != nil {
		log.Printf("Join message read error: %v", err)
		return
	}
	var joinData struct {
		Type     string `json:"type"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(joinMsg, &joinData); err != nil {
		log.Printf("Invalid join message: %v", err)
		return
	}
	if joinData.Type != "join" || joinData.Username == "" {
		log.Printf("Invalid join payload")
		return
	}
	user := User{
		ID:       uuid.New().String(),
		Username: joinData.Username,
	}

	// 2️⃣ Send assigned userId back to client
	userIDPayload := fmt.Sprintf(`{"type":"joined","userId":"%s"}`, user.ID)
	if err := ws.WriteMessage(websocket.TextMessage, []byte(userIDPayload)); err != nil {
		log.Printf("Write joined error: %v", err)
		return
	}

	mutex.Lock()
	clients[ws] = user
	for _, saved := range messageHistory {
		err := ws.WriteMessage(websocket.TextMessage, []byte(saved))
		if err != nil {
			log.Printf("Write error (replay): %v", err)
			ws.Close()
			delete(clients, ws)
			mutex.Unlock()
			return
		}
	}
	broadcastUserCount()
	mutex.Unlock()

	log.Println("New client connected", user.Username, user.ID)

	for {
		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		// 3️⃣ Wrap message to include user info
		enriched := map[string]interface{}{}
		if err := json.Unmarshal(msg, &enriched); err != nil {
			log.Printf("Invalid message JSON: %v", err)
			continue
		}
		enriched["userId"] = user.ID
		enriched["username"] = user.Username

		if enriched["type"] == "draw" {
			enriched["strokeId"] = uuid.New().String()
		}

		encoded, _ := json.Marshal(enriched)

		mutex.Lock()
		messageHistory = append(messageHistory, string(encoded))
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
	log.Println("Client disconnected", user.Username)
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

func broadcastUserCount() {
	userCountMsg := []byte(fmt.Sprintf(`{"type":"userCount","count":%d}`, len(clients)))
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, userCountMsg)
		if err != nil {
			log.Printf("Write error (userCount): %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

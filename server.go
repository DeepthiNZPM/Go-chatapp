package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader to handle WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Store connected clients
var broadcast = make(chan string)            // Channel to broadcast messages
var mu sync.Mutex

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading WebSocket:", err)
		return
	}
	defer conn.Close()

	// Register the new client
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	// Listen for messages from the client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		broadcast <- string(msg) // Send the message to the broadcast channel
	}

	// Remove the client when disconnected
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}

func handleMessages() {
	for {
		// Get the message from the broadcast channel
		msg := <-broadcast
		mu.Lock()
		// Send the message to all connected clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func main() {
	// Set up WebSocket route
	http.HandleFunc("/ws", handleConnections)

	// Start message handler in a new goroutine
	go handleMessages()

	// Start the server
	port := "8080"
	fmt.Println("Server started on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

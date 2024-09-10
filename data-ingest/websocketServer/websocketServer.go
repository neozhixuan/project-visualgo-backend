package websocketServer

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Set up the WebSocket upgrader
// - The connection starts as a HTTP request; an upgrader upgrades it to a WebSocket
// - Configure buffer sizes
// - Allow any origin using a callback
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Initialise an empty boolean dictionary for WebSocket connections
var clients = make(map[*websocket.Conn]bool)

// Handle all clients that are trying to connect to our WSS server
// Read messages from each client
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Use our upgrader to upgrade the request and response into WSS
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Track our new client in our boolean map
	clients[ws] = true

	// Read each message from the client
	// - if there is an error, remove this client from our server
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
	}
}

// Listens for messages from the `broadcast` channel
// Then for each client, we write the message to them
func HandleMessages(broadcast chan []byte) {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			// If there is an error writing to a client, we remove that client
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

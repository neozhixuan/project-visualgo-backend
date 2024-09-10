package websocketServer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func StartWebSocketServer(ema9Channel chan []float64) {
	// Configure WebSocket upgrade
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
	}

	// Handle WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		conn2, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection to WebSocket: %v", err)
			return
		}
		defer conn2.Close()

		// Broadcast ema9 data to this WebSocket client
		for {
			ema9 := <-ema9Channel
			// Convert ema9 to JSON
			jsonBytes, err := json.Marshal(ema9)
			if err != nil {
				log.Printf("Error marshaling ema9 to JSON: %v", err)
				return
			}

			// Write the JSON data to the WebSocket connection
			fmt.Println("Writing JSON data to websocket")
			err = conn2.WriteMessage(websocket.TextMessage, jsonBytes)
			if err != nil {
				log.Printf("Error sending ema9 data to WebSocket client: %v", err)
				return
			}
		}
	})

	// Start HTTP server for WebSocket connections
	log.Println("Starting WebSocket server on port 8090")
	http.ListenAndServe(":8090", nil)
}

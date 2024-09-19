package main

import (
	"fmt"
	"net/http"

	"github.com/neozhixuan/project-visualgo-backend/trading-algo/grpcClient"
	"github.com/neozhixuan/project-visualgo-backend/trading-algo/websocketServer"
)

// Define a global channel to send ema9 data to the WebSocket server
var ema9Channel = make(chan []float64, 10)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Server is up and running!")
}

func main() {
	// Start gRPC client in a separate goroutine
	go grpcClient.StartGRPCClient(ema9Channel)

	go http.HandleFunc("/health", healthCheckHandler)

	// Start WebSocket server
	// - this is not a goroutine so the server does not stop
	websocketServer.StartWebSocketServer(ema9Channel)

	// - Alternatively, create a blocking channel that triggers upon closure of client -
	// done := make(chan bool)

	// go func() {
	// 	if err := s.Serve(lis); err != nil {
	// 		log.Fatalf("failed to serve gRPC: %v", err)
	// 	}
	// 	// After gRPC server finishes, signal done
	// 	done <- true
	// }()

	// // Block until we receive a signal from the goroutine
	// <-done
}

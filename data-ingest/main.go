package main

import (
	"log"
	"net/http"

	"github.com/neozhixuan/project-visualgo-backend/data-ingest/grpcServer"
	"github.com/neozhixuan/project-visualgo-backend/data-ingest/websocketClient"
	"github.com/neozhixuan/project-visualgo-backend/data-ingest/websocketServer"

	"github.com/neozhixuan/project-visualgo-backend/pb"
)

func main() {
	// Introduction to Goroutines
	// - only starts when the code execution reaches that line
	// - after which, it executes concurrently with current flow
	// - if current function completes, goroutine is shut down

	// Introduction to Channels
	// - Channels in Go are unbuffered by default.
	// - This means any send operation (ema9Channel <- "message") will block until another goroutine reads from the channel
	// - We can specify the size we want (eg. make(chan T, 5)) which will only block after 5 messages of type T.

	// When to use buffer in channels?
	// - Unbuffered: synchronize communication between goroutines (sender will block until the receiver reads from the channel)
	// - Buffered: allow non-blocking sends up to a certain capacity // if you expect that there might be a delay in receiving the message
	// -- Cons of buffered: high memory usage

	// Initialise an empty channel to send messages to WSS clients
	var broadcast = make(chan []byte)

	// Initialise our tradeDataChan to send data to our gRPC clients
	tradeDataChan := make(chan *pb.KlineData)

	// Write a message from the `broadcast` channel to each client
	// - NOTE: Golang will process the code above before processing this goroutine
	go websocketServer.HandleMessages(broadcast)

	//////////////////////////////////////////////////////////////////////////
	// Start a HTTP server on port 8080
	// - Use a WSS Upgrader (handleConnections(r, w)) to upgrade the requests and response from HTTP to WSS
	// -- e.g. upgrade HTTP requests, handle WSS data
	//////////////////////////////////////////////////////////////////////////
	go http.ListenAndServe(":8080", nil)
	log.Println("WebSocket server started on port 8080...")

	// Alternatively, we can listen and serve separately as well
	// - create a server, create a listener, serve the listener using server
	//    server := &http.Server{Addr: ":8080"}
	//    listener, err := net.Listen("tcp", ":8080")
	//    if err != nil {
	//     	log.Fatalf("Failed to listen: %v", err)
	//    }
	//    err = server.Serve(listener)
	//    if err != nil {
	//    	log.Fatalf("Failed to serve: %v", err)
	//    }

	http.HandleFunc("/ws", websocketServer.HandleConnections)
	//////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////
	// Create a client WSS connection to Binance's WSS server
	// - uses `websocket.DefaultDialer.Dial(websocketURL, requestHeader)`
	// - read each message from websocket connection `c`
	// - the messages are sent into `broadcast` for our WSS clients
	// - the messages are also sent to `tradeDataChan` for our gRPC clients
	//////////////////////////////////////////////////////////////////////////
	// Introduction to WebSockets
	// - WebSocket upgrade happens via a standard HTTP header negotiation (e.g., Upgrade: websocket)
	// - The use of upgrade mechanic is common even in Node.js socketIO, Python Django Channels

	go websocketClient.BinanceConnection(broadcast, tradeDataChan)
	//////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////
	// Start our own gRPC server on port 50051
	// - NOTE: This cannot be a goroutine, else the application stops completely
	//////////////////////////////////////////////////////////////////////////
	grpcServer.StartgrpcServer(tradeDataChan)
	//////////////////////////////////////////////////////////////////////////
}

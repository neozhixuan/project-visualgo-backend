package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/neozhixuan/project-visualgo-backend/pb"
	"google.golang.org/grpc"
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

// The gRPC server has this type
// 1. The Protobuf service server (that is not implemented yet)
// 2. The channel where data (kline information) is sent to and streamed to clients
type server struct {
	pb.UnimplementedKlineServiceServer
	tradeDataChan chan *pb.KlineData
}

// gRPC method to start streaming trade data to the client
// - data comes from the server's `tradeDataChan`
func (s *server) StreamKlines(req *pb.TradeRequest, stream pb.KlineService_StreamKlinesServer) error {
	log.Printf("Client requested to start streaming trades: %s", req.Message)
	for tradeData := range s.tradeDataChan {
		if err := stream.Send(tradeData); err != nil {
			return err // handle gRPC stream errors
		}
	}
	return nil
}

// Initialise an empty boolean dictionary for WebSocket connections
var clients = make(map[*websocket.Conn]bool)

// Initialise an empty channel to send messages to all connected clients
var broadcast = make(chan []byte)

// Helper function to safely parse price values from string -> float64
func parsePrice(val interface{}) float64 {
	if str, ok := val.(string); ok {
		if price, err := strconv.ParseFloat(str, 64); err == nil {
			return price
		}
	}
	return 0
}

// Track all the clients that are trying to connect to our WSS server
func handleConnections(w http.ResponseWriter, r *http.Request) {
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
func handleMessages() {
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

func main() {
	// Write a message from the `broadcast` channel to each client
	go handleMessages()

	//////////////////////////////////////////////////////////////////////////
	// Create a gRPC server
	//////////////////////////////////////////////////////////////////////////
	s := grpc.NewServer()

	// Initialise our tradeDataChan to send data to our clients
	tradeDataChan := make(chan *pb.KlineData)

	// Register our server + channel as a service server
	pb.RegisterKlineServiceServer(s, &server{tradeDataChan: tradeDataChan})
	//////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////
	// Create a WSS connection to Binance's WebSocket
	// - uses `websocket.DefaultDialer.Dial(websocketURL, requestHeader)`
	// - all messages go into the channel `c`
	// - read each message from channel
	//////////////////////////////////////////////////////////////////////////
	binanceURL := url.URL{
		Scheme: "wss",
		Host:   "stream.binance.com:9443",
		Path:   "/ws/bnbbtc@trade",
	}
	log.Printf("Connecting to Binance WebSocket at %s", binanceURL.String())
	c, _, err := websocket.DefaultDialer.Dial(binanceURL.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to Binance WebSocket: %v", err)
	}
	defer c.Close() // Close the channel when the function stops

	// Handle incoming messages from Binance
	go func() {
		for {
			// Read messages indefinitely
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// Add message into the `broadcast` channel
			log.Printf("Received raw message: %s", string(message))
			broadcast <- message

			// Unmarshal the message into a map
			var msgMap map[string]interface{}
			if err := json.Unmarshal(message, &msgMap); err != nil {
				log.Printf("Error unmarshaling into map: %v", err)
				continue
			}

			// If the value in msgMap["e"] is a string
			// - it is stored in `eventType`
			// - ok = True
			// Then, if the `eventType` is a kline,
			// - we take the kline data from the map
			if eventType, ok := msgMap["e"].(string); ok && eventType == "kline" {
				kData, ok := msgMap["k"].(map[string]interface{})
				if !ok {
					log.Println("Error extracting kline data from message")
					continue
				}

				// Create KlineData protobuf message and populate fields
				klineData := &pb.KlineData{
					Symbol:        msgMap["s"].(string),
					OpenTime:      int64(kData["t"].(float64)),
					CloseTime:     int64(kData["T"].(float64)),
					OpenPrice:     parsePrice(kData["o"]),
					ClosePrice:    parsePrice(kData["c"]),
					HighPrice:     parsePrice(kData["h"]),
					LowPrice:      parsePrice(kData["l"]),
					Volume:        parsePrice(kData["v"]),
					NumTrades:     int32(kData["n"].(float64)),
					IsKlineClosed: kData["x"].(bool),
				}

				// Send it to the gRPC client via the channel
				tradeDataChan <- klineData
			}
		}
	}()
	//////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////
	// Start our own gRPC server, after dialing Binance WSS
	//////////////////////////////////////////////////////////////////////////
	// Listen for gRPC server on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Start gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	//////////////////////////////////////////////////////////////////////////
	// Start frontend WebSocket server
	//////////////////////////////////////////////////////////////////////////
	// Starts our HTTP server on localhost:8080
	go http.ListenAndServe(":8080", nil)
	log.Println("WebSocket server started on :8080...")

	// When the client calls localhost:8080/ws, the request and response is handled by handleConnections(r, w)
	// - e.g. upgrade HTTP requests, handle WSS data
	http.HandleFunc("/ws", handleConnections)

	// Send subscription message to the Binance WSS Connection
	subscribe := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			"bnbbtc@kline_1m", // Stream name
		},
		"id": 1,
	}
	if err := c.WriteJSON(subscribe); err != nil {
		log.Fatal("write:", err)
	}
}

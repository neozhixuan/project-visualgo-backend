package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/neozhixuan/project-visualgo-backend/pb"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

type server struct {
	pb.UnimplementedKlineServiceServer
	tradeDataChan chan *pb.KlineData
}

// gRPC method to start streaming trade data to the client
func (s *server) StreamKlines(req *pb.TradeRequest, stream pb.KlineService_StreamKlinesServer) error {
	log.Printf("Client requested to start streaming trades: %s", req.Message)
	for tradeData := range s.tradeDataChan {
		if err := stream.Send(tradeData); err != nil {
			return err // handle gRPC stream errors
		}
	}
	return nil
}

// // Handle incoming WebSocket messages and forward them to the gRPC stream
// func handleWebSocketConn(ws *websocket.Conn, tradeDataChan chan *pb.TradeData) {
// 	defer ws.Close()
// 	for {
// 		_, _, err := ws.ReadMessage()
// 		if err != nil {
// 			log.Println("WebSocket read error:", err)
// 			break
// 		}
// 		// Assume message is JSON that needs to be unmarshaled into pb.TradeData
// 		tradeData := &pb.TradeData{}
// 		// Unmarshal message to tradeData (skipped for brevity)
// 		tradeDataChan <- tradeData // send it to the gRPC streaming channel
// 	}
// }

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)            // broadcast channel
// Helper function to safely parse price values
func parsePrice(val interface{}) float64 {
	if str, ok := val.(string); ok {
		if price, err := strconv.ParseFloat(str, 64); err == nil {
			return price
		}
	}
	return 0
}

// Handle connections from the frontend
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
	}
}

// Broadcast messages to all connected clients
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	err := godotenv.Load() // Load .env file
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

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
	defer c.Close()

	go handleMessages()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tradeDataChan := make(chan *pb.KlineData)
	pb.RegisterKlineServiceServer(s, &server{tradeDataChan: tradeDataChan})
	// Handle incoming messages
	// Handle incoming messages
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			log.Printf("Received raw message: %s", string(message))
			broadcast <- message

			var msgMap map[string]interface{}
			if err := json.Unmarshal(message, &msgMap); err != nil {
				log.Printf("Error unmarshaling into map: %v", err)
				continue
			}

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

				tradeDataChan <- klineData
			}
		}
	}()

	// Start frontend WebSocket server
	http.HandleFunc("/ws", handleConnections)
	go http.ListenAndServe(":8080", nil)

	log.Println("WebSocket server started on :8080...")

	// Subscribe to the stream
	subscribe := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			"bnbbtc@kline_1m", // Stream name
		},
		"id": 1,
	}

	// Send subscription message
	if err := c.WriteJSON(subscribe); err != nil {
		log.Fatal("write:", err)
	}

	// Start gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
	}
}

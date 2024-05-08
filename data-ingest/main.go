package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"

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
	pb.UnimplementedTradeServiceServer
	tradeDataChan chan *pb.TradeData
}

// gRPC method to start streaming trade data to the client
func (s *server) StreamTrades(req *pb.TradeRequest, stream pb.TradeService_StreamTradesServer) error {
	log.Printf("Client requested to start streaming trades: %s", req.Message)
	for tradeData := range s.tradeDataChan {
		if err := stream.Send(tradeData); err != nil {
			return err // handle gRPC stream errors
		}
	}
	return nil
}

// Handle incoming WebSocket messages and forward them to the gRPC stream
func handleWebSocketConn(ws *websocket.Conn, tradeDataChan chan *pb.TradeData) {
	defer ws.Close()
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		// Assume message is JSON that needs to be unmarshaled into pb.TradeData
		tradeData := &pb.TradeData{}
		// Unmarshal message to tradeData (skipped for brevity)
		tradeDataChan <- tradeData // send it to the gRPC streaming channel
	}
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)            // broadcast channel

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

	// Setup WebSocket connection to Finnhub.
	apiKey := os.Getenv("FINNHUB_API_KEY") // Replace with your actual Finnhub API key
	u := url.URL{
		Scheme:   "wss",
		Host:     "ws.finnhub.io",
		Path:     "/",
		RawQuery: "token=" + apiKey,
	}
	log.Printf("Connecting to websocket")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go handleMessages()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tradeDataChan := make(chan *pb.TradeData)
	pb.RegisterTradeServiceServer(s, &server{tradeDataChan: tradeDataChan})

	// Handle incoming messages from Finnhub
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Println("Data")
			// Broadcast message to all connected frontend clients
			broadcast <- message
			// Parse JSON message into struct
			var tradeDataJSON struct {
				Data []struct {
					C float64 `json:"c"` // Assuming "c" is a float64 value
					P float64 `json:"p"`
					S string  `json:"s"`
					T int64   `json:"t"`
					V float64 `json:"v"`
				} `json:"data"`
			}
			if err := json.Unmarshal(message, &tradeDataJSON); err != nil {
				log.Printf("Error unmarshaling trade data JSON: %v", err)
				continue // Skip processing this message
			}

			// Convert JSON data to TradeData protobuf messages
			for _, data := range tradeDataJSON.Data {
				// Create TradeData protobuf message and populate fields
				tradeData := &pb.TradeData{
					Price:     data.P,
					Symbol:    data.S,
					Timestamp: data.T,
					Volume:    data.V,
					// Populate other fields as needed
				}

				// Send tradeData to gRPC streaming channel
				tradeDataChan <- tradeData
			}
		}
	}()

	// Start frontend WebSocket server
	http.HandleFunc("/ws", handleConnections)
	go http.ListenAndServe(":8080", nil)

	log.Println("WebSocket server started on :8080...")

	// Send subscription message to Finnhub
	ticker := "BINANCE:BTCUSDT"
	msg := `{"type":"subscribe","symbol":"` + ticker + `"}`
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
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

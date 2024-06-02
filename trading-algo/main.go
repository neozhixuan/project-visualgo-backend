package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	pb "github.com/neozhixuan/project-visualgo-backend/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Candlestick struct {
	Open, High, Low, Close float64
	Volume                 float64
}

// Define a global channel to send ema9 data to the WebSocket server
var ema9Channel = make(chan []float64)

func main() {
	// Start gRPC client in a separate goroutine
	go startGRPCClient()

	// Start WebSocket server
	startWebSocketServer()
}

func startGRPCClient() {
	// Set up a connection to the gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewKlineServiceClient(conn)

	// Contact the server and print out its response.
	ctx := context.Background() // No timeout
	cancel := func() {}         // No-op cancel function since there's no timeout
	defer cancel()              // It's still good practice to call defer cancel() even if it does nothing in this case

	stream, err := c.StreamKlines(ctx, &pb.TradeRequest{Message: "start_stream"})
	if err != nil {
		log.Fatalf("could not stream trades: %v", err)
	}
	var candlesticks []Candlestick

	// Handle streamed responses
	for {
		tradeData, err := stream.Recv()
		if err == io.EOF {
			// End of stream
			break
		}
		if err != nil {
			log.Printf("error while receiving: %v", err)
			// Handle the error (e.g., retry, log, etc.)
			continue // or return, depending on your error handling strategy
		}
		log.Printf("Received: %s", tradeData)
		if tradeData.IsKlineClosed {
			// Create Candlestick struct and append to candlesticks slice
			candle := Candlestick{
				Open:   tradeData.OpenPrice,
				High:   tradeData.HighPrice,
				Low:    tradeData.LowPrice,
				Close:  tradeData.ClosePrice,
				Volume: tradeData.Volume,
			}
			candlesticks = append(candlesticks, candle)

			// Calculate 9-EMA
			ema9 := CalculateEMA(candlesticks, 9)

			// Send ema9 data to the WebSocket server via channel
			fmt.Println("Attempting to send ema data")
			ema9Channel <- ema9
		}
	}
}

func startWebSocketServer() {
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

// CalculateEMA calculates the Exponential Moving Average (EMA)
func CalculateEMA(candlesticks []Candlestick, period int) []float64 {
	k := 2 / float64(period+1)
	prices := extractClosingPrices(candlesticks)
	ema := make([]float64, len(prices))
	ema[0] = prices[0]

	for i := 1; i < len(prices); i++ {
		ema[i] = prices[i]*k + ema[i-1]*(1-k)
	}
	return ema
}

// CalculateVWAP calculates the Volume Weighted Average Price (VWAP)
func CalculateVWAP(candlesticks []Candlestick) []float64 {
	vwap := make([]float64, len(candlesticks))
	cumulativePriceVolume := 0.0
	cumulativeVolume := 0.0

	for i, candle := range candlesticks {
		typicalPrice := (candle.High + candle.Low + candle.Close) / 3
		priceVolume := typicalPrice * candle.Volume
		cumulativePriceVolume += priceVolume
		cumulativeVolume += candle.Volume
		vwap[i] = cumulativePriceVolume / cumulativeVolume
	}
	return vwap
}

// Helper function to extract closing prices from candlesticks
func extractClosingPrices(candlesticks []Candlestick) []float64 {
	prices := make([]float64, len(candlesticks))
	for i, candle := range candlesticks {
		prices[i] = candle.Close
	}
	return prices
}

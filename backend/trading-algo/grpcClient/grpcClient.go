package grpcClient

import (
	"context"
	"io"
	"log"
	"os"

	pb "github.com/neozhixuan/project-visualgo-backend/pb"

	"github.com/joho/godotenv"
	"github.com/neozhixuan/project-visualgo-backend/trading-algo/financeFunctions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartGRPCClient(ema9Channel chan []float64) {
	log.Println("Hi, trying to start gRPC client")
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Define a candlestick slice to store all candlesticks
	var candlesticks []financeFunctions.Candlestick

	// Production
	stage := os.Getenv("STAGE")
	var wssUrl = "localhost:50051"
	if stage == "production" {
		wssUrl = "host.docker.internal:50051"
	}
	log.Println("Dialing now")

	// Set up a connection to the gRPC server
	conn, err := grpc.Dial(wssUrl, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	log.Println("Connected to gRPC server.")
	defer conn.Close()
	log.Println("Aftermath")

	// Set up a gRPC client using the connection object
	c := pb.NewKlineServiceClient(conn)

	// Contact the gRPC server and print out its response.
	ctx := context.Background() // No timeout
	cancel := func() {}         // No-op cancel function since there's no timeout
	defer cancel()              // It's still good practice to call defer cancel() even if it does nothing in this case
	log.Println("Sending message")

	// Send a start_stream message to the gRPC server to request a stream of data
	stream, err := c.StreamKlines(ctx, &pb.TradeRequest{Message: "start_stream"})
	if err != nil {
		log.Fatalf("could not stream trades: %v", err)
	}
	log.Println("Message shld be sent")

	// Indefinitely read the messages from gRPC server
	for {
		tradeData, err := stream.Recv()

		// End of stream
		if err == io.EOF {
			break
		}

		// Error in stream
		// - can either handle or do nothing in our case
		if err != nil {
			log.Printf("Error while receiving: %v", err)
			continue
		}

		// Received message
		log.Printf("Received: %s", tradeData)

		// If we hit the minute candlestick, calculate 9-EMA
		if tradeData.IsKlineClosed {
			// Create Candlestick struct and append to candlesticks slice
			candle := financeFunctions.Candlestick{
				Open:   tradeData.OpenPrice,
				High:   tradeData.HighPrice,
				Low:    tradeData.LowPrice,
				Close:  tradeData.ClosePrice,
				Volume: tradeData.Volume,
			}
			candlesticks = append(candlesticks, candle)

			// Calculate 9-EMA
			ema9 := financeFunctions.CalculateEMA(candlesticks, 9)

			// Send ema9 data to the WebSocket server via channel
			select {
			case ema9Channel <- ema9:
				// Successfully sent to broadcast
				log.Println("Sent a message to WSS client")
			default:
				// Handle when no one is reading from broadcast (could log or handle differently)
				log.Println("Warning: channel to WSS client is full, dropping message")
			}
		}
	}
}

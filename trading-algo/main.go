package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/neozhixuan/project-visualgo-backend/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set up a connection to the server.
	fmt.Println("hi")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewKlineServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	stream, err := c.StreamKlines(ctx, &pb.TradeRequest{Message: "start_stream"})
	if err != nil {
		log.Fatalf("could not stream trades: %v", err)
	}

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
	}

}

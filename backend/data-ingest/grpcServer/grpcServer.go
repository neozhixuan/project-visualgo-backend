package grpcServer

import (
	"log"
	"net"

	"github.com/neozhixuan/project-visualgo-backend/pb"
	"google.golang.org/grpc"
)

// The gRPC server has this type
// 1. The Protobuf service server (that is not implemented yet)
// 2. The channel where data (kline information) is sent to and streamed to clients
type server struct {
	pb.UnimplementedKlineServiceServer
	tradeDataChan chan *pb.KlineData
}

// gRPC method to start streaming trade data to the client
// - data comes from the server's `tradeDataChan`
// - NOTE: this is triggered when the gRPC client starts and sends a message to us
func (s *server) StreamKlines(req *pb.TradeRequest, stream pb.KlineService_StreamKlinesServer) error {
	log.Printf("Client requested to start streaming trades: %s", req.Message)
	for tradeData := range s.tradeDataChan {
		if err := stream.Send(tradeData); err != nil {
			return err
		}
	}
	return nil
}

func StartgrpcServer(tradeDataChan chan *pb.KlineData) {
	s := grpc.NewServer()

	// Register our server + channel as a service server
	pb.RegisterKlineServiceServer(s, &server{tradeDataChan: tradeDataChan})

	// Opens a TCP listener on port 50051
	// - When a server "listens," it waits for incoming connections on a specific port (e.g., port 8080 or 50051).
	// - It is the first step where the server is ready to accept requests but does not actually handle them yet.
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start serving incoming gRPC requests on that port
	// - Once a connection is established, the server "serves" it by handling incoming requests,
	//   executing the appropriate logic (like processing WebSocket connections, responding to HTTP requests, or handling gRPC calls),
	//   and sending back responses
	log.Println("gRPC server started on port 50051... (check for incoming error)")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

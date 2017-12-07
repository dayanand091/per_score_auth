package server

import (
	"fmt"
	"log"
	"net"
	pb "perScoreAuth/perScoreProto/user"

	"google.golang.org/grpc"
)

const address = "0.0.0.0:6060"

func StartServer() {
	fmt.Println("Server Started...")
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Creates a new gRPC server
	s := grpc.NewServer()
	pb.RegisterUserServer(s, &Server{})
	s.Serve(lis)
}

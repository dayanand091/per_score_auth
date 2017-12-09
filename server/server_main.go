package server

import (
	"fmt"
	"log"
	"net"
	pb "perScoreAuth/perScoreProto/user"

	"google.golang.org/grpc"
)

const address = "localhost:6050"

func StartServer() {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("perScoreAuth server started on :6050 ...")

	// Creates a new gRPC server
	s := grpc.NewServer()
	pb.RegisterUserServer(s, &Server{})
	s.Serve(lis)

}

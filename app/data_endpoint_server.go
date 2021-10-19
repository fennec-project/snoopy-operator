package main

import (
	"log"
	"net"

	"github.com/fennec-project/snoopy-operator/pkg/endpoint"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatalf("Failed to listen on port 5555: %v", err)
	}
	s := endpoint.Server{}

	grpcServer := grpc.NewServer()
	endpoint.RegisterDataEndpointServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000 %v", err)
	}

}

package main

import (
	"log"

	pb "github.com/fennec-project/snoopy-operator/pkg/endpoint"
	"google.golang.org/grpc"
)

func main() {

	// dail server
	conn, err := grpc.Dial(":5555", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewDataEndpointClient(conn)
	stream, err := client.ExportPodData()
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

}

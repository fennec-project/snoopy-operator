package main

import (
	"log"
	"net"
	"os"

	pb "github.com/fennec-project/snoopy-operator/pkg/endpoint"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedDataEndpointServer
}

func (s Server) ExportPodData(p *pb.PodData) *pb.Response {
	var r *pb.Response

	// open a file to write and append podData
	f, err := os.OpenFile(p.Name, os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		r.Message = err.Error()
		r.Status = 1
		log.Print(err.Error())
		return r
	}
	defer f.Close()

	_, err = f.Write(p.Data)
	if err != nil {
		r.Message = err.Error()
		r.Status = 1
		log.Print(err.Error())
		return r
	}

	r.Message = "Received data from pod " + p.Name
	r.Status = 0

	return r
}

func main() {
	lis, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatalf("Failed to listen on port 5555: %v", err)
	}
	s := Server{}

	grpcServer := grpc.NewServer()
	pb.RegisterDataEndpointServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 5555 %v", err)
	}

}

package endpoint

import (
	"log"

	"golang.org/x/net/context"
)

type Server struct {
	UnimplementedDataEndpointServer
}

func (s *Server) Write(ctx context.Context, data *Data) (*Response, error) {
	log.Printf("Received pod: %s", data)

	return &Response{Message: "Received data", Status: 0}, nil
}

func (s *Server) SendMetadata(ctx context.Context, podData *Metadata) (*Response, error) {

	return &Response{Message: "Received data", Status: 0}, nil
}

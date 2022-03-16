// Copyright The Snoopy Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	pb "github.com/fennec-project/snoopy-operator/endpoint/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedDataEndpointServer
}

func (s server) ExportPodData(srv pb.DataEndpoint_ExportPodDataServer) error {

	log.Println("start new server")

	ctx := srv.Context()

	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		pd, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		// Send response back to client
		resp := pb.Response{Message: "Received data for pod " + pd.Name}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		// log.Printf("Received data for pod %v", pd.Name)

		// open a file to write and append podData
		f, err := os.OpenFile("/pcap/"+pd.Name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			fmt.Print(err.Error())
			log.Fatal("Error openning file to write data on server.")
		}
		_, err = f.Write(pd.Data)
		if err != nil {
			fmt.Print(err.Error())
			log.Fatal("Error writing data to file on server.")
		}
		f.Close()
	}
}

func main() {

	// Get os args
	// address := os.Args[1]
	port := os.Args[1]

	// create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Listening on port %s", port)
	// create grpc server
	s := grpc.NewServer()
	pb.RegisterDataEndpointServer(s, &server{})

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

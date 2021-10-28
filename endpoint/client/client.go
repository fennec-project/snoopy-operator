package main

import (
	"context"
	"io"
	"log"

	pb "github.com/fennec-project/snoopy-operator/endpoint/proto"

	"time"

	"google.golang.org/grpc"
)

func main() {

	// dail server
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewDataEndpointClient(conn)
	stream, err := client.ExportPodData(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

	ctx := stream.Context()
	done := make(chan bool)

	// first goroutine sends random increasing numbers to stream
	// and closes it after 10 iterations
	go func() {
		for i := 1; i <= 10; i++ {
			// generates random number and sends it to stream
			mydata := []byte("That is my test string")
			pd := pb.PodData{Name: "podtest", Data: mydata}
			if err := stream.Send(&pd); err != nil {
				log.Fatalf("can not send %v", err)
			}
			log.Printf("%v sent data for pod ", pd.Name)
			time.Sleep(time.Millisecond * 200)
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives data from stream
	// and saves result in max variable
	//
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			message := resp.Message
			log.Printf("Received: %s", message)
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Print("finished with pod data")
}

package main

import (
	"context"
	"flag"
	"log"
	"main/pb"
	"main/sample"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	ctx := context.Background()
	serverAddr := flag.String("addr", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddr)

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect to server with address: %s. Error: (%v)", *serverAddr, err)
	}
	client := pb.NewLaptopServiceClient(conn)
	laptop := sample.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	response, err := client.CreateLaptop(ctx, req)
	if err != nil {
		state, ok := status.FromError(err)
		if ok && state.Code() == codes.AlreadyExists {
			log.Println("laptop already exists")
		} else {
			log.Fatalf("cannot save laptop: (%v)", err)
		}
	}

	log.Printf("laptop saved with id: %s", response.Id)
}

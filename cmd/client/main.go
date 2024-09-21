package main

import (
	"context"
	"flag"
	"io"
	"log"
	"main/pb"
	"main/sample"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	parentCtx := context.Background()
	serverAddr := flag.String("addr", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddr)

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect to server with address: %s. Error: (%v)", *serverAddr, err)
	}
	client := pb.NewLaptopServiceClient(conn)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			createLaptop(parentCtx, client)
		}()
	}
	wg.Wait()
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	searchLaptop(parentCtx, client, filter)
}

func searchLaptop(parCtx context.Context, client pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("seacrh filter: %v\n", filter)
	ctx, cancel := context.WithTimeout(parCtx, 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := client.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatalf("cannot seacrh laptop: (%v)", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("cannot recieve response: (%v)", err)
		}

		laptop := response.GetLaptop()
		log.Printf("- found: %v\n", laptop.GetId())
		log.Printf("\t+ brand: %v\n", laptop.GetBrand())
		log.Printf("\t+ name: %v\n", laptop.GetName())
		log.Printf("\t+ cpu cores: %v\n", laptop.GetCpu().GetCores())
		log.Printf("\t+ ram: %v (%v)\n", laptop.GetRAM().GetValue(), laptop.GetRAM().GetUnit())
		log.Printf("\t+ price: %v USD\n", laptop.GetPriceUsd())
	}
}

func createLaptop(parCtx context.Context, client pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	ctx, cancel := context.WithTimeout(parCtx, time.Second*5)
	defer cancel()
	response, err := client.CreateLaptop(ctx, req)
	if err != nil {
		state, ok := status.FromError(err)
		if ok && state.Code() == codes.AlreadyExists {
			log.Println("laptop already exists")
		} else {
			log.Fatalf("cannot save laptop: (%v)", err)
		}
	}

	log.Printf("laptop saved with id: %s\n", response.Id)
}

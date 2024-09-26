package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/client"
	"main/pb"
	"main/sample"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	username        = "admin"
	password        = "passwordQWERTY1"
	refreshDuration = time.Second * 30
)

func main() {
	parentCtx := context.Background()
	serverAddr := flag.String("addr", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddr)

	authConn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect to server with address: %s. Error: (%v)", *serverAddr, err)
	}
	authClient := client.NewAuthClient(authConn, username, password)
	interceptor, err := client.NewAuthIntercepter(parentCtx, authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := grpc.NewClient(
		*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal(err)
	}
	laptopClient := client.NewLaptopClient(conn)
	testRateLaptop(parentCtx, laptopClient)
}

func authMethods() map[string]bool {
	const laptopServicePath = "/pc.LaptopService/"
	return map[string]bool{
		fmt.Sprintf("%v%v", laptopServicePath, "CreateLaptop"): true,
		fmt.Sprintf("%v%v", laptopServicePath, "UploadImage"):  true,
		fmt.Sprintf("%v%v", laptopServicePath, "RateLaptop"):   true,
	}
}

func testCreateNLaptopsAndSearchOneOf(ctx context.Context, client *client.LaptopClient) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			client.CreateLaptop(ctx, sample.NewLaptop())
		}()
	}
	wg.Wait()
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	client.SearchLaptop(ctx, filter)
}

func testUploadImage(ctx context.Context, client *client.LaptopClient) {
	laptop := sample.NewLaptop()
	client.CreateLaptop(ctx, laptop)
	client.UploadImage(ctx, laptop.GetId(), "tmp/laptop.jpg")
}

func testRateLaptop(ctx context.Context,
	client *client.LaptopClient,
) {
	n := 3
	laptopIds := make([]string, n)
	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIds[i] = laptop.GetId()
		client.CreateLaptop(ctx, laptop)
	}
	scores := make([]float64, n)
	for i := 0; i < n; i++ {
		scores[i] = sample.RadnomLaptopScore()
	}
	err := client.RateLaptop(ctx, laptopIds, scores)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

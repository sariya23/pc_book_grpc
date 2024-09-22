package main

import (
	"flag"
	"fmt"
	"log"
	"main/pb"
	"main/service"
	"main/storage"
	"net"

	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Println("starting grpc server")
	laptopStorage := storage.NewInMemoryLaptopStorage()
	imageStorage := storage.NewImageStorage("img")
	server := service.NewLaptopServer(laptopStorage, imageStorage)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, server)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot start server: (%v)", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("cannot start server")
	}
}

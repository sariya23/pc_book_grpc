package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/pb"
	"main/service"
	"main/storage"
	"net"

	"google.golang.org/grpc"
)

func unaryIntercepter(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	log.Println("--> unary intercepter: ", info.FullMethod)
	return handler(ctx, req)
}

func streamIntercepter(srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	log.Println("--> stream intercepter: ", info.FullMethod)
	return handler(srv, ss)
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Println("starting grpc server")
	laptopStorage := storage.NewInMemoryLaptopStorage()
	imageStorage := storage.NewImageStorage("img")
	ratingStorage := storage.NewRatingStorage()
	server := service.NewLaptopServer(laptopStorage, imageStorage, ratingStorage)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryIntercepter),
		grpc.StreamInterceptor(streamIntercepter),
	)
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

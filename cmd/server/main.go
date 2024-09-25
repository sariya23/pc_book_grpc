package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"main/models"
	"main/pb"
	"main/service"
	"main/storage"
	"net"
	"time"

	"google.golang.org/grpc"
)

// TODO: secret in .env
const (
	jwtKey = "secret"
	TTL    = 15 * time.Minute
)

func main() {
	const op = "cmd.server.main"
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("%v: starting grpc server\n", op)
	laptopStorage := storage.NewInMemoryLaptopStorage()
	imageStorage := storage.NewImageStorage("img")
	ratingStorage := storage.NewRatingStorage()
	userStorage := storage.NewUserStorage()
	err := seedUsers(userStorage)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v: users created\n", op)
	jwtManager := service.NewJWTManager(jwtKey, TTL)
	authServer := service.NewAuthServer(userStorage, jwtManager)
	server := service.NewLaptopServer(laptopStorage, imageStorage, ratingStorage)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryIntercepter),
		grpc.StreamInterceptor(streamIntercepter),
	)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, server)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("%v: cannot start server: (%v)", op, err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("%v: cannot start server", op)
	}
}

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

func seedUsers(userStorage service.UserStorager) error {
	err := createUser(userStorage, "admin", "passwordQWERTY1", "admin")
	if err != nil {
		return err
	}
	return createUser(userStorage, "user1", "pass2", "user")
}

func createUser(userStorage service.UserStorager, username, password, role string) error {
	user, err := models.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStorage.Save(user)
}

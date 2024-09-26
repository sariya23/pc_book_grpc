package main

import (
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

	interceptor := service.NEwAuthInterceptor(jwtManager, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
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

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/pc.LaptopService/"
	return map[string][]string{
		fmt.Sprintf("%v%v", laptopServicePath, "CreateLaptop"): {"admin"},
		fmt.Sprintf("%v%v", laptopServicePath, "UploadImage"):  {"admin"},
		fmt.Sprintf("%v%v", laptopServicePath, "RateLaptop"):   {"admin", "user"},
	}
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

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"main/models"
	"main/pb"
	"main/service"
	"main/storage"
	"net"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TODO: secret in .env
const (
	jwtKey = "secret"
	TTL    = 15 * time.Minute
)

func main() {
	const op = "cmd.server.main"
	port := flag.Int("port", 0, "the server port")
	enableTLS := flag.Bool("tsl", false, "enable SSL/TLS")
	flag.Parse()
	log.Printf("%v: starting grpc server, TLS: %v\n", op, *enableTLS)
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
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if *enableTLS {
		tlsCreds, err := loadTLSCreds()
		if err != nil {
			log.Fatalf("cannot load TLS creds: %v", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(tlsCreds))
	}

	grpcServer := grpc.NewServer(serverOpts...)
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

func loadTLSCreds() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(filepath.Join("cert", "server-cert.pem"), filepath.Join("cert", "server-key.pem"))
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}
	return credentials.NewTLS(config), nil
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

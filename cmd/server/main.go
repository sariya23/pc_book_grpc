package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"main/models"
	"main/pb"
	"main/service"
	"main/storage"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: secret in .env
const (
	jwtKey = "secret"
	TTL    = 15 * time.Minute
)

var (
	serverCertFile       = filepath.Join("cert", "server-cert.pem")
	serverPriviteKeyFile = filepath.Join("cert", "server-key.pem")
)

func main() {
	ctx := context.Background()
	const op = "cmd.server.main"

	port := flag.Int("port", 0, "the server port")
	enableTLS := flag.Bool("tsl", false, "enable SSL/TLS")
	serverType := flag.String("type", "grpc", "type os server: grpc/rest")
	grpcEndpoint := flag.String("endpoint", "", "gRPC endpoint")

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
	laptopServer := service.NewLaptopServer(laptopStorage, imageStorage, ratingStorage)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("%v: cannot start server: (%v)", op, err)
	}
	if *serverType == "grpc" {
		err = runGRPCServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
	} else {
		err = runRESTServer(ctx, authServer, laptopServer, *enableTLS, listener, *grpcEndpoint)
	}
	if err != nil {
		log.Fatalf("op: %v, error: %v", op, err)
	}
}

func runGRPCServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	const op = "cmd.server.runGRPCServer"
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCreds, err := loadTLSCreds()
		if err != nil {
			return fmt.Errorf("%v: cannot load TLS creds: %w", op, err)
		}
		serverOpts = append(serverOpts, grpc.Creds(tlsCreds))
	}

	grpcServer := grpc.NewServer(serverOpts...)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	log.Printf("%v: start GRPC server at %v, TLS: %t\n", op, listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}

func runRESTServer(
	ctx context.Context,
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	enableTLS bool,
	listener net.Listener,
	grpcEndpoint string,
) error {
	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return nil
	}

	err = pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return nil
	}
	log.Printf("start REST server at %v, TLS: %t\n", listener.Addr().String(), enableTLS)
	if enableTLS {
		return http.ServeTLS(listener, mux, serverCertFile, serverPriviteKeyFile)
	}
	return http.Serve(listener, mux)
}

func loadTLSCreds() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverPriviteKeyFile)
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

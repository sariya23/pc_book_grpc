package service

import (
	"context"
	"log"
	"main/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userStorage UserStorager
	jwtManager  *JWTManager
	pb.UnimplementedAuthServiceServer
}

func NewAuthServer(userStorage UserStorager, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{
		userStorage: userStorage,
		jwtManager:  jwtManager,
	}
}

func (au *AuthServer) Login(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	user, err := au.userStorage.Get(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}
	if user == nil || !user.IsPasswordCorrect(req.GetPassword()) {
		log.Printf("invalid creds for %v \n", user)
		return nil, status.Errorf(codes.NotFound, "invalid creds")
	}

	token, err := au.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot generate jwt")
	}

	resp := &pb.LogResponse{
		Token: token,
	}
	return resp, nil
}

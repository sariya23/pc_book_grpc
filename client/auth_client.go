package client

import (
	"context"
	"main/pb"
	"time"

	"google.golang.org/grpc"
)

type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

func (ac *AuthClient) Login(ctx context.Context) (token string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.LogRequest{
		Username: ac.username,
		Password: ac.password,
	}
	resp, err := ac.service.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.GetToken(), nil
}

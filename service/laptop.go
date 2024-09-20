package service

import (
	"context"
	"errors"
	"main/pb"
	"main/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopStorage interface {
	Save(laptop *pb.Laptop) error
	Get(id string) (*pb.Laptop, error)
}

type LaptopServer struct {
	Store LaptopStorage
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(store LaptopStorage) *LaptopServer {
	return &LaptopServer{Store: store}
}

func (s *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "laprot id is not correct")
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Error(codes.Internal, "cannot generate id")
		}
		laptop.Id = id.String()
	}

	err := s.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, storage.ErrAlreadyExist) {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, "cennot save laptop")
	}
	return &pb.CreateLaptopResponse{Id: laptop.Id}, nil
}

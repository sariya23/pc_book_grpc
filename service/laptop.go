package service

import (
	"context"
	"errors"
	"log"
	"main/pb"
	"main/storage"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopStorage interface {
	Save(laptop *pb.Laptop) error
	Get(id string) (*pb.Laptop, error)
	Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type LaptopServer struct {
	Storage LaptopStorage
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(storage LaptopStorage) *LaptopServer {
	return &LaptopServer{Storage: storage}
}

func (s *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a laptop with id: %s", laptop.Id)
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
	time.Sleep(6 * time.Second)

	if ctx.Err() == context.Canceled {
		log.Println("context cancelled")
		return nil, status.Error(codes.Canceled, "cancelled context")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("deadline exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	err := s.Storage.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, storage.ErrAlreadyExist) {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, "cennot save laptop")
	}
	return &pb.CreateLaptopResponse{Id: laptop.Id}, nil
}

func (s *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream grpc.ServerStreamingServer[pb.SearchLaptopResponse],
) error {
	filter := req.GetFilter()
	log.Printf("recieve a seacrh laptop request with filter %v\n", filter)
	err := s.Storage.Search(
		filter,
		func(laptop *pb.Laptop) error {
			response := &pb.SearchLaptopResponse{
				Laptop: laptop,
			}
			err := stream.Send(response)
			if err != nil {
				return nil
			}
			log.Printf("send laptop with id %v", laptop.GetId())
			return nil
		})
	if err != nil {
		return status.Error(codes.Internal, "internal error")
	}
	return nil
}

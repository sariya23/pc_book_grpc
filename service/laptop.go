package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"main/pb"
	"main/storage"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxImageSize = 1 << 20
)

type LaptopStorager interface {
	Save(laptop *pb.Laptop) error
	Get(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type ImageStorager interface {
	Save(
		laptopID string,
		imageType string,
		imageData bytes.Buffer,
	) (string, error)
}

type RatingStorager interface {
	Add(laptopId string, score float64) (*storage.Rating, error)
}

type LaptopServer struct {
	LaptopStorage LaptopStorager
	ImageStorage  ImageStorager
	RatingStorage RatingStorager
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(laptopStorage LaptopStorager, imageStorage ImageStorager, ratingStorage RatingStorager) *LaptopServer {
	return &LaptopServer{LaptopStorage: laptopStorage, ImageStorage: imageStorage, RatingStorage: ratingStorage}
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
	time.Sleep(time.Second)

	if ctx.Err() == context.Canceled {
		log.Println("context cancelled")
		return nil, status.Error(codes.Canceled, "cancelled context")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("deadline exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	err := s.LaptopStorage.Save(laptop)
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
	err := s.LaptopStorage.Search(
		stream.Context(),
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

func (s *LaptopServer) UploadImage(
	stream grpc.ClientStreamingServer[pb.UploadImageRequest, pb.UploadImageResponse]) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannon receive image info: %v", err)
	}
	laptopId := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive laptop with id: %v image type: %v", laptopId, imageType)

	laptop, err := s.LaptopStorage.Get(laptopId)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot get laptop with id %v: %v", laptopId, err)
	}

	if laptop == nil {
		return status.Errorf(codes.Internal, "laptop with id: %v does not exists", laptopId)
	}

	imageData := bytes.Buffer{}
	var imageSize int

	for {
		if stream.Context().Err() == context.Canceled {
			log.Println("context cancelled")
			return status.Error(codes.Canceled, "cancelled context")
		}

		if stream.Context().Err() == context.DeadlineExceeded {
			log.Println("deadline exceeded")
			return status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		log.Println("waiting image data")
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("no more data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "cannot receive data: %v", err)
		}
		chunk := req.GetChunkData()
		size := len(chunk)
		imageSize += size
		if imageSize > maxImageSize {
			return status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize)
		}
		_, err = imageData.Write(chunk)
		if err != nil {
			status.Errorf(codes.Internal, "cannot write data: %v", err)
		}
	}
	imageId, err := s.ImageStorage.Save(laptopId, imageType, imageData)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save image: %v", err)
	}

	resp := &pb.UploadImageResponse{
		Id:       imageId,
		ByteSize: uint32(imageSize),
	}
	err = stream.SendAndClose(resp)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}
	log.Println("image save successfully")
	return nil
}

func (s *LaptopServer) RateLaptop(stream grpc.BidiStreamingServer[pb.RateLaptopRequest, pb.RateLaptopResponse]) error {
	for {
		if stream.Context().Err() == context.Canceled {
			log.Println("context cancelled")
			return status.Error(codes.Canceled, "cancelled context")
		}

		if stream.Context().Err() == context.DeadlineExceeded {
			log.Println("deadline exceeded")
			return status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("no data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}
		laptopId := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("reveive a rate for laptop with id: %s, score: %.2f", laptopId, score)

		found, err := s.LaptopStorage.Get(laptopId)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot find laptop with id: %v: (%v)", laptopId, err)
		}
		if found == nil {
			return status.Errorf(codes.NotFound, "laptop id %v is not found", laptopId)
		}

		rating, err := s.RatingStorage.Add(laptopId, score)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot rate the laptop with id: %v: (%v)", laptopId, err)
		}

		resp := &pb.RateLaptopResponse{
			LaptopId:     laptopId,
			RatedCount:   rating.Count,
			AvarageScore: rating.Sum / float64(rating.Count),
		}
		err = stream.Send(resp)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot send response: %v", err)
		}
	}
	return nil
}

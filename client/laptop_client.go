package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"main/pb"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(cc)
	return &LaptopClient{service}
}

func (client *LaptopClient) CreateLaptop(parCtx context.Context, laptop *pb.Laptop) {
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	ctx, cancel := context.WithTimeout(parCtx, time.Second*5)
	defer cancel()
	response, err := client.service.CreateLaptop(ctx, req)
	if err != nil {
		state, ok := status.FromError(err)
		if ok && state.Code() == codes.AlreadyExists {
			log.Println("laptop already exists")
		} else {
			log.Fatalf("cannot save laptop: (%v)", err)
		}
	}

	log.Printf("laptop saved with id: %s\n", response.Id)
}

func (client *LaptopClient) SearchLaptop(parCtx context.Context, filter *pb.Filter) {
	log.Printf("seacrh filter: %v\n", filter)
	ctx, cancel := context.WithTimeout(parCtx, 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := client.service.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatalf("cannot seacrh laptop: (%v)", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("cannot recieve response: (%v)", err)
		}

		laptop := response.GetLaptop()
		log.Printf("- found: %v\n", laptop.GetId())
		log.Printf("\t+ brand: %v\n", laptop.GetBrand())
		log.Printf("\t+ name: %v\n", laptop.GetName())
		log.Printf("\t+ cpu cores: %v\n", laptop.GetCpu().GetCores())
		log.Printf("\t+ ram: %v (%v)\n", laptop.GetRAM().GetValue(), laptop.GetRAM().GetUnit())
		log.Printf("\t+ price: %v USD\n", laptop.GetPriceUsd())
	}
}

func (client *LaptopClient) RateLaptop(ctx context.Context, laptopIds []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("no more data")
				waitResponse <- nil
				break
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %w", err)
			}

			log.Printf("recieved data: %v\n", resp)

		}
	}()
	for i, laptopId := range laptopIds {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopId,
			Score:    scores[i],
		}
		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request %w: %w", err, stream.RecvMsg(nil))
		}

		log.Printf("sent request: %v\n", req)
	}
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %w", err)
	}
	err = <-waitResponse
	return err
}

func (client *LaptopClient) UploadImage(patCtx context.Context, laptopId string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("cannot open file: %v", err)
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(patCtx, time.Second*5)
	defer cancel()

	stream, err := client.service.UploadImage(ctx)
	if err != nil {
		log.Fatalf("cannot upload image: %v", err)
	}
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopId,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatalf("cannot send request: %v", err)
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot read chunk to buffer: %v", err)
		}
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatalf("cannot send chunk to server: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("cannot receive response: %v", err)
	}
	log.Printf("image uploaded with id: %s, size: %d", resp.GetId(), resp.GetByteSize())
}

package service_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"main/pb"
	"main/sample"
	"main/serializer"
	"main/service"
	"main/storage"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	laptopStorage := storage.NewInMemoryLaptopStorage()
	addr := startTestLaptopServer(t, laptopStorage, nil, nil)
	client := newTestLaptopClient(t, addr)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}
	res, err := client.CreateLaptop(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	other, err := laptopStorage.Get(laptop.Id)
	require.NoError(t, err)
	require.NotNil(t, other)
	requireSameLaptops(t, other, laptop)
}

func TestClientSearchLaptop(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	storage := storage.NewInMemoryLaptopStorage()
	excpectedIDs := make(map[string]bool, 2)
	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.Cores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.RAM = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.Cores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhx = 4.5
			laptop.RAM = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			excpectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.Cores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhx = 4.5
			laptop.RAM = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			excpectedIDs[laptop.Id] = true
		}
		err := storage.Save(laptop)
		require.NoError(t, err)
	}
	addr := startTestLaptopServer(t, storage, nil, nil)
	client := newTestLaptopClient(t, addr)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.SearchLaptop(ctx, req)
	require.NoError(t, err)
	var found int
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, excpectedIDs, response.GetLaptop().GetId())
		found++
	}
	require.Equal(t, len(excpectedIDs), found)
}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testImageFolder := "../tmp"
	laptopStorage := storage.NewInMemoryLaptopStorage()
	imageStorage := storage.NewImageStorage(testImageFolder)

	laptop := sample.NewLaptop()
	err := laptopStorage.Save(laptop)
	require.NoError(t, err)

	serverAddr := startTestLaptopServer(t, laptopStorage, imageStorage, nil)
	client := newTestLaptopClient(t, serverAddr)
	imagePath := fmt.Sprintf("%s/laptop.jpg", testImageFolder)
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()

	stream, err := client.UploadImage(ctx)
	require.NoError(t, err)

	imageType := filepath.Ext(imagePath)

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	var size int
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		size += n

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		require.NoError(t, err)
	}
	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, resp.GetId())
	require.EqualValues(t, size, resp.GetByteSize())
	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, resp.GetId(), imageType)
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))
}

func TestClientRateLaptop(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	laptopStorage := storage.NewInMemoryLaptopStorage()
	ratingStorage := storage.NewRatingStorage()
	laptop := sample.NewLaptop()
	err := laptopStorage.Save(laptop)
	require.NoError(t, err)
	serverAddr := startTestLaptopServer(t, laptopStorage, nil, ratingStorage)
	client := newTestLaptopClient(t, serverAddr)

	stream, err := client.RateLaptop(ctx)
	require.NoError(t, err)
	scores := []float64{8, 7.5, 10}
	avg := []float64{8, 7.75, 8.5}
	n := len(scores)
	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}

		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for i := 0; ; i++ {
		resp, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, i)
			return
		}
		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), resp.LaptopId)
		require.Equal(t, uint32(i+1), resp.GetRatedCount())
		require.Equal(t, avg[i], resp.GetAvarageScore())
	}
}

func startTestLaptopServer(t *testing.T,
	laptopStorage service.LaptopStorager,
	imageStorage service.ImageStorager,
	ratingStorage service.RatingStorager,
) string {
	server := service.NewLaptopServer(laptopStorage, imageStorage, ratingStorage)
	grpsServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpsServer, server)

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go grpsServer.Serve(l)

	return l.Addr().String()

}

func newTestLaptopClient(t *testing.T, serverAddr string) pb.LaptopServiceClient {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptops(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.Marshaler(laptop1)
	require.NoError(t, err)

	json2, err := serializer.Marshaler(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}

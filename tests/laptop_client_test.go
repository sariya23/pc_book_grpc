package service_test

import (
	"context"
	"io"
	"main/pb"
	"main/sample"
	"main/serializer"
	"main/service"
	"main/storage"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	laptopStorage := storage.NewInMemoryLaptopStorage()
	server, addr := startTestLaptopServer(t, laptopStorage, nil)
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

	other, err := server.LaptopStorage.Get(laptop.Id)
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
	_, addr := startTestLaptopServer(t, storage, nil)
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

func startTestLaptopServer(t *testing.T, laptopStorage service.LaptopStorager, imageStorage service.ImageStorager) (*service.LaptopServer, string) {
	server := service.NewLaptopServer(laptopStorage, imageStorage)
	grpsServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpsServer, server)

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go grpsServer.Serve(l)

	return server, l.Addr().String()

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

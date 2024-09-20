package service_test

import (
	"context"
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
	server, addr := startTestLaptopServer(t)
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

	other, err := server.Store.Get(laptop.Id)
	require.NoError(t, err)
	require.NotNil(t, other)
	requireSameLaptops(t, other, laptop)
}

func startTestLaptopServer(t *testing.T) (*service.LaptopServer, string) {
	server := service.NewLaptopServer(storage.NewInMemoryLaptopStorage())
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

package service_test

import (
	"context"
	"main/pb"
	"main/sample"
	"main/service"
	"main/storage"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()
	laptopWithNoId := sample.NewLaptop()
	laptopWithNoId.Id = ""

	laptopWithInvalidId := sample.NewLaptop()
	laptopWithInvalidId.Id = "invelid uuid"

	laptopForTestDuplicateId := sample.NewLaptop()
	storageWithDuplicateLaptop := storage.NewInMemoryLaptopStorage()
	err := storageWithDuplicateLaptop.Save(laptopForTestDuplicateId)
	require.NoError(t, err)

	testCasess := []struct {
		name         string
		laptop       *pb.Laptop
		storage      service.LaptopStorager
		expectedCode codes.Code
	}{
		{
			name:         "success with id",
			laptop:       sample.NewLaptop(),
			storage:      storage.NewInMemoryLaptopStorage(),
			expectedCode: codes.OK,
		},
		{
			name:         "success no id",
			laptop:       laptopWithNoId,
			storage:      storage.NewInMemoryLaptopStorage(),
			expectedCode: codes.OK,
		},
		{
			name:         "invalid uuid",
			laptop:       laptopWithInvalidId,
			storage:      storage.NewInMemoryLaptopStorage(),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "duplicate id",
			laptop:       laptopForTestDuplicateId,
			storage:      storageWithDuplicateLaptop,
			expectedCode: codes.AlreadyExists,
		},
	}

	for _, tc := range testCasess {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}
			ctx := context.Background()
			server := service.NewLaptopServer(tc.storage, nil, nil)
			res, err := server.CreateLaptop(ctx, req)
			if tc.expectedCode == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res)
				require.Equal(t, tc.laptop.Id, res.Id)
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				state, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, state.Code())
			}
		})
	}
}

package serializer_test

import (
	"main/pb"
	"main/sample"
	"main/serializer"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCanConvertProtoToBinAndBinToProto(t *testing.T) {
	t.Parallel()

	binFile := "../tmp/laptop.bin"
	laptop1 := sample.NewLaptop()
	err := serializer.ProtobufToBinFile(binFile, laptop1)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = serializer.BinFileToProtobuf(binFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))
}

func TestCanConvertProtoToJSONAndJSONToProto(t *testing.T) {
	laptop1 := sample.NewLaptop()
	jsonFile := "../tmp/laptop.json"

	err := serializer.PotobobufToJSON(jsonFile, laptop1)
	require.NoError(t, err)

}

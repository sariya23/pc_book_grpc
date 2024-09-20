package serializer

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
)

func ProtobufToBinFile(filename string, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("cannnot marshal proto: %w", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write data in file: %w", err)
	}
	return nil
}

func BinFileToProtobuf(filename string, message proto.Message) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot write file: %w", err)
	}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	return nil
}

func PotobobufToJSON(filename string, message proto.Message) error {
	data, err := Marshaler(message)
	if err != nil {
		return fmt.Errorf("cannot marhsal data: %w", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write data: %w", err)
	}
	return nil
}

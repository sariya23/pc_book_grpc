package sample

import (
	"main/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewKeyboard() *pb.Keyboard {
	keyboard := &pb.Keyboard{
		Layout:   randomKeyboardLayout(),
		Backlist: randomBool(),
	}
	return keyboard
}

func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)
	cores := randomInt(2, 8)
	threads := randomInt(cores, 12)
	minGhx := randomFloat64(2.0, 3.5)
	maxGhx := randomFloat64(minGhx, 5.0)
	cpu := &pb.CPU{
		Brand:   brand,
		Name:    name,
		Threads: uint32(threads),
		MinGhz:  minGhx,
		MaxGhx:  maxGhx,
	}
	return cpu
}

func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minGhx := randomFloat64(2.0, 3.5)
	maxGhx := randomFloat64(minGhx, 5.0)
	memory := &pb.Memory{
		Value: uint64(randomInt(2, 32)),
		Unit:  pb.Memory_GIGABYTE,
	}
	gpu := &pb.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: maxGhx,
		MaxGhx: maxGhx,
		Memory: memory,
	}
	return gpu
}

func NewRAM() *pb.Memory {
	ram := &pb.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit:  pb.Memory_GIGABYTE,
	}
	return ram
}

func NewHDD() *pb.Storage {
	hdd := &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pb.Memory_TERABYTE,
		},
	}
	return hdd
}

func NewSSD() *pb.Storage {
	ssd := &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(512, 1024)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
	return ssd
}

func NewScreen() *pb.Screen {
	screen := &pb.Screen{
		Inch:       randomFloat64(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}
	return screen
}

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)
	laptop := &pb.Laptop{
		Id:          randomUUID(),
		Brand:       brand,
		Name:        name,
		Cpu:         NewCPU(),
		Gpus:        []*pb.GPU{NewGPU()},
		RAM:         NewRAM(),
		Storages:    []*pb.Storage{NewHDD(), NewHDD()},
		Screen:      NewScreen(),
		Keyboard:    NewKeyboard(),
		Weight:      &pb.Laptop_WeightKg{WeightKg: randomFloat64(1, 3)},
		PriceUsd:    randomFloat64(1000, 3000),
		ReleaseYear: uint32(randomInt(2015, 2024)),
		UpdatedAt:   timestamppb.Now(),
	}
	return laptop
}

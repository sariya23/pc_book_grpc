package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/pb"
	"sync"

	"github.com/jinzhu/copier"
)

var (
	ErrAlreadyExist = errors.New("laptop with this id already exist")
)

type InMemoryLaptopStore struct {
	mu   sync.RWMutex
	data map[string]*pb.Laptop
}

func NewInMemoryLaptopStorage() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (m *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[laptop.Id]; ok {
		return ErrAlreadyExist
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("cannot copy laptop data: %w", err)
	}

	m.data[other.Id] = other
	return nil
}

func (m *InMemoryLaptopStore) Get(id string) (*pb.Laptop, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	laptop, ok := m.data[id]
	if !ok {
		return nil, nil
	}
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy object: %w", err)
	}
	return other, nil
}

func (m *InMemoryLaptopStore) Search(
	ctx context.Context,
	filter *pb.Filter,
	found func(laptop *pb.Laptop) error,
) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, laptop := range m.data {
		if err := ctx.Err(); err == context.Canceled || err == context.DeadlineExceeded {
			log.Println("context is cancelled")
			return errors.New("context is cancelled")
		}
		if isQualified(filter, laptop) {
			other := &pb.Laptop{}
			err := copier.Copy(other, laptop)
			if err != nil {
				return fmt.Errorf("cannot copy object: %w", err)
			}
			err = found(other)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}
	if laptop.GetCpu().GetCores() < filter.GetMinCpuCores() {
		return false
	}
	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}
	if toBit(laptop.GetRAM()) < toBit(filter.GetMinRam()) {
		return false
	}
	return true
}

func toBit(memory *pb.Memory) uint64 {
	val := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return val
	case pb.Memory_KILOBYTE:
		return val << 13
	case pb.Memory_MEGABYTE:
		return val << 23
	case pb.Memory_GIGABYTE:
		return val << 33
	case pb.Memory_TERABYTE:
		return val << 43
	default:
		return 0
	}
}

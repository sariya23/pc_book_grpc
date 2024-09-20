package storage

import (
	"errors"
	"fmt"
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

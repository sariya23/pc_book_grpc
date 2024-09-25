package storage

import (
	"main/models"
	"sync"
)

type UserStorage struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		users: make(map[string]*models.User),
	}
}

func (s *UserStorage) Save(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[user.UserName]; ok {
		return ErrAlreadyExist
	}

	s.users[user.UserName] = user.Clone()
	return nil
}

func (s *UserStorage) Get(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return nil, nil
	}
	return user.Clone(), nil
}

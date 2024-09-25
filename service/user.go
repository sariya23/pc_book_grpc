package service

import (
	"fmt"
	"main/models"

	"golang.org/x/crypto/bcrypt"
)

type UserStorager interface {
	Save(user *models.User) error
	Get(username string) (*models.User, error)
}

func NewUser(username string, password string, role string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}
	return &models.User{
		UserName:       username,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}, nil
}

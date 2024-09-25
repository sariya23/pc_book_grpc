package models

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserName       string
	HashedPassword string
	Role           string
}

func (u *User) IsPasswordCorrect(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	log.Println("models.user.IsPasswordCorrect", err)
	return err == nil
}

func (u *User) Clone() *User {
	return &User{
		UserName:       u.UserName,
		HashedPassword: u.HashedPassword,
		Role:           u.Role,
	}
}

func NewUser(username string, password string, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}
	return &User{
		UserName:       username,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}, nil
}

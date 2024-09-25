package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	UserName       string
	HashedPassword string
	Role           string
}

func (u *User) IsPasswordCorrect(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(u.HashedPassword))
	return err == nil
}

func (u *User) Clone() *User {
	return &User{
		UserName:       u.UserName,
		HashedPassword: u.HashedPassword,
		Role:           u.Role,
	}
}

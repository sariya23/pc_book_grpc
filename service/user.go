package service

import (
	"main/models"
)

type UserStorager interface {
	Save(user *models.User) error
	Get(username string) (*models.User, error)
}

package service

import (
	"fmt"
	"main/models"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTManager struct {
	secretKey string
	TTL       time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewJWTManager(secretKey string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		TTL:       ttl,
	}
}

func (j *JWTManager) Generate(user *models.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.TTL).Unix(),
		},
		Username: user.UserName,
		Role:     user.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (any, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(j.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

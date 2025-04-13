package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var SecretKey = []byte("your_secret_key") // тот же ключ, что и в VerifyToken

// GenerateToken генерирует JWT токен по email
func GenerateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// SecretKey используется для подписи токенов
var SecretKey = []byte("your_secret_key") // Замените на ваш секретный ключ
// ErrInvalidToken ошибка для недействительного токена
var ErrInvalidToken = errors.New("invalid token")

// GenerateToken генерирует JWT токен для пользователя
func GenerateToken(email string) (string, error) {
	// Создание нового токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Токен истекает через 24 часа
	})

	// Подпись токена
	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken проверяет действительность токена
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	// Парсинг токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Извлечение данных из токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

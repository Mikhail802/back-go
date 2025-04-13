package utils

import (
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)


var ErrInvalidToken = errors.New("invalid token")

// VerifyToken проверяет JWT и возвращает claims, если всё ок
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	log.Printf("🧪 Старт проверки токена: %s", tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Println("⛔ Неподдерживаемый метод подписи")
			return nil, ErrInvalidToken
		}
		return SecretKey, nil
	})

	if err != nil {
		log.Printf("❌ Ошибка парсинга токена: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("⛔ Токен невалиден или не содержит MapClaims")
		return nil, ErrInvalidToken
	}

	// Выводим все claims
	log.Printf("📥 Получены claims: %+v", claims)

	// Проверка срока действия
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			log.Printf("⏰ Токен истёк: %v (exp=%v)", expTime, exp)
			return nil, errors.New("token expired")
		}
	} else {
		log.Println("⚠️ В токене отсутствует поле exp")
		return nil, errors.New("token has no exp")
	}

	log.Println("✅ Токен прошёл проверку")
	return claims, nil
}

package middleware

import (
	"go_back/internal/utils"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware проверяет токен
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("AuthMiddleware: Missing token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Missing token"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		log.Println("AuthMiddleware: Missing token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Missing token"})
	}

	claims, err := utils.VerifyToken(token)
	if err != nil {
		log.Printf("AuthMiddleware: Invalid token: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid token"})
	}

	c.Locals("user", claims["email"])
	return c.Next()
}

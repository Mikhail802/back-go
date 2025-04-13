package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go_back/internal/initializers"
	"go_back/internal/models"
	"go_back/internal/utils"
)

// AuthMiddleware проверяет токен авторизации
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		log.Println("❌ AuthMiddleware: Missing Authorization header")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "Missing Authorization header",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Printf("❌ AuthMiddleware: Malformed Authorization header: %s", authHeader)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "Malformed Authorization header",
		})
	}

	token := parts[1]
	if token == "" {
		log.Println("❌ AuthMiddleware: Token part is empty")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "Empty token",
		})
	}

	log.Printf("🔐 Проверка токена: %s", token)

	claims, err := utils.VerifyToken(token)
	if err != nil {
		log.Printf("❌ AuthMiddleware: Invalid token: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid token",
		})
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		log.Println("❌ AuthMiddleware: Token payload missing email")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid token payload (missing email)",
		})
	}

	log.Printf("📨 Email из токена: %s", email)

	var user models.User
	if err := initializers.DB.First(&user, "email = ?", email).Error; err != nil {
		log.Printf("❌ AuthMiddleware: User not found with email: %s", email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "fail",
			"message": "User not found",
		})
	}

	log.Printf("✅ AuthMiddleware: Authenticated user %s (ID: %v)", user.Email, user.ID)

	c.Locals("user", &user)
	return c.Next()
}

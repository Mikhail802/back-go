package controllers

import (
	"go_back/internal/initializers"
	"go_back/internal/models"
	"go_back/internal/utils"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// VerifyPassword проверяет правильность пароля
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func ResetPassword(c *fiber.Ctx) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат данных"})
	}

	email := strings.ToLower(body.Email)

	var user models.User
	if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при хешировании пароля"})
	}

	user.Password = string(hashedPassword)
	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Не удалось обновить пароль"})
	}

	return c.JSON(fiber.Map{"message": "Пароль успешно обновлён"})
}


// CreateUser  создает нового пользователя
func CreateUser(c *fiber.Ctx) error {
	var payload *models.CreateUserSchema

	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Create:User  BodyParser error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	newUser := models.User{
		ID:       uuid.New(),
		Name:     payload.Name,
		Username: payload.Username,
		Email:    payload.Email,
		Password: utils.GeneratePassword(payload.Password),
	}

	result := initializers.DB.Create(&newUser)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "(SQLSTATE 23505)") {
			log.Printf("Create:User  Email already exists: %s", payload.Email)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "Email already exists"})
		}
		log.Printf("Create:User  Database error: %v", result.Error)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
	}

	log.Printf("Create:User  User created successfully: %v", newUser)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": newUser}})
}

// LoginUser  авторизует пользователя
func LoginUser(c *fiber.Ctx) error {
	var payload *models.LoginUserSchema

	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Login:User  BodyParser error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	var user models.User
	result := initializers.DB.First(&user, "email = ?", payload.Email)
	if result.Error != nil {
		log.Printf("Login:User  Invalid credentials for email: %s", payload.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	if err := VerifyPassword(user.Password, payload.Password); err != nil {
		log.Printf("Login:User  Invalid password for email: %s", payload.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	// Генерация токена
	token, err := utils.GenerateToken(user.Email)
	if err != nil {
		log.Printf("Login:User  Could not generate token for email: %s, error: %v", payload.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not generate token"})
	}

	log.Printf("Login:User  User logged in successfully: %s", payload.Email)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user, "token": token}})
}

// DeleteUser   удаляет пользователя
func DeleteUser(c *fiber.Ctx) error {
	userId := c.Params("userId")

	result := initializers.DB.Delete(&models.User{}, "id = ?", userId)

	if result.RowsAffected == 0 {
		log.Printf("Delete:User  No user found with ID: %s", userId)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "No user with that Id exists"})
	} else if result.Error != nil {
		log.Printf("Delete:User  Database error: %v", result.Error)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error})
	}

	log.Printf("Delete:User  User deleted successfully with ID: %s", userId)
	return c.SendStatus(fiber.StatusNoContent)
}

// FindUsers возвращает список пользователей с пагинацией
func FindUsers(c *fiber.Ctx) error {
	var page = c.Query("page", "1")
	var limit = c.Query("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var users []models.User
	results := initializers.DB.Limit(intLimit).Offset(offset).Find(&users)
	if results.Error != nil {
		log.Printf("FindUsers: Database error: %v", results.Error)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": results.Error})
	}

	log.Printf("FindUsers: Retrieved %d users", len(users))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "results": len(users), "users": users})
}

// FindUser ById возвращает пользователя по ID
func FindUserById(c *fiber.Ctx) error {
	userId := c.Params("userId")

	var user models.User
	result := initializers.DB.First(&user, "id = ?", userId)
	if result.Error != nil {
		log.Printf("FindUser ById: User not found with ID: %s", userId)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "User   not found"})
	}

	log.Printf("FindUser ById: User found with ID: %s", userId)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": user})
}

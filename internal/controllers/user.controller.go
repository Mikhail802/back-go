package controllers

import (
	"go_back/internal/initializers"
	"go_back/internal/models"
	"go_back/internal/utils"
	"log"
	"strconv"
	"strings"
	"encoding/json"
	"net/http"
	"fmt"
	"os"


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

	// определяем: email или логин
	query := initializers.DB
	if strings.Contains(payload.Identifier, "@") {
		query = query.Where("email = ?", payload.Identifier)
	} else {
		query = query.Where("username = ?", payload.Identifier)
	}

	result := query.First(&user)
	if result.Error != nil {
		log.Printf("Login:User  Invalid credentials for: %s", payload.Identifier)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	if err := VerifyPassword(user.Password, payload.Password); err != nil {
		log.Printf("Login:User  Invalid password for: %s", payload.Identifier)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	token, err := utils.GenerateToken(user.Email)
	if err != nil {
		log.Printf("Login:User  Could not generate token for: %s, error: %v", payload.Identifier, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not generate token"})
	}

	log.Printf("Login:User  User logged in successfully: %s", payload.Identifier)
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

type GoogleAuthRequest struct {
	IDToken string `json:"idToken"`
}

func GoogleLogin(c *fiber.Ctx) error {
	var body GoogleAuthRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
	}

	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + body.IDToken)
	if err != nil || resp.StatusCode != 200 {
		return c.Status(401).JSON(fiber.Map{"error": "Невалидный токен Google"})
	}

	var googleData struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleData); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка обработки данных Google"})
	}

	// Найти или создать пользователя
	var user models.User
	initializers.DB.FirstOrCreate(&user, models.User{Email: googleData.Email, Name: googleData.Name})

	// Генерация токена
	token, _ := utils.GenerateToken(user.Email)

	return c.JSON(fiber.Map{"token": token, "user": user})
}
// UpdateUserName обновляет имя пользователя
func UpdateUserName(c *fiber.Ctx) error {
	type request struct {
		Name string `json:"name"`
	}

	var body request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат запроса",
		})
	}

	user := c.Locals("user").(*models.User)

	// Обновляем имя
	if err := initializers.DB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("name", body.Name).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось обновить имя",
		})
	}

	// Получаем обновлённого пользователя
	var updatedUser models.User
	if err := initializers.DB.First(&updatedUser, "id = ?", user.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось загрузить обновлённого пользователя",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"user":   updatedUser,
	})
}

// UpdateUserUsername обновляет логин пользователя
func UpdateUserUsername(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
	}

	var body request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат запроса",
		})
	}

	user := c.Locals("user").(*models.User)

	// Проверка на уникальность логина
	var count int64
	initializers.DB.Model(&models.User{}).
		Where("username = ?", body.Username).
		Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Логин уже занят",
		})
	}

	// Обновляем логин
	if err := initializers.DB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("username", body.Username).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось обновить логин",
		})
	}

	// Получаем обновлённого пользователя
	var updatedUser models.User
	if err := initializers.DB.First(&updatedUser, "id = ?", user.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось загрузить обновлённого пользователя",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"user":   updatedUser,
	})
}

// UpdateUserAvatar обновляет аватар пользователя
func UpdateUserAvatar(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	userID := user.ID.String()

	contentType := c.Get("Content-Type")

	// 🎨 Обработка JSON-запроса с кастомной аватаркой (цвет + буква)
	if strings.HasPrefix(contentType, "application/json") {
		var body struct {
			Picture string `json:"Picture"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Неверный JSON-формат",
			})
		}

		if err := initializers.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Update("picture", body.Picture).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Не удалось сохранить Picture",
			})
		}

		return c.JSON(fiber.Map{
			"status":  "ok",
			"avatar":  body.Picture,
			"message": "JSON-аватар успешно сохранён",
		})
	}

	// 🖼️ Обработка multipart/form-data (фото)
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Файл аватара не найден",
		})
	}

	// Создание директории при необходимости
	saveDir := "/uploads/avatars"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось создать директорию для сохранения",
		})
	}

	// Сохранение файла
	filename := fmt.Sprintf("%s_%s", userID, fileHeader.Filename)
	filepath := fmt.Sprintf("%s/%s", saveDir, filename)

	if err := c.SaveFile(fileHeader, filepath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось сохранить файл",
		})
	}

	if err := initializers.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("picture", filename).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось сохранить имя файла в базе",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Аватар успешно обновлён",
		"avatar":  filename,
	})
}


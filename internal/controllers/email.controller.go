package controllers

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"
	"log"
	"go_back/internal/models"       
	"go_back/internal/initializers"

	"github.com/gofiber/fiber/v2"
)

var verificationCodes = make(map[string]string)
var user models.User

func SendVerificationCode(c *fiber.Ctx) error {
	type Request struct {
		Email string `json:"email"`
		From  string `json:"from"` // "register" или "recover"
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный запрос"})
	}

	email := strings.ToLower(body.Email)

	// Если это восстановление, проверяем что пользователь существует
	if body.From == "recover" {
		var user models.User
		if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Пользователь не найден"})
		}
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	verificationCodes[email] = code

	from := "ryzhovcodesender@gmail.com"
	password := "uwmmiwuqjexxjhob"
	msg := []byte("Subject: Код подтверждения\n\nВаш код: " + code)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{email}, msg)
	if err != nil {
		log.Println("Ошибка SMTP:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка отправки письма"})
	}

	

	return c.JSON(fiber.Map{"message": "Код отправлен"})
}



func VerifyCode(c *fiber.Ctx) error {
	type Request struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат"})
	}

	expected := verificationCodes[strings.ToLower(body.Email)]
	if expected == "" || expected != body.Code {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный код"})
	}

	delete(verificationCodes, body.Email) // Удаляем код после подтверждения
	return c.JSON(fiber.Map{"message": "Email подтверждён"})
}

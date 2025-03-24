package controllers

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var verificationCodes = make(map[string]string)

func SendVerificationCode(c *fiber.Ctx) error {
	type Request struct {
		Email string `json:"email"`
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный запрос"})
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	verificationCodes[strings.ToLower(body.Email)] = code

	// Тут отправка письма
	from := "ryzhovcodesender@gmail.com"
	password := "exrgjhlfgebyufka"
	to := body.Email
	msg := []byte("Subject: Код подтверждения\n\nВаш код: " + code)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
	if err != nil {
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

package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go_back/internal/models"
	"go_back/internal/initializers"
)

// Получить все колонки в комнате
func GetColumns(c *fiber.Ctx) error {
	roomID := c.Query("roomId")
	var columns []models.Column
	initializers.DB.Preload("Tasks").Where("room_id = ?", roomID).Find(&columns)
	return c.JSON(columns)
}

// Создать колонку
func CreateColumn(c *fiber.Ctx) error {
	var column models.Column
	if err := c.BodyParser(&column); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	column.ID = uuid.New()
	initializers.DB.Create(&column)
	return c.JSON(column)
}

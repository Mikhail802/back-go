package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go_back/internal/models"
	"go_back/internal/initializers"
)

// Получение записей по комнате
func GetEntries(c *fiber.Ctx) error {
	roomID := c.Query("roomId")
	var entries []models.Entry
	initializers.DB.Where("room_id = ?", roomID).Find(&entries)
	return c.JSON(entries)
}

// Создание записи
func CreateEntry(c *fiber.Ctx) error {
	var entry models.Entry
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	entry.ID = uuid.New()
	initializers.DB.Create(&entry)
	return c.JSON(entry)
}

// Удаление записи
func DeleteEntry(c *fiber.Ctx) error {
	id := c.Params("entryId")
	initializers.DB.Delete(&models.Entry{}, "id = ?", id)
	return c.JSON(fiber.Map{"message": "Entry deleted"})
}
